package springaop

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const fixture = "../../testdata/"

// fixtureHierarchy reads testdata/aop (scripts/aop-oracle.sh) over the
// Spring 5.3.6 jars of testdata/tx.
func fixtureHierarchy(t *testing.T) *hierarchy.Hierarchy {
	t.Helper()
	deps, err := filepath.Glob(fixture + "tx/deps/*.jar")
	if err != nil || len(deps) == 0 {
		t.Fatalf("no deps: %v", err)
	}
	var dj []input.DepJar
	for _, d := range deps {
		dj = append(dj, input.DepJar{Path: d})
	}
	h, _, err := hierarchy.Build(fixture+"aop/app.jar", dj)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// The fixture's bean types: the stereotypes @ComponentScan finds.
var fixtureBeans = []string{
	"fx.app.AopConfig", "fx.aop.svc.OrderService", "fx.aop.svc.JdbcRepo", "fx.aop.svc.ReportService", "fx.aop.aspect.TracingAspect",
}

func TestFindOnFixture(t *testing.T) {
	h := fixtureHierarchy(t)
	r := Find(h, fixtureBeans)
	if r.NoCreator || r.Truncated {
		t.Fatalf("NoCreator %v Truncated %v", r.NoCreator, r.Truncated)
	}
	if got := strings.Join(r.Aspects, ","); got != "fx.aop.aspect.TracingAspect,fx.aop.aspect.Unregistered" {
		t.Errorf("aspects %s", got)
	}
	if r.RegisteredAspects() != 1 {
		t.Errorf("registered aspects %d, want 1", r.RegisteredAspects())
	}

	// Advice → advised join points (bean#method).
	matched := map[string][]string{}
	for _, m := range matchesOf(h, r, fixtureBeans) {
		matched[m[0]] = append(matched[m[0]], m[1])
	}
	want := map[string]string{
		"before":    "fx.aop.svc.OrderService#cancel,fx.aop.svc.OrderService#internal,fx.aop.svc.OrderService#place,fx.aop.svc.ReportService#render",
		"afterSave": "fx.aop.svc.JdbcRepo#save",
		"onError":   "fx.aop.svc.JdbcRepo#find,fx.aop.svc.JdbcRepo#save",
		"audit":     "fx.aop.svc.OrderService#cancel",
		"monitor":   "fx.aop.svc.ReportService#render",
		"ghost":     "",
		"withArgs":  "",
		"never":     "",
	}
	for adv, w := range want {
		got := matched[adv]
		sort.Strings(got)
		if strings.Join(got, ",") != w {
			t.Errorf("%s advises %v, want %s", adv, got, w)
		}
	}
	for _, a := range r.Advice {
		switch a.Name {
		case "withArgs":
			if a.Undecided != 2 || len(a.Unknown) != 1 || a.Unknown[0] != "args(id)" {
				t.Errorf("withArgs: undecided %d, unknown %v; want 2 (Repo's save and find), [args(id)]", a.Undecided, a.Unknown)
			}
		case "never":
			if a.Registered {
				t.Error("Unregistered#never: the aspect is no bean")
			}
		default:
			if a.Undecided != 0 || len(a.Unknown) != 0 {
				t.Errorf("%s: undecided %d, unknown %v", a.Name, a.Undecided, a.Unknown)
			}
		}
	}

	// Edges: from the call sites, with conditions; none from the
	// self-invocation in place().
	edges := map[string]string{}
	for _, e := range r.Edges {
		edges[e.CallerFQCN+"#"+e.CallerName+" -> "+e.TargetName] = e.Condition
	}
	for k, cond := range map[string]string{
		"fx.app.Main#main -> before":    "",
		"fx.app.Main#main -> afterSave": "returned",
		"fx.app.Main#main -> onError":   "exceptional",
		"fx.app.Main#main -> audit":     "",
		"fx.app.Main#main -> monitor":   "",
	} {
		got, ok := edges[k]
		if !ok || got != cond {
			t.Errorf("edge %s: present %v condition %q, want %q", k, ok, got, cond)
		}
	}
	for k := range edges {
		if strings.HasPrefix(k, "fx.aop.svc.OrderService#place") {
			t.Errorf("edge %s: a self-invocation bypasses the proxy", k)
		}
		if strings.HasSuffix(k, "-> ghost") || strings.HasSuffix(k, "-> withArgs") || strings.HasSuffix(k, "-> never") {
			t.Errorf("edge %s: the advice matches nothing decided", k)
		}
	}
	if r.SelfCalls != 1 {
		t.Errorf("self calls %d, want 1 (place → internal)", r.SelfCalls)
	}
}

// matchesOf re-evaluates the registered advice to list advice → join
// point pairs.
func matchesOf(h *hierarchy.Hierarchy, r Result, beans []string) [][2]string {
	aspects := map[string]bool{}
	for _, a := range r.Aspects {
		aspects[a] = true
	}
	var out [][2]string
	for _, a := range r.Advice {
		if !a.Registered {
			continue
		}
		res := newResolver(h, a.Aspect)
		n := bind(h, parse(a.Expression), a.Desc)
		for _, c := range beanClasses(h, beans, aspects) {
			for _, jp := range joinPoints(h, c) {
				jp := jp
				if n.eval(&jp, res) == yes {
					out = append(out, [2]string{a.Name, jp.C + "#" + jp.m.Name})
				}
			}
		}
	}
	return out
}

func TestParseExecution(t *testing.T) {
	for _, tc := range []struct {
		expr                  string
		ok                    bool
		ret, decl, name, mods string
		params                string
	}{
		{"execution(* *(..))", true, "*", "", "*", "", ".."},
		{"execution(public * com.x..*Service.find*(..))", true, "*", "com.x..*Service", "find*", "public", ".."},
		{"execution(* com.x..*(..))", true, "*", "com.x..*", "*", "", ".."},
		{"execution(void com.x.A.set(String, int))", true, "void", "com.x.A", "set", "", "String,int"},
		{"execution(* com.x.A+.m()) ", true, "*", "com.x.A+", "m", "", ""},
		{"execution(* com.x.A.m(..) throws java.io.IOException)", true, "*", "com.x.A", "m", "", ".."},
		{"execution(@com.x.Ann * *(..))", false, "", "", "", "", ""},
		{"execution(*)", false, "", "", "", "", ""},
	} {
		n := parse(tc.expr)
		e, ok := n.(execNode)
		if ok != tc.ok {
			t.Errorf("%s: parsed %#v, want ok=%v", tc.expr, n, tc.ok)
			continue
		}
		if !ok {
			continue
		}
		if e.ret != tc.ret || e.decl != tc.decl || e.name != tc.name || strings.Join(e.mods, " ") != tc.mods || strings.Join(e.params, ",") != tc.params {
			t.Errorf("%s: %+v", tc.expr, e)
		}
	}
}

func TestParseOperatorsAndUnknown(t *testing.T) {
	n := parse("services() && !within(com.x.Skip) || @annotation(com.x.Ann)")
	o, ok := n.(orNode)
	if !ok {
		t.Fatalf("%#v", n)
	}
	a, ok := o.l.(andNode)
	if !ok {
		t.Fatalf("%#v", o.l)
	}
	if _, ok := a.l.(refNode); !ok {
		t.Errorf("left %#v", a.l)
	}
	if _, ok := a.r.(notNode); !ok {
		t.Errorf("right %#v", a.r)
	}
	if x, ok := o.r.(annNode); !ok || x.ann != "com.x.Ann" || x.onType {
		t.Errorf("annotation %#v", o.r)
	}
	for _, e := range []string{"bean(orderService)", "args(java.lang.String)", "this(com.x.A)", "execution(* *(..)", "", "within(com.x.A) &&"} {
		why := map[string]bool{}
		unknownReasons(parse(e), &resolver{cache: map[string]node{}, active: map[string]bool{}}, why, 0)
		if len(why) == 0 {
			t.Errorf("%q: no unknown reason", e)
		}
	}
}

func TestTriLogic(t *testing.T) {
	if and3(no, unknown) != no || and3(yes, unknown) != unknown || or3(yes, unknown) != yes || or3(no, unknown) != unknown || not3(unknown) != unknown {
		t.Error("three-valued logic")
	}
}

func TestTypeAndParamPatterns(t *testing.T) {
	h := fixtureHierarchy(t)
	for _, tc := range []struct {
		pat, typ string
		want     tri
	}{
		{"*", "int", yes},
		{"*", "fx.aop.svc.Repo", yes},
		{"String", "java.lang.String", yes},
		{"String", "com.x.String", no},
		{"fx..*", "fx.aop.svc.Repo", yes},
		{"fx.*", "fx.aop.svc.Repo", no},
		{"fx.aop.svc.*Service", "fx.aop.svc.OrderService", yes},
		{"fx.aop.svc.Repo+", "fx.aop.svc.JdbcRepo", yes},
		{"fx.aop.svc.Repo", "fx.aop.svc.JdbcRepo", no},
		{"int", "long", no},
		{"Object[]", "java.lang.Object[]", yes},
		{"Object", "java.lang.Object[]", no},
	} {
		if got := typeMatches(h, tc.pat, tc.typ); got != tc.want {
			t.Errorf("typeMatches(%s, %s) = %v, want %v", tc.pat, tc.typ, got, tc.want)
		}
	}
	for _, tc := range []struct {
		pats, types string
		want        tri
	}{
		{"..", "", yes},
		{"..", "int,long", yes},
		{"", "int", no},
		{"*", "int", yes},
		{"*", "int,int", no},
		{"..,long", "int,long", yes},
		{"String,..", "java.lang.String", yes},
		{"String,..", "int", no},
	} {
		split := func(s string) []string {
			if s == "" {
				return nil
			}
			return strings.Split(s, ",")
		}
		if got := paramsMatch(h, split(tc.pats), split(tc.types)); got != tc.want {
			t.Errorf("paramsMatch(%s; %s) = %v, want %v", tc.pats, tc.types, got, tc.want)
		}
	}
}

func TestDescParts(t *testing.T) {
	p, r := descParts("(I[Ljava/lang/String;J[[D)Lfx/A$B;")
	if strings.Join(p, ",") != "int,java.lang.String[],long,double[][]" || r != "fx.A$B" {
		t.Errorf("%v %s", p, r)
	}
}

// No advice objects ever created (no spring-aop): advice is cataloged,
// no edges.
func TestNoCreatorWithoutSpringAOP(t *testing.T) {
	h, _, err := hierarchy.Build(fixture+"aop/app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	r := Find(h, fixtureBeans)
	if !r.NoCreator || len(r.Edges) != 0 || len(r.Advice) == 0 {
		t.Errorf("NoCreator %v, %d edges, %d advice", r.NoCreator, len(r.Edges), len(r.Advice))
	}
}
