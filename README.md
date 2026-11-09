
# ariadne
<p align="center"><img src="docs/ariadne-icon.png" width="160" alt="Ariadne logo"></p>

CVE reachability analysis for JVM applications. ariadne takes a built Java
application and its resolved Maven dependencies, builds a static call graph
straight from bytecode, cross-references the dependencies against known
vulnerabilities (OSV), and reports which of those vulnerabilities are
actually **reachable** from the application's own code — as opposed to
merely present on the classpath.

No JDK is required to run it: ariadne parses `.class` files itself. A JDK
is only needed to build the Java *target* applications and test fixtures
this repository ships with.

See [supported capabilities and limitations](docs/SUPPORT.MD) for the current
scope and evidence behind support claims. Future implementation plans are
listed separately in [RECOMMENDATION](RECOMMENDATION/README.MD).

See [PLAN.md](PLAN.md) for the full design rationale, phase-by-phase
history, and every finding made while building this (including two real
bugs found by running the tool end-to-end, with before/after numbers).
This file describes how the finished pipeline actually works.

## The pipeline

Running `ariadne` executes eight steps in order. Each one is a separate
Go package under `internal/`, independently testable, and the CLI
(`cmd/ariadne`) just wires them together. One more package,
`internal/profile` (see step 7), isn't a numbered step of its own: it
runs right after step 3 to pick the active profile, then feeds step 4
(synthetic edges) and step 7 (the framework-package list and, with
`-roots profile`, the root set) rather than sitting in the middle of the
pipeline by itself.

```
-app, -deps, -deps-list
        │
        ▼
 1. internal/input      Normalize -app into a single jar; resolve every
                         dependency jar to a Maven coordinate (GAV).
        │
        ▼
 2. internal/classfile  Parse every .class file in the app + dependency
                         jars (no JDK); decode each method's bytecode into
                         call-site edges.
        │
        ▼
 3. internal/hierarchy  Index every class read: superclass, interfaces,
                         methods, annotations, source jar/GAV, app-vs-dep.
        │
        ▼
 4. internal/callgraph  Intern every method into an integer id; build the
                         directed call graph; expand virtual/interface
                         calls via class-hierarchy analysis (CHA).
        │
        ▼
 5. internal/cves       Resolved GAVs → OSV querybatch → per-id details →
                         deduplicated CVE list (cves.json-shaped).
        │
        ▼
 6. internal/vulnmap    CVE → the methods that are actually vulnerable
                         (manual table, then GitHub fix-commit diffs).
        │
        ▼
 7. internal/reach      BFS from every app method over the call graph;
                         assign each CVE a status; reconstruct the
                         shortest path for REACHABLE/FRAMEWORK ones.
        │
        ▼
 8. internal/report     Render report.md (Markdown) and results.json
                         (deterministic) from the finished analysis.
```

### 1. Input normalization (`internal/input`)

`-app` accepts five forms, auto-detected:

- a **directory** of `.class` files (e.g. `target/classes`) — packed into
  a temporary jar;
- a **plain jar** — used as-is;
- a **Spring Boot fat jar** — detected by a `BOOT-INF/` entry.
  `BOOT-INF/classes/` is repacked into a temporary jar (with the
  `BOOT-INF/classes/` prefix stripped) and every `BOOT-INF/lib/*.jar` is
  extracted to a temporary directory, which is then used as the
  dependency source automatically (`-deps` is not needed in this case);
- a **WAR** — detected by a `WEB-INF/` entry. `WEB-INF/classes/` and
  `WEB-INF/lib/*.jar` are handled exactly like a fat jar's
  `BOOT-INF/classes`/`BOOT-INF/lib`;
- an **EAR** — detected by `META-INF/application.xml`, or (since that
  file has been optional since Java EE 5) by a nested `*.war` entry when
  it's missing. Every nested WAR module's `WEB-INF/classes`/`WEB-INF/lib`
  is merged the same way (a class name shared by two modules keeps
  whichever module was read first, same "first one wins" rule
  `internal/hierarchy` already applies to a class duplicated across two
  dependency jars), and every top-level `*.jar` is extracted as a
  dependency jar — an EAR's own EJB-module jars are not distinguished
  from shared library jars, so an EJB module's own classes are not
  picked up as app code (see Known limitations).

Every dependency jar (from `-deps`, or extracted from a fat
jar/WAR/EAR) is then resolved to a Maven coordinate, in priority order:

1. `META-INF/maven/<group>/<artifact>/pom.properties` **inside the jar
   itself** — works regardless of which build tool produced it. If a jar
   contains more than one such file (a shaded/uber jar bundling other
   libraries' poms), the one whose own `artifact-version` matches the
   jar's file name is preferred, and the rest are recorded as a warning.
   When *none* of them matches the file name but the `-deps-list` entry
   for that exact file does, the list wins (groovy 2.4.x carries only
   commons-cli's `pom.properties`, repackaged inside it; taking it would
   query OSV for the wrong artifact). The other entries are examined as
   **embedded components** (CVE:F6a): a module of the
   jar's own project (same version, related group) is part of the jar; a
   separate library is compared class by class with its jar of exactly
   that version from Maven Central (`ARIADNE_MAVEN_CENTRAL` overrides the
   repository root). When its classes are there (same bytes, or the same
   content after a shading tool rewrote them), it becomes its own
   component instance inside the jar, with its own OSV lookup and its
   own findings; otherwise it is listed as an unconfirmed candidate
   (`content_differs`, `partial`, `classes_absent` — e.g. relocated,
   `reference_unavailable`) and not looked up. The reference jars are
   cached; an offline run reads them from the cache.

   **Relocated copies (CVE:F7a).** A library shaded under another
   package is found when the relocation is known: from the jar's own
   `pom.xml` (maven-shade-plugin `<relocation>` rules, read
   automatically) or from an `-embedded-components` file (pinned to the
   jar's SHA-256; also for a candidate without `pom.properties`). The
   reference classes are then looked up under their relocated names and
   compared with every relocated reference mapped back; a confirmed
   relocated instance shows both names and the rule's source, and an
   advisory's methods are matched under their relocated names. A wrong
   map, contradicting rules, two candidates claiming the same classes
   (`collision`) or a minimized copy (`partial`) never confirm anything.

   **Candidates without `pom.properties`, SBOM, candidate versions
   (CVE:F6b).** Two more sources name candidates: the jar's own
   `pom.xml` — every dependency (compile/runtime, not optional, and
   transitively through its Central pom) that a maven-shade
   `<artifactSet><includes>` pattern names, with `${…}` versions resolved
   through parent poms and `dependencyManagement` (jruby-core's shaded
   ASM 9.0 is found this way) — and an optional `-sbom` (CycloneDX JSON
   1.4–1.6), whose components nested in a component carrying a jar's
   SHA-256 become candidates of that jar (nothing else in the SBOM is
   used). Every candidate is still confirmed only by its classes. When
   the declared version does not fit the classes, or there is none, up
   to 8 other published versions (Central's version list) are compared:
   one fitting version is confirmed from the classes, several give an
   instance with a set of candidate versions (`g:a:{v1,v2}`; advisories
   of every candidate, findings marked `VERSION_CANDIDATES`, applicability
   `undetermined` unless all are affected), none is `version_unknown` — a
   coverage gap, not "no CVE".
2. `-deps-list` (the output of `mvn dependency:list`), matched by file
   name (`artifactId-version[-classifier].jar`). The parser tolerates
   ANSI color codes and the ` -- module x [auto]` suffix newer versions
   of the plugin append.
3. If neither source identifies a jar, it is reported as **unidentified**
   in the final report and excluded from the OSV query — it is never
   silently dropped.

**Gradle builds** need no dedicated code path: `-app` already accepts a
directory of classes (e.g. `build/classes/java/main`), and most
Gradle-resolved jars from Maven Central already carry their own
`pom.properties`. The one piece Maven's wrapper scripts give for free —
`mvn dependency:copy-dependencies`/`dependency:list` — has a Gradle
equivalent in `scripts/gradle-copy-deps.init.gradle`: an init script with
two tasks, `ariadneCopyDeps` (copies every `runtimeClasspath` jar into one
directory) and `ariadneDepsList` (writes a `deps.txt` in the exact format
step 2 above already parses), so `-deps`/`-deps-list` consume a Gradle
build exactly as they do a Maven one — no new Go parser. Verified
end-to-end against spring-petclinic built with Gradle instead of Maven
(same Spring Boot 2.4.5 era as the Maven-built target below): 80/80
dependency jars identified, and `REACHABLE` CVE count matched the
Maven-built baseline exactly (3/3). See
[PLAN.md](PLAN.md)'s Phase 7 for the full verification, including a
build quirk the script works around (a dead JCenter-only Gradle plugin in
that exact petclinic commit).

**WAR and EAR support was verified the same way**, by repackaging the
same Maven-built petclinic (`target/classes` + `target/dependency/*.jar`)
as a WAR (`WEB-INF/classes`, `WEB-INF/lib`) and as an EAR wrapping that
WAR: both ran with no code path other than `-app`/`-deps-list`, and both
reproduced the Maven-built baseline exactly — 87/87 dependency jars
identified, 148 CVEs, `REACHABLE` 3/3.

### 2–3. Bytecode parsing and the class hierarchy (`internal/classfile`, `internal/hierarchy`)

`internal/classfile` is a from-scratch, pure-Go JVM class-file parser: the
constant pool (every tag through Java 21, including `MethodHandle`,
`Dynamic`/`InvokeDynamic`), class/field/method structure,
`RuntimeVisibleAnnotations`, and a full bytecode walker over the `Code`
attribute's instructions. For each method it extracts every call-site
edge:

| Bytecode instruction | Edge kind | Resolution |
|---|---|---|
| `invokevirtual` | `virtual` | Methodref's owner/name/descriptor, as compiled (not necessarily the declaring class — see below) |
| `invokespecial` | `special` | same (constructors, super calls, private-method calls) |
| `invokestatic` | `static` | same |
| `invokeinterface` | `interface` | same |
| `invokedynamic` | `dynamic` | resolved through `BootstrapMethods`: if the bootstrap is `LambdaMetafactory`/`altMetafactory`, the real target method is read from its 2nd bootstrap argument (the implementation `MethodHandle`) — this is how lambdas and method references become graph edges instead of disappearing. Any other bootstrap (e.g. `StringConcatFactory` for `+` string concatenation on newer bytecode) is counted as **unresolved** rather than guessed at. |

`internal/hierarchy` reads every class from the app and every dependency
jar (in parallel), and indexes: superclass, interfaces, declared methods
(with their extracted call edges and annotations), which jar it came
from, its resolved GAV, and whether it's app code or a dependency. It
also answers two questions the next two steps need:

- **`Subtypes(fqcn)`** — every transitive subtype of a class/interface
  (memoized), for CHA expansion.
- **`ResolveMethod(fqcn, name)`** — which class in the superclass/
  superinterface chain actually declares a given method name (JVM lookup
  order: the class itself, then its superclass chain, then interfaces
  reachable from that chain, breadth-first).

A duplicate class found in two different jars (a classpath conflict) is
kept from whichever jar was read first and reported as a warning, not an
error.

### 4. The call graph and CHA expansion (`internal/callgraph`)

**Strategy, settings and comparison (GRAPH_PRECISION P0).** Which
implementations a virtual/interface call may run is chosen by a strategy
(`internal/callgraph/strategy.go`); everything else below — interning,
direct edges, class initialization, evidence — is the same for every
strategy. CHA is the default (`-callgraph-mode cha`); RTA is an opt-in
second mode (`-callgraph-mode rta`, P2, below), never a silent change of
the default. The graph's settings — mode, builder version, skip rule,
unknown-allocation policy, roots, target Java — are printed in the report
header and recorded in v2 (`callgraph`), and two graphs are compared edge
by edge with `-graph-out` / `-graph-compare`: by identity, not by size,
with the findings whose verdict or path changed and the edges that explain
it. Fewer edges or fewer REACHABLE results are not by themselves an
improvement. **Created types (P1):** every run also catalogs the classes
the input may create — by `new`, as a ServiceLoader provider, through a
resolved reflective constructor, as a Spring bean, as a deserialization
candidate, or as the class of an entry point — each with the sites whose
reachability creates it, plus the gaps where a class may be created by a
mechanism the catalog does not model (dynamic proxies, unresolved
reflection, `Unsafe`, constructor handles, unknown services,
container-generated repositories, and with `-roots all` the root methods'
parameters). Report-only in CHA mode ("Created types" section, v2
`callgraph.created_types`, the class list in `-graph-out`). Checked against
the real JVM on `testdata/alloc` (`scripts/alloc-oracle.sh`). **RTA (P2,
opt-in `-callgraph-mode rta`):** the graph is built exactly as for CHA and
every model edge is added; then a fixed point over the methods reachable
from the roots and the classes they create (the catalog) decides which
classes are live, and a CHA edge stays only if a live class dispatches the
call to its target. Direct, class-init and model edges are untouched. A
reachable gap applies the unknown-allocation policy
(`cha-fallback-with-gap`): its fallback classes become live — for
unresolved reflection, `Unsafe` and constructor handles the classes the
code can name at run time (class literals, class-name strings, annotated
classes, field/parameter/return types), for an unknown ServiceLoader
service every registered provider, for `-roots all` every concrete subtype
of a root's reference parameter — and the gap is reported; a class named
only in a configuration file is not covered (the gap stays, as a model gap
of `NOT_REACHABLE` findings). The "RTA" report section and v2
`callgraph.rta` give the counts, including the CHA edges kept only by the
fallback; compare an RTA run with a CHA run with `-graph-out` /
`-graph-compare`. `-roots all` and `-roots profile` are separate scenarios:
with `-roots all` every method is a root, so RTA removes little. **The
default stays CHA** (GRAPH_PRECISION P4,
[testdata/graphbench/RESULTS.MD](testdata/graphbench/RESULTS.MD)): on
petclinic and WebGoat RTA costs under a second and changes no verdict with
`-roots all`; with `-roots profile` it moved 8 findings from FRAMEWORK to
NOT_REACHABLE in P4, 5 correctly, 1 undetermined and 2 wrongly — container →
servlet dispatch was a skipped `javax.*` call in both modes. With the servlet
container dispatch model (below) petclinic's wrong one (CVE-2024-38809) and
the undetermined one are `FRAMEWORK` through the real request path, leaving
4 correct changes there; WebGoat's analyzed input does not contain its Boot
launcher, so nothing registers `DispatcherServlet` and its CVE-2024-38809
change remains. Use RTA as a second opinion next to a CHA run
(`-graph-compare`) and check the `servlet dispatch` log line. Checked
against the real JVM on `testdata/rta` (`scripts/rta-oracle.sh`). Checked against the real JVM on
`testdata/callgraph` (`scripts/callgraph-oracle.sh`: overloads, a default
method, generic and covariant bridges, an inherited body, a super call).

Every method is interned into a dense `uint32` id. Building the graph is
two passes:

1. **Intern every declared method** first, so a method with no outgoing
   calls at all (a getter, an empty override) still gets a node —
   `internal/reach` needs it there to use as a root.
2. **Add edges.** Every call-site edge found by `classfile` becomes a
   direct graph edge. For a `virtual` or `interface` call to `T#m`, an
   extra edge is added to `S#m` for **every transitive subtype `S` of `T`
   that itself declares `m` with the exact same descriptor** — this is
   class-hierarchy analysis (CHA): the call graph can't know which
   subtype a given object actually is at run time, so it conservatively
   assumes it could be any of them. Matching the exact descriptor
   (JVM_GRAPH G2), not just the name, matters in practice: on real
   petclinic, the CHA-added edge count roughly halved once two unrelated
   overloads of the same name stopped getting wired together.

**CHA is deliberately not applied to JDK types.** A call declared on
`java.*`/`javax.*`/`jdk.*`/`sun.*` (`Object#toString`, `Runnable#run`,
`Map#get`, ...) is never expanded to its subtypes. Without this guard,
CHA fans out into a huge fraction of the entire classpath through
nothing more than the framework's own bootstrap calling a few common
`Object`/collection methods — measured on real petclinic: **1.2M edges
with the guard vs. 11.4M without it (9.3×), built in 1.9s vs. 2m43s
(85× slower)**. The practical effect without the guard is that nearly
every CVE in every dependency would show up as "reachable", which is not
useful.

**One bounded exception: servlet container dispatch.** The skip also hides
the one `javax.*`/`jakarta.*` call every web request goes through: the
container's `servlet.service(req, res)` (`Servlet#service`) and
`filter.doFilter(req, res, chain)` (`Filter#doFilter`) — Tomcat's
`ApplicationFilterChain#internalDoFilter`, Undertow's handlers. Without
it, an app's `DispatcherServlet` is unreachable from `main` in both
`-callgraph-mode cha` and `rta`. The `servlet-container:dispatch` model
(`internal/servletdispatch`, version `s1`, a B3 registry model, so its edges
have basis `model`) expands exactly these calls, plus `HttpServlet`'s own
template self-calls (`service(HttpServletRequest, …)`, `doGet`/`doPost`/…,
`getLastModified`, only when the caller is `HttpServlet` itself), to the
**servlets and filters the app registers** — never to every subtype on the
class path: classes named in `web.xml`; application classes annotated
`@WebServlet`/`@WebFilter`; and, for a Spring Boot application (an
application class annotated `@SpringBootApplication` /
`@EnableAutoConfiguration`), Spring beans whose class is a servlet or filter
(stereotype classes, concrete `@Bean` return types — Boot registers every
such bean; plain Spring MVC does not) and a fixed catalog of auto-configured
registrations — `DispatcherServlet`, the character-encoding, form-content
and request-context filters, and Spring Security's `DelegatingFilterProxy`
with its `FilterChainProxy` delegate — each only when the
auto-configuration and its `@ConditionalOnClass` classes are on the class
path and the application does not `exclude` it. Each target is the method
the JVM selects on the registered class (an inherited body such as
`HttpServlet#service` for `DispatcherServlet` included). The log line
`servlet dispatch (s1): N registered servlets/filters (class [kind, basis];
…)` lists them with their basis, and registered classes count as
container-created types for RTA. A `@Bean` returning a
`ServletRegistrationBean`/`FilterRegistrationBean` (or a bare
`Servlet`/`Filter`) cannot be resolved statically: it makes every container
dispatch an incomplete model fact with the reason (a `model analysis:`
warning). Not modeled: servlet/filter `init`/`destroy`, listeners, servlets
and filters a container or library registers itself
(`ServletContainerInitializer`s such as Tomcat's WebSocket filter, Boot
Actuator's metrics filter), Spring Security's internal filter chain behind
`FilterChainProxy`, and programmatic registration
(`ServletContext#addServlet`). Checked against the real JVM on
`testdata/servletdispatch` (`scripts/servletdispatch-oracle.sh`: a
container that creates web.xml-declared classes by name and dispatches
through the real `javax.servlet` API; every observed body reached from
`main`, the two unregistered classes never reached — before the model,
none of the 5 observed bodies was reached).

**Request, response and filter-chain objects (model version `s2`).** Calls
*on* the objects the container passes along — `request.getHeader(…)`,
`response.sendError(…)`, a filter's `chain.doFilter(…)`, on
`javax.servlet.ServletRequest`/`ServletResponse`, their `Http*` and
`*Wrapper` types and `FilterChain` — are skipped receivers too, so they
never reached the container's or a wrapper's body. Since `s2` each such
virtual/interface call gets an edge (basis `model`) to the body every
concrete implementation in the input runs for it, **once a method creating
that implementation (`new`) is reachable from the roots** — the same
creator gate as the SAM calls: Tomcat's `RequestFacade`/`Request`,
`ApplicationFilterChain`, Undertow's `HttpServletRequestImpl`, and the
wrappers of filters that run; not the wrappers only an unreachable filter
creates (`servlet objects (…): N candidates, N edges added, …` in the log).
`jakarta.servlet` is not listed: it is not under the skip rule, CHA already
expands it. `s2` also adds Thymeleaf's `ResourceUrlEncodingFilter` to the
Boot catalog (when webjars-locator enables the resource chain; the
`spring.web.resources.chain.enabled` property is not read). Measured on a
petclinic trace: 175 → 139 relations missing from the graph
([TRACE_GAPS.MD](testdata/graphbench/TRACE_GAPS.MD)). The container's own
internals stay CHA-expanded, so a path from application code into them can
continue through CHA over-approximation (on WebGoat, 15 Undertow advisories
became `REACHABLE`, 11 of them through such paths —
[RESULTS.MD](testdata/graphbench/RESULTS.MD)).

**Undertow features a Spring Boot application leaves off (`internal/bootfeatures`,
`u1`).** Inside the container, CHA dispatches to every implementation on the
class path — a stream close reaches every XNIO `ChannelListener` (AJP, HTTP/2,
PROXY protocol, the HTTP client's), an exchange end reaches
`Http2ServerConnection` — and those classes are created by code CHA also
considers reachable, so neither a creator gate nor RTA removes them. Whether
they run is configuration: Spring Boot's Undertow factory turns on HTTP/2
only with `server.http2.enabled=true`, TLS only with `server.ssl.*` (and
`server.ssl.enabled` not false), a write timeout only through
`server.undertow.options`, and never configures AJP, the PROXY protocol,
Undertow's own authentication mechanisms or its HTTP client. When Undertow
and Boot's Undertow factory (or a `@SpringBootApplication`) are in the input,
ariadne reads the packaged `application*.properties`/`application*.yml`
(every profile; a feature any file enables is on) and the application
classes (a call, `new` or static read of the feature's API, e.g.
`Undertow$Builder#addAjpListener`, turns it on); a feature nothing turns on
is not a CHA/RTA target (direct calls are untouched). A switch whose value
is a placeholder (`server.ssl.enabled=${WEBGOAT_SSLENABLED:false}`) is the
environment's to decide, so that feature stays on. The log line
`Undertow features (Spring Boot, u1; config …): http2 off (…); …` gives
each decision; the graph's skip rule names the features left out.
Configuration outside the packaged files (environment variables, command
line, an external config directory) is not read — a deployment that
enables HTTP/2 that way is outside what the input shows. On WebGoat: 4
Undertow advisories leave `REACHABLE` (HTTP/2 HPACK, AJP, PROXY protocol,
Undertow form authentication).

**A second bounded exception: JDK/javax service-provider interfaces.**
A provider-based API — Bean Validation, JDBC, JCache — calls its provider
through an interface in a skipped package (`javax.validation.spi.
ValidationProvider#buildValidatorFactory`, `java.sql.Driver#connect`), and
hands out further objects through other interfaces of the same API
(`Configuration` → `ValidatorFactory` → `Validator`, `Connection` →
`Statement` → `ResultSet`, `CacheManager` → `Cache`). The
`service-provider:dispatch` model (`internal/spidispatch`, version `p1`, a
B3 registry model with basis `model`) expands a call whose static receiver
is such an interface:

- a **registered service** (a `META-INF/services` file in the input for a
  skipped interface) → to its registered providers present in the input;
- a **product** of a service — an interface of the same API family (the
  service's package without `.spi`) reached from the service through return
  types of its methods, transitively — → to the implementations of that
  interface found in the **jars of the registered providers** only.

Each target is the method the JVM selects on that class. An interface the
service does not hand out (a `ServletContainerInitializer` returns
nothing) has no products, so the servlet API stays with the servlet model
above. Not modeled: implementations outside the providers' jars,
providers registered only through a module `provides`, and which provider
a given call site really receives (the edges are "may run", like CHA's;
a never-created class in a provider jar is a target too). The log line
`service-provider dispatch (service-provider:dispatch p1): …` counts the
services and call signatures. Found by a runtime trace: on petclinic the
JVM ran Hibernate Validator's `HibernateValidator#buildValidatorFactory`
while the graph stopped at the `ValidationProvider` call
(CVE-2025-35036 was `NOT_REACHABLE`). Checked against the real JVM on
`testdata/spidispatch` (`scripts/spidispatch-oracle.sh`: a `javax.*` SPI
in its own jar, a registered provider and a product, an unregistered
provider and a product implementation in another jar — every observed body
reached, the other jar never a target).

CHA's own targets follow the same rule (builder version `3`): for a
virtual/interface call on `T`, every subtype of `T` declaring the method,
plus — for each concrete subtype that only inherits it — the body JVM
method selection finds for that subtype, also when it is declared in a
superclass that is not itself a subtype of `T` (a class implementing an
interface with a method inherited from an unrelated base class). Without
it, CHA missed 17 relations a recorded petclinic run took
([testdata/graphbench/TRACE_GAPS.MD](testdata/graphbench/TRACE_GAPS.MD)).

**SAM calls (builder version `4`, DYNAMIC_LINKAGE L1, `internal/samcall`).**
A lambda or method reference gets an edge from the method creating it
(the `invokedynamic` above); when it is *run* somewhere else — passed to a
helper, returned, stored in a field and called later through its
functional interface (`Supplier#get`, `Runnable#run`, a library callback
such as Spring's `ReflectionUtils$MethodCallback`) — that calling site gets
a `sam` edge to the implementation too. The value is followed with
DATA_FLOW: the receiver of every call of a functional interface's method
(an interface some lambda of the input implements, or a supertype of one),
through the method's callers (the union over call sites, also through a
supertype declaration it overrides, at most 64 callers per method), the
values a lambda body captured from its creator, the writers of a field
(every write in the field's class and nested classes), and the returns of
static/private helpers — each link resolved once and memoized; a parameter
or field whose declared type cannot hold a lambda is not followed, and a
site needing more than 20 000 links is a gap. A value handed to a method
outside the input (`List#forEach`, `Map#computeIfAbsent`,
`Optional#orElseGet`) gets the same edge from the call handing it over:
the JDK code that may run it is not in the graph. An instance-method
reference (`obj::m`, `Type::m`) is dispatch-expanded like a virtual call,
so an override is a target. Two rules keep it honest: a receiver the
analysis cannot follow adds **no** edge — never "every lambda of the
interface" — and is counted with its reason in the log and the report
header (`SAM calls (DYNAMIC_LINKAGE L1): …`); and an edge is added only
once some method creating the value is reachable from the roots (a fixed
point with the library models), so a lambda whose creator never runs does
not reach a shared helper through other callers. On a recorded petclinic
run, functional-interface calls the graph missed fell from 50 to 22
([testdata/graphbench/TRACE_GAPS.MD](testdata/graphbench/TRACE_GAPS.MD));
checked against the real JVM on `testdata/samcall`
(`scripts/samcall-oracle.sh`).

A call to a subtype's inherited (not overridden) method resolves to the
class that actually supplies the body, not a dead-end node under the
subclass's own name: `classfile.ExtractCalls` reports a Methodref's
owner exactly as the bytecode names it (for a call through a subclass
reference, the subclass — even when it never declares the method
itself), but `internal/callgraph.Build` (JVM_GRAPH G2) resolves that
against the actual declaring class first, the same superclass-then-
interface order the JVM itself uses. `internal/reach`'s method matching
(see below) keeps its own inheritance check too, as a second line of
defense for a declaring class the input doesn't actually include.

**Class initialization edges** (JVM_GRAPH G3): a `new`, `getstatic`,
`putstatic` or `invokestatic` adds an edge to the target class's own
`<clinit>`, if it has one, since the JVM runs it first (JVMS 5.5); a
class with its own `<clinit>` also gets an edge to its superclass's
`<clinit>`. These show up in the report as a third edge category next
to "direct" and "added by CHA".

**Synthetic edges** (`internal/callgraph.ApplySyntheticEdges`, applied by
the CLI right after `Build`) cover the one case CHA alone cannot: a proxy
with *no bytecode implementation at all*. A Spring Data repository
interface (`OwnerRepository extends JpaRepository<...>`) has zero
subtypes in the analyzed bytecode — Spring generates the implementation
at runtime — so without help, a call into it is a dead end even though it
really dispatches to `SimpleJpaRepository`. The active profile's
`synthetic_edges` (see step 7's "Profiles" paragraph below) name this
explicitly: every transitive subtype of a marker interface gets an edge,
for each method it declares, to the same-named method on the real
implementation class.

**Deserialization edges** (DESERIALIZATION Z1, `internal/deserialize`):
every Java serialization entry — a call of `ObjectInputStream#readObject`/
`#readUnshared` (also on a subclass) or `ObjectInput#readObject` — gets a
"may run" edge (basis `deserialization`) to every callback serialization
may run on a concrete serializable class of the input: a private,
non-static `readObject(ObjectInputStream)`/`readObjectNoData()` of each
class of the serializable chain, `readResolve()` resolved with
serialization's access rules, an Externalizable class's public no-arg
constructor and `readExternal(ObjectInput)`, and the no-arg constructor
of the first non-serializable superclass. A method with another
descriptor, or a public/static `readObject`, is not a callback; enums and
records get none. The stream decides which classes appear (a cast after
`readObject` does not bound it), so the edges say *may run*; whether the
data is attacker-controlled is not modeled. A filter in the same method
(`setObjectInputFilter`) or a `resolveClass`-overriding stream is shown,
evaluated when it is a literal (DESERIALIZATION Z2, below). Every entry is
also an unbounded model gap (`deserialization`): the JDK's own callbacks (a
`HashMap` hashing its keys) are not modeled. The report's
"Deserialization entry points" section lists every site with its policy.

**Filters and XStream (DESERIALIZATION Z2).** Each entry gets a policy
state. *Java serialization*: a literal
`setObjectInputFilter(ObjectInputFilter.Config.createFilter("a.B;a.**;!*"))`
on the same stream is evaluated with JEP 290 class patterns — rejected
classes lose their edges — while a non-literal filter or a
`resolveClass` override is `unknown` (edges kept). *XStream* `fromXML`
connects to `readResolve` of any class (declared in it or a superclass),
the private `readObject` of serializable classes and Externalizable's
constructor/`readExternal`, restricted by its security framework as the
release and the same instance's configuration in the entry's method say:
none before 1.4.7 (`unrestricted`), any type by default in 1.4.7–1.4.17,
a default allowlist of JDK basic types from 1.4.18
(`default_allowlist`); `addPermission(NONE/ANY)`, `setupDefaultSecurity`,
`allowTypes`/`denyTypes`, wildcards and type hierarchies with literal
arguments are evaluated, the last added permission first; anything else
is `unknown`. Checked on the real JVM against XStream 1.4.5, 1.4.10 and
1.4.20 (`scripts/prepare-xstream-demo.sh`). On WebGoat the
vulnerable-components lesson's `new XStream().fromXML(...)` on XStream
1.4.5 is `unrestricted`; no verdict changed (its advisories were already
reachable through the entry point). Measured: petclinic 130 entry sites, 227 callback
targets; WebGoat 142 and 296. Two advisories whose code *is* a
serialization callback became reachable on both apps — CVE-2021-46877
(jackson-databind `NodeSerialization#readExternal`) and CVE-2023-6378
(logback `LoggingEventVO#readObject`) — through Spring's
`DefaultDeserializer` (petclinic, FRAMEWORK) and JRuby's
`JavaProxy#marshal_load` (WebGoat, REACHABLE); on petclinic also
CVE-2022-45868 (H2, through `JdbcDataSource#readObject` and a long CHA
path, FRAMEWORK). Control app: `scripts/prepare-deserialization-demo.sh`.

**Candidate gadget chains (DESERIALIZATION Z3).** A small, versioned
catalog of documented chains (`internal/deserialize/chains.go`, catalog
`z3-1`): CommonsCollections6 (commons-collections 3.x), CommonsCollections2/4
(commons-collections4 4.0), CommonsBeanutils1, Groovy1 and XStream's
`java.beans.EventHandler` dynamic proxy (CVE-2013-7285). For each, the
required classes are looked up by name or under a relocation of a
confirmed embedded component (CVE:F7a), and must belong to the expected
component (a class with the right name in an unidentified or other jar
is *not confirmed*); the fixed release's **guard is read from the
analyzed bytecode**, not from a version string (3.2.2's
`InvokerTransformer#readObject` calling `FunctorUtils#checkUnsafeSerialization`,
4.1's non-Serializable `InvokerTransformer`, 2.4.4's `MethodClosure#readResolve`,
1.4.7's `ReflectionConverter.eventHandlerType`); a JDK-bound chain uses
`-target-java` (Groovy1 needs ≤ 8u71, CommonsBeanutils1's `TemplatesImpl`
sink ≤ 15). Then, per entry site of the chain's format, the site's Z2
policy: an evaluated policy rejecting any class of the chain (JDK trigger
included) blocks it there. States: `candidate_chain`, `conditions_unresolved`
(an undecided premise: unconfirmed identity, policy `unknown`, no
`-target-java`), `candidate_blocked` (this chain only — never the advisory),
`classes_absent`, `no_entry`. Report-only: no edge and no verdict changes,
and none of it is a proof of exploitability. **Data flow (Z3):** each
chain × entry carries that entry's taint result as `flow` — `flow_found`,
`no_flow_in_model`, `undetermined`, `not_assessed` with `-taint` (see
"Request data reaching the entries" below), `not_analyzed` without it — and
the chain's own `flow` is the strongest among the entries where it is not
blocked (all entries when it is blocked everywhere). Flow is a separate
reason: it never changes a chain's state, and `candidate_chain` with
`flow_found` still leaves the object graph, field values and runtime
configuration unchecked; only request data is checked (files, sockets,
queues are not). A process-wide `jdk.serialFilter`
is invisible, and an entry with no catalog chain is not free of chains.
Control app: `scripts/prepare-deserialization-chains-demo.sh` (a vulnerable
and a fixed build; the guards are observed on the real JVM by
round-tripping the single guarded class — no chain is built; with `-taint`
its request-fed `demo.Web` entries are `flow_found` — vulnerable build:
CC6/CC2-4/XStream `candidate_chain` with request data; fixed build:
blocked with request data). On
WebGoat the XStream EventHandler chain is a `candidate_chain` at the
vulnerable-components lesson's entry (XStream 1.4.5) — with `-taint`
also `flow_found` (`@RequestParam payload`), so the CVE-2013-7285 row reads
"xstream:eventhandler candidate_chain, request data flow_found"; the other four are
`classes_absent` on both apps.

**What the data read depends on (DATA_FLOW D3).** Each deserialization
entry also states what the data it reads depends on *within its own
method*: for Java serialization the stream object (with whatever was
wrapped into it — `new ObjectInputStream(new ByteArrayInputStream(
Base64.getDecoder().decode(p)))` depends on `p`), for XStream `fromXML`'s
argument — expressed as the method's parameters, `this`, fields, or
results of calls the analysis does not model. It comes from dependency
summaries (`internal/flowsum`): per method, what its return value, field
writes and call arguments depend on, applied at each call site with that
site's own arguments (two calls of one helper do not mix), through a
single analyzed target (exact, or the one CHA implementation), a small
JDK catalog, or a conservative rule (depends on everything it is given,
recorded as widened). Recursion is iterated to a fixed point with a
limit; an incomplete summary is marked `?`. Where those parameters come
from — an HTTP request — is assessed with `-taint` (below). On
WebGoat the vulnerable-components lesson's `fromXML` reads its `String`
parameter. Control app: `scripts/prepare-dataflow-demo.sh`.

**Request data reaching the entries (DATA_FLOW D4, opt-in `-taint`).**
A first taint pilot (`internal/taint`, catalog `taint-1`): does request
data reach the data a deserialization entry reads? Sources are Spring MVC
handler parameters bound from the request (`@RequestParam`,
`@RequestBody`, `@RequestHeader`, `@PathVariable`, `@CookieValue`,
`@RequestPart`, `@MatrixVariable`, and unannotated `String`/`String[]`/
`byte[]`/stream/`MultipartFile` parameters of a mapped method), JAX-RS
`@QueryParam`/`@FormParam`/`@PathParam`/`@HeaderParam`/`@CookieParam`/
`@MatrixParam`/`@BeanParam`, and the results of `ServletRequest` accessors
(`getParameter*`, `getInputStream`, `getReader`, `getHeader(s)`,
`getQueryString`, `getCookies`, `getPart(s)`, `getRequestURI`/`URL`,
`getPathInfo`; `javax` and `jakarta`). The search starts at the entry's
data in D3's terms and walks backwards: a parameter (or `this`) of the
method leads to every call site of that method in application classes,
where that site's argument is followed in the caller's terms — so a helper
called once with request data and once with a literal does not mix them —
at most 8 calls deep. Each entry gets a state: `flow_found` (with the path
from the source to the entry), `no_flow_in_model` (every input ends at a
constant or at a parameter of a method no application code calls — e.g.
`main`'s arguments, which are not a cataloged source), `undetermined`
(something on the way could not be followed: a field, an unmodeled call's
result, a lambda caller, a limit; listed as gaps), or `not_assessed` (a
library entry no application code calls). Only the entry's own Z2 type
policy can protect a deserialization entry: the guard is `effective` when
it rejects every candidate class (or XStream's default allowlist applies),
`partial` or `unknown` otherwise; string replacement, escaping/encoding
and checks that only decide a branch are listed but give `unrelated` —
they never clear a flow. With `-json-format v2`: `deserialization.taint_catalog`,
`entries[].taint` (state, rule, flows with steps, guard state, guards, gaps,
outside inputs) and `summary.entries_by_taint`; the report adds "Request
data reaching the entries (taint)". Report-only: no verdict changes, and a
flow is not proof of exploitability (nor is "no flow" proof of safety). On
WebGoat the vulnerable-components lesson's `fromXML` is `flow_found` from
`@RequestParam payload` through five `String#replace` calls (guard
`unrelated`, policy `unrestricted`); petclinic has no application entry.
Control app: `scripts/prepare-taint-demo.sh` (nine entries, one per case,
each with its expected state). Qualification (DATA_FLOW D5):
[docs/TAINT_QUALIFICATION.MD](docs/TAINT_QUALIFICATION.MD) — a real-JVM
oracle (`scripts/taint-oracle.sh`: every fixture handler is called with a
serialized canary and a process-wide filter records which entries read it)
finds no flow the model calls "no flow" and no `flow_found` the JVM did not
show, with four observed flows the model reports `undetermined` (depth
limit, a field, a lambda caller, an exception's message); the search costs
under a millisecond per run. `-taint` stays opt-in.

**Qualification and v2 (DESERIALIZATION Z4).** A chain is evaluated once
per copy of its classes — the original names and each confirmed embedded
instance carrying them relocated — so a shaded copy blocked by its guard
does not hide a plain vulnerable one (or the reverse). With
`-json-format v2` the run's deserialization model is in the document:
`deserialization` (model catalog `z2-1`, chain catalog `z3-1`, every
entry with its policy, every chain copy with its classes, reasons,
per-entry states and flows, `flow`/`flows_by_state`) and, on a finding whose advisory a chain names on the
same instance, `deserialization_chains`; the same link is shown on the
advisory's row of the report's CVE table ("gadget chain: …", with
", request data flow_found" etc. when `-taint` ran). Evidence
only: no verdict and no v1 field changes. A data snapshot records both
catalog versions (`deserialize.models`, `deserialize.chains`), so a rule
change invalidates a replay or shows up in `-snapshot-compare`.
Verified against the real JVM with `testdata/deserialize/oracle`
(`scripts/deserialize-oracle.sh`: private callbacks, a public look-alike,
Externalizable, a record, an unknown-serializability class, two
instances, a filter — no model edge the JVM did not run, two declared
gaps) and against a source-documented CVE set
(`testdata/deserialize/cves.yaml`, compared in
`testdata/deserialize/QUALIFICATION.MD`).

### 5. CVE resolution (`internal/cves`)

Every unique `(groupId:artifactId, version)` pair is queried against
`POST https://api.osv.dev/v1/querybatch` (chunked at 1000 per request,
OSV's documented limit; pagination via `/v1/query` + `page_token` is
implemented for a package with more results than one page, though no
real Maven package has been observed to need it). Each returned
vulnerability id is then fetched individually via `GET /v1/vulns/{id}`,
in parallel (8 workers).

Records that alias each other (a GHSA id and the CVE id it corresponds
to, say — OSV cross-links these in each record's own `aliases` field)
are merged into one entry via union-find, preferring a `CVE-*` id as the
primary key when the group has one. Each entry keeps: id, aliases,
summary, a severity label, which of the app's dependencies it affects,
and its references (particularly `FIX` links to a commit), plus (CVE:F2)
`published`/`modified` and every merged OSV source record with its own
dates, `withdrawn` time and affected ranges. A withdrawn record still links
its aliases: if any record of the group is active (OSV's alias records are
fetched when the query returned only the withdrawn one), the advisory stays
active; if every record is withdrawn, it is listed in the report's
"Withdrawn advisories" and not analyzed.

Every HTTP request goes through `internal/httpcache`, a disk cache keyed
by a hash of the request. `-offline` makes a cache miss a hard, clear
error instead of silently reaching the network. **Data freshness
(CVE:F2):** each body has a sidecar `<key>.meta.json` (kind, method, URL,
fetch and validation time, body SHA-256, HTTP validators — never request
headers). OSV responses, empty answers included, are fetched again once
older than `-data-ttl` (default 24h; conditionally when the server sent an
ETag/Last-Modified); GitHub content addressed by a full 40-hex commit SHA
never expires; LLM answers are never refreshed. Offline, old data is used
with a warning; online, a failed refresh is an error unless `-allow-stale`;
`-max-data-age` is a hard limit everywhere. Entries cached before F2 (no
sidecar, or one whose hash does not match) are of **unknown age**: never
given a made-up date, revalidated online, flagged offline, refused by
`-max-data-age`. The report header says `Data freshness: complete | stale |
unknown` (with a banner for the last two) and v2 carries it in
`coverage.data_freshness`.

### 6. CVE → vulnerable methods (`internal/vulnmap`)

For each CVE, in priority order:

1. **The manual table** (`internal/vulnmap/vuln_methods.yaml`, embedded
   in the binary; override with `-vuln-methods`). Matched by the CVE's
   own id or any of its aliases. This always wins when present. An entry
   is either the legacy list of `"Class#method"` strings or (CVE:F1b) an
   object `{methods: [...], role: vulnerable|entrypoint|sink, coverage:
   complete_under_declared_assumptions|partial|unknown, review:
   reviewed|unreviewed}`; the list form means partial and reviewed. An
   `entrypoint` role adds `MAPPING_ENTRYPOINT_ONLY` to the result;
   diff-derived mappings are vulnerable/partial/unreviewed, LLM ones
   vulnerable/unknown/unreviewed (or `entrypoint` when the model labelled
   every method `entry_point`; per-method roles are kept, CVE:F4).
   **Schema v2 (CVE:F4):** instead of `methods`, an entry may give
   `variants: [{package: g:a, versions: "[1.0,2.0)", methods: [...]}]`;
   each component instance uses the one variant whose package and Maven
   version range cover it. An irregular version (SNAPSHOT, non-numeric) is
   `UNKNOWN`/`VERSION_UNKNOWN`, no covering variant
   `MAPPING_VERSION_UNCOVERED`, several with different methods
   `MAPPING_CONFLICT`. A selector may carry a JVM descriptor:
   `Class#name` matches every overload, `Class#name(params)` one parameter
   list, `Class#name(params)ret` one method; an inherited call still
   reaches the declaring class's body, and a subclass overrides only the
   same descriptor. Unknown fields, bad ids, ranges and descriptors fail
   loading. Versions compare by Maven 3.9's rules (`internal/mavenversion`,
   checked against the installed Maven's own `ComparableVersion` on 38809
   pairs and `VersionRange` on 396 cases). The embedded table's
   CVE-2022-1471 (SnakeYAML) is versioned to `(,2.0)`.
2. **A GitHub fix-commit diff.** For every reference matching
   `github.com/<owner>/<repo>/commit/<sha>`, ariadne downloads
   `<url>.diff`, keeps only non-test `.java` files
   (excluding test code), and for each one fetches the full file *as it
   existed at that commit* from `raw.githubusercontent.com`. A
   brace-aware scanner (`internal/javadiff`) — which first blanks out
   every comment, string, text block and char literal so a stray `{` or
   `;` inside one never desyncs brace matching — finds which
   method/constructor encloses each line the diff actually changed.
   `if`/`for`/`while`/lambda/anonymous-class bodies are transparent (a
   changed line inside one is attributed to the *enclosing* method); a
   method whose name matches its enclosing class is renamed to `<init>`,
   matching bytecode's own convention.

   **Both sides of the fix (CVE:F5a).** The code in *your* release is the
   code **before** the fix, so ariadne also reads the commit's metadata
   from the GitHub REST API (`api.github.com/repos/<o>/<r>/commits/<sha>`
   — full SHA and parent; set `GITHUB_TOKEN` for a higher rate limit,
   the token is sent but never cached) and, for a single-parent commit,
   fetches each file at the **parent** revision too (under its old path
   for a renamed file). Every method the diff touched is then *changed*
   (declared before and after), *removed* (before only — a deleted file's
   methods included) or *added* (after only — a new helper or file). The
   mapping is the changed + removed methods under their pre-fix class
   name; added methods are not code of the vulnerable release and are
   kept only as evidence (`-json-format v2`: `mapping.fix_commits`). When
   the old side cannot be read — no metadata (e.g. an offline cache filled
   before F5), a merge or root commit, an unreadable pre-fix file — that
   commit falls back to the post-fix side and the row says so with
   `MAPPING_NEW_SIDE_ONLY`. A fix commit that cannot be fetched no longer
   fails the run: it is recorded, and an advisory none of whose commits
   could be read is `UNKNOWN` + `SOURCE_UNAVAILABLE`.

   **Checked against your bytecode (CVE:F5b).** Methods of member classes
   are named as in bytecode (`Outer$Inner#m`); lines inside anonymous or
   local classes and lambdas belong to the enclosing method. Each
   candidate is looked up in the analyzed classes (bridge/synthetic
   methods ignored): when exactly one overload matches the parameter
   types read from the source, the selector gets that descriptor
   (`Class#m(Ljava/lang/String;)V`); otherwise it stays a name selector —
   ariadne never guesses an overload. A fix commit none of whose methods
   exists in your classes is dropped from the mapping when another commit
   of the same advisory does match (typically a backport to another
   release line). v2 records per method the outcome (`bytecode`), per
   commit the OSV reference type and `release_fit`, and pairs a removed
   and an added method with near-identical bodies as `possible_rename`
   (evidence only).

   **Matched to your release line through git history (CVE:F5c).** For
   each fix commit ariadne looks up the tag of your version in the
   commit's GitHub repository (exact names `v1.2.3`, `1.2.3`,
   `artifact-1.2.3`, `repo-1.2.3`, `release-1.2.3`, `v_1.2.3`,
   `version-1.2.3`) and asks GitHub's compare API how the two relate:
   `descends_from_release`, `in_release` (your version already contains
   the commit) or `other_line`. A tag on a release commit made off the
   branch (Tomcat, Spring) counts as descending. When one commit of an
   advisory descends from your release, commits on another line are left
   out of the mapping (`excluded: release_line` in v2); the bytecode fit
   rule then applies to the rest. No tag or no answer leaves the line
   unknown and changes nothing. These lookups are GitHub API requests —
   set `GITHUB_TOKEN` when filling a cache; offline runs read them from
   the cache.
3. **An LLM (opt-in, `-llm`).** Only for a CVE sources 1 and 2 found
   nothing for. The model gets the advisory (summary + details) and,
   when the CVE links a fix commit, that commit's non-test `.java` diff;
   it answers with a JSON list of methods. Since CVE:F5c the prompt also
   lists the methods the fix changed, removed or added (read from both
   sides of the fix), whether each exists in your classes, and which
   commits are on another release line; the model is told never to name
   an added method. The reply is validated
   against a schema (one corrective retry), and every method is then
   checked to actually exist in the analyzed bytecode — an invented one
   is dropped. The report marks these `llm (diff)` or `llm (advisory)`.
   If the model saw only the advisory text, a `NOT_REACHABLE` verdict is
   reported as `UNKNOWN` instead (it could have named a method that
   exists but is not the vulnerable one). Prompts, answers and the model
   id are cached under `<cache-dir>/llm/`. Full design, including the
   exact prompt: [docs/phase4-llm.md](docs/phase4-llm.md).

A CVE vulnmap finds no methods for at all becomes `UNKNOWN` in the next
step, not an error.

### 7. Reachability (`internal/reach`)

Two multi-source breadth-first searches are run **once**, from every
application method simultaneously (not once per CVE — this is what keeps
analyzing 100+ CVEs under 100ms even on a 2.6M-edge graph):

- **Pass 1 (REACHABLE):** a node whose class's GAV matches a
  framework-package pattern is still visited (so a one-hop call straight
  from app code into a framework method still counts), but its own
  outgoing edges are never expanded — execution can't be traced *through*
  it. Frameworks dispatch via reflection/proxies the static graph can't
  see, so continuing past one would be guessing.
- **Pass 2 (FRAMEWORK):** the same roots, no such restriction.

A CVE's status is decided by where its vulnerable method(s) were first
found:

| Status | Meaning |
|---|---|
| `REACHABLE` | Found in pass 1: a path exists that never has to pass *through* framework-internal code to get there. The shortest such path is recorded, along with the first app → dependency edge on it (the "entry call"). |
| `FRAMEWORK` | Found only in pass 2: the only path runs through framework code the static graph can't follow past. Needs manual review. |
| `NOT_REACHABLE` | vulnmap has methods for this CVE, at least one of them exists in the analyzed graph, but neither pass ever visits any of them. Since CVE:F1b the row carries `reasons`: `MAPPING_PARTIAL` when the mapping is not known to be complete (every mapping today), and `MODEL_INCOMPLETE` plus `model_gaps` (counts by kind) when reachable unresolved reflection/MethodHandle/callback/ServiceLoader sites or deserialization entries could dispatch into the affected component. That gap rule is **report-only**: on the reference apps it applies to every `NOT_REACHABLE` finding (16 of 16 petclinic, 10 of 10 WebGoat, after DESERIALIZATION Z1), so turning them into `UNKNOWN` would carry no information ([docs/RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD) §4). Read the verdict as "no path in the model", not "safe". |
| `UNKNOWN` | vulnmap has no entry for this CVE at all; or (CVE:F1b, with `reasons`) **none of the mapped methods exists in the analyzed graph** (`METHOD_NOT_FOUND` — e.g. a diff-derived method absent under that name in the analyzed version; 7 petclinic and 6 WebGoat rows that v1 used to report as `NOT_REACHABLE`), or the affected instance's own copy of a mapped class lost to another jar's copy under the classpath order (`CLASSPATH_CONFLICT` — a path through the other copy is not that instance's code). |
| `NOT_APPLICABLE` | Every affected dependency is a WebJar (no `.class` file, assets under `META-INF/resources/webjars/`): client-side JavaScript/CSS a call graph cannot analyze. Checked before everything else. **Not** a verdict that the app is safe: the code runs in the browser. An empty jar alone does not qualify, since Maven aggregators such as `spring-boot-starter-web` are class-free too and CVEs filed against them (Spring4Shell) concern JVM code in other jars. |

Matching a vulnmap method name (`"Class#method"`, no descriptor — matching
is overload-insensitive everywhere in ariadne) against the graph also
checks every non-overriding transitive subtype of the named class, to
catch the inherited-method-called-through-a-subclass-reference case
described above.

**Profiles (`internal/profile`)** decide the framework-package list, and
optionally the root set, instead of a single hardcoded list.
`-profile auto` (the default) selects every embedded profile
(`internal/profile/profiles/*.yaml`) whose `detect` list names a class
present anywhere in the app or its dependencies, and merges them
(petclinic and WebGoat, both Spring Boot apps on an embedded Tomcat,
auto-detect and merge as `spring+servlet` — the embedded Tomcat jar
bundles the `javax.servlet.*` classes `servlet`'s own `detect` list
matches, confirmed not to change either app's CVE status counts, since
`spring`'s own `org.apache.tomcat.embed:*` pattern already covers that
exact dependency); nothing matching falls back to `generic`, which has
no framework packages at all, so every CVE's status for a plain
non-Spring, non-servlet, non-Jakarta-EE, non-Micronaut, non-Quarkus app
is `REACHABLE`/`NOT_REACHABLE`/`UNKNOWN`/`NOT_APPLICABLE` only, never
`FRAMEWORK`. `-profile spring`/`-profile servlet`/`-profile jakarta`/
`-profile micronaut`/`-profile quarkus`/`-profile generic` picks one
explicitly; `-profile-file path.yaml` loads a custom profile (same YAML
shape) and adds it to the catalog, or replaces one of the same name.

FRAMEWORKS F0 separates *detection* from *activation*: `-profile spring`
forces `spring` active even when `servlet` is also detected, and before
F0 that detection fact was invisible once a specific `-profile` name was
passed (only `-profile auto`'s own merge showed it, as part of the
merged name). The report header's "Frameworks detected" line now always
shows every profile whose markers were actually found, independent of
`-profile`'s own choice.

`spring`'s own list is section 4's original seed plus
`org.springframework:spring-beans`/`spring-core`/`spring-expression`/
`spring-aop` — added after a real petclinic run showed several
`REACHABLE` paths routing through deep Spring Beans/SpEL internals
(`CachedIntrospectionResults`, `SpelExpression`) that are arguably
framework-mediated, not genuine app-code reachability (see
[PLAN.md](PLAN.md)'s Phase 5/7 notes); `com.fasterxml.jackson.core:jackson-databind`
is conditional on the app carrying `@RequestBody`/`@ResponseBody`/
`@RestController` anywhere (evaluated once for the whole app, not
per call site — a precise per-call-site version would need
`internal/reach`'s BFS sink test to see the calling node as well as the
callee, a much bigger change than this simplification buys back).

`-roots profile` (default: `-roots all`, every application method, as
above) narrows the BFS roots to the active profile's own entry points
instead: a method carrying one of `roots.annotations`
(`@RequestMapping`/`@GetMapping`/.../`@Scheduled`/`@EventListener`/`@Bean`
for `spring`), every public method of a class itself carrying one,
(when `roots.main` is true) every public static `main(String[])`, and
every `roots.overrides` entry (`"fqcn#method"`, e.g. `javax.servlet.
http.HttpServlet#doGet`): every app class that is that class itself or
one of its transitive subtypes **and declares that method itself**
(an actual override, not an inherited no-op) — `servlet.yaml`'s own use
of this (`HttpServlet#do*`/`Filter#doFilter`, both the `javax.servlet.*`
and `jakarta.servlet.*` namespaces) is the first shipped profile to need
it; `spring.yaml`'s own list is still empty.

`jakarta.yaml` (JAX-RS `@GET`/`@POST`/`@PUT`/`@DELETE`/`@PATCH`/`@HEAD`/
`@OPTIONS`, EJB `@Schedule`/`@Schedules`, both namespaces) is a second
`roots.annotations` user, deliberately *not* including `@Path` itself:
`@Path` is often class-level (marking a JAX-RS resource class), and the
class-level rule above ("every public method of a class carrying one
becomes a root") would then wrongly turn every public helper method into
an entry point, not just the ones actually exposed as endpoints (which
also carry an HTTP-verb annotation). **CDI's `@Observes`** needs a third
kind of root, since it annotates a method *parameter*, not the method or
the class: `roots.parameter_annotations` (`javax.enterprise.event.Observes`/
`jakarta.enterprise.event.Observes`) matches a method whose own
`RuntimeVisibleParameterAnnotations` attribute — a separate, per-parameter
attribute `internal/classfile` now also parses, alongside the
class-/method-level `RuntimeVisibleAnnotations` it already read — carries
one of these on any parameter.

`micronaut.yaml` (`@Get`/`@Post`/`@Put`/`@Delete`/`@Patch`/`@Head`/
`@Options`, `@Scheduled`, `@EventListener`) is a third `roots.annotations`
user, with the same `@Controller`-is-class-level-only reasoning as
`@Path`/`@WebServlet` above for why `@Controller` itself is detect-only,
not a root. Unlike `spring.yaml`, it needs no `synthetic_edges`: a
Micronaut Data repository interface's implementation is generated by
Micronaut's own annotation processor *at compile time*, as a real class
in the app's own compiled output, not a JDK runtime proxy with zero
bytecode the way Spring Data's are — CHA already finds it like any other
subtype.

`quarkus.yaml` (`@Scheduled`, MicroProfile Reactive Messaging's
`@Incoming`/`@Outgoing`) is a fourth `roots.annotations` user. Quarkus's
own REST layer is plain JAX-RS, already covered by `jakarta.yaml` — a
Quarkus REST app auto-detects as `jakarta+quarkus`, the same way a Spring
Boot app auto-detects as `spring+servlet` — so this profile only adds
what's genuinely Quarkus-/MicroProfile-specific on top. No
`roots.overrides`/`synthetic_edges` either: Quarkus builds its bean/route
wiring at build time via Jandex indexing and bytecode recording, the same
"real class, not a runtime proxy" story as Micronaut.

### Runtime evidence (`internal/runtimeevents`, optional)

RUNTIME_NATIVE N0–N1. `-runtime-trace FILE` imports what one real run of
the application did — loaded classes, methods seen executing, measured
caller → callee relations — in the versioned format
[docs/RUNTIME_TRACE.MD](docs/RUNTIME_TRACE.MD), and shows it next to the
static results ("Runtime evidence" report section, v2 `runtime` and
`findings[].runtime_observed`). It **never** changes the graph or a
verdict: the absence of an event proves nothing, and a vulnerable method
seen executing is evidence that the code ran, not that an attacker's data
reached it. Evidence counts only for classes whose class file (or jar) has
the recorded SHA-256 — another build of the code, or a class known only by
name, is reported, never attached — and a class defined by two loaders
stays two definitions. Measured calls are checked against the static
graph: `not_in_graph` lists relations the run took that the static
analysis has no edge for, each with a cause read from the caller's
bytecode at the measured index (a lambda run through a functional
interface, a skipped `java.*`/`javax.*` receiver, reflection, a body
inherited from outside the call's type, a proxy, …), grouped by receiver
type — measured input for the next model rather than a guess
([testdata/graphbench/TRACE_GAPS.MD](testdata/graphbench/TRACE_GAPS.MD):
on petclinic CHA misses 7.7% of the relations the run took, RTA with
`-roots profile` 11.4%).

No Java agent is needed: `scripts/JfrToTrace.java` converts a Java Flight
Recorder recording. Stack sampling (`jdk.ExecutionSample`) gives methods
and measured relations, class loading (`jdk.ClassLoad`) the loaded
classes, and on JDK 25+ `jdk.MethodTrace` exact entries of chosen methods.
`-runtime-filter-out` writes those choices for you — every mapped
vulnerable method of the findings (`-runtime-filter-status` narrows the
verdicts) — as a JFR settings file:

```sh
bin/ariadne … -runtime-filter-out vulnerable.jfc
java -XX:StartFlightRecording=filename=run.jfr,settings=profile,settings=vulnerable.jfc,+jdk.ClassLoad#enabled=true -cp … your.Main
java scripts/JfrToTrace.java --jfr run.jfr --out run.trace --scenario "what you ran" \
     --artifact app.jar --artifact-dir deps/
bin/ariadne … -runtime-trace run.trace
```

The report then lists the advisories whose vulnerable methods were
entered, with exact counts, and those **traced but not entered in this
run** — under that scenario only, never a negative.

On petclinic (`scripts/runtime-trace-petclinic.sh`: pages and static
resources) the generated settings trace 505 vulnerable methods of 91
advisories: 26 advisories' methods are entered — all 3 statically
`REACHABLE` ones, 23 of the 74 `FRAMEWORK` ones (CVE-2024-38809's
`ServletWebRequest#validateIfNoneMatch` about 1 400 times), and
CVE-2025-35036's `ValidatorFactoryImpl#<init>` while the static verdict is
`NOT_REACHABLE` (it is reached through a `javax.validation` SPI call the
graph does not expand) — and 65 are traced but not entered in that
scenario. Runtime evidence points at a static gap; it does not override
the verdict.

### 8. The report (`internal/report`)

`report.md` is rendered from an embedded template and includes: a header
(app, timestamp, ariadne version, graph size before/after CHA, unresolved
`invokedynamic` count, dependency identification rate), a status/severity
summary, the full CVE table, every `REACHABLE` CVE's full call path with
its entry call highlighted, every `FRAMEWORK` CVE flagged for manual
review, the list of unidentified dependency jars, a "Components" inventory
(CVE:F1a: every jar's SHA-256 and stable location, its component instance with
identity evidence, embedded components and the status of every embedded `pom.properties`
candidate, and every class present in
several jars with the selected and non-selected copy), a `META-INF/services`
provider inventory, every resolved `ServiceLoader.load(...)` call site,
and recognized ServiceLoader consumption/dispatch sites with their linked
load interface or gap; the registered providers with constructible no-argument
constructors are also listed (the B3 model adds graph edges only when a reachable
`Iterator.next()` or `Provider.get()` can be traced through a bounded same-method receiver chain
to the exact `ServiceLoader.load(...)` call; the linear `stream().findFirst().get().get()`
path is also modeled at `Provider.get()`. A stream without a terminal operation remains lazy;
unlinked calls and unsupported callback/terminal flows remain explicit model gaps;
SERVICE_LOADING S0/S1/S2 — see "Known limitations" below for the limits),
every Spring MVC HTTP mapping found with its real
path(s)/verb(s), every `@Component`-family stereotype's own declared
bean name and `@Primary` status, and every `@Profile(...)`'s declared
name(s) with whether it's active (`unknown` unless `-active-profiles`
was given — FRAMEWORKS F1, sourced from real annotation element values,
not just the bare presence of the annotation), a limited, safe reading
of every `WEB-INF/web.xml` found in a WAR/EAR input (servlets, filters,
listeners, context params, with every other element named as unmodeled
rather than silently dropped — FRAMEWORKS F1), every Spring classic XML
`<bean>` found on the app's own classpath, every bean name claimed by
more than one source with a different class (flagged, never
auto-resolved), every `@Autowired` injection point resolved against the
known beans by type/qualifier/primary where possible (FRAMEWORKS F2),
and a fixed Limitations section (see below).

The entry call line also names the call kind, the resolution rule
(`direct`: the edge is exactly the call site's own symbolic target; `cha`:
class-hierarchy-analysis expanded it to a subtype, so trust it less;
`synthetic`: a profile's `synthetic_edges`, with no real call site at
all; `model`: a reachable call site connected to a modeled target
(currently the ServiceLoader classpath provider model, ROADMAP B3);
`callback`: a resolved `Thread`/`Executor`/`AccessController.doPrivileged` registration connected to the
callback's own `run`/`call` method, ASYNC_CALLBACKS A2; `methodhandle`:
a resolved `MethodHandle#invoke`/`invokeExact` connected to its target,
DYNAMIC_LINKAGE L2; `deserialization`: a Java serialization entry
connected to a serialization callback, DESERIALIZATION Z1) and, when there
is one, the bytecode offset of the actual invoke
instruction — e.g. `(virtual call, resolution: direct, pc=77)` — so a
reader can jump straight to `javap -c -p` output instead of trusting the
method names alone (`internal/reach.EntryEdge`, `internal/callgraph.Edge`;
RECOMMENDATION's JVM_GRAPH:G0, now archived — see
[archive/2026-10-22/jvm_graph/COMPLETION.md](archive/2026-10-22/jvm_graph/COMPLETION.md)).
The underlying call graph also keeps every call site that produced an
edge (not just the first one that happened to be seen) as evidence,
even when several calls to the same target collapse into one edge.

**Outputs are written together (CVE:F9a).** The report, the results
file and a `-snapshot-out` manifest are rendered first and written only
at the end of a successful run, all or nothing — a run that fails
anywhere (an OSV lookup offline without cache, a broken snapshot, a
results path that cannot be written) replaces none of them, so a report
on disk is never a partial one. The report header shows an **analysis
ID** (the v2 document's `analysis_id`, a hash of its content) and the
results file's name, format and SHA-256: a report and a results file
belong to the same run only when these match.

`results.json` is the same analysis as structured data — sorted by CVE
id, with each result's own method list sorted — so that two runs against
the same cached inputs produce **byte-for-byte identical output**
(verified: see [PLAN.md](PLAN.md)'s Phase 6 notes for a real
non-determinism bug this guarantee caught and how it was fixed; the new
`entry.kind`/`entry.basis`/`entry.pc` fields were checked for the same
byte-identical-rerun property against real petclinic/WebGoat runs).
v2 additionally records each mapping's `match_precision` (`name`,
`params`, `exact`, `mixed`), `variants` and LLM `method_roles`, and per
finding the precision of the selector a positive path reached and an
`applicability_check` — the advisory's OSV ranges evaluated with the
Maven comparator next to OSV's own answer (on the reference apps: 323 of
323 findings agree). CVE:F1b added two optional fields, `reasons` (stable reason codes, see
[docs/RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD) §3) and `model_gaps`
(relevant gap counts by kind); no existing field changed. An advisory
that affects several component instances is evaluated per instance
(the report's "Findings per component instance" section) and aggregated
into its one v1 row: a positive path wins, otherwise any unresolved
instance makes the row `UNKNOWN`. With `-json-format v2` the file is instead one
object (`schema_version: "2"`) holding those per-instance findings with
identity, applicability, mapping and reachability as separate fields;
v1 stays the default and is unchanged. Every run checks that the v2
findings aggregate to exactly the v1 verdicts.

## Usage

```bash
go build -o bin/ariadne ./cmd/ariadne

# From a build directory plus its resolved dependencies:
bin/ariadne -app target/classes -deps target/dependency -deps-list deps.txt \
  -out report.md -json results.json

# Straight from a Spring Boot fat jar (no -deps needed):
bin/ariadne -app target/myapp-1.0.0.jar -out report.md -json results.json

# Re-run against a warm cache only, never touching the network:
bin/ariadne -app ... -deps ... -out report.md -json results.json -offline
```

Flags:

| Flag | Required | Meaning |
|---|---|---|
| `-app` | yes | A directory, a plain jar, a Spring Boot fat jar, a WAR, or an EAR. |
| `-deps` | no* | Directory of dependency jars. *Not needed for a fat jar/WAR/EAR, which supply their own.* |
| `-deps-list` | no | `mvn dependency:list` output (or its Gradle equivalent, see `scripts/gradle-copy-deps.init.gradle`), used as a fallback for jars with no `pom.properties`. |
| `-vuln-methods` | no | Override the embedded `vuln_methods.yaml`. |
| `-embedded-components` | no | YAML file declaring components embedded in dependency jars and the package relocations they sit under, pinned to each jar's SHA-256 (CVE:F7a; example: `testdata/embedded/petclinic-ehcache.yaml`). Declarations are verified against the classes; one for a jar with another hash is reported and ignored. |
| `-sbom` | no | CycloneDX JSON SBOM (spec 1.4–1.6; CVE:F6b; example: `testdata/sbom/webgoat-jruby.cdx.json`). A component whose SHA-256 is an analyzed dependency jar's is tied to it; Maven components nested in it become embedded candidates of that jar, verified by their classes. Untied components and the dependency graph are not used; a disagreeing identity is reported, never applied. |
| `-profile` | no | `auto` (default, auto-detect and merge), `generic`, `spring`, `servlet`, `jakarta`, `micronaut`, `quarkus`, or a name from `-profile-file`. |
| `-profile-file` | no | Load/override one profile from a YAML file (same shape as the embedded defaults). |
| `-roots` | no | `all` (default: every application method) or `profile` (the active profile's own entry points). |
| `-target-java` | no | Java feature version (e.g. `17`) to resolve multi-release JAR variants against (JVM_GRAPH G4). Unset: every jar's base classes, regardless of any `META-INF/versions/` present. |
| `-jpms-scenario` | no | ServiceLoader launch scenario (SERVICE_LOADING S3): `classpath` (default — every jar in the unnamed module, `module-info.class` ignored, providers from `META-INF/services`) or `module-path` (every jar, the app included, on the module path: explicit modules provide only through `provides`, a provider with `public static provider()` is created through it, automatic modules provide through `META-INF/services` with a constructor, and a caller in an explicit module must declare `uses`), or `mixed` (only the `-jpms-module-path` jars on the module path, every other jar on the class path; module-path modules assumed resolved as with `--add-modules ALL-MODULE-PATH`). Recorded in the report header; `results.json`'s format is unchanged. |
| `-jpms-module-path` | with `mixed` | Comma-separated jar file-name patterns (`path.Match` syntax, e.g. `app,commons-text-*.jar`) selecting the module-path jars for `-jpms-scenario mixed`; `app` names the analyzed application. Rejected with any other scenario; a pattern matching no jar is a warning. |
| `-active-profiles` | no | Comma-separated Spring profile names to treat as active when evaluating `@Profile` conditions (FRAMEWORKS F1 — see "8. The report"). Unset: every `@Profile` condition is reported `unknown`, never a guessed `false`. |
| `-callgraph-mode` | no | Dispatch strategy for virtual/interface calls (GRAPH_PRECISION P0/P2): `cha` (the default) or `rta` (opt-in: CHA edges narrowed to the classes created by code reachable from the roots, with the unknown-allocation fallback; see section 4). Recorded with the graph's other settings in the report header and v2 `callgraph`. |
| `-graph-out` | no | Write the final call graph — every edge with a stable identity (source/target method, call kind, basis) and its call-site PCs, the graph's settings, the findings' paths and the created-type catalog (GRAPH_PRECISION P1) — as sorted text (`ariadne-callgraph-2`; large: ≈ 370 MB for WebGoat). |
| `-runtime-trace` | no | Import a runtime trace (RUNTIME_NATIVE N1, [docs/RUNTIME_TRACE.MD](docs/RUNTIME_TRACE.MD); e.g. from `scripts/JfrToTrace.java`): loaded classes, executed methods and measured calls, matched to the analyzed classes by content hash and shown next to the results (report "Runtime evidence", v2 `runtime`, `findings[].runtime_observed`); never changes the graph or a verdict. Validated before the analysis; recorded as a snapshot input. |
| `-runtime-filter-out` | no | Write a JFR settings file (`.jfc`) tracing on entry (`jdk.MethodTrace`, JDK 25+) the mapped vulnerable methods of this run's findings whose class is in the input; record the next run with `-XX:StartFlightRecording=settings=profile,settings=FILE` and import it with `-runtime-trace`. |
| `-runtime-filter-status` | no | With `-runtime-filter-out`: comma-separated verdicts whose mapped methods are traced (default `REACHABLE,FRAMEWORK,NOT_REACHABLE,UNKNOWN`). |
| `-graph-compare` | no | Compare this run's graph with a `-graph-out` file: edges added/removed by basis, findings whose verdict or path changed and the edges that explain it, and any settings difference (report "Call graph compared with …", v2 `callgraph.comparison`). |
| `-taint` | no | DATA_FLOW D4 taint pilot (opt-in, default off, report-only): assess whether request data (Spring MVC / JAX-RS handler parameters, `ServletRequest` accessors) reaches the data each deserialization entry reads — see "Request data reaching the entries". Never changes a verdict. |
| `-out` | yes | `report.md` output path. |
| `-json` | yes | `results.json` output path. |
| `-json-format` | no | `v1` (default: one row per advisory) or `v2` (CVE:F1c: components, one finding per advisory × component instance, coverage, inputs reference — schema in [docs/RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD) §5). |
| `-cache-dir` | no | Disk cache directory for OSV/GitHub HTTP responses (default `.cache`). |
| `-offline` | no | Serve only from the cache; a miss is a clear, immediate error. Old data is used with a warning (see `-max-data-age`). |
| `-data-ttl` | no | How long an OSV/advisory response counts as fresh before it is fetched again (default `24h`). |
| `-refresh` | no | Revalidate every cached OSV/advisory response now, whatever its age; never re-asks an LLM. Not with `-offline`. |
| `-allow-stale` | no | When an online refresh fails, use the cached copy and mark the report "stale" (never past `-max-data-age`). |
| `-max-data-age` | no | Hard limit on the age of any OSV/advisory data used, offline included; data of unknown age fails it. |
| `-snapshot-out` | no | Write a data snapshot to a new directory (CVE:F3; see "Data snapshots"). |
| `-snapshot-in` | no | Replay a data snapshot: no cache, no network; inputs, analyzer and algorithm versions must match. Not with `-refresh`/`-snapshot-out`. |
| `-snapshot-compare` | no | Compare this run with an earlier snapshot: which advisories' derived results were invalidated, and why (report section and v2 `coverage.snapshot_changes`). |
| `-llm-reask-changed` | no | With `-llm`: ask again when a CVE's advisory text or fix diff changed since its cached answer. Default: reuse the previous answer with a warning, so a data refresh never triggers paid requests by itself. |
| `-llm` | no | Enable the LLM source for CVEs the manual table and fix-commit diffs could not map. Needs a model and `ARIADNE_LLM_API_KEY`. |
| `-llm-model` | no | LLM model id; overrides `ARIADNE_LLM_MODEL` (e.g. `z-ai/glm-5.3`). |
| `-llm-base-url` | no | OpenAI-compatible base URL; overrides `ARIADNE_LLM_BASE_URL` (default `https://openrouter.ai/api/v1`). |
| `-exploit-check` | no | For each REACHABLE CVE, ask an LLM whether its call path plausibly carries attacker-controlled input. Uses the same `ARIADNE_LLM_*` settings as `-llm` (and works without `-llm` itself); one extra request per REACHABLE CVE. |
| `-framework-check` | no | For each FRAMEWORK CVE, ask an LLM whether its call path looks like a real framework call chain or a CHA over-approximation artifact. Same `ARIADNE_LLM_*` settings; one extra request per FRAMEWORK CVE. |
| `-version` | no | Print the ariadne version and exit. |

### Latest full test run with the LLM source

[docs/llm-test-report.md](docs/llm-test-report.md) records a complete live run: all unit
tests, one live LLM call, and full analyses of **spring-petclinic 2.4.5** and **WebGoat
8.2.2**, each once without and once with `-llm` (model `z-ai/glm-5.3`), including the
`NOT_APPLICABLE` status for WebJar-only CVEs. It covers the before/after status counts, how
the LLM answered by evidence type (fix diff vs. advisory text only), how repeatable the
answers were against an earlier run, and links to the full reports
([petclinic](docs/llm-run/petclinic.md), [WebGoat](docs/llm-run/webgoat.md)) and their
`results.json` files. A later re-run (its own "Update" section at the top of that page)
confirms the LLM source still behaves correctly against the current `spring`-profile
framework-package list and the widened source-2 file filter — both landed after the
original run, in unrelated work — with every answer served from the existing cache (no new
cost).

### Using the LLM source

The API key is read **only** from the environment (never a flag, so it
stays out of shell history and `ps`), and ariadne does not read `.env`
itself. Copy `.env.example` to `.env`, fill it in, then:

```bash
set -a; . ./.env; set +a
bin/ariadne -app ... -deps ... -out report.md -json results.json -llm
```

With `-offline`, `-llm` needs no key: every answer must already be in
`<cache-dir>/llm/`. To discard cached LLM answers and ask again, delete
that directory. After a data refresh changes a CVE's advisory text, the
previous answer for that CVE is reused (counted in the `llm:` log line,
with a warning) unless `-llm-reask-changed` is given.

### Data snapshots (CVE:F3)

```bash
bin/ariadne -app ... -deps ... -out report.md -json results.json -snapshot-out snap/
bin/ariadne -app ... -deps ... -out report.md -json results.json -snapshot-in snap/
bin/ariadne -app ... -deps ... -out report.md -json results.json -snapshot-compare snap/
```

`-snapshot-out` writes `manifest.json` and a content-addressed `objects/` store holding
only what the run read (petclinic: 23 MB, 1082 responses, against a 56 MB shared cache).
The manifest records the outcome of every cache lookup (hits, misses, errors such as an
alias's 404), the input artifacts' SHA-256 by stable location, the configuration, the tool
version, Go version, binary hash, algorithm versions, a **derivation key** per advisory
(advisory content, affected instances, mapping, the hashes of the fix commits/LLM answer it
came from, resolver/prompt versions), and the v1 verdicts. Its `data_set_id` hashes the
observations, not the time, so it names the logical data set.

`-snapshot-in` answers every lookup from the snapshot: no cache directory, no network (the
replay passes inside a network-less namespace). It refuses a different tool or algorithm
version, a changed, missing or extra input artifact, an edited manifest, and an object that
does not match its hash; a lookup the snapshot cannot answer fails the run. A different
configuration is allowed and reported (v2 `snapshot.data_set.config_matches: false`).

**What a snapshot does and does not reproduce:** it reproduces the *data* of a run. It does
not archive the analyzed binaries (they must be supplied; their hashes are checked) or the
analyzer: the same analyzer version is required, and a locally modified ("dirty") build is
marked `reproducible: false` and can replay only with the identical binary. The Go
toolchain and OS are recorded, not enforced.

`-snapshot-compare` explains, per advisory, why a result may differ from the earlier
snapshot: new or gone advisory, advisory content, affected instances, mapping or its
sources, resolver/LLM/algorithm versions, analyzer version, configuration, input artifact,
and the verdict change. Algorithm versions (`callgraph.builder`, `reach.engine`,
`vulnmap.resolver`, prompt versions) are constants bumped by hand when semantics change;
the analyzer version check is the backstop.

### Exploit-check (PLAN.md Phase 8)

`-exploit-check` asks the same kind of LLM, once per REACHABLE CVE, whether its
statically-computed call path plausibly carries attacker-controlled input to the
vulnerable call — a heuristic second opinion, not a proof, since ariadne has no
decompiler and shows the model method signatures along the path, not Java source. It is
opt-in: the flag must be passed explicitly, and it is independent of `-llm` (which only
controls CVE-to-method mapping) though it reads the same `ARIADNE_LLM_*` settings:

```bash
set -a; . ./.env; set +a
bin/ariadne -app ... -deps ... -out report.md -json results.json -exploit-check
```

Each REACHABLE CVE's verdict (`likely`/`unlikely`/`unknown`) and reason are written next to
its full call path in `report.md`'s "REACHABLE — full call paths" section, and as
`exploitability`/`exploit_reason`/`exploit_model` fields on that CVE's entry in
`results.json`. Answers are cached per `(model, prompt)` under `<cache-dir>/exploit/`, same
as `-llm`'s own cache. Verified against spring-petclinic 2.4.5's 3 REACHABLE CVEs: the model
correctly told apart `CVE-2022-22968`/`CVE-2025-22233` (`likely` — `OwnerController`'s
`@InitBinder` method wires `WebDataBinder#setDisallowedFields` from a request-bound
controller) from `CVE-2024-38820` (`unlikely` — the same call site but with a hardcoded
field name, not attacker-supplied data). NOT_REACHABLE/UNKNOWN CVEs have no checked path
to ask about and are left out of this check entirely; FRAMEWORK CVEs do have one, but get
a different question — see `-framework-check` below.

### Framework-check

`-framework-check` asks the same kind of LLM about each FRAMEWORK CVE instead: not "is this
exploitable" but "does this specific call path look like a real call chain the named
framework would make, or like a CHA over-approximation artifact" — the exact risk the
report's own "FRAMEWORK — needs manual review" section names (a virtual/interface call is
treated as reaching every subtype that declares a same-named method, whether or not that
override is ever really invoked for this call site). It is opt-in and independent of both
`-llm` and `-exploit-check`, though it reads the same `ARIADNE_LLM_*` settings:

```bash
set -a; . ./.env; set +a
bin/ariadne -app ... -deps ... -out report.md -json results.json -framework-check
```

Each FRAMEWORK CVE's verdict (`plausible`/`artifact`/`unknown`) and reason are written next
to its full call path in `report.md`'s "FRAMEWORK — needs manual review" section, and as
`path_plausibility`/`path_plausibility_reason`/`path_plausibility_model` fields on that
CVE's entry in `results.json`. Answers are cached under `<cache-dir>/framework-check/`, a
separate namespace from `-exploit-check`'s cache so the two never collide even for the same
CVE. Verified against spring-petclinic 2.4.5's 61 FRAMEWORK CVEs (a real OpenRouter run):
36 came back `artifact`, and the model independently found the same real pattern in most of
them — a spurious `StandardHost#<init> → StandardPipeline#setBasic → LifecycleBase#stop`
hop (a constructor/setter does not trigger a lifecycle shutdown) stitching together
unrelated Tomcat startup and request-processing code, a textbook CHA over-approximation.
This is a genuine finding about ariadne's own call graph, not just a plausible-sounding
guess, and a candidate for a future `cha_skip.yaml` entry to cut this specific false-positive
pattern at the source. WebGoat's 45 FRAMEWORK CVEs split 26 `plausible`/19 `artifact`.

## Building and testing

```bash
make build   # -> bin/ariadne
make test    # go test ./... — no JDK, no network required
make vet
make e2e-llm # one real LLM call with the real prompts; needs .env, costs ~nothing
```

`go test ./...` never requires a JDK or network access: every package's
test fixtures are pre-compiled `.jar` files committed under `testdata/`,
and every test that would otherwise need the network (OSV, GitHub) runs
against a `net/http/httptest` server instead.

Regenerating the fixtures (only needed if you change
`testdata/*/src/**/*.java`) does need a JDK:

```bash
make fixtures JAVAC8=/path/to/jdk11-or-later/bin/javac JAVAC21=/path/to/jdk21-or-later/bin/javac
./scripts/build-reachfixture.sh /path/to/javac
```

Preparing the target/control applications this project was developed and
verified against (also JDK-dependent; not needed to run or test ariadne
itself) uses the scripts in `scripts/`:

```bash
./scripts/prepare-petclinic.sh        /path/to/jdk11/JAVA_HOME         # Maven, Spring Boot
./scripts/prepare-webgoat.sh          /path/to/jdk17-or-later/JAVA_HOME # Maven, Spring Boot, control case
./scripts/prepare-petclinic-gradle.sh /path/to/jdk11/JAVA_HOME         # same app, Gradle instead
./scripts/prepare-servlet-demo.sh     /path/to/jdk11-or-later/JAVA_HOME # plain Servlet, control case
./scripts/prepare-jakarta-demo.sh     /path/to/jdk11-or-later/JAVA_HOME # JAX-RS+EJB+CDI, control case
./scripts/prepare-micronaut-demo.sh   /path/to/jdk11-or-later/JAVA_HOME # Micronaut, control case
./scripts/prepare-quarkus-demo.sh     /path/to/jdk17-or-later/JAVA_HOME # Quarkus, control case
./scripts/prepare-jpms-demo.sh        /path/to/jdk11-or-later/JAVA_HOME # JPMS provides/provider(), -jpms-scenario control case
./scripts/prepare-taint-demo.sh       /path/to/jdk11-or-later/JAVA_HOME # request data → deserialization entries, -taint control case
./scripts/taint-oracle.sh             /path/to/jdk17-or-later/JAVA_HOME # real-JVM oracle for the -taint fixture (testdata/taint/oracle)
./scripts/callgraph-oracle.sh         /path/to/jdk17-or-later/JAVA_HOME # real-JVM known-target oracle for the call graph (testdata/callgraph)
./scripts/alloc-oracle.sh             /path/to/jdk17-or-later/JAVA_HOME # real-JVM created-type oracle (testdata/alloc)
./scripts/rta-oracle.sh               /path/to/jdk17-or-later/JAVA_HOME # real-JVM RTA oracle: method bodies that ran (testdata/rta)
./scripts/runtime-trace-fixture.sh    /path/to/jdk25-or-later/JAVA_HOME # JFR-recorded runtime trace of testdata/runtime (rtm.jar, rtm-v2.jar, run.trace)
./scripts/runtime-trace-petclinic.sh  [OUT_DIR] [PORT]                  # write petclinic's vulnerable-method JFR settings, run it under JFR with them, send a fixed scenario, convert, import (java 25+ on PATH)
./scripts/spidispatch-oracle.sh       /path/to/jdk17-or-later/JAVA_HOME # real-JVM oracle for service-provider dispatch (testdata/spidispatch)
./scripts/servletdispatch-oracle.sh   /path/to/jdk17-or-later/JAVA_HOME [servlet-api.jar] # real-JVM servlet container dispatch oracle (testdata/servletdispatch; API jar defaults to the prepared WebGoat's)
./scripts/graph-benchmark.sh          [--repeat N]                      # CHA vs RTA on prepared petclinic/WebGoat: time, memory, edges, gaps, verdict changes (testdata/graphbench)
./scripts/serviceloader-oracle.sh                                      # real-JVM oracle for the ServiceLoader fixtures (java 11+ on PATH)
```

## Project layout

```
cmd/ariadne/            CLI entry point; wires every package below together
internal/
  input/                -app/-deps normalization, GAV resolution
  cves/                 OSV querying, alias deduplication
  httpcache/             disk cache + --offline support (shared by cves and vulnmap)
  classfile/             .class parser + bytecode walker
  hierarchy/              class universe: super/interfaces/subtypes/methods
  callgraph/              interning, edges, CHA expansion, synthetic edges
  servletdispatch/        servlet container dispatch model: registered servlets/filters
  profile/                framework knowledge: detect, roots, framework packages, synthetic edges
    profiles/              spring.yaml, servlet.yaml, jakarta.yaml, micronaut.yaml,
                             quarkus.yaml, generic.yaml (embedded defaults)
  vulnmap/                CVE -> vulnerable methods (manual table + fix-commit diffs)
    vuln_methods.yaml      the embedded manual table
  javadiff/               unified-diff parsing + the brace-aware Java method scanner
  reach/                  BFS, status assignment, path reconstruction
  report/                 Markdown/JSON rendering
    templates/
scripts/                 prepare-{petclinic,webgoat,petclinic-gradle,servlet-demo,
                         jakarta-demo,micronaut-demo,quarkus-demo,jpms-demo}.sh, build-*fixture*.sh,
                         gradle-copy-deps.init.gradle (Gradle equivalent of
                         mvn dependency:copy-dependencies/dependency:list)
testdata/
  fixture/                hand-written Java project, compiled for Java 8 and 21
  reachfixture/           a 3-jar fixture (app + framework dep + plain dep) covering
                          all four reachability statuses
work/                     cloned/built target apps (gitignored; scripts/prepare-*.sh output)
.cache/                   OSV/GitHub HTTP cache (gitignored)
```

## Known limitations

These are structural, not bugs, and are also printed in every generated
report:

- **Runtime traces are evidence of one run** (`-runtime-trace`): nothing
  not seen in a trace is less reachable, sampled traces differ between
  runs, and a trace recorded on another build of the code is not used.

- **Reflection, dynamic proxies (AOP/CGLIB, Spring Data repositories, JDK
  `Proxy`), and any `invokedynamic` site other than a lambda/method
  reference** are invisible to a static call graph. A Spring Data
  repository interface, for instance, has *no bytecode implementation at
  all* — Spring generates it at run time — so a call to one can only be
  seen reaching the interface method itself.
- **`MethodHandle`**: a `MethodHandles.Lookup#findStatic/findVirtual/
  findSpecial/findConstructor` call with a literal target, and an
  `invoke()`/`invokeExact()` call whose receiver is that same handle —
  found *earlier in the same method*, cached in a static or *final
  instance* field and read from a different method, or returned by a
  direct (`invokestatic`/`invokespecial`) helper call, with fields and
  calls chased through each other where one resolves to the other
  (DATA_FLOW D3's three slices) — gets a real call-graph edge
  (DYNAMIC_LINKAGE L2 — "MethodHandle call-graph edges" section).
  Measured honestly on real petclinic/WebGoat: this still adds zero
  edges on these two specific apps, but the chase now reaches each
  site's *true* root cause instead of stopping early — e.g. one
  petclinic field-backed receiver resolves all the way through to a
  `findVirtual` call whose own argument isn't a literal, genuinely
  unresolvable, not a missed field write. The remaining real gaps:
  non-final instance-field handle caches (WebGoat's JRuby dependency
  reassigns its cache fields, so the deliberately final-only field
  slice correctly declines them), and calls into JDK-internal classes
  outside the analyzed hierarchy.
- **Callbacks and async execution** (`Thread`, `Executor`/
  `ExecutorService`, and `AccessController.doPrivileged` — all six
  `PrivilegedAction`/`PrivilegedExceptionAction` overloads, modeled as
  running `run()` synchronously and returning its result): every call
  site registering one is inventoried
  (ASYNC_CALLBACKS A0 — "Callback registration call sites" section),
  the concrete `Runnable`/`Callable` class or lambda/method reference
  passed directly is resolved where possible — a lambda resolves
  straight to its own implementation method (DYNAMIC_LINKAGE L0/L1)
  (ASYNC_CALLBACKS A1 — "Callback identities resolved" section), and a
  real call-graph edge connects the registration site to that method
  (ASYNC_CALLBACKS A2 — "Callback call-graph edges" section). Nothing
  beyond the registration→callback edge itself is modeled
  (`CompletionStage`, listeners, reactive operators), and a lambda whose
  own bootstrap isn't LambdaMetafactory (e.g. `StringConcatFactory`) is
  left unresolved rather than guessed at.
- **Lambdas run in another method than the one creating them** get a `sam`
  edge only where DATA_FLOW follows the value (DYNAMIC_LINKAGE L1, above):
  not through the result of an instance call, an array element, a
  collection, a value live across an exception handler, a field written
  outside its class, a method overriding one declared outside the input,
  or past the caller and work limits — those sites are counted as gaps in the
  report header, with no edge. Resolution is context-insensitive (a
  helper's parameter gets every lambda any reachable creator passes), and a
  value handed to a JDK method is taken to run there even when the JDK
  only stores it (`Map#put`).
- **Serialization gadget chains** are only matched against a five-chain
  catalog (DESERIALIZATION Z3a: present classes, guards, JDK, site policy
  — never the data flow, never a search for new chains); otherwise not
  modeled: Java serialization
  entries reach the callbacks of the input's own serializable classes
  (DESERIALIZATION Z1), not the JDK's callbacks nor what those call on the
  objects read (the classic `HashMap` → `hashCode` chain); filters are
  evaluated only when literal and in the entry's own method (a process-wide
  `jdk.serialFilter` is not visible at all); XStream's converters for JDK
  types are not modeled, and other libraries' deserializers get no
  callback edges; classes
  whose superclass chain leaves the input at a JDK class not in the
  built-in table are not candidates. **Taint** (`-taint`, DATA_FLOW D4) is
  a pilot: deserialization entries are its only sinks, its sources are a
  small Spring MVC / JAX-RS / Servlet catalog, callers are searched in
  application classes only (8 calls deep), flows through fields, lambda
  callers and unmodeled calls are gaps (`undetermined`), control
  dependence is not tracked, and a path lists methods, not the steps
  inside each method. **`web.xml`** (a WAR's own
  deployment descriptor, and an EAR's per-module ones) and **Spring
  classic XML bean definitions** (`<bean id="..." class="...">`,
  wherever a `<beans>`-rooted `.xml` resource is found on the app's own
  classpath) are both read in a limited, safe way (FRAMEWORKS F1 — the
  report's own "web.xml" / "Spring XML bean definitions" sections):
  `<servlet>`/`<servlet-mapping>`/`<filter>`/`<filter-mapping>`/
  `<listener>`/`<context-param>` for the former, each bean's own
  identity/class/scope/primary for the latter (`<property>`/
  `<constructor-arg>` wiring is not modeled), with every other element
  named as unmodeled rather than silently dropped. A bean name claimed
  by both an annotation and an XML `<bean>` with a *different* class is
  flagged as a conflict (the report's own "Configuration conflicts"
  section) — detected, never resolved: Spring's own real
  bean-definition-override behavior is configurable and not modeled.
  No call-graph edges come from either reader — it is inventory, not
  wiring, same reasoning as ServiceLoader's own declarations below.
  **DI**: every `@Autowired` field/constructor injection point is
  resolved against the known Spring bean names by type, then
  `@Qualifier`, then `@Primary` (FRAMEWORKS F2 — the report's own
  "Dependency injection" section), against annotation-declared beans
  (`@Component`-family stereotypes) *and* `@Bean` factory methods on
  `@Configuration` classes — a factory method's bean FQCN is its own
  declared return type, its name is `@Bean("...")`/`value()`/`name()`
  or (Spring's own default) the method name — and, now, every
  application interface that is a (transitive) subtype of Spring
  Data's own `Repository` marker interface, reported as a bean whose
  FQCN is the interface itself (the same real proxy Spring Data's own
  container generates implements exactly that interface, nothing
  else). A framework-provided bean (`JdbcTemplate`, `DataSource`,
  ...) is still not a known candidate, so those injection points are
  correctly reported unresolved rather than guessed. A `@Value`-
  annotated field or constructor parameter (a config value, e.g.
  `@Value("${spring.datasource.driver-class-name}")`) is recognized as
  its own, separate case: never run through the bean search at all, so
  it reports as "config value: `${...}`" rather than a misleading
  "unresolved: no known bean" — Spring resolves it against property
  sources, not the bean catalog, so there genuinely is no bean to find
  (confirmed on petclinic: its own 8 injection points, all Spring Data
  repository interfaces, resolve 8/8, no `@Value` anywhere in its own
  source; WebGoat: **45** real injection points (**43** — the
  implicit-sole-constructor rule only applies to a class Spring
  actually container-manages, a stereotype or `@Configuration` class,
  not every class with exactly one constructor — plus **2** `@Value`
  fields that weren't scanned at all before `@Value` field recognition
  was added, since Spring treats `@Value` as its own injection
  annotation, independent of `@Autowired`), **37** genuinely resolved, with all 8
  remaining honestly explained: **3** real `@Value` config values
  (`DatabaseConfiguration`'s own `String driverClassName` constructor
  parameter, plus `LessonMenuService`'s own `excludeCategories`/
  `excludeLessons` fields — SpEL expressions splitting a property into
  a `List<String>`, found to be `@Value`, not `List<Bean>`, once field
  recognition existed to catch them), **4** real `List<Bean>`
  collection injections (out of scope, see below), and **1** real
  framework-provided bean (`UserService`'s own `JdbcTemplate`) —
  nothing left unaccounted for). A factory method whose own
  construction is ambiguous or programmatic (anything beyond its
  declared return type) is not resolved further. No AOP/pointcut
  model, and no call-graph edges from a resolved injection or a
  `@Value` config value — neither is a call. **Web.xml roots**: a
  `web.xml`-declared servlet/filter/listener's own lifecycle method(s)
  (`doGet`, `doFilter`, `contextInitialized`, ... — only the
  well-known container-callback names, never every public method)
  become real `-roots profile` entry points (FRAMEWORKS F2 — the
  report's own "web.xml-declared roots" section); previously a plain
  servlet with no annotation contributed nothing under `-roots
  profile`.
  **Entry point catalog**: a single labeled inventory combining the
  HTTP routes above with the web.xml lifecycle methods, each tagged
  with its own basis (`http-mapping`/`servlet`/`filter`/`listener`)
  and activation condition — cross-referenced against the `@Profile`
  conditions above, never a second condition model (FRAMEWORKS F2 —
  the report's own "Entry point catalog" section). Purely a labeling
  pass over facts already computed elsewhere: it adds no new
  `-roots profile` entry point and no new call-graph edge, since an
  external parser/dispatcher's own call into application code is
  never fabricated, only shown to exist ahead of the listed method.
  **Proxy/interceptor identity**: every application class implementing
  `java.lang.reflect.InvocationHandler` is listed with its own
  candidate target field(s) — a `final` instance field of a reference
  type (FRAMEWORKS F3's own first slice — the report's own "Proxy/
  interceptor identity" section). Identity only, no call-graph edge:
  binding a proxy call to its real dispatch needs `Method.invoke`
  receiver reconstruction (REFLECTION R2, not built), and reporting
  one without it would fabricate a dispatch this slice cannot prove.
  CGLIB-generated proxy subclasses are never listed — they exist only
  as bytecode the JVM generates at runtime, never as a `.class` file
  in the analyzed jars. Confirmed on real petclinic/WebGoat: WebGoat
  has exactly one (`LessonConnectionInvocationHandler`, with its own
  `targetConnection: java.sql.Connection` field); petclinic has none.
  **ServiceLoader** provider *declarations*
  are inventoried (`META-INF/services/*`, SERVICE_LOADING S0 — see the
  report's own "ServiceLoader provider registrations" section), every
  `ServiceLoader.load(...)`/`loadInstalled(...)` call site is found with
  its service-interface argument resolved where possible (SERVICE_LOADING
  S1 — "ServiceLoader.load call sites" section), and each resolved
  interface's registered providers can add model-derived call-graph edges
  to their public no-argument constructors at a linked `Iterator.next()` or
  `Provider.get()` call. `Iterator.hasNext()` is recognized but does not
  materialize an instance. The link follows bounded receiver provenance
  exact `invokestatic`/`invokespecial` helper return summaries, including a
  returned formal parameter substituted from that call site's argument, and
  D3 field summaries (static fields and final instance fields written by `this`
  in a constructor) to the
  `load()` site. A subsequent call through that exact service interface can
  add descriptor-exact edges to the registered provider implementations;
  existing CHA edges remain as their broader conservative result. The
  `load()` call alone stays lazy and adds no constructor edge. The report's
  provider table lists materializable candidates; it
  does not claim every listed provider was instantiated. Unsupported
  receiver flows produce an explicit incomplete-model warning rather than
  a guessed edge. Callback-based stream pipelines and unsupported terminal flows, receiver
  flows through arbitrary mutable fields, parameter values not resolvable from
  the analyzed callers (DATA_FLOW D3 now resolves a `Class`/`ClassLoader`
  parameter of the method calling `load` from its callers, up to 3 levels and
  16 call sites, for statically exact methods only; a consumption through one
  helper call uses that call's own argument; a `final` field set from a
  constructor parameter — e.g. an anonymous class's captured `val$x` — is
  resolved through that constructor's callers),
  virtual helper returns and other unresolved results,
  and caller-supplied loaders remain unsupported or incomplete, except a
  `load(Class, ClassLoader)` loader recognized in the calling method itself —
  `Thread#getContextClassLoader()`, `ClassLoader.getSystemClassLoader()`,
  `null`, or any `Class#getClassLoader()` — which is treated as the modeled
  class path (one application class loader over every input jar; not true in
  servlet containers or nested-jar launchers) and shown in the load-site
  table. Loaders passed as parameters or held in fields keep the gap.
  JPMS (SERVICE_LOADING S3): `-jpms-scenario` picks the launch the
  ServiceLoader model assumes. The default `classpath` matches the JVM's
  class-path behavior (every `module-info.class` ignored). The report's
  "JPMS module descriptors" section lists, per input jar, how it is treated on
  a module path (explicit module from `module-info.class`, including a
  multi-release `META-INF/versions/N/` descriptor selected by `-target-java`,
  or an automatic module), every `provides` pair (provider found,
  `public static provider()` present, also in `META-INF/services`), and
  providers declared by only one of the two mechanisms. With
  `-jpms-scenario module-path` every input jar is on the module path: an
  explicit module's providers are its `provides` entries (its
  `META-INF/services` is ignored), a `provider()` method is the model target
  instead of a constructor, automatic modules keep their `META-INF/services`
  providers (constructor only), and a caller in an explicit module without
  `uses` for the service gets no provider edge and an incomplete-model
  warning. A `requires` on a module absent from the input is a warning (the
  JVM would refuse to start). No exact dispatch edge follows a
  `provider()`-created object (its runtime type is not tracked; CHA edges
  remain). The exact supported signatures are listed in `docs/SUPPORT.MD`
  ("ServiceLoader: exact supported signatures and scenarios"); the model is
  checked against a real-JVM oracle over the fixtures
  (`scripts/serviceloader-oracle.sh`, `testdata/oracle/`).
  `-jpms-scenario mixed` puts only the `-jpms-module-path` jars on
  the module path: class-path jars keep their `META-INF/services` providers
  (visible to `ServiceLoader.load(Class)` alongside module-path providers), a
  class-path jar's `module-info.class` is ignored, the `uses` rule applies only
  to callers in explicit modules on the module path, and a `requires` satisfied
  only by a class-path jar is a warning (named modules cannot read the unnamed
  module). `load(ModuleLayer.boot(), S)` — recognized when the argument is
  provably `ModuleLayer.boot()` — sees only providers in named modules on the
  module path (none from the input in the `classpath` scenario, reported as a
  gap since JDK providers are not analyzed); any other `ModuleLayer` argument
  keeps every candidate provider with a visibility gap. Full module resolution,
  qualified exports, a class-path main class not resolving module-path modules
  without `--add-modules`, and layers other than the boot layer are not
  modeled.
- **Shaded/relocated dependencies**: a library embedded with its own
  packages and its `pom.properties` is detected and assessed separately
  (CVE:F6a), and so is a relocated one whose relocation is explicit — in
  the jar's own maven-shade configuration or in `-embedded-components`
  (CVE:F7a). An unknown relocation is not discovered (`classes_absent`).
  A library embedded without its `pom.properties` is found only when the
  jar's maven-shade `artifactSet` includes, a hash-tied SBOM or a
  declaration names it (CVE:F6b); a shade configuration without includes
  names nothing. A copy of another version than declared is resolved by
  comparing at most 8 published versions — several indistinguishable
  ones stay a candidate set, none is `version_unknown`. Diff-mapping verification (CVE:F5b) and LLM
  grounding still look methods up under their original names.
- **CHA over-approximates**: a virtual/interface call is treated as
  reaching every subtype that overrides it (as of JVM_GRAPH G2, matched
  by exact descriptor, not just name — so two unrelated overloads of
  the same name no longer get conflated), whether or not that specific
  override is ever invoked for that call site at run time. This can
  only produce false positives.
- **Reflection under-approximates**: anything invoked purely via
  `java.lang.reflect`, a `MethodHandle` lookup outside a recognized
  `LambdaMetafactory` call, or dependency injection wiring beans by
  name/type at runtime is invisible, and can produce false negatives (a
  real path exists that ariadne cannot see). One exception, new: a
  `Class.forName("literal")`/`forName("literal", bool, loader)` call
  whose class-name argument is a provable string literal (directly or
  through a local variable) resolves to the same Class fact a `.class`
  literal already gives (REFLECTION R0/R1's own first slice) — so any
  existing consumer reading a Class-typed value (e.g. `ServiceLoader.
  load(Class)`'s own argument) now also sees one reached via
  `forName`. A second slice: `Class#getMethod`/`getDeclaredMethod`
  called with a literal name and a literal `Class[]` parameter-types
  array (`getMethod("foo", String.class)` and the like) resolves to a
  known owner/name/parameter-types fact, `getMethod` and
  `getDeclaredMethod` kept distinct. A runtime-computed name, or an
  array with any non-literal/primitive-typed element (`int.class`
  compiles to a field read, not a literal, and is correctly left
  unresolved), still resolves to nothing, never guessed. This is
  value-knowledge only — no call-graph edge comes from either slice.
  A third slice, REFLECTION R2: every `java.lang.reflect.Method#invoke`
  call site whose Method receiver resolves back to one of the above
  `getMethod`/`getDeclaredMethod` facts now gets a real "may run"
  call-graph edge (`BasisReflection`) to every overload of the resolved
  method name on the real dispatch target (the invoked object's own
  concrete allocation type when provable, the `Class` the lookup was
  performed on otherwise) — the same reasoning as a profile's synthetic
  edge or `MethodHandle.invoke`'s own `BasisMethodHandle`. A receiver
  that isn't one of those resolved facts (a parameter, a field read, an
  unresolved call result) is correctly left with no edge, not guessed.
  A fourth slice, REFLECTION R2's own other half: `Constructor#
  newInstance` whose Constructor receiver resolves back to a
  `getConstructor`/`getDeclaredConstructor` fact gets a real
  call-graph edge too — but exact, not "every overload": a constructor
  is never inherited, so the resolved parameter types pick exactly one
  `<init>` on the resolved class, with a related class-initialization
  edge alongside it if that class has a `<clinit>` (the same pairing
  a `new` instruction itself already gets). `getDeclared*`'s own
  access-check bypass/`setAccessible` (R3), a Method/Constructor
  receiver that arrives as a method parameter rather than a field or a
  helper-method return value, and any runtime-computed class/method
  name remain unmodeled, so the rest of this limitation still stands.
- **Six profiles exist: `spring`, `servlet`, `jakarta`, `micronaut`,
  `quarkus` and `generic`** — every framework PLAN.md section 4 originally
  named for Phase 7, plus the `spring`/`generic` pair from the base plan.
  The profile system itself (auto-detection and merging, `-roots profile`,
  synthetic edges, the conditional jackson-databind rule,
  `roots.overrides`/`roots.annotations`/`roots.parameter_annotations`,
  including CDI's `@Observes`) is implemented (see step 7 above); there is
  no `--max-dep-depth`. See [PLAN.md](PLAN.md)'s Phase 7 for what's left.
- **An EAR's own top-level modules are never read as app code**, only as
  dependency jars — there is no reliable signal in a zip listing alone to
  tell an EJB module's jar apart from a shared library jar without
  parsing `META-INF/application.xml`, which is itself optional. A
  multi-WAR EAR has only been verified with one WAR module; each
  additional module's classes/jars are merged in, but no multi-module
  EAR has actually been run through this.
- **LLM-assisted vulnerable-method mapping is opt-in and unverified.**
  Without `-llm` the third `internal/vulnmap` source is off. With it, the
  methods are a machine's suggestion (checked to exist, not to be the
  vulnerable ones); see step 6 above.

See [docs/SUPPORT.MD](docs/SUPPORT.MD) for a mechanism-by-mechanism status table and
[RECOMMENDATION](RECOMMENDATION/README.MD) for plans to extend beyond the limitations
above — this list is not independently maintained from those two.


