
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
   repository root; `ARIADNE_MAVEN_LOCAL=~/.m2/repository` reads files a
   local Maven repository downloaded from a remote one before the network —
   Central rate-limits automated clients (HTTP 429, met 2026-11-18): run
   online once, then `-offline`; after a 429 nothing more is sent to it in
   that run, and a failed URL is not asked twice). When its classes are there (same bytes, or the same
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
second mode (`-callgraph-mode rta`, P2, below) and receiver narrowing an
experimental third (`-callgraph-mode pta`, P3, below), never a silent
change of the default. The graph's settings — mode, builder version, skip rule,
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
4 correct changes there; WebGoat's too since its prepared input includes its
Boot launcher (`StartWebGoat`, 2026-11-13) and `DispatcherServlet` is
registered. An input without the class that enables Boot auto-configuration
registers nothing — check the `servlet dispatch` log line. Use RTA as a second opinion next to a CHA run
(`-graph-compare`) and check the `servlet dispatch` log line. Checked
against the real JVM on `testdata/rta` (`scripts/rta-oracle.sh`). Checked against the real JVM on
`testdata/callgraph` (`scripts/callgraph-oracle.sh`: overloads, a default
method, generic and covariant bridges, an inherited body, a super call).

**Receiver narrowing (GRAPH_PRECISION P3 `p3-1`, experimental opt-in
`-callgraph-mode pta`, `internal/pta`):** RTA first, then for every CHA
call site in a reachable method the analysis tries to prove which classes
the receiver can be, over the shared value domain (DATA_FLOW D2/D3) — a
`new` in the same method (the allocation key is the `new` site), a small
join of such values, a static factory's result, a `final` field's value
(only `<clinit>` or the constructors write it), a parameter of a static or
private method or a constructor whose every caller is in the input (not a
root, no reflection/model edge into it). No call-site context; an array
element, a virtual call's result, a non-final field, a field written from
a constructor parameter, an unresolvable class or more than 8 candidate
classes are no proof, and the site keeps every target RTA kept. With a
proof, a CHA edge keeps the site only if a proven class selects its
target (JVM method selection); direct edges are untouched. Real-JVM
oracle `testdata/pta` (`scripts/pta-oracle.sh`, `TestPTAOracle`): two
receivers, a final field, a factory, a parameter and a two-allocation
factory are narrowed, the array, non-final field and constructor-parameter
field are not, and every call the JVM made stays an edge. Measured
(`-roots profile` / `all`): petclinic 7 062 / 7 991 more CHA edges removed
than RTA (3.4 % of the reachable ones), WebGoat 9 555 / 12 302, JPetStore
292 / 299; **no verdict changed** (two WebGoat paths moved to another real
route); cost 2.7–3.7× RTA's time (petclinic 22 s → 61 s) and +30–45 %
memory. The log line "receiver narrowing" and v2 `callgraph.pta` give the
counts. It stays experimental: it removes real over-approximation but has
not yet changed a result, at a large cost.

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

**log4j 1.2 appenders the packaged configuration leaves out
(`internal/log4jconfig`, `l1`).** log4j 1.2 creates its appenders
reflectively from `log4j.properties`/`log4j.xml`, so every call log4j makes
on an appender (`Appender#doAppend`, `OptionHandler#activateOptions`,
`AppenderSkeleton`'s own `this.append`) is expanded by CHA to all 18 appenders
log4j ships — `JDBCAppender`'s SQL, `JMSAppender`'s JNDI lookup, the socket
appenders. When log4j 1.2 itself is in the input (not the log4j-over-slf4j
bridge) and a packaged configuration is found, ariadne reads every
`log4j.properties`/`log4j.xml` at a class-path root of the application and of
its dependency jars (`log4j.appender.<name>=<class>`, `<appender class="…">`)
and the application classes (a call, `new` or static read of an appender);
log4j's own appenders that neither names are not CHA/RTA targets (direct calls
are untouched; an application's own appender classes are never left out). The
model leaves every appender on when the configuration may come from
elsewhere: no packaged file (`-Dlog4j.configuration`, an external file), a
`${…}` class name, an unparsable XML file, a `web.xml` `log4jConfigLocation`
other than a `classpath:` root file it read, or application code calling
`PropertyConfigurator`/`DOMConfigurator` `configure*` or Spring's
`Log4jConfigurer#initLogging`. Log line: `log4j 1.2 appenders (l1-1; config
…): on …; off …` (or why every appender is kept); the graph's skip rule names
those left out. Control app: `scripts/prepare-log4j1-demo.sh` (the same
classes with a console-only and a `JDBCAppender` configuration, run on the
JVM: CVE-2022-23305 `REACHABLE` only where `JDBCAppender` runs). On
JPetStore (CHA): CVE-2021-4104 and CVE-2022-23305 `REACHABLE` →
`NOT_REACHABLE`, CVE-2022-41853 `REACHABLE` → `FRAMEWORK`.

**logback appenders the configuration leaves out (`internal/logbackconfig`,
`l2`).** The same for logback: CHA sends `Appender#doAppend` and the base
classes' `this.append` to all 20 appenders logback-core/classic ship
(`DBAppender`, `SMTPAppender` with its evaluator, `RollingFileAppender`'s
rollover…). An appender stays a target when a packaged `logback.xml`,
`logback-test.xml`, `logback-spring.xml` or `logback-test-spring.xml` names it
(`<appender class>` anywhere, `<include resource>` followed into the input's
jars — Spring Boot's `base.xml` includes its file appender — short names
through `<import>`), when application code references it, or — with no
packaged file and Spring Boot's logging system in the input — when Boot's
default configuration creates it: `ConsoleAppender`, plus
`RollingFileAppender` if a packaged `application*` file sets
`logging.file.name`/`logging.file.path` (or the older
`logging.file`/`logging.path`; a placeholder counts). Every appender stays on
with no packaged file and no Spring Boot, with Boot's `logging.config` set, a
Groovy configuration, a `Configurator` service, a `${…}` or unimported class,
an include by file or URL, or application code running Joran. Log line:
`logback appenders (l2-2; …): on …; off …`. Since builder `5`, a body an
abstract class declares (`SMTPAppenderBase#append`, `DBAppenderBase#append`)
is not a CHA target when every concrete subtype that inherits it is excluded
by such a model. Control app: `scripts/prepare-logback-demo.sh` (Boot 2.4.5
defaults, `logging.file.name`, a `logback-spring.xml` with `SMTPAppender` —
the model's appenders match the JVM's in all three). On petclinic and
WebGoat (Boot defaults: console only) the appender paths are gone, but only
CVE-2023-6378's logback-core instance changes (`FRAMEWORK` → `NOT_REACHABLE`):
CVE-2021-42550 and CVE-2024-12798 stayed `FRAMEWORK` through CHA over logback's
`LifeCycle#start`/`#stop` (e.g. `StatusListenerConfigHelper` → every
`LifeCycle`).

**logback's other Joran components (`l2-2`).** The same configuration now also
decides the rest of what Joran creates: a *Joran-only component* is a concrete
logback class (`ch.qos.logback.*`) implementing `LifeCycle` — not an appender,
not a pattern converter (`PatternLayout`/`FileNamePattern` create those from
their own maps) — that no code in the input creates with `new`: filters,
turbo filters, evaluators, connection sources, rolling/triggering policies,
layouts, receivers, discriminators, status listeners. It is a CHA/RTA target
only when a class attribute anywhere in the packaged configuration names it,
when it is Joran's default class of an element present without one
(`<layout>` → `PatternLayout`, `<encoder>` → `PatternLayoutEncoder`,
`<evaluator>` → `JaninoEventEvaluator`, `<discriminator>` →
`MDCBasedDiscriminator`), when application code references it, or when Joran
could create it from a one-argument setter's parameter type. Log line:
`logback Joran-only components (l2-2): N on, M off (…)`. On petclinic and
WebGoat (Boot defaults) 30 such components are off, and CVE-2024-12798
(`JaninoEventEvaluator`) is `NOT_REACHABLE` instead of `FRAMEWORK`; the control
app's `<evaluator>`-without-class variant keeps it (`FRAMEWORK`), as the JVM
does (`Assuming default type [JaninoEventEvaluator]`). CVE-2021-42550 stays
`FRAMEWORK` through `ContextJNDISelector`, which logback creates only under
`-Dlogback.ContextSelector=JNDI` — a JVM setting the input does not show.
Builder `6`: an excluded class's body is still a target when a kept subtype
inherits it (`ConsoleAppender` runs `OutputStreamAppender#append`; builder
`5` dropped such bodies).

**A second bounded exception: JDK/javax service-provider interfaces.**
A provider-based API — Bean Validation, JDBC, JCache — calls its provider
through an interface in a skipped package (`javax.validation.spi.
ValidationProvider#buildValidatorFactory`, `java.sql.Driver#connect`), and
hands out further objects through other interfaces of the same API
(`Configuration` → `ValidatorFactory` → `Validator`, `Connection` →
`Statement` → `ResultSet`, `CacheManager` → `Cache`). The
`service-provider:dispatch` model (`internal/spidispatch`, version `p3`, a
B3 registry model with basis `model`) expands a call whose static receiver
is such an interface:

- a **registered service** (a `META-INF/services` file in the input for a
  skipped interface) → to its registered providers present in the input;
- a **product** of a service — an interface of the same API family (the
  service's package without `.spi`) reached from the service through return
  types of its methods, transitively — → to the implementations of that
  interface found in the **jars of the registered providers** only.
- a **facade** (since `p2`) — a static JDK method that calls the service
  itself: `java.sql.DriverManager#getConnection` (every overload) → the
  registered drivers' `Driver#connect`; its return type makes `Connection`
  a product as above. Without it a driver reached only through
  `DriverManager` (Spring `DriverManagerDataSource`, MyBatis
  `UnpooledDataSource`, Hibernate `DriverManagerConnectionCreator`) was
  not reached at all (measured by `testdata/jdbc`, `scripts/jdbc-oracle.sh`).
  Since `p3` a product's implementations include those of its JDK
  subinterfaces, as for `jdbc:objects` `j2`.

Each target is the method the JVM selects on that class. An interface the
service does not hand out (a `ServletContainerInitializer` returns
nothing) has no products, so the servlet API stays with the servlet model
above. Not modeled: implementations outside the providers' jars,
providers registered only through a module `provides`, and which provider
a given call site really receives (the edges are "may run", like CHA's;
a never-created class in a provider jar is a target too). The log line
`service-provider dispatch (service-provider:dispatch p3): …` counts the
services and call signatures. Found by a runtime trace: on petclinic the
JVM ran Hibernate Validator's `HibernateValidator#buildValidatorFactory`
while the graph stopped at the `ValidationProvider` call
(CVE-2025-35036 was `NOT_REACHABLE`). Checked against the real JVM on
`testdata/spidispatch` (`scripts/spidispatch-oracle.sh`: a `javax.*` SPI
in its own jar, a registered provider and a product, an unregistered
provider and a product implementation in another jar — every observed body
reached, the other jar never a target).

**One fixed point for models and gates (ROADMAP B3, coordinator `b3-2`).**
The registered library models (ServiceLoader, servlet container dispatch,
service-provider dispatch) and the creator gates (SAM calls, servlet
objects, JDBC objects, framework proxies) can unlock each other: a gate
edge can make a modeled call site reachable, a model edge can make a
creator reachable. `internal/analysis.Run` iterates both kinds together —
every gate, then one model pass — until an iteration adds nothing, within
a budget it reports (the remaining participants are named when the budget
ends). The order of the participants does not change the result. Log line:
`models and gates (ROADMAP B3 fixed point): N iterations of at most M; …
model edges; gate edges dynamic-linkage:sam-calls N, …`; snapshots record
`analysis.coordinator`. Before `b3-2` the CLI ran the gates, then the
models to their own fixed point, then the gates once more, which misses an
edge behind a model → gate → model chain (`TestRun_stepsAndModelsShareOneFixedPoint`).
On petclinic, WebGoat, JPetStore and the log4j/logback control apps the
joint fixed point gives the same graph and results (1–5 iterations): those
inputs had no such chain.

**Choosing the participants (coordinator `b3-3`).** A profile can turn
models and gates off by ID — data only, no code:

```yaml
models:
  disable:
    - servlet-container:dispatch
    - servlet-container:objects
```

The IDs are `jdk:service-loader-classpath-provider`,
`jdk:service-loader-consumption`, `jdk:service-loader-provider-dispatch`,
`servlet-container:dispatch`, `service-provider:dispatch` (models) and
`dynamic-linkage:sam-calls`, `servlet-container:objects`, `jdbc:objects`,
`jpa:objects`, `jdbc:drivers`, `framework:proxies`, `spring:events`,
`spring:aop`, `async-callbacks:returned-tasks`, `async-callbacks:listeners`,
`async-callbacks:reactive`, `async-callbacks:completion`,
`async-callbacks:rejection` (gates). An ID the analyzer does not know stops the run
with the list of known ones; profiles merged by auto-detection disable the
union of their lists; the embedded profiles disable nothing. A disabled
model is removed from the registry and a disabled gate adds no edge. The
participants that ran are recorded as `id@version` in the report's call
graph line, in v2 `callgraph.participants` (with
`disabled_participants`, so the `analysis_id` differs), in `-graph-out`
settings (a `-graph-compare` against another selection says so) and in a
snapshot's configuration (`models`, `models-disabled`). Disabling the two
servlet participants on petclinic removes 5 918 edges and turns 13
`FRAMEWORK` into `NOT_REACHABLE` — a way to see what a model contributes,
not a recommended setting.

**Framework proxies (`framework:proxies`, `internal/proxycatalog`, version `x3`).**
Some framework objects exist only as JDK dynamic proxies: a MyBatis mapper
interface (`CategoryMapper`) has no implementation in the input — MyBatis
hands out a proxy whose `MapperProxy#invoke` runs the SQL — and neither has a
Spring Data repository interface (Spring AOP's `JdkDynamicAopProxy`) nor the
`EntityManager`/`Query`/`EntityManagerFactory` Spring ORM injects. A general
rule cannot tell which interfaces a `Proxy.newProxyInstance` call implements
(FRAMEWORKS F3 slice 2 measured: bounded at none of the sites behind the
traced relations), so this model takes them from a catalog: calls on a
**MyBatis mapper** (an interface named as the `namespace` of a mapper XML in
the input, annotated `@Mapper`, or in a package a `@MapperScan`, a
`<mybatis:scan base-package>` or a `MapperScannerConfigurer` in the input
scans) → `MapperProxy#invoke`; `SqlSessionTemplate`'s calls on its
`SqlSession` proxy → `SqlSessionTemplate$SqlSessionInterceptor#invoke`; a
**Spring Data repository** interface → `JdkDynamicAopProxy#invoke`;
`EntityManager` → Spring ORM's shared and extended EntityManager handlers,
`Query`/`TypedQuery`/`StoredProcedureQuery` → its deferred-query handler,
`EntityManagerFactory` → the managed factory handler (javax and jakarta;
all four handlers reached exactly as Spring ORM 5.3.6 runs them on a real
JVM, `testdata/jpa`; the provider behind them is reached by `jpa:objects`
below).
Since `x2` a Spring Data repository call also reaches the **base
repository** body (`spring-data:base-repository`): Spring Data's proxy
runs it through `RepositoryMethodInvoker` → `Method#invoke`, which no
literal resolves (measured by `testdata/springdata`). A base class is a
concrete `Repository` subtype in the input (`SimpleJpaRepository`, a
store's own), gated by a reachable method that `new`s it or belongs to a
class naming it as a literal (a factory's `getRepositoryBaseClass`); the
method is matched as Spring Data does — exact descriptor, else name and
assignable parameters (`void save(Owner)` → `save(Object)`); query methods
(an annotation `*.Query`, or no base-class method of that name) and default
methods get no such edge.
Since `x3` a call on a method Spring runs in a transaction reaches
`TransactionInterceptor#invoke` (`spring-tx:transactional`): `@Transactional`
(Spring's, or JTA's javax/jakarta one) on the method, its class, or a
supertype declaring it or that supertype itself; public, non-static,
non-final methods; the named method on the call's owner or on any concrete
subtype; not a call from inside the same class (self-invocation bypasses
the proxy). Gate: a reachable method that `new`s the interceptor
(`@EnableTransactionManagement`'s configuration) or names the class (the
`<tx:annotation-driven>` parser). Real-JVM oracle `testdata/tx`
(`scripts/tx-oracle.sh`, `TestTxOracle`): class-level, method-level and
interface-proxied services and a rollback — the graph missed the
interceptor, `TransactionAspectSupport` and the transaction manager's
`getTransaction`/`commit`/`rollback` (9 bodies), now reaches all of them.
Reference apps: petclinic 20 call sites (its `@Transactional` Spring Data
repositories), WebGoat 10, JPetStore 4 (`AccountService`, `OrderService`);
the edges make spring-tx's internals reachable from application code, and
CHA inside them reaches far more than an application runs (JPetStore:
`TransactionSynchronization#beforeCommit` → the `@TransactionalEventListener`
adapter it never registers → SpEL). So the `spring` profile also lists
`org.springframework:spring-tx` as a framework package (like spring-beans,
core, expression and aop): a path through it is `FRAMEWORK`, to review by
hand. With both, every verdict on petclinic, WebGoat and JPetStore is what
it was before `x3`; only JPetStore's `-roots all` `FRAMEWORK` paths and
model-gap counts change (they now go through the real interceptor). Without
spring-tx in the list, CVE-2026-41850 (JPetStore `-roots all`) and, with
RTA, CVE-2022-41853 became `REACHABLE` through those internals.
Each edge (basis `model`) waits for the same creator gate as the JDBC
objects: a method creating the handler must be reachable from the roots.
Checked in the bytecode of mybatis 3.4.1–3.5.16, mybatis-spring
1.3.0–3.0.3, spring-aop/spring-orm 4.3.4–6.1.14. Log line `framework
proxies (framework:proxies x3): N MyBatis mappers, N Spring Data
repositories; N candidates (…), N edges added, …`; snapshots record
`framework.proxies`. Measured on runtime traces: petclinic's 13 `proxy`
relations and 6 of JPetStore's 8 are graph edges now; JPetStore's
CVE-2022-41853 (HSQLDB) is `REACHABLE` under `-roots all` through the
application's own mapper calls (it was `FRAMEWORK`, through Spring's
database shutdown only). Limits: the edge stops at the handler —
what it then calls through `Method#invoke` is reflection; a mapper
registered only in code (`Configuration#addMapper`) is not found (since
2026-11-15 a `<mybatis:scan>`/`MapperScannerConfigurer` in a WAR's
`WEB-INF/*.xml` is read too — `input.PreparedApp.WebInfBeansXMLs`; on
JPetStore it finds no mapper the mapper XML namespaces did not already
find, so nothing changed there);
`SqlSessionTemplate`'s interceptor is gated out when Spring creates the
template only reflectively (JPetStore); behind `JdkDynamicAopProxy#invoke`
CHA expands Spring AOP's interceptor chain and target sources, so some paths
through repositories are CHA over-approximations (petclinic/WebGoat: paths
changed, no verdict).

**Spring AOP advice (`spring:aop`, `internal/springaop`, version `a2`;
FRAMEWORKS F3).** A bean advised by an aspect is replaced in the container
by a proxy (JDK or CGLIB); a call through the bean reference reaches the
proxy, which runs the matching advice through
`AbstractAspectJAdvice#invokeAdviceMethodWithGivenArgs` → `Method#invoke` —
nothing a reflection model resolves, so no path reached an advice method.
The model adds an edge (basis `model`) from each call site that may invoke
an advised bean method to the advice method:
- an aspect is an application class annotated `@Aspect` that is a bean
  (related to a bean type the spring profile found — a stereotype, a
  `@Bean` product, an XML bean), or the bean an `<aop:aspect ref>` of an
  `<aop:config>` in the input's bean XML names; an `@Aspect` that is no bean
  advises nothing;
- advice is a method annotated `@Before`, `@After`, `@AfterReturning`,
  `@AfterThrowing` or `@Around` (its pointcut is the annotation's `pointcut`
  or `value` element), or the method an `<aop:before|after|after-returning|
  after-throwing|around>` names (an inline `pointcut` or a `pointcut-ref`;
  `and`/`or`/`not` allowed; the overload with the fewest parameters, as
  Spring resolves it);
- the advice's variables are bound as Spring's parameter name discovery
  does: `argNames`/`arg-names`; else the names in the class file
  (`MethodParameters` from `-parameters`, which Spring Boot's parent sets,
  or the `LocalVariableTable` from `-g`); else inference from the
  expression (`AspectJAdviceParameterNameDiscoverer`): a leading
  `JoinPoint` binds nothing, the `throwing` variable takes the one
  `Throwable` parameter, the `returning` variable the one remaining
  parameter, an annotation variable the one annotation-typed parameter,
  one `this`/`target`/`args` variable the one remaining parameter;
  anything ambiguous stays unbound;
- the pointcut language evaluated, three-valued: `execution(…)` (method
  annotation patterns `@a.B`, modifier, return, declaring-type, name and
  parameter patterns with `*`, `..`, `+`; the declaring type matches any
  type of the bean's hierarchy declaring the method), `within`,
  `@annotation`, `@within`, `@target`, `this`, `target`, `args` (types or
  bound variables; an argument whose declared type may hold an instance of
  the type matches — Spring tests the value at run time — primitives
  boxed), `bean(name*)` (the stereotype's explicit or default name, the
  `@Bean` method's name, the XML id), references to `@Pointcut` methods
  (parameterized ones bound by their own parameters), `&&`, `||`, `!`.
  Anything else (`@args`, `cflow`, annotation patterns on parameters, an
  unbound variable, …) is *unknown*: an advice whose match depends on it
  gets no edge for that method and is counted as *undecided* — never a
  guessed match;
- `@AfterReturning` with a `returning` variable runs only when the
  method's return type is compatible with the variable's type (`Object`
  takes every method, `void` included);
- a join point is a method a proxy can intercept on a bean class: an
  application class or bean type related to a bean type (not an aspect),
  its own and inherited methods except `java.lang.Object`'s, neither
  static, private, final, abstract, a constructor nor synthetic;
- a call site matches when it is a virtual or interface call of that name
  and descriptor whose owner the bean class is assignable to; a call on
  `this` from the bean's own code (self-invocation, DATA_FLOW D2's receiver
  fact) bypasses the proxy and gets no edge;
- gate: for `@Aspect` advice a reachable method creating one of Spring's
  AspectJ advice objects (`AbstractAspectJAdvice` subtypes, made by
  `ReflectiveAspectJAdvisorFactory#getAdvice`); for `<aop:config>` advice,
  whose advice objects the container creates reflectively, a reachable
  `ConfigBeanDefinitionParser#parse`.

`@AfterReturning` edges carry the condition `returned`, `@AfterThrowing`
edges `exceptional` (v2 path evidence `"condition"`, the report's arrow
`→[library_model: model, if returned]`). For `@Around` the target runs only
if the advice calls `proceed()`; the existing edge to the target is kept (an
over-approximation). Advice order (`@Order`), the `throwing` type
restriction, `<aop:advisor>` (advice beans such as `<tx:advice>`), the
AspectJ compiler or load-time weaving, `this()` on a JDK proxy (tested
against the bean class, as with Spring Boot's default CGLIB proxies) and the
JDK-proxy case of a final method reached through an interface are not
modeled. Real-JVM oracle `testdata/aop` (`scripts/aop-oracle.sh`,
`TestAOPOracle`, Spring 5.3.6 with aspectjweaver 1.9.6, compiled without
debug info so variables are bound by inference): a CGLIB-proxied service
(with a self-invocation, a final method and an `@Audited` method), a
repository behind a JDK proxy (a normal return, a thrown exception, a
`String` return), a `@Monitored` class, 17 annotated advice methods
covering every designator above and an `<aop:config>` aspect
(`fx/aop/aop.xml`, `and not`); every one of the 15 advice bodies the JVM
ran is reached (a1 reached 6, without the model none), and the four that
never run — patterns matching no method, `bean()` naming no bean, an
`@Aspect` that is no bean — get no edge; the self-invocation gets none
(the JVM ran no advice for it nor for the final method). Log line `Spring
AOP (spring:aop a2): N aspects (N beans), N advice methods (N from
<aop:config>, N unresolved), N advised bean methods, N advice × method
pairs undecided …, N self-invocations left out; N candidates, N edges
added, …`; snapshots record `spring.aop`. Control app
`scripts/prepare-aop-demo.sh` (Spring Boot 2.4.3, `spring-boot-starter-aop`,
a `@Before` advice with `args(text)` calling commons-text 1.9
`StringSubstitutor#replace`): under `-roots profile` CVE-2022-42889 is
`REACHABLE` through `main` → the advice when its pointcut names the called
`OrderService#place` (the JVM runs the advice; a1 left the `args(text)`
advice undecided and reported `NOT_REACHABLE`, without the model it was
`NOT_REACHABLE` too), stays `NOT_REACHABLE` when it names `refund`, which
nothing calls (the JVM does not run it), and is `REACHABLE` in the variant
declaring the same advice in `<aop:config>` (the JVM runs it; a1:
`NOT_REACHABLE`); under `-roots all` it is `REACHABLE` in every variant
(every application method is a root). petclinic, WebGoat and JPetStore have no `@Aspect` class and no `<aop:config>`: their v1 results are byte-identical to a1 (`-roots all`/`profile`, CHA/RTA).

**Spring application events (`spring:events`, `internal/springevents`,
version `ev1`; FRAMEWORKS F4 slice 1).** `publishEvent` goes through the
context's multicaster to every listener whose event type matches. A class
implementing `ApplicationListener` is called through that interface, and
CHA already reaches it; an annotated listener method is not — Spring wraps
it in an `ApplicationListenerMethodAdapter` whose `doInvoke` calls it
through `Method#invoke`. The model adds an edge (basis `model`) from a
publishing call site to every annotated listener method that may receive
the event:
- a publishing site is a call of `publishEvent(Object)` or
  `publishEvent(ApplicationEvent)` on `ApplicationEventPublisher` or a
  subtype, anywhere in the input (the context itself publishes
  `ContextRefreshedEvent`/`ContextClosedEvent`); a publisher forwarding its
  own parameter (a child context to its parent) is left out, and the
  adapter's own re-publishing of a listener's return value publishes the
  registered listeners' return types (arrays: the element type;
  collections and futures: unknown);
- the published type is the argument's local value: `new T(…)` is exactly
  `T`, a parameter is followed to its callers, anything else is its static
  type — or the `publishEvent` parameter's type (`Object`: every listener)
  when nothing narrower is known (counted as an *unbounded* site);
- a listener method is a method of an application class annotated
  `@EventListener` or with an annotation meta-annotated with it
  (`@TransactionalEventListener`); its event types are the annotation's
  `classes`/`value`, else its one parameter's type;
- it matches when Spring could deliver the event: the listener type is a
  supertype of the published type (a payload that is not an
  `ApplicationEvent` arrives wrapped in a `PayloadApplicationEvent`, so a
  listener on that type receives it too), or, for a static bound, also a
  subtype of it;
- the listener's class must be related (equal, subtype, supertype) to a
  bean type the spring profile found — a stereotype class, a `@Bean`
  product, an XML bean; a listener on a class that is no bean gets nothing;
- gate: a reachable method creating an `ApplicationListenerMethodAdapter`
  (the `EventListenerFactory` implementations).

The edge says the listener *may* run: a SpEL `condition`, a
`@TransactionalEventListener`'s phase (it runs only for an event published
in a transaction, after commit by default), an `@Async` executor and the
event's generic type arguments are not evaluated. Listener methods of
library classes and Spring Boot's early events (sent through the
multicaster directly, before the context exists) are out of scope.
The profile's annotation roots are unchanged: under `-roots profile` every
`@EventListener` method is still a root by itself, so the model adds
mostly paths (publisher → listener) there; a `@TransactionalEventListener`
is no annotation root and is reached only through this edge. Since the same
slice the `spring` profile also lists `@Schedules` as a root annotation:
a method with two `@Scheduled` holds only their `@Repeatable` container in
the class file, and was no root. Real-JVM oracle `testdata/events`
(`scripts/events-oracle.sh`, `TestEventsOracle`, Spring 5.3.6): before the
model the graph missed 5 of the 7 listener bodies the JVM ran (all but the
`ApplicationListener` and nothing on the `@EventListener` adapter's last
hop) and the repeated `@Scheduled` method was no root; now every one is
reached, and the two listeners the JVM never runs (an event type nothing
publishes; a class that is no bean) get no edge. Log line `Spring events
(spring:events ev1): N listener methods (N on beans), N publishing sites
(N unbounded, N forwarding left out); N candidates, N edges added, …`;
snapshots record `spring.events`. Control app `scripts/prepare-events-demo.sh`
(Spring Boot 2.4.3, a `@TransactionalEventListener` calling commons-text
1.9 `StringSubstitutor#replace`): under `-roots profile` CVE-2022-42889 was
`NOT_REACHABLE` and is now `REACHABLE` through `main` → `OrderService#place`
→ the listener when the event is published (the JVM runs the listener), and
stays `NOT_REACHABLE` in the variant publishing another type (the JVM does
not run it). petclinic, WebGoat and JPetStore have no annotated listener:
their v1 results are byte-identical (`-roots all`/`profile`, CHA/RTA).

**JDBC objects (`jdbc:objects`, `internal/jdbcdispatch`, version `j2`).**
The service-provider model reaches a driver's classes, but a real
application rarely talks to the driver directly: Hibernate or Spring's
`JdbcTemplate` call `Connection`/`PreparedStatement`/`ResultSet` methods on
a pool's proxies (Hikari's `HikariProxyConnection`,
`HikariProxyPreparedStatement`, …), which live outside the driver jars, and
the pool's proxy calls the driver through the same interfaces. Since `j1`
a virtual/interface call on `java.sql` `Connection`, `Statement`,
`PreparedStatement`, `CallableStatement`, `ResultSet`, `DatabaseMetaData`,
`ResultSetMetaData`, `ParameterMetaData`, `Array`, `Blob`, `Clob`,
`NClob`, `SQLXML`, `Savepoint`, `Wrapper` or `javax.sql.DataSource` gets
an edge (basis `model`) to the body every concrete implementation in the
input runs for it, **once a method creating that implementation (`new`)
is reachable from the roots** — the same creator gate as the servlet
objects above: the pool's proxies once its proxy factory runs, the driver's
connection once its `connect` runs. A pool or wrapper class nothing
reachable creates is not a target, nor is one created only reflectively (a
`DataSource` class Spring Boot instantiates by name gets its edges only if
some reachable method also `new`s it — on petclinic Boot creates
`HikariDataSource` reflectively, and the gate admits it through Hikari's own
`new HikariDataSource` in `HikariConnectionProvider`/`HikariJNDIFactory`,
not through Boot's creation). Since `j2` an implementation counts also
when it implements only a JDK subinterface of the receiver type
(`internal/jdkapi`: `PreparedStatement` extends `Statement`, `NClob` extends
`Clob`, every JDBC interface extends `Wrapper`, … — JDK facts the hierarchy
does not read): HSQLDB's `JDBCPreparedStatement` implements only
`PreparedStatement`, so before `j2` no call through `java.sql.Statement`
(`close`, `executeBatch`) reached it (found by `testdata/mybatis`).
`java.sql.Driver` is left to the
service-provider model. The log line `JDBC objects (jdbc:objects j2): N
candidates, N edges added, …` counts them; snapshots record
`jdbc.objects`. Measured on a petclinic trace: every one of the 22
`java.sql`/`javax.sql` relations the graph missed is an edge now, and the
`java.sql.Driver#connect` relation behind them follows
([TRACE_GAPS.MD](testdata/graphbench/TRACE_GAPS.MD)).

**Library mode (`-roots api`, `internal/libapi`, LIBRARY_MODE LB1,
2026-11-16).** A library asks a different question than an application:
can anyone using its public API reach vulnerable code in its
dependencies, and through which methods? Pass the library jar as `-app`
and its dependencies as `-deps`; with `-roots api` the roots are its
public API ([docs/RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD) §8):
the `public` methods and constructors of accessible classes (a member
class only when every enclosing class is accessible; `protected` too,
unless the class is `final`), public interface methods, every accessible
class's `<clinit>`, the public (protected in non-`final` classes)
instance methods an accessible class inherits from a dependency supertype
without overriding them (`lb2-3`: `YAMLMapper#readValue` is
`ObjectMapper#readValue`; v2 `exposed_api[].inherited_by` names the API
classes, `coverage.api.inherited_methods` counts them), and the providers the library registers in its own
`META-INF/services`, and (`lb2-5`) what it announces to Spring whatever its access — the
classes in `META-INF/spring.factories` and `META-INF/spring/*.imports`: an
autoconfiguration's constructors, `@Bean` methods (package-private ones
too), injection/lifecycle methods, member `@Configuration` classes and
`@Import`s, transitively; any other factory's constructors and public
methods (v2 `exposed_api[].announced_by` gives the file, key and steps,
`coverage.api.announced_classes`/`announcements` count them) — otherwise never `private`/package-private code, synthetic or
bridge methods. `-api-exclude` drops package prefixes the user calls
internal; with `-jpms-scenario module-path` and a `module-info`, only the
unqualified `exports` count (and its `provides` replace
`META-INF/services`). Each positive finding then lists, in v2
`findings[].exposed_api[]` and the report's "Exposed API" section, every
API method from which its mapped code is reached — with the pass
(`app_only`/`full`), an example path from that method and the path's
character (`heuristic` = depends on a CHA expansion); `coverage.api`
counts the API. Under RTA a class counts as created when it is an
accessible class (a client can construct it). **Client code (LB2,
`lb2-1`/`lb2-2`, 2026-11-17):** a call the library makes on a method a client can
implement or override — declared, with `public`/`protected` access and not
`static`/`final`, in an accessible, non-excluded library type the client
can subtype (an interface, or a non-`final` class with a public/protected
constructor) — is a `client_code` gap (site code `client`, in
`coverage.gaps`; its candidate set is empty, so it never makes a finding
`MODEL_INCOMPLETE`: client code reaches the library only through its API,
which is already the roots). No edge is invented for it. An
`exposed_api` entry is marked `client_dispatch` (report: "client
dispatch") when every path from that method passes such a site dispatched
to a body other than the called method's own, in a class whose objects
exist only as client code (the input never creates it and a client cannot
construct it — `lb2-2`) — a stand-in for the client's implementation. Calls into
dependency/JDK types are left to the other mechanisms. Measured on three real libraries (`scripts/library-measure.sh`:
commons-configuration2 2.7, jackson-dataformat-yaml 2.12.3, jjwt-jackson
0.11.2): same verdicts as `-roots all` before `lb2-3`, `exposed_api` from 6 to
778 methods per finding. With inherited methods as roots (`lb2-3`)
jackson-dataformat-yaml goes from 1 to 17 of 26 advisories `REACHABLE`
(databind deserialization through the inherited `readValue`/`readValueAs`);
**receiver specialization (`lb2-4`):** each inherited root is a copy of
the dependency method per inheriting API class, and in that copy a call
on the method's own unchanged `this` (local-value analysis of its
bytecode) goes only to the bodies the API class and its subtypes select;
a call on `this` to another method the class inherits goes to that
method's copy for the same class, transitively. Other calls (on fields,
parameters, results) dispatch as before. On jackson-dataformat-yaml 682
copies drop 1007 CHA edges (777 under RTA): `JsonFactory#createJsonParser`
on a `YAMLFactory` no longer reaches the JSON parser, and `exposed_api`
of CVE-2025-49128 shrinks from 101 to 87 entries (CVE-2026-68494 and
CVE-2026-89425: 61 → 53), verdicts unchanged — what stays comes through
fields and parameters or is the mapped code itself (`JsonLocation#_appendSourceDesc`
in a YAML exception's message). The field path `ObjectMapper#createParser`
→ `_jsonFactory.createParser` → JSON parser is real, not an
approximation: `new YAMLMapper((YAMLFactory) null)` makes `ObjectMapper`
store a `MappingJsonFactory`, and the JVM then returns a JSON parser
(`scripts/YAMLMapperFactoryProbe.java`, 2026-11-18); a field-type rule
would have to keep it. `coverage.api.receiver_specialization`
counts copies, `this` sites, dropped edges and unanalyzed methods (their
edges kept). Log lines
`library API (lb2-8): …, N methods (M inherited from dependency
supertypes, K more from C framework-announced classes in A
announcements), …, N client-code call sites (M to abstract methods)`, `receiver specialization (LIBRARY_MODE lb2-4): …`,
`client-code dispatch (LIBRARY_MODE LB2): …` and `roots: api
(…)`. Verified on `testdata/library` (cases in
[EXPECTED.MD](testdata/library/EXPECTED.MD), written before the
implementation; real-JVM oracle `scripts/library-oracle.sh`,
`TestLibraryMode`, `internal/libapi` `TestRoots`): `-roots all` lists nine
library methods reaching CVE-2022-42889's `StringSubstitutor#replace`,
three of which no client can call; `-roots api` exposes exactly the 5 (CHA)
/ 4 (RTA) / 3 (`-api-exclude lib.internal`) API methods expected, and under
RTA only methods the JVM ran. Its three client-code sites are reported;
under CHA only `Pipeline#run` (case (d), where the JVM runs the client's
listener) is `client_dispatch`. Case (i), `StrictExpander` overriding the
protected `substitute` every inherited `replace` calls on `this`, is in
`inherited_by` of the `replace` overloads for CVE-2022-42889 (they are its
mapped code) but not for CVE-2025-48924, whose lookups only
`StringSubstitutor#substitute` reaches — as the JVM shows
(`observed-strict.txt`, `internal/libapi` `TestSpecialize`). Framework
announcements (`lb2-5`) are verified on `testdata/library-autoconfig`, a
control starter whose vulnerable code only Spring can reach (package-private
`@Bean` methods, a nested configuration class, an `@Import`, a
`spring.factories` listener, an `AutoConfiguration.imports` entry, a
never-announced configuration; [EXPECTED.MD](testdata/library-autoconfig/EXPECTED.MD),
real-JVM oracle `scripts/library-autoconfig-oracle.sh` on Spring Boot 2.4.5,
`TestLibraryAutoconfig`, `TestAnnouncedRoots`): before, both advisories were
`NOT_REACHABLE` although the JVM runs `replace`; now `REACHABLE`, with exactly
the five announced entries exposed and the unannounced one not. On a real
starter (mybatis-spring-boot-autoconfigure 2.1.2, `scripts/library-measure.sh`)
9 classes are entries, verdicts and `exposed_api` are unchanged (its
autoconfigurations are public already) and the entries gain `announced_by`.
**CDI bean archives (`lb2-6`, 2026-11-18):** a library jar with
`META-INF/beans.xml` (or only a Jandex index, as Quarkus reads it) is a
bean archive; its beans — every class with a bean-defining annotation
(`@ApplicationScoped`, `@Dependent`, `@Singleton`, …) in `annotated`
mode, every concrete class with a no-argument or `@Inject` constructor in
`all` mode (also an empty or pre-4.0 beans.xml) — are announcements, and
their constructors, `@Produces`/`@Inject`/`@PostConstruct`/`@PreDestroy`/
interceptor methods and `@Observes`/`@Disposes` methods (any access) and
public methods are entries, with `announced_by` naming the archive, mode
and annotation. Verified on `testdata/library-cdi`
([EXPECTED.MD](testdata/library-cdi/EXPECTED.MD), real CDI container
oracle `scripts/library-cdi-oracle.sh` on Weld SE 5.1.2, `TestLibraryCDI`,
`TestCDIAnnouncements`, `TestCDIRoots`): before, both advisories were
`NOT_REACHABLE` although Weld runs `replace` from a package-private
producer, observer, `@PostConstruct` and `@Inject` method; now `REACHABLE`
with exactly those four exposed, and a class without a bean-defining
annotation never. On real jars (Quarkus 3.9.5 dependencies)
smallrye-context-propagation 2.1.0 has one bean (a public interceptor:
no new roots) and mutiny 2.5.8 a Jandex index but no bean; verdicts
unchanged. **Micronaut bean definitions (`lb2-9`, 2026-11-20):** a library
jar that lists bean definitions for Micronaut
(`META-INF/micronaut/io.micronaut.inject.BeanDefinitionReference/*`, 3.x;
`META-INF/services/io.micronaut.inject.BeanDefinitionReference`, 1.x/2.x)
announces them: Micronaut's generated `$X$Definition` classes
(package-private, synthetic) call the bean's constructor, factory,
`@Inject`, `@PostConstruct` and — through `$Exec#dispatch` —
`@EventListener` methods directly, whatever their access, so every library
method a definition's code calls directly is an entry (`announced_by` names
the listing entry), with the bean's public instance methods; the
definitions' own superclasses are not. Private bean methods, which
Micronaut calls through reflection, are not found. Verified on
`testdata/library-micronaut` ([EXPECTED.MD](testdata/library-micronaut/EXPECTED.MD),
Micronaut 3.9.7, JVM oracle `scripts/library-micronaut-oracle.sh`,
`TestLibraryMicronaut`, `TestMicronautRoots`): before, both advisories were
`NOT_REACHABLE` although Micronaut runs `replace` from four package-private
bean methods; now `REACHABLE` with exactly those four exposed, and a class
without a definition never. On Micronaut's own jars (micronaut-inject,
-context, -http 3.9.7: 11/24/23 announced classes, +0/+11/+9 entries)
verdicts and `exposed_api` are unchanged — their beans are public. The 4.x
layout is not verified.
**Caller-supplied objects (`lb2-7`, 2026-11-19):** a call the
library makes on an object of a dependency type that its caller hands it —
a parameter of the API method the path starts at, or a library field set
only by roots the library never calls itself (a Spring `ObjectProvider` or
`BeanFactory`, a commons-text `StringLookup`) — runs whatever
implementation the caller chose; CHA takes it to be any. Such calls are a
`client_code` gap (site code `supplied`, never `MODEL_INCOMPLETE`), the
`exposed_api` entries reached only through one are `supplied_dispatch`
(report: "caller-supplied"), and a finding all of whose entries are carries
`SUPPLIED_DISPATCH_ONLY`; verdicts are kept. Verified on
`testdata/library-supplied` ([EXPECTED.MD](testdata/library-supplied/EXPECTED.MD),
`TestLibrarySupplied`, `TestSuppliedSites`). Motivated by a JVM oracle of
the mybatis Spring Boot starter (`testdata/library-starter`,
`scripts/library-starter-oracle.sh`): of its 26 `REACHABLE` findings 5 run
in the smallest application, 7 more from Boot's own YAML loading, 14 never.
On the five measured libraries `lb2-7` changes no verdict or exposed method;
it marks the starter's `MybatisAutoConfiguration#<init>` (its
`ObjectProvider`s) in all 26 findings, but no finding becomes
`SUPPLIED_DISPATCH_ONLY` — the other entries reach the same code through
logging initialization (slf4j → logback's configurator → Spring's
`Environment` → SpEL / `BeanFactory`), interface calls inside the
dependencies whose implementation configuration selects.
**Spring-announcing libraries are framework-mediated (`lb2-8`,
2026-11-19):** a library that announces classes to Spring
(`spring.factories`, `META-INF/spring/*.imports`) runs inside a Spring
container, so with `-roots api` the spring profile's framework packages
(Spring Boot, beans, core, expression, aop, tx, logback, log4j, …) count as
framework code whatever profile is selected, as `-profile spring` already
does for applications; the added patterns are in
`coverage.api.framework_packages_added` and the log. On the mybatis starter
(`-profile generic`) 25 of 26 `REACHABLE` become `FRAMEWORK` and one stays
`REACHABLE` — mybatis → OGNL through its XML configuration builder, which
runs only with an XML configuration. Why not a model: a JVM path-divergence
oracle (`testdata/library-starter`, `scripts/path-divergence.py`) showed
that at 11 of 13 diverging CHA sites the container's implementation is
chosen by configuration, and that even keeping only the executed targets
everywhere leaves all 26 reached through configuration- and data-dependent
branches in container code. Other libraries, CDI bean archives and
`-roots all` are unchanged. **Dependency declarations (LB3, `lb3-1`, 2026-11-18):** the findings are for
the versions in `-deps`, but a user of the library may get others, and some
dependencies exist only if the user adds them. With `-roots api` the
library's own pom — embedded in the jar (`META-INF/maven/<g>/<a>/pom.xml`)
or given with `-library-pom` — is read with its parents (properties,
`dependencyManagement`), and the tree a consumer gets is built as Maven
does: compile/runtime, non-optional declarations, transitively through
their poms from Central (via the cache), nearest first, exclusions
honoured. Those dependencies are `supplied_by: library`; `provided`,
`system` and `optional` ones, and what comes only through them, are
`user`; an input jar the library does not declare is `undeclared`. v2
`coverage.api.dependencies` lists every declaration (scope, `via`,
declared and resolved version, the matching instance and
`version_match`: `declared`, `differs`, `in_range`, `out_of_range`,
`absent`, …), `components[].declaration` and `findings[].supplied_by`
repeat it per instance; the report has a "Library dependencies" table,
and a warning names each dependency a consumer gets through the library
that is missing from the input (its advisories are not assessed). Log
line `library dependencies (lb3-1): <pom> (<gav>): N supplied by the
library (M not in the input), …`. Verified on `testdata/library-deps`
([EXPECTED.MD](testdata/library-deps/EXPECTED.MD), `TestLibraryDeps`,
`internal/component` `TestLibraryDeclarations`): the set the library
supplies equals what Maven 3.9 resolves for a consumer
(`scripts/library-deps-oracle.sh`). Versions managed by an imported BOM
(`scope import`, e.g. Spring Boot's) are resolved. A declared dependency
whose jar is in the input without an identity (no `pom.properties`) is
`unidentified`, with a warning. On the four measured libraries
(`scripts/library-measure.sh`, offline, `ARIADNE_MAVEN_LOCAL`), verdicts
and `exposed_api` are identical to `lb2-5`; commons-configuration2 has 3
dependencies supplied by the library and 17 optional ones by the user;
jackson-dataformat-yaml and jjwt-jackson 4 each, all supplied by the
library at the declared versions (jackson-core/-annotations through
databind); for the mybatis autoconfigure jar 8 Spring Boot/Spring jars
come through `spring-boot-autoconfigure` and are `unidentified` in the
input (their advisories were never assessed in that measurement), mybatis
and mybatis-spring are optional (user), and 9 input jars are undeclared
(the starter brings them). Since 2026-11-18 the measurement passes the
`mvn dependency:list` of each library as `-deps-list`: the Spring jars are
identified, and the starter has 39 findings instead of 19 — 20 new on
spring-boot/-autoconfigure/-beans/-context/-core/-expression, all supplied
by the library, 13 `REACHABLE` from the API under CHA (11 under RTA),
among them CVE-2022-22965 (spring-beans) and 8 SpEL advisories; the other
three libraries are unchanged. Not yet:
advisories for the other versions a user can get (the declared range).
`-roots all` and `profile` results are unchanged.

**JDBC drivers (`jdbc:drivers`, `internal/jdbcurl`, version `u1`,
2026-11-16).** `DriverManager#getConnection(url)` offers the URL to every
registered driver, and the service-provider model links `Driver#connect`
to every driver in `META-INF/services` — so an application on HSQLDB
reached pgjdbc's and MySQL's whole connection code just because their jars
are on the class path. A driver that does not accept the URL checks it
first thing and returns `null`. For a driver in the model's catalog
(pgjdbc 42.x, MySQL Connector/J 8.0.23+, H2 1.4/2.x), when no URL the
application can use starts with one of its prefixes, its `connect` is cut
after the URL check (`callgraph.Graph.CutSitesAfter`: later call sites are
removed, and a model adding an edge from one later is refused), before any
model or creator gate reads the graph. The URLs come from the application
jar — its classes' `jdbc:` string constants, its configuration files
(every Spring profile's `.properties`/`.yml`, Spring XML, a `web.xml`) —
plus H2, HSQLDB and Derby when Spring's embedded-database support is in
the input (Boot starts one without a URL). **Assumption:** the packaged
configuration is what runs. A URL set only from outside cannot be seen,
so a URL placeholder the packaged configuration does not set
(`${DATABASE_URL}`), a JNDI `DataSource` or a URL with no literal
`jdbc:<subprotocol>:` beginning turns the model off for the run (every
driver kept, the log says why); an environment variable overriding a
packaged URL is not seen. A driver outside the catalog, or a catalog
version whose `connect` has no URL check (Connector/J 8.0.11), is left as
it is. Log line `JDBC drivers (jdbc:drivers u1; URL prefixes …): kept …;
cut after the URL check …`; snapshots record `jdbc.drivers`; it can be
switched off in a profile (`models.disable: [jdbc:drivers]`). Real-JVM
oracle: `testdata/jdbcdrivers` (`scripts/jdbcdrivers-oracle.sh`, traced
through JDI by `scripts/MethodTrace.java`, `TestJDBCDriversOracle`) — the
JVM enters pgjdbc's and Connector/J's `connect` and URL check for a URL
neither accepts, and nothing past it; the graph reaches both checks and
nothing past them. WebGoat (HSQLDB): pgjdbc cut, its CVE-2022-21724,
CVE-2022-26520 and CVE-2026-42198 `REACHABLE` → `NOT_REACHABLE`; petclinic
(H2, MySQL profile) and JPetStore (HSQLDB only) unchanged.

**JPA objects (`jpa:objects`, `internal/jpadispatch`, version `e1`).** The
same gate for `javax.persistence` `EntityManager`, `EntityManagerFactory`,
`Query`, `TypedQuery`, `StoredProcedureQuery` and `EntityTransaction`: a
virtual/interface call on one of them gets a `model` edge to the body every
concrete implementation in the input runs for it, once a method creating
that implementation is reachable — the provider's entity manager once its
factory's `createEntityManager` runs, a query once `createQuery` runs.
Behind Spring ORM's proxies (the `framework:proxies` `spring-orm:*`
entries) the handler reaches the provider's object through `Method#invoke`
on its target, which no literal resolves: the real-JVM oracle
`testdata/jpa` (`scripts/jpa-oracle.sh`, `TestJPAOracle`; Spring ORM
5.3.6 over a minimal provider instead of Hibernate) ran all four handlers
and four provider bodies, and the graph reached the handlers only. The
service-provider model already covers a provider registered in
`META-INF/services/javax.persistence.spi.PersistenceProvider`, but only in
its own jars and only for calls it reaches; this model is not tied to the
registration. `jakarta.persistence` is not listed: it is not under the
skip rule, so CHA already expands it. On petclinic and WebGoat (Hibernate
5.4) it adds 139/138 edges (Hibernate's queries, `SessionImpl#find`/
`remove`/`flush`, …); verdicts are byte-identical in v1, also under RTA.
The log line is `JPA objects (jpa:objects e1): …`; snapshots record
`jpa.objects`.

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
   The table has 109 entries (audit:
   `testdata/cve-control/MANUAL_MAPPING_AUDIT.MD`); 22 of them (2026-11-13)
   cover advisories with neither a fix commit nor a class name — SnakeYAML
   DoS, Spring MVC static resources/SSE/script templates, SpEL, Spring log
   injection, Thymeleaf — each read by hand from a release diff and
   versioned over the releases where the methods were checked. 15 more
   (2026-11-16, the same way) cover the Spring advisories the reference
   apps left `UNKNOWN`: Spring Framework (the HTTP-invoker deserialization
   entry points of CVE-2016-1000027, four 4.3-line advisories whose fix
   commits name methods 4.3 does not have, `@RequestBody byte[]`, static
   resources on non-compliant containers), Spring Boot (`EndpointRequest.to()`,
   DevTools' secret check), Spring Data Commons (three DoS through the web
   argument resolvers) and Spring Security (OAuth2 authorization requests,
   header writing, X.509 subject parsing). Where the fix changed a shared
   utility that runs at startup anyway (`StringUtils#cleanPath`,
   `UrlResource#createRelative`, Spring Data's property-path caches), the
   entry names the entry point the advisory describes (role `entrypoint`).
   `UNKNOWN` on petclinic/WebGoat/JPetStore: 27/26/5 → 19/17/3. 14 more
   (2026-11-16) cover Tomcat (default-servlet writes, the JSON access log,
   DIGEST nonce counts), Jackson (async number parsing, `JsonNode#toString`,
   `InetSocketAddress`, case-insensitive properties), logback (`<if>` and
   complex-property instantiation in configuration files, the socket
   receivers' object stream), Undertow (HTTP/2 framing and flow control,
   multipart parsing) and XNIO (per-thread selectors). Where the fix-commit
   diff names a class only a newer line has (Jackson's 2.14+ async parser,
   logback 1.3+'s model processors), the older line's variant names the
   method that holds the same code there. `UNKNOWN`: 19/17/3 → 10/6/3.
   9 more (2026-11-16) cover H2 Console, MySQL Connector/J, pgjdbc, jsoup,
   hibernate-validator's `@SafeHtml` and Spring Boot's CloudFoundry
   actuator paths. An advisory that concerns no JVM code at all gets a
   `not_code: "<why>"` entry instead of methods: the finding is
   `NOT_APPLICABLE` with reason `NOT_CODE_MAPPABLE` (Tomcat's Windows
   installer, CVE-2025-49124). `UNKNOWN`: 10/6/3 → 3/1/3. Mapped code
   the analyzed classpath does not ship at all is `NOT_REACHABLE` with
   `MAPPED_CODE_ABSENT` (2026-11-16; 3/1/3 → 1/0/0); the one left is
   CVE-2022-21363, which neither Oracle's advisory nor the release diff
   attributes to any code. 2 more (2026-11-19, LIBRARY_MODE LB4) cover
   Jetty 9.4's HTTP parser (`+`-prefixed Content-Length, CVE-2023-40167)
   and `HttpURI` authority parsing (CVE-2022-2047), whose OSV records
   carry no fix commit; checked in the bytecode of 9.4.36 against 9.4.57.
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
   **Pull requests and field initializers (CVE:F10).** A reference to a
   GitHub pull request (`github.com/<o>/<r>/pull/<n>`) is resolved through
   the API to the commit that merged it (a two-parent merge commit is read
   against its first parent, the change the pull request brought); a pull
   request that was not merged is recorded with an error; references into
   `github/advisory-database` (edits of the advisory itself) are skipped.
   Unlike a commit the advisory names directly, a pull request whose
   metadata or diff cannot be read (an issue number, a rate limit — or,
   offline, a lookup that failed online and so was never cached) is
   recorded per reference instead of stopping the run; a snapshot replay
   still refuses a lookup it did not record. Changed lines of
   a field initializer or an initializer block count for the method their
   code compiles into — `<clinit>` for static ones, every constructor
   (`C#<init>`) for instance ones — but only when the fix changed no
   method: next to changed methods such a change is usually a constant
   those methods use, and its `<clinit>` runs whenever the class is used.

   **Static-initializer-only mappings (2026-11-17).** For the same reason
   a mapping (fix diff or release diff) that names only `<clinit>`s says
   little: reaching one means the class was used. Such a row keeps its
   verdict and carries the mapping qualifier `MAPPING_INITIALIZER_ONLY`
   (v2 `mapping.initializer_only`). One family is handled by hand
   (`internal/vulnmap/initializers.go`): jackson-databind's
   polymorphic-typing "gadget" advisories, whose fixes only add class
   names to the deserialization blocklist (`SubTypeValidator`'s static
   set, `BeanDeserializerFactory`'s before 2.9). That initializer runs on
   any Jackson deserialization, so every Jackson user used to get all of
   them as `REACHABLE`; the blocklist matters only when default typing
   (or class-name type ids) is enabled. A mapping made only of those two
   initializers is replaced by the API that enables default typing —
   `ObjectMapper#enableDefaultTyping`/`enableDefaultTypingAsProperty`/
   `activateDefaultTyping`/`activateDefaultTypingAsProperty`/
   `setDefaultTyping` and the `MapperBuilder` equivalents — with role
   `entrypoint` (`MAPPING_ENTRYPOINT_ONLY`), the initializers kept as
   evidence (v2 `mapping.initializer_redirect`) and a row note.
   `@JsonTypeInfo(use = CLASS)` on an application type enables the same
   resolution without any of those calls and is **not** covered: a
   `NOT_REACHABLE` there carries `MAPPING_PARTIAL` like every partial
   mapping. On jjwt-jackson 0.11.2 (databind 2.9.10.4) 21 of 36
   `REACHABLE` advisories were this family; they are now `NOT_REACHABLE`
   (`REACHABLE` 36 → 15 with `-roots all`, CHA and RTA; with library
   mode's `-roots api` 36 → 15 under CHA, 35 → 14 under RTA).
   On petclinic/WebGoat/JPetStore no verdict changed; the qualifier marks
   2/2/3 rows (`UriComponentsBuilder#<clinit>`, `MimeTypeUtils#<clinit>`).

3. **The released bytecode (CVE:F10, `-release-diff`, on by default).** For
   a CVE sources 1 and 2 left unmapped, the methods whose code differs
   between the last release before the fix and the fixing release, read
   from Maven Central (through the cache), **in the classes the advisory
   text names** — CamelCase identifiers with at least two humps, or any
   identifier the text marks as code with backticks. Only patch releases
   are compared (the release before the fix must be in the same
   major.minor line; the OSV version is matched to the published one,
   `2.2.10` → `2.2.10.Final`); a changed constructor or `<clinit>` counts
   only when no other method of the class changed; more than 40 methods is
   too broad to map. Labelled `source: release-diff`, vulnerable/partial/
   unreviewed, with `release_diffs` evidence (releases compared, named
   classes, counts) in v2. Validated against 130 advisories that have a
   fix-commit mapping: 25 of the 80 that name a class get a mapping, 24 of
   them contain a method of the real fix; 52 of 55 methods are methods the
   fix commit changed (95 %), and 52 of the 61 fix methods in the named
   classes are found ([validation](testdata/cve-control/RELEASE_DIFF_VALIDATION.MD)).
   Central requests are serialized and spaced (750 ms, `User-Agent:
   ariadne`); offline, a jar the cache lacks is read from `~/.m2/repository`
   (a snapshot replay never does). An advisory that names no class, or a
   fix first released in a new minor line, stays unmapped.

4. **An LLM (opt-in, `-llm`).** Only for a CVE sources 1–3 found
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
| `NOT_REACHABLE` | vulnmap has methods for this CVE, at least one of them exists in the analyzed graph, but neither pass ever visits any of them. Since CVE:F1b the row carries `reasons`: `MAPPING_PARTIAL` when the mapping is not known to be complete (every mapping today), and `MODEL_INCOMPLETE` plus `model_gaps` (counts by kind) when reachable unresolved reflection/MethodHandle/callback/ServiceLoader sites or deserialization entries could dispatch into the affected component. That gap rule is **report-only**: on the reference apps it applies to every `NOT_REACHABLE` finding (16 of 16 petclinic, 10 of 10 WebGoat, after DESERIALIZATION Z1), so turning them into `UNKNOWN` would carry no information ([docs/RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD) §4). Read the verdict as "no path in the model", not "safe". Also (2026-11-16) when **no class the mapping names is on the analyzed classpath** (`MAPPED_CODE_ABSENT`, the note names the classes): the version or the application does not ship that code — JPetStore without `spring-webmvc`, a `tomcat-embed-core` without the JSON access log valve. This assumes, like every `NOT_REACHABLE`, that the inputs are the deployed classpath. |
| `UNKNOWN` | vulnmap has no entry for this CVE at all; or (CVE:F1b, with `reasons`) **none of the mapped methods exists in the analyzed graph** while at least one mapped class does (`METHOD_NOT_FOUND` — e.g. a diff-derived method absent under that name in the analyzed version; 7 petclinic and 6 WebGoat rows that v1 used to report as `NOT_REACHABLE`), or the affected instance's own copy of a mapped class lost to another jar's copy under the classpath order (`CLASSPATH_CONFLICT` — a path through the other copy is not that instance's code). |
| `NOT_APPLICABLE` | Every affected dependency is a WebJar (no `.class` file, assets under `META-INF/resources/webjars/`): client-side JavaScript/CSS a call graph cannot analyze. Checked before everything else. **Not** a verdict that the app is safe: the code runs in the browser. Also (2026-11-16) when the manual table declares, with a reason, that the advisory concerns no JVM code at all (`not_code`; reason `NOT_CODE_MAPPABLE`) — Tomcat's Windows installer filed against `tomcat-embed-core` — with that reason as the note. An empty jar alone does not qualify, since Maven aggregators such as `spring-boot-starter-web` are class-free too and CVEs filed against them (Spring4Shell) concern JVM code in other jars. |

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
above, plus the `web.xml` lifecycle roots described below) narrows the BFS roots to the active profile's own entry points
instead: a method carrying one of `roots.annotations`
(`@RequestMapping`/`@GetMapping`/.../`@Scheduled`/`@EventListener`/`@Bean`
and the message listeners `@JmsListener`/`@KafkaListener`/`@KafkaHandler`/
`@RabbitListener`/`@RabbitHandler` with their `@Repeatable` containers
for `spring`), every public method of a class itself carrying one,
(when `roots.main` is true) every public static `main(String[])`, and
every `roots.overrides` entry (`"fqcn#method"`, e.g. `javax.servlet.
http.HttpServlet#doGet`): every app class that is that class itself or
one of its transitive subtypes **and declares that method itself**
(an actual override, not an inherited no-op) — `servlet.yaml`'s own use
of this (`HttpServlet#do*`/`Filter#doFilter`, both the `javax.servlet.*`
and `jakarta.servlet.*` namespaces) is the first shipped profile to need
it; `spring.yaml`'s own list is still empty.

**Message listeners (`spring` profile, since 2026-11-16).** A broker's
listener container calls a listener method through
`MessagingMessageListenerAdapter` → `InvocableHandlerMethod` →
`Method#invoke`; nothing in the application calls it, so under `-roots
profile` it is an entry point like an HTTP mapping. The `spring` profile's
`roots.annotations` list `@JmsListener`/`@JmsListeners` (spring-jms),
`@KafkaListener`/`@KafkaListeners`/`@KafkaHandler` (spring-kafka) and
`@RabbitListener`/`@RabbitListeners`/`@RabbitHandler` (spring-rabbit); a
class-level `@KafkaListener`/`@RabbitListener` roots its public methods
(its handlers, and its public constructor, by the class-annotation rule).
Roots are matched by annotation name: whether the listener container is
configured (`@EnableJms`/`@EnableKafka`/`@EnableRabbit`, Boot
autoconfiguration) is not checked, like `@Scheduled` without
`@EnableScheduling`. Spring Cloud Stream functions, `@SqsListener` and
programmatic `MessageListener` registrations are not covered (a class
implementing `MessageListener` is reached by CHA only when a container
call is reachable). Verified: `TestMessageListenerRoots` (name-only
stand-ins of the annotations, `testdata/listeners`, method- and
class-level, repeated); control app `scripts/prepare-listeners-demo.sh`
(Spring Boot 2.4.3, in-memory ActiveMQ, a `@JmsListener` calling
commons-text 1.9 `StringSubstitutor#replace`): under `-roots profile`
CVE-2022-42889 was `NOT_REACHABLE` and is now `REACHABLE` with the
listener as the root (the JVM's listener container runs it); the variant
without the annotation stays `NOT_REACHABLE` (nothing runs it); under
`-roots all` it is `REACHABLE` in both (every application method is a
root). petclinic, WebGoat and JPetStore have no message listener: their v1 results are byte-identical (`-roots all`/`profile`, CHA/RTA).

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
(the B3 registry models — ServiceLoader, servlet container dispatch,
service-provider dispatch — and the creator-gated models: servlet and JDBC
objects, framework proxies);
`callback`: a resolved `Thread`/`Executor`/`AccessController.doPrivileged`/`CompletionStage` registration connected to the
callback's own method, ASYNC_CALLBACKS A2/A3; `methodhandle`:
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

**Every step of a path, not only the entry (ROADMAP B0/B1, 2026-11-13).**
Each step of a REACHABLE/FRAMEWORK path is rebuilt from the graph with
every edge between its two methods (kind, basis, call-site offsets), and
gets a character: `jvm_rule` (a bytecode call or class initialization),
`library_model` (a modeled API), `input_declaration` (a profile's declared
edge) or `heuristic` (a CHA expansion only — some subtype's override
assumed possible); a step a bytecode call also backs is `jvm_rule`. The
report marks the other steps in the path (`→[heuristic: cha]`) and adds a
"Path evidence" line; v2 findings carry `path_edges` and `character`.
These are kinds of evidence, not a confidence score. **Since `b0-2`** the
reported path is one with no CHA-only step whenever such a path exists
(even a longer one; the verdict itself is decided as before), so a
`heuristic` character means the method is reachable only through some
CHA expansion. Measured on the default runs: 99 of 109 positive findings on
petclinic, 81 of 137 on WebGoat and 16 of 16 on JPetStore depend on a CHA
expansion (before `b0-2` the shortest path showed 106, 89 and 16).
An advisory no source could map is `UNKNOWN` with `MAPPING_MISSING` (also
in v1's `reasons`).

**One gap vocabulary (ROADMAP B0 b0-3).** Each analyzer that reports model
gaps — the unresolved sites of the model-gap rule, the created-type
catalog's unknown allocations, the SAM-call gaps, the B3 coordinator's
incomplete facts, a trace's causes of calls missing from the graph — keeps
its own codes, and each code maps to one mechanism (`reflection`,
`method_handle`, `callback_async`, `service_loader`, `deserialization`,
`dynamic_proxy`, `container_proxy`, `unknown_allocation`, `external_input`,
`client_code`, `lambda`, `skipped_receiver`, `virtual_dispatch`, `class_init`,
`model_incomplete`, …; `internal/gaps`, table in
[docs/RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD)). The report's "Model
gaps by mechanism" table and v2 `coverage.gaps`/`findings[].gaps` use it.
It is a reading aid: only the `site` gaps take part in the model-gap rule,
and no verdict or reason code changes (v1 is byte-identical).

**Gap-rule probe (ROADMAP B0 b0-4).** Every run also measures, report-only,
which `NOT_REACHABLE` findings a narrower or wider model-gap rule would mark:
only gaps bounded to the instance (`scoped`), plus unresolved lambda calls
whose interface has a lambda created in the instance, plus calls on
`java.*`/`javax.*` types whose skipped CHA expansion would land in the
instance. The log line "gap-rule probe", the report's "Gap-rule probe"
section and v2 `findings[].gap_probe`/`coverage.gap_probe` show it. On
petclinic, WebGoat and JPetStore the wider variants mark exactly what the
current rule marks (every `NOT_REACHABLE` finding), and the skipped receivers
reach every instance — so the rule is unchanged ([contract
§4](docs/RESULTS_CONTRACT.MD)). No verdict or reason code depends on it; v1
is byte-identical.

**Unbounded gap sites (ROADMAP B0 b0-5).** The sites that make the rule
mark everything are the reachable ones with no bounded candidate set —
unresolved `Method#invoke`/`Constructor#newInstance`, `MethodHandle`
calls, deserialization entries. Every run reports where they are: the log
line "unbounded gap sites", the report's "Unbounded gap sites" section (by
the calling class's component and by calling method) and v2
`coverage.unbounded_sites`. A site could only be bounded by a model of
what its calling method dispatches to, so the measurement asks whether a
few dispatchers hold most of them. They do not: petclinic 611 sites in 477
distinct methods across 40 components (Tomcat 98, Byte Buddy 88, Hibernate
79, …; the 15 busiest methods hold 82), WebGoat 681 in 549 across 55,
JPetStore 161 in 129 across 14. The framework dispatchers ariadne already
models (Spring bean creation and lifecycle, Spring Data's
`RepositoryMethodInvoker`, Spring ORM's handlers, Spring MVC's handler
invocation, Stripes) hold 25, 25 and 19 of them — bounding all of those
would leave every `NOT_REACHABLE` finding marked, so no such catalog was
built. Report-only; v1 is byte-identical.

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
| `-profile` | no | `auto` (default, auto-detect and merge), `generic`, `spring`, `servlet`, `jakarta`, `micronaut`, `quarkus`, `stripes`, or a name from `-profile-file`. |
| `-profile-file` | no | Load/override one profile from a YAML file (same shape as the embedded defaults). |
| `-roots` | no | `all` (default: every application method, plus the lifecycle methods of `web.xml`-registered servlets/filters/listeners), `profile` (the active profile's own entry points, plus the same `web.xml` methods) or `api` (library mode, below: the `-app` jar is a library and its public API is the roots). |
| `-library-pom` | no | With `-roots api`: the library's `pom.xml`, read for its dependency declarations (LIBRARY_MODE LB3) — for a library jar that embeds none (e.g. a Gradle build); default: the pom in the jar's `META-INF/maven/`. |
| `-api-exclude` | no | With `-roots api`: comma-separated package prefixes left out of the API (`com.example.internal`). |
| `-target-java` | no | Java feature version (e.g. `17`) to resolve multi-release JAR variants against (JVM_GRAPH G4). Unset: every jar's base classes, regardless of any `META-INF/versions/` present. |
| `-jpms-scenario` | no | ServiceLoader launch scenario (SERVICE_LOADING S3): `classpath` (default — every jar in the unnamed module, `module-info.class` ignored, providers from `META-INF/services`) or `module-path` (every jar, the app included, on the module path: explicit modules provide only through `provides`, a provider with `public static provider()` is created through it, automatic modules provide through `META-INF/services` with a constructor, and a caller in an explicit module must declare `uses`), or `mixed` (only the `-jpms-module-path` jars on the module path, every other jar on the class path; module-path modules assumed resolved as with `--add-modules ALL-MODULE-PATH`). Recorded in the report header; `results.json`'s format is unchanged. |
| `-jpms-module-path` | with `mixed` | Comma-separated jar file-name patterns (`path.Match` syntax, e.g. `app,commons-text-*.jar`) selecting the module-path jars for `-jpms-scenario mixed`; `app` names the analyzed application. Rejected with any other scenario; a pattern matching no jar is a warning. |
| `-active-profiles` | no | Comma-separated Spring profile names to treat as active when evaluating `@Profile` conditions (FRAMEWORKS F1 — see "8. The report"). Unset: every `@Profile` condition is reported `unknown`, never a guessed `false`. |
| `-callgraph-mode` | no | Dispatch strategy for virtual/interface calls (GRAPH_PRECISION P0/P2/P3): `cha` (the default), `rta` (opt-in: CHA edges narrowed to the classes created by code reachable from the roots, with the unknown-allocation fallback; see section 4) or `pta` (experimental: RTA, then every CHA call site whose receiver's allocated classes are proven keeps only the targets they select). Recorded with the graph's other settings in the report header and v2 `callgraph`. |
| `-graph-out` | no | Write the final call graph — every edge with a stable identity (source/target method, call kind, basis) and its call-site PCs, the graph's settings, the findings' paths and the created-type catalog (GRAPH_PRECISION P1) — as sorted text (`ariadne-callgraph-2`; large: ≈ 370 MB for WebGoat). |
| `-jdk-home` | no | The target runtime's installation directory (CVE:F8a): its `release` file gives vendor, version, OS and architecture; nothing is run, and the scanner's own JDK is never assumed. Report header "Target runtime", v2 `target_runtime`. For a Temurin runtime the OpenJDK Vulnerability Group advisories of the curated catalog are matched (CVE:F8b): report section "Target runtime advisories", v2 `target_runtime.advisories` — applicability only, reachability `not_assessed`, never part of the v1 verdicts; any other distribution is `not_performed` with the reason. The embedded catalog (version 1, curated 2026-11-20) holds the eight advisories from 2026-01-20 to 2026-11-17; a runtime older than the first of them gets a note that earlier advisories are not included. Exclusive with `-runtime-file`. See [docs/RUNTIME_SOURCES.MD](docs/RUNTIME_SOURCES.MD). |
| `-openjdk-catalog` | no | Use this curated OpenJDK advisory catalog (YAML from `go run ./cmd/openjdk-catalog`) instead of the embedded one — to check a draft before committing it. ariadne never fetches openjdk.org (its terms of use do not allow automated access); see [docs/RUNTIME_SOURCES.MD](docs/RUNTIME_SOURCES.MD). |
| `-runtime-file` | no | A YAML description of the target runtime (`vendor`, `distribution`, `version`, `os`, `arch`, `libc`, `package: {name, version, distro}`; unknown fields are errors), e.g. for a runtime that is not at hand. Exclusive with `-jdk-home`. |
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

### Library mode in a library's CI (LIBRARY_MODE LB4)

For a library artifact (a jar others depend on), run the analysis on the
built jar with its runtime dependencies and `-roots api`:

```bash
mvn -q dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory=target/deps
mvn -q dependency:list -DincludeScope=runtime -DoutputFile=target/deps.txt
bin/ariadne -app target/mylib-1.2.3.jar -deps target/deps -deps-list target/deps.txt \
  -roots api -json-format v2 -out ariadne.md -json ariadne.json
```

- Add `-api-exclude com.example.mylib.internal` for packages that are public
  to the JVM but not API (repeatable). The library's own pom (embedded in the
  jar) is read for `supplied_by`; set `ARIADNE_MAVEN_LOCAL` (default
  `~/.m2/repository`) so parent poms come from the local repository, not
  Maven Central.
- Read a finding as: **can a user of this library's API reach the vulnerable
  code** — not whether a particular application does. `exposed_api` lists the
  API methods that reach it, each with an example path; `supplied_by` says who
  picks the dependency version (`library`: the library's declaration;
  `user`/`undeclared`: the application decides, so upgrading it is theirs).
- Qualification (`testdata/library-lb4/EXPECTED.MD`, five real libraries:
  facade, factory/subclass, SPI, Spring Boot starter, callbacks): `-roots api`
  never lost a `REACHABLE` of `-roots all`; a JVM oracle per library
  (`TestLibraryJetty`, `TestLibraryLB4Oracles`: of the `REACHABLE` findings
  2/2, 15/17, 14/15 and 2/2 run from the public API, no `NOT_REACHABLE` one
  ever ran) confirms the verdicts, with these limits:
  example paths across asynchronous hand-offs (NIO selectors, executors) are
  not execution traces; public-but-internal classes count as API; a Spring
  Boot starter's findings mostly run through the container it plugs into —
  since `lb2-8` a library that announces classes to Spring gets the spring
  profile's framework packages whatever `-profile` says, so those are
  `FRAMEWORK` (reachable only through framework code; on the mybatis starter
  25 of 26) and `REACHABLE` keeps what the library's own code reaches
  (`coverage.api.framework_packages_added` lists what was added);
  `exposed_api` can list hundreds of methods for a facade or client library.

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
the analyzer version check is the backstop. Every kind of call-graph edge has the version
of the rules that add it in the snapshot (ROADMAP B3 `b3-4`, 2026-11-20: the fixed-point
models and gates through the `models` configuration with their versions, the one-shot
passes — base callback rules `callback.resolver`, `methodhandle.edges`,
`reflection.invoke` — and `jdk.api`, `gaps.rules` as algorithm versions);
`TestAlgorithmVersions_everyEdgeBasis` reads every edge basis `internal/callgraph`
declares and fails until a new one is mapped to a version. Roots are not edges, so the
rules that turn a profile into roots and framework facts have their own entries
(FRAMEWORKS F5 `f5-1`, 2026-11-22): `profile.rules` (root kinds, `roots.beans`, `web.xml`
and Spring XML roots, entry catalog, injection points, `@Value` facts) and
`profile.builtin` (a digest of the built-in profile YAMLs, so a YAML change needs no bump;
a `-profile-file` is hashed as an input as before). `TestRulesVersion_tracksRuleSource`
hashes that rule code without comments and fails until `profile.rules` is bumped.
The comparison is per run,
not per model: a version change invalidates every advisory, not only those whose paths
use the changed edges.

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
./scripts/prepare-jpetstore.sh        [JAVA_HOME]                       # Maven, WAR (web.xml, Stripes, Spring 4.3 XML, MyBatis, HSQLDB); analyze target/jpetstore.war with -deps-list
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
./scripts/async-oracle.sh             [JAVA_HOME]                       # real-JVM oracle for callbacks (ASYNC A0–A2) and MethodHandle (L2): method bodies that ran (testdata/async)
./scripts/jdbc-oracle.sh              [JAVA_HOME]                       # real-JVM oracle for JDBC objects and driver dispatch (testdata/jdbc)
./scripts/mybatis-oracle.sh           [JAVA_HOME]                       # real-JVM oracle for MyBatis mapper proxies down to the driver, from stack frames (testdata/mybatis)
./scripts/springdata-oracle.sh        [JAVA_HOME]                       # real-JVM oracle for Spring Data repository proxies, from stack frames (testdata/springdata; jars from prepare-petclinic.sh)
./scripts/runtime-trace-fixture.sh    /path/to/jdk25-or-later/JAVA_HOME # JFR-recorded runtime trace of testdata/runtime (rtm.jar, rtm-v2.jar, run.trace)
./scripts/runtime-trace-petclinic.sh  [OUT_DIR] [PORT]                  # write petclinic's vulnerable-method JFR settings, run it under JFR with them, send a fixed scenario, convert, import (java 25+ on PATH)
./scripts/runtime-trace-jpetstore.sh  [OUT_DIR] [PORT]                  # the same for JPetStore's WAR in Jetty (jetty-runner fetched by Maven)
./scripts/spidispatch-oracle.sh       /path/to/jdk17-or-later/JAVA_HOME # real-JVM oracle for service-provider dispatch (testdata/spidispatch)
./scripts/reflection-oracle.sh        [JAVA_HOME]                       # real-JVM oracle for reflection R0–R2: method bodies that ran (testdata/reflection)
./scripts/reflectparam-oracle.sh      [JAVA_HOME]                       # real-JVM oracle for reflection R3: Methods passed as parameters (testdata/reflectparam)
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
                             quarkus.yaml, stripes.yaml, generic.yaml (embedded defaults)
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
  (ASYNC_CALLBACKS A2 — "Callback call-graph edges" section). Since
  ASYNC_CALLBACKS A3 slice 1 the catalog also holds every
  callback-taking method of `CompletionStage`, on the interface and as
  `CompletableFuture` overrides it (`thenApply`/`thenAccept`/`thenRun`,
  `thenCombine`/`thenAcceptBoth`/`runAfterBoth` and the `*Either` forms —
  the callback is then the second argument — `thenCompose`, `handle`,
  `whenComplete`, `exceptionally`, each with its `*Async` and
  `*Async(…, Executor)` forms), `CompletableFuture`'s
  `exceptionallyAsync`/`exceptionallyCompose*` (Java 12+) and
  `supplyAsync`/`runAsync`. A plain continuation is `immediate-or-async`
  (it runs on the registering thread when the stage is already complete),
  an `*Async` one and a source `async-maybe`. `exceptionally*` callbacks
  run only when the previous stage fails: their edges carry the condition
  `exceptional`, kept in the v2 path evidence (`"condition"`) and shown on
  the report's arrow (`→[library_model: callback, if exceptional]`);
  `handle`/`whenComplete` run on both outcomes and carry none. The real-JVM
  oracle `testdata/continuations` (`scripts/continuations-oracle.sh`,
  `TestContinuationsOracle`) measured that a lambda or method reference
  handed to a continuation was already reached — the edge from the
  `invokedynamic` creating it — so for those the model adds the
  registration edge and its condition, i.e. the explanation; what it adds
  to reachability is an object of a class handed to a continuation
  (`runAsync(new Task(), executor)`, `thenApply(new Render())`), whose
  method nothing reached before. Whether a stage ever completes, and how,
  was not decided by that slice (since `a3-7` it is for hand-completed
  `new CompletableFuture`s, below): the normal branch of a `supplyAsync`
  stage that always fails and the exceptional branch of one that always
  succeeds keep their "may run" edge (the oracle has one of each). Edges are per registration site, so a long pipeline needs
  no budget. Control app `scripts/prepare-continuations-demo.sh`
  (commons-text 1.9, `thenApply(new Render())`): CVE-2022-42889 under
  `-profile generic -roots profile` was `NOT_REACHABLE` and is now
  `REACHABLE` (`main` → `Render#apply(Object)` bridge → `apply(String)` →
  `StringSubstitutor#replace`), as on the JVM; the variant creating the
  `Render` without registering it stays `NOT_REACHABLE`, as on the JVM.
  Reference apps (petclinic, WebGoat, JPetStore; `-roots all`/`profile`,
  CHA/RTA): +9/+9/+2 registration sites and `callback` edges, all in
  libraries; v1 results byte-identical.
  Since `a3-4` a task a method *returns* to the code running it
  (`task = producer.produce(); task.run()`, or handed to an executor) is
  followed through the return values of the called method — its exact
  target, or the implementations of a virtual/interface call declared to
  return `Runnable`/`Callable`, narrowed by the receiver's field/parameter
  values when those are known — and final fields: a `callback` edge with
  condition `produced` from the running site to the task's body, gated on
  a reachable method returning it (log line `returned tasks (ASYNC_CALLBACKS
  a3-4)`; `testdata/async-returned`, `scripts/async-returned-oracle.sh`).
  On jetty-client this is the edge from Jetty's `EatWhatYouKill` to
  `ChannelEndPoint$1#run` (`_runFillable`).
  Since `a3-5` a listener runs when its source notifies: a listener
  stored into a JDK collection field and a method walking that field
  without removing or clearing it (a broadcast), or
  `PropertyChangeSupport#addPropertyChangeListener` with
  `firePropertyChange` on the same field, give a `callback` edge
  (condition `notified`) from the notifying call to the listener's
  method, gated on the method registering it (log line `listeners
  (ASYNC_CALLBACKS a3-5)`; `testdata/async-listeners`,
  `scripts/async-listeners-oracle.sh`). Listener arrays, lists walked
  through a parameter, `forEach` with a lambda and `Observable` are not
  modeled.
  Since `a3-6` reactive streams of one library — the JDK's `Flow` with
  `SubmissionPublisher` — are modeled: `subscribe(s)` gives
  `s.onSubscribe`; `submit`/`offer` on the same publisher (same `new`
  site, or `this` in a processor subclass) give `onNext` and `onError`;
  `close` gives `onComplete`, `closeExceptionally` `onError`; `consume(c)`
  gives `c.accept` — `callback` edges with conditions `subscribed`,
  `published`, `completed`, `exceptional`, gated on the method subscribing
  (log line `reactive streams (ASYNC_CALLBACKS a3-6 …)`;
  `testdata/async-reactive`, `scripts/async-reactive-oracle.sh`). A
  pipeline nothing emits into gives no `onNext`; demand is not modeled.
  Since `a3-7` the completion of a future created with
  `new CompletableFuture<>()` and completed by hand is decided: its
  continuations are gated on reachable completion calls of the right kind
  on that future (`complete`… for `then*`, `completeExceptionally`/
  `cancel`… for `exceptionally*`, either for `handle`/`whenComplete`),
  and a future nothing completes gets no `callback` edge (log line
  `future completion (ASYNC_CALLBACKS a3-7)`; `testdata/async-completion`).
  Stages from `supplyAsync`, other continuations or libraries keep their
  "may run" edge.
  Since `a4-1` `start()` on a `Thread` subclass of the input reaches the
  subclass's own `run()` (condition `started`), and a
  `RejectedExecutionHandler` given to a `ThreadPoolExecutor` may run on
  `execute`/`submit` on that executor (condition `rejected`, gated on the
  registering method; `testdata/async-threads`).
  Not modeled: other reactive libraries, Spring's `ListenableFuture`,
  a callback passed through a parameter or field (a reported gap, as
  before), and a lambda whose own bootstrap isn't LambdaMetafactory (e.g.
  `StringConcatFactory`), left unresolved rather than guessed at.
- **What an asynchronous edge means (ASYNC_CALLBACKS A4 `a4-2`).** Every
  edge the async models add (basis `callback`) is a "may run" relation from
  the place where control may pass — a registration, an emission, a
  notification, a drain, a run of a returned task — to the body that may
  run. It is not a call stack of one thread and not a time order: the
  body may run on another thread, later, or never (a pool that rejects
  the task, a future nothing completes, a source that never notifies —
  the gated models drop what the input proves cannot happen, the rest
  stays may-run). Conditions on the edge say what has to happen first:
  `exceptional`, `queued`, `selected`, `produced`, `notified`,
  `subscribed`, `published`, `completed`, `started`, `rejected`; v2 path
  evidence and the report's arrow carry them. Not modeled: thread
  identity, interleavings, scheduling and cancellation policy, demand in
  reactive streams, memory visibility — none of these can make an edge
  disappear, and an unknown possibility of cancellation never proves a
  task never runs. Qualification: `testdata/async-qualify/RESULTS.MD`
  (`TestAsyncQualification`): the whole pipeline on all nine async
  fixtures against their real-JVM oracles — every body that ran is
  reached (one miss, a `MethodHandle` name chosen at run time —
  DYNAMIC_LINKAGE), and of the 6 of 60 `callback` edges to bodies the
  oracles did not record, 2 lead to bodies that ran unmarked and 4 are
  documented may-run cases — none a wrong association;
  `TestAsyncTwoTasks`: two tasks at two sites never mix. The gated async
  models (`async-callbacks:*`) and their versions are recorded in every
  snapshot's configuration (`models`) and in v2 `callgraph.participants`,
  so a comparison names a model change; the plain-edge rules (A2,
  `a3-2`, `a3-3`, `a4-1`'s `start()`) are not participants and show only
  as edge differences.
- **Lambdas run in another method than the one creating them** get a `sam`
  edge only where DATA_FLOW follows the value (DYNAMIC_LINKAGE L1, above):
  not through the result of an instance call (except a `Runnable`/`Callable`
  run or submitted after being returned — ASYNC_CALLBACKS `a3-4`, above), an array element, a
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
  profile`. Since the JPetStore measurement they are roots of the default
  `-roots all` too (every application method **plus** these): a WAR's
  container entry points are often library classes named in `web.xml`
  (Spring's `ContextLoaderListener`, Stripes' filter), which "every
  application method" alone left out — on JPetStore
  (`scripts/prepare-jpetstore.sh`) 8 advisories were `NOT_REACHABLE` under
  `-roots all` and `FRAMEWORK` under `-roots profile`, 3 of them with their
  vulnerable method executing in a recorded run; both scenarios now agree
  ([RESULTS.MD](testdata/graphbench/RESULTS.MD), "JPetStore"). The log
  line `roots: all (N methods: every application method, M more from
  web.xml)` counts them.
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
  a `new` instruction itself already gets). A fifth slice, REFLECTION R3 (`r3-3`, 2026-11-15): a `Method#invoke`
  whose Method receiver is a parameter of the calling method is resolved
  from what its callers in the input pass (`dataflow.CallerIndex`) — a
  resolved `getMethod`/`getDeclaredMethod` fact directly, or one read from
  a field or returned by a static helper one step away. Each resolved
  value is an edge; when every value resolves the site stops being a
  model gap, otherwise it stays one (partial). Measured first: on
  petclinic and WebGoat exactly 2 sites have this shape — Thymeleaf's
  `BinaryOperationExpression#doComposeBinaryOperationExpression`, whose
  `isLeftAllowed`/`isRightAllowed` Methods six subclasses keep in static
  fields (12 edges; reflection gaps 510 → 508 and 536 → 534; no verdict
  changed); JPetStore has one partial site and no new edge. Real-JVM
  oracle `testdata/reflectparam` (`scripts/reflectparam-oracle.sh`,
  `TestReflectParamOracle`): 4 → 9 of 10 bodies the JVM ran; the tenth is
  looked up by a run-time name. Every run also reports where the remaining
  unresolved receivers come from — log "reflection census", report
  "Unresolved reflection by receiver provenance", v2
  `coverage.reflection_census`: mostly values the analysis does not track
  (results of virtual calls, array elements) and fields holding them; on
  the three apps no further site is convertible. `getDeclared*`'s own
  access-check bypass/`setAccessible`, `Constructor#newInstance` through
  parameters (no such site measured), and any runtime-computed
  class/method name remain unmodeled, so the rest of this limitation still
  stands. A sixth slice, REFLECTION R4 (`r4-1`, 2026-11-21):
  `Class.forName` with a literal name initializes the class — `forName(String)`
  always, `forName(String, boolean, ClassLoader)` unless the boolean is a
  constant `false` (a non-constant one gives a "may run" edge) — so the call
  gets a `BasisReflection` class-initialization edge to the class's
  `<clinit>`, or to its nearest superclass's when it has none (the JVM
  initializes the superclass first). Log line "Class.forName call sites".
  Real-JVM oracle: `testdata/reflection` (direct, through the superclass,
  non-constant, and `initialize=false` with no edge). Measured: 22 / 27 / 8
  edges on petclinic / WebGoat / JPetStore; v1 results byte-identical with
  and without them in all 12 configurations (`-roots profile`/`all` ×
  CHA/RTA).
  A seventh slice, REFLECTION R4 (`r4-2`, 2026-11-22): a lookup the JVM
  itself rejects with `NoSuchMethodException` — `getMethod` of a
  non-public method, `getDeclaredMethod` of an inherited one,
  `getConstructor` of a non-public constructor — is checked on the class
  the lookup was made on and gives a `reflection` gap instead of an edge
  (it gave a false edge before). `setAccessible` is not tracked, so a
  private method invoked without it keeps its edge (an extra edge, never a
  missed one). The real-JVM oracle adds these cases and a lookup with no
  `invoke`; on petclinic / WebGoat / JPetStore no edge changes and v1
  results are byte-identical in all 12 configurations.
- **Seven profiles exist: `spring`, `servlet`, `jakarta`, `micronaut`,
  `quarkus`, `stripes` and `generic`** — every framework PLAN.md section 4
  originally named for Phase 7, the `spring`/`generic` pair from the base
  plan, and `stripes` (2026-11-14: Stripes action beans' event handlers,
  through the convention rule `roots.handlers`, found missing by a recorded
  JPetStore run). Since 2026-11-15 the `spring` profile also sets
  `roots.beans`: the methods the Spring container calls on its beans by
  reflection are roots in both scenarios — the constructor it picks (the
  only one, else the `@Autowired`/`@Inject` one, else the no-argument one),
  `@Autowired`/`@Inject`/`@Resource`/`@Value` methods, `@PostConstruct`/
  `@PreDestroy` (also in application superclasses), and a `@Bean`'s
  `initMethod`/`destroyMethod`, including the inferred `close()`/
  `shutdown()` on the class the factory method returns or creates. A bean
  is a concrete application class with a stereotype, `@Configuration` or
  `@ControllerAdvice`; one whose `@Profile` `-active-profiles` shows
  inactive is left out, while a `@Conditional` or `@Lazy` bean counts (the
  container decides at run time; the rule never guesses false). Found by a
  real-JVM oracle (`testdata/di`, `scripts/di-oracle.sh`, `TestDIOracle`,
  Spring 5.3.6): before, the graph missed every such body — constructors,
  an `@Autowired` setter, `@PostConstruct`, `@PreDestroy`, `@Bean`
  init/destroy — while the callback interfaces (`InitializingBean`, the
  `Aware` interfaces, `FactoryBean`, post-processors, `ApplicationListener`)
  were already reached through CHA. The log line `spring bean lifecycle
  roots: N methods (…), N new roots` counts them. Reference apps: v1
  verdicts unchanged (petclinic 8, WebGoat 165, JPetStore 3 such methods);
  under `-roots profile` two WebGoat paths now start at a lifecycle method
  (`WebSecurityConfig#configureGlobal`, `@Autowired`;
  `HSQLDBDatabaseConfig#shutdown`, `@PreDestroy`). Since `b4-8` the same
  rule covers **Spring XML bean definitions**, read from the application's
  class path and — new — from a WAR's `WEB-INF/*.xml` (also per EAR
  module): for every non-abstract `<bean>` (after merging its `parent`
  chain; inner beans too) whose class is in the input, application or
  library, the constructors taking as many parameters as it has
  `<constructor-arg>`/`c:` arguments, the setter of each `<property>`/`p:`
  property (in superclasses too), the `init-method` or
  `default-init-method` and the `destroy-method` (`(inferred)`: `close`/
  `shutdown`) or `default-destroy-method`; a `factory-method` definition
  roots the factory method. Real-JVM oracle `testdata/xmlbeans`
  (`scripts/xmlbeans-oracle.sh`, `TestXMLBeansOracle`): before, 1 of 18
  bodies the JVM ran was reached, now all 18. On JPetStore,
  `WEB-INF/applicationContext.xml` is read for the first time: `SqlSessionFactoryBean`
  and `DataSourceTransactionManager` constructors and setters become roots
  (unreachable before); no verdict changed. Namespace elements
  (`<jdbc:…>`, `<tx:…>`, `<mybatis:scan>`, `<context:component-scan>`),
  `<import>`ed files outside those places, and XML outside `WEB-INF/` and
  the class path are not read.
  Since FRAMEWORKS F5 `f5-2` (2026-11-22) the `spring` profile also roots
  what **Spring MVC** calls around a controller: `@ExceptionHandler`,
  `@ModelAttribute` and `@InitBinder` methods, every `WebMvcConfigurer`
  callback an application class overrides (`addInterceptors`,
  `addFormatters`, …) and `HandlerInterceptor`'s `preHandle`/`postHandle`/
  `afterCompletion` (`roots.overrides`; an interceptor nothing registers is
  still a root), and — new `roots.web_bodies` — the JSON bodies Jackson
  handles: the getters of a `@ResponseBody`/`@RestController` handler's
  return type and the no-argument constructor and setters of a
  `@RequestBody` parameter's type, recursively over application types
  (generic element types such as `List<Item>`, field access, builders and
  mix-ins are not modeled). Log line `spring web body roots: N methods (…)`.
  Real-JVM oracle `testdata/mvc` (`scripts/mvc-oracle.sh`, `TestMVCOracle`):
  an embedded Tomcat 9.0.45 runs a `DispatcherServlet` and real HTTP
  requests reach the controllers — before, 14 of the 26 application bodies
  the JVM ran were unreached, now all 26; an unmapped public method stays
  unreached, while a controller outside the component scan is reached (the
  profile does not read the scan). Reference apps: no verdict changed in
  12 configurations; under `-roots profile` 3 petclinic paths now start at
  the `@InitBinder` method and 2 WebGoat paths at
  `MvcConfiguration#addViewControllers` (petclinic 1, WebGoat 35 web body
  roots); `-roots all` byte-identical. Since `f5-3` a `@Profile` value is
  evaluated as a Spring 5.1+ profile expression (`!`, `&`, `|`,
  parentheses), and the report's DI resolution matches a `@Qualifier`
  against a bean's explicit or default name and its own `@Qualifier` (no
  match: unresolved instead of the `@Primary` bean); the DI oracle runs in
  two profile scenarios and checks every injected bean against the JVM.
  The profile system itself (auto-detection and merging, `-roots profile`,
  synthetic edges, the conditional jackson-databind rule,
  `roots.overrides`/`roots.annotations`/`roots.parameter_annotations`/
  `roots.handlers`/`roots.beans`/`roots.web_bodies`,
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


