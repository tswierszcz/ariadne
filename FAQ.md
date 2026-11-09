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
seven builds have actually been tested and verified**:

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
- a tiny hand-written **Jakarta EE (JAX-RS + EJB + CDI, no Spring)** app
  (`scripts/prepare-jakarta-demo.sh`): a `@Path`/`@GET` JAX-RS resource,
  an `@Schedule`d EJB bean, and a CDI observer method
  (`void onEvent(@Observes String payload)`), all three calling the same
  vulnerable SnakeYAML — the control case for the `jakarta` profile,
  including CDI's `@Observes`.
- a tiny hand-written **Micronaut (no Spring)** app
  (`scripts/prepare-micronaut-demo.sh`): one `@Controller`/`@Get` method
  calling the same vulnerable SnakeYAML — the control case for the
  `micronaut` profile. Pinned to Micronaut 3.9.7, since the current
  4.x/5.x line's jars are compiled for a newer Java than this project's
  demo scripts assume (a `javac` constraint on the demo app, not a limit
  of ariadne's own bytecode parser).
- a tiny hand-written **Quarkus** app (`scripts/prepare-quarkus-demo.sh`):
  a `jakarta.ws.rs` `@Path`/`@GET` REST resource, an
  `io.quarkus.scheduler` `@Scheduled` job, and a MicroProfile Reactive
  Messaging `@Incoming` consumer, all three calling the same vulnerable
  SnakeYAML — the control case for the `quarkus` profile. Needs JDK 17+.
- a tiny hand-written **JPMS** app (`scripts/prepare-jpms-demo.sh`): one
  module whose only ServiceLoader provider is declared in `module-info`
  `provides` and created through `public static provider()`, which calls
  commons-text 1.9's `StringSubstitutor#replace` (CVE-2022-42889). It is
  the control case for `-jpms-scenario`: with `-profile generic -roots
  profile`, `classpath` gives NOT_REACHABLE (the JVM ignores module-info
  there) and `module-path` gives REACHABLE through `provider()`; the script
  also launches it with the real JVM (0 vs 1 provider found), plus two mixed
  launches for `-jpms-scenario mixed`: commons-lang3 alone on the class path
  starts and finds the provider (REACHABLE in ariadne too), commons-text on
  the class path fails at startup (`Module org.apache.commons.text not
  found`), which ariadne reports as a `requires` warning.

Everything else (Kotlin, Scala, Groovy) **should work mechanically**,
but has never been run or tested.

### Build tools and profiles: what's supported, and what's still missing

Everything in this list started life as "not supported"; each bullet below records what
actually got built, how it was verified, and what — if anything — is still missing from
that one piece.

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
- **A `jakarta` profile is supported too**, for JAX-RS (the HTTP-verb
  annotations `@GET`/`@POST`/`@PUT`/`@DELETE`/`@PATCH`/`@HEAD`/`@OPTIONS`
  as entry points — deliberately not `@Path` itself, which is routinely
  class-level and would otherwise turn every public helper method of a
  resource class into an entry point, not just its actual endpoints) and
  EJB (`@Schedule`/`@Schedules`), across both the `javax.*` and
  post-Jakarta-EE9 `jakarta.*` namespaces — plus **CDI's `@Observes`**,
  which needed real new code, not just YAML: `@Observes` annotates a
  method *parameter*
  (`void onEvent(@Observes SomeEvent e)`), not the method or the class,
  so it needed a new kind of root (`roots.parameter_annotations`) and a
  parser for the `RuntimeVisibleParameterAnnotations` attribute in
  `internal/classfile`, which ariadne didn't read at all before (only
  class-/method-level `RuntimeVisibleAnnotations`). Turned out smaller
  than first estimated: the new parser reuses the existing
  `RuntimeVisibleAnnotations` parser's own `skipAnnotation` helper
  (~15 lines), and `hierarchy.Method` already embeds `classfile.Method`
  directly, so the new field needed zero changes there. Verified the same
  two ways as `servlet`/`micronaut`: against both existing baselines
  (petclinic, WebGoat — still auto-detect `spring+servlet` only,
  confirming `jakarta`'s own detect markers don't accidentally fire on a
  Spring app); and against the control app above (now three classes: the
  JAX-RS resource, the EJB bean, and a CDI observer method), with both
  `-roots all` and `-roots profile` (which narrows to exactly those three
  methods — confirmed directly via a one-off call to `Profile.AppRoots`,
  not just the log's entry-point count, and the observer method's own
  path is the one the report records) finding the same `REACHABLE`
  CVE-2022-1471 either way. Full details: [PLAN.md](PLAN.md)'s Phase 7.
- **A `micronaut` profile is supported too**, for HTTP routing
  (`@Get`/`@Post`/`@Put`/`@Delete`/`@Patch`/`@Head`/`@Options`) and
  `@Scheduled`/`@EventListener` as entry points — again deliberately not
  `@Controller` itself (confirmed via `javap` to be class-level-only,
  `@Target(TYPE, ANNOTATION_TYPE)`), for the same reason `@Path` is
  excluded from `jakarta`. Also confirmed via `javap`: Micronaut's own
  annotations are real `@Retention(RUNTIME)`, not a compile-time-only
  marker ariadne's parser would miss — worth checking given Micronaut's
  whole design philosophy is avoiding runtime reflection elsewhere.
  Verified the same two ways: against both existing baselines
  (unaffected); and against a new Micronaut-only control app
  (`scripts/prepare-micronaut-demo.sh`, pinned to Micronaut 3.9.7 since
  the current line's jars need a newer JDK than this project's demo
  scripts assume), with both `-roots all` and `-roots profile` (narrows
  to exactly the one `@Get` method) finding the same `REACHABLE`
  CVE-2022-1471. **No `synthetic_edges` needed, unlike `spring`'s Spring
  Data bridge**: a Micronaut Data repository's implementation is
  generated by Micronaut's own annotation processor *at compile time*, as
  a real class in the app's own output — not a Spring-Data-style runtime
  proxy with zero bytecode — so CHA already finds it like any other
  subtype. Full details: [PLAN.md](PLAN.md)'s Phase 7.
- **A `quarkus` profile is supported too** — the last of PLAN.md section
  4's originally-named Phase 7 frameworks. Quarkus's own REST layer is
  plain JAX-RS, already covered by `jakarta.yaml`: a Quarkus REST app
  auto-detects as `jakarta+quarkus`, the same way a Spring Boot app
  auto-detects as `spring+servlet`, so this profile only adds what's
  genuinely Quarkus-/MicroProfile-specific — `@Scheduled`
  (`io.quarkus.scheduler`) and reactive messaging's `@Incoming`/
  `@Outgoing` (`org.eclipse.microprofile.reactive.messaging`) — as entry
  points, both confirmed via `javap` to be real
  `@Target(METHOD)`/`@Retention(RUNTIME)`, the same diligence applied to
  Micronaut's annotations. Verified against both existing baselines
  (unaffected) and against a new, real, three-class Quarkus control app
  (`scripts/prepare-quarkus-demo.sh`, needs JDK 17+): auto-detects
  `jakarta+quarkus` (confirming both profiles' detect markers fire on the
  same app and merge correctly), and `-roots profile` narrows to exactly
  the three annotated methods (confirmed directly via `Profile.AppRoots`,
  not just the log's count), finding the same `REACHABLE` CVE-2022-1471
  either way. No `roots.overrides`/`synthetic_edges`: Quarkus, like
  Micronaut, builds its wiring at build time via Jandex indexing and
  bytecode recording, not runtime proxies with zero bytecode.

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

See [docs/SUPPORT.MD](docs/SUPPORT.MD) for a mechanism-by-mechanism current-status table
and [RECOMMENDATION](RECOMMENDATION/README.MD) for plans to extend what's still missing
above — this section is a narrative companion to those, not a separately maintained list.

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

   **Why confirm an embedded library by its classes (CVE:F6a)?** Those
   extra `pom.properties` are claims, not facts: in `ehcache-3.9.2` five
   are ehcache's own modules (one more is the jar's own identity), two (`org.terracotta:offheap-store`,
   `statistics`) describe code that was relocated to
   `org.ehcache.shadow` — no class of theirs exists under their own
   names — and two (`org.ehcache:sizeof`, `terracotta-utilities-tools`)
   really are there. Looking advisories up for all of them would invent
   findings; ignoring them hides real libraries. So a library candidate
   is checked against its jar of exactly that version from Maven
   Central, class by class. Byte comparison alone is not enough: shading
   rewrites class files even when it doesn't rename them (constant-pool
   order changes every index) — `jnr-ffi` inside `jruby-core` had 0 of
   656 byte-identical classes and 644 identical by the
   serialization-independent fingerprint. Only a confirmed library gets
   an instance, owns its classes and gets looked up; everything else
   stays visible as a coverage gap.

   **Why only explicit relocation (CVE:F7a)?** A relocated copy has
   none of its original class names, so "is it there, in which
   version?" needs the map from old to new package. Where the map is
   stated — jruby-core ships its own `pom.xml` whose maven-shade
   configuration says `me.qmx.jitescript` → `org.jruby.me.qmx.jitescript`
   and `org.objectweb` → `org.jruby.org.objectweb` — ariadne uses it
   (and with it jnr-ffi's last 12 classes, which only differed by their
   references to the relocated ASM, match too: 656 of 656). Where it is
   not (ehcache is built with Gradle), the user can state it in an
   `-embedded-components` file pinned to the jar's hash; ariadne still
   checks every claim against the classes. Guessing relocations from
   code similarity would be a heuristic without a version, so it is left
   to a later, experimental stage.

   **Why read the build configuration and an SBOM, and why candidate
   versions (CVE:F6b)?** jruby-core relocates ASM but ships no
   `pom.properties` for it — yet its own `pom.xml` says which artifacts
   the shade plugin copies (`org.ow2.asm:*`) and depends on asm,
   asm-commons and asm-util at `${asm.version}`; the parent pom says
   9.0, and asm-commons' own pom adds asm-tree and asm-analysis. All
   five, 140 classes, match ASM 9.0 exactly. An SBOM is used the same
   way, but only where it is tied to the input by a jar's SHA-256: a
   list of dependencies says what the build used, not what is inside a
   jar. Metadata can also name the wrong version. Then the classes are
   compared with neighbouring releases instead of trusting the label —
   and when several releases are indistinguishable (asm-util did not
   change between 9.0-beta and 9.0), ariadne keeps both as candidates and
   reports every candidate's advisories as `undetermined` rather than
   picking one.
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
recursively, `RuntimeVisibleParameterAnnotations` — needed for CDI's
`@Observes`, a *parameter*-level annotation — parsed the same way, the
raw `Code` bytes), and class attributes, especially `BootstrapMethods`
(needed for lambdas — see below).

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

That edge goes from the method that **creates** the lambda. When the
lambda is run somewhere else — handed to a helper, returned, kept in a
field and later called as `supplier.get()` — the calling method needs an
edge too, or a path through it is lost. Since builder version `4`
(DYNAMIC_LINKAGE L1, `internal/samcall`) every call of a functional
interface's method asks DATA_FLOW where its receiver came from (the
method's callers, values a lambda captured, a field's writers, a helper's
return) and links the site to the lambdas found (`sam` edges). It never
guesses "every lambda of `Supplier`": what it cannot follow is a counted
gap with no edge. And it only adds an edge once a method creating that
lambda is itself reachable — otherwise a helper called from many places
would "run" lambdas whose creators never execute. Why bother? A recorded
petclinic run took 50 such relations the graph did not have; now 22
remain (most of the rest are not lambdas at all, but JDK collections
calling `equals`/`hashCode`).

Every part of this was checked by hand against `javap -c -p -v` on the
same file compiled for Java 8 and Java 21. On real data: 29154 classes
from petclinic and 40811 from WebGoat parsed with zero errors.

### Step 3: `internal/hierarchy` — the class hierarchy

Input: the app jar + `[]input.DepJar` from step 1. Output: a `Hierarchy`
structure that knows every class read, from the app and from the
dependencies.

For every jar, for every `.class` entry (skipping `module-info.class` — module
descriptors are read separately, as an inventory only, by
`serviceload.InventoryModules` — and
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
itself declares that method — and, for a concrete subtype that only
inherits it, to the body the JVM would select for it, even when that body
sits in a superclass outside the interface's hierarchy (Spring's
`DefaultListableBeanFactory` implements `ConfigurableListableBeanFactory`
but inherits `getBean` from `AbstractBeanFactory`). Until 2026-11-08 that
last case was missed; a runtime trace measured it (17 relations petclinic
really took, `testdata/graphbench/TRACE_GAPS.MD`).

**Guarding against explosion.** CHA never expands calls on
`java.*`/`javax.*`/`jdk.*`/`sun.*` types (`DefaultSkip`) — without this
guard, the framework's own bootstrap alone would pull in almost the
whole classpath. Measured on petclinic: with the guard, the graph has
1.2M edges and builds in 1.9s; without it — 11.4M edges (9.3×) and 2m43s
(85× slower). Without it, practically every CVE in every dependency
would come out "reachable".

**But then how does a request ever reach `DispatcherServlet`?** Through
one bounded exception. The container hands every request to the app with
`servlet.service(req, res)` and `filter.doFilter(req, res, chain)` — calls
on `javax.servlet.Servlet`/`Filter` (or `jakarta.servlet.*`), which the
guard above never expands, so without help the whole Spring MVC stack is
cut off from `main`. Lifting the guard for these two interfaces would fan
out to every servlet and filter class in every jar (dozens nobody
registers). Instead, the `servlet-container:dispatch` model
(`internal/servletdispatch`) adds edges from those calls — and from
`HttpServlet`'s own `this.service(...)`/`this.doGet(...)`/... template calls
— only to the servlets and filters the app actually registers: `web.xml`,
`@WebServlet`/`@WebFilter`, and in a Spring Boot application the beans
that are servlets or filters plus a short, source-checked catalog of what Spring Boot's auto-configuration
registers (`DispatcherServlet`, a few standard filters, Spring Security's
`DelegatingFilterProxy`/`FilterChainProxy`), each with its basis in the
log. What it cannot name (a `@Bean` returning a `FilterRegistrationBean`)
is reported as an incomplete model fact, not guessed. Its real-JVM check is
`testdata/servletdispatch`; its effect on petclinic and WebGoat is in
[testdata/graphbench/RESULTS.MD](testdata/graphbench/RESULTS.MD).

**Why doesn't a runtime trace change any verdict?** (RUNTIME_NATIVE N1,
`-runtime-trace`.) A trace is what one run did under one scenario: it can
show that a vulnerable method ran, but not that it can't run — traffic,
configuration, timing and sampling decide what a run shows. Turning
"not seen" into `NOT_REACHABLE` would make a short or narrow test run look
like proof. And "seen executing" is still not "exploitable": it says the
code ran, not that attacker data got there. So the trace stays evidence
next to the static result. Where the two disagree — petclinic's
CVE-2025-35036 runs although the graph says `NOT_REACHABLE` — the trace
points at a static gap to model (there, a `javax.validation` SPI call the
graph never expands), it doesn't overwrite the verdict. Traces are matched
to the analyzed classes by content hash, so a trace from another build of
the code is reported but not used.

**Why is RTA opt-in, and why doesn't it remove more?** (GRAPH_PRECISION
P2, `-callgraph-mode rta`.) RTA keeps a CHA edge only if some class that
reachable code creates dispatches the call there — but a class can also
be created by mechanisms the analysis does not see: reflection with a
computed name, `Unsafe`, a framework instantiating a DTO or an entity.
Dropping those targets silently would turn a false positive into a false
negative, which is worse for this tool. So every such site, once
reachable, keeps a justified fallback set of classes alive (classes the
code can name at run time, annotated classes, member types, registered
service providers, root parameter types with `-roots all`) and is
reported as a gap; the report says how many edges were kept only because
of that fallback. With `-roots all` every method is a root, so almost
every class is created by reachable code and RTA removes little —
compare scenarios, not modes alone, and always against a CHA run
(`-graph-compare`).

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
cache miss is a clear error, not a silent skip. Why does a cached OSV answer
expire (CVE:F2)? Because an *empty* answer ("no advisories for this
version") is the one that goes wrong silently: a week-old empty answer
hides every advisory published since. OSV responses are fetched again
after `-data-ttl` (24h); content pinned to a full commit SHA never
changes and never expires; LLM answers are paid results and are never
re-asked because of a refresh. On the reference apps the first online
refresh added 5 (petclinic) and 4 (WebGoat) advisories that the old
cache did not have. And when a result changes, why? `-snapshot-out`
records exactly the data a run used, `-snapshot-in` replays it without
cache or network, and `-snapshot-compare` names, per advisory, what moved:
the advisory text, the mapping or its fix commits, an input jar, an
algorithm version. A snapshot reproduces the data, not the analyzer: the
same analyzer version is required. Why are versions compared by Maven's
rules and not SemVer? Because Maven versions are not SemVer: `2.0.0.RC1`
is older than `2.0.0`, `1-sp` newer than `1`, `1.0.0.Final` equal to `1`.
`internal/mavenversion` follows Maven 3.9 and is checked against the
installed Maven's own comparator; a version it cannot place (a SNAPSHOT,
a non-numeric string) is "undetermined", never guessed textually.

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
fetch its **full content** as it existed at the fix commit (and, since CVE:F5a, at its parent — see below)
(`raw.githubusercontent.com`), and use a scanner (`internal/javadiff`) to
find which method encloses each changed line. The scanner first blanks
out comments/strings/text blocks (keeping line numbers intact) so a
stray `{` never desyncs brace matching, then tracks a stack of
class/method/"something else" frames (`if`/`for`/lambda/anonymous-class
bodies are transparent), and renames a method to `<init>` when its name
matches the enclosing class (a constructor).

**Why read the file before the fix, too (CVE:F5a)?** The analyzed
application contains the *vulnerable* release, i.e. the code before the
fix. Reading only the post-fix file had two failure modes: a method the
fix **deleted** never became a candidate (it is not in the new file), and
a helper the fix **added** (`checkInput`, a whole new firewall class)
became one although the vulnerable release does not contain it — so a
`NOT_REACHABLE` could rest on the wrong side of the diff. The `.diff`
format does not name the commit's parent, so ariadne asks the GitHub REST
API for it (metadata only — the patch still comes from the `.diff`, since
the API may truncate per-file patches), reads the file at the parent, and
maps changed + removed methods while keeping added ones as evidence.
Without the parent (no API answer, a merge commit with two parents) it
does not guess a base: it falls back to the post-fix side and marks the
row `MAPPING_NEW_SIDE_ONLY`. On the CVE:F0 inputs this removed 232/402
helper-only methods from petclinic/WebGoat mappings and added 120/116
pre-fix ones; it turned logback's CVE-2024-12798 (the fix deletes the
Janino evaluator) from `UNKNOWN` into a real `FRAMEWORK` path, and
CVE-2024-38821 on WebGoat from a `REACHABLE` that ended in an *added*
method into `FRAMEWORK` through the actually changed
`WebFilterChainProxy#filter`.

**Why check diff candidates against the bytecode (CVE:F5b)?** A source
diff knows method names, not JVM descriptors, and the scanner used to put
every method of a member class under the top-level class — `AntPathMatcher#matchStrings`
for code that really lives in `AntPathMatcher$AntPathStringMatcher`. Such a
selector either matched nothing (and the advisory looked `NOT_REACHABLE`
on the remaining methods) or matched an unrelated method of the outer
class. Now member classes keep their `$` names, and each candidate is
looked up in the analyzed classes: exactly one overload compatible with
the source parameter types gives an exact selector; anything less certain
stays a name selector. On the CVE:F0 inputs this turned CVE-2023-20873
from `NOT_REACHABLE` into `FRAMEWORK` (its fix is in a nested
`WebFilterChainPostProcessor`) and CVE-2026-50193 from `NOT_REACHABLE`
into `UNKNOWN` + `METHOD_NOT_FOUND` (its fixed class does not exist in the
analyzed jackson version) — in both, the old verdict rested on the wrong
class.

**Why ask git which release line a fix commit is on (CVE:F5c)?** An
advisory often links one fix commit per maintained branch. The bytecode
fit rule only drops a commit whose methods don't exist in your classes;
a commit on another branch whose methods *do* exist (most of Tomcat's
fixes touch the same classes on 9.0.x, 10.1.x and main) still added its
methods. Comparing your version's tag with each commit separates them:
on Tomcat 9.0.45, CVE-2023-46589 went from 13 mapped methods (mostly from
a main-line refactoring of request parameter handling) to the one method
the 9.0.x fix changed. Two git facts needed care: Tomcat and Spring tag
a release commit that is not on the branch, so every later fix looks
"diverged" by exactly that commit (treated as descending); and in a
repository merged forward (Jackson, Spring) the next major line
descends from every older tag, so "descends" means "fixed after your
release in history containing it", not "your line" — which is why it
never overrides the fit rule. No verdict changed on the CVE:F0 inputs;
the paths now rest on the right line's methods.

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
step 5, the mapping from step 6. Output: one of five statuses per CVE,
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
path). Visited only in search 2 → `FRAMEWORK`. Neither → `NOT_REACHABLE`
— unless no mapped method exists in the graph at all, which is `UNKNOWN`
with `METHOD_NOT_FOUND` (CVE:F1b; "method not found" is not "no path").
The advisory is evaluated per affected component instance and aggregated;
the instance whose own copy of a class lost to another jar's copy gets
`CLASSPATH_CONFLICT` instead of borrowing that copy's path. Why are model
gaps only an annotation (`MODEL_INCOMPLETE`) and not `UNKNOWN`? Because,
measured, hundreds of reachable unresolved reflection sites in Spring have
no bounded target set, so the rule applied to every `NOT_REACHABLE` row
of both reference apps — a label on everything distinguishes nothing.
`-json-format v2` writes those per-instance findings themselves (one per
advisory × component instance), with identity, version applicability,
mapping precision and reachability as separate fields instead of one
status — the v1 row is their aggregate, and each run checks that it is.

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
(reflection, dynamic proxies, serialization gadget chains, CHA
over-approximation — the same content as in the README), and since
DESERIALIZATION Z1 a "Deserialization entry points" section.

**Why does a `readObject` call connect to every serializable class's
callbacks, not just the type it is cast to?** Because the cast runs
*after* deserialization: by the time `(Foo) in.readObject()` checks the
type, every object in the stream has been created and its `readObject`,
`readResolve` or `readExternal` has already run. The stream — the data —
decides which classes appear, not the code that reads it, which is
exactly what makes deserialization dangerous. So the honest static answer
is "any concrete serializable class the input contains may run its
callbacks here", unless a filter restricts it. Since DESERIALIZATION Z2 a
filter (or XStream permission) written as a literal in the same method is
evaluated, and a class it rejects loses its edges; one ariadne cannot
read is reported as `unknown` and changes nothing — an unreadable policy
is not evidence that a class is blocked. For XStream the release matters
as much as the code: the same `new XStream().fromXML(xml)` admits every
type on 1.4.5, still every type on 1.4.10 (calling `allowTypes` alone
does not remove the default "any type" permission — observed on the real
JVM), and none of the application's classes on 1.4.20. On WebGoat this made two advisories reachable whose
vulnerable code is itself a serialization callback (jackson-databind's
`NodeSerialization#readExternal`, logback's `LoggingEventVO#readObject`) —
code no ordinary call ever reaches.

**Why does ariadne say "candidate chain" and not "exploitable"?** Because
DESERIALIZATION Z3 checks only what the bytecode can settle: the chain's
classes are there, they belong to the component the chain is about (a
class with the right name inside another, unidentified jar does not
count), the fixed release's guard is absent (read from the class itself —
commons-collections 3.2.2's `InvokerTransformer#readObject` refuses
unless a system property is set; a shaded copy carries the guard or not
regardless of its version string), the target JDK allows it, and the
entry's own policy does not reject any of its classes. With `-taint` it
also says whether request data reaches the bytes read at each entry: the
chain carries that entry's taint result as its *flow* (`flow_found`,
`no_flow_in_model`, `undetermined`), so "candidate chain with request data
flow_found" is the strongest statement it makes — still not an exploit:
the object graph, the field values and the runtime configuration are not
checked, and only request data is (a file or a message queue is not a
source). It cannot know whether a process-wide `jdk.serialFilter` is set, or
whether some chain outside its five-entry catalog exists. So a blocked
chain clears only that chain — never the advisory — and an entry with no
catalog chain is not "safe". The demo `scripts/prepare-deserialization-chains-demo.sh`
shows each guard on the real JVM by round-tripping just the guarded class;
no chain is ever built.

**What does "Data read depends on: parameter 0" mean — is that taint?**
No. It says the bytes or XML a deserialization entry reads are
computed from the method's parameter 0 (directly, or through decoding,
wrapping and helper calls), found by dependency summaries that follow
each call with that call's own arguments. It does not say who calls the
method with what — whether parameter 0 is an HTTP request parameter is
the taint pilot's job (`-taint`, next question). "result of X#m" means the data passes
through a call ariadne does not model, and it conservatively assumes the
result depends on everything that call was given; "—" means the data
depends on nothing the summary sees (a constant).

**What does `-taint` add, and how far can I trust `flow_found`?**
With `-taint` (DATA_FLOW D4, off by default) each deserialization entry
also gets a taint result: does data from a cataloged request source — a
Spring MVC handler parameter such as `@RequestParam`, a JAX-RS
`@QueryParam`, a `ServletRequest#getParameter` result — reach the data the
entry reads? ariadne starts from the entry's "data read depends on"
inputs and walks to every call site of that method in the application,
following the argument passed *there*, until it meets a source (`flow_found`,
with the path), only constants or inputs outside the catalog
(`no_flow_in_model`), or something it cannot follow — a field, a call it
does not model, a lambda caller, a depth limit (`undetermined`, with the
gaps listed). The guard column is specific to deserialization: only the
entry's own type policy (an evaluated filter rejecting every candidate
class, XStream's default allowlist) is `effective`; `String#replace`,
HTML escaping or an `isEmpty()` check on the way are listed as
`unrelated` — they change or test the text, they do not restrict what
classes the stream creates, and a check deciding one branch never clears
a flow. `flow_found` means the model sees request data reaching the
entry; it is not a demonstrated exploit (a working gadget chain and the
right classes are still needed — see the chain results). `no_flow_in_model`
is not proof of safety either: it is relative to the source catalog and
to callers in the application's own classes. The demo
`scripts/prepare-taint-demo.sh` has one entry per case and runs on the
real JVM. How it was measured — a real-JVM oracle that sends a
serialized canary through every handler of the test fixture and records
which entries read it — is in `docs/TAINT_QUALIFICATION.MD`: no observed
flow was reported as "no flow", every `flow_found` was observed, and the
four observed flows the model cannot follow (a depth limit, a field, a
lambda caller, an exception's message) come out `undetermined`.

**How do I know the deserialization model is right?** Partly by running
it against the real JVM: `testdata/deserialize/oracle` writes one object
of each scenario (private and public `readObject`, Externalizable, a
record, a class serializable only through an unlisted JDK base, the same
library class plain and relocated) and reads it back through a plain and
a filtered stream on a real JDK; a test compares what ran with what
ariadne connects. Today: no edge the JVM did not run, and exactly two
callbacks the JVM ran that the model does not connect — the record's
canonical constructor and the unknown-serializability class — both
declared gaps, counted in the report. And partly by sources:
`testdata/deserialize/cves.yaml` writes down, from each advisory's own
references, what should happen for seven cases, and
`testdata/deserialize/QUALIFICATION.MD` compares that with ariadne's
output. Neither proves the model complete; they pin down where it agrees
and where it knowingly does not.

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
