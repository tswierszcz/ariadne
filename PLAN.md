# ariadne — CVE reachability analysis for Java (JVM) apps

> Starter brief for a new coding session. Copy this file into an empty repo
> (as `PLAN.md`, `CLAUDE.md` or `AGENTS.md`, depending on the harness) and
> tell the agent: "Read PLAN.md and implement it phase by phase."

## 1. Task (from the manager)

1. Take an open-source Java application that has open-source dependencies.
2. Build the application's call graph and find the calls it makes into its dependencies.
3. Check those dependency methods against known vulnerabilities (CVE / GHSA).
4. Report **reachability**: which vulnerabilities in the dependencies can actually be
   called from this application.

**Final output:** a list of CVEs in the dependencies, each checked against the call graph
and given a status, plus the call path for the reachable ones.

## 2. Decisions already made

- **Scope:** any application compiled to JVM bytecode (Java, and also Kotlin/Scala, because
  we analyse bytecode) with Maven-ecosystem dependencies. Framework-specific knowledge lives
  in **profiles** (section 4). The first ones to implement are `spring` (Spring Boot) and
  `generic` (plain Java: CLI apps, libraries, anything without a dedicated profile).
- **Main target app:** `spring-projects/spring-petclinic` at an **old commit**
  (Spring Boot 2.x, ~2020–2021), so it has known-vulnerable dependencies.
  Realistic case: many CVEs, few reachable from app code.
- **Control app:** `WebGoat/WebGoat` (a v8.x tag). It is intentionally vulnerable: it uses an old
  XStream and calls its deserialization, so we expect at least one clear REACHABLE result.
  It proves the tool does not always say "not reachable".
- **CVE source:** OSV (`api.osv.dev`), queried with the exact resolved Maven coordinates (GAV).
- **Call graph: built by ariadne itself, in Go, from bytecode.** We already parse `.class`
  files for the hierarchy, so we also walk the `Code` attribute and collect `invoke*`
  instructions. As a result:
  - no JDK is needed to run ariadne;
  - we support every bytecode version ourselves;
  - lambdas and method references are resolved via `BootstrapMethods`/`LambdaMetafactory`.

  `gousiosg/java-callgraph` is kept **only as a reference** for an optional cross-check test.
- **Language of the tool:** **Go 1.22+**. A single static binary with no cgo.
  Use the standard library where possible (`net/http`, `encoding/json`, `encoding/binary`,
  `archive/zip`, `text/template`, `flag`, `log/slog`).
  Allowed external deps: `gopkg.in/yaml.v3`. Optional: `github.com/spf13/cobra`
  (only if the CLI grows beyond one command). No tree-sitter and no other cgo dependencies.
- **Known-vulnerability approach only.** No hunting for new 0-days. An LLM (GLM) is used
  only to help map CVE fix diffs to vulnerable methods.

## 3. Inputs and analysis model (generic)

1. **Input artifact.** ariadne accepts one of:
   - a directory of classes or a plain jar for the app, plus a directory of dependency jars;
   - a Spring Boot fat jar: classes are read from `BOOT-INF/classes/` and dependencies from
     `BOOT-INF/lib/*.jar`, straight from the zip with no extraction to disk;
   - (Phase 7) a WAR (`WEB-INF/classes`, `WEB-INF/lib`) or an EAR (nested WARs/jars).

   The `internal/input` package turns every form into one model:
   `App{Classes []ClassSource; Deps []DepJar}`.
2. **What is app code.** By default, the app classes are the ones from the app part of
   the input (`target/classes`, `BOOT-INF/classes`, ...). `--app-package com.example`
   (repeatable) overrides this, which is useful for a plain fat/uber jar.
3. **Dependency identification (GAV)**, in priority order:
   1. `META-INF/maven/<g>/<a>/pom.properties` inside the jar. This works for any build
      tool, as long as the jar came from Maven Central;
   2. `--deps-list deps.txt` (output of `mvn dependency:list`), matched by file name
      `artifactId-version[-classifier].jar`;
   3. nothing found → the jar is "unidentified" and listed in the report (no CVEs for it).

   If a jar has several `pom.properties` (shaded/uber jar), record all of them and warn.
   Relocated packages are not supported; list this under Limitations.
4. **Entry points (roots).** By default, roots are **all methods of all app classes**. This
   is generic and handles DI/reflection frameworks without knowing them.
   `--roots profile` takes only the entry points defined by the active profile
   (annotations, overridden framework methods, `main`).
5. **CHA expansion.** A call to `Interface.method` / a virtual `Class.method` gets extra
   edges to every implementation/override of that method in subtypes, using the class
   hierarchy.
6. **CHA explosion guard (important).** With "all app methods" as roots, a single call into
   the framework bootstrap (e.g. `main` → `SpringApplication.run`) leads into the whole
   framework. Naive CHA on calls like `Object#toString`, `Runnable#run`, `Map#get` or
   `Iterator#next` then connects almost the whole classpath, and nearly every CVE becomes
   `REACHABLE`. Mitigations:
   - **Do not CHA-expand** calls whose declared receiver is a `java.*`/`javax.*`/`jdk.*`/`sun.*`
     type, or a type on a configurable "too generic" list (`cha_skip.yaml`).
   - Paths that pass through a framework package from the active profile are reported as
     `FRAMEWORK`, not `REACHABLE`, unless the app → dependency entry call itself goes
     directly into the vulnerable package (see Phase 5).
   - Cap the BFS depth inside dependency code (`--max-dep-depth`, default e.g. 15).
7. **Framework-reachable vulns:** some CVEs live in code driven by the framework, not by
   the app (HTTP parsing in Tomcat/Jetty, Jackson deserialization of request bodies, the
   web framework itself). They will not show up on paths from app code. Give them a
   separate status instead of "not reachable", otherwise the report is falsely reassuring.
8. **Known blind spots** (must be listed in the report): reflection, dynamic proxies
   (AOP/CGLIB, Spring Data, JDK `Proxy`), `invokedynamic` other than lambdas/method refs,
   serialization gadgets, config-driven behaviour (XML wiring, `web.xml`), ServiceLoader,
   shaded/relocated dependencies. CHA over-approximates (false positives); reflection
   under-approximates (false negatives).

## 4. Profiles

A profile is a YAML file (defaults embedded with `//go:embed` in `internal/profile/profiles/`,
overridable with `--profile-file`). `--profile auto` (default) picks a profile by detection
markers. If several profiles match, they are merged (e.g. spring + servlet). Nothing
matches → `generic`.

```yaml
name: spring
detect:                       # any of these classes present in app or deps
  - org.springframework.boot.SpringApplication
  - org.springframework.context.ApplicationContext
roots:
  annotations:                # method-level, or class-level = all public methods
    - org.springframework.web.bind.annotation.RequestMapping
    - org.springframework.web.bind.annotation.GetMapping
    - org.springframework.web.bind.annotation.PostMapping
    - org.springframework.web.bind.annotation.PutMapping
    - org.springframework.web.bind.annotation.DeleteMapping
    - org.springframework.web.bind.annotation.PatchMapping
    - org.springframework.scheduling.annotation.Scheduled
    - org.springframework.context.event.EventListener
    - org.springframework.context.annotation.Bean
  overrides: []               # e.g. javax.servlet.http.HttpServlet#doGet in a servlet profile
  main: true
framework_packages:           # groupId:artifactId globs
  - org.apache.tomcat.embed:*
  - org.springframework:spring-web
  - org.springframework:spring-webmvc
  - org.springframework.boot:*
  - org.hibernate:*
  - ch.qos.logback:*
  - org.apache.logging.log4j:*
  - pkg: com.fasterxml.jackson.core:jackson-databind
    when_annotation: [org.springframework.web.bind.annotation.RequestBody,
                      org.springframework.web.bind.annotation.ResponseBody,
                      org.springframework.web.bind.annotation.RestController]
synthetic_edges:              # proxies with no implementation in bytecode
  - from_subtypes_of: org.springframework.data.repository.Repository
    to: org.springframework.data.jpa.repository.support.SimpleJpaRepository
```

- `generic`: roots = `main` + (for libraries, `--library`) all public methods of public
  classes. No framework packages and no synthetic edges.
- Spring Data note: in petclinic, `OwnerRepository` etc. are interfaces with
  runtime-generated proxies. Without the synthetic edge, CHA finds no implementation.
  This is expected.
- Later profiles (Phase 7): `jakarta` (JAX-RS `@Path`/`@GET`/..., CDI `@Observes`, EJB `@Schedule`),
  `servlet` (`HttpServlet#do*`, `Filter#doFilter`), `micronaut`, `quarkus`.

## 5. Statuses

| Status | Meaning |
|---|---|
| `REACHABLE` | A path exists from an app method to a vulnerable method and does not depend on the framework bootstrap (see 3.6). Report the shortest path. |
| `FRAMEWORK` | The vulnerable code is in a framework package of the active profile and is driven by the framework, or the only paths found go through the framework. Needs manual review. |
| `NOT_REACHABLE` | Vulnerable methods are known, no path found. |
| `UNKNOWN` | No mapping from the CVE to methods, so only package-level information. |

## 6. Phases

### Phase 0 — Environment (~30 min)
- Go 1.22+, plus JDK 11 (old petclinic) and JDK 17 (WebGoat) **only for building the target
  apps and the test fixtures**. ariadne itself does not need Java. Also Maven and git.
- Check what is already installed (`go version`, `java -version`, `mvn -v`) and check network
  access to `github.com` and `api.osv.dev`.
- `go mod init ariadne` (rename the module once the remote is known), with a `Makefile`
  containing `build`, `test`, `lint` (`go vet`, plus `staticcheck` if available) and `fixtures`.
- Test fixture `testdata/fixture/`: a tiny hand-written Java project (an interface + 2
  implementations, a lambda, a method reference, a static call, a constructor, a class
  annotated `@RestController`-like with an annotation from the fixture itself). Compile it
  with `javac --release 8` **and** `--release 21`, and commit the compiled `.class`/`.jar`
  files so that `go test` does not need a JDK.
- Optional: build `gousiosg/java-callgraph` as a reference for the cross-check test in
  Phase 3.

**Done when:** `go build ./...` passes on an empty skeleton and the fixture jars exist in `testdata/`.

### Phase 1 — Target apps + input normalization (`internal/input`, ~1 h)
```bash
git clone https://github.com/spring-projects/spring-petclinic && cd spring-petclinic
git checkout $(git rev-list -1 --before=2021-06-01 main)
./mvnw -q package -DskipTests
./mvnw -q dependency:copy-dependencies -DincludeScope=runtime
./mvnw -q dependency:list -DincludeScope=runtime -DoutputFile=deps.txt
```
Same for WebGoat (multi-module: pass the `target/classes` directories of the relevant
modules, including the vulnerable-components lesson module, as several `--app` inputs).
Check which JDK the chosen 8.x tag needs (the 8.2 line may need a JDK newer than 11).
If the old commit does not build, try a neighbouring commit or another JDK.

Keep these steps in `scripts/prepare-petclinic.sh` and `scripts/prepare-webgoat.sh` so they
can be repeated. Work in a gitignored `work/` directory.

`internal/input`:
- Input sources: a directory, a jar, a Spring Boot fat jar (detected by `BOOT-INF/`).
  Nested jars are read in memory (`zip.NewReader` over `bytes.Reader`).
- GAV resolution as in 3.3 (`pom.properties` → `deps.txt` → unidentified).
- `deps.txt` parser: the format is `group:artifact:type[:classifier]:version:scope`. Newer
  versions of the plugin append ` -- module x` or `(optional)`, so cut off everything after
  the coordinates. Skip the header and empty lines.

**Done when:**
- petclinic gives the same GAV list from `target/classes` + `target/dependency` as from
  the fat jar `target/*.jar`;
- for petclinic, GAVs from `pom.properties` agree with `deps.txt` (any differences are
  logged);
- WebGoat inputs are ready.

**Status: done.** `scripts/prepare-petclinic.sh` and `scripts/prepare-webgoat.sh` reproduce
this; artifacts live under the gitignored `work/`. Findings:
- Petclinic pinned commit `e7c879e` (2021-05-14), Spring Boot 2.4.5, Java 8 source level.
  Built clean with JDK 11 (`mvnw package -DskipTests`). 87 runtime dependency jars.
  `deps.txt` confirmed both quirks the plan anticipated: ANSI color codes in the plugin
  output and ` -- module x [auto]` suffixes — the parser must strip both.
- WebGoat tag `v8.2.2`, Java 15 source level, built with JDK 17. Two problems hit and
  worked around (now baked into `prepare-webgoat.sh`):
  1. `webgoat-lessons/cross-site-scripting` runs `asciidoctor-maven-plugin` at build time
     to render its lesson plan to HTML; its embedded JRuby fails under JDK 17's module
     system (`InaccessibleObjectException` even with `--add-opens`). We don't need that
     lesson, so we build only `webgoat-container` and
     `webgoat-lessons/vulnerable-components` via `-pl`, which skips the plugin instead of
     fighting it.
  2. `vulnerable-components` depends on `webgoat-container` as `provided`, which is only
     resolvable from the local repo, so `webgoat-container` needs `mvn install` (not just
     `package`) before `dependency:copy-dependencies`/`dependency:list` work on the lesson.
  - `app-classes.jar` merges both modules' `target/classes`. 126 runtime dependency jars.
    `deps.txt` confirms `com.thoughtworks.xstream:xstream:jar:1.4.5` — old XStream, as
    section 2 expects for the control case.

`internal/input` (the Go code, written alongside Phase 2 since Phase 2 needs it as input):
implemented as planned — `pom.properties` read from inside each jar first, `deps.txt`
parsing as a fallback for jars with none, unidentified jars reported rather than erroring.
One real-world wrinkle not in the original plan: a jar can contain *several*
`pom.properties` files (`ehcache-3.9.2.jar` has 10 — it bundles its internal modules'
poms), and the first one found is not necessarily the published artifact's own
coordinates. Fixed by preferring the entry whose `artifact-version` matches the jar's own
file name, with a warning when there's more than one candidate. Verified on both real
app dependency sets: petclinic's 87 jars and WebGoat's 126 jars are each 100% identified
with zero pom.properties/deps.txt disagreements.

### Phase 2 — CVE list (`internal/cves`, ~45 min)
- Input: the GAV list from `internal/input`.
- `POST https://api.osv.dev/v1/querybatch` with
  `{"queries": [{"package": {"ecosystem": "Maven", "name": "groupId:artifactId"}, "version": "x.y.z"}, ...]}`.
  Mind the batch limit (chunks of ≤1000) and pagination (`next_page_token`).
- For each returned id: `GET https://api.osv.dev/v1/vulns/{id}` for details.
  Fetch in parallel (worker pool, e.g. 8 goroutines).
- Keep: id, aliases (CVE ids), summary, severity (CVSS vector and score if present), affected
  package and version, references (especially type `FIX` and GitHub commit URLs).
- Deduplicate GHSA/CVE aliases into one entry (union-find over aliases; the key is the
  CVE id if present, otherwise the GHSA id).
- Cache: `.cache/osv/<sha256(request)>.json`. Use a shared `internal/httpcache` package,
  which Phase 4 uses too. Add an `--offline` mode that only reads from the cache and
  fails clearly on a cache miss.
- Cross-check option: `osv-scanner` (v2: `osv-scanner scan source -L pom.xml --format json`).

**Done when:** `cves.json` for petclinic is produced, with a non-trivial number of entries,
and `go test ./internal/cves` covers alias deduplication.

**Status: done.** Pagination support added (`/v1/query` with `page_token` as a fallback
for packages with more vulns than one querybatch page) but is untested against anything
real, since no Maven package comes close to that limit. Verified end-to-end against the
live OSV API: petclinic → **148 CVEs** across 87 resolved GAVs, including the plan's
SnakeYAML seed example `CVE-2022-1471`; WebGoat → 160 CVEs, including 34 for
`com.thoughtworks.xstream:xstream:1.4.5` alone (control case confirmed, `CVE-2021-39144`
among them). Re-running against a warm cache is fast (~0.2s, no network) and produces
byte-identical `cves.json` output.

### Phase 3 — Bytecode, call graph, hierarchy (`internal/classfile`, `internal/callgraph`, `internal/hierarchy`, ~2.5 h)
- `classfile`: a pure-Go `.class` parser (`encoding/binary`, big-endian) that reads:
  - the magic and the version (log the max version seen);
  - the constant pool: all tags up to the newest Java, including `MethodHandle`,
    `MethodType`, `Dynamic`, `InvokeDynamic`, `Module` and `Package`. `Long`/`Double` take
    2 slots;
  - access flags, this/super class and interfaces;
  - fields and methods (name, descriptor, flags);
  - the attributes `Code`, `RuntimeVisibleAnnotations` (class and method level) and
    `BootstrapMethods`.

  Skip `module-info.class` and `META-INF/versions/**` (multi-release jars: take the base
  version).
- **Bytecode walker** (in `classfile` or `callgraph`): walk the `Code` byte array opcode by
  opcode, using an operand-length table for all opcodes. Handle `tableswitch`/`lookupswitch`
  (padding to 4 bytes) and `wide`. Collect:
  - `invokevirtual` (182), `invokespecial` (183), `invokestatic` (184), `invokeinterface` (185)
    → an edge to `Owner#name(desc)` from `Methodref`/`InterfaceMethodref`, with kind
    `virtual|special|static|interface`;
  - `invokedynamic` (186) → look at the bootstrap method. If it is
    `java/lang/invoke/LambdaMetafactory#metafactory|altMetafactory`, take bootstrap
    argument #2 (`MethodHandle` → impl method) and add an edge of kind `dynamic` to it.
    Ignore `StringConcatFactory` and record others as unresolved (count them in the report);
  - optional: `new` (187) → class instantiation, to restrict CHA to instantiated types
    (RTA). Leave this as a later improvement and keep it behind a flag.
- `callgraph`:
  - **Interning:** method key `fully/qualified/Class#name(desc)`, kept in JVM internal form
    and rendered as `a.b.C#name(args)` in reports. The key maps to a `uint32` id
    (`map[string]uint32` + `[]string`). Adjacency as `[][]uint32` with the edge kind kept.
  - Name-only index `Class#method` → `[]uint32` (for overload-insensitive matching).
  - Expect millions of edges. Process jars in parallel and merge into one graph. Log
    memory, time and graph size (`slog`).
- `hierarchy`: superclass map, interfaces map, transitive subtypes map (computed lazily,
  with a memo), declared methods per class, source (app vs dependency jar → GAV) per class,
  and annotations per class/method.
- **Method resolution**: an edge to `C#m(d)` where `C` does not declare `m` → walk up the
  superclasses/superinterfaces to the declaring class, as the JVM does.
- **CHA expansion:** for each `virtual`/`interface` edge to `T#m(d)`, add edges to `S#m(d)`
  for every subtype `S` of `T` that declares that method. Respect the skip list from 3.6.
  Apply the synthetic edges from the active profile.
- Cross-check (optional test, skipped when the jar is missing): on the fixture, the edges
  we find ⊇ the edges from java-callgraph, ignoring JDK-internal ones.

**Done when:** unit tests on the fixture (both the `--release 8` and `--release 21` builds)
pass:
- the parser reads every class and its annotations;
- the walker finds static/virtual/interface/special calls, plus a lambda → impl method edge
  and a method ref → target edge;
- CHA: an interface call reaches both implementations, and `java.lang.Runnable#run` is not
  expanded.

On petclinic, the graph contains an edge `OwnerController → OwnerRepository#findById`
(or similar).

### Phase 4 — CVE → vulnerable methods (`internal/vulnmap`, ~2 h; the hardest part)
Sources, in priority order:
1. **Manual table** `vuln_methods.yaml` — always wins. Seed entries
   (verify each against the advisory / fix commit before relying on it):
   ```yaml
   CVE-2022-1471:   # SnakeYAML, unsafe Constructor
     - org.yaml.snakeyaml.Yaml#load
     - org.yaml.snakeyaml.Yaml#loadAs
     - org.yaml.snakeyaml.constructor.Constructor#<init>
   CVE-2021-44228:  # Log4Shell
     - org.apache.logging.log4j.core.lookup.JndiLookup#lookup
   CVE-2022-42889:  # Text4Shell
     - org.apache.commons.text.StringSubstitutor#replace
   # XStream CVEs (WebGoat control case) — all deserialization entry points
   # e.g. CVE-2021-39144 etc.:
   #   - com.thoughtworks.xstream.XStream#fromXML
   ```
   Embed the default file in the binary (`//go:embed`). Allow an override with `--vuln-methods`.
2. **Fix-commit diffs:** for references pointing to `github.com/<o>/<r>/commit/<sha>`,
   download `https://github.com/<o>/<r>/commit/<sha>.diff` (cached by `httpcache`).
   Only `.java` files under `src/main`. The Go diff parser (`internal/javadiff`) works in
   two steps:
   - parse the unified diff (files, hunks, `+`/`-`/context lines with line numbers);
   - extract method names with a **simple brace-aware scanner**: skip comments and string
     and char literals (text blocks too), track `{`/`}` depth, and recognise
     method/constructor headers with a regex on the text before `{`. Use the
     `@@ ... @@ <signature>` hunk context as a hint. For a reliable result, fetch the
     full file at the fix commit (`raw.githubusercontent.com/<o>/<r>/<sha>/<path>`) and
     map the changed lines to method ranges in that file.
   - Map file path → FQCN using the `package` declaration (and nested-class names `Outer$Inner`).
3. **LLM verification (GLM 5.3 via OpenRouter):** give the model the advisory summary + the
   diff, ask for a JSON list of methods that contain the vulnerability or are the public
   entry points to it (not methods touched only by refactoring or tests). Use a plain HTTP
   client against the OpenAI-compatible chat completions API (`POST <base_url>/chat/completions`).
   Configuration (all three are config variables; flags override env; nothing is hard-coded
   except the base URL default):

   | Env var | Flag | Default | Notes |
   |---|---|---|---|
   | `ARIADNE_LLM_BASE_URL` | `-llm-base-url` | `https://openrouter.ai/api/v1` | any OpenAI-compatible endpoint |
   | `ARIADNE_LLM_MODEL` | `-llm-model` | — (required with `-llm`) | the OpenRouter model id of GLM 5.3; check the exact slug on openrouter.ai/models |
   | `ARIADNE_LLM_API_KEY` | — (env only) | — (required with `-llm`) | never logged, never written to `.cache/` or the report |

   - The API key is read **only from the environment**, so it does not end up in shell
     history or `ps`. For local work, keep it in `.env`: gitignored, with a committed
     `.env.example` and loaded by the `Makefile`/shell, not by ariadne.
   - OpenRouter extras: send the `X-Title: ariadne` header. On HTTP 429/5xx, retry with
     backoff (max 3 attempts). Use `temperature: 0` and ask for JSON output
     (`response_format: {"type": "json_object"}` if the model supports it, otherwise parse
     the first JSON block from the answer).
   - `-llm` without the model or the key → fail at startup with a clear message, not in the
     middle of the run.
   - Store the prompt, the model id and the answer in `.cache/llm/<sha256(model+prompt)>.json`
     so results are reproducible. Validate the answer against a schema; on invalid JSON,
     retry once and then fall back to the `diff` source.
   - Flag this source in the report, together with the model id. Without `-llm`, this step
     is skipped.

Output: `vulnmap.json`: `{cve_id: {"methods": [...], "source": "manual|diff|llm"}}`.
CVEs without methods → status `UNKNOWN`.

**Done when:** a unit test on a stored sample `.diff` in `testdata/diffs/` returns the expected
methods, and the seed CVEs from the YAML show up in `vulnmap.json`.

### Phase 5 — Reachability engine (`internal/reach`, ~1 h)
- Roots: all app methods (default) or profile entry points (`--roots profile`).
- Multi-source BFS over the CHA-expanded graph with parent pointers (`[]int32`, -1 = none)
  and a `[]bool` visited set (or a bitset).
- Match vulnerable methods by `Class#method` (ignore overloads, conservative). Also treat
  a call to a public method of a vulnerable class's subclass as a match if the method is
  inherited (resolved via the hierarchy).
- `REACHABLE` vs `FRAMEWORK` split, done with two BFS passes:
  1. over a graph without edges *into* the framework packages of the active profile, except
     when the vulnerable method itself is in that package and the edge is an app → dependency
     entry call → `REACHABLE`;
  2. over the full graph → anything found only here is `FRAMEWORK`.

  Vulnerable methods in framework packages with no path at all are also `FRAMEWORK`.
- For each CVE, assign a status (section 5). For `REACHABLE`, reconstruct the shortest path
  and also record the **first app → dependency edge** on it (the "entry call").
- Cap the dependency depth (`--max-dep-depth`) and the node count. Log the graph size and
  the BFS time.

**Done when:** unit tests on a hand-built fixture graph cover path reconstruction, the depth
cap, and the `REACHABLE` vs `FRAMEWORK` split.

### Phase 6 — CLI + report (`cmd/ariadne`, `internal/report`, ~1 h)
```bash
go build -o bin/ariadne ./cmd/ariadne

# petclinic from build output
bin/ariadne -app target/classes -deps target/dependency [-deps-list deps.txt] \
  [-profile auto|spring|generic] [-roots all|profile] [-max-dep-depth 15] \
  -out report.md -json results.json [-html report.html] [-llm [-llm-model ID] [-llm-base-url URL]] [-offline]

# or straight from the Spring Boot fat jar
bin/ariadne -app target/spring-petclinic-2.4.5.jar -out report.md -json results.json
```
Report (`text/template` for Markdown, `html/template` for the optional HTML; templates embedded with `//go:embed`):
- Header: app, commit, date, ariadne version (`debug.ReadBuildInfo`), active profile(s),
  max bytecode version, graph size (nodes/edges, before/after CHA), number of unresolved
  `invokedynamic` sites, number of deps (identified / unidentified).
- Summary: total CVEs and counts per status, plus counts by severity.
- Table: CVE | package:version | severity | status | vulnerable method(s) | mapping source.
- Section per `REACHABLE` CVE: full call path, e.g.
  `org.springframework.samples.petclinic.owner.OwnerController#processFindForm → … → org.yaml.snakeyaml.Yaml#load`,
  with the entry call highlighted.
- Section `FRAMEWORK`: what to check manually.
- Section "Unidentified jars" (no GAV, so no CVE check).
- **Limitations** section (section 3, point 8).
- Control-case result from WebGoat (expected: XStream deserialization `REACHABLE`).

`results.json` must be deterministic: sort by CVE id, then by method. This makes snapshot
tests possible.

### Phase 7 — More inputs, profiles and build tools (~2–3 h)
- Input: WAR (`WEB-INF/classes`, `WEB-INF/lib`) and EAR.
- Profiles: `servlet`, `jakarta` (JAX-RS/CDI/EJB), and optionally `micronaut` and `quarkus`.
- Gradle: `scripts/gradle-copy-deps.init.gradle`, an init script that copies
  `runtimeClasspath` into a directory (`gradle -I ... ariadneCopyDeps`). With GAV coming
  from `pom.properties`, no list is needed.
- A non-Spring control app (e.g. an old Jakarta EE/JAX-RS sample, or a plain Java CLI app
  with a known-vulnerable dependency), to show that the `generic`/`jakarta` profiles work.
- Optional: RTA (restrict CHA to types instantiated with `new`) as `--cha rta`.

### Phase 8 (optional) — Exploitability check
For 1–2 `REACHABLE` CVEs: ask the LLM whether user-controlled input actually flows to the
vulnerable call (look at the path and the code), or run Strix against a **locally running**
instance with an instruction targeting that CVE. Never run it against third-party systems.

## 7. Project layout
```
ariadne/
  go.mod
  Makefile
  cmd/ariadne/main.go        # flags, orchestration
  internal/
    input/                   # dirs, jars, Boot fat jar, (WAR/EAR); GAV from pom.properties / deps.txt
    cves/                    # GAVs -> OSV -> cves.json
    httpcache/               # disk cache for OSV, diffs, raw files; --offline
    classfile/               # pure-Go .class parser + bytecode walker
    callgraph/               # interning, edges, adjacency
    hierarchy/               # super/interfaces/subtypes, method resolution, annotations
    profile/                 # profile loading, detection, merging
      profiles/              # spring.yaml, generic.yaml (later: servlet, jakarta, ...)
    vulnmap/                 # CVE -> methods (manual yaml, diffs, LLM)
      vuln_methods.yaml      # embedded default
    javadiff/                # unified-diff parser + brace-aware Java method scanner
    llm/                     # OpenAI-compatible client (OpenRouter + GLM 5.3 by default), with cache
    reach/                   # CHA expansion, BFS, statuses, paths
    report/                  # Markdown/HTML templates + JSON
      templates/
  config/
    cha_skip.yaml
  scripts/
    prepare-petclinic.sh
    prepare-webgoat.sh
  testdata/
    fixture/                 # Java sources + committed .class/.jar (release 8 and 21)
    diffs/                   # sample fix-commit diffs
work/                        # cloned target apps (gitignored)
.cache/                      # OSV, diffs, LLM answers (gitignored)
.env                         # ARIADNE_LLM_* incl. the API key (gitignored)
.env.example                 # template without secrets (committed)
```

Conventions: return errors (`fmt.Errorf("...: %w", err)`) and never panic in library code;
pass `context.Context` to everything that does network I/O; use table-driven tests.

## 8. Acceptance criteria
- [ ] `go build ./...`, `go vet ./...` and `go test ./...` pass, and `go test` needs no JDK
      and no network.
- [ ] ariadne runs without Java installed.
- [ ] Runs end-to-end on old petclinic (both from `target/classes`+deps and from the fat
      jar, with the same result) and produces `report.md` + `results.json`.
- [ ] Every CVE has exactly one status; `REACHABLE` ones have a path and an entry call.
- [ ] Petclinic is not "everything REACHABLE" (the CHA guard from 3.6 works).
- [ ] WebGoat control run shows at least one `REACHABLE` CVE (XStream).
- [ ] Unit tests for the `.class` parser and the bytecode walker on fixtures compiled for
      Java 8 and Java 21 (incl. a lambda edge).
- [ ] Unit test for CHA expansion (interface call reaches implementation; JDK types not expanded).
- [ ] Unit test for BFS path reconstruction on a fixture graph.
- [ ] Profile auto-detection picks `spring` for petclinic and `generic` for the fixture.
- [ ] Report contains the Limitations and "Unidentified jars" sections.
- [ ] Re-running with `-offline` and a warm cache works without network.

## 9. Risks and fallbacks
| Risk | Fallback |
|---|---|
| Bug in our bytecode walker (missed/garbled edges) | cross-check test against java-callgraph on the fixture and on petclinic app classes; fixtures for every invoke kind |
| New class-file features in future Java versions | the parser rejects unknown constant-pool tags with a clear error naming the class and the version; skip such a class, do not abort the run |
| Old petclinic does not build | neighbouring commit, other JDK, or a different old Spring Boot sample |
| CHA makes almost everything reachable | extend `cha_skip.yaml`, lower `--max-dep-depth`, rely on the FRAMEWORK split; later RTA |
| Jar without `pom.properties` (non-Central, custom builds) | `--deps-list`; otherwise "unidentified" in the report |
| Shaded/relocated dependencies | detect several `pom.properties`, warn; relocation under Limitations |
| OSV has no FIX commit for a CVE | manual entry in `vuln_methods.yaml` for important ones, else `UNKNOWN` |
| Brace-aware scanner misreads Java (generics, annotations, lambdas) | map changed lines to method ranges in the full file from the fix commit; LLM verification; manual table |
| Call graph too big / BFS slow | interning + `uint32` ids, parallel jar parsing; restrict to jars with CVEs plus their transitive callees; cap depth |
| LLM mapping is wrong | manual table overrides; mark `llm` source in the report |

## 10. Time budget (~10–11 h for Phases 0–6, plus ~2–3 h for Phase 7)
Phase 0: ~0.5 h · Phases 1–2: ~1.75 h · Phase 3: ~2.5 h · Phase 4: ~2 h ·
Phases 5–6: ~2 h · Phase 7: ~2–3 h · Phase 8: optional.
