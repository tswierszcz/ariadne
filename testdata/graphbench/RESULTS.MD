# GRAPH_PRECISION P4 — call-graph strategy benchmark (CHA vs RTA)

Measured 2026-11-08 with `scripts/graph-benchmark.sh --repeat 3` (real applications) and
`go test ./cmd/ariadne -run TestGraphBenchmarkFixtures -v` (fixtures with a real-JVM
oracle). **A measurement, not ground truth:** fewer edges or fewer `REACHABLE`/`FRAMEWORK`
results are not an improvement by themselves; every verdict change below is explained by
hand from the bytecode, not by an LLM's opinion of a shorter path.

| File | Content |
|---|---|
| `results.tsv` | one row per app × roots scenario × mode: wall time and peak RSS (min–max of 3 measured runs), graph size, CHA edges removed / kept only by the unknown-allocation fallback, live classes, active unknown-allocation gaps, created-type catalog size and gaps, model-gap sites (all / reachable), v1 verdict counts `R/F/NR/U/NA`, and against the CHA run of the same scenario the edges removed/added and the findings whose verdict or path changed |
| `verdict-changes.tsv` | every verdict change against CHA, with the removed edges that explain it (from `-graph-compare`) |
| `versions.txt` | analyzer commit, Go version, and per mode the graph settings and rule versions: builder `1`, skip rule (`java.`, `javax.`, `jdk.`, `sun.` receivers not expanded), unknown-allocation policy, catalog `p1-1`, RTA `p2-1` |

## Setup

- Inputs: the prepared petclinic (Spring Boot 2.4.5, embedded Tomcat 9.0.45) and WebGoat
  (Spring Boot, Undertow) of `scripts/prepare-petclinic.sh` / `scripts/prepare-webgoat.sh`,
  `-profile spring`, offline on the same OSV cache, `-deps-list` given.
- Each configuration: 3 measured runs that write and read no graph file, plus 1 unmeasured
  run that writes (CHA) or compares against (RTA) the graph — reading a several-hundred-MB
  graph file costs ~1 GB of memory, which must not be charged to the strategy. Every figure
  other than time and memory was identical across the 4 runs (the script checks).
- `versions.txt` names a dirty commit: the uncommitted changes at measurement time were this
  phase's benchmark script, tests and the snapshot algorithm versions — none of them in the
  analysis path.
- No external reference engine (e.g. SootUp) was used: none is pinned in this repository,
  and its result would not be ground truth either. The ground truth available is the
  real-JVM oracles of the fixtures and hand-checked bytecode for each verdict change.

## Fixtures (real-JVM oracle)

`TestGraphBenchmarkFixtures`: per fixture, the *traced* method bodies (those that record
when they run) the JVM ran (`observed`), how many traced bodies the graph reaches from
`main`, the observed bodies it does not reach (**missed expected targets**) and the reached
traced bodies the JVM did not run (**extra** — an over-approximation, or code this run did
not exercise).

| Fixture | Roots | Mode | Observed | Reached | Missed | Extra |
|---|---|---|---|---|---|---|
| `testdata/callgraph` (overloads, default method, bridges, inherited body, super call) | profile / all | cha | 13 | 13 of 14 | 0 | 0 |
| | profile / all | rta | 13 | 13 of 14 | 0 | 0 |
| `testdata/rta` (3 implementations, factory, dispatch cycle with late allocation, ServiceLoader provider, reflective factory) | profile | cha | 9 | 11 of 11 | 0 | 2 (`Ghost#area` never created, `Hidden#area` created only in uncalled code) |
| | profile | rta | 9 | 8 of 11 | 1 (`Unnamed#area`: created by reflection from a name built at run time, named nowhere — covered by the active `reflection-unresolved` gap) | 0 |
| | all | cha | 9 | 11 of 11 | 0 | 2 (same) |
| | all | rta | 9 | 11 of 11 | 0 | 2 (with `-roots all`, `total(Shape)` is a root: any `Shape` may come in) |

The test fails if CHA misses anything, if RTA misses anything but the documented case, or if
RTA reaches a body CHA does not.

## Real applications

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 13.7–17.9 | 1335–1367 | 991 328 | — | — | 596 of 1096 | 3/71/16/57/6 |
| petclinic | profile | rta | 14.1–14.7 | 1315–1385 | 938 767 | 52 561 of 310 418 (50 305) | 15 818 (7 562) | 532 of 1097 | 3/65/22/57/6 |
| petclinic | all | cha | 13.5–14.9 | 1387–1394 | 991 328 | — | — | 596 of 1096 | 3/71/16/57/6 |
| petclinic | all | rta | 13.9–14.0 | 1359–1435 | 985 215 | 6 113 of 310 418 (78 194) | 21 603 (11 609) | 594 of 1097 | 3/71/16/57/6 |
| WebGoat | profile | cha | 23.1–23.4 | 1779–1829 | 1 409 574 | — | — | 676 of 1194 | 42/51/10/61/0 |
| WebGoat | profile | rta | 23.8–24.1 | 1783–1806 | 1 314 196 | 95 378 of 556 769 (178 188) | 24 398 (11 752) | 590 of 1201 | 42/49/12/61/0 |
| WebGoat | all | cha | 23.1–23.4 | 1772–1810 | 1 409 575 | — | — | 677 of 1194 | 42/51/10/61/0 |
| WebGoat | all | rta | 23.8–24.3 | 1776–1819 | 1 394 752 | 14 823 of 556 769 (122 892) | 32 123 (13 998) | 680 of 1201 | 42/51/10/61/0 |

- **Cost:** RTA adds about 0.4–0.7 s; peak memory is within run-to-run noise.
- **Edges:** `-roots profile` removes 17% of the CHA edges in both apps; `-roots all` 2–3%.
  No edge is ever added. Under RTA, 1 (petclinic) / 7 (WebGoat) more gap sites exist — the
  `Unsafe`/constructor-handle `allocation` gaps RTA introduces.
- **Unknown sites:** the fallback carries a large share of what RTA keeps (in `-roots
  profile`, 50 305 of the 257 857 CHA edges kept on petclinic, 178 188 of 461 391 on
  WebGoat): most of the narrowing RTA could do is given back to unresolved reflection
  (130–159 reachable sites).
- **Verdicts:** `-roots all` (the default roots) changes no verdict in either app. `-roots
  profile` moves 6 (petclinic) and 2 (WebGoat) findings from `FRAMEWORK` to
  `NOT_REACHABLE`, all carrying `MODEL_INCOMPLETE`. No `REACHABLE` result changes anywhere.

## Every verdict change, checked by hand

`FRAMEWORK` → `NOT_REACHABLE` under RTA, `-roots profile` (`verdict-changes.tsv`):

| App | Advisory | Vulnerable code | Why CHA reached it | Check | Verdict |
|---|---|---|---|---|---|
| petclinic | CVE-2022-45143 | `JsonErrorReportValve#report` | `ErrorReportValve#invoke` → CHA to the subclass | No class of the input references `JsonErrorReportValve` (bytecode scan of every jar, slashed and dotted names): Tomcat creates valves only from `server.xml`/`context.xml`, which embedded Spring Boot does not use | **correct** |
| petclinic | CVE-2025-55752, CVE-2026-53404 | `RewriteValve#invoke/#parse`, `RewriteRule#evaluate` | `LifecycleBase#start` → CHA to `RewriteValve#startInternal` | Same: `RewriteValve` is referenced by no other class; configuration-only | **correct** |
| petclinic | CVE-2026-55957 | `JNDIRealm#bindAsUser` | `BasicAuthenticator#doAuthenticate` → CHA to `JNDIRealm#authenticate` | Only `MBeanFactory#createJNDIRealm` (a JMX management operation, unreachable) does `new JNDIRealm`; realms otherwise come from `server.xml` | **correct** |
| WebGoat | CVE-2021-22118 | reactive `DefaultPartHttpMessageReader` | WebFlux `DefaultServerWebExchange#initFormData` → CHA into the reactive multipart readers | The reactive stack is present in `spring-web` but WebGoat is a servlet MVC application: no reachable code creates a WebFlux exchange or these readers | **correct** |
| petclinic | CVE-2024-22259 | `UriComponentsBuilder#fromHttpUrl/#fromOriginHeader` | WebFlux `DefaultWebFilterChain#invokeFilter` → CHA to `CorsWebFilter#filter` | The CHA path was an artifact (WebFlux, unused). Whether a servlet path exists cannot be told from either graph: servlet request dispatch is cut in both modes (see below) | artifact removed; real reachability **undetermined** |
| petclinic, WebGoat | CVE-2024-38809 | `ServletWebRequest#validateIfNoneMatch` (ETag parsing) | reactive `ChannelSendOperator$WriteBarrier#emitCachedSignals` → CHA on `Subscriber#onNext` to the servlet-functional `PublisherEntityResponse$DeferredResultSubscriber` | The CHA path was an artifact (that subscriber is created only by WebMvc.fn publisher responses, which neither app uses). **But the method really runs:** `ResourceHttpRequestHandler#handleRequest` → `ServletWebRequest#checkNotModified(J)` → `validateIfNoneMatch` on every static-resource request. Neither graph reaches it, because `ApplicationFilterChain#internalDoFilter` → `javax.servlet.Servlet#service` is a `javax.*` receiver the skip rule never expands, so `DispatcherServlet` is unreachable in both modes | **false negative** under RTA (CHA's `FRAMEWORK` was right only by accident) |

So of 8 changes: 5 correct, 1 undetermined, 2 (one advisory, both apps) a false negative.
The false negative is not RTA's fixed point going wrong — RTA removed exactly the absurd
edge it should — but it shows that a narrower graph turns a gap that CHA's
over-approximation happened to paper over (the unexpanded servlet dispatch) into a
`NOT_REACHABLE`. `MODEL_INCOMPLETE` is attached, but its gap counts name reflection,
deserialization and the like, not the skipped servlet dispatch.

## Decision: the default stays CHA

RTA stays opt-in (`-callgraph-mode rta`); the default strategy is not changed:

1. **A real false negative** (CVE-2024-38809 on both apps) appears with `-roots profile`,
   from a pre-existing modeling gap (container → servlet dispatch through a skipped `javax.*`
   receiver) that RTA exposes. Until that dispatch is modeled, narrowing the default would
   trade a few correct `NOT_REACHABLE`s for a wrong one.
2. **With the default `-roots all`, RTA changes no verdict** on either app: the default user
   would pay the extra pass and gain nothing visible.
3. **The fallback is a heuristic:** a class created by reflection and named only in a
   configuration file is not live (it is the reason the four Tomcat removals above are
   right for these apps — and would be wrong for an app whose `server.xml` declares those
   valves).

Re-evaluate after container → servlet/filter dispatch is modeled (a FRAMEWORKS/JVM_GRAPH
item: a bounded exemption from the `javax.`/`jakarta.` skip for `Servlet#service` and
`Filter#doFilter` to the registered servlets and filters). Points-to (GRAPH_PRECISION P3)
is not justified by these numbers yet: on both apps the fallback, not the dispatch
precision, decides most of what RTA keeps.

## Using RTA (migration)

- Opt in with `-callgraph-mode rta`; nothing changes for a run without it. Use it as a
  second opinion with `-roots profile`, always next to a CHA run of the same input and
  configuration: `-graph-out` from the CHA run, `-graph-compare` in the RTA run, then read
  the "Call graph compared" section — every finding whose verdict or path changed, with the
  removed edges that explain it.
- What changes in the outputs: the report gets an "RTA" section; v2 gets `callgraph.mode =
  "rta"`, `callgraph.unknown_allocation_policy = "cha-fallback-with-gap"` and
  `callgraph.rta` (`docs/RESULTS_CONTRACT.MD`); reachable `Unsafe`/constructor-handle sites
  count as `allocation` model gaps of `NOT_REACHABLE` findings; snapshots record
  `alloc.catalog` and `callgraph.rta` among the algorithm versions (only in RTA mode), so a
  snapshot comparison flags a change of those rules.
- Treat an RTA-only `NOT_REACHABLE` as "no path in the narrowed model", never as proof:
  check whether the removed edge's class is created by configuration (XML, properties,
  `META-INF` files other than `services`) or reached through a framework dispatch the graph
  does not expand (servlet containers).
