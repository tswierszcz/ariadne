package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/component"
	"ariadne/internal/hierarchy"
	"ariadne/internal/reach"
	"ariadne/internal/report"
	"ariadne/internal/runtimeevents"
)

// loadRuntimeTrace reads and validates -runtime-trace before the analysis,
// so a bad file fails early (RUNTIME_NATIVE N1).
func loadRuntimeTrace(path string) (*runtimeevents.Trace, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("-runtime-trace: %w", err)
	}
	if st.Size() > runtimeevents.MaxBytes {
		return nil, fmt.Errorf("-runtime-trace %s: %d bytes, more than the %d-byte limit", path, st.Size(), runtimeevents.MaxBytes)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("-runtime-trace: %w", err)
	}
	sum := sha256.Sum256(b)
	t, err := runtimeevents.Parse(bytes.NewReader(b), hex.EncodeToString(sum[:]))
	if err != nil {
		return nil, fmt.Errorf("-runtime-trace %s: %w", path, err)
	}
	return t, nil
}

// classLookup answers runtimeevents.Inputs.Lookup from the analyzed
// classes: the SHA-256 of the class file actually read (its jar entry) and
// of the jar it was read from. Jars are opened once.
func classLookup(h *hierarchy.Hierarchy) (func(string) (string, string, bool), func()) {
	jars := map[string]*zip.ReadCloser{}
	jarSHA := map[string]string{}
	classSHA := map[string]string{}
	closeAll := func() {
		for _, z := range jars {
			z.Close()
		}
	}
	return func(fqcn string) (string, string, bool) {
		ci, ok := h.Lookup(fqcn)
		if !ok {
			return "", "", false
		}
		if _, done := jarSHA[ci.SourceJar]; !done {
			b, err := os.ReadFile(ci.SourceJar)
			if err == nil {
				s := sha256.Sum256(b)
				jarSHA[ci.SourceJar] = hex.EncodeToString(s[:])
			} else {
				jarSHA[ci.SourceJar] = ""
			}
		}
		if cs, done := classSHA[fqcn]; done {
			return cs, jarSHA[ci.SourceJar], true
		}
		cs := ""
		z, ok := jars[ci.SourceJar]
		if !ok {
			if zr, err := zip.OpenReader(ci.SourceJar); err == nil {
				z = zr
				jars[ci.SourceJar] = zr
			}
		}
		if z != nil {
			for _, f := range z.File {
				if f.Name != ci.SourceEntry {
					continue
				}
				if rc, err := f.Open(); err == nil {
					hsh := sha256.New()
					if _, err := io.Copy(hsh, rc); err == nil {
						cs = hex.EncodeToString(hsh.Sum(nil))
					}
					rc.Close()
				}
				break
			}
		}
		classSHA[fqcn] = cs
		return cs, jarSHA[ci.SourceJar], true
	}, closeAll
}

// runtimeObserved: per advisory and component instance, the mapped
// vulnerable methods seen executing in the trace.
func runtimeObserved(rt runtimeevents.Result, results []reach.Result, h *hierarchy.Hierarchy, inv *component.Inventory) map[string]map[string][]runtimeevents.MethodObservation {
	out := map[string]map[string][]runtimeevents.MethodObservation{}
	for _, r := range results {
		for _, o := range rt.ObservedVulnerable(r.Methods) {
			owner, _, _ := strings.Cut(o.Method, "#")
			inst := ""
			if ci, ok := h.Lookup(owner); ok && inv != nil {
				inst = inv.InstanceForClass(ci)
			}
			if out[r.CVE] == nil {
				out[r.CVE] = map[string][]runtimeevents.MethodObservation{}
			}
			out[r.CVE][inst] = append(out[r.CVE][inst], o)
		}
	}
	return out
}

func runtimeFindings(obs map[string]map[string][]runtimeevents.MethodObservation, results []reach.Result) []report.RuntimeFinding {
	var out []report.RuntimeFinding
	for _, r := range results {
		var all []runtimeevents.MethodObservation
		for _, os := range obs[r.CVE] {
			all = append(all, os...)
		}
		if len(all) == 0 {
			continue
		}
		sort.Slice(all, func(i, j int) bool { return all[i].Evidence < all[j].Evidence })
		out = append(out, report.RuntimeFinding{Advisory: r.CVE, Status: string(r.Status), Observed: all})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Advisory < out[j].Advisory })
	return out
}

// Default statuses whose mapped methods -runtime-filter-out traces: every
// verdict with a mapping that could still be wrong either way.
const defaultRuntimeFilterStatus = "REACHABLE,FRAMEWORK,NOT_REACHABLE,UNKNOWN"

// RuntimeFilter is the exact-entry tracing selection for a later recorded
// run (RUNTIME_NATIVE N1): JFR jdk.MethodTrace filters ("Class::method")
// for the mapped vulnerable methods of the chosen verdicts whose class is
// in the analyzed input.
type RuntimeFilter struct {
	Filters    []string // sorted, deduplicated "Class::method"
	Advisories int
	Skipped    int // mapped methods whose class is not in the input
}

func runtimeFilter(results []reach.Result, h *hierarchy.Hierarchy, statuses string) RuntimeFilter {
	want := map[string]bool{}
	for _, s := range strings.Split(statuses, ",") {
		if s = strings.TrimSpace(s); s != "" {
			want[s] = true
		}
	}
	set := map[string]bool{}
	var f RuntimeFilter
	for _, r := range results {
		if !want[string(r.Status)] || len(r.Methods) == 0 {
			continue
		}
		added := false
		for _, m := range r.Methods {
			owner, rest, ok := strings.Cut(m, "#")
			if !ok {
				continue
			}
			name := rest
			if i := strings.IndexByte(rest, '('); i >= 0 {
				name = rest[:i]
			}
			if _, ok := h.Lookup(owner); !ok || name == "" {
				f.Skipped++
				continue
			}
			set[owner+"::"+name] = true
			added = true
		}
		if added {
			f.Advisories++
		}
	}
	for k := range set {
		f.Filters = append(f.Filters, k)
	}
	sort.Strings(f.Filters)
	return f
}

// jfcFile renders the filter as a JFR settings file, to be given next to
// the usual settings: -XX:StartFlightRecording=settings=profile,settings=FILE.
func (f RuntimeFilter) jfcFile(statuses string) []byte {
	var b strings.Builder
	esc := func(s string) string {
		return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
	}
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, `<configuration version="2.0" label="ariadne vulnerable methods" description="%s">`+"\n",
		esc(fmt.Sprintf("ariadne -runtime-filter-out: exact entries (jdk.MethodTrace, JDK 25+) of %d mapped vulnerable methods of %d advisories with verdicts %s", len(f.Filters), f.Advisories, statuses)))
	b.WriteString(`  <event name="jdk.MethodTrace">` + "\n")
	b.WriteString(`    <setting name="enabled">` + fmt.Sprint(len(f.Filters) > 0) + "</setting>\n")
	b.WriteString(`    <setting name="stackTrace">true</setting>` + "\n")
	b.WriteString(`    <setting name="threshold">0 ms</setting>` + "\n")
	b.WriteString(`    <setting name="filter">` + esc(strings.Join(f.Filters, ";")) + "</setting>\n")
	b.WriteString("  </event>\n</configuration>\n")
	return []byte(b.String())
}

// tracedBy reports whether a header filter ("Class::method" or a bare
// class) covers the mapped method "Class#method[(desc)]".
func tracedBy(filters []string, mapped string) bool {
	owner, rest, ok := strings.Cut(mapped, "#")
	if !ok {
		return false
	}
	name := rest
	if i := strings.IndexByte(rest, '('); i >= 0 {
		name = rest[:i]
	}
	for _, f := range filters {
		if f == owner+"::"+name || f == owner {
			return true
		}
	}
	return false
}

// runtimeTraced: per advisory and instance, whether exact-entry tracing
// covered a mapped method in the recorded run. Instances with the
// advisory's mapped class are those the finding is about; without an
// inventory every instance of the advisory is marked.
func runtimeTraced(rt runtimeevents.Result, results []reach.Result, h *hierarchy.Hierarchy, inv *component.Inventory) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	if len(rt.MethodFilters) == 0 {
		return out
	}
	for _, r := range results {
		for _, m := range r.Methods {
			if !tracedBy(rt.MethodFilters, m) {
				continue
			}
			owner, _, _ := strings.Cut(m, "#")
			inst := ""
			if ci, ok := h.Lookup(owner); ok && inv != nil {
				inst = inv.InstanceForClass(ci)
			}
			if out[r.CVE] == nil {
				out[r.CVE] = map[string]bool{}
			}
			out[r.CVE][inst] = true
		}
	}
	return out
}

// tracedNoEntry: advisories whose mapped methods were traced on entry in
// the recorded run and never entered — not evidence of unreachability.
func tracedNoEntry(traced map[string]map[string]bool, obs map[string]map[string][]runtimeevents.MethodObservation, results []reach.Result) []report.RuntimeFinding {
	var out []report.RuntimeFinding
	for _, r := range results {
		if len(traced[r.CVE]) == 0 {
			continue
		}
		entered := false
		for _, os := range obs[r.CVE] {
			for _, o := range os {
				if o.Basis == "traced" {
					entered = true
				}
			}
		}
		if !entered {
			out = append(out, report.RuntimeFinding{Advisory: r.CVE, Status: string(r.Status)})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Advisory < out[j].Advisory })
	return out
}

// gapClassifier decides why the static graph has no edge for a measured
// call, from the caller's bytecode at the measured index (runtimeevents
// GapCauseVersion). mode is the graph's -callgraph-mode.
func gapClassifier(h *hierarchy.Hierarchy, mode string) func(caller, callee string, bci int, basis string) (string, string) {
	split := func(key string) (owner, name, desc string) {
		owner, rest, _ := strings.Cut(key, "#")
		i := strings.IndexByte(rest, '(')
		if i < 0 {
			return owner, rest, ""
		}
		return owner, rest[:i], rest[i:]
	}
	isSubtype := func(cls, of string) bool {
		if cls == of {
			return true
		}
		for _, s := range h.Subtypes(of) {
			if s == cls {
				return true
			}
		}
		return false
	}
	implements := func(cls, iface string) bool {
		seen := map[string]bool{}
		var walk func(string) bool
		walk = func(c string) bool {
			if c == "" || seen[c] {
				return false
			}
			if c == iface {
				return true
			}
			seen[c] = true
			ci, ok := h.Lookup(c)
			if !ok {
				return false
			}
			if walk(ci.SuperClass) {
				return true
			}
			for _, i := range ci.Interfaces {
				if walk(i) {
					return true
				}
			}
			return false
		}
		return walk(cls)
	}
	asyncOwners := map[string]bool{
		"java.lang.Thread": true, "java.util.concurrent.Executor": true, "java.util.concurrent.ExecutorService": true,
		"java.util.concurrent.ScheduledExecutorService": true, "java.util.concurrent.CompletableFuture": true,
		"java.util.concurrent.ThreadPoolExecutor": true, "java.util.concurrent.ForkJoinPool": true, "java.util.Timer": true,
	}
	return func(caller, callee string, bci int, basis string) (string, string) {
		cOwner, cName, cDesc := split(caller)
		eOwner, eName, eDesc := split(callee)
		if eName == "<clinit>" {
			return runtimeevents.GapClassInit, eOwner
		}
		ci, ok := h.Lookup(cOwner)
		if !ok {
			return runtimeevents.GapUnclassified, "caller not in the input"
		}
		m, ok := ci.MethodByNameAndDesc(cName, cDesc)
		if !ok || bci < 0 {
			return runtimeevents.GapNoCallAtSite, "no bytecode index"
		}
		var site *classfile.Call
		for i := range m.Calls {
			if m.Calls[i].PC == bci {
				site = &m.Calls[i]
				break
			}
		}
		if site == nil {
			if eName == "<init>" {
				return runtimeevents.GapClassInit, eOwner
			}
			return runtimeevents.GapNoCallAtSite, cOwner + "#" + cName
		}
		target := site.Owner + "#" + site.Name
		switch {
		case site.Kind == classfile.CallDynamic:
			return runtimeevents.GapInvokeDynamic, site.Owner
		case target == "java.lang.reflect.Method#invoke" || target == "java.lang.reflect.Constructor#newInstance" || target == "java.lang.Class#newInstance":
			return runtimeevents.GapReflection, target
		case site.Owner == "java.lang.invoke.MethodHandle":
			return runtimeevents.GapMethodHandle, target
		case eName == "invoke" && implements(eOwner, "java.lang.reflect.InvocationHandler") && site.Owner != eOwner:
			return runtimeevents.GapProxy, site.Owner
		case asyncOwners[site.Owner] && basis == runtimeevents.BasisThroughRuntime:
			return runtimeevents.GapAsync, target
		}
		virtual := site.Kind == classfile.CallVirtual || site.Kind == classfile.CallInterface
		// A lambda body, or a method-reference target whose name differs
		// from the interface method: a functional-interface call.
		if virtual && (strings.HasPrefix(eName, "lambda$") || (eName != site.Name && !isSubtype(eOwner, site.Owner))) {
			return runtimeevents.GapSAMCall, site.Owner
		}
		if callgraph.DefaultSkip(site.Owner) {
			return runtimeevents.GapSkippedReceiver, site.Owner
		}
		// The body the JVM selected comes from a superclass that does not
		// itself implement the call's type: some subtype inherits it.
		if virtual && site.Name == eName && site.Desc == eDesc && !isSubtype(eOwner, site.Owner) {
			for _, s := range h.Subtypes(site.Owner) {
				if decl, ok := h.ResolveExact(s, site.Name, site.Desc); ok && decl == eOwner {
					return runtimeevents.GapInherited, site.Owner
				}
			}
		}
		if site.Name == eName && site.Desc == eDesc && isSubtype(eOwner, site.Owner) {
			if site.Kind == classfile.CallVirtual || site.Kind == classfile.CallInterface {
				if mode == callgraph.ModeRTA {
					return runtimeevents.GapRTAPruned, site.Owner
				}
				return runtimeevents.GapDispatchMissing, site.Owner
			}
		}
		return runtimeevents.GapIndirect, target
	}
}
