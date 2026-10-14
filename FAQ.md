# FAQ

Notes explaining ariadne's design decisions, written while working through
the tool with the user. A companion to [PLAN.md](PLAN.md) (phase-by-phase
history) and [README.md](README.md) (how the pipeline works) — this file
is about *why*, not *what*.

A Polish version of this file exists as [FAQ_PL.MD](FAQ_PL.MD).

## Why is analyzing bytecode (`.class`) better than analyzing Java source?

A few reasons, starting with the most important:

### 1. Dependencies have no source code available — only bytecode

This is the decisive argument for this particular tool. The goal is to
trace a call **from the application into a library** (e.g. into
`XStream#fromXML`). You have your own application's source, but you
**almost never have your dependencies' source** in a normal build —
Maven only pulls compiled jars (`target/dependency/*.jar`), not
`-sources.jar`. To analyze the source of the whole dependency tree (87
jars for petclinic, 126 for WebGoat), you'd have to separately pull
`-sources` for each one, match versions exactly, and trust that they
downloaded correctly. Bytecode is always there — it's literally what
actually runs.

### 2. The compiler has already resolved every ambiguity for you

Java source is full of things that have to be *inferred*: which
overload of a method gets called, what a generic erases to, whether
autoboxing inserts an implicit call, whether the compiler adds a
synthetic accessor for a private field in a nested class. Doing that at
the source level would mean writing practically your own Java compiler
front end.

In bytecode, all of this is already resolved and written down
explicitly. We saw this literally while building the tool: when checking
`javap` on identical source compiled by two different compilers (JDK11
vs. JDK17), the bytecode differed (`invokespecial` vs. `invokevirtual`
for a call to the class's own private method, `StringBuilder` vs.
`invokedynamic` for string concatenation) — the compiler made the call,
the tool just reads it instead of guessing.

### 3. The format is small, stable, and precisely specified

The JVM spec (the `.class` format, the instruction set) is narrow and
has barely changed in decades — the whole parser (`internal/classfile`)
came in at a few hundred lines of Go, with no external library. The Java
language grammar is huge and keeps growing (records, `sealed`, pattern
matching, text blocks...) — parsing source robustly against all of that
is a much bigger and more fragile project.

### 4. Bytecode is the common denominator across the whole JVM

Kotlin, Scala, Groovy — all of it compiles to the same `.class` format.
Analyzing bytecode handles a mixed-language project automatically (say,
a module written in Kotlin) without writing a separate parser per
language.

### Where a different choice was made deliberately

In Phase 4 (`internal/vulnmap`, mapping a CVE to a method) the tool
analyzes **source code**, not bytecode — because there the data source
is a GitHub diff (a fix commit), which is inherently Java text, not a
compiled artifact. The general rule: **use whichever representation the
data naturally exists in**, rather than dogmatically sticking to one.

### The cost of this decision

- Bytecode is less readable for debugging — which is why unit tests kept
  cross-checking the parser's own output against `javap` as ground
  truth.
- You lose variable names and comments (irrelevant to a call graph, so
  it doesn't hurt).
- You have to write your own instruction decoder (`invokevirtual`,
  `invokeinterface`, etc.) instead of using a ready-made Java AST parser
  — which is exactly what `internal/classfile` does.

## Which projects does ariadne support, and how to prepare one before running it

**At the mechanism level:** any application compiled to JVM bytecode —
technically Java, Kotlin, Scala, Groovy, since they all compile to the
same `.class` format (see point 4 above). In practice, though, **only
five builds have actually been tested and verified**:

- **old spring-petclinic** (Spring Boot 2.4.5, a commit from before June
  2021), built with **Maven** — the realistic case: lots of CVEs, few
  actually reachable.
- **WebGoat v8.2.2**, built with **Maven** — the control case,
  deliberately vulnerable (old XStream), there to confirm the tool
  doesn't always say "safe".
- the same old spring-petclinic, built with **Gradle** instead (commit
  `4df621b`, 2021-05-10, still Spring Boot 2.4.5 — the oldest commit in
  the repo's history with a `build.gradle` at all), via
  `scripts/gradle-copy-deps.init.gradle` — see "Gradle" just below.
- a tiny hand-written, **plain Servlet (no Spring)** app
  (`scripts/prepare-servlet-demo.sh`): one `HttpServlet` calling straight
  into an old, vulnerable SnakeYAML — the control case for the `servlet`
  profile, the same role WebGoat plays for `spring`. See the `servlet`
  profile entry just below.
- a tiny hand-written **Jakarta EE (JAX-RS + EJB, no Spring)** app
  (`scripts/prepare-jakarta-demo.sh`): a `@Path`/`@GET` JAX-RS resource
  and an `@Schedule`d EJB bean, both calling the same vulnerable
  SnakeYAML — the control case for the `jakarta` profile.

Everything else (Kotlin, CDI's own `@Observes`, Micronaut, Quarkus)
**should work mechanically** (CDI's framework-package classification
aside — see the `jakarta` entry below), but has never been run or
tested.

### What's genuinely NOT supported

- **Gradle is supported** as of `scripts/gradle-copy-deps.init.gradle` +
  `scripts/prepare-petclinic-gradle.sh`, with **no changes to any Go
  package**: the init script's two tasks reproduce `mvn
  dependency:copy-dependencies`/`dependency:list` closely enough
  (`ariadneCopyDeps` copies every `runtimeClasspath` jar to a directory;
  `ariadneDepsList` writes a `deps.txt` in the exact format
  `internal/input.ParseDepsList` already parses) that `-deps`/
  `-deps-list` need no Gradle-specific code at all. Verified against
  spring-petclinic built with Gradle (see above): 80/80 dependency jars
  identified, 144 CVEs, and the same 3/3 `REACHABLE` count as the
  Maven-built baseline. One real build quirk hit and worked around (not
  upstream-fixable): that exact commit's `build.gradle` depends on a
  Gradle plugin (`wro4j`, for LESS→CSS bundling, irrelevant to bytecode
  analysis) that was only ever published on JCenter, which shut down in
  2021–2022 — the prep script drops that plugin and its CSS-bundling
  config from its own clone rather than trying to resolve a now-gone
  artifact. Full details, including why this needed no new Go test:
  [PLAN.md](PLAN.md)'s Phase 7. **Still missing:** a multi-module Gradle
  project has not been tried (the init script's tasks would need
  invoking once per subproject, same as `prepare-webgoat.sh` already does
  for Maven), and there's no Gradle equivalent of a Spring Boot fat jar's
  one-input convenience.
- **WAR and EAR are supported**, inside `internal/input` alone (no other
  package changed): a WAR (detected by a `WEB-INF/` entry) is handled
  exactly like a Spring Boot fat jar, with `WEB-INF/classes`/
  `WEB-INF/lib` in place of `BOOT-INF/classes`/`BOOT-INF/lib`; an EAR
  (detected by `META-INF/application.xml`, or by a nested `*.war` entry
  when that optional file is missing) merges every nested WAR module's
  classes/libs the same way, plus every top-level `*.jar` as a dependency
  jar. Verified by repackaging the existing Maven-built petclinic as a
  real WAR and as a real EAR wrapping that WAR: both ran through
  `bin/ariadne` unmodified and reproduced the Maven baseline exactly
  (87/87 dependencies, 148 CVEs, `REACHABLE` 3/3) — not just a synthetic
  fixture test. **Still missing:** an EAR's own top-level jars are always
  treated as dependency jars, never as app code, so an EJB module's own
  classes aren't picked up as app code (there's no reliable way to tell
  an EJB jar from a shared library jar without parsing
  `application.xml`'s module list); and a multi-WAR-module EAR has not
  actually been tried, only a single-WAR one. Full details:
  [PLAN.md](PLAN.md)'s Phase 7.
- **A plain-Servlet `servlet` profile is supported**, alongside `spring`
  and the `generic` fallback (`internal/profile`, auto-detected by
  default, or picked with `-profile`/extended with `-profile-file`).
  Covers both the pre-Jakarta-EE9 `javax.servlet.*` namespace and the
  post-EE9 `jakarta.servlet.*` one (same spec, renamed packages), with
  `HttpServlet#do*`/`Filter#doFilter`/lifecycle-listener methods as entry
  points (via the newly-wired-up `roots.overrides`, not
  `roots.annotations` — `@WebServlet` is class-level-only, unlike
  `@RequestMapping`, so it can't mark "every public method" an entry
  point the way `spring.yaml` does for `@Controller`) and
  `javax.servlet:*`/`jakarta.servlet:*`/`org.apache.tomcat:*`/
  `org.eclipse.jetty:*` as framework packages. Verified two ways: against
  both existing baselines (petclinic, WebGoat — both now auto-detect
  `spring+servlet`, since their embedded Tomcat bundles the
  `javax.servlet.*` classes too, with **no change to either app's CVE
  status counts** — confirmed, not assumed); and against the new
  `servlet`-only control app above, both with `-roots all` and with
  `-roots profile` (which narrows to exactly the one `doGet` override and
  still finds the same `REACHABLE` CVE). Full details: [PLAN.md](PLAN.md)'s
  Phase 7.
  - **Still missing:** `micronaut` and `quarkus` profiles — on an app
    using one of those, nothing in `-profile auto`'s `detect` lists
    matches, so it falls back to `generic` (no framework packages at
    all), and the core of the analysis
    (`REACHABLE`/`NOT_REACHABLE`/`UNKNOWN`) still works correctly, but
    every CVE's status stays `REACHABLE`/`NOT_REACHABLE` rather than
    being downgraded to `FRAMEWORK` where that other framework's own
    internals are driving the call, even though in practice they're
    reflection-driven just like in Spring or plain Servlet.
- **A `jakarta` profile is supported too**, for JAX-RS (the HTTP-verb
  annotations `@GET`/`@POST`/`@PUT`/`@DELETE`/`@PATCH`/`@HEAD`/`@OPTIONS`
  as entry points — deliberately not `@Path` itself, which is routinely
  class-level and would otherwise turn every public helper method of a
  resource class into an entry point, not just its actual endpoints) and
  EJB (`@Schedule`/`@Schedules`), across both the `javax.*` and
  post-Jakarta-EE9 `jakarta.*` namespaces. Verified the same two ways as
  `servlet`: against both existing baselines (petclinic, WebGoat — still
  auto-detect `spring+servlet` only, confirming `jakarta`'s own detect
  markers don't accidentally fire on a Spring app); and against the new
  JAX-RS+EJB control app above, with both `-roots all` and `-roots
  profile` (which narrows to exactly the resource's `@GET` method and the
  EJB's `@Schedule` method — confirmed directly via a one-off call to
  `Profile.AppRoots`, not just the log's entry-point count) finding the
  same `REACHABLE` CVE-2022-1471 either way.
  - **CDI's `@Observes` is NOT implemented, and this is a real gap, not
    an oversight**: it annotates a method *parameter*
    (`void onEvent(@Observes SomeEvent e)`), not the method or the class,
    and ariadne's bytecode parser never reads the
    `RuntimeVisibleParameterAnnotations` attribute at all — only
    class-/method-level `RuntimeVisibleAnnotations`. Implementing it
    would touch three packages (a new attribute parser in
    `internal/classfile`, a per-parameter-annotations field on
    `hierarchy.Method`, and a new matching rule in `internal/profile`),
    not just a YAML addition like the rest of this profile. CDI's own
    framework packages (Weld and friends) are still listed, so CDI
    container internals are classified `FRAMEWORK` on any path that
    reaches them — just not via an observer-method root. Full details:
    [PLAN.md](PLAN.md)'s Phase 7.

### How to prepare a project

Ariadne analyzes **compiled bytecode**, so the project has to be built
first with a standard tool (Maven):

```bash
# 1. Build the application
mvn package -DskipTests

# 2. Copy runtime dependencies to a separate directory
mvn dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory=target/dependency

# 3. (optional, but recommended) generate a dependency list as an identification fallback
mvn dependency:list -DincludeScope=runtime -DoutputFile=target/deps.txt
```

Step 3 matters in practice — on petclinic, `pom.properties` alone (from
inside the jars) identified only 42/87 dependencies; with `deps.txt` as
a fallback, 87/87. Without that list, some jars end up "unidentified" in
the report and never get checked against any CVE at all.

**Alternative — a Spring Boot fat jar:** if the build produces one fat
jar (`target/myapp-1.0.0.jar`), you can just point ariadne at that alone,
no `-deps`/`-deps-list` needed — ariadne pulls `BOOT-INF/classes` and
`BOOT-INF/lib/*.jar` out itself. Downside: dependency identification is
then lower (only from `pom.properties`, since there's no `deps.txt` to
fall back on), so this mode finds fewer CVEs than the `-deps-list` mode.

**Alternative — a Gradle build:** the equivalent three steps use
`scripts/gradle-copy-deps.init.gradle` in place of Maven's
`dependency:copy-dependencies`/`dependency:list`:

```bash
./gradlew classes
./gradlew -I /path/to/ariadne/scripts/gradle-copy-deps.init.gradle \
  ariadneCopyDeps ariadneDepsList \
  -PariadneOutputDir=build/ariadne-dependency \
  -PariadneDepsListFile=build/deps.txt
```

then `-app build/classes/java/main -deps build/ariadne-dependency
-deps-list build/deps.txt`, same as the Maven case. See
`scripts/prepare-petclinic-gradle.sh` for a complete, working example
(including a real build quirk it has to work around for that specific
target).

**What you do NOT need on the machine running `ariadne`:** a JDK. It's
only needed to build the application being analyzed — `ariadne` itself
parses `.class` files with its own parser.

## How the pipeline works, step by step

### Step 1: `internal/input` — normalizing the input

The job: turn whatever was given in `-app`/`-deps`/`-deps-list` into two
uniform things the rest of the program never has to tell apart again:
one path to the app's jar, and a `[]input.DepJar` list (each entry is
`{Path, GAV, Source, Warnings}`).

**`PrepareAppInput(appPath)`** detects what it's looking at:

- a **directory** (`target/classes`) → `packDirAsJar`: walks the tree
  (`filepath.Walk`), adds every file to a new ZIP archive keeping its
  relative path, writes it to a temp file;
- a **plain jar** → used as-is, no copying;
- a **Spring Boot fat jar** (detected by an entry starting with
  `BOOT-INF/`), a **WAR** (`WEB-INF/`), or an **EAR** (`META-INF/
  application.xml`, or a nested `*.war` entry when that optional file is
  missing) → `prepareFromFatArchive`/`prepareFromEar`: rewrites
  `BOOT-INF/classes/**` or `WEB-INF/classes/**` into a new jar **stripping
  the prefix** (without that, the `.class` parser would look for the
  class under the wrong path), and extracts every `*/lib/*.jar`
  byte-for-byte into its own file — as if someone had run `mvn
  dependency:copy-dependencies` by hand. An EAR repeats this once per
  nested WAR module (opened as its own `zip.Reader` read fully into
  memory first, since a zip entry can only be opened as a one-shot
  stream) and additionally pulls every top-level `*.jar` in as a
  dependency jar.

**`ResolveDepJars(depsDir, depsList)`** for every dependency jar, in this
order:

1. Look inside the jar itself — every jar from Maven Central has
   `META-INF/maven/<groupId>/<artifactId>/pom.properties` inside it.
   Works regardless of the build tool. A complication found on
   petclinic: some jars (e.g. `ehcache-3.9.2.jar`) have **several** such
   files (their own bundled submodules) — the one whose
   `artifactId-version` matches the jar's own file name wins, and the
   rest are recorded as a warning.
2. If the jar has no `pom.properties`, match the file name
   (`artifactId-version[-classifier].jar`) against the `-deps-list`
   (`ParseDepsList` has to strip ANSI color codes and the
   ` -- module x [auto]` suffixes that newer `mvn dependency:list`
   versions litter the output with, on the way).
3. If nothing worked, the jar goes on the "unidentified" list in the
   report — it is not silently dropped.

### Step 2: `internal/classfile` — parsing bytecode

Input: the raw bytes of one `.class` file. Output: the class's structure
(superclass, interfaces, methods, annotations) plus, for every method, a
list of its call sites.

**`Parse()`** reads the file sequentially per the JVM spec: the magic
number `0xCAFEBABE`, the bytecode version (collected for the report as
"highest version seen"), the **constant pool** (a table of everything
the class refers to — class names, method signatures, literals,
`MethodHandle`s; every entry type through Java 21; a `Long`/`Double`
entry takes two slots, not one — a spec quirk), access flags/superclass/
interfaces, methods (name, signature, `RuntimeVisibleAnnotations` parsed
recursively, the raw `Code` bytes), and class attributes, especially
`BootstrapMethods` (needed for lambdas — see below).

**`ExtractCalls()`** walks a method's bytecode instruction by
instruction — each instruction is an opcode byte plus an operand whose
length depends on the specific instruction (a table for ~200 opcodes,
with special cases for `tableswitch`/`lookupswitch` and the `wide`
prefix). Four kinds of calls matter: `invokevirtual`, `invokespecial`,
`invokestatic`, `invokeinterface` — all resolved straight from the
constant pool.

**`invokedynamic`** (lambdas, method references, or on newer bytecode
string concatenation) can't be read directly — the real target is
hidden inside `BootstrapMethods`. If the bootstrap is
`LambdaMetafactory`/`altMetafactory`, the program reads its second
argument (a handle to the real implementation) and turns it into a
normal graph edge. If the bootstrap is something else (e.g.
`StringConcatFactory`), the program **deliberately doesn't guess** —
it counts it as "unresolved" and reports it that way.

Every part of this was checked by hand against `javap -c -p -v` on the
same file compiled for Java 8 and Java 21. On real data: 29154 classes
from petclinic and 40811 from WebGoat parsed with zero errors.

### Step 3: `internal/hierarchy` — the class hierarchy

Input: the app jar + `[]input.DepJar` from step 1. Output: a `Hierarchy`
structure that knows every class read, from the app and from the
dependencies.

For every jar, for every `.class` entry (skipping `module-info.class` and
`META-INF/versions/**` — multi-release jars): parse via
`classfile.Parse`, and **right there** call `ExtractCalls` for every
method — because `classfile.ClassFile` deliberately doesn't expose the
constant pool or `BootstrapMethods` outside the package, so this is the
only moment they're still available. This runs in parallel (8
goroutines); results land in a `FQCN → *ClassInfo` map.

**Classpath conflicts:** if the same class exists in two different jars
(a real case on petclinic: `jakarta.activation` vs.
`jakarta.activation-api`), the program doesn't error out — the first one
read wins, and the second goes into the warnings.

**Inverting the inheritance relation:** the program builds a
`direct: supertype → list of subtypes` map, walking every class in
**sorted** name order (not in Go's randomized map iteration order — a
fix for a real bug: without the sort, two runs of the exact same program
on the exact same data produced different, though equally valid, paths
in `results.json`).

**`Subtypes(fqcn)`** returns every transitive subtype (a BFS over the
`direct` map, memoized in `subtypesCache`, since the same type can be
asked about thousands of times in one run). A type that was never read
(typically something from the JDK, e.g. `java.lang.Runnable`) simply
returns empty — this is exactly the mechanism that protects step 4 from
a graph explosion.

**`ResolveMethod(fqcn, name)`** is implemented and tested (it replays the
JVM's own method lookup order: the class, then its superclasses, then
interfaces), but **no other package calls it yet** — a ready building
block waiting for a later phase, not part of today's working pipeline.

### Step 4: `internal/callgraph` — building the call graph and CHA

Input: the `Hierarchy` from step 3 — every class already has its methods
with their calls (`Calls`) extracted. Output: `Graph` — a directed graph
where nodes are methods and edges are "who calls whom", extended with
the polymorphic calls bytecode alone doesn't fully show.

**Interning.** Every method has a text key,
`"fqcn#name(signature)"`. Holding millions of edges as pairs of strings
would be slow and memory-heavy, so every key gets one small number
(`uint32`) once, and from then on the whole graph operates on numbers. A
second index is also built: `"fqcn#name"` (no signature) → every method
with that name regardless of overload — needed in step 7 to match a
vulnerable method from `vuln_methods.yaml`, which never gives an exact
signature.

**Two build passes.** Pass 1: intern every declared method, even one
with no calls at all — step 7 needs a starting point for every app
method as a potential BFS root. Pass 2: add edges — straight to the
resolved target for `invokespecial`/`invokestatic`/`invokedynamic`; to
the bytecode's own target **and** to the CHA expansion for
`invokevirtual`/`invokeinterface`.

**CHA (class-hierarchy analysis).** Bytecode records an interface call
exactly as the compiler saw it statically — at run time it could be any
subtype. The static graph doesn't know which, so it plays it safe: it
adds an edge to **every** subtype (from `hierarchy.Subtypes()`) that
itself declares that method.

**Guarding against explosion.** CHA never expands calls on
`java.*`/`javax.*`/`jdk.*`/`sun.*` types (`DefaultSkip`) — without this
guard, the framework's own bootstrap alone would pull in almost the
whole classpath. Measured on petclinic: with the guard, the graph has
1.2M edges and builds in 1.9s; without it — 11.4M edges (9.3×) and 2m43s
(85× slower). Without it, practically every CVE in every dependency
would come out "reachable".

Both passes walk the classes in **sorted** name order (not Go's
randomized map iteration order) — the other half of step 3's determinism
fix: without it, two runs on the same data produced different, though
equally valid, paths in `results.json`.

Numbers: petclinic — 273,180 nodes, 1,229,600 edges; WebGoat — 339,786
nodes, 2,591,404 edges.

### Step 5: `internal/cves` — querying OSV

Input: the unique `(groupId:artifactId, version)` pairs from step 1.
Output: a CVE list, each one deduplicated, with a summary, severity,
which dependencies it affects, and references (especially links to fix
commits).

**Batch query.** `POST /v1/querybatch` on `api.osv.dev` takes many
queries at once (a 1000/request limit; larger lists get chunked). The
response for each package+version pair is just a **list of ids**, no
details — deliberately lightweight. Pagination is also handled
(`/v1/query` + `page_token`), though it's never actually been needed in
practice.

**Details.** For the set of unique ids, `GET /v1/vulns/{id}` — in
parallel, 8 at a time.

**Alias deduplication.** OSV often has two separate records for the same
vulnerability (one under a GHSA id, one under a CVE id, each pointing at
the other via its `aliases` field). Union-find merges them into one
group, with the `CVE-*` id as the primary one when there is one.

**Severity.** The `severity[].score` field for `CVSS_V3` is the whole
CVSS vector as text, not a number — the first version of the report
showed this as a table of ~30 nearly-unique rows. Fix: the program uses
`database_specific.severity` — a readable label (`LOW`/`HIGH`/
`CRITICAL`...) that OSV fills in for GitHub Advisory Database-sourced
vulnerabilities — falling back to the CVSS vector only when that label
is missing.

Every request goes through `internal/httpcache`; with `-offline`, a
cache miss is a clear error, not a silent skip.

Numbers: petclinic — 148 CVEs for 87 dependencies; WebGoat — 160 CVEs
for 126 dependencies (including 34 for `xstream:1.4.5` alone).

### Step 6: `internal/vulnmap` — CVE → specific methods

Input: the CVE list from step 5. Output: a map of CVE → `{Methods,
Source}` — for every CVE that got mapped, a concrete list of vulnerable
methods. A CVE with no entry here gets status `UNKNOWN` in step 7.

**Source 1 — the manual table** (`internal/vulnmap/vuln_methods.yaml`,
embedded via `//go:embed`, overridable with `-vuln-methods`). Matched by
the CVE's own id **and** by every alias. Always wins when it has an
entry.

**Source 2 — a GitHub fix-commit diff.** For a reference shaped like
`github.com/<owner>/<repo>/commit/<sha>`: download `<url>.diff`, keep
only `.java` files under `src/main/` (excluding tests), for each one
fetch its **full content** as it existed at the fix commit
(`raw.githubusercontent.com`), and use a scanner (`internal/javadiff`) to
find which method encloses each changed line. The scanner first blanks
out comments/strings/text blocks (keeping line numbers intact) so a
stray `{` never desyncs brace matching, then tracks a stack of
class/method/"something else" frames (`if`/`for`/lambda/anonymous-class
bodies are transparent), and renames a method to `<init>` when its name
matches the enclosing class (a constructor).

Source 3 is an LLM reading the advisory + fix diff, opt-in with `-llm`
(design and exact prompt: [docs/phase4-llm.md](docs/phase4-llm.md)). It
runs only for CVEs sources 1 and 2 left unmapped, and each method it
names is checked to exist in the analyzed bytecode. When it saw only the
advisory text (no diff), a `NOT_REACHABLE` result is downgraded to
`UNKNOWN`, since the model may have named a real method that is not the
vulnerable one.

Numbers: petclinic — 91/148 mapped (1 from the manual table, 90 from a
diff); WebGoat — 104/160 (37 from the manual table — the whole XStream
family, 67 from diffs). Source 2's file filter keeps any non-test `.java`
file rather than requiring a Maven `src/main/` layout, which is what
raised petclinic's diff-mapped count: Apache Tomcat lays its code out as
`java/org/apache/...`, not `src/main/java/...` (see [PLAN.md](PLAN.md)'s
Phase 4 follow-up).

### Step 7: `internal/reach` — searching the graph and assigning statuses

Input: the graph from step 4, the hierarchy from step 3, the CVEs from
step 5, the mapping from step 6. Output: one of four statuses per CVE,
and for `REACHABLE`/`FRAMEWORK`, a concrete path.

**Roots:** every method in the application's own code (`IsApp == true`),
no exceptions — private methods included.

**Two BFS searches, computed once** (not per CVE — which is why
analyzing 148-160 CVEs takes under 100ms even on a 2.6M-edge graph):

- **Search 1 (REACHABLE):** a node from a "framework" package (Spring
  Web/WebMVC, Spring Boot, Tomcat, Jackson-databind, Hibernate, Logback,
  Log4j) is marked visited, but the program **never expands** its own
  outgoing edges — the framework is a dead end you're allowed to reach
  in one hop, but not allowed to trace further through.
- **Search 2 (FRAMEWORK):** the same roots, no such restriction.

This one trick ("mark visited, but don't expand") implements exactly the
exception from the plan: a direct call from app code into a framework
method still counts as REACHABLE, but a path that needs several hops
*through* the framework doesn't.

**Matching a YAML method to graph nodes** (`matchVulnerableMethod`):
look up nodes by class+method name directly (ignoring overloads), **and
additionally** check every transitive subtype that doesn't override that
method — because a call through a subclass reference records the
subclass's name in bytecode even when the method is only inherited, not
overridden.

**Status for a CVE:** first, if every affected dependency is a WebJar
(no `.class` file at all, assets under `META-INF/resources/webjars/`) →
`NOT_APPLICABLE`: client-side JavaScript/CSS a call graph cannot analyze,
which is not the same as safe. (A merely empty jar does not count: Maven
aggregators like `spring-boot-starter-web` are empty too, and Spring4Shell
is filed against exactly that one.) Then: no entry from step 6 → `UNKNOWN`. Otherwise:
matched nodes visited in search 1 → `REACHABLE` (the shortest such
path). Visited only in search 2 → `FRAMEWORK`. Neither → `NOT_REACHABLE`.

**Path and "entry call":** the BFS remembers, for every node, which node
it was first reached from (a parent pointer) — reconstructing a path is
just walking from the target back to a root and reversing it. The "entry
call" is the first edge on that path where the caller is app code and
the callee isn't — the single most useful line for a human reviewing the
result, since it shows exactly where in your own code the problem
starts.

Numbers (current, under `-profile spring`, no `-llm`): petclinic — 3
REACHABLE / 59 FRAMEWORK / 29 NOT_REACHABLE / 51 UNKNOWN / 6
NOT_APPLICABLE (of 148). WebGoat (control case) — 40/43/21/56 (of 160),
with the whole XStream family still REACHABLE via the two-hop path
`VulnerableComponentsLesson#completed → XStream#fromXML`. (Before the
`spring` profile's `framework_packages` list added `spring-beans`/
`spring-core`/`spring-expression`/`spring-aop` — see
[PLAN.md](PLAN.md)'s Phase 5 follow-up note — these were 13/18/25/92 and
51/32/21/56: several paths that only reach a vulnerable method by
continuing through Spring Beans/SpEL internals correctly moved from
REACHABLE to FRAMEWORK.)

### Step 8: `internal/report` — assembling the result

Input: everything produced so far. Output: `report.md` (for a human) and
`results.json` (for a machine). This step computes nothing — it only
lays out already-finished data in readable form.

**Preparing the data.** Results are sorted by CVE id, and each result's
own method list too — this is exactly the mechanism that guarantees two
runs on the same data produce an identical output file. What gets
computed: status counts, severity counts (bucketed by the readable label
from step 5), table rows (one per CVE+dependency combination), and
separate `REACHABLE`/`FRAMEWORK` lists for the full-path sections.

**Rendering.** The Markdown template is embedded (`//go:embed`,
`templates/report.md.tmpl`), filled in via `text/template` — all the
logic happened earlier, the template just formats. An HTML version
(`-html`) was an option in the plan but was never built.

The report contains: a header, a summary (statuses + severity), the full
CVE table, a full-path section for every `REACHABLE` CVE (with the
"entry call" highlighted), a `FRAMEWORK` section labeled "needs manual
review", the list of unidentified jars, and a fixed Limitations section
(reflection, dynamic proxies, serialization gadgets, CHA
over-approximation — the same content as in the README).

`results.json` is the same list of results as indented JSON, with no
rendering logic — easy to diff between runs or load into another tool.

## Can the call graph be saved/persisted anywhere?

**No — it isn't saved anywhere today.** It only lives in the RAM of one
`ariadne` run and is discarded when the process exits.
`internal/callgraph` has no save/load function at all (no JSON tags, no
`Save`/`Load`).

What actually does get saved: only the final analysis result
(`results.json` — status, methods, path per CVE) and the HTTP response
cache for OSV/GitHub (`.cache/`). Neither of those is the graph itself.

**Why nobody saves it:** the graph is deterministically reproducible
from the same inputs, and building it takes about 2 seconds even for
270K nodes and 1.2M edges (measured on petclinic) — like a compiler's
intermediate representation: compute it, use it, forget it.

**Could it be saved — technically, yes,** since `Graph` is plain data
inside (a list of strings, a string→int map, an array of edge lists).
Two options: `encoding/gob` (compact, binary, readable only by Go) or
JSON (readable by a human/other tools, but at 1.2M edges the file would
run tens to hundreds of MB). It would make sense if you needed to: (1)
run several analyses against the same build without rebuilding the
graph each time, (2) build a visualization of the graph in a separate
tool (exporting to `.dot`), or (3) debug the graph itself without going
through the whole pipeline. None of these have come up yet, so it hasn't
been implemented.
