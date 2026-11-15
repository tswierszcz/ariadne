// Package jpadispatch models calls on JPA objects — entity managers,
// their factories, queries and resource-local transactions, the
// javax.persistence/jakarta.persistence interfaces a persistence provider
// (Hibernate, EclipseLink) implements (FRAMEWORKS F2, JVM_GRAPH skip rule;
// ROADMAP B4 b4-6).
//
// The call graph never CHA-expands a receiver under java./javax.
// (callgraph.DefaultSkip), so a DAO calling `em.find(...)` on an injected
// EntityManager kept only its edge to the interface method. The
// framework:proxies catalog then links the call to Spring ORM's
// invocation handler (the shared or extended entity manager, the deferred
// query, the managed factory), but the handler reaches the provider
// through Method#invoke on its target, which no literal resolves: the b4-6
// oracle (testdata/jpa) measured every provider body the JVM ran
// unreached.
//
// Only javax.persistence (JPA ≤ 2.2) is listed: jakarta.persistence
// (Jakarta Persistence ≥ 3.0) is not under the skip rule, so CHA already
// expands calls on it to every implementation in the input.
//
// Like jdbc:objects, the model links each such call to the body every
// concrete implementation in the input runs, gated by creation: an
// implementation counts once a method creating it (`new`) is reachable
// from the roots — the provider's entity manager once its factory's
// createEntityManager runs, a query once the entity manager's createQuery
// runs. An implementation created only reflectively (a provider factory
// instantiated from META-INF/persistence.xml) has no creator and gets no
// edge.
package jpadispatch

import (
	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// ModelID identifies the model in logs; its edges carry
// callgraph.BasisModel.
const ModelID = "jpa:objects"

// Version versions the model's semantics (the receiver types). Snapshots
// record it.
const Version = "e1"

// Types are the JPA interfaces whose calls the model links. A call on
// Query reaches TypedQuery and StoredProcedureQuery implementations
// through the hierarchy; the subinterfaces are listed for calls made on
// them.
var Types = []string{
	"javax.persistence.EntityManager",
	"javax.persistence.EntityManagerFactory",
	"javax.persistence.EntityTransaction",
	"javax.persistence.Query",
	"javax.persistence.StoredProcedureQuery",
	"javax.persistence.TypedQuery",
}

// ObjectEdges returns the candidate edges for every virtual/interface
// call on a JPA type (Types), each to an implementation's body with the
// methods creating that implementation; apply them with
// callgraph.NewCreatorGate.
func ObjectEdges(h *hierarchy.Hierarchy) []callgraph.SAMEdge {
	return callgraph.ObjectEdges(h, Types)
}
