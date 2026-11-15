package springevents

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
	"ariadne/internal/input"
)

const fixture = "../../testdata/"

// fixtureHierarchy reads testdata/events (scripts/events-oracle.sh) over
// the Spring 5.3.6 jars of testdata/tx.
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
	h, _, err := hierarchy.Build(fixture+"events/app.jar", dj)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// The fixture's bean types: the stereotypes and EventsConfig's @Bean products.
var fixtureBeans = []string{
	"fx.ev.SignupService", "fx.ev.OrderService", "fx.ev.MailListener", "fx.ev.MetricsListener",
	"fx.ev.ShippingListener", "fx.ev.CancelListener", "fx.ev.StartupListener", "fx.ev.GhostListener", "fx.ev.Ticker",
	"org.springframework.transaction.PlatformTransactionManager", "fx.ev.AuditListener",
}

func TestFindOnFixture(t *testing.T) {
	h := fixtureHierarchy(t)
	r := Find(h, fixtureBeans)

	got := map[string]Listener{}
	for _, l := range r.Listeners {
		got[l.FQCN+"#"+l.Name] = l
	}
	want := map[string]struct {
		ann        string
		events     []string
		registered bool
	}{
		"fx.ev.MailListener#onSignup":     {eventListener, []string{"fx.ev.UserSignedUp"}, true},
		"fx.ev.MetricsListener#any":       {eventListener, []string{"fx.ev.DomainEvent"}, true},
		"fx.ev.AuditListener#onPlaced":    {eventListener, []string{"fx.ev.OrderPlaced"}, true},
		"fx.ev.ShippingListener#onPlaced": {"org.springframework.transaction.event.TransactionalEventListener", []string{"fx.ev.OrderPlaced"}, true},
		"fx.ev.StartupListener#ready":     {eventListener, []string{"org.springframework.context.event.ContextRefreshedEvent"}, true},
		"fx.ev.GhostListener#onGhost":     {eventListener, []string{"fx.ev.NeverPublished"}, true},
		"fx.ev.Unregistered#onSignup":     {eventListener, []string{"fx.ev.UserSignedUp"}, false},
	}
	if len(got) != len(want) {
		t.Errorf("listeners %v, want %d", got, len(want))
	}
	for k, w := range want {
		l, ok := got[k]
		if !ok {
			t.Errorf("listener %s not found", k)
			continue
		}
		if l.Annotation != w.ann || !slices.Equal(l.Events, w.events) || l.Registered != w.registered {
			t.Errorf("%s = %+v, want %+v", k, l, w)
		}
	}

	sites := map[string]Site{}
	for _, s := range r.Sites {
		sites[s.CallerFQCN+"#"+s.CallerName] = s
	}
	for k, exact := range map[string]string{
		"fx.ev.SignupService#signup": "fx.ev.UserSignedUp",
		"fx.ev.OrderService#place":   "fx.ev.OrderPlaced",
		"fx.ev.OrderService#cancel":  "fx.ev.OrderCancelled",
		"org.springframework.context.support.AbstractApplicationContext#finishRefresh": "org.springframework.context.event.ContextRefreshedEvent",
	} {
		s, ok := sites[k]
		if !ok || !slices.Equal(s.Exact, []string{exact}) || s.Unbounded {
			t.Errorf("site %s = %+v (found %t), want exactly %s", k, s, ok, exact)
		}
	}
	if s, ok := sites["org.springframework.context.event.ApplicationListenerMethodAdapter#publishEvent"]; !ok || len(s.Bounds) != 0 {
		t.Errorf("adapter result site = %+v (found %t), want no types (every fixture listener is void)", s, ok)
	}
	if r.Forwarding == 0 {
		t.Error("the context's forwarding to its parent was not left out")
	}

	edges := map[string]bool{}
	for _, e := range r.Edges {
		edges[e.CallerFQCN+"#"+e.CallerName+" → "+e.TargetFQCN+"#"+e.TargetName] = true
		if len(e.Creators) == 0 {
			t.Errorf("edge %+v has no gate", e)
		}
	}
	for _, e := range []string{
		"fx.ev.SignupService#signup → fx.ev.MailListener#onSignup",
		"fx.ev.SignupService#signup → fx.ev.MetricsListener#any",
		"fx.ev.OrderService#place → fx.ev.AuditListener#onPlaced",
		"fx.ev.OrderService#place → fx.ev.ShippingListener#onPlaced",
		"org.springframework.context.support.AbstractApplicationContext#finishRefresh → fx.ev.StartupListener#ready",
	} {
		if !edges[e] {
			t.Errorf("missing edge %s", e)
		}
	}
	for e := range edges {
		switch e {
		case "fx.ev.SignupService#signup → fx.ev.MailListener#onSignup",
			"fx.ev.SignupService#signup → fx.ev.MetricsListener#any",
			"fx.ev.OrderService#place → fx.ev.AuditListener#onPlaced",
			"fx.ev.OrderService#place → fx.ev.ShippingListener#onPlaced",
			"org.springframework.context.support.AbstractApplicationContext#finishRefresh → fx.ev.StartupListener#ready":
		default:
			// Any other edge must come from a site with only a static bound
			// (EventPublicationInterceptor publishes an ApplicationEvent it
			// instantiates reflectively).
			caller := e[:strings.Index(e, " → ")]
			if s := sites[caller]; len(s.Exact) != 0 || len(s.Bounds) == 0 {
				t.Errorf("unexpected edge %s from site %+v", e, s)
			}
		}
	}
	for e := range edges {
		if strings.HasSuffix(e, "fx.ev.GhostListener#onGhost") || strings.HasSuffix(e, "fx.ev.Unregistered#onSignup") {
			t.Errorf("edge to a listener the JVM never runs: %s", e)
		}
	}

	// Without bean types nothing is registered: no edge.
	if r := Find(h, nil); len(r.Edges) != 0 {
		t.Errorf("no beans: %d edges", len(r.Edges))
	}
}

func TestFindWithoutAdapterCreator(t *testing.T) {
	ev := "demo.Ev"
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: publisher, SuperClass: objectFQCN, AccessFlags: classfile.AccInterface | classfile.AccAbstract},
		{FQCN: ev, SuperClass: objectFQCN, IsApp: true},
		{FQCN: "demo.L", SuperClass: objectFQCN, IsApp: true, Methods: []hierarchy.Method{{Method: classfile.Method{
			Name: "on", Descriptor: "(Ldemo/Ev;)V", AnnotationDetails: []classfile.Annotation{{Type: eventListener}},
		}}}},
	})
	r := Find(h, []string{"demo.L"})
	if !r.NoCreator || len(r.Edges) != 0 || len(r.Listeners) != 1 {
		t.Errorf("got %+v, want one listener, no creator, no edges", r)
	}
}

func TestMatches(t *testing.T) {
	h := hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		{FQCN: applicationEvent, SuperClass: "java.util.EventObject"},
		{FQCN: payloadEvent, SuperClass: applicationEvent},
		{FQCN: "demo.Created", SuperClass: applicationEvent},
		{FQCN: "demo.Sub", SuperClass: "demo.Created"},
		{FQCN: "demo.Domain", SuperClass: objectFQCN, AccessFlags: classfile.AccInterface | classfile.AccAbstract},
		{FQCN: "demo.Signed", SuperClass: objectFQCN, Interfaces: []string{"demo.Domain"}},
		{FQCN: "demo.Other", SuperClass: objectFQCN},
	})
	for _, c := range []struct {
		name          string
		exact, bounds []string
		listener      string
		want          bool
	}{
		{"exact type", []string{"demo.Signed"}, nil, "demo.Signed", true},
		{"supertype", []string{"demo.Signed"}, nil, "demo.Domain", true},
		{"unrelated", []string{"demo.Signed"}, nil, "demo.Other", false},
		{"payload wrapper", []string{"demo.Signed"}, nil, payloadEvent, true},
		{"any ApplicationEvent receives a payload", []string{"demo.Signed"}, nil, applicationEvent, true},
		{"event not a payload", []string{"demo.Created"}, nil, payloadEvent, false},
		{"event to a subtype listener", []string{"demo.Created"}, nil, "demo.Sub", false},
		{"bound to a subtype listener", nil, []string{"demo.Created"}, "demo.Sub", true},
		{"bound to an unrelated listener", nil, []string{"demo.Created"}, "demo.Other", false},
		{"Object bound reaches all", nil, []string{objectFQCN}, "demo.Other", true},
		{"ApplicationEvent bound skips payloads", nil, []string{applicationEvent}, "demo.Other", false},
		{"interface bound may be a payload", nil, []string{"demo.Domain"}, payloadEvent, true},
		{"java.util.EventObject listener", []string{"demo.Created"}, nil, "java.util.EventObject", true},
	} {
		if got := matches(h, Site{Exact: c.exact, Bounds: c.bounds}, []string{c.listener}); got != c.want {
			t.Errorf("%s: matches = %t, want %t", c.name, got, c.want)
		}
	}
}

func TestParamTypes(t *testing.T) {
	got := paramTypes("(Ljava/lang/String;I[Lfoo/Bar;[[J)V")
	want := []string{"java.lang.String", "I", "[Lfoo/Bar;", "[[J"}
	if !slices.Equal(got, want) {
		t.Errorf("paramTypes = %v, want %v", got, want)
	}
}
