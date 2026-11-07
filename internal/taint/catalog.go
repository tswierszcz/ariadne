package taint

import "strings"

// CatalogVersion versions taint's catalogs (sources, guards) and rules; a
// data snapshot records it, so a change invalidates a replay.
const CatalogVersion = "taint-1"

// Source rule IDs.
const (
	RuleSpringParam    = "src:spring-mvc-param"    // an annotated Spring MVC handler parameter
	RuleSpringImplicit = "src:spring-mvc-implicit" // an unannotated, request-bound handler parameter
	RuleJaxRSParam     = "src:jaxrs-param"         // a JAX-RS resource parameter
	RuleServletCall    = "src:servlet-request"     // the result of a ServletRequest accessor
)

// springParamAnnotations bind a handler parameter to request data.
var springParamAnnotations = map[string]string{
	"org.springframework.web.bind.annotation.RequestParam":   "@RequestParam",
	"org.springframework.web.bind.annotation.RequestBody":    "@RequestBody",
	"org.springframework.web.bind.annotation.RequestHeader":  "@RequestHeader",
	"org.springframework.web.bind.annotation.PathVariable":   "@PathVariable",
	"org.springframework.web.bind.annotation.CookieValue":    "@CookieValue",
	"org.springframework.web.bind.annotation.RequestPart":    "@RequestPart",
	"org.springframework.web.bind.annotation.MatrixVariable": "@MatrixVariable",
}

// springMappings mark a Spring MVC handler method.
var springMappings = map[string]bool{
	"org.springframework.web.bind.annotation.RequestMapping": true,
	"org.springframework.web.bind.annotation.GetMapping":     true,
	"org.springframework.web.bind.annotation.PostMapping":    true,
	"org.springframework.web.bind.annotation.PutMapping":     true,
	"org.springframework.web.bind.annotation.DeleteMapping":  true,
	"org.springframework.web.bind.annotation.PatchMapping":   true,
}

// springImplicitTypes are the parameter types Spring MVC binds from the
// request without an annotation (a simple value as a request parameter,
// the body as a stream). Other unannotated parameters (Model, the
// request object, a @ModelAttribute bean) are not sources themselves:
// data read from the request object is, through servletCalls.
var springImplicitTypes = map[string]bool{
	"Ljava/lang/String;": true, "[Ljava/lang/String;": true, "[B": true,
	"Ljava/io/InputStream;": true, "Ljava/io/Reader;": true,
	"Lorg/springframework/web/multipart/MultipartFile;": true,
}

// jaxrsParamAnnotations bind a JAX-RS resource parameter to request data
// (javax.ws.rs and jakarta.ws.rs).
var jaxrsParamAnnotations = map[string]string{}

func init() {
	for _, p := range []string{"javax.ws.rs.", "jakarta.ws.rs."} {
		for _, a := range []string{"QueryParam", "FormParam", "PathParam", "HeaderParam", "CookieParam", "MatrixParam", "BeanParam"} {
			jaxrsParamAnnotations[p+a] = "@" + a
		}
	}
}

// servletCalls are the request accessors whose result is request data,
// by the owner the bytecode names (any descriptor).
var servletCalls = map[string]bool{}

func init() {
	for _, pkg := range []string{"javax.servlet.", "jakarta.servlet."} {
		for _, owner := range []string{"ServletRequest", "http.HttpServletRequest", "ServletRequestWrapper", "http.HttpServletRequestWrapper"} {
			for _, m := range []string{"getParameter", "getParameterValues", "getParameterMap", "getParameterNames",
				"getInputStream", "getReader", "getHeader", "getHeaders", "getQueryString", "getCookies",
				"getPart", "getParts", "getRequestURI", "getRequestURL", "getPathInfo"} {
				servletCalls[pkg+owner+"#"+m] = true
			}
		}
	}
}

// Opaque is flowsum's Options.Opaque for taint: the source accessors are
// never resolved into an implementation's body.
func Opaque(owner, name string) bool { return servletCalls[owner+"#"+name] }

// servletSource: whether a flowsum call source ("owner#name(desc)…") is a
// cataloged request accessor.
func servletSource(call string) bool {
	if i := strings.IndexByte(call, '('); i > 0 {
		call = call[:i]
	}
	return servletCalls[call]
}

// Guard kinds.
const (
	// GuardPolicy: the entry's own type policy (DESERIALIZATION Z2) — the
	// only guard relevant to a deserialization sink.
	GuardPolicy = "policy"
	// GuardUnrelated: a transformation or encoder the data passes
	// through that does not restrict what is deserialized.
	GuardUnrelated = "unrelated_sanitizer"
	// GuardControlCheck: a check of the data that only decides a branch
	// (control dependence, not tracked): it does not clean the data.
	GuardControlCheck = "control_check"
)

// Guard states of a result.
const (
	GuardEffective = "effective" // the entry's policy rejects every candidate class
	GuardPartial   = "partial"   // the entry's policy rejects some candidate classes
	GuardUnknown   = "unknown"   // a policy exists that is not evaluated
	GuardOnlyOther = "unrelated" // only unrelated sanitizers / branch checks
	GuardNone      = "none"
)

type guardSpec struct{ kind, what string }

// guardCalls: calls a value on the path may pass through that look like
// protection but are not one for a deserialization sink, keyed by owner
// (as named in the bytecode) and name.
var guardCalls = map[string]guardSpec{}

func init() {
	add := func(kind, what, owner string, names ...string) {
		for _, n := range names {
			guardCalls[owner+"#"+n] = guardSpec{kind, what}
		}
	}
	u, c := GuardUnrelated, GuardControlCheck
	add(u, "string replacement", "java.lang.String", "replace", "replaceAll", "replaceFirst")
	add(u, "HTML escaping", "org.springframework.web.util.HtmlUtils", "htmlEscape", "htmlEscapeDecimal", "htmlEscapeHex")
	add(u, "JavaScript escaping", "org.springframework.web.util.JavaScriptUtils", "javaScriptEscape")
	add(u, "URI encoding", "org.springframework.web.util.UriUtils", "encode", "encodeQueryParam", "encodePath", "encodePathSegment")
	add(u, "URL encoding", "java.net.URLEncoder", "encode")
	for _, o := range []string{"org.apache.commons.text.StringEscapeUtils", "org.apache.commons.lang3.StringEscapeUtils", "org.apache.commons.lang.StringEscapeUtils"} {
		add(u, "string escaping", o, "escapeHtml4", "escapeHtml3", "escapeHtml", "escapeXml", "escapeXml10", "escapeXml11",
			"escapeJava", "escapeEcmaScript", "escapeJavaScript", "escapeSql", "escapeCsv")
	}
	add(u, "OWASP encoder", "org.owasp.encoder.Encode", "forHtml", "forHtmlContent", "forHtmlAttribute", "forJavaScript", "forUri", "forUriComponent", "forXml", "forXmlContent", "forXmlAttribute")
	add(u, "HTML cleaning", "org.jsoup.Jsoup", "clean")
	add(c, "emptiness check", "java.lang.String", "isEmpty", "isBlank")
	add(c, "content check", "java.lang.String", "matches", "startsWith", "endsWith", "contains", "equals", "equalsIgnoreCase")
	for _, o := range []string{"org.apache.commons.lang3.StringUtils", "org.apache.commons.lang.StringUtils"} {
		add(c, "emptiness check", o, "isEmpty", "isNotEmpty", "isBlank", "isNotBlank")
		add(c, "content check", o, "isAlphanumeric", "isAlpha", "isNumeric", "startsWith", "endsWith", "contains")
	}
	add(c, "emptiness check", "org.springframework.util.StringUtils", "isEmpty", "hasText", "hasLength")
	add(c, "pattern check", "java.util.regex.Pattern", "matches")
}
