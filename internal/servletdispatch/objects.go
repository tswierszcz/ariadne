package servletdispatch

import (
	"ariadne/internal/callgraph"
	"ariadne/internal/hierarchy"
)

// objectTypes are the Servlet API types whose instances the container and
// the registered filters pass along a request: the request, the response,
// their wrappers and the filter chain. A call on one of them is a
// javax./jakarta. receiver CHA never expands (callgraph.DefaultSkip), so
// a filter's chain.doFilter(...) or a controller's request.getHeader(...)
// keeps only its edge to the interface method and never reaches the
// container's or a wrapper's body. jakarta.* is not under the skip rule —
// CHA already expands those calls — so only javax.servlet is listed.
var objectTypes = func() []string {
	var out []string
	for _, p := range []string{"javax.servlet."} {
		for _, t := range []string{"FilterChain", "ServletRequest", "ServletResponse", "ServletRequestWrapper", "ServletResponseWrapper",
			"http.HttpServletRequest", "http.HttpServletResponse", "http.HttpServletRequestWrapper", "http.HttpServletResponseWrapper"} {
			out = append(out, p+t)
		}
	}
	return out
}()

// ObjectEdges returns, for every virtual/interface call on a request,
// response, wrapper or filter-chain type (objectTypes), a candidate edge to
// the body each concrete implementation in the input runs for it
// (hierarchy.ResolveExact), with the methods that create that
// implementation (`new`) as its creators. The candidates go through a
// callgraph creator gate: an implementation counts only once a method
// creating it is reachable — the container's own request/response and
// chain classes, and the wrappers of the filters that run, not the
// wrappers of filters nobody registers (dozens on a Spring/Tomcat class
// path). The result is sorted.
func ObjectEdges(h *hierarchy.Hierarchy) []callgraph.SAMEdge {
	return callgraph.ObjectEdges(h, objectTypes)
}
