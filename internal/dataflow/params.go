package dataflow

import (
	"fmt"
	"sort"
	"strings"
	"sync"

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
	m   *hierarchy.Method // into ci.Methods: a copy per call site cost ~300 MB on petclinic
	pc  int
	via classfile.CallKind
	// refKind is, for an invokedynamic creation site (creators), the
	// implementation handle's reference kind: whether the first captured
	// value is the implementation's receiver.
	refKind byte
}

// CallerOptions tune a CallerIndex; a zero field keeps the package
// default (MaxParameterDepth, MaxParameterCallers, captures not followed).
type CallerOptions struct {
	MaxDepth, MaxCallers int
	// FollowCaptures resolves a parameter of a lambda/method-reference
	// implementation that holds a captured value through the
	// invokedynamic sites creating it (the value the creator captured),
	// instead of the "invokedynamic callers are not traced" unknown. A
	// parameter that is an argument of the functional-interface call
	// itself stays unknown, with that reason (DYNAMIC_LINKAGE L1).
	FollowCaptures bool
	// VirtualCallers accepts a virtual method whose call sites are not
	// statically exact: its callers are the union of the call sites
	// naming it and naming any declaration it overrides in the input
	// (over-approximating — a site through the supertype may dispatch
	// elsewhere), instead of the "overridden"/"overrides" unknown. A
	// supertype outside the input still makes the set unknown unless it
	// cannot declare the method: a JDK type and a descriptor naming a
	// type outside the JDK, or a marker interface with no methods.
	VirtualCallers bool
	// Direct stops at the immediate callers (and creators): an argument
	// that is itself a formal parameter of the caller is returned as is,
	// with the caller in ParameterValues.In, for the consumer to resolve
	// (and memoize) one level up — instead of being expanded here, anew
	// for every chain through it.
	Direct bool
	// MaxFacts bounds the cached method facts (oldest evicted first, then
	// recomputed on demand); 0 keeps every method's facts.
	MaxFacts int
}

// CallerIndex records, for every method in the analyzed input, the
// call sites that reach it with a statically exact target, plus methods
// whose callers cannot all be seen (referenced by invokedynamic as a
// lambda/method-reference implementation). It is built once per run.
type CallerIndex struct {
	h       *hierarchy.Hierarchy
	callers map[MethodID][]callerSite
	hidden  map[MethodID]bool
	// creators: per lambda/method-reference implementation, the
	// invokedynamic sites creating a functional object for it.
	creators map[MethodID][]callerSite
	opts     CallerOptions
	facts    map[MethodID]*MethodFacts
	failed   map[MethodID]string
	reader   classfile.JarClassReader
	// classes caches the last classCacheSize parsed class files: the
	// analysis of one class's methods, and caller chains, re-read the
	// same classes, and parsing dominates the cost.
	classes    map[string]*classfile.ClassFile
	classOrder []string
	factOrder  []MethodID
}

const classCacheSize = 512

func (idx *CallerIndex) storeFacts(id MethodID, f *MethodFacts) {
	if _, ok := idx.facts[id]; !ok {
		idx.factOrder = append(idx.factOrder, id)
	}
	idx.facts[id] = f
	for idx.opts.MaxFacts > 0 && len(idx.factOrder) > idx.opts.MaxFacts {
		delete(idx.facts, idx.factOrder[0])
		idx.factOrder = idx.factOrder[1:]
	}
}

// Forget drops the cached facts of methods (they are recomputed if asked
// for again).
func (idx *CallerIndex) Forget(methods []MethodID) {
	drop := map[MethodID]bool{}
	for _, m := range methods {
		if _, ok := idx.facts[m]; ok {
			delete(idx.facts, m)
			drop[m] = true
		}
	}
	if len(drop) == 0 {
		return
	}
	kept := idx.factOrder[:0]
	for _, m := range idx.factOrder {
		if !drop[m] {
			kept = append(kept, m)
		}
	}
	idx.factOrder = kept
}

func (idx *CallerIndex) readClass(ci *hierarchy.ClassInfo) (*classfile.ClassFile, error) {
	if cf, ok := idx.classes[ci.FQCN]; ok {
		return cf, nil
	}
	cf, err := idx.reader.ReadClass(ci.SourceJar, ci.SourceEntry)
	if err != nil {
		return nil, err
	}
	if idx.classes == nil {
		idx.classes = map[string]*classfile.ClassFile{}
	}
	if len(idx.classOrder) >= classCacheSize {
		delete(idx.classes, idx.classOrder[0])
		idx.classOrder = idx.classOrder[1:]
	}
	idx.classes[ci.FQCN] = cf
	idx.classOrder = append(idx.classOrder, ci.FQCN)
	return cf, nil
}

// NewCallerIndex scans every method's call list. Call targets are keyed
// by the class that actually declares them (hierarchy.ResolveExact), so a
// call naming a subclass still reaches the inherited declaration.
func NewCallerIndex(h *hierarchy.Hierarchy) *CallerIndex {
	idx := &CallerIndex{
		h: h, callers: map[MethodID][]callerSite{}, hidden: map[MethodID]bool{}, creators: map[MethodID][]callerSite{},
		facts: map[MethodID]*MethodFacts{}, failed: map[MethodID]string{},
	}
	for _, fqcn := range h.SortedFQCNs() {
		ci, _ := h.Lookup(fqcn)
		for mi := range ci.Methods {
			m := &ci.Methods[mi]
			for _, c := range m.Calls {
				decl, ok := h.ResolveExact(c.Owner, c.Name, c.Desc)
				if !ok {
					continue
				}
				target := MethodID{decl, c.Name, c.Desc}
				if c.Kind == classfile.CallDynamic {
					idx.hidden[target] = true
					idx.creators[target] = append(idx.creators[target], callerSite{ci: ci, m: m, pc: c.PC, via: c.Kind, refKind: c.RefKind})
					continue
				}
				idx.callers[target] = append(idx.callers[target], callerSite{ci: ci, m: m, pc: c.PC, via: c.Kind})
			}
		}
	}
	return idx
}

// SetOptions replaces the index's limits and options (CallerOptions).
func (idx *CallerIndex) SetOptions(o CallerOptions) { idx.opts = o }

func (idx *CallerIndex) maxDepth() int {
	if idx.opts.MaxDepth > 0 {
		return idx.opts.MaxDepth
	}
	return MaxParameterDepth
}

func (idx *CallerIndex) maxCallers() int {
	if idx.opts.MaxCallers > 0 {
		return idx.opts.MaxCallers
	}
	return MaxParameterCallers
}

// Facts returns the local-value facts of method (with parameter
// references: a formal parameter is a KindTyped value carrying its
// index), analyzed once and cached, or why they are not available.
func (idx *CallerIndex) Facts(method MethodID) (*MethodFacts, string) {
	ci, ok := idx.h.Lookup(method.Owner)
	if !ok {
		return nil, fmt.Sprintf("%s is outside the analyzed hierarchy", method)
	}
	m, ok := ci.MethodByNameAndDesc(method.Name, method.Desc)
	if !ok {
		return nil, fmt.Sprintf("%s is not declared in the analyzed hierarchy", method)
	}
	return idx.methodFacts(ci, m)
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
	// In is, with CallerOptions.Direct, the method each of Values was
	// read in (parallel to Values); nil otherwise.
	In []MethodID
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
	if depth >= idx.maxDepth() {
		out.Unknown = append(out.Unknown, fmt.Sprintf("parameter chain deeper than %d caller levels at %s", idx.maxDepth(), method))
		return
	}
	sites := idx.callers[method]
	if idx.opts.FollowCaptures && idx.hidden[method] {
		visiting[key] = true
		idx.resolveCaptured(method, index, depth, visiting, out)
		delete(visiting, key)
		if len(sites) == 0 {
			return
		}
		// Also called directly (a method reference to an ordinary
		// method): those callers count too, when they are exact.
		if idx.opts.VirtualCallers {
			more, reason := idx.virtualCallers(method)
			if reason != "" {
				out.Unknown = append(out.Unknown, reason)
				return
			}
			sites = append(append([]callerSite(nil), sites...), more...)
		} else if reason := idx.closedCallersIgnoringHidden(method); reason != "" {
			out.Unknown = append(out.Unknown, reason)
			return
		}
	} else if idx.opts.VirtualCallers {
		more, reason := idx.virtualCallers(method)
		if reason != "" {
			out.Unknown = append(out.Unknown, reason)
			return
		}
		sites = append(append([]callerSite(nil), sites...), more...)
	} else if reason := idx.closedCallers(method); reason != "" {
		out.Unknown = append(out.Unknown, reason)
		return
	}
	if len(sites) == 0 {
		out.Unknown = append(out.Unknown, fmt.Sprintf("%s has no caller in the analyzed input (it may be called by a framework, reflection or external code)", method))
		return
	}
	if len(sites) > idx.maxCallers() {
		out.Unknown = append(out.Unknown, fmt.Sprintf("%s has %d callers, above the limit of %d", method, len(sites), idx.maxCallers()))
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

// resolveCaptured resolves parameter index of a lambda/method-reference
// implementation through the invokedynamic sites creating it: the
// creator's captured values come first in the implementation's
// parameters (after the receiver, when the handle has one); a parameter
// past them is an argument of the functional-interface call.
func (idx *CallerIndex) resolveCaptured(method MethodID, index, depth int, visiting map[string]bool, out *ParameterValues) {
	sites := idx.creators[method]
	if len(sites) > idx.maxCallers() {
		out.Unknown = append(out.Unknown, fmt.Sprintf("%s is created at %d invokedynamic sites, above the limit of %d", method, len(sites), idx.maxCallers()))
		return
	}
	for _, s := range sites {
		out.Callers++
		facts, reason := idx.methodFacts(s.ci, *s.m)
		if reason != "" {
			out.Unknown = append(out.Unknown, fmt.Sprintf("creator %s#%s%s: %s", s.ci.FQCN, s.m.Name, s.m.Descriptor, reason))
			continue
		}
		site := bytecodeir.Site{Method: s.ci.FQCN + "#" + s.m.Name, Descriptor: s.m.Descriptor, PC: s.pc, Op: bytecodeir.OpCall}
		captured, ok := facts.calls[site]
		if !ok {
			out.Unknown = append(out.Unknown, fmt.Sprintf("creator %s#%s%s: no invokedynamic recorded at pc=%d", s.ci.FQCN, s.m.Name, s.m.Descriptor, s.pc))
			continue
		}
		j := index
		switch s.refKind {
		case classfile.RefInvokeVirtual, classfile.RefInvokeSpecial, classfile.RefInvokeInterface:
			j++ // captured[0] is the implementation's receiver
		}
		if j >= len(captured) {
			out.Unknown = append(out.Unknown, fmt.Sprintf("parameter %d of %s is an argument of the functional-interface call, not a captured value", index, method))
			continue
		}
		idx.collect(captured[j], s, depth, visiting, out)
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
	case v.Kind == valueflow.KindTyped && v.Parameter != nil && !idx.opts.Direct:
		m := MethodID{s.ci.FQCN, s.m.Name, s.m.Descriptor}
		if len(owner) > 0 {
			m = owner[0]
		}
		idx.resolve(m, v.Parameter.Index, depth+1, visiting, out)
	case v.Kind == valueflow.KindUnknown:
		out.Unknown = append(out.Unknown, fmt.Sprintf("caller %s#%s%s: %s", s.ci.FQCN, s.m.Name, s.m.Descriptor, v.Reason))
	default:
		var in MethodID
		if idx.opts.Direct {
			in = MethodID{s.ci.FQCN, s.m.Name, s.m.Descriptor}
			if len(owner) > 0 {
				in = owner[0]
			}
		}
		for i, existing := range out.Values {
			if existing.Equal(v) && (!idx.opts.Direct || out.In[i] == in) {
				return
			}
		}
		out.Values = append(out.Values, v)
		if idx.opts.Direct {
			out.In = append(out.In, in)
		}
	}
}

// closedCallers returns "" when every caller of method is a statically
// exact call site, or why not.
func (idx *CallerIndex) closedCallers(method MethodID) string {
	if idx.hidden[method] {
		return fmt.Sprintf("%s is used as a lambda/method-reference implementation; its invokedynamic callers are not traced", method)
	}
	return idx.closedCallersIgnoringHidden(method)
}

func (idx *CallerIndex) closedCallersIgnoringHidden(method MethodID) string {
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

// virtualCallers returns, for CallerOptions.VirtualCallers, the call
// sites naming a declaration method overrides in the input (transitively),
// or why the callers cannot all be seen.
func (idx *CallerIndex) virtualCallers(method MethodID) ([]callerSite, string) {
	ci, ok := idx.h.Lookup(method.Owner)
	if !ok {
		return nil, fmt.Sprintf("%s is outside the analyzed hierarchy", method)
	}
	m, ok := ci.MethodByNameAndDesc(method.Name, method.Desc)
	if !ok {
		return nil, fmt.Sprintf("%s is not declared in the analyzed hierarchy", method)
	}
	if m.IsStatic() || m.AccessFlags&accPrivate != 0 || method.Name == "<init>" {
		return nil, ""
	}
	var out []callerSite
	seen := map[string]bool{method.Owner: true}
	queue := []*hierarchy.ClassInfo{ci}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		supers := append([]string{}, cur.Interfaces...)
		if cur.SuperClass != "" {
			supers = append(supers, cur.SuperClass)
		}
		for _, s := range supers {
			if seen[s] || s == "java.lang.Object" {
				continue
			}
			seen[s] = true
			sci, ok := idx.h.Lookup(s)
			if !ok {
				if mayDeclareOutside(s, method.Desc) {
					return nil, fmt.Sprintf("%s may override a method of %s, which is outside the analyzed input; calls through it are not seen", method, s)
				}
				continue
			}
			if _, declared := sci.MethodByNameAndDesc(method.Name, method.Desc); declared {
				out = append(out, idx.callers[MethodID{s, method.Name, method.Desc}]...)
			}
			queue = append(queue, sci)
		}
	}
	return out, ""
}

// markerInterfaces declare no methods.
var markerInterfaces = map[string]bool{"java.io.Serializable": true, "java.lang.Cloneable": true, "java.util.RandomAccess": true, "java.util.EventListener": true}

// mayDeclareOutside reports whether a supertype outside the input may
// declare a method with desc: not a marker interface, and not a JDK type
// whose methods could not mention a type outside the JDK that desc names.
func mayDeclareOutside(super, desc string) bool {
	if markerInterfaces[super] {
		return false
	}
	if !strings.HasPrefix(super, "java.") && !strings.HasPrefix(super, "jdk.") {
		return true
	}
	for i := 0; i < len(desc); i++ {
		if desc[i] != 'L' {
			continue
		}
		end := strings.IndexByte(desc[i:], ';')
		if end < 0 {
			break
		}
		t := desc[i+1 : i+end]
		if !strings.HasPrefix(t, "java/") && !strings.HasPrefix(t, "javax/") && !strings.HasPrefix(t, "jdk/") && !strings.HasPrefix(t, "sun/") {
			return false
		}
		i += end
	}
	return true
}

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
	facts, reason := idx.methodFacts(s.ci, *s.m)
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
	cf, err := idx.readClass(ci)
	if err != nil {
		idx.failed[id] = fmt.Sprintf("re-reading the class failed: %v", err)
		return nil, idx.failed[id]
	}
	f, err := AnalyzeWithParameterRefs(cf, ci.FQCN, m.Method)
	if err != nil {
		idx.failed[id] = fmt.Sprintf("local-value analysis failed: %v", err)
		return nil, idx.failed[id]
	}
	idx.storeFacts(id, f)
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
	cf, err := idx.readClass(ci)
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
			idx.collect(v, callerSite{ci: ci, m: &m}, -1, map[string]bool{}, &out, ctor)
		}
	}
	if writes == 0 {
		out.Unknown = append(out.Unknown, fmt.Sprintf("no `this.%s = ...` write in any constructor of %s", ref.Name, declaring))
	}
	sort.Strings(out.Unknown)
	out.Unknown = uniqueSorted(out.Unknown)
	return out
}

// Prefetch analyzes methods with workers goroutines (each with its own
// jar reader) and caches their facts, as Facts would one at a time —
// the facts of a method do not depend on any other's, so only the order
// of the work changes, not a result.
func (idx *CallerIndex) Prefetch(methods []MethodID, workers int) {
	byClass := map[string][]MethodID{}
	var classes []string
	for _, m := range methods {
		if _, done := idx.facts[m]; done {
			continue
		}
		if _, done := idx.failed[m]; done {
			continue
		}
		if _, ok := byClass[m.Owner]; !ok {
			classes = append(classes, m.Owner)
		}
		byClass[m.Owner] = append(byClass[m.Owner], m)
	}
	if workers < 1 {
		workers = 1
	}
	type result struct {
		id     MethodID
		facts  *MethodFacts
		reason string
	}
	jobs := make(chan string)
	results := make(chan []result)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var reader classfile.JarClassReader
			defer reader.Close()
			for fqcn := range jobs {
				var out []result
				ci, ok := idx.h.Lookup(fqcn)
				if !ok {
					results <- nil
					continue
				}
				cf, err := reader.ReadClass(ci.SourceJar, ci.SourceEntry)
				for _, id := range byClass[fqcn] {
					r := result{id: id}
					m, found := ci.MethodByNameAndDesc(id.Name, id.Desc)
					switch {
					case !found:
						continue // Facts reports it
					case err != nil:
						r.reason = fmt.Sprintf("re-reading the class failed: %v", err)
					default:
						f, aerr := AnalyzeWithParameterRefs(cf, ci.FQCN, m.Method)
						if aerr != nil {
							r.reason = fmt.Sprintf("local-value analysis failed: %v", aerr)
						} else {
							r.facts = f
						}
					}
					out = append(out, r)
				}
				results <- out
			}
		}()
	}
	go func() {
		for _, c := range classes {
			jobs <- c
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()
	for rs := range results {
		for _, r := range rs {
			if r.facts != nil {
				idx.storeFacts(r.id, r.facts)
			} else {
				idx.failed[r.id] = r.reason
			}
		}
	}
}
