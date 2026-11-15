package jpadispatch

import (
	"testing"

	"ariadne/internal/callgraph"
	"ariadne/internal/classfile"
	"ariadne/internal/hierarchy"
)

const (
	emf        = "javax.persistence.EntityManagerFactory"
	em         = "javax.persistence.EntityManager"
	query      = "javax.persistence.Query"
	typedQuery = "javax.persistence.TypedQuery"
	createEM   = "()Ljavax/persistence/EntityManager;"
	createQ    = "(Ljava/lang/String;)Ljavax/persistence/Query;"
	listDesc   = "()Ljava/util/List;"
	sfImpl     = "org.hibernate.internal.SessionFactoryImpl"
	session    = "org.hibernate.internal.SessionImpl"
	queryImpl  = "org.hibernate.query.internal.QueryImpl"
	daoFind    = "()V"
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

func iface(fqcn string, supers []string, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: "java.lang.Object", Interfaces: supers, AccessFlags: classfile.AccInterface | classfile.AccAbstract, Methods: methods}
}

func class(fqcn string, ifaces []string, methods ...hierarchy.Method) *hierarchy.ClassInfo {
	return &hierarchy.ClassInfo{FQCN: fqcn, SuperClass: "java.lang.Object", Interfaces: ifaces, Methods: methods}
}

func call(owner, name, desc string, pc int) classfile.Call {
	return classfile.Call{Kind: classfile.CallInterface, Owner: owner, Name: name, Desc: desc, PC: pc}
}

// fixture, shaped like Hibernate: the provider's bootstrap creates the
// factory, the factory creates the session (the entity manager), the
// session creates the query, which is also a TypedQuery. The DAO goes
// factory → entity manager → query → result list through the javax
// interfaces only. Another entity manager nothing creates is in the input.
func fixture() *hierarchy.Hierarchy {
	return hierarchy.NewForTesting([]*hierarchy.ClassInfo{
		iface(emf, nil, method("createEntityManager", createEM, true)),
		iface(em, nil, method("createQuery", createQ, true)),
		iface(query, nil, method("getResultList", listDesc, true)),
		iface(typedQuery, []string{query}, method("getResultList", listDesc, true)),
		class("org.hibernate.jpa.HibernatePersistenceProvider", nil,
			creating(method("createEntityManagerFactory", "()V", false), sfImpl)),
		class(sfImpl, []string{emf},
			creating(method("createEntityManager", createEM, false), session)),
		class(session, []string{em},
			creating(method("createQuery", createQ, false), queryImpl)),
		class(queryImpl, []string{typedQuery}, method("getResultList", listDesc, false)),
		class("org.example.OrphanEntityManager", []string{em}, method("createQuery", createQ, false)),
		class("app.Dao", nil,
			method("find", daoFind, false,
				call(emf, "createEntityManager", createEM, 1),
				call(em, "createQuery", createQ, 5),
				call(query, "getResultList", listDesc, 9),
				call(typedQuery, "getResultList", listDesc, 13))),
	})
}

// TestObjectEdgesGatedByCreator: with the provider's bootstrap reachable,
// the DAO's calls reach the factory's, the session's and the query's
// bodies — a fixed point: each edge makes the next class's creator
// reachable — through Query and TypedQuery alike; without it, none. An
// entity manager nothing creates is never a candidate.
func TestObjectEdgesGatedByCreator(t *testing.T) {
	h := fixture()
	edges := ObjectEdges(h)
	for _, e := range edges {
		if e.TargetFQCN == "org.example.OrphanEntityManager" {
			t.Errorf("candidate to a class nothing creates: %+v", e)
		}
	}
	type target struct{ owner, name, desc string }
	want := []struct {
		callOwner string
		to        target
	}{
		{emf, target{sfImpl, "createEntityManager", createEM}},
		{em, target{session, "createQuery", createQ}},
		{query, target{queryImpl, "getResultList", listDesc}},
		{typedQuery, target{queryImpl, "getResultList", listDesc}},
	}
	for _, bootstrap := range []bool{false, true} {
		g := callgraph.Build(h, nil)
		id := func(owner, name, desc string) uint32 {
			n, ok := g.ID(owner, name, desc)
			if !ok {
				t.Fatalf("no node %s#%s", owner, name)
			}
			return n
		}
		from := id("app.Dao", "find", daoFind)
		roots := []uint32{from}
		if bootstrap {
			roots = append(roots, id("org.hibernate.jpa.HibernatePersistenceProvider", "createEntityManagerFactory", "()V"))
		}
		callgraph.NewCreatorGate(g, h, edges, callgraph.BasisModel).Apply(g, roots)
		for _, w := range want {
			e, ok := g.EdgeBetween(from, id(w.to.owner, w.to.name, w.to.desc))
			if ok != bootstrap {
				t.Errorf("bootstrap %v: edge Dao#find → %s#%s: %v", bootstrap, w.to.owner, w.to.name, ok)
			}
			if ok && e.Basis != callgraph.BasisModel {
				t.Errorf("edge to %s has basis %v, want model", w.to.owner, e.Basis)
			}
		}
		if _, ok := g.EdgeBetween(from, id("org.example.OrphanEntityManager", "createQuery", createQ)); ok {
			t.Error("edge to an entity manager nothing creates")
		}
	}
}

// TestTypesAreSkipped: every listed type is a receiver CHA never expands;
// a type CHA already expands (jakarta.persistence) would get duplicate
// edges.
func TestTypesAreSkipped(t *testing.T) {
	for _, ty := range Types {
		if !callgraph.DefaultSkip(ty) {
			t.Errorf("%s is not under the CHA skip rule", ty)
		}
	}
	if callgraph.DefaultSkip("jakarta.persistence.EntityManager") {
		t.Error("jakarta.persistence is under the skip rule: list its types too")
	}
}
