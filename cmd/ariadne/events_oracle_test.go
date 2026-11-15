package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
	"ariadne/internal/profile"
)

const eventsFixture = "../../testdata/events/"

// eventsGraph runs the spring profile over testdata/events and returns the
// exported graph and the input's hierarchy.
func eventsGraph(t *testing.T, extra ...string) (callgraph.Export, *hierarchy.Hierarchy) {
	t.Helper()
	dir := t.TempDir()
	graph := filepath.Join(dir, "g.txt")
	args := []string{"-offline", "-cache-dir", filepath.Join(dir, "cache"), "-app", eventsFixture + "app.jar",
		"-deps", txFixture + "deps", "-profile", "spring", "-roots", "profile",
		"-out", filepath.Join(dir, "r.md"), "-json", filepath.Join(dir, "r.json"), "-graph-out", graph}
	args = append(args, extra...)
	if code, out := runWithEnv(args, nil); code != 0 {
		t.Fatalf("exit %d: %s", code, out)
	}
	f, err := os.Open(graph)
	if err != nil {
		t.Fatal(err)
	}
	ex, err := callgraph.Read(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	h, _, err := hierarchy.Build(eventsFixture+"app.jar", nil)
	if err != nil {
		t.Fatal(err)
	}
	return ex, h
}

// TestEventsOracle (spring:events): testdata/events (scripts/
// events-oracle.sh) publishes events through services on a Spring 5.3.6
// context — a payload object, one inside a transaction, an
// ApplicationEvent — and the container publishes ContextRefreshedEvent.
// The JVM runs the matching listeners: @EventListener methods (on the
// exact type, a supertype, classes= with no parameter, a @Bean-registered
// class) through ApplicationListenerMethodAdapter#doInvoke → Method#invoke,
// the @TransactionalEventListener after commit, and the
// ApplicationListener through the multicaster.
//
// The listener methods are reached here without the spring profile's
// annotation roots (an @EventListener is a root by itself under -roots
// profile; a @TransactionalEventListener is not): from main, the @Bean
// methods and the bean lifecycle roots, so the edges must come from the
// publishing call sites. The two listeners the JVM never runs — an event
// type nothing creates, a class that is not a bean — get no model edge.
// The @Scheduled methods (one repeated: @Schedules in the class file) are
// annotation roots of the profile.
func TestEventsOracle(t *testing.T) {
	ex, h := eventsGraph(t)
	starts := []string{"fx.app.Main#main([Ljava/lang/String;)V"}
	for _, m := range []string{"transactionManager()Lorg/springframework/transaction/PlatformTransactionManager;", "audit()Lfx/ev/AuditListener;"} {
		starts = append(starts, "fx.app.EventsConfig#"+m)
	}
	for _, r := range profile.BeanLifecycleRoots(h, nil) {
		starts = append(starts, r.FQCN+"#"+r.Name+r.Desc)
	}
	reach := map[string]bool{}
	for _, s := range starts {
		for b := range reachableFrom(ex, s) {
			reach[b] = true
		}
	}
	all := (benchFixture{observed: eventsFixture + "observed.txt", bodyColumn: 0}).observedBodies(t)
	var missed []string
	n := 0
	for b := range all {
		if strings.Contains(b, "$$") || strings.HasPrefix(b, "jdk.proxy") || strings.HasPrefix(b, "fx.ev.Ticker#") {
			continue // generated at run time; scheduled methods are checked below
		}
		if strings.HasPrefix(b, "fx.") || strings.HasPrefix(b, "org.springframework.context.event.") {
			n++
			if !reach[b] {
				missed = append(missed, b)
			}
		}
	}
	if n < 15 {
		t.Fatalf("observed.txt has %d input bodies; regenerate with scripts/events-oracle.sh", n)
	}
	sort.Strings(missed)
	for _, b := range missed {
		t.Errorf("the JVM ran %s, the graph does not reach it", b)
	}

	model := map[string]map[string]bool{} // caller → model targets
	for id := range ex.Edges {
		if id.Basis == string(callgraph.BasisModel) && strings.HasPrefix(id.To, "fx.ev.") {
			if model[id.From] == nil {
				model[id.From] = map[string]bool{}
			}
			model[id.From][id.To] = true
		}
	}
	const (
		signup  = "fx.ev.SignupService#signup(Ljava/lang/String;)V"
		place   = "fx.ev.OrderService#place()V"
		mail    = "fx.ev.MailListener#onSignup(Lfx/ev/UserSignedUp;)V"
		metrics = "fx.ev.MetricsListener#any(Lfx/ev/DomainEvent;)V"
		audit   = "fx.ev.AuditListener#onPlaced(Lfx/ev/OrderPlaced;)V"
		ship    = "fx.ev.ShippingListener#onPlaced(Lfx/ev/OrderPlaced;)V"
	)
	for _, w := range []struct{ from, to string }{{signup, mail}, {signup, metrics}, {place, audit}, {place, ship}} {
		if !model[w.from][w.to] {
			t.Errorf("no model edge %s → %s", w.from, w.to)
		}
	}
	for from, tos := range model {
		for to := range tos {
			if strings.HasPrefix(to, "fx.ev.GhostListener#") || strings.HasPrefix(to, "fx.ev.Unregistered#") {
				t.Errorf("model edge %s → %s: the JVM never runs it", from, to)
			}
			if from == signup && (to == audit || to == ship) || from == place && (to == mail || to == metrics) {
				t.Errorf("model edge %s → %s: a listener of another event type", from, to)
			}
		}
	}

	cat, err := profile.LoadDefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	prof, _ := cat.Get("spring")
	g := callgraph.Build(h, nil)
	roots := map[string]bool{}
	for _, id := range prof.AppRoots(g, h) {
		roots[g.Key(id)] = true
	}
	for _, s := range []string{"fx.ev.Ticker#tick()V", "fx.ev.Ticker#twice()V"} {
		if !all[s] {
			t.Fatalf("observed.txt lacks %s; regenerate with scripts/events-oracle.sh", s)
		}
		if !roots[s] {
			t.Errorf("the JVM ran the scheduled %s; the spring profile's roots do not include it", s)
		}
	}
}
