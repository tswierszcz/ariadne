package deserialize

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/dataflow"
	"ariadne/internal/hierarchy"
	"ariadne/internal/mavenversion"
	"ariadne/internal/valueflow"
)

// DESERIALIZATION Z2: what an entry's filter or security configuration
// allows, as far as the bytecode states it. Only a policy read from
// literals in the entry's own method, on the same object, is evaluated;
// anything else that could restrict types is "unknown" and keeps every
// edge (a policy that cannot be read is not a proof that a class is
// rejected). A class the evaluated policy rejects gets no edge from
// that entry.

// Policy states of a site (Site.Filter).
const (
	// FilterNoneSeen: no filter or configuration visible (Java: a
	// process-wide jdk.serialFilter is still possible).
	FilterNoneSeen = "none_seen"
	// FilterUnknown: a filter/configuration exists that is not evaluated
	// (dynamic, a resolveClass override, an object configured or created
	// elsewhere); every edge is kept.
	FilterUnknown = "unknown"
	// FilterEvaluated: a policy read from literals decides each class.
	FilterEvaluated = "evaluated"
	// FilterUnrestricted (XStream): no type restriction — a release
	// without the security framework (< 1.4.7) or one whose default
	// allows any type (1.4.7–1.4.17) with nothing configured.
	FilterUnrestricted = "unrestricted"
	// FilterDefaultAllowlist (XStream >= 1.4.18 with nothing added): the
	// default allowlist admits only JDK basic types, so no class of the
	// input.
	FilterDefaultAllowlist = "default_allowlist"
	// FilterVersionUnknown (XStream): the release could not be read;
	// every edge is kept.
	FilterVersionUnknown = "version_unknown"
)

// FilterNotEvaluated is the Z1 name of FilterUnknown.
const FilterNotEvaluated = FilterUnknown

// policy decides one class: true (allowed) or false (rejected).
type policy func(fqcn string) bool

const (
	xstreamClass = "com.thoughtworks.xstream.XStream"
	xsPerm       = "com.thoughtworks.xstream.security."
	filterConfig = "java.io.ObjectInputFilter$Config"
)

// --- Java: ObjectInputFilter patterns (JEP 290) ---------------------

type javaPattern struct {
	reject bool
	kind   string // exact | pkg | subpkg | prefix | any
	text   string
}

// parseJavaFilter parses a filter pattern string; limit patterns
// (maxdepth=…) do not decide classes and are skipped; a module-qualified
// pattern is not supported (ok=false).
func parseJavaFilter(s string) (pats []javaPattern, ok bool) {
	for _, p := range strings.Split(s, ";") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "=") {
			continue // a limit: maxdepth, maxrefs, maxbytes, maxarray
		}
		jp := javaPattern{}
		if strings.HasPrefix(p, "!") {
			jp.reject, p = true, p[1:]
		}
		if strings.Contains(p, "/") {
			return nil, false
		}
		switch {
		case p == "*":
			jp.kind = "any"
		case strings.HasSuffix(p, ".**"):
			jp.kind, jp.text = "subpkg", strings.TrimSuffix(p, ".**")
		case strings.HasSuffix(p, ".*"):
			jp.kind, jp.text = "pkg", strings.TrimSuffix(p, ".*")
		case strings.HasSuffix(p, "*"):
			jp.kind, jp.text = "prefix", strings.TrimSuffix(p, "*")
		default:
			jp.kind, jp.text = "exact", p
		}
		pats = append(pats, jp)
	}
	return pats, true
}

// javaFilterPolicy: the first matching pattern decides; no match is
// UNDECIDED, which an ObjectInputStream treats as allowed when no other
// filter decides.
func javaFilterPolicy(pats []javaPattern) policy {
	return func(fqcn string) bool {
		pkg := ""
		if i := strings.LastIndexByte(fqcn, '.'); i >= 0 {
			pkg = fqcn[:i]
		}
		for _, p := range pats {
			match := false
			switch p.kind {
			case "any":
				match = true
			case "exact":
				match = fqcn == p.text
			case "pkg":
				match = pkg == p.text
			case "subpkg":
				match = pkg == p.text || strings.HasPrefix(pkg, p.text+".")
			case "prefix":
				match = strings.HasPrefix(fqcn, p.text)
			}
			if match {
				return !p.reject
			}
		}
		return true
	}
}

// --- XStream: TypePermission list ----------------------------------

type xsPermission struct {
	deny bool
	// any matches every type (AnyTypePermission.ANY / NoTypePermission.NONE)
	any       bool
	names     map[string]bool
	wildcards []*regexp.Regexp
	hierarchy string
}

func (p xsPermission) matches(h *hierarchy.Hierarchy, fqcn string) bool {
	if p.any || p.names[fqcn] {
		return true
	}
	for _, re := range p.wildcards {
		if re.MatchString(fqcn) {
			return true
		}
	}
	if p.hierarchy != "" {
		if fqcn == p.hierarchy {
			return true
		}
		for _, s := range h.Subtypes(p.hierarchy) {
			if s == fqcn {
				return true
			}
		}
	}
	return false
}

// xsWildcard converts an XStream wildcard ("**" any characters, "*" any
// but '.', "?" one character but '.') to a regexp.
func xsWildcard(w string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(w); i++ {
		switch {
		case strings.HasPrefix(w[i:], "**"):
			b.WriteString(".*")
			i++
		case w[i] == '*':
			b.WriteString(`[^.]*`)
		case w[i] == '?':
			b.WriteString(`[^.]`)
		default:
			b.WriteString(regexp.QuoteMeta(w[i : i+1]))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

// xsPolicy: XStream's SecurityMapper checks the permissions from the
// last added to the first; the first that matches decides (a deny
// permission rejects); none matching rejects. NONE and ANY clear the
// earlier ones when added.
func xsPolicy(h *hierarchy.Hierarchy, perms []xsPermission) policy {
	return func(fqcn string) bool {
		for i := len(perms) - 1; i >= 0; i-- {
			if perms[i].matches(h, fqcn) {
				return !perms[i].deny
			}
		}
		return false
	}
}

func addPerm(perms []xsPermission, p xsPermission) []xsPermission {
	if p.any {
		perms = nil
	}
	return append(perms, p)
}

// --- resolution -----------------------------------------------------

func sameAlloc(a, b valueflow.Value) bool {
	return a.Kind == valueflow.KindAlloc && b.Kind == valueflow.KindAlloc && a.AllocSite != nil && b.AllocSite != nil && *a.AllocSite == *b.AllocSite
}

func dotted(s string) string { return strings.ReplaceAll(s, "/", ".") }

// constStrings returns the elements of a literal array of strings or
// class literals.
func constStrings(v valueflow.Value) ([]string, bool) {
	if v.Kind != valueflow.KindArray || v.Array == nil {
		return nil, false
	}
	var out []string
	for _, e := range v.Array.Elements {
		s, ok := e.Const.(string)
		if e.Kind != valueflow.KindConst || !ok {
			return nil, false
		}
		out = append(out, dotted(s))
	}
	return out, true
}

func argCount(desc string) int {
	n := 0
	for i := 1; i < len(desc) && desc[i] != ')'; i++ {
		switch desc[i] {
		case 'L':
			for desc[i] != ';' {
				i++
			}
			n++
		case '[':
			for desc[i] == '[' {
				i++
			}
			if desc[i] == 'L' {
				for desc[i] != ';' {
					i++
				}
			}
			n++
		default:
			n++
		}
	}
	return n
}

// ResolvePolicies sets each site's policy (Filter, FilterDetail,
// Version) from its own method's bytecode. It re-reads the class files
// of the methods with sites; a class or method it cannot analyze leaves
// the Z1 detection (unknown when something was seen).
func ResolvePolicies(h *hierarchy.Hierarchy, sites []Site) []string {
	var warnings []string
	var jr classfile.JarClassReader
	defer jr.Close()
	xsVersion := ""
	if ci, ok := h.Lookup(xstreamClass); ok && ci.GAV != nil {
		xsVersion = ci.GAV.Version
	}
	type key struct{ fqcn, name, desc string }
	facts := map[key]*dataflow.MethodFacts{}
	failed := map[key]string{}
	factsFor := func(s *Site) (*dataflow.MethodFacts, *hierarchy.Method, string) {
		k := key{s.CallerFQCN, s.CallerName, s.CallerDesc}
		ci, _ := h.Lookup(s.CallerFQCN)
		m, ok := ci.MethodByNameAndDesc(s.CallerName, s.CallerDesc)
		if !ok {
			return nil, nil, "method not found"
		}
		if f, ok := facts[k]; ok {
			return f, &m, ""
		}
		if r, ok := failed[k]; ok {
			return nil, &m, r
		}
		cf, err := jr.ReadClass(ci.SourceJar, ci.SourceEntry)
		if err != nil {
			failed[k] = err.Error()
			warnings = append(warnings, fmt.Sprintf("deserialization policy: re-reading %s: %v", s.CallerFQCN, err))
			return nil, &m, failed[k]
		}
		f, err := dataflow.Analyze(cf, s.CallerFQCN, m.Method)
		if err != nil {
			failed[k] = err.Error()
			return nil, &m, failed[k]
		}
		if ab, reason := f.Aborted(); ab {
			failed[k] = "local-value analysis aborted: " + reason
			return nil, &m, failed[k]
		}
		facts[k] = f
		return f, &m, ""
	}
	for i := range sites {
		s := &sites[i]
		switch s.Model {
		case "jdk:java-serialization":
			resolveJava(h, s, factsFor)
		case "xstream:fromxml":
			s.Version = xsVersion
			resolveXStream(h, s, xsVersion, factsFor)
		}
	}
	return warnings
}

type factsFunc func(*Site) (*dataflow.MethodFacts, *hierarchy.Method, string)

// argAt/recvAt read a call's facts; a call the analysis did not record
// (e.g. in unreachable code) is unknown, not a crash.
func argAt(f *dataflow.MethodFacts, at bytecodeir.Site, i int) (v valueflow.Value) {
	defer func() {
		if recover() != nil {
			v = valueflow.Unknown("call not recorded by the local-value analysis")
		}
	}()
	return f.ArgumentAt(at, i)
}

func recvAt(f *dataflow.MethodFacts, at bytecodeir.Site) (v valueflow.Value) {
	defer func() {
		if recover() != nil {
			v = valueflow.Unknown("call not recorded by the local-value analysis")
		}
	}()
	return f.ReceiverAt(at)
}

func callSite(s *Site, pc int) bytecodeir.Site {
	return bytecodeir.Site{Method: s.CallerFQCN + "#" + s.CallerName, Descriptor: s.CallerDesc, PC: pc, Op: bytecodeir.OpCall}
}

func resolveJava(h *hierarchy.Hierarchy, s *Site, factsFor factsFunc) {
	f, m, reason := factsFor(s)
	if f == nil {
		if s.Filter == FilterUnknown && reason != "" {
			s.FilterDetail += "; not evaluated: " + reason
		}
		return
	}
	var recv valueflow.Value
	if s.Owner != objectInput {
		recv = recvAt(f, callSite(s, s.PC))
	}
	var details []string
	if s.FilterDetail != "" {
		for _, d := range strings.Split(s.FilterDetail, "; ") {
			if !strings.HasPrefix(d, "setObjectInputFilter") {
				details = append(details, d+" (not evaluated)")
			}
		}
	}
	var evaluated policy
	unknown := len(details) > 0
	for _, c := range m.Calls {
		if c.PC >= s.PC || c.Name != "setObjectInputFilter" || c.Desc != "(Ljava/io/ObjectInputFilter;)V" || !isOISSubtype(h, c.Owner) {
			continue
		}
		fr := recvAt(f, callSite(s, c.PC))
		if !sameAlloc(fr, recv) {
			if recv.Kind == valueflow.KindAlloc && fr.Kind == valueflow.KindAlloc {
				continue // a filter on another stream
			}
			unknown = true
			details = append(details, "setObjectInputFilter on a stream not proven to be this one")
			continue
		}
		arg := argAt(f, callSite(s, c.PC), 0)
		if arg.Kind == valueflow.KindCallResult && arg.Call != nil && arg.Call.Owner == filterConfig && arg.Call.Name == "createFilter" && arg.Call.Desc == "(Ljava/lang/String;)Ljava/io/ObjectInputFilter;" {
			lit := argAt(f, arg.Call.Site, 0)
			if str, ok := lit.Const.(string); lit.Kind == valueflow.KindConst && ok {
				if pats, ok := parseJavaFilter(str); ok {
					evaluated = javaFilterPolicy(pats)
					details = append(details, fmt.Sprintf("setObjectInputFilter(createFilter(%q))", str))
					continue
				}
				details = append(details, fmt.Sprintf("createFilter(%q): module-qualified patterns not evaluated", str))
				unknown = true
				continue
			}
		}
		unknown = true
		details = append(details, "setObjectInputFilter with a filter that is not a literal createFilter(\"…\")")
	}
	switch {
	case evaluated != nil:
		s.Filter, s.allow = FilterEvaluated, evaluated
		if unknown {
			details = append(details, "other restrictions not evaluated: only the literal filter's rejections remove edges")
		}
	case unknown:
		s.Filter = FilterUnknown
	default:
		s.Filter = FilterNoneSeen
	}
	sort.Strings(details)
	s.FilterDetail = strings.Join(details, "; ")
}

func resolveXStream(h *hierarchy.Hierarchy, s *Site, version string, factsFor factsFunc) {
	switch {
	case version == "" || !mavenversion.Regular(version):
		s.Filter, s.FilterDetail = FilterVersionUnknown, "XStream release not identified: its default security is not known"
		return
	case mavenversion.Compare(version, "1.4.7") < 0:
		s.Filter, s.FilterDetail = FilterUnrestricted, "XStream "+version+" has no security framework (added in 1.4.7): any type"
		return
	}
	modern := mavenversion.Compare(version, "1.4.18") >= 0
	var perms []xsPermission
	if modern {
		perms = []xsPermission{{deny: true, any: true}} // default allowlist: JDK basic types only
	} else {
		perms = []xsPermission{{any: true}}
	}
	f, m, reason := factsFor(s)
	if f == nil {
		s.Filter, s.FilterDetail = FilterUnknown, "not evaluated: "+reason
		return
	}
	recv := recvAt(f, callSite(s, s.PC))
	if recv.Kind != valueflow.KindAlloc {
		s.Filter, s.FilterDetail = FilterUnknown, "the XStream instance is not created in this method (configured elsewhere?)"
		return
	}
	var details []string
	unknown := false
	configured := false
	for _, c := range m.Calls {
		if c.Kind == classfile.CallDynamic || c.PC == s.PC {
			continue
		}
		at := callSite(s, c.PC)
		n := argCount(c.Desc)
		isConfig := c.Owner == xstreamClass && (strings.HasPrefix(c.Name, "allow") || strings.HasPrefix(c.Name, "deny") || c.Name == "addPermission" || c.Name == "setupDefaultSecurity")
		if !isConfig {
			// escape: the instance handed to other code may be configured there
			for a := 0; a < n; a++ {
				if sameAlloc(argAt(f, at, a), recv) {
					unknown = true
					details = append(details, fmt.Sprintf("the instance is passed to %s#%s", c.Owner, c.Name))
				}
			}
			continue
		}
		if c.PC > s.PC {
			continue
		}
		if c.Name == "setupDefaultSecurity" && c.Kind == classfile.CallStatic {
			if n == 1 && sameAlloc(argAt(f, at, 0), recv) {
				perms = addPerm(perms, xsPermission{deny: true, any: true})
				configured = true
				details = append(details, "setupDefaultSecurity")
			}
			continue
		}
		if c.Kind == classfile.CallStatic || !sameAlloc(recvAt(f, at), recv) {
			continue
		}
		configured = true
		deny := strings.HasPrefix(c.Name, "deny")
		var arg valueflow.Value
		if n == 1 {
			arg = argAt(f, at, 0)
		}
		switch {
		case c.Name == "addPermission" && arg.Kind == valueflow.KindField && arg.Field != nil && arg.Field.Owner == xsPerm+"NoTypePermission" && arg.Field.Name == "NONE":
			perms = addPerm(perms, xsPermission{deny: true, any: true})
			details = append(details, "addPermission(NoTypePermission.NONE)")
		case c.Name == "addPermission" && arg.Kind == valueflow.KindField && arg.Field != nil && arg.Field.Owner == xsPerm+"AnyTypePermission" && arg.Field.Name == "ANY":
			perms = addPerm(perms, xsPermission{any: true})
			details = append(details, "addPermission(AnyTypePermission.ANY)")
		case (c.Name == "allowTypes" || c.Name == "denyTypes") && (c.Desc == "([Ljava/lang/Class;)V" || c.Desc == "([Ljava/lang/String;)V"):
			names, ok := constStrings(arg)
			if !ok {
				unknown = true
				details = append(details, c.Name+" with a non-literal array")
				continue
			}
			set := map[string]bool{}
			for _, n := range names {
				set[n] = true
			}
			perms = addPerm(perms, xsPermission{deny: deny, names: set})
			details = append(details, fmt.Sprintf("%s(%s)", c.Name, strings.Join(names, ", ")))
		case (c.Name == "allowTypesByWildcard" || c.Name == "denyTypesByWildcard") && c.Desc == "([Ljava/lang/String;)V":
			ws, ok := constStrings(arg)
			if !ok {
				unknown = true
				details = append(details, c.Name+" with a non-literal array")
				continue
			}
			p := xsPermission{deny: deny}
			for _, w := range ws {
				p.wildcards = append(p.wildcards, xsWildcard(w))
			}
			perms = addPerm(perms, p)
			details = append(details, fmt.Sprintf("%s(%s)", c.Name, strings.Join(ws, ", ")))
		case (c.Name == "allowTypeHierarchy" || c.Name == "denyTypeHierarchy") && c.Desc == "(Ljava/lang/Class;)V":
			cls, ok := arg.Const.(string)
			if arg.Kind != valueflow.KindConst || !ok {
				unknown = true
				details = append(details, c.Name+" with a non-literal class")
				continue
			}
			perms = addPerm(perms, xsPermission{deny: deny, hierarchy: dotted(cls)})
			details = append(details, fmt.Sprintf("%s(%s)", c.Name, dotted(cls)))
		default:
			unknown = true
			details = append(details, c.Name+c.Desc+" not evaluated")
		}
	}
	sort.Strings(details)
	switch {
	case unknown:
		s.Filter = FilterUnknown
	case !configured && modern:
		s.Filter = FilterDefaultAllowlist
		details = append(details, "XStream "+version+" default allowlist (JDK basic types only)")
		s.allow = xsPolicy(h, perms)
	case !configured:
		s.Filter = FilterUnrestricted
		details = append(details, "XStream "+version+" allows any type unless restricted")
	default:
		s.Filter, s.allow = FilterEvaluated, xsPolicy(h, perms)
		details = append([]string{"XStream " + version}, details...)
	}
	s.FilterDetail = strings.Join(details, "; ")
}
