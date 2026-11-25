# How ariadne works

*Polish version: [JAK_TO_DZIALA.md](JAK_TO_DZIALA.md). How to run it: [README](../README.md).*

This page describes what ariadne does with its inputs, step by step, and how to read what it produces. What is supported, and the evidence behind each claim, is in [SUPPORT.MD](SUPPORT.MD). The full development history, with the detailed write-up of every model, is archived in [archive/2026-11-25/docs](../archive/2026-11-25/docs/README-2026-11-25.md).

## 1. The question it answers

ariadne answers two kinds of question about a built JVM application or library:

1. **CVE scan** (`ariadne`): which known vulnerabilities in the dependencies can actually be reached from this code, as opposed to merely being on the classpath.
2. **Research** (`ariadne research …`), with no advisory: is an arbitrary method reachable, from which entry points, and by which path. The answer comes from a frozen analysis that other tools (Tekmerion) can query.

Both start from the same **static pipeline**: a call graph built straight from bytecode. No JDK is needed to run ariadne, because it parses `.class` files itself.

```
 -app, -deps, -deps-list, -roots, -profile …
                     │
                     ▼
        ┌─────── static pipeline (prepare) ───────┐
        │ 1 input and components                  │
        │ 2 bytecode and class hierarchy          │
        │ 3 profiles and roots                    │
        │ 4 call graph and models                 │
        │ 5 reachability engine                   │
        └─────────────────────────────────────────┘
              │                          │
              ▼                          ▼
     CVE stage (ariadne)         research bundle (research index)
     6 advisories (OSV)          → research query / check
     7 CVE → vulnerable methods  → research trace-config / correlate
     8 verdicts
     9 report.md, results.json
```

## 2. The static pipeline

### Step 1. Input and components (`internal/input`, `internal/component`)

`-app` can be a directory of classes, a plain jar, a Spring Boot fat jar, a WAR or an EAR. Archives that carry their own libraries (`BOOT-INF/lib`, `WEB-INF/lib`, EAR modules) supply their dependencies themselves. `-deps` is a directory of dependency jars. `-deps-list` (the output of `mvn dependency:list`) identifies jars that have no `pom.properties`.

Every jar becomes a **component instance** with a Maven identity (group:artifact:version, GAV) and a SHA-256. Components embedded inside another jar (shaded or relocated copies) are found from their metadata and confirmed by comparing their classes with the published jar. `-embedded-components` and `-sbom` can declare such components; a declaration is checked against the classes, never trusted blindly. Advisories are later matched per instance, not per jar name.

### Step 2. Bytecode and the class hierarchy (`internal/classfile`, `internal/hierarchy`)

ariadne parses every `.class` file of the application and its dependencies with its own parser. For each method it records the call instructions (invokevirtual, invokeinterface, invokestatic, invokespecial, invokedynamic), field and allocation sites, annotations and line numbers. The class hierarchy indexes superclasses, interfaces, subtypes and declared methods.

Classes form one flat class path: the application first, then the dependency jars in the order listed (the first copy of a class wins). A losing copy is kept as **shadowed**; its code is never used for a path. Multi-release jars are resolved against `-target-java`.

### Step 3. Profiles and roots (`internal/profile`, `internal/libapi`)

A **profile** holds the framework knowledge: which packages are framework code, which methods the framework calls (controllers, listeners, bean lifecycle methods), and the synthetic edges the framework adds. `-profile auto` (the default) detects the frameworks present (spring, servlet, jakarta, micronaut, quarkus, stripes) and merges their profiles; `generic` has no framework knowledge.

The **roots** are where execution may start (`-roots`):

| `-roots` | Roots |
|---|---|
| `all` (default) | every application method, plus the lifecycle methods of `web.xml`-registered servlets, filters and listeners |
| `profile` | only the entry points the active profiles know (handlers, listeners, beans), plus the same `web.xml` methods |
| `api` | **library mode**: the public API of the `-app` jar. This includes methods its API classes inherit from dependencies and classes the library announces to a framework (Spring, CDI, Micronaut); `-api-exclude` removes packages |

In library mode the library's dependency declarations (its `pom.xml`, or `-library-pom`) say who supplies each dependency version (`supplied_by`). With `-alt-versions`, the other versions a user of the library can get are assessed as well (section 3).

### Step 4. The call graph and its models (`internal/callgraph` and the model packages)

Every method becomes a node. Direct calls become edges. A virtual or interface call becomes edges to every implementation in the hierarchy (**CHA**, the default). `-callgraph-mode rta` narrows these to classes that reachable code actually creates; `pta` narrows further where the receiver's allocations are known.

Calls on JDK and `javax` types are not expanded by CHA, because the JDK would otherwise connect everything to everything. Instead, bounded **models** add the edges the JVM really takes through such calls. Each model is versioned, and most are checked against a real-JVM oracle:

- class initialization (`<clinit>`), lambdas and method references (SAM calls), MethodHandles,
- reflection with known names (literals, values that flow into the call, configuration),
- ServiceLoader and JPMS `provides`, JDBC driver and object dispatch, servlet container dispatch (to the servlets and filters of `web.xml`, `@WebServlet`/`@WebFilter`, Spring Boot, and those start-up code registers with `ServletContext#addServlet`/`addFilter`),
- framework proxies (Spring Data repositories, MyBatis mappers, transactional proxies), Spring AOP advice, Spring events, JPA,
- callbacks and asynchronous execution (threads, executors, CompletableFuture, listeners, reactive),
- deserialization entry points and candidate gadget chains.

Models that depend on each other (one creates objects another dispatches to) run together until nothing changes (a **fixed point**). A model adds an edge only when code that creates the needed object is reachable.

What no model can resolve is recorded as a **model gap**, with a site and a mechanism: unresolved reflection, a native method, an unknown callback, and so on. A gap never adds an edge. It marks where the graph may be incomplete.

### Step 5. The reachability engine (`internal/reach`)

Two breadth-first searches run once from all roots:

- **Pass 1 (`app_only`)** visits framework methods but does not continue through them. Frameworks dispatch through reflection and proxies that a static graph cannot follow faithfully.
- **Pass 2 (`full`)** has no such restriction.

A method found in pass 1 is reachable without passing through framework internals. A method found only in pass 2 is reachable only through framework code. Paths are reconstructed shortest-first, preferring steps that do not depend on CHA alone. Every step of a path carries its evidence: the edge's basis, its character (`jvm_rule`, `library_model`, `input_declaration`, `heuristic`) and its call site.

This is where the static pipeline ends (`prepare` in `cmd/ariadne`). `ariadne` continues with the CVE stage; `ariadne research index` writes the state to a bundle instead.

## 3. The CVE stage

### Step 6. Advisories (`internal/cves`)

The identified components are sent to OSV (`api.osv.dev`), in batches. The answers are cached on disk (`-cache-dir`, default `.cache`) with their fetch time:

- `-offline` serves only from the cache.
- `-data-ttl`, `-refresh`, `-allow-stale` and `-max-data-age` control freshness.
- `-snapshot-out` writes everything a run read to a self-checking snapshot that `-snapshot-in` replays without network, and `-snapshot-compare` explains differences between runs.

### Step 7. From an advisory to vulnerable methods (`internal/vulnmap`)

An advisory names a package and versions, not code. ariadne finds the vulnerable methods from these sources, in order:

1. the embedded manual table (`internal/vulnmap/vuln_methods.yaml`), curated with version variants,
2. the methods changed by the advisory's fix commits (GitHub; `GITHUB_TOKEN` raises the rate limit),
3. for an advisory with no fix commit: the methods whose bytecode changed between the last release before the fix and the fixing patch release, in the classes the advisory names, read from Maven Central (`-release-diff`, on by default),
4. opt-in, `-llm`: a language model for advisories the other sources could not map.

### Step 8. Verdicts

Each advisory gets one verdict per affected component instance. In results v1 there is one row per advisory.

| Verdict | Meaning |
|---|---|
| `REACHABLE` | A vulnerable method is reached in pass 1. The shortest path and its first call from application code into the dependency are reported. |
| `FRAMEWORK` | A vulnerable method is reached only in pass 2, through framework code. This needs review. |
| `NOT_REACHABLE` | The mapped methods exist in the graph but no root reaches them. **Read this as "no path in the model", not "safe".** Reasons say why the verdict may be weaker: `MAPPING_PARTIAL`, `MODEL_INCOMPLETE` (reachable model gaps could dispatch into the component), `MAPPED_CODE_ABSENT`. |
| `UNKNOWN` | There is no mapping, or no mapped method exists in the analyzed version (`METHOD_NOT_FOUND`), or the instance's class lost to another copy (`CLASSPATH_CONFLICT`). |
| `NOT_APPLICABLE` | The component has no JVM code (a WebJar), or the advisory concerns no JVM code at all. |

Further assessments are attached without changing the verdicts:

- **target JDK advisories** (`-jdk-home` / `-runtime-file`): applicability only,
- **other dependency versions** in library mode (`-alt-versions`):
  - `approx`: that version's advisories, assessed on the analyzed code,
  - `rerun`: the whole static pipeline runs again with that version's jar,
- **exploit and framework checks** by a language model (`-exploit-check`, `-framework-check`), as a second opinion,
- **runtime evidence** from a recorded run (`-runtime-trace`, section 5).

### Step 9. Outputs (`internal/report`)

- `report.md`: a human-readable report with every finding, its path, reasons, model gaps and coverage.
- `results.json`: deterministic machine-readable results. `v1` (default) has one row per advisory. `v2` (`-json-format v2`) has components, one finding per advisory × instance, coverage, frameworks and the runtime section. The schema is in [RESULTS_CONTRACT.MD](RESULTS_CONTRACT.MD).
- Optionally: the call graph (`-graph-out`), a comparison with an earlier graph (`-graph-compare`), a JFR filter for the run's vulnerable methods (`-runtime-filter-out`) and a snapshot.

All outputs are written together, or none is.

## 4. The research interface

`ariadne research index` runs the same static pipeline with no advisory and no network. It writes a **bundle**: a directory with a manifest, an identity (inputs with hashes, configuration, coverage limits) and one JSONL dataset per record type:

- components,
- classes (selected and shadowed),
- graph nodes (with receiver specializations and source locations),
- edges (with call sites and conditions),
- roots,
- model gaps,
- every bytecode call site.

`-scope reachable` keeps only what the roots or the input's own methods reach. `-sources` adds the SHA-256 of each method's source file.

`ariadne research query` answers one request from the bundle alone. The operations are:

- `entries`, `methods`, `calls`: what the API is and what it calls,
- `callers`, `callees`: one step of the graph,
- `reachability`: up to 256 target methods per request. Each target is resolved (`matched`, `missing`, `shadowed`, `ambiguous`), searched within a work budget, and returned with its entry points, a witness path per entry with every edge as evidence, and the model gaps relevant to it.

The contract (identifiers, schemas, rules, errors) is owned by ariadne and documented in [research-interface](research-interface/README.md). Two further commands support the consumer:

- `ariadne research check` gives the producer's verdict on any response. The acceptance package of real responses and deliberately broken ones is described in [ACCEPTANCE.md](research-interface/ACCEPTANCE.md).
- `research trace-config` and `research correlate` (experimental) connect a recorded run to the bundle's nodes; see [RUNTIME.md](research-interface/RUNTIME.md).

## 5. Runtime evidence

A static graph can over-approximate (CHA) and can miss what no model covers. A recorded run gives independent evidence:

1. `-runtime-filter-out` (CVE scan) or `research trace-config` (research) writes JFR settings that trace exact entries of chosen methods (`jdk.MethodTrace`, JDK 25+).
2. The application runs under Java Flight Recorder with those settings. `scripts/JfrToTrace.java` converts the recording to the trace format [`ariadne-runtime-trace-1`](RUNTIME_TRACE.MD): loaded classes, executed methods and measured caller → callee relations.
3. `-runtime-trace` (CVE scan) or `research correlate` attaches the trace to the analyzed classes **by content hash**. A class with other bytes gets nothing. A measured call the graph has no edge for is classified by cause (reflection, proxy, async, skipped receiver, …).

Runtime evidence never changes an edge or a verdict. **The absence of an event proves nothing**, because another scenario may run the code. An entry proves that the code ran, not that attacker-controlled data reached it.

## 6. How to read the results

- `REACHABLE` and `FRAMEWORK` come with a path you can check step by step. Steps marked `heuristic` depend on CHA alone and may be over-approximations.
- `NOT_REACHABLE` is conditional on the model. Check its reasons and the model gaps, and remember that the inputs must be the deployed class path.
- In library mode a finding means that **a user of the library's API** can reach the vulnerable code. `exposed_api` lists the API methods that do, and `supplied_by` says who chooses the dependency version.
- Everything ariadne knows it could not see (gaps, limits, unresolved dispatch, native boundaries) is listed in the report's coverage sections rather than hidden.
