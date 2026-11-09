package jdbcdispatch

import (
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

const (
	conn         = "java.sql.Connection"
	ps           = "java.sql.PreparedStatement"
	prepareDesc  = "(Ljava/lang/String;)Ljava/sql/PreparedStatement;"
	execDesc     = "()Ljava/sql/ResultSet;"
	connectDesc  = "()Ljava/sql/Connection;"
	hikariPS     = "com.zaxxer.hikari.pool.ProxyPreparedStatement"
	hikariPSImpl = "com.zaxxer.hikari.pool.HikariProxyPreparedStatement"
)

func method(name, desc string, abstract bool, calls ...classfile.Call) hierarchy.Method {
	m := hierarchy.Method{Method: classfile.Method{Name: name, Descriptor: desc}, Calls: calls}
	if abstract {
		m.AccessFlags = classfile.AccAbstract
	}
	return m
}

func creating(m hierarchy.Method, classes ...string) hierarchy.Method {
	for i, c := range classes {
		m.InitRefs = append(m.InitRefs, classfile.InitRef{Kind: classfile.InitNew, Owner: c, PC: 100 + i})
	}
	return m
}

func iface(fqcn string, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: "java.lang.Object", AccessFlags: classfile.AccInterface | classfile.AccAbstract, Methods: methods}
}

func class(fqcn, super string, ifaces []string, abstract bool, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	c := &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: super, Interfaces: ifaces, Methods: methods}
	if abstract {
		c.AccessFlags = classfile.AccAbstract
	}
	return c
}

func call(owner, name, desc string, pc int) classfile.Call {
	return classfile.Call{Kind: classfile.CallInterface, Owner: owner, Name: name, Desc: desc, PC: pc}
}

// fixture, shaped like Hibernate over Hikari over H2: the application
// prepares and executes a statement through the java.sql interfaces; the
// pool's proxy statement (created by its proxy factory) inherits
// executeQuery from an abstract class outside the interface and delegates
// to the driver's statement; the driver's statement is created by its
// connection, created by its connect. Another driver's statement is
// created only by a method nothing calls, and one implementation has no
// `new` anywhere.
func fixture() *hierarchy.Hierarchy {
	return hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		iface(conn, method("prepareStatement", prepareDesc, true)),
		iface(ps, method("executeQuery", execDesc, true)),
		class(hikariPS, "java.lang.Object", []string{ps}, true,
			method("executeQuery", execDesc, false, call(ps, "executeQuery", execDesc, 3))),
		class(hikariPSImpl, hikariPS, nil, false),
		class("com.zaxxer.hikari.pool.ProxyFactory", "java.lang.Object", nil, false,
			creating(method("getProxyPreparedStatement", "()V", false), hikariPSImpl)),
		class("org.h2.Driver", "java.lang.Object", nil, false,
			creating(method("connect", connectDesc, false), "org.h2.jdbc.JdbcConnection")),
		class("org.h2.jdbc.JdbcConnection", "java.lang.Object", []string{conn}, false,
			creating(method("prepareStatement", prepareDesc, false), "org.h2.jdbc.JdbcPreparedStatement")),
		class("org.h2.jdbc.JdbcPreparedStatement", "java.lang.Object", []string{ps}, false,
			method("executeQuery", execDesc, false)),
		class("com.mysql.cj.jdbc.ClientPreparedStatement", "java.lang.Object", []string{ps}, false,
			method("executeQuery", execDesc, false)),
		class("com.mysql.cj.jdbc.Unused", "java.lang.Object", nil, false,
			creating(method("make", "()V", false), "com.mysql.cj.jdbc.ClientPreparedStatement")),
		class("org.example.NeverCreated", "java.lang.Object", []string{ps}, false,
			method("executeQuery", execDesc, false)),
		class("app.Dao", "java.lang.Object", nil, false,
			method("find", "()V", false,
				classfile.Call{Kind: classfile.CallStatic, Owner: "com.zaxxer.hikari.pool.ProxyFactory", Name: "getProxyPreparedStatement", Desc: "()V", PC: 1},
				call(conn, "prepareStatement", prepareDesc, 5),
				call(ps, "executeQuery", execDesc, 9))),
	})
}

// TestObjectEdgesGatedByCreator: a call on a JDBC interface reaches the
// body of every implementation a reachable method creates — the pool's
// proxy (its inherited body), and the driver's statement once the
// driver's connect is reachable (a fixed point: the connection's edge
// makes the statement's creator reachable) — and never one only an
// unreachable method creates or one nothing creates.
func TestObjectEdgesGatedByCreator(t *testing.T) {
	h := fixture()
	edges := ObjectEdges(h)
	for _, e := range edges {
		if e.TargetFQCN == "org.example.NeverCreated" {
			t.Errorf("candidate to a class nothing creates: %+v", e)
		}
	}
	for _, tc := range []struct {
		name       string
		withDriver bool
		want       map[string]bool // target class of executeQuery from Dao#find → edge expected
	}{
		{"pool only", false, map[string]bool{hikariPS: true, "org.h2.jdbc.JdbcPreparedStatement": false, "com.mysql.cj.jdbc.ClientPreparedStatement": false}},
		{"pool and driver", true, map[string]bool{hikariPS: true, "org.h2.jdbc.JdbcPreparedStatement": true, "com.mysql.cj.jdbc.ClientPreparedStatement": false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := callgraph.Build(h, nil)
			id := func(owner, name, desc string) uint32 {
				n, ok := g.ID(owner, name, desc)
				if !ok {
					t.Fatalf("no node %s#%s", owner, name)
				}
				return n
			}
			roots := []uint32{id("app.Dao", "find", "()V")}
			if tc.withDriver {
				roots = append(roots, id("org.h2.Driver", "connect", connectDesc))
			}
			gate := callgraph.NewCreatorGate(g, h, edges, callgraph.BasisModel)
			gate.Apply(g, roots)
			from := id("app.Dao", "find", "()V")
			for target, want := range tc.want {
				e, ok := g.EdgeBetween(from, id(target, "executeQuery", execDesc))
				if ok != want {
					t.Errorf("edge Dao#find → %s#executeQuery: %v, want %v", target, ok, want)
				}
				if ok && e.Basis != callgraph.BasisModel {
					t.Errorf("edge to %s has basis %v, want model", target, e.Basis)
				}
			}
			// The pool's proxy delegates to the driver: that edge too only
			// with a reachable creator of the driver's statement.
			_, ok := g.EdgeBetween(id(hikariPS, "executeQuery", execDesc), id("org.h2.jdbc.JdbcPreparedStatement", "executeQuery", execDesc))
			if ok != tc.withDriver {
				t.Errorf("proxy → driver edge: %v, want %v", ok, tc.withDriver)
			}
		})
	}
}

// TestTypesAreSkipped: every listed type is a receiver CHA never expands;
// a type CHA already expands would get duplicate, ungated edges.
func TestTypesAreSkipped(t *testing.T) {
	for _, ty := range Types {
		if !callgraph.DefaultSkip(ty) {
			t.Errorf("%s is not under the CHA skip rule", ty)
		}
	}
}
