package deserialize

import (
	"strings"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/relocate"
)

const (
	ccPkg  = "org.apache.commons.collections."
	cc4Pkg = "org.apache.commons.collections4."
	ccComp = "commons-collections:commons-collections"
)

// cc3 returns the CommonsCollections6 classes under prefix (the original
// package or a relocated one); guarded adds 3.2.2's InvokerTransformer
// readObject guard.
func cc3(prefix string, guarded bool) []*hierarchy.ClassInfo {
	var inv []hierarchy.Method
	if guarded {
		inv = append(inv, meth("readObject", readObj, priv,
			classfile.Call{Kind: classfile.CallStatic, Owner: prefix + "functors.FunctorUtils", Name: "checkUnsafeSerialization", Desc: "(Ljava/lang/Class;)V"}))
	}
	return []*hierarchy.ClassInfo{
		class(prefix+"keyvalue.TiedMapEntry", "java.lang.Object", pub, ser),
		class(prefix+"map.LazyMap", "java.lang.Object", pub, ser),
		class(prefix+"functors.ChainedTransformer", "java.lang.Object", pub, ser),
		class(prefix+"functors.ConstantTransformer", "java.lang.Object", pub, ser),
		class(prefix+"functors.InvokerTransformer", "java.lang.Object", pub, ser, inv...),
	}
}

func evalChains(t *testing.T, classes []*hierarchy.ClassInfo, owner func(*hierarchy.ClassInfo) ClassOwner, relocated []ClassOwner, targetJava int, sites []Site) map[string]ChainPresence {
	t.Helper()
	h := hierarchy.NewForTesting(classes)
	out := map[string]ChainPresence{}
	for _, p := range EvaluateChains(ChainInputs{H: h, Owner: owner, Relocated: relocated, TargetJava: targetJava}, sites) {
		out[p.Spec.ID] = p
	}
	if len(out) != len(ChainCatalog) {
		t.Fatalf("got %d chains, want the whole catalog (%d)", len(out), len(ChainCatalog))
	}
	return out
}

// ownedBy: classes whose name starts with a prefix belong to that
// component; others are unidentified.
func ownedBy(prefixes map[string]string) func(*hierarchy.ClassInfo) ClassOwner {
	return func(ci *hierarchy.ClassInfo) ClassOwner {
		for p, comp := range prefixes {
			if strings.HasPrefix(ci.FQCN, p) {
				return ClassOwner{Instance: comp, Component: comp, Version: "x"}
			}
		}
		return ClassOwner{Instance: "unidentified"}
	}
}

func javaSite(caller, filter string) Site {
	return Site{CallerFQCN: caller, CallerName: "read", CallerDesc: "()V", Format: FormatJava, Filter: filter}
}

func siteState(t *testing.T, p ChainPresence, caller string) ChainSite {
	t.Helper()
	for _, s := range p.Sites {
		if s.Site.CallerFQCN == caller {
			return s
		}
	}
	t.Fatalf("%s: no result for site %s", p.Spec.ID, caller)
	return ChainSite{}
}

// TestChainCandidateAndSitePolicies: the vulnerable 3.2.1 form, confirmed
// as commons-collections, is a candidate at a site with nothing seen,
// unresolved at a site whose policy is not evaluated, and blocked —
// at that site only — by an evaluated filter that rejects HashMap.
func TestChainCandidateAndSitePolicies(t *testing.T) {
	open := javaSite("app.Open", FilterNoneSeen)
	unknown := javaSite("app.Unknown", FilterUnknown)
	filtered := javaSite("app.Filtered", FilterEvaluated)
	pats, _ := parseJavaFilter("app.*;!*")
	filtered.allow = javaFilterPolicy(pats)
	xs := Site{CallerFQCN: "app.XS", Format: FormatXStream, Filter: FilterUnrestricted}
	got := evalChains(t, cc3(ccPkg, false), ownedBy(map[string]string{ccPkg: ccComp}), nil, 0, []Site{open, unknown, filtered, xs})
	p := got["java:commons-collections3:lazymap"]
	if p.State != ChainCandidate {
		t.Fatalf("chain state %s (blocks %v, unresolved %v), want %s", p.State, p.Blocks, p.Unresolved, ChainCandidate)
	}
	if len(p.Sites) != 3 {
		t.Fatalf("%d sites, want the 3 Java sites only (not XStream)", len(p.Sites))
	}
	if s := siteState(t, p, "app.Open"); s.State != ChainCandidate || s.Flow != FlowNotAnalyzed {
		t.Errorf("open site: %s flow %s", s.State, s.Flow)
	}
	if s := siteState(t, p, "app.Unknown"); s.State != ChainUnresolved {
		t.Errorf("unknown-policy site: %s, want %s", s.State, ChainUnresolved)
	}
	s := siteState(t, p, "app.Filtered")
	if s.State != ChainBlocked || !strings.Contains(s.SiteReason(), "java.util.HashMap") {
		t.Errorf("filtered site: %s (%s), want blocked by its filter rejecting HashMap", s.State, s.SiteReason())
	}
	// Other chains: their classes are not in the input.
	for _, id := range []string{"java:commons-collections4:transformingcomparator", "java:commons-beanutils:beancomparator", "java:groovy:convertedclosure"} {
		if got[id].State != ChainAbsent || len(got[id].Sites) != 0 || len(got[id].Missing) == 0 {
			t.Errorf("%s: %s with %d sites, missing %v; want classes_absent with none", id, got[id].State, len(got[id].Sites), got[id].Missing)
		}
	}
}

// TestChainGuardBlocks: the fixed release's bytecode guard blocks the
// chain at every site, whatever the site's policy; a readObject that
// does not call the guard does not.
func TestChainGuardBlocks(t *testing.T) {
	sites := []Site{javaSite("app.Open", FilterNoneSeen)}
	got := evalChains(t, cc3(ccPkg, true), ownedBy(map[string]string{ccPkg: ccComp}), nil, 0, sites)
	p := got["java:commons-collections3:lazymap"]
	if p.State != ChainBlocked || p.Sites[0].State != ChainBlocked || !strings.Contains(strings.Join(p.Blocks, ""), "enableUnsafeSerialization") {
		t.Fatalf("guarded 3.2.2 form: chain %s, site %s, blocks %v", p.State, p.Sites[0].State, p.Blocks)
	}
	// A readObject without the guard call.
	classes := cc3(ccPkg, false)
	classes[4] = class(ccPkg+"functors.InvokerTransformer", "java.lang.Object", pub, ser, meth("readObject", readObj, priv))
	if p := evalChains(t, classes, ownedBy(map[string]string{ccPkg: ccComp}), nil, 0, sites)["java:commons-collections3:lazymap"]; p.State != ChainCandidate {
		t.Errorf("readObject without the guard call: %s, want %s", p.State, ChainCandidate)
	}
	// collections4 4.1: InvokerTransformer is not Serializable.
	cc4 := func(serializable bool) []*hierarchy.ClassInfo {
		var invIfaces []string
		if serializable {
			invIfaces = ser
		}
		return []*hierarchy.ClassInfo{
			class(cc4Pkg+"comparators.TransformingComparator", "java.lang.Object", pub, ser),
			class(cc4Pkg+"functors.ChainedTransformer", "java.lang.Object", pub, ser),
			class(cc4Pkg+"functors.ConstantTransformer", "java.lang.Object", pub, ser),
			class(cc4Pkg+"functors.InvokerTransformer", "java.lang.Object", pub, invIfaces),
		}
	}
	own4 := ownedBy(map[string]string{cc4Pkg: "org.apache.commons:commons-collections4"})
	if p := evalChains(t, cc4(true), own4, nil, 0, sites)["java:commons-collections4:transformingcomparator"]; p.State != ChainCandidate {
		t.Errorf("collections4 4.0 form: %s (%v %v)", p.State, p.Blocks, p.Unresolved)
	}
	if p := evalChains(t, cc4(false), own4, nil, 0, sites)["java:commons-collections4:transformingcomparator"]; p.State != ChainBlocked {
		t.Errorf("collections4 4.1 form: %s, want blocked", p.State)
	}
}

// TestChainIdentity: a class with the right name in an unidentified
// component is not confirmed (unresolved, never a candidate); a copy
// under a relocation of a confirmed embedded instance is found and
// confirmed; a relocated copy whose instance is another component is not
// looked up.
func TestChainIdentity(t *testing.T) {
	sites := []Site{javaSite("app.Open", FilterNoneSeen)}
	p := evalChains(t, cc3(ccPkg, false), ownedBy(nil), nil, 0, sites)["java:commons-collections3:lazymap"]
	if p.State != ChainUnresolved || !strings.Contains(strings.Join(p.Unresolved, ""), "unidentified component") {
		t.Errorf("unidentified copy: %s %v", p.State, p.Unresolved)
	}
	for _, c := range p.Classes {
		if c.Confirmed {
			t.Errorf("%s confirmed without an identified owner", c.Found)
		}
	}

	shaded := "shaded.cc."
	rules := []relocate.Rule{{From: "org.apache.commons.collections", To: "shaded.cc"}}
	emb := ClassOwner{Instance: "emb", Component: ccComp, Version: "3.2.1", Relocations: rules}
	owner := func(ci *hierarchy.ClassInfo) ClassOwner {
		if strings.HasPrefix(ci.FQCN, shaded) {
			return emb
		}
		return ClassOwner{}
	}
	p = evalChains(t, cc3(shaded, false), owner, []ClassOwner{emb}, 0, sites)["java:commons-collections3:lazymap"]
	if p.State != ChainCandidate {
		t.Fatalf("relocated confirmed copy: %s (missing %v, unresolved %v)", p.State, p.Missing, p.Unresolved)
	}
	if p.Classes[4].Found != shaded+"functors.InvokerTransformer" || !p.Classes[4].Confirmed {
		t.Errorf("relocated class: %+v", p.Classes[4])
	}
	// The guard is read from the relocated class too.
	p = evalChains(t, cc3(shaded, true), owner, []ClassOwner{emb}, 0, sites)["java:commons-collections3:lazymap"]
	if p.State != ChainBlocked {
		t.Errorf("relocated guarded copy: %s, want blocked", p.State)
	}
	other := ClassOwner{Instance: "o", Component: "x:y", Relocations: rules}
	if p := evalChains(t, cc3(shaded, false), owner, []ClassOwner{other}, 0, sites)["java:commons-collections3:lazymap"]; p.State != ChainAbsent {
		t.Errorf("relocation of another component: %s, want absent", p.State)
	}
}

// TestChainTargetJDK: a chain that needs an old JDK is unresolved
// without -target-java, blocked above its limit, unresolved on the
// limit when the update release matters.
func TestChainTargetJDK(t *testing.T) {
	groovy := []*hierarchy.ClassInfo{
		class("groovy.lang.Closure", "java.lang.Object", pub|abstr, ser),
		class("org.codehaus.groovy.runtime.ConvertedClosure", "java.lang.Object", pub, ser),
		class("org.codehaus.groovy.runtime.MethodClosure", "groovy.lang.Closure", pub, nil),
	}
	own := ownedBy(map[string]string{"org.codehaus.groovy.": "org.codehaus.groovy:groovy-all"})
	sites := []Site{javaSite("app.Open", FilterNoneSeen)}
	for _, c := range []struct {
		java int
		want string
	}{{0, ChainUnresolved}, {8, ChainUnresolved}, {7, ChainCandidate}, {11, ChainBlocked}} {
		if p := evalChains(t, groovy, own, nil, c.java, sites)["java:groovy:convertedclosure"]; p.State != c.want {
			t.Errorf("target java %d: %s (%v %v), want %s", c.java, p.State, p.Blocks, p.Unresolved, c.want)
		}
	}
	// 2.4.4: MethodClosure declares readResolve.
	groovy[2] = class("org.codehaus.groovy.runtime.MethodClosure", "groovy.lang.Closure", pub, nil, meth("readResolve", "()Ljava/lang/Object;", priv))
	if p := evalChains(t, groovy, own, nil, 7, sites)["java:groovy:convertedclosure"]; p.State != ChainBlocked {
		t.Errorf("groovy 2.4.4 form: %s, want blocked", p.State)
	}
}

// TestChainXStream: the EventHandler chain needs only JDK classes; its
// guard is a field of XStream's ReflectionConverter (1.4.7+); without
// the class the guard is not checked; a site whose policy rejects
// EventHandler blocks it there.
func TestChainXStream(t *testing.T) {
	rc := "com.thoughtworks.xstream.converters.reflection.ReflectionConverter"
	open := Site{CallerFQCN: "app.Open", Format: FormatXStream, Filter: FilterUnrestricted}
	allowlist := Site{CallerFQCN: "app.Allow", Format: FormatXStream, Filter: FilterDefaultAllowlist}
	allowlist.allow = func(string) bool { return false }
	sites := []Site{open, allowlist, javaSite("app.Java", FilterNoneSeen)}
	old := []*hierarchy.ClassInfo{class(rc, "java.lang.Object", pub, nil)}
	p := evalChains(t, old, nil, nil, 0, sites)["xstream:eventhandler"]
	if p.State != ChainCandidate || len(p.Sites) != 2 {
		t.Fatalf("1.4.6 form: %s with %d sites", p.State, len(p.Sites))
	}
	if s := siteState(t, p, "app.Allow"); s.State != ChainBlocked || !strings.Contains(s.SiteReason(), "java.beans.EventHandler") {
		t.Errorf("default-allowlist site: %s %s", s.State, s.SiteReason())
	}
	fixed := []*hierarchy.ClassInfo{{FQCN: rc, SuperClass: "java.lang.Object", AccessFlags: pub, Fields: []classfile.Field{{Name: "eventHandlerType", Descriptor: "Ljava/lang/Class;", AccessFlags: priv | static}}}}
	if p := evalChains(t, fixed, nil, nil, 0, sites)["xstream:eventhandler"]; p.State != ChainBlocked {
		t.Errorf("1.4.7 form: %s, want blocked", p.State)
	}
	if p := evalChains(t, nil, nil, nil, 0, sites)["xstream:eventhandler"]; p.State != ChainUnresolved || p.Sites[0].State != ChainUnresolved {
		t.Errorf("XStream classes absent: %s, want unresolved (guard not checked)", p.State)
	}
	if p := evalChains(t, old, nil, nil, 0, nil)["xstream:eventhandler"]; p.State != ChainNoEntry || len(p.Sites) != 0 {
		t.Errorf("no XStream entry: %s, want %s", p.State, ChainNoEntry)
	}
}

// TestChainCatalogWellFormed: every chain names a known format, required
// or JDK classes, a source, and guards of a known kind.
func TestChainCatalogWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range ChainCatalog {
		if seen[c.ID] {
			t.Errorf("duplicate chain %s", c.ID)
		}
		seen[c.ID] = true
		if c.Format != FormatJava && c.Format != FormatXStream {
			t.Errorf("%s: format %q", c.ID, c.Format)
		}
		if len(c.Required)+len(c.JDKClasses) == 0 || c.Source == "" || c.Trigger == "" || c.Sink == "" {
			t.Errorf("%s: incomplete", c.ID)
		}
		if len(c.Required) > 0 && len(c.Components) == 0 {
			t.Errorf("%s: required classes without a component", c.ID)
		}
		for _, g := range c.Guards {
			switch g.Kind {
			case "declares-method", "not-serializable", "declares-field":
			default:
				t.Errorf("%s: guard kind %q", c.ID, g.Kind)
			}
		}
	}
}

// TestChainTwoInstances: the chain's classes under their original names
// and a relocated copy in a confirmed embedded instance are two results;
// the shaded copy's guard blocks it alone, the plain copy stays a
// candidate (and the other way round).
func TestChainTwoInstances(t *testing.T) {
	shaded := "shaded.cc."
	emb := ClassOwner{Instance: "emb", Component: ccComp, Version: "3.2.2",
		Relocations: []relocate.Rule{{From: "org.apache.commons.collections", To: "shaded.cc"}}}
	owner := func(ci *hierarchy.ClassInfo) ClassOwner {
		if strings.HasPrefix(ci.FQCN, shaded) {
			return emb
		}
		return ClassOwner{Instance: "plain", Component: ccComp, Version: "3.2.1"}
	}
	sites := []Site{javaSite("app.Open", FilterNoneSeen)}
	for _, c := range []struct {
		plainGuard, shadedGuard bool
		want                    map[string]string
	}{
		{false, true, map[string]string{"": ChainCandidate, "emb": ChainBlocked}},
		{true, false, map[string]string{"": ChainBlocked, "emb": ChainCandidate}},
	} {
		classes := append(cc3(ccPkg, c.plainGuard), cc3(shaded, c.shadedGuard)...)
		h := hierarchy.NewForTesting(classes)
		got := map[string]string{}
		for _, p := range EvaluateChains(ChainInputs{H: h, Owner: owner, Relocated: []ClassOwner{emb}}, sites) {
			if p.Spec.ID == "java:commons-collections3:lazymap" {
				got[p.Copy] = p.State
				wantInst := map[string]string{"": "plain", "emb": "emb"}[p.Copy]
				if p.Instance != wantInst || len(p.Sites) != 1 {
					t.Errorf("copy %q: instance %q, %d sites", p.Copy, p.Instance, len(p.Sites))
				}
			}
		}
		if len(got) != 2 || got[""] != c.want[""] || got["emb"] != c.want["emb"] {
			t.Errorf("guards plain=%v shaded=%v: got %v, want %v", c.plainGuard, c.shadedGuard, got, c.want)
		}
	}
	// Only the shaded copy: one result, no absent original.
	h := hierarchy.NewForTesting(cc3(shaded, false))
	n := 0
	for _, p := range EvaluateChains(ChainInputs{H: h, Owner: owner, Relocated: []ClassOwner{emb}}, sites) {
		if p.Spec.ID == "java:commons-collections3:lazymap" {
			n++
			if p.Copy != "emb" || p.State != ChainCandidate {
				t.Errorf("shaded only: copy %q state %s", p.Copy, p.State)
			}
		}
	}
	if n != 1 {
		t.Errorf("shaded only: %d results, want 1", n)
	}
}
