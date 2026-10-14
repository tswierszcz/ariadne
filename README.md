
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

Every method is interned into a dense `uint32` id. Building the graph is
two passes:

1. **Intern every declared method** first, so a method with no outgoing
   calls at all (a getter, an empty override) still gets a node —
   `internal/reach` needs it there to use as a root.
2. **Add edges.** Every call-site edge found by `classfile` becomes a
   direct graph edge. For a `virtual` or `interface` call to `T#m`, an
   extra edge is added to `S#m` for **every transitive subtype `S` of `T`
   that itself declares `m`** — this is class-hierarchy analysis (CHA):
   the call graph can't know which subtype a given object actually is at
   run time, so it conservatively assumes it could be any of them.

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

A call to a subtype's inherited (not overridden) method still resolves
correctly: `classfile.ExtractCalls` reports a Methodref's owner exactly
as the bytecode names it, which for a call through a subclass reference
is the subclass — even when that subclass never declares the method
itself. `internal/reach`'s method matching (see below) accounts for this.

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
and its references (particularly `FIX` links to a commit).

Every HTTP request goes through `internal/httpcache`, a disk cache keyed
by a hash of the request. `-offline` makes a cache miss a hard, clear
error instead of silently reaching the network.

### 6. CVE → vulnerable methods (`internal/vulnmap`)

For each CVE, in priority order:

1. **The manual table** (`internal/vulnmap/vuln_methods.yaml`, embedded
   in the binary; override with `-vuln-methods`). Matched by the CVE's
   own id or any of its aliases. This always wins when present.
2. **A GitHub fix-commit diff.** For every reference matching
   `github.com/<owner>/<repo>/commit/<sha>`, ariadne downloads
   `<url>.diff`, keeps only `.java` files under a `src/main/` path
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
3. **An LLM (opt-in, `-llm`).** Only for a CVE sources 1 and 2 found
   nothing for. The model gets the advisory (summary + details) and,
   when the CVE links a fix commit, that commit's non-test `.java` diff;
   it answers with a JSON list of methods. The reply is validated
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
| `NOT_REACHABLE` | vulnmap has methods for this CVE, but neither pass ever visits any of them. |
| `UNKNOWN` | vulnmap has no entry for this CVE at all. |
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
non-Spring, non-servlet, non-Jakarta-EE app is
`REACHABLE`/`NOT_REACHABLE`/`UNKNOWN`/`NOT_APPLICABLE` only, never
`FRAMEWORK`. `-profile spring`/`-profile servlet`/`-profile jakarta`/
`-profile generic` picks one explicitly; `-profile-file path.yaml` loads
a custom profile (same YAML shape) and adds it to the catalog, or
replaces one of the same name.

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
also carry an HTTP-verb annotation). **CDI's `@Observes` is not
implemented**: it annotates a method *parameter*, not the method or the
class, and ariadne's bytecode parser does not read
`RuntimeVisibleParameterAnnotations` at all — only class-/method-level
`RuntimeVisibleAnnotations` (see Known limitations).

### 8. The report (`internal/report`)

`report.md` is rendered from an embedded template and includes: a header
(app, timestamp, ariadne version, graph size before/after CHA, unresolved
`invokedynamic` count, dependency identification rate), a status/severity
summary, the full CVE table, every `REACHABLE` CVE's full call path with
its entry call highlighted, every `FRAMEWORK` CVE flagged for manual
review, the list of unidentified dependency jars, and a fixed Limitations
section (see below).

`results.json` is the same analysis as structured data — sorted by CVE
id, with each result's own method list sorted — so that two runs against
the same cached inputs produce **byte-for-byte identical output**
(verified: see [PLAN.md](PLAN.md)'s Phase 6 notes for a real
non-determinism bug this guarantee caught and how it was fixed).

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
| `-app` | yes | A directory, a plain jar, or a Spring Boot fat jar. |
| `-deps` | no* | Directory of dependency jars. *Not needed for a fat jar, which supplies its own via `BOOT-INF/lib/`.* |
| `-deps-list` | no | `mvn dependency:list` output, used as a fallback for jars with no `pom.properties`. |
| `-vuln-methods` | no | Override the embedded `vuln_methods.yaml`. |
| `-profile` | no | `auto` (default, auto-detect and merge), `generic`, `spring`, or a name from `-profile-file`. |
| `-profile-file` | no | Load/override one profile from a YAML file (same shape as the embedded defaults). |
| `-roots` | no | `all` (default: every application method) or `profile` (the active profile's own entry points). |
| `-out` | yes | `report.md` output path. |
| `-json` | yes | `results.json` output path. |
| `-cache-dir` | no | Disk cache directory for OSV/GitHub HTTP responses (default `.cache`). |
| `-offline` | no | Serve only from the cache; a miss is a clear, immediate error. |
| `-llm` | no | Enable the LLM source for CVEs the manual table and fix-commit diffs could not map. Needs a model and `ARIADNE_LLM_API_KEY`. |
| `-llm-model` | no | LLM model id; overrides `ARIADNE_LLM_MODEL` (e.g. `z-ai/glm-5.3`). |
| `-llm-base-url` | no | OpenAI-compatible base URL; overrides `ARIADNE_LLM_BASE_URL` (default `https://openrouter.ai/api/v1`). |
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
that directory.

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

Preparing the two target applications this project was developed and
verified against (also JDK-dependent; not needed to run or test ariadne
itself) uses the scripts in `scripts/`:

```bash
./scripts/prepare-petclinic.sh /path/to/jdk11/JAVA_HOME
./scripts/prepare-webgoat.sh   /path/to/jdk17-or-later/JAVA_HOME
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
  profile/                framework knowledge: detect, roots, framework packages, synthetic edges
    profiles/              spring.yaml, servlet.yaml, jakarta.yaml, generic.yaml (embedded defaults)
  vulnmap/                CVE -> vulnerable methods (manual table + fix-commit diffs)
    vuln_methods.yaml      the embedded manual table
  javadiff/               unified-diff parsing + the brace-aware Java method scanner
  reach/                  BFS, status assignment, path reconstruction
  report/                 Markdown/JSON rendering
    templates/
scripts/                 prepare-{petclinic,webgoat,petclinic-gradle,servlet-demo,
                         jakarta-demo}.sh, build-*fixture*.sh, gradle-copy-deps.init.gradle
                         (Gradle equivalent of mvn dependency:copy-dependencies/dependency:list)
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

- **Reflection, dynamic proxies (AOP/CGLIB, Spring Data repositories, JDK
  `Proxy`), and any `invokedynamic` site other than a lambda/method
  reference** are invisible to a static call graph. A Spring Data
  repository interface, for instance, has *no bytecode implementation at
  all* — Spring generates it at run time — so a call to one can only be
  seen reaching the interface method itself.
- **Serialization gadgets and config-driven behavior** (XML wiring,
  `web.xml`, `ServiceLoader`) are not modeled.
- **Shaded/relocated dependencies**: a jar whose classes were relocated
  under a different package than its own Maven coordinates declare is
  not detected as such.
- **CHA over-approximates**: a virtual/interface call is treated as
  reaching every subtype that declares the same-named method, whether or
  not that specific override is ever invoked for that call site at run
  time. This can only produce false positives.
- **Reflection under-approximates**: anything invoked purely via
  `java.lang.reflect`, a `MethodHandle` lookup outside a recognized
  `LambdaMetafactory` call, or dependency injection wiring beans by
  name/type at runtime is invisible, and can produce false negatives (a
  real path exists that ariadne cannot see).
- **Only four profiles exist: `spring`, `servlet`, `jakarta` and
  `generic`.** The profile system itself (auto-detection and merging,
  `-roots profile`, synthetic edges, the conditional jackson-databind
  rule, `roots.overrides`/`roots.annotations`) is implemented (see step 7
  above), but `micronaut` and `quarkus` are not — an app using one of
  those gets the `generic` fallback (no framework packages at all) unless
  it also embeds Spring or a plain servlet/JAX-RS/EJB container. There is
  also no `--max-dep-depth`. **`jakarta`'s own CDI support is
  incomplete:** `@Observes` is a method-*parameter* annotation, and
  ariadne's bytecode parser never reads
  `RuntimeVisibleParameterAnnotations` — only class-/method-level
  annotations — so a CDI observer method is not added as an entry point
  (CDI container internals are still classified `FRAMEWORK`, just not via
  an observer-method root). See [PLAN.md](PLAN.md)'s Phase 7 for what the
  rest of this would involve.
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


