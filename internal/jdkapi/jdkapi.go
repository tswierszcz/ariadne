// Package jdkapi holds facts about JDK interfaces the analyzed hierarchy
// cannot show: the hierarchy reads only the input's classes, so that
// java.sql.PreparedStatement extends java.sql.Statement is unknown to it,
// and a class implementing only PreparedStatement is not found among the
// subtypes of Statement (ROADMAP B4 b4-4 measured this: MyBatis closes a
// driver's statement through Statement#close).
package jdkapi

import (
	"sort"

	"ariadne/internal/hierarchy"
)

// Version versions the table.
const Version = "k1"

// superinterfaces are the direct java.sql/javax.sql superinterfaces of the
// JDBC interfaces, read by reflection from JDK 25 (Class#getInterfaces);
// unchanged since JDBC 4.1 (Java 7).
var superinterfaces = map[string][]string{
	"java.sql.CallableStatement":         {"java.sql.PreparedStatement"},
	"java.sql.Connection":                {"java.sql.Wrapper"},
	"java.sql.DatabaseMetaData":          {"java.sql.Wrapper"},
	"java.sql.NClob":                     {"java.sql.Clob"},
	"java.sql.ParameterMetaData":         {"java.sql.Wrapper"},
	"java.sql.PreparedStatement":         {"java.sql.Statement"},
	"java.sql.ResultSet":                 {"java.sql.Wrapper"},
	"java.sql.ResultSetMetaData":         {"java.sql.Wrapper"},
	"java.sql.Statement":                 {"java.sql.Wrapper"},
	"javax.sql.DataSource":               {"javax.sql.CommonDataSource", "java.sql.Wrapper"},
	"javax.sql.ConnectionPoolDataSource": {"javax.sql.CommonDataSource"},
	"javax.sql.XADataSource":             {"javax.sql.CommonDataSource"},
	"javax.sql.XAConnection":             {"javax.sql.PooledConnection"},
	"javax.sql.RowSet":                   {"java.sql.ResultSet"},
}

// subinterfaces is superinterfaces inverted.
var subinterfaces = func() map[string][]string {
	out := map[string][]string{}
	for sub, supers := range superinterfaces {
		for _, s := range supers {
			out[s] = append(out[s], sub)
		}
	}
	return out
}()

// Subinterfaces returns iface's transitive JDK subinterfaces from the
// table, sorted; nil when it has none.
func Subinterfaces(iface string) []string {
	seen := map[string]bool{}
	var out []string
	queue := []string{iface}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, s := range subinterfaces[cur] {
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
				queue = append(queue, s)
			}
		}
	}
	sort.Strings(out)
	return out
}

// Subtypes is h.Subtypes(iface) plus the input's subtypes of every JDK
// subinterface of iface (the table), without duplicates: iface's own first,
// in h.Subtypes order, then each subinterface's in turn.
func Subtypes(h *hierarchy.Hierarchy, iface string) []string {
	out := append([]string(nil), h.Subtypes(iface)...)
	subs := Subinterfaces(iface)
	if len(subs) == 0 {
		return out
	}
	seen := map[string]bool{}
	for _, s := range out {
		seen[s] = true
	}
	for _, sub := range subs {
		for _, s := range h.Subtypes(sub) {
			if !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	return out
}
