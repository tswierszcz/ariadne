// Package jdbcdispatch models calls on JDBC objects — connections,
// statements, result sets and the other java.sql/javax.sql interfaces a
// driver or a connection pool implements (FRAMEWORKS F2, JVM_GRAPH skip
// rule).
//
// The call graph never CHA-expands a receiver under java./javax.
// (callgraph.DefaultSkip): Spring's JdbcTemplate calling
// `ps.executeQuery()` or Hibernate calling `rs.getString(...)` keeps only
// its edge to the interface method, so the driver's and the pool's bodies
// (H2's JdbcPreparedStatement, Hikari's HikariProxyPreparedStatement) are
// never reached. Lifting the skip would fan out to every JDBC class on the
// class path: every driver jar, every pool, every wrapper. The model
// instead links each such call to the body every concrete implementation
// in the input runs, gated by creation: an implementation counts once a
// method creating it (`new`) is reachable from the roots — the pool's
// proxies once its proxy factory runs, the driver's connection once the
// driver's connect runs (the service-provider dispatch model links
// java.sql.Driver#connect to the drivers registered in META-INF/services).
//
// java.sql.Driver itself is not listed: it is a service interface the
// service-provider dispatch model already handles. An implementation
// created only reflectively (a DataSource class named in configuration
// and instantiated by Spring Boot's DataSourceBuilder) has no creator and
// gets no edge.
package jdbcdispatch

import (
	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// ModelID identifies the model in logs; its edges carry
// callgraph.BasisModel.
const ModelID = "jdbc:objects"

// Version versions the model's semantics (the receiver types). Snapshots
// record it.
const Version = "j1"

// Types are the JDBC interfaces whose calls the model links. All are
// interfaces under java.sql/javax.sql, implemented by drivers and pools.
var Types = []string{
	"java.sql.Array",
	"java.sql.Blob",
	"java.sql.CallableStatement",
	"java.sql.Clob",
	"java.sql.Connection",
	"java.sql.DatabaseMetaData",
	"java.sql.NClob",
	"java.sql.ParameterMetaData",
	"java.sql.PreparedStatement",
	"java.sql.ResultSet",
	"java.sql.ResultSetMetaData",
	"java.sql.SQLXML",
	"java.sql.Savepoint",
	"java.sql.Statement",
	"java.sql.Wrapper",
	"javax.sql.DataSource",
}

// ObjectEdges returns the candidate edges for every virtual/interface
// call on a JDBC type (Types), each to an implementation's body with the
// methods creating that implementation; apply them with
// callgraph.NewCreatorGate.
func ObjectEdges(h *hierarchy.Hierarchy) []callgraph.SAMEdge {
	return callgraph.ObjectEdges(h, Types)
}
