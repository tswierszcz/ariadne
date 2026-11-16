package springaop

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
	"ariadne/internal/springxml"
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

// The fixture's beans: the stereotypes @ComponentScan finds and
// fx/aop/aop.xml's bean, with their names, and aop.xml's <aop:config>.
func fixtureInput(t *testing.T) Input {
	t.Helper()
	f, err := os.Open(fixture + "aop/app/fx/aop/aop.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	x, err := springxml.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	return Input{
		Beans: []string{"fx.app.AopConfig", "fx.aop.svc.OrderService", "fx.aop.svc.JdbcRepo", "fx.aop.svc.ReportService", "fx.aop.aspect.TracingAspect", "fx.aop.aspect.XmlAudit"},
		Names: map[string][]string{
			"fx.app.AopConfig": {"aopConfig"}, "fx.aop.svc.OrderService": {"orderService"}, "fx.aop.svc.JdbcRepo": {"jdbcRepo"},
			"fx.aop.svc.ReportService": {"reportService"}, "fx.aop.aspect.TracingAspect": {"tracingAspect"}, "fx.aop.aspect.XmlAudit": {"xmlAudit"},
		},
		XML: x.AOP,
	}
}

func TestFindOnFixture(t *testing.T) {
	h := fixtureHierarchy(t)
	in := fixtureInput(t)
	r := Find(h, in)
	if r.NoCreator || r.Truncated {
		t.Fatalf("NoCreator %v Truncated %v", r.NoCreator, r.Truncated)
	}
	if got := strings.Join(r.Aspects, ","); got != "fx.aop.aspect.TracingAspect,fx.aop.aspect.Unregistered,fx.aop.aspect.XmlAudit" {
		t.Errorf("aspects %s", got)
	}
	if r.RegisteredAspects() != 2 || r.XMLCount() != 2 || len(r.Unresolved) != 0 {
		t.Errorf("registered aspects %d (want 2), XML advice %d (want 2), unresolved %v", r.RegisteredAspects(), r.XMLCount(), r.Unresolved)
	}

	// Advice (Aspect.method) → advised join points (Bean#method).
	short := func(fqcn string) string { return fqcn[strings.LastIndexByte(fqcn, '.')+1:] }
	matched := map[string][]string{}
	for _, p := range r.Pairs {
		k := short(p.Aspect) + "." + p.Advice
		matched[k] = append(matched[k], short(p.Bean)+"#"+p.Name)
	}
	if testing.Verbose() {
		for k, v := range matched {
			sort.Strings(v)
			t.Logf("%s: %s", k, strings.Join(v, ","))
		}
		for _, a := range r.Advice {
			t.Logf("%s.%s undecided %d unknown %v returning %q", a.Aspect, a.Name, a.Undecided, a.Unknown, a.Returning)
		}
	}
	want := map[string]string{
		"TracingAspect.before":    "OrderService#cancel,OrderService#internal,OrderService#place,ReportService#render",
		"TracingAspect.afterSave": "JdbcRepo#save",
		"TracingAspect.onError":   "JdbcRepo#find,JdbcRepo#name,JdbcRepo#save",
		"TracingAspect.audit":     "OrderService#cancel",
		"TracingAspect.monitor":   "ReportService#render",
		"TracingAspect.ghost":     "",
		// args(id) bound to long: find's long, and save's Object may hold
		// a Long (Spring tests the value at run time).
		"TracingAspect.withArgs":    "JdbcRepo#find,JdbcRepo#save",
		"TracingAspect.beforeFind":  "JdbcRepo#find",
		"TracingAspect.byBean":      "JdbcRepo#name",
		"TracingAspect.noBean":      "",
		"TracingAspect.byTarget":    "ReportService#render",
		"TracingAspect.byThis":      "ReportService#render",
		"TracingAspect.byTargetAnn": "ReportService#render",
		"TracingAspect.byExecAnn":   "OrderService#cancel",
		// returning String: name, and find's Object return (save is void).
		"TracingAspect.onString":  "JdbcRepo#find,JdbcRepo#name",
		"TracingAspect.stringArg": "JdbcRepo#save,OrderService#internal,OrderService#place",
		"Unregistered.never":      "",
		"XmlAudit.audit":          "ReportService#render",
		"XmlAudit.ghost":          "",
	}
	for adv, w := range want {
		got := matched[adv]
		sort.Strings(got)
		if strings.Join(got, ",") != w {
			t.Errorf("%s advises %v, want %s", adv, got, w)
		}
	}
	if len(r.Advice) != len(want) {
		t.Errorf("%d advice, want %d", len(r.Advice), len(want))
	}
	for _, a := range r.Advice {
		if a.Name == "never" && a.Registered {
			t.Error("Unregistered#never: the aspect is no bean")
		}
		if a.Undecided != 0 || len(a.Unknown) != 0 {
			t.Errorf("%s.%s: undecided %d, unknown %v", short(a.Aspect), a.Name, a.Undecided, a.Unknown)
		}
	}

	// Edges: from the call sites, with conditions; none from the
	// self-invocation in place().
	edges := map[string]string{}
	for _, e := range r.Edges {
		edges[e.CallerFQCN+"#"+e.CallerName+" -> "+short(e.TargetFQCN)+"."+e.TargetName] = e.Condition
	}
	for k, cond := range map[string]string{
		"fx.app.Main#main -> TracingAspect.before":     "",
		"fx.app.Main#main -> TracingAspect.afterSave":  "returned",
		"fx.app.Main#main -> TracingAspect.onError":    "exceptional",
		"fx.app.Main#main -> TracingAspect.audit":      "",
		"fx.app.Main#main -> TracingAspect.monitor":    "",
		"fx.app.Main#main -> TracingAspect.withArgs":   "",
		"fx.app.Main#main -> TracingAspect.beforeFind": "",
		"fx.app.Main#main -> TracingAspect.onString":   "returned",
		"fx.app.Main#main -> XmlAudit.audit":           "",
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
		for _, never := range []string{"ghost", "never", "noBean"} {
			if strings.HasSuffix(k, "."+never) {
				t.Errorf("edge %s: the advice matches nothing", k)
			}
		}
	}
	if r.SelfCalls != 1 {
		t.Errorf("self calls %d, want 1 (place → internal)", r.SelfCalls)
	}
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
		{"execution(@com.x.Ann * *(..))", true, "*", "", "*", "", ".."},
		{"execution(* *(@com.x.Ann *))", false, "", "", "", "", ""},
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
	if x, ok := o.r.(annNode); !ok || x.raw != "com.x.Ann" || x.designator != onMethod {
		t.Errorf("annotation %#v", o.r)
	}
	for _, e := range []string{"@args(com.x.A)", "cflow(execution(* *(..)))", "execution(* *(..)", "", "within(com.x.A) &&",
		"execution(* *(@com.x.A *))", "args(id)", "this(svc)", "@annotation(ann)"} {
		why := map[string]bool{}
		unknownReasons(bind(parse(e), scope{}), &resolver{cache: map[string]node{}, active: map[string]bool{}}, why, 0)
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
	r := Find(h, fixtureInput(t))
	if !r.NoCreator || len(r.Edges) != 0 || len(r.Advice) == 0 {
		t.Errorf("NoCreator %v, %d edges, %d advice", r.NoCreator, len(r.Edges), len(r.Advice))
	}
}

func TestParseXMLWords(t *testing.T) {
	n := parseXML("execution(* a.B.*(..)) and not within(a.android..*) or bean(order)")
	o, ok := n.(orNode)
	if !ok {
		t.Fatalf("%#v", n)
	}
	a, ok := o.l.(andNode)
	if !ok {
		t.Fatalf("%#v", o.l)
	}
	if w, ok := a.r.(notNode); !ok || w.x.(withinNode).pat != "a.android..*" {
		t.Errorf("%#v", a.r)
	}
}

func TestMethodScope(t *testing.T) {
	h := fixtureHierarchy(t)
	m := func(desc string, names ...string) classfile.Method {
		return classfile.Method{Descriptor: desc, ParameterNames: names}
	}
	for _, tc := range []struct {
		expr, argNames, returning, throwing string
		m                                   classfile.Method
		want                                string
	}{
		// names from the class file
		{"args(a, b)", "", "", "", m("(JLjava/lang/String;)V", "a", "b"), "a=long b=java.lang.String"},
		// argNames, JoinPoint omitted
		{"args(id)", "id", "", "", m("(Lorg/aspectj/lang/JoinPoint;J)V"), "id=long"},
		// inference: one args variable, one parameter
		{"args(id)", "", "", "", m("(Lorg/aspectj/lang/JoinPoint;J)V"), "id=long"},
		// inference: two variables are ambiguous
		{"args(a, b)", "", "", "", m("(JJ)V"), ""},
		// inference: returning takes the one parameter
		{"execution(* *(..))", "", "r", "", m("(Ljava/lang/String;)V"), "r=java.lang.String"},
		// inference: throwing takes the Throwable, the args variable the rest
		{"args(x)", "", "", "e", m("(Ljava/lang/Exception;I)V"), "e=java.lang.Exception x=int"},
		// inference: the annotation variable takes the annotation-typed parameter
		{"@annotation(a) && args(x)", "", "", "", m("(Lfx/aop/svc/Audited;I)V"), "a=fx.aop.svc.Audited x=int"},
	} {
		p := parse(tc.expr)
		sc := methodScope(h, p, tc.m, tc.argNames, tc.returning, tc.throwing)
		var got []string
		for k, v := range sc {
			got = append(got, k+"="+v)
		}
		sort.Strings(got)
		if strings.Join(got, " ") != tc.want {
			t.Errorf("%s %+v: %v, want %s", tc.expr, tc.m, got, tc.want)
		}
	}
}

func TestArgsAndReturn(t *testing.T) {
	h := fixtureHierarchy(t)
	for _, tc := range []struct {
		pats, types string
		want        tri
	}{
		{"long", "long", yes},
		{"long", "int", no},
		{"java.lang.Long", "long", yes},
		{"long", "java.lang.Object", yes}, // may hold a Long
		{"java.lang.String", "java.lang.Integer", no},
		{"*,..", "int,long", yes},
		{",..", "int", unknown}, // an unbound variable
	} {
		split := strings.Split(tc.pats, ",")
		if got := argsMatch(h, split, strings.Split(tc.types, ",")); got != tc.want {
			t.Errorf("argsMatch(%s; %s) = %v, want %v", tc.pats, tc.types, got, tc.want)
		}
	}
	if returnMatches(h, "()V", "java.lang.Object") != yes || returnMatches(h, "()V", "java.lang.String") != no ||
		returnMatches(h, "()Ljava/lang/Object;", "java.lang.String") != yes || returnMatches(h, "()I", "java.lang.Integer") != yes {
		t.Error("returnMatches")
	}
}
