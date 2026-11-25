# ariadne
<p align="center"><img src="docs/ariadne-icon.png" width="160" alt="Ariadne logo"></p>

CVE reachability analysis for JVM applications and libraries. ariadne takes a built Java
application (or library) and its resolved dependencies, builds a call graph straight from
bytecode, matches the dependencies against known vulnerabilities (OSV), and reports which
vulnerabilities are **reachable** from the code, as opposed to merely present on the classpath.
Without any advisory, it can also answer whether an arbitrary method is reachable (the
research interface).

**How it works, step by step:** [English](docs/HOW_IT_WORKS.md) · [Polski](docs/JAK_TO_DZIALA.md)

| Document | Content |
|---|---|
| [docs/HOW_IT_WORKS.md](docs/HOW_IT_WORKS.md) / [docs/JAK_TO_DZIALA.md](docs/JAK_TO_DZIALA.md) | The pipeline, the verdicts and how to read the results (EN / PL) |
| [docs/SUPPORT.MD](docs/SUPPORT.MD) | What is supported, with limits and the evidence behind each claim |
| [docs/RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD) | `results.json` v1/v2 schema and the meaning of statuses and reasons |
| [docs/research-interface](docs/research-interface/README.md) | `ariadne research`: contract, schemas, acceptance package, runtime extension |
| [docs/RUNTIME_TRACE.MD](docs/RUNTIME_TRACE.MD) | Runtime trace format and how to record one |
| [RECOMMENDATION](RECOMMENDATION/README.MD) | Plans still open and the implementation log |
| [archive](archive/2026-11-25/docs/README-2026-11-25.md) | History: the original plan, FAQ, earlier README with every model described in detail |

## Requirements and build

- Go 1.22 or later. Nothing else is needed to **run** ariadne: no JDK, no Maven.
- A JDK is needed only to build the applications you analyze, to rebuild the test fixtures,
  and (JDK 25+) to record runtime traces.

```bash
make build        # → bin/ariadne   (same as: go build -o bin/ariadne ./cmd/ariadne)
bin/ariadne -version
```

## Quick start

```bash
# 1. Build your application and collect its runtime dependencies (Maven):
mvn -q package -DskipTests
mvn -q dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory=target/dependency
mvn -q dependency:list -DincludeScope=runtime -DoutputFile=target/deps.txt

# 2. Scan:
bin/ariadne -app target/classes -deps target/dependency -deps-list target/deps.txt \
  -out report.md -json results.json

# 3. Read report.md (findings with paths) or results.json (machine-readable).
```

The first run queries OSV and GitHub over the network and fills the cache (`.cache/`); later
runs can use `-offline`.

## Preparing the input

| What you have | `-app` | `-deps` |
|---|---|---|
| Maven build | `target/classes` or the built jar | `mvn dependency:copy-dependencies -DincludeScope=runtime` output; add `-deps-list` from `mvn dependency:list` |
| Gradle build | the built jar or `build/classes/java/main` | copy the runtime classpath with `scripts/gradle-copy-deps.init.gradle` (also writes a `-deps-list` file) |
| Spring Boot fat jar | the fat jar | not needed (`BOOT-INF/lib` is read) |
| WAR / EAR | the archive | not needed (`WEB-INF/lib`, EAR modules); `-deps-list` still helps identify jars |
| A library (jar others depend on) | the library jar | its runtime dependencies; use `-roots api` (see [Library mode](#library-mode)) |

`-deps-list` identifies jars that carry no `pom.properties`. Without identification a jar is
analyzed but cannot be matched to advisories.

## Scanning for CVEs

```bash
bin/ariadne -app APP -deps DEPS_DIR [-deps-list DEPS.txt] -out report.md -json results.json [flags]
```

Common variants:

```bash
# Spring Boot fat jar
bin/ariadne -app target/myapp-1.0.0.jar -out report.md -json results.json

# Only the framework's entry points as roots, v2 results
bin/ariadne -app ... -deps ... -roots profile -json-format v2 -out report.md -json results.json

# Narrower call graph (opt-in): classes actually created by reachable code
bin/ariadne -app ... -deps ... -callgraph-mode rta -out report.md -json results.json

# No network: everything from the cache
bin/ariadne -app ... -deps ... -offline -out report.md -json results.json
```

**Exit status:** `0` the analysis completed (whatever the verdicts), `1` the analysis failed
(input, network, cache or data error; nothing is written), `2` invalid flags.

### All flags

| Flag | Required | Meaning |
|---|---|---|
| `-app` | yes | A directory, a plain jar, a Spring Boot fat jar, a WAR, or an EAR. |
| `-deps` | no* | Directory of dependency jars. *Not needed for a fat jar/WAR/EAR, which supply their own.* |
| `-deps-list` | no | `mvn dependency:list` output (or its Gradle equivalent, see `scripts/gradle-copy-deps.init.gradle`), used as a fallback for jars with no `pom.properties`. |
| `-vuln-methods` | no | Override the embedded `vuln_methods.yaml`. |
| `-release-diff` | no | Default `true`. For an advisory with no fix commit: the methods changed between the last release before the fix and the fixing patch release, in the classes the advisory names, read from Maven Central through the cache (source `release-diff`, unreviewed). `-release-diff=false` turns it off. |
| `-embedded-components` | no | YAML file declaring components embedded in dependency jars and the package relocations they sit under, pinned to each jar's SHA-256 (CVE:F7a; example: `testdata/embedded/petclinic-ehcache.yaml`). Declarations are verified against the classes; one for a jar with another hash is reported and ignored. |
| `-sbom` | no | CycloneDX JSON SBOM (spec 1.4–1.6; CVE:F6b; example: `testdata/sbom/webgoat-jruby.cdx.json`). A component whose SHA-256 is an analyzed dependency jar's is tied to it; Maven components nested in it become embedded candidates of that jar, verified by their classes. Untied components and the dependency graph are not used; a disagreeing identity is reported, never applied. |
| `-profile` | no | `auto` (default, auto-detect and merge), `generic`, `spring`, `servlet`, `jakarta`, `micronaut`, `quarkus`, `stripes`, or a name from `-profile-file`. |
| `-profile-file` | no | Load/override one profile from a YAML file (same shape as the embedded defaults). |
| `-roots` | no | `all` (default: every application method, plus the lifecycle methods of `web.xml`-registered servlets/filters/listeners), `profile` (the active profile's own entry points, plus the same `web.xml` methods) or `api` (library mode, see [Library mode](#library-mode): the `-app` jar is a library and its public API is the roots). |
| `-alt-versions` | no | With `-roots api`: `off` (default) or `newest` — also assess other versions of the library's dependencies a user can get (LIBRARY_MODE LB3 `lb3-2`): the highest published version in a declared range and the newest of the analyzed version's major.minor line. v2 `coverage.api.alternate_versions` only. |
| `-alt-versions-mode` | no | How `-alt-versions` assesses a version: `approx` (default: its advisories and mapping variant, reachability on the analyzed version's code) or `rerun` (`lb3-3`: the static pipeline runs again with that version's jar in place of the analyzed one, the advisories assessed on its own code; the jar from the cache, `ARIADNE_MAVEN_LOCAL` or Central). |
| `-library-pom` | no | With `-roots api`: the library's `pom.xml`, read for its dependency declarations (LIBRARY_MODE LB3) — for a library jar that embeds none (e.g. a Gradle build); default: the pom in the jar's `META-INF/maven/`. |
| `-api-exclude` | no | With `-roots api`: comma-separated package prefixes left out of the API (`com.example.internal`). |
| `-target-java` | no | Java feature version (e.g. `17`) to resolve multi-release JAR variants against (JVM_GRAPH G4). Unset: every jar's base classes, regardless of any `META-INF/versions/` present. |
| `-jpms-scenario` | no | ServiceLoader launch scenario (SERVICE_LOADING S3): `classpath` (default — every jar in the unnamed module, `module-info.class` ignored, providers from `META-INF/services`) or `module-path` (every jar, the app included, on the module path: explicit modules provide only through `provides`, a provider with `public static provider()` is created through it, automatic modules provide through `META-INF/services` with a constructor, and a caller in an explicit module must declare `uses`), or `mixed` (only the `-jpms-module-path` jars on the module path, every other jar on the class path; module-path modules assumed resolved as with `--add-modules ALL-MODULE-PATH`). Recorded in the report header; `results.json`'s format is unchanged. |
| `-jpms-module-path` | with `mixed` | Comma-separated jar file-name patterns (`path.Match` syntax, e.g. `app,commons-text-*.jar`) selecting the module-path jars for `-jpms-scenario mixed`; `app` names the analyzed application. Rejected with any other scenario; a pattern matching no jar is a warning. |
| `-active-profiles` | no | Comma-separated Spring profile names to treat as active when evaluating `@Profile` conditions. Unset: every `@Profile` condition is reported `unknown`, never a guessed `false`. |
| `-callgraph-mode` | no | Dispatch strategy for virtual/interface calls (GRAPH_PRECISION P0/P2/P3): `cha` (the default), `rta` (opt-in: CHA edges narrowed to the classes created by code reachable from the roots, with the unknown-allocation fallback) or `pta` (experimental: RTA, then every CHA call site whose receiver's allocated classes are proven keeps only the targets they select). Recorded with the graph's other settings in the report header and v2 `callgraph`. |
| `-graph-out` | no | Write the final call graph — every edge with a stable identity (source/target method, call kind, basis) and its call-site PCs, the graph's settings, the findings' paths and the created-type catalog (GRAPH_PRECISION P1) — as sorted text (`ariadne-callgraph-2`; large: ≈ 370 MB for WebGoat). |
| `-jdk-home` | no | The target runtime's installation directory (CVE:F8a): its `release` file gives vendor, version, OS and architecture; nothing is run, and the scanner's own JDK is never assumed. Report header "Target runtime", v2 `target_runtime`. For a Temurin runtime the OpenJDK Vulnerability Group advisories of the curated catalog are matched (CVE:F8b): report section "Target runtime advisories", v2 `target_runtime.advisories` — applicability only, reachability `not_assessed`, never part of the v1 verdicts; any other distribution is `not_performed` with the reason. The embedded catalog (version 1, curated 2026-11-20) holds the eight advisories from 2026-01-20 to 2026-11-17; a runtime older than the first of them gets a note that earlier advisories are not included. Exclusive with `-runtime-file`. See [docs/RUNTIME_SOURCES.MD](docs/RUNTIME_SOURCES.MD). |
| `-openjdk-catalog` | no | Use this curated OpenJDK advisory catalog (YAML from `go run ./cmd/openjdk-catalog`) instead of the embedded one — to check a draft before committing it. ariadne never fetches openjdk.org (its terms of use do not allow automated access); see [docs/RUNTIME_SOURCES.MD](docs/RUNTIME_SOURCES.MD). |
| `-runtime-file` | no | A YAML description of the target runtime (`vendor`, `distribution`, `version`, `os`, `arch`, `libc`, `package: {name, version, distro}`; unknown fields are errors), e.g. for a runtime that is not at hand. Exclusive with `-jdk-home`. |
| `-runtime-trace` | no | Import a runtime trace (RUNTIME_NATIVE N1, [docs/RUNTIME_TRACE.MD](docs/RUNTIME_TRACE.MD); e.g. from `scripts/JfrToTrace.java`): loaded classes, executed methods and measured calls, matched to the analyzed classes by content hash and shown next to the results (report "Runtime evidence", v2 `runtime`, `findings[].runtime_observed`); never changes the graph or a verdict. Validated before the analysis; recorded as a snapshot input. |
| `-runtime-filter-out` | no | Write a JFR settings file (`.jfc`) tracing on entry (`jdk.MethodTrace`, JDK 25+) the mapped vulnerable methods of this run's findings whose class is in the input; record the next run with `-XX:StartFlightRecording=settings=profile,settings=FILE` and import it with `-runtime-trace`. |
| `-runtime-filter-status` | no | With `-runtime-filter-out`: comma-separated verdicts whose mapped methods are traced (default `REACHABLE,FRAMEWORK,NOT_REACHABLE,UNKNOWN`). |
| `-graph-compare` | no | Compare this run's graph with a `-graph-out` file: edges added/removed by basis, findings whose verdict or path changed and the edges that explain it, and any settings difference (report "Call graph compared with …", v2 `callgraph.comparison`). |
| `-taint` | no | DATA_FLOW D4 taint pilot (opt-in, default off, report-only): assess whether request data (Spring MVC / JAX-RS handler parameters, `ServletRequest` accessors) reaches the data each deserialization entry reads. Never changes a verdict. |
| `-out` | yes | `report.md` output path. |
| `-json` | yes | `results.json` output path. |
| `-json-format` | no | `v1` (default: one row per advisory) or `v2` (CVE:F1c: components, one finding per advisory × component instance, coverage, inputs reference — schema in [docs/RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD) §5). Since 2026-11-23 v2 also carries `frameworks`: the profile's beans, `@Profile` conditions, injection points and their resolution, entry points and framework roots (facts only, never part of reachability). |
| `-cache-dir` | no | Disk cache directory for OSV/GitHub HTTP responses (default `.cache`). |
| `-offline` | no | Serve only from the cache; a miss is a clear, immediate error. Old data is used with a warning (see `-max-data-age`). |
| `-data-ttl` | no | How long an OSV/advisory response counts as fresh before it is fetched again (default `24h`). |
| `-refresh` | no | Revalidate every cached OSV/advisory response now, whatever its age; never re-asks an LLM. Not with `-offline`. |
| `-allow-stale` | no | When an online refresh fails, use the cached copy and mark the report "stale" (never past `-max-data-age`). |
| `-max-data-age` | no | Hard limit on the age of any OSV/advisory data used, offline included; data of unknown age fails it. |
| `-snapshot-out` | no | Write a data snapshot to a new directory ([Data snapshots](#data-snapshots)). |
| `-snapshot-in` | no | Replay a data snapshot: no cache, no network; inputs, analyzer and algorithm versions must match. Not with `-refresh`/`-snapshot-out`. |
| `-snapshot-compare` | no | Compare this run with an earlier snapshot: which advisories' derived results were invalidated, and why (report section and v2 `coverage.snapshot_changes`). |
| `-llm-reask-changed` | no | With `-llm`: ask again when a CVE's advisory text or fix diff changed since its cached answer. Default: reuse the previous answer with a warning, so a data refresh never triggers paid requests by itself. |
| `-llm` | no | Enable the LLM source for CVEs the manual table and fix-commit diffs could not map. Needs a model and `ARIADNE_LLM_API_KEY`. |
| `-llm-model` | no | LLM model id; overrides `ARIADNE_LLM_MODEL` (e.g. `z-ai/glm-5.3`). |
| `-llm-base-url` | no | OpenAI-compatible base URL; overrides `ARIADNE_LLM_BASE_URL` (default `https://openrouter.ai/api/v1`). |
| `-exploit-check` | no | For each REACHABLE CVE, ask an LLM whether its call path plausibly carries attacker-controlled input. Uses the same `ARIADNE_LLM_*` settings as `-llm` (and works without `-llm` itself); one extra request per REACHABLE CVE. |
| `-framework-check` | no | For each FRAMEWORK CVE, ask an LLM whether its call path looks like a real framework call chain or a CHA over-approximation artifact. Same `ARIADNE_LLM_*` settings; one extra request per FRAMEWORK CVE. |
| `-version` | no | Print the ariadne version and exit. |

### Environment variables

| Variable | Meaning |
|---|---|
| `GITHUB_TOKEN` | Authenticates GitHub requests for fix-commit metadata (higher rate limit). |
| `ARIADNE_MAVEN_LOCAL` | A local Maven repository (e.g. `~/.m2/repository`) read before the network for poms and reference jars (embedded components, library poms, `-alt-versions-mode rerun`). Only files downloaded from a remote repository are used. |
| `ARIADNE_MAVEN_CENTRAL` | Repository root used instead of Maven Central for reference jars. |
| `ARIADNE_LLM_API_KEY`, `ARIADNE_LLM_MODEL`, `ARIADNE_LLM_BASE_URL` | Settings for `-llm`, `-exploit-check`, `-framework-check` (OpenAI-compatible endpoint; default `https://openrouter.ai/api/v1`). The key is read only from the environment; ariadne never reads `.env` itself: `set -a; . ./.env; set +a` (template: `.env.example`). |

### Network, cache and freshness

ariadne contacts `api.osv.dev` (advisories), GitHub (fix commits) and Maven Central (version
lists, reference jars, release diffs). Every response is cached under `-cache-dir`
(default `.cache`).

- `-offline` uses only the cache; a missing entry is an immediate error.
- `-data-ttl` (default `24h`) decides when advisory data is fetched again. `-refresh` revalidates everything now. `-allow-stale` falls back to the cache when a refresh fails. `-max-data-age` is a hard limit.
- Maven Central rate-limits heavy use: run online once, then `-offline`, and set `ARIADNE_MAVEN_LOCAL` so poms and jars come from your local repository.

### Data snapshots

```bash
bin/ariadne ... -snapshot-out snap/       # record exactly the data this run read
bin/ariadne ... -snapshot-in snap/        # replay it: no cache, no network
bin/ariadne ... -snapshot-compare snap/   # explain, per advisory, what changed since
```

A replay requires the same analyzer version and the same input artifacts (their hashes are
checked). A different configuration is allowed and reported.

### Library mode

For a jar others depend on, scan its public API with its runtime dependencies:

```bash
mvn -q dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory=target/deps
mvn -q dependency:list -DincludeScope=runtime -DoutputFile=target/deps.txt
bin/ariadne -app target/mylib-1.2.3.jar -deps target/deps -deps-list target/deps.txt \
  -roots api -json-format v2 -out ariadne.md -json ariadne.json
```

- `-api-exclude com.example.mylib.internal` leaves public-but-internal packages out of the API.
- The library's own pom (inside the jar, or `-library-pom pom.xml`) says who supplies each dependency (`supplied_by`). Set `ARIADNE_MAVEN_LOCAL` so parent poms come from the local repository.
- A finding means **a user of the library's API can reach the vulnerable code**. `exposed_api` lists the API methods that do.
- `-alt-versions newest` also assesses the other dependency versions a user can get. With `-alt-versions-mode approx` (the default) they are assessed on the analyzed code. With `-alt-versions-mode rerun`, the whole graph is rebuilt with each version's jar.

### Runtime evidence (optional, JDK 25+)

```bash
bin/ariadne ... -runtime-filter-out vulnerable.jfc        # JFR settings for the findings' methods
java -XX:StartFlightRecording=filename=run.jfr,settings=profile,settings=vulnerable.jfc -jar app.jar
#   … exercise the application, stop it …
java scripts/JfrToTrace.java --jfr run.jfr --out run.trace --scenario "what was exercised" \
  --artifact app.jar --artifact-dir deps/
bin/ariadne ... -runtime-trace run.trace                   # shown next to the static results
```

The trace never changes a verdict. A method not seen running proves nothing. See
[docs/RUNTIME_TRACE.MD](docs/RUNTIME_TRACE.MD).

### Language-model options

`-llm` maps advisories that the manual table, fix commits and release diffs could not map.
`-exploit-check` and `-framework-check` give a second opinion on each `REACHABLE` or
`FRAMEWORK` path. All three are opt-in, need the `ARIADNE_LLM_*` variables, and cache their
answers under the cache directory (with `-offline` only cached answers are used).

## Research interface (no advisories)

```bash
# Index once: the static pipeline only, offline, no advisory.
bin/ariadne research index -app library.jar -deps deps/ -roots api -profile generic -out bundle/
#   optional: -scope reachable (smaller bundle), -sources src/main/java (source hashes)

# Ask questions: entries, methods, calls, callers, callees, reachability.
bin/ariadne research query -bundle bundle/ -request request.json -out response.json [-verify full]

# The producer's verdict on any response (for a consumer's acceptance tests).
bin/ariadne research check -bundle bundle/ -request request.json -response response.json [-replay]

# Runtime evidence for chosen targets (experimental, JDK 25+ for recording).
bin/ariadne research trace-config -bundle bundle/ -request targets.json -out config.json -jfc trace.jfc
bin/ariadne research correlate -bundle bundle/ -trace run.trace -config config.json -out correlation.json
```

`research index` accepts the static flags of the scan (`-app`, `-deps`, `-deps-list`,
`-roots` (required), `-profile`, `-profile-file`, `-target-java`, `-active-profiles`,
`-jpms-scenario`, `-jpms-module-path`, `-api-exclude`, `-library-pom`, `-callgraph-mode`,
`-embedded-components`, `-sbom`, `-cache-dir`) plus `-scope` and `-sources`. It never uses
the network and refuses an existing `-out`. Request and response formats, examples and rules:
[docs/research-interface](docs/research-interface/README.md).

Exit status of `research query`: `0` an answer (including `missing`, `not_found`,
`incomplete`), `2` an invalid argument or request, `1` a bundle, I/O or internal failure.
`research check`: `0` valid, `1` a broken rule, `2` invalid arguments.

## Outputs

| File | Content |
|---|---|
| `report.md` (`-out`) | Findings with verdict, reasons, example path step by step, entry call, model gaps; coverage, frameworks, runtime evidence |
| `results.json` (`-json`) | `v1`: one row per advisory. `v2`: components, findings per advisory × component instance, coverage, frameworks, runtime. Schema: [RESULTS_CONTRACT.MD](docs/RESULTS_CONTRACT.MD) |
| `-graph-out FILE` | The call graph as sorted text (large) |
| `-snapshot-out DIR` | The data the run read, for an exact replay |

The verdicts (`REACHABLE`, `FRAMEWORK`, `NOT_REACHABLE`, `UNKNOWN`, `NOT_APPLICABLE`) are explained in
[How it works § 8](docs/HOW_IT_WORKS.md#step-8-verdicts). `NOT_REACHABLE` means "no path in the
model", not "safe".

## Development

```bash
make test         # go test ./...  (no JDK, no network: fixtures are committed jars)
make vet
make fixtures     # rebuild testdata fixtures (needs a JDK)
```

- `cmd/ariadne`: the CLI (`analyze` = `prepare` + `scanCVEs`, and `research`).
- `internal/`: one package per step and per model (see [How it works](docs/HOW_IT_WORKS.md)).
- `testdata/`: committed fixtures, each with an `EXPECTED.MD`, many with real-JVM oracles.
- `scripts/`: preparing the reference applications (`prepare-*.sh`), JVM oracles (`*-oracle.sh`), measurements (`pipeline-compare.sh`, `graph-benchmark.sh`, `runtime-trace-*.sh`), fixture builders.
- `work/` (target applications) and `.cache/` are git-ignored.
