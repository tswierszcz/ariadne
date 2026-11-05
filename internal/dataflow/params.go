package dataflow

import (
	"fmt"
	"sort"
	"strings"

	"ariadne/internal/bytecodeir"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/valueflow"
)

// Parameter-propagation limits (DATA_FLOW D3, slice "parameters from
// callers"). Exceeding either yields an explicit unknown, never a
// truncated set presented as complete.
const (
	MaxParameterDepth   = 3  // caller levels followed when an argument is itself a parameter
	MaxParameterCallers = 16 // distinct call sites accepted for one method
)

// MethodID is one exact method: its declaring class, name and descriptor.
type MethodID struct{ Owner, Name, Desc string }

func (m MethodID) String() string { return m.Owner + "#" + m.Name + m.Desc }

type callerSite struct {
	ci  *hierarchy.ClassInfo
	m   hierarchy.Method
	pc  int
	via classfile.CallKind
}

// CallerIndex records, for every method in the analyzed input, the
// call sites that reach it with a statically exact target, plus methods
// whose callers cannot all be seen (referenced by invokedynamic as a
// lambda/method-reference implementation). It is built once per run.
type CallerIndex struct {
	h       *hierarchy.Hierarchy
	callers map[MethodID][]callerSite
	hidden  map[MethodID]bool
	facts   map[MethodID]*MethodFacts
	failed  map[MethodID]string
	reader  classfile.JarClassReader
}

// NewCallerIndex scans every method's call list. Call targets are keyed
// by the class that actually declares them (hierarchy.ResolveExact), so a
// call naming a subclass still reaches the inherited declaration.
func NewCallerIndex(h *hierarchy.Hierarchy) *CallerIndex {
	idx := &CallerIndex{
		h: h, callers: map[MethodID][]callerSite{}, hidden: map[MethodID]bool{},
		facts: map[MethodID]*MethodFacts{}, failed: map[MethodID]string{},
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for _, m := range ci.Methods {
			for _, c := range m.Calls {
				decl, ok := h.ResolveExact(c.Owner, c.Name, c.Desc)
				if !ok {
					continue
				}
				target := MethodID{decl, c.Name, c.Desc}
				if c.Kind == classfile.CallDynamic {
					idx.hidden[target] = true
					continue
				}
				idx.callers[target] = append(idx.callers[target], callerSite{ci: ci, m: m, pc: c.PC, via: c.Kind})
			}
		}
	}
	return idx
}

// Close releases the jars opened while analyzing callers.
func (idx *CallerIndex) Close() error { return idx.reader.Close() }

// ParameterValues is what the analyzed callers pass for one formal
// parameter: the union of their argument values (context-insensitive —
// one set per parameter, whichever call site), followed through
// arguments that are themselves parameters up to MaxParameterDepth.
// Unknown lists why some contribution could not be determined; when it
// is non-empty, Values is a partial set and consumers must keep a gap.
type ParameterValues struct {
	Values  []valueflow.Value
	Unknown []string
	Callers int // distinct call sites contributing, across all levels
}

// Complete reports whether every contribution was determined.
func (p ParameterValues) Complete() bool { return len(p.Unknown) == 0 && len(p.Values) > 0 }

// ResolveParameter returns the values callers in the analyzed input pass
// as parameter index (zero-based, receiver excluded — the same indexing
// as valueflow.ParameterRef) of method. It assumes the analyzed input
// holds every caller (whole-program view); callers through reflection,
// frameworks or code outside the input are not seen. A method is
// resolvable only when its call sites are statically exact: static,
// private and constructor methods, or a virtual method that neither
// overrides nor is overridden. Anything else — and a method used as a
// lambda/method-reference implementation, a method with no analyzed
// caller, too many callers, or too deep a chain — contributes
// an Unknown reason instead of a guess. A recursive edge (a parameter
// passed back into itself) converges: it adds nothing to the union.
func (idx *CallerIndex) ResolveParameter(method MethodID, index int) ParameterValues {
	var out ParameterValues
	idx.resolve(method, index, 0, map[string]bool{}, &out)
	sort.Strings(out.Unknown)
	out.Unknown = uniqueSorted(out.Unknown)
	return out
}

func (idx *CallerIndex) resolve(method MethodID, index, depth int, visiting map[string]bool, out *ParameterValues) {
	key := fmt.Sprintf("%s/%d", method, index)
	if visiting[key] {
		// A parameter flowing back into itself adds no value the
		// non-recursive callers do not already contribute: the union is
		// the fixed point, so the cycle converges without a guess.
		return
	}
	if depth >= MaxParameterDepth {
		out.Unknown = append(out.Unknown, fmt.Sprintf("parameter chain deeper than %d caller levels at %s", MaxParameterDepth, method))
		return
	}
	if reason := idx.closedCallers(method); reason != "" {
		out.Unknown = append(out.Unknown, reason)
		return
	}
	sites := idx.callers[method]
	if len(sites) == 0 {
		out.Unknown = append(out.Unknown, fmt.Sprintf("%s has no caller in the analyzed input (it may be called by a framework, reflection or external code)", method))
		return
	}
	if len(sites) > MaxParameterCallers {
		out.Unknown = append(out.Unknown, fmt.Sprintf("%s has %d callers, above the limit of %d", method, len(sites), MaxParameterCallers))
		return
	}
	visiting[key] = true
	defer delete(visiting, key)
	for _, s := range sites {
		out.Callers++
		v, reason := idx.argumentAt(s, method, index)
		if reason != "" {
			out.Unknown = append(out.Unknown, reason)
			continue
		}
		idx.collect(v, s, depth, visiting, out)
	}
}

// collect adds one argument value; a formal parameter of the method the
// value was read in (s, or owner when given) is resolved one caller
// level up.
func (idx *CallerIndex) collect(v valueflow.Value, s callerSite, depth int, visiting map[string]bool, out *ParameterValues, owner ...MethodID) {
	switch {
	case v.Kind == valueflow.KindSet:
		for _, c := range v.Candidates {
			idx.collect(c, s, depth, visiting, out, owner...)
		}
	case v.Kind == valueflow.KindTyped && v.Parameter != nil:
		m := MethodID{s.ci.FQCN, s.m.Name, s.m.Descriptor}
		if len(owner) > 0 {
			m = owner[0]
		}
		idx.resolve(m, v.Parameter.Index, depth+1, visiting, out)
	case v.Kind == valueflow.KindUnknown:
		out.Unknown = append(out.Unknown, fmt.Sprintf("caller %s#%s%s: %s", s.ci.FQCN, s.m.Name, s.m.Descriptor, v.Reason))
	default:
		for _, existing := range out.Values {
			if existing.Equal(v) {
				return
			}
		}
		out.Values = append(out.Values, v)
	}
}

// closedCallers returns "" when every caller of method is a statically
// exact call site, or why not.
func (idx *CallerIndex) closedCallers(method MethodID) string {
	if idx.hidden[method] {
		return fmt.Sprintf("%s is used as a lambda/method-reference implementation; its invokedynamic callers are not traced", method)
	}
	ci, ok := idx.h.Lookup(method.Owner)
	if !ok {
		return fmt.Sprintf("%s is outside the analyzed hierarchy", method)
	}
	m, ok := ci.MethodByNameAndDesc(method.Name, method.Desc)
	if !ok {
		return fmt.Sprintf("%s is not declared in the analyzed hierarchy", method)
	}
	if m.IsStatic() || m.AccessFlags&accPrivate != 0 || method.Name == "<init>" {
		return ""
	}
	if why := overrides(idx.h, ci, method, map[string]bool{}); why != "" {
		return why
	}
	for _, sub := range idx.h.Subtypes(method.Owner) {
		if sub == method.Owner {
			continue
		}
		if sci, ok := idx.h.Lookup(sub); ok {
			if _, declared := sci.MethodByNameAndDesc(method.Name, method.Desc); declared {
				return fmt.Sprintf("%s is overridden in %s; its virtual call sites do not have one exact target", method, sub)
			}
		}
	}
	return ""
}

const accPrivate = 0x0002

// overrides returns why method may override a supertype method — a
// supertype in the input declaring it, or a supertype outside the input
// (other than java.lang.Object) that might — or "".
func overrides(h *hierarchy.Hierarchy, ci *hierarchy.ClassInfo, method MethodID, seen map[string]bool) string {
	supers := append([]string{}, ci.Interfaces...)
	if ci.SuperClass != "" {
		supers = append(supers, ci.SuperClass)
	}
	for _, s := range supers {
		if seen[s] || s == "java.lang.Object" {
			continue
		}
		seen[s] = true
		sci, ok := h.Lookup(s)
		if !ok {
			return fmt.Sprintf("%s may override a method of %s, which is outside the analyzed input; calls through it are not seen", method, s)
		}
		if _, declared := sci.MethodByNameAndDesc(method.Name, method.Desc); declared {
			return fmt.Sprintf("%s overrides %s; calls through the supertype may reach it", method, s)
		}
		if why := overrides(h, sci, method, seen); why != "" {
			return why
		}
	}
	return ""
}

func (idx *CallerIndex) argumentAt(s callerSite, target MethodID, index int) (v valueflow.Value, reason string) {
	id := MethodID{s.ci.FQCN, s.m.Name, s.m.Descriptor}
	facts, reason := idx.methodFacts(s.ci, s.m)
	if reason != "" {
		return valueflow.Value{}, fmt.Sprintf("caller %s: %s", id, reason)
	}
	if aborted, why := facts.Aborted(); aborted {
		return valueflow.Value{}, fmt.Sprintf("caller %s: local-value analysis aborted: %s", id, why)
	}
	site := bytecodeir.Site{Method: s.ci.FQCN + "#" + s.m.Name, Descriptor: s.m.Descriptor, PC: s.pc, Op: bytecodeir.OpCall}
	defer func() {
		if r := recover(); r != nil {
			v, reason = valueflow.Value{}, fmt.Sprintf("caller %s: no argument recorded at pc=%d for %s", id, s.pc, target)
		}
	}()
	return facts.ArgumentAt(site, index), ""
}

func (idx *CallerIndex) methodFacts(ci *hierarchy.ClassInfo, m hierarchy.Method) (*MethodFacts, string) {
	id := MethodID{ci.FQCN, m.Name, m.Descriptor}
	if f, ok := idx.facts[id]; ok {
		return f, ""
	}
	if r, ok := idx.failed[id]; ok {
		return nil, r
	}
	cf, err := idx.reader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		idx.failed[id] = fmt.Sprintf("re-reading the class failed: %v", err)
		return nil, idx.failed[id]
	}
	f, err := AnalyzeWithParameterRefs(cf, ci.FQCN, m.Method)
	if err != nil {
		idx.failed[id] = fmt.Sprintf("local-value analysis failed: %v", err)
		return nil, idx.failed[id]
	}
	idx.facts[id] = f
	return f, ""
}

func uniqueSorted(in []string) []string {
	var out []string
	for i, s := range in {
		if i > 0 && s == in[i-1] {
			continue
		}
		out = append(out, s)
	}
	return out
}

// Describe renders a short human summary, for reports and reasons.
func (p ParameterValues) Describe() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("%d value(s) from %d caller site(s)", len(p.Values), p.Callers))
	if len(p.Unknown) > 0 {
		parts = append(parts, "incomplete: "+strings.Join(p.Unknown, "; "))
	}
	return strings.Join(parts, ", ")
}

// ResolveFinalField returns what a *final* instance field holds, for the
// shape ResolveInstanceField already accepts (written only by
// `this.field = ...` in its declaring class's constructors), extending it
// to a value that is a constructor parameter: that parameter is resolved
// through the constructor's callers (ResolveParameter, same limits). The
// union covers every `new` site — instance-insensitive, which finality
// and self-receiver writes make safe: each instance holds one of these
// values. Typical case: the synthetic `val$x` fields of an anonymous or
// local class capturing its enclosing method's parameters. A non-final
// field, or no constructor self-write, is an Unknown reason.
func (idx *CallerIndex) ResolveFinalField(ref valueflow.FieldRef) ParameterValues {
	var out ParameterValues
	h := idx.h
	declaring, ok := h.ResolveField(ref.Owner, ref.Name, ref.Desc)
	if !ok {
		out.Unknown = append(out.Unknown, fmt.Sprintf("%s does not resolve field %s:%s", ref.Owner, ref.Name, ref.Desc))
		return out
	}
	ci, ok := h.Lookup(declaring)
	if !ok {
		out.Unknown = append(out.Unknown, fmt.Sprintf("declaring class %s is outside the analyzed hierarchy", declaring))
		return out
	}
	field, ok := ci.FieldByNameAndDesc(ref.Name, ref.Desc)
	if !ok || field.AccessFlags&classfile.AccFinal == 0 || field.AccessFlags&classfile.AccStatic != 0 {
		out.Unknown = append(out.Unknown, fmt.Sprintf("%s#%s is not a final instance field", declaring, ref.Name))
		return out
	}
	cf, err := idx.reader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		out.Unknown = append(out.Unknown, fmt.Sprintf("re-reading %s failed: %v", declaring, err))
		return out
	}
	writes := 0
	for _, m := range ci.Methods {
		if m.Name != "<init>" {
			continue
		}
		pcs, err := fieldPutPCs(cf, m.Method, declaring, ref.Name, ref.Desc)
		if err != nil || len(pcs) == 0 {
			continue
		}
		facts, err := AnalyzeWithParameterRefs(cf, declaring, m.Method)
		if err != nil {
			out.Unknown = append(out.Unknown, fmt.Sprintf("%s#<init>%s: local-value analysis: %v", declaring, m.Descriptor, err))
			continue
		}
		if aborted, why := facts.Aborted(); aborted {
			out.Unknown = append(out.Unknown, fmt.Sprintf("%s#<init>%s: local-value analysis aborted: %s", declaring, m.Descriptor, why))
			continue
		}
		ctor := MethodID{declaring, "<init>", m.Descriptor}
		for _, pc := range pcs {
			site := bytecodeir.Site{Method: declaring + "#" + m.Name, Descriptor: m.Descriptor, PC: pc, Op: bytecodeir.OpFieldPut}
			v, ok := facts.fieldWrites[site]
			if !ok {
				continue // not a self write
			}
			writes++
			idx.collect(v, callerSite{ci: ci, m: m}, -1, map[string]bool{}, &out, ctor)
		}
	}
	if writes == 0 {
		out.Unknown = append(out.Unknown, fmt.Sprintf("no `this.%s = ...` write in any constructor of %s", ref.Name, declaring))
	}
	sort.Strings(out.Unknown)
	out.Unknown = uniqueSorted(out.Unknown)
	return out
}
