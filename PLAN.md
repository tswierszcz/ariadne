# ariadne — CVE reachability analysis for Java (JVM) apps

> Starter brief for a new coding session. Copy this file into an empty repo
> (as `PLAN.md`, `CLAUDE.md` or `AGENTS.md`, depending on the harness) and
> tell the agent: "Read PLAN.md and implement it phase by phase."

## 1. Task

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
| `NOT_APPLICABLE` | Added after Phase 6. Every affected dependency is a WebJar (no `.class`, assets under `META-INF/resources/webjars/`), i.e. client-side JS/CSS outside a JVM call graph. Not "safe": check whether the app's pages use the component. Empty aggregator jars (`spring-boot-starter-*`, `org.jruby:jruby`) deliberately do not qualify: Spring4Shell (CVE-2022-22965) is filed against `spring-boot-starter-web`. Effect: petclinic's 6 `org.webjars:bootstrap` CVEs move from UNKNOWN to NOT_APPLICABLE; WebGoat unchanged. |

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

**Status: done.** All three packages implemented and verified against both the fixture
and the real petclinic/WebGoat jars (see the Phase 3a/3b/3c commits for the detailed
per-package findings). Headline numbers on real petclinic (87 dep jars):
- `classfile`: 29154 classes, 822218 call edges, 0 parse errors, 0 unresolved dynamic
  sites (WebGoat: 40811 classes, 1023924 edges, 13 unresolved — all StringConcatFactory).
- `hierarchy`: built in ~4s; confirmed `OwnerRepository` has 0 bytecode subtypes (the
  Spring Data proxy blind spot) and `java.lang.Object` has 26367 of 29123 classes as
  subtypes.
- `callgraph`: 273180 nodes; **1229600 edges with the CHA skip list vs. 11452847
  without it (9.3x more), and 1.9s to build vs. 2m43s (85x slower)** — measured
  confirmation that the skip list from section 3.6 is load-bearing, not cosmetic.
- The exact acceptance-test edge from this section,
  `OwnerController#initUpdateOwnerForm → OwnerRepository#findById`, is present (as a
  direct `interface` edge, since OwnerRepository has no CHA-reachable implementation —
  Spring generates its proxy at runtime).
- The CHA/synthetic-edges-from-profile item is deferred: no profile system exists yet
  (that's Phase 4/5/6 territory per section 4), so `callgraph.Build` currently takes a
  plain `SkipFunc` with `DefaultSkip` (java./javax./jdk./sun.) as the only rule; profile
  package wiring (including Spring Data's synthetic edges) happens when internal/reach
  and the CLI are built.

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

**Status: done — sources 1, 2 and 3 (LLM, opt-in) implemented; see the Source 3 notes below.**
- `vuln_methods.yaml` seeded with the plan's own examples (SnakeYAML, Log4Shell, Text4Shell)
  plus every XStream 1.4.5 CVE id OSV actually returned for WebGoat (37 entries via a
  YAML anchor/alias, since they share one attack surface: an unconfigured XStream's
  `fromXML`/`unmarshal`). `LoadManual("")` loads this embedded default;
  `LoadManual(path)` loads an override.
- Diff source implemented as `internal/javadiff` (unified-diff parsing + a brace-aware,
  comment/string-stripping Java method scanner that correctly treats `if`/`for`/lambda/
  anonymous-class bodies as transparent and renames a same-named method to `<init>`) plus
  `internal/vulnmap`'s orchestration: every GitHub `.../commit/<sha>` reference → fetch
  `<url>.diff` → keep `.java` files under `src/main/` → fetch each at
  `raw.githubusercontent.com/.../<sha>/<path>` → map changed lines to methods. All fetches
  go through `internal/httpcache`.
  **Deviation from the plan's own description:** rather than a regex/`javalang` pass over
  the diff's `@@` hunk context, the implementation always fetches the full post-fix file
  and runs the same brace-aware scanner on it, using the diff only for *which lines*
  changed — this was simpler to get right than trying to parse Java out of hunk context
  lines alone, and the plan already allows fetching the full file "for a reliable result".
- Verified against the real `cves.json` for both apps (unit tests use a fake HTTP server
  instead, so no test depends on the network): petclinic resolved 56/148 CVEs (1 manual,
  55 diff); WebGoat resolved 104/160 (37 manual, 67 diff). WebGoat's `CVE-2021-39144` (the
  plan's own seed) resolves to `com.thoughtworks.xstream.XStream#fromXML`/`#unmarshal` via
  the manual table, confirmed in `vulnmap.json`.
- Not yet wired into a CLI/`vulnmap.json` output path (no `cmd/ariadne` flags exist yet);
  that happens with Phase 6.
- **Source 3 (LLM), opt-in with `-llm`.** Design, exact prompt text and test plan:
  [docs/phase4-llm.md](docs/phase4-llm.md). Files: `internal/vulnmap/llm.go` (config + HTTP
  client: retries, 401/403 fatal), `llmprompt.go` + `llm_system_prompt.txt` +
  `llm_user_prompt.tmpl` (the prompt; pinned by golden tests in `testdata/llm/`),
  `llmanswer.go` (schema validation + bytecode grounding), `llmsource.go` (cache in
  `<cache-dir>/llm/`, one corrective retry). Runs only when sources 1 and 2 found nothing.
  Decisions made while implementing:
  - the model slug is `z-ai/glm-5.3` (`.env` originally held the display name "GLM 5.3",
    which is not an OpenRouter id);
  - **advisory-only LLM results never report NOT_REACHABLE**: `reach` downgrades them to
    UNKNOWN with a `note` (REACHABLE/FRAMEWORK stay). A named method that exists but is not
    the vulnerable one would otherwise read as a false "safe";
  - `cves.CVE` gained `Details` (OSV's long advisory text, previously dropped);
  - source 2's `src/main/` filter hides Tomcat's `java/org/apache/...` layout (35 of
    petclinic's 43 "has a fix commit but unmapped" CVEs). Source 3 reads those diffs
    instead; widening source 2's filter is a possible later follow-up.
  - `-llm`/`-llm-model`/`-llm-base-url` in `cmd/ariadne`; the key is read only from
    `ARIADNE_LLM_API_KEY`. Unit tests use a fake HTTP server (no key, no network); the one
    real-endpoint test is skipped unless `ARIADNE_LLM_E2E=1` (`make e2e-llm`).
  - **Live check (one real CVE, `z-ai/glm-5.3`, CVE-2025-61795 / Tomcat in petclinic, diff
    evidence).** The first attempt returned an empty reply: GLM 5.3 is a reasoning model and
    spent its whole 8192-token budget on hidden reasoning. Fixes: request
    `reasoning: {"effort": "low"}`, `max_tokens` 16384, and no corrective retry on
    `finish_reason: "length"` (it would just repeat and double the cost; the CVE is skipped
    with a warning and nothing is cached). After the fix: valid reply, 189 completion tokens
    (25 reasoning), ~$0.0013, and `Request#parseParts` / `#getParts` / `#getPart` all passed
    bytecode grounding. A full petclinic run with `-llm` (then ~92 unmapped CVEs) came later
    — see [docs/llm-test-report.md](docs/llm-test-report.md), including its 2026-10-14
    re-run against the current baseline (91/148 already mapped by sources 1/2 by then, down
    from 56/148 at the time of this live check, after the source-2 filter widening noted
    just below).

- **Follow-up done: widen source 2's file filter.** `isMainJavaSource`
  (`internal/vulnmap/diffsource.go`) used to keep only `.java` files whose path contained
  `src/main/`. Repositories that do not use the Maven layout, notably Apache Tomcat
  (`java/org/apache/...` for code, `test/...` for tests), lost every file this way: source 2
  fetched the diff but mapped nothing. In petclinic that was 35 of the 43 CVEs with a fix
  commit that stayed unmapped after sources 1 and 2.
  - *Change made:* the `src/main/` requirement was replaced by the same rule `keepForLLM`
    already used (a `.java` file that is not test code: no `test`/`tests` path segment, no
    `*Test.java`/`*Tests.java`), as one shared helper, `isNonTestJavaFile`
    (`internal/vulnmap/javasource.go`), used by both source 2 and the LLM source. `keepForLLM`
    is now a one-line wrapper around it.
  - *Test added:* `TestResolve_diffSourceMapsNonMavenLayout`
    (`internal/vulnmap/vulnmap_test.go`) — a Tomcat-style diff fixture (`java/...` code,
    `test/...` tests) asserts source 2 now maps the `java/` file and that the `test/` file's
    raw URL is never even fetched. The existing `src/main/`-based test
    (`TestResolve_fallsBackToFixCommitDiff`) still passes unchanged.
  - *Effect, measured end-to-end* (both apps re-run online against live OSV/GitHub,
    `bin/ariadne`, before vs. after this change):
    - **petclinic: 56/148 → 91/148 CVEs mapped** (diff: 55 → 90). 37 CVEs changed status, all
      away from `UNKNOWN` (86 → 51 UNKNOWN): 32 to `FRAMEWORK` (18 → 49) — expected, since
      nearly all of them are Tomcat CVEs, and Tomcat is a framework package — and 5 to
      `NOT_REACHABLE` (25 → 29, e.g. `CVE-2025-24813`, `CVE-2025-46701`). `REACHABLE` stayed
      exactly 13/148: the change surfaces more framework-level detail, it doesn't change what
      the app's own code can reach.
    - **WebGoat: 0 CVEs changed status** (104/160 mapped, unchanged) — expected, since
      WebGoat's unmapped CVEs weren't a Tomcat-filter artifact to begin with.
  - *Trade-off, as anticipated:* source 2 takes every method a fix touched (not just the one
    judged vulnerable), so these new Tomcat mappings are coarser than an LLM's would be — the
    same trade-off source 2 already has for every other library.

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

**Status: done, except `--max-dep-depth`/node cap (deferred — see below).** Implemented as a
single multi-source BFS pair (not per-CVE: both passes run once in `reach.New`, independent of
how many CVEs `Analyze` is then asked about):
- Pass 1 (REACHABLE): BFS from every app method, treating every framework-package node as a
  dead end — visited (so a direct one-hop app→framework call still counts, which is exactly
  the plan's "entry call lands directly in the vulnerable package" exception) but never
  expanded past. Pass 2 (FRAMEWORK): the same roots, no restriction at all. A CVE's matched
  methods found in pass 1 → `REACHABLE`; found only in pass 2 → `FRAMEWORK`; not found in
  either (but vulnmap has methods for it) → `NOT_REACHABLE`; no vulnmap entry at all →
  `UNKNOWN`.
- Inherited-method matching implemented and verified by hand against `javap`: a call through
  a subclass that does not override the vulnerable method keeps the subclass as the
  bytecode's `Methodref` owner (confirmed, not assumed), so `matchVulnerableMethod` also
  checks every non-overriding transitive subtype of the vulnerable class.
- A minimal, non-profile `FrameworkMatcher` (`"group:artifact"` glob patterns) stands in for
  the full `framework_packages.yaml`-driven profile system section 4 describes; seeded with
  exactly that section's initial list. Profile wiring (incl. the conditional
  jackson-databind/`@RequestBody` rule) is still CLI/Phase 6+ territory.
- `--max-dep-depth`/a node cap is not implemented; on real petclinic/WebGoat graphs (up to
  340k nodes, 2.6M edges) both BFS passes together finish in under 100ms, so there's no
  evidence yet that a cap is needed. Revisit if a larger target app changes that.
- New `testdata/reachfixture/` (3 separate jars: an app and two dependencies, one tagged as
  a framework package) built via `scripts/build-reachfixture.sh`, specifically to give each
  status a real, distinct call-graph shape to test against instead of a status asserted only
  by construction.
- Verified end-to-end on both real apps (current `cves.json`/manual-table-only `vulnmap.json`,
  no diff/LLM sources re-run): analyzing all CVEs takes well under 100ms for either app.
  - **WebGoat (control case): `CVE-2021-39144` and every other XStream CVE → `REACHABLE`**,
    with the exact 2-hop path
    `VulnerableComponentsLesson#completed → XStream#fromXML`. 51 REACHABLE, 32 FRAMEWORK,
    21 NOT_REACHABLE, 56 UNKNOWN (of 160).
  - **petclinic: 13 REACHABLE, 18 FRAMEWORK, 25 NOT_REACHABLE, 92 UNKNOWN (of 148)** — "many
    CVEs, few reachable", exactly as section 2 predicts. `CVE-2022-22965` (Spring4Shell) is
    correctly found REACHABLE via `PetController`'s data-binding path.
  - **Observed limitation worth flagging for the profile phase:** several petclinic
    REACHABLE paths route through deep Spring Beans/SpEL internals (`CachedIntrospectionResults`,
    `SpelExpression`, ...) because section 4's initial framework list only names
    `spring-web`/`spring-webmvc`/`spring-boot:*`, not `spring-beans`/`spring-core`/
    `spring-expression`/`spring-aop`. Those are exactly the plan's own stated defaults, so
    this isn't a bug — just a sign the seed list may be too narrow once a real profile
    system assigns it, since some of these paths are arguably framework/reflection-mediated
    rather than genuine app-code reachability.
  - **Fixed in Phase 7's profile system.** `internal/profile/profiles/spring.yaml` adds
    exactly those four packages to `framework_packages`. Measured effect (live OSV/GitHub,
    `bin/ariadne`, comparing `-profile spring` against the same run with those four packages
    left out — i.e. this isn't a change to the 13/18/25/92 numbers printed just above, which
    predate the Phase 4 follow-up that later moved 18 to 49): **petclinic REACHABLE 13 → 3,
    FRAMEWORK 49 → 59**. `CVE-2022-22965` (Spring4Shell) is one of the ten that moved: its
    shortest path now reaches
    `OwnerController#processFindForm → ... → AbstractNestablePropertyAccessor#getPropertyValue
    → BeanWrapperImpl#getCachedIntrospectionResults → CachedIntrospectionResults#forClass
    → CachedIntrospectionResults#<init>`, i.e. the vulnerable constructor is only reached by
    continuing *through* `spring-beans`, which the old (too-narrow) framework list didn't
    stop at — so this line's own former claim ("correctly found REACHABLE") no longer holds
    under the `spring` profile, and `FRAMEWORK` is the more honest status: the same
    JavaBean-introspection path still needs the manual review Spring4Shell actually requires,
    it just isn't claimed as reachable from app code alone. WebGoat (which also auto-detects
    as `spring`, since it embeds Spring Boot too): REACHABLE 51 → 40, FRAMEWORK 32 → 43, same
    ten-ish-CVE family; the XStream control case (`CVE-2021-39144`) is unaffected — its path
    never goes near Spring internals. See Phase 7 below for the rest of the profile system
    this fix shipped as part of.

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

**Status: done**, with some scope deferred (no `-profile`/`-roots`/`-max-dep-depth`/`-html`/
`-llm` flags — all CLI/profile-system territory noted as deferred in earlier phases' status
notes). What's implemented:
- `cmd/ariadne` wires every package built so far into one command:
  `-app` (directory, plain jar, **or Spring Boot fat jar** — `internal/input.PrepareAppInput`
  detects `BOOT-INF/` and repacks `BOOT-INF/classes/`/extracts `BOOT-INF/lib/*.jar` to temp
  files automatically) → `ResolveDepJars` → `hierarchy.Build` → `callgraph.Build` →
  `cves.Resolve` → `vulnmap.ResolveAll` → `reach.Analyze` → `report.Build` + `WriteMarkdown`
  + `WriteResultsJSON`, logging progress at each step.
- `internal/report`: a Markdown report from an embedded `text/template` (no `-html`/
  `html/template` yet) covering every item this section lists, including the Limitations
  section and an "Unidentified jars" section. Severity counts bucket by OSV's
  `database_specific.severity` qualitative label (HIGH/CRITICAL/...) rather than the raw
  CVSS vector string — the vector string is what OSV's `severity[].score` actually contains
  for `CVSS_V3`/`CVSS_V4` entries, and a table of ~30 nearly-unique vectors was worse than
  useless as a summary (found by actually looking at the rendered report, not assumed).

**Two real bugs found and fixed by actually running the finished CLI end-to-end** (not just
unit tests) that earlier phases' unit-test-only verification had not caught:
1. **A genuine concurrency bug in `internal/httpcache`**: `Cache.Set`'s temp file name was
   `key+".tmp"` — deterministic per key, not per call. Two different CVEs can legitimately
   share one GitHub fix-commit URL (one commit fixing several CVEs), so two of
   `vulnmap.ResolveAll`'s worker goroutines raced to cache the very same key; one's `rename`
   stole the temp file out from under the other's, failing with "no such file or directory"
   on real data (not on the unit tests' synthetic single-key scenarios). Fixed with
   `os.CreateTemp` for a guaranteed-unique temp name per call; added a regression test run
   under `-race`.
2. **`results.json`'s reported paths were not deterministic between two runs of the same
   analysis**, violating this section's own explicit requirement. Root cause: `hierarchy`'s
   subtype inversion, both of `callgraph.Build`'s passes, and `reach.Engine.appRoots` all
   ranged over a `map[string]*ClassInfo` directly; Go deliberately randomizes map iteration
   order per `range`, so whichever of several equally-short BFS paths got recorded first
   varied run to run whenever a CVE's vulnerable method had more than one shortest path (real
   and common on WebGoat's graph — spotted by diffing two real runs' `results.json`, not by
   reasoning about the code). Fixed by having `Hierarchy` compute a sorted FQCN order once
   and having all three call sites iterate that instead of the map; added a dedicated
   regression test (`internal/callgraph/determinism_test.go`) confirmed to both fail against
   the original code and pass against the fix.

**Verified end-to-end** (the actual `bin/ariadne` binary, not ad hoc scripts), all three
input modes:
- petclinic via `-app target/classes` (packed to a jar in memory) `-deps target/dependency
  -deps-list deps.txt`: 148 CVEs, 13/18/25/92 REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN,
  matching every earlier phase's ad hoc verification exactly.
- petclinic via `-app target/spring-petclinic-2.4.5.jar` alone (the Spring Boot fat jar,
  no `-deps`/`-deps-list` at all): resolves 42/75 dependency jars from `pom.properties`
  alone (fewer than the 87/87 the directory + `deps.txt` path reaches — expected, and now
  documented behavior rather than a surprise, since Boot repackaging also drops some
  jars `target/dependency` has, e.g. `provided`-scope ones).
- WebGoat (control case): `CVE-2021-39144` and every other XStream CVE REACHABLE via
  `VulnerableComponentsLesson#completed → XStream#fromXML` in the rendered report.
  **Three full runs with `-offline` and a warm cache produced byte-identical
  `results.json`** (after the determinism fix) and reports identical except for the
  generation timestamp — section 8's "re-running with a warm cache works without
  network" acceptance criterion, confirmed.

### Phase 7 — More inputs, profiles and build tools (~2–3 h)
- Input: WAR (`WEB-INF/classes`, `WEB-INF/lib`) and EAR.
- Profiles: `servlet`, `jakarta` (JAX-RS/CDI/EJB), and optionally `micronaut` and `quarkus`.
- Gradle: `scripts/gradle-copy-deps.init.gradle`, an init script that copies
  `runtimeClasspath` into a directory (`gradle -I ... ariadneCopyDeps`). With GAV coming
  from `pom.properties`, no list is needed.
- A non-Spring control app (e.g. an old Jakarta EE/JAX-RS sample, or a plain Java CLI app
  with a known-vulnerable dependency), to show that the `generic`/`jakarta` profiles work.
- Optional: RTA (restrict CHA to types instantiated with `new`) as `--cha rta`.

**Status: the profile system, Gradle input, and WAR/EAR input are done; every profile besides
`spring`/`generic` is not.** `internal/profile` implements section 4 as described:
- YAML profiles (`internal/profile/profiles/{spring,generic}.yaml`), embedded via
  `//go:embed`, overridable/extendable with `-profile-file` (`Catalog.AddFile`, which loads
  one profile and replaces any existing one of the same name — not a partial per-field
  merge, matching how `-vuln-methods` already overrides a whole table rather than patching
  one entry).
- `-profile auto` (the default): every profile whose `detect` list names a class present
  anywhere in the analyzed app or its dependencies is merged (`profile.Select`/`merge`);
  `generic` is the fallback when nothing matches and is never itself auto-detected into.
  Verified: petclinic and WebGoat (which also embeds Spring Boot) both auto-detect as
  `spring` at the time this was first written; after the `servlet` profile below was added,
  both auto-detect as `spring+servlet` instead (their embedded Tomcat jar bundles the
  `javax.servlet.*` classes, which `servlet`'s own `detect` list matches too — confirmed not
  to change either app's CVE status counts, since `spring.yaml`'s own
  `org.apache.tomcat.embed:*` pattern already covered that exact dependency). The fixture
  still auto-detects as `generic`, since its `fixture.Entry` annotation is a
  stand-in for a route annotation, not a real framework marker class.
- `-roots profile` (default remains `-roots all`, every application method, exactly as
  before): `Profile.AppRoots` computes the entry-point set section 4 point 4 describes —
  every method carrying one of `roots.annotations`, every public method of a class itself
  carrying one, (when `roots.main` is true, as in every shipped profile) every public
  static `main(String[])`, and (added along with the `servlet` profile below)
  `roots.overrides`: see that profile's own status note for the matching rule and how it was
  verified.
- The conditional jackson-databind `framework_packages` entry (`pkg`/`when_annotation`) is
  implemented, but evaluated once for the whole app (`Profile.ResolvePatterns`), not
  per call site: true if *any* app class anywhere carries one of
  `@RequestBody`/`@ResponseBody`/`@RestController`, not only when the specific caller on a
  specific path does. A precise per-call-site version would need `internal/reach`'s BFS sink
  test (`bfs.go`) to see the calling node as well as the callee, which is a substantially
  bigger change than this simplification buys back — an app using `@RequestBody` anywhere is,
  in practice, exactly the kind of app where Jackson deserializes attacker-controlled JSON
  somewhere.
- Synthetic edges (`internal/callgraph.ApplySyntheticEdges`, called from `cmd/ariadne` after
  `callgraph.Build`): every transitive subtype of `from_subtypes_of` gets, for each method it
  declares, an edge to the same-named method on `to`. Verified on real petclinic/WebGoat:
  spring.yaml's `Repository -> SimpleJpaRepository` entry adds 766/774 edges respectively
  (every Spring Data repository interface method, e.g. `OwnerRepository#findByLastName`,
  now has somewhere to go instead of being a bytecode dead end).
- **Effect on the Phase 5 "observed limitation": fixed.** See that section's own follow-up
  note for the measured before/after (petclinic REACHABLE 13 → 3, WebGoat 51 → 40,
  `CVE-2022-22965` specifically moving from REACHABLE to FRAMEWORK) — adding
  spring-beans/-core/-expression/-aop to `framework_packages` was the one-line YAML change
  the whole profile system existed to make possible.
- **Gradle input: done**, as `scripts/gradle-copy-deps.init.gradle` + `scripts/
  prepare-petclinic-gradle.sh`, with **no changes to any Go package**. Decision made while
  implementing: rather than write a second Go parser for Gradle's own dependency-report
  format, the init script's two tasks (`ariadneCopyDeps`, `ariadneDepsList`) deliberately
  reproduce what `mvn dependency:copy-dependencies`/`dependency:list` already produce —
  every `runtimeClasspath` jar copied into one directory, plus a `deps.txt` in the exact
  `group:artifact:type[:classifier]:version:scope` shape `internal/input.ParseDepsList`
  already parses — so `-deps`/`-deps-list` consume a Gradle build exactly as they do a
  Maven one.
  - **Verified end-to-end** against spring-petclinic at commit `4df621b` (2021-05-10, "Make
    build work with Gradle") — the oldest commit in the repo's history with a `build.gradle`
    at all, and still Spring Boot 2.4.5, the same version `prepare-petclinic.sh`'s Maven
    target uses (commit `e7c879e`, 4 days later) — chosen specifically so the run could be
    checked against that already-known baseline instead of against an unfamiliar app.
    `go build ./cmd/ariadne` unchanged; `bin/ariadne -app app-classes.jar -deps dependency
    -deps-list deps.txt` ran with no code changes and no errors. Result, reproduced
    identically from a clean `git clone` (and byte-identical to the cached run): 80/80
    dependency jars identified, 144 CVEs, **3/56/29/50/6**
    REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN/NOT_APPLICABLE under `-profile spring` (auto
    still selects `spring`). **REACHABLE matches the Maven baseline exactly (3/3)**; the
    other counts differ in step with the dependency-count gap (80 vs. the Maven build's 87,
    itself expected — see the pom.xml note just below), not because anything in `internal/
    reach` behaves differently for a Gradle-built input.
  - **A genuine build problem found and worked around, not upstream-fixed.** As committed,
    this commit's `build.gradle` pulls in `ro.isdc.wro4j.gradle:wro4j-gradle-plugin:
    1.8.0.Beta4` (for LESS→CSS compilation) with no repository configured to resolve it
    from — and the artifact no longer exists on Maven Central or the Gradle Plugin Portal
    either; it was a JCenter-only release, and JCenter shut down in 2021–2022, so there is
    no live source to restore it from even by fixing the missing repository declaration.
    `prepare-petclinic-gradle.sh` drops the `wro4j` plugin and its `webResources` block
    (CSS/LESS bundling — irrelevant to bytecode analysis) from its own clone before
    building, and remaps the dependencies it declared via the plugin's own `webjarsRuntime`
    configuration (the webjar jars themselves — relevant, since e.g.
    `org.webjars:bootstrap` is exactly what the `NOT_APPLICABLE` status above covers) to the
    stock `runtimeOnly` configuration, which is all `ariadneCopyDeps`/`ariadneDepsList`
    need them on.
  - **Also found:** this commit's `pom.xml` had already moved on to Spring Boot 2.5.6 by
    the time `build.gradle` was added (2.4.5) — the two build files were briefly out of
    sync during the Maven→Gradle migration — so the Gradle- and Maven-built dependency sets
    for "the same commit" are not expected to match jar-for-jar; the 80-vs-87 count above is
    that gap, not a bug in dependency resolution.
  - **Not covered by `go test ./...`** (deliberately, consistent with `prepare-petclinic.sh`/
    `prepare-webgoat.sh` having no unit tests of their own): the init script's output format
    is already exercised by `internal/input`'s existing `ParseDepsList`/`ResolveDepJars`
    tests, since it was designed to match that format byte-for-byte, so a dedicated
    regression test would just re-run those same tests against another sample of the same
    shape. What's actually new here — the Groovy init script and the Gradle build itself —
    can't be driven by `go test` without a JDK and Gradle on the test machine, which would
    break that package's "no JDK, no network" guarantee; this end-to-end run is the
    verification instead, the same evidentiary standard Phase 1's prepare scripts already
    use.
  - **Multi-module Gradle projects are untested.** The init script's two tasks are written
    per-subproject (`allprojects { afterEvaluate { ... } }`), so a multi-module build would
    need them invoked once per module (`./gradlew :sub:ariadneCopyDeps :sub:ariadneDepsList
    ...`), the same way `prepare-webgoat.sh` already runs Maven's
    `dependency:copy-dependencies` once per WebGoat module — but no multi-module Gradle
    project has actually been run through this yet.
- **WAR/EAR input: done**, entirely inside `internal/input/source.go`, with **no changes to
  any other Go package** — `PrepareAppInput` already turned a Spring Boot fat jar into one
  app-classes jar + one extracted lib directory, and a WAR (`WEB-INF/classes`,
  `WEB-INF/lib`) is the exact same shape with a different prefix, so the fat-jar extraction
  logic (`prepareFromBootJar`) was generalized into `prepareFromFatArchive(files,
  classesPrefix, libPrefix)`, parameterized over the prefix pair, and reused for both.
  - **Detection** (`detectArchiveKind`): an EAR is checked *before* a WAR, since an EAR's
    own top-level entries are nested `*.war`/`*.jar` modules rather than a `WEB-INF:`/
    `BOOT-INF:` prefix of their own — `META-INF/application.xml` if present, else (that
    file has been optional since Java EE 5) a nested `*.war` entry at any depth.
  - **EAR** (`prepareFromEar`): opens every nested `*.war` entry as its own `zip.Reader`
    (`readNestedZip` — reads it fully into memory first, since `archive/zip` can only open
    a `*zip.File` as a one-shot stream, not re-seek into a nested archive in place) and
    merges its `WEB-INF/classes`/`WEB-INF/lib` the same way a standalone WAR would; every
    top-level `*.jar` entry is extracted as a dependency jar. A class name two modules both
    declare keeps whichever module was processed first (`extractModule`'s `written` map) —
    the same "first one wins" rule `internal/hierarchy` already applies to a class
    duplicated across two dependency jars, now also documented as the rule for this case.
  - **Decision made while implementing:** an EAR's own top-level module jars are always
    treated as dependency jars, never as app code, even though a top-level jar could in
    principle be an EJB module carrying real application classes — there is no reliable
    signal in a zip listing alone to tell that apart from a shared library jar without
    parsing `META-INF/application.xml`'s `<module><ejb>` entries (itself optional), which
    was out of scope for this pass. Recorded under Known limitations, not treated as a bug.
  - **Verified end-to-end**, not just on synthetic test fixtures: repackaged the existing
    Maven-built petclinic (`target/classes` + `target/dependency/*.jar`, the same input
    Phase 1's baseline uses) as a real WAR and as a real EAR wrapping that WAR, and ran both
    through `bin/ariadne` unmodified. Both reproduced the Maven-built baseline exactly:
    87/87 dependency jars identified (with `-deps-list`; 42/87 from `pom.properties` alone,
    the same fat-jar-style gap the Spring Boot case already has and already documents), 148
    CVEs, **REACHABLE 3/3**. (The WAR-only run's own class hierarchy was additionally
    cross-checked against the directory-based baseline: 29123 classes, 273180 call-graph
    nodes, 766 synthetic edges — identical to the Maven baseline in README.md/FAQ.md, which
    is exactly what byte-for-byte repackaging should produce.)
  - **New unit tests** (`internal/input/source_test.go`, no JDK or network needed, synthetic
    in-memory zips via a new `zipBytes`/`writeZip`/`jarEntryNames` test-helper trio):
    `TestPrepareAppInput_war`; `TestPrepareAppInput_ear` (one nested WAR module plus one
    top-level shared-lib jar, both ending up correctly merged); and
    `TestPrepareAppInput_ear_detectedWithoutApplicationXml`, confirming detection falls back
    to the nested-`*.war`-entry signal when `application.xml` is absent.
  - **Not covered:** a multi-WAR-module EAR (the merge logic is written to support it —
    `extractModule`'s dedup — but has not actually been run against one), and an EJB module
    whose classes should count as app code (see the decision above).
- **`servlet` profile: done** (`internal/profile/profiles/servlet.yaml`), covering plain
  (non-Spring) Servlet API apps — `HttpServlet#do*`/`Filter#doFilter`/lifecycle listeners —
  across both the pre-Jakarta-EE9 `javax.servlet.*` namespace and the post-EE9
  `jakarta.servlet.*` one (same spec, renamed packages; one profile lists both rather than
  forcing an app onto one).
  - **`Roots.Overrides` wired up**, the one piece this section's own earlier status note
    left as "parses but matches nothing yet": a new `Profile.overrideRoots` (`internal/
    profile/profile.go`) matches each `"fqcn#method"` spec (e.g.
    `javax.servlet.http.HttpServlet#doGet`) against every app class that is that class
    itself or one of its transitive subtypes (`hierarchy.Subtypes` — this works even though
    `HttpServlet` itself is never read: subtype inversion only needs an app class's own
    recorded superclass name, not the superclass's own class file) **and declares that
    method itself** — an actual override, not an inherited no-op. A new
    `classfile.Method.IsAbstract()` (mirroring the existing `IsStatic`/`IsPublic`) excludes
    an interface's own abstract declaration from matching as a root — caught by a unit test
    using the existing fixture's `Greeter` interface (two app implementations,
    `FormalGreeter`/`CasualGreeter`, both become roots for `Roots.Overrides:
    ["fixture.Greeter#greet"]`; `fixture.Greeter#greet` itself, the abstract declaration,
    does not).
  - **No `roots.annotations`** for `servlet.yaml`: `@WebServlet`/`@WebFilter` are
    class-level-only markers (unlike `@RequestMapping`), and the container only ever
    dispatches to the specific lifecycle methods below, not "every public method" the way
    `spring.yaml`'s annotation handling assumes for `@Controller` — `roots.overrides` is the
    tool actually built for this, not a misuse of `roots.annotations`.
  - `framework_packages`: `javax.servlet:*`, `jakarta.servlet:*`, `org.apache.tomcat:*`,
    `org.apache.tomcat.embed:*`, `org.eclipse.jetty:*`. No `synthetic_edges`: a servlet
    container instantiates the app's own servlet/filter classes directly and calls their
    real, analyzable lifecycle methods — there is no Spring-Data-style interface-with-
    zero-bytecode-implementation case here to bridge.
  - **Verified against both existing baselines first** (petclinic, WebGoat — both Spring
    Boot apps on an embedded Tomcat, so both now auto-detect `servlet` too): both now report
    `-profile auto` as `spring+servlet`, and **both apps' CVE status counts are
    byte-identical to before** (petclinic 3/59/29/51/6, WebGoat 40/43/21/56/0) — the merge is
    additive with no regression, confirmed by running `bin/ariadne` against both, not
    assumed.
  - **Verified against a new, real, non-Spring control app** —
    `scripts/prepare-servlet-demo.sh` generates and builds a tiny one-class, one-dependency
    Maven project (`demo.YamlServlet`, whose `doGet` calls straight into
    `org.yaml.snakeyaml:1.30`'s `Yaml#load`, the manual-table seed CVE-2022-1471 this file's
    own section 4 example already names) — small enough to generate from a script rather
    than clone, filling the "non-Spring control app" gap this section's own earlier status
    note left open, at least for `servlet` (`jakarta`/`micronaut`/`quarkus` still have none).
    Run through `bin/ariadne` unmodified, twice:
    - **`-roots all`** (the default): profile auto-detects `servlet` (not `generic`) purely
      from `javax.servlet.http.HttpServlet` being present on the classpath;
      `demo.YamlServlet#doGet → org.yaml.snakeyaml.Yaml#load` is the reported `REACHABLE`
      path for CVE-2022-1471.
    - **`-roots profile`**: `roots: profile "servlet" (1 entry-point methods)` — exactly
      `demo.YamlServlet#doGet`, matched via `Roots.Overrides`'
      `javax.servlet.http.HttpServlet#doGet` entry, nothing else — and CVE-2022-1471 is
      still `REACHABLE` via the identical path, confirming the narrower, profile-driven root
      set finds the same vulnerability the unrestricted "every app method" root set does.
- **`jakarta` profile: done for JAX-RS and EJB; CDI is partial by design**
  (`internal/profile/profiles/jakarta.yaml`). Covers JAX-RS's HTTP-verb annotations
  (`@GET`/`@POST`/`@PUT`/`@DELETE`/`@PATCH`/`@HEAD`/`@OPTIONS`) and EJB's
  `@Schedule`/`@Schedules`, across both the `javax.*` and `jakarta.*` namespaces, as
  `roots.annotations` — no new matching code was needed, since `Profile.AppRoots` already
  implements that mechanism generically for `spring.yaml`; `jakarta.yaml` is simply its
  second real user.
  - **Deliberately excludes `@Path` from `roots.annotations`**: `@Path` is routinely
    class-level (marking a JAX-RS resource class), and the class-level rule ("every public
    method of a class carrying one of these becomes a root") would then wrongly turn every
    public helper method of that class into an entry point, not just the ones actually
    exposed as endpoints — which always also carry an HTTP-verb annotation, the actual
    entry-point signal. `@Path` is listed only in `detect`.
  - **CDI's `@Observes` is not implemented as a root**, and this is a real gap, not an
    oversight to paper over: `@Observes` annotates a method *parameter*
    (`void onEvent(@Observes SomeEvent e)`), not the method or the class, and ariadne's
    bytecode parser (`internal/classfile`) does not read the `RuntimeVisibleParameterAnnotations`
    attribute at all — only class- and method-level `RuntimeVisibleAnnotations`.
    Implementing it would need: parsing that attribute in `internal/classfile`, a new
    per-parameter-annotations field on `hierarchy.Method`, and a new `Profile` YAML field
    (e.g. `roots.parameter_annotations`) plus matching logic in `internal/profile` — a
    three-package change, not a YAML-only addition like the rest of this profile, so it was
    scoped out rather than rushed. CDI's own framework packages
    (`javax.enterprise:*`/`jakarta.enterprise:*`/`org.jboss.weld:*`) are still listed, so CDI
    container internals are at least classified `FRAMEWORK` on any path that reaches them,
    just not via an observer-method root.
  - `framework_packages`: `javax.ws.rs:*`/`jakarta.ws.rs:*` plus the three common JAX-RS
    implementations (`org.glassfish.jersey:*`, `org.jboss.resteasy:*`, `org.apache.cxf:*`,
    picked since no single one is "the" JAX-RS runtime the way embedded Tomcat is for
    `spring`/`servlet`), `javax.ejb:*`/`jakarta.ejb:*`, and
    `javax.enterprise:*`/`jakarta.enterprise:*`/`org.jboss.weld:*` for CDI. No
    `synthetic_edges`: no interface-with-zero-bytecode-implementation case like Spring Data's
    here to bridge.
  - **Verified against both existing baselines first**: petclinic and WebGoat still
    auto-detect `spring+servlet` only — `jakarta`'s own `detect` markers
    (`javax.ws.rs.Path`/`javax.ejb.Stateless`/`javax.enterprise.inject.spi.Extension`, plus
    the `jakarta.*` namespace equivalents) are absent from both, confirmed by running
    `bin/ariadne` against both (identical status counts to before this profile existed).
  - **Verified against a new, real, non-Spring, non-servlet control app** —
    `scripts/prepare-jakarta-demo.sh` generates and builds a tiny two-class Maven project:
    `demo.YamlResource` (a JAX-RS `@Path("/yaml")` resource, `@GET`-annotated `get` method)
    and `demo.ScheduledYamlJob` (an EJB `@Stateless` bean, `@Schedule`-annotated `run`
    method), both calling straight into `org.yaml.snakeyaml:1.30`'s `Yaml#load` — the same
    manual-table seed CVE-2022-1471 the `servlet` profile's own control app uses. Run through
    `bin/ariadne` unmodified, twice:
    - **`-roots all`**: profile auto-detects `jakarta` (not `generic`) purely from
      `javax.ws.rs.Path` being present on the classpath; CVE-2022-1471 is `REACHABLE`.
    - **`-roots profile`**: `roots: profile "jakarta" (2 entry-point methods)` — confirmed
      directly (not just by the log line) to be exactly `demo.YamlResource#get` and
      `demo.ScheduledYamlJob#run`, via a one-off debug test calling `Profile.AppRoots`
      straight — and CVE-2022-1471 is still `REACHABLE`. (The report's own recorded path for
      both of this demo's two CVEs happens to go through `ScheduledYamlJob#run` rather than
      `YamlResource#get`: both roots reach `Yaml#load` in exactly one hop, and
      `internal/reach`'s BFS deterministically breaks same-distance ties by sorted FQCN
      order, which puts `ScheduledYamlJob` first — not a sign `YamlResource#get` wasn't
      actually a root, just that the report only ever records one shortest path per CVE by
      design.)
- **`micronaut` profile: done** (`internal/profile/profiles/micronaut.yaml`). Covers HTTP
  routing (`@Get`/`@Post`/`@Put`/`@Delete`/`@Patch`/`@Head`/`@Options`) and
  `@Scheduled`/`@EventListener` as `roots.annotations` entries, the same mechanism
  `spring.yaml`/`jakarta.yaml` already use — again, no new matching code, just a third user
  of it. `@Controller` itself is confirmed (via `javap` on the published
  `micronaut-http` jar) `@Target(TYPE, ANNOTATION_TYPE)` — class-level only — so it is
  `detect`-only, not a root, for the same reason `@Path`/`@WebServlet` are: the class-level
  rule would otherwise turn every public helper method of a controller into an entry point,
  not just its actual routes. Also confirmed via `javap`: Micronaut's own annotations
  (`@Get`, `@Controller`) are real `@Retention(RUNTIME)` — not a compile-time-only marker
  ariadne's `RuntimeVisibleAnnotations`-only parser would silently miss, which was worth
  checking rather than assuming, since Micronaut's whole selling point is avoiding runtime
  reflection elsewhere in its own design.
  - **No `synthetic_edges`, by design, not by omission:** unlike Spring Data's
    interface-with-zero-bytecode-implementation repositories, a Micronaut Data repository
    interface's implementation is generated by Micronaut's own annotation processor *at
    compile time*, landing as a real class in the app's own compiled output — CHA already
    finds it like any other subtype, with nothing to bridge. (Not independently verified
    against a real Micronaut Data repository in this pass — the control app below doesn't
    use one — but it follows directly from how Micronaut's AOT compilation model works and
    is the reason no `from_subtypes_of`/`to` entry was added speculatively.)
  - **Verified against both existing baselines first**: petclinic and WebGoat still
    auto-detect `spring+servlet` only (`micronaut`'s own detect markers,
    `io.micronaut.context.ApplicationContext`/`io.micronaut.runtime.Micronaut`, are absent
    from both), confirmed by running `bin/ariadne` against both (identical status counts).
  - **Verified against a new, real, non-Spring, non-servlet, non-JAX-RS control app** —
    `scripts/prepare-micronaut-demo.sh` generates and builds a one-class Maven project
    (`demo.YamlController`, a `@Controller("/yaml")` with a `@Get`-annotated `get` method)
    calling straight into `org.yaml.snakeyaml:1.30`'s `Yaml#load` — the same manual-table
    seed CVE-2022-1471 the `servlet`/`jakarta` control apps use. Pinned to **Micronaut
    3.9.7**, not the current 4.x/5.x line: Micronaut 5.x's published jars (as of this
    writing) are compiled for Java 25, newer than the JDK 11 this project's other demo/prep
    scripts assume — a `javac` ceiling on the demo app, not a limitation of ariadne's own
    bytecode parser, which has no version ceiling. Run through `bin/ariadne` unmodified,
    twice:
    - **`-roots all`**: profile auto-detects `micronaut` (not `generic`) purely from
      `io.micronaut.context.ApplicationContext` being present on the classpath;
      CVE-2022-1471 is `REACHABLE` via `demo.YamlController#get → Yaml#load`.
    - **`-roots profile`**: `roots: profile "micronaut" (1 entry-point methods)` — exactly
      `demo.YamlController#get` — and CVE-2022-1471 is still `REACHABLE` via the identical
      path.
- **`quarkus` profile: done** (`internal/profile/profiles/quarkus.yaml`). Quarkus's own REST
  layer is plain JAX-RS (`javax.ws.rs`/`jakarta.ws.rs`), already covered by `jakarta.yaml` —
  a Quarkus REST app auto-detects as `jakarta+quarkus`, the same way a Spring Boot app on
  embedded Tomcat auto-detects as `spring+servlet` — so this profile only adds what's
  genuinely Quarkus-/MicroProfile-specific on top: `io.quarkus.scheduler.Scheduled` and
  MicroProfile Reactive Messaging's `@Incoming`/`@Outgoing`
  (`org.eclipse.microprofile.reactive.messaging`), as a fourth `roots.annotations` user —
  again no new matching code needed. Confirmed via `javap` on the published jars (the same
  diligence applied to `micronaut.yaml`): all three are `@Target(METHOD)`,
  `@Retention(RUNTIME)` — real `RuntimeVisibleAnnotations` ariadne's parser already reads,
  not compile-time-only markers. No `roots.overrides`/`synthetic_edges`: Quarkus, like
  Micronaut, builds its bean/route wiring at build time via Jandex indexing and bytecode
  recording, not JDK runtime proxies with zero bytecode.
  - **Verified against both existing baselines first**: petclinic and WebGoat still
    auto-detect `spring+servlet` only (`quarkus`'s own detect markers,
    `io.quarkus.runtime.Quarkus`/`io.quarkus.runtime.Application`, are absent from both),
    confirmed by running `bin/ariadne` against both (identical status counts to before this
    profile existed).
  - **Verified against a new, real, Quarkus control app** —
    `scripts/prepare-quarkus-demo.sh` generates and builds a three-class Maven project:
    `demo.YamlResource` (a `jakarta.ws.rs` `@Path("/yaml")`/`@GET` REST resource),
    `demo.ScheduledYamlJob` (`@Scheduled(every = "15m")`), and
    `demo.YamlMessageConsumer` (`@Incoming("yaml-in")`) — all three calling straight into
    `org.yaml.snakeyaml:1.30`'s `Yaml#load`, the same manual-table seed CVE-2022-1471 the
    other control apps use. Needs JDK 17+ (Quarkus 3.x's own jars require it) and pulls
    ~100 transitive dependencies (mostly Vert.x/Netty, Quarkus's own runtime — not anything
    this demo's three classes use directly). Run through `bin/ariadne` unmodified, twice:
    - **`-roots all`**: profile auto-detects `jakarta+quarkus` (not `generic`), confirming
      both profiles' own detect markers fire on the same real app and merge correctly;
      CVE-2022-1471 is `REACHABLE`.
    - **`-roots profile`**: `roots: profile "jakarta+quarkus" (3 entry-point methods)` —
      confirmed directly via a one-off call to `Profile.AppRoots` (not just the log's count)
      to be exactly `demo.YamlResource#get`, `demo.ScheduledYamlJob#run`, and
      `demo.YamlMessageConsumer#consume` — all three annotation families correctly
      recognized as entry points in one run. CVE-2022-1471 is still `REACHABLE`.
- **Not implemented:** RTA and `--max-dep-depth`; CDI's `@Observes` remains unimplemented
  (see the `jakarta` note above). `internal/reach`'s BFS passes stay unchanged (still
  computed once for every CVE, same as Phase 5 left them) — only the root set and the
  framework-package list going into them changed.

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
    input/                   # dirs, jars, Boot fat jar, WAR, EAR; GAV from pom.properties / deps.txt
    cves/                    # GAVs -> OSV -> cves.json
    httpcache/               # disk cache for OSV, diffs, raw files; --offline
    classfile/               # pure-Go .class parser + bytecode walker
    callgraph/               # interning, edges, adjacency
    hierarchy/               # super/interfaces/subtypes, method resolution, annotations
    profile/                 # profile loading, detection, merging
      profiles/              # spring.yaml, servlet.yaml, jakarta.yaml, micronaut.yaml,
                             #   quarkus.yaml, generic.yaml
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
- [x] `go build ./...`, `go vet ./...` and `go test ./...` pass, and `go test` needs no JDK
      and no network.
- [x] ariadne runs without Java installed (the whole point of the own-bytecode-walker
      decision in section 2; nothing in `cmd/ariadne` or any `internal/*` package shells
      out to `java`/`javac`/`javap` — only `scripts/prepare-*.sh` and `scripts/build-
      *fixture*.sh`, which build the *test* apps/fixtures, need a JDK).
- [x] Runs end-to-end on old petclinic from both `app-classes.jar`+deps and the fat jar,
      producing `report.md` + `results.json` — **but not "with the same result"**: the fat
      jar resolves fewer dependency jars (42/75 vs 87/87) since Spring Boot repackaging
      drops some scopes `target/dependency` still has, so it finds fewer CVEs (47 vs 148).
      This is correct, expected behavior given the two inputs genuinely differ, not a bug,
      but it means this criterion as originally phrased doesn't hold and is corrected here.
- [x] Every CVE has exactly one status; `REACHABLE` ones have a path and an entry call
      (`reach.Result`/`EntryEdge`, rendered in the report and `results.json`).
- [x] Petclinic is not "everything REACHABLE" (13/148, not 148/148 — the CHA guard works).
- [x] WebGoat control run shows at least one `REACHABLE` CVE (XStream) — in fact every
      XStream CVE in `vuln_methods.yaml` is REACHABLE via the same 2-hop path.
- [x] Unit tests for the `.class` parser and the bytecode walker on fixtures compiled for
      Java 8 and Java 21 (incl. a lambda edge) — `internal/classfile`, cross-checked against
      `javap` ground truth.
- [x] Unit test for CHA expansion (interface call reaches implementation; JDK types not
      expanded) — `internal/callgraph`.
- [x] Unit test for BFS path reconstruction on a fixture graph — `internal/reach`, on a
      purpose-built 3-jar fixture covering all four statuses, not just CHA.
- [x] Profile auto-detection picks `spring+servlet` for petclinic and `generic` for the
      fixture — `internal/profile`, wired into `cmd/ariadne` (Phase 7); `go test
      ./internal/profile` verifies both, plus merging when several profiles match and
      `-profile-file` overriding/adding one by name. `--roots profile`, synthetic edges
      (Spring Data), and `Roots.Overrides`-based roots (the `servlet` profile) are
      implemented too; the `jakarta` profile is implemented for JAX-RS and EJB (not CDI's
      `@Observes` — see Phase 7's own status note for why); the `micronaut` and `quarkus`
      profiles are implemented too; only RTA and `--max-dep-depth` are not. Gradle input
      (`scripts/gradle-copy-deps.init.gradle`)
      and WAR/EAR input (`internal/input.PrepareAppInput`) *are* implemented — see Phase 7's
      own status note for the line between all of these.
- [x] Report contains the Limitations and "Unidentified jars" sections.
- [x] Re-running with `-offline` and a warm cache works without network, and — beyond what
      this criterion originally asked — is now also verified byte-identical across repeated
      runs (`results.json`), after fixing the map-iteration-order determinism bug Phase 6
      found.

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
