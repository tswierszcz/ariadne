# GRAPH_PRECISION P4 — call-graph strategy benchmark (CHA vs RTA)

**Rerun 2026-11-08 after the servlet container dispatch model** (`servlet-container:dispatch`,
`internal/servletdispatch`, FRAMEWORKS F2): the files here and the tables below are that
rerun; the P4 figures without the model are kept in "P4 baseline" for comparison, and the
section "Servlet container dispatch" explains what changed and why.

Measured 2026-11-08 with `scripts/graph-benchmark.sh --repeat 3` (real applications) and
`go test ./cmd/ariadne -run TestGraphBenchmarkFixtures -v` (fixtures with a real-JVM
oracle). **A measurement, not ground truth:** fewer edges or fewer `REACHABLE`/`FRAMEWORK`
results are not an improvement by themselves; every verdict change below is explained by
hand from the bytecode, not by an LLM's opinion of a shorter path.

| File | Content |
|---|---|
| `results.tsv` | one row per app × roots scenario × mode: wall time and peak RSS (min–max of 3 measured runs), graph size, CHA edges removed / kept only by the unknown-allocation fallback, live classes, active unknown-allocation gaps, created-type catalog size and gaps, model-gap sites (all / reachable), v1 verdict counts `R/F/NR/U/NA`, registered servlets/filters and edges added by B3 models, and against the CHA run of the same scenario the edges removed/added and the findings whose verdict or path changed |
| `verdict-changes.tsv` | every verdict change against CHA, with the removed edges that explain it (from `-graph-compare`) |
| `versions.txt` | analyzer commit, Go version, and per mode the graph settings and rule versions: builder `2`, skip rule (`java.`, `javax.`, `jdk.`, `sun.` receivers not expanded — the servlet dispatch model is the bounded exception), unknown-allocation policy, catalog `p1-2`, RTA `p2-1`, servlet dispatch `s1` |

## Setup

- Inputs: the prepared petclinic (Spring Boot 2.4.5, embedded Tomcat 9.0.45) and WebGoat
  (Spring Boot, Undertow) of `scripts/prepare-petclinic.sh` / `scripts/prepare-webgoat.sh`,
  `-profile spring`, offline on the same OSV cache, `-deps-list` given.
- Each configuration: 3 measured runs that write and read no graph file, plus 1 unmeasured
  run that writes (CHA) or compares against (RTA) the graph — reading a several-hundred-MB
  graph file costs ~1 GB of memory, which must not be charged to the strategy. Every figure
  other than time and memory was identical across the 4 runs (the script checks).
- `versions.txt` names a dirty commit (`a64edf7`, P4): the uncommitted changes at measurement
  time were the servlet dispatch model and its wiring, docs and tests. One later change (Spring
  servlet/filter beans count only in a Spring Boot application) does not apply to either app
  — neither declares such a bean; the per-finding comparison below was run with the final
  binary and gives the same counts.
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
| `testdata/servletdispatch` (a container that creates web.xml-declared servlets/filters by name and dispatches through the real `javax.servlet` API: an `HttpServlet` via its template dispatch, a `GenericServlet`, a filter; an unregistered servlet and filter on the class path) | profile / all | cha | 5 | 5 of 8 | 0 | 0 |
| | profile / all | rta | 5 | 5 of 8 | 0 | 0 |

The test fails if CHA misses anything, if RTA misses anything but the documented case, if
RTA reaches a body CHA does not, or if the graph reaches the unregistered `sd.UnusedServlet` /
`sd.UnusedFilter` (the 3 traced bodies not reached). Without the servlet dispatch model
(analyzer `a64edf7`) none of the 5 observed `servletdispatch` bodies is reached from `main`
in either mode.

## Real applications

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | Model edges | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 17.3–18.2 | 1320–1435 | 996 446 | — | — | 678 of 1096 | 4 | 5 127 | 6/72/12/57/6 |
| petclinic | profile | rta | 16.6–16.7 | 1350–1403 | 956 465 | 39 981 of 310 418 (35 374) | 17 106 (6 071) | 625 of 1097 | 4 | 5 127 | 6/68/16/57/6 |
| petclinic | all | cha | 14.1–16.5 | 1363–1397 | 996 446 | — | — | 678 of 1096 | 4 | 5 127 | 6/72/12/57/6 |
| petclinic | all | rta | 16.5–18.7 | 1380–1403 | 990 536 | 5 910 of 310 418 (59 454) | 21 835 (9 923) | 676 of 1097 | 4 | 5 127 | 6/72/12/57/6 |
| WebGoat | profile | cha | 28.0–32.6 | 1783–1841 | 1 411 339 | — | — | 742 of 1194 | 0 | 1 843 | 42/52/9/61/0 |
| WebGoat | profile | rta | 27.8–29.8 | 1774–1821 | 1 327 988 | 83 351 of 556 769 (158 682) | 25 211 (10 671) | 662 of 1201 | 0 | 1 843 | 42/50/11/61/0 |
| WebGoat | all | cha | 27.2–29.8 | 1735–1802 | 1 411 344 | — | — | 743 of 1194 | 0 | 1 848 | 42/52/9/61/0 |
| WebGoat | all | rta | 27.8–32.8 | 1756–1795 | 1 396 558 | 14 786 of 556 769 (101 407) | 32 210 (13 181) | 745 of 1201 | 0 | 1 848 | 42/52/9/61/0 |

<details><summary>Before the service-provider dispatch model (servlet dispatch only, analyzer <code>9a460a6</code>)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 15.2–15.8 | 1381–1411 | 991 340 | — | — | 606 of 1096 | 4 | 3/74/13/57/6 |
| petclinic | profile | rta | 15.2–15.4 | 1351–1419 | 941 815 | 49 525 of 310 418 (48 701) | 16 011 (7 379) | 541 of 1097 | 4 | 3/70/17/57/6 |
| petclinic | all | cha | 15.2–15.3 | 1321–1367 | 991 340 | — | — | 606 of 1096 | 4 | 3/74/13/57/6 |
| petclinic | all | rta | 14.1–15.4 | 1373–1415 | 985 273 | 6 067 of 310 418 (71 771) | 21 662 (11 168) | 604 of 1097 | 4 | 3/74/13/57/6 |
| WebGoat | profile | cha | 23.0–24.9 | 1760–1784 | 1 409 574 | — | — | 676 of 1194 | 0 | 42/51/10/61/0 |
| WebGoat | profile | rta | 23.8–25.8 | 1766–1798 | 1 314 196 | 95 378 of 556 769 (178 188) | 24 398 (11 752) | 590 of 1201 | 0 | 42/49/12/61/0 |
| WebGoat | all | cha | 23.1–23.8 | 1766–1797 | 1 409 575 | — | — | 677 of 1194 | 0 | 42/51/10/61/0 |
| WebGoat | all | rta | 24.0–25.1 | 1772–1817 | 1 394 752 | 14 823 of 556 769 (122 892) | 32 123 (13 998) | 680 of 1201 | 0 | 42/51/10/61/0 |

</details>

<details><summary>P4 baseline (analyzer <code>a64edf7</code>, without the servlet dispatch model)</summary>

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

</details>

- **Cost:** RTA adds about 0–2 s; peak memory is within run-to-run noise. The machine was
  busier during this run (load average ~5) than for the tables below, so absolute times are
  not comparable across tables; measured back to back on WebGoat, the service-provider
  dispatch model's own cost is 0.6–2 s, inside the main analyzer's own spread (23.5–27.2 s).
- **Edges:** `-roots profile` removes 13% (petclinic) / 15% (WebGoat) of the CHA edges; `-roots
  all` 2–3%. No CHA edge is ever added by RTA. Under RTA, 1 (petclinic) / 7 (WebGoat) more gap
  sites exist — the `Unsafe`/constructor-handle `allocation` gaps RTA introduces.
- **Unknown sites:** the fallback carries a large share of what RTA keeps (in `-roots
  profile`, 35 374 of the 270 437 CHA edges kept on petclinic, 158 682 of 473 418 on
  WebGoat): most of the narrowing RTA could do is given back to unresolved reflection.
- **Verdicts:** `-roots all` (the default roots) changes no verdict in either app. `-roots
  profile` moves 4 (petclinic) and 2 (WebGoat) findings from `FRAMEWORK` to
  `NOT_REACHABLE`, all carrying `MODEL_INCOMPLETE` — the same findings as before the
  service-provider dispatch model. No `REACHABLE` result changes anywhere.

## Servlet container dispatch (model `servlet-container:dispatch`, `s1`)

**Registered servlets/filters.** petclinic (a Spring Boot application, `@SpringBootApplication`
on `PetClinicApplication`): `DispatcherServlet` (DispatcherServletAutoConfiguration),
`OrderedCharacterEncodingFilter` (HttpEncodingAutoConfiguration), `OrderedFormContentFilter`
and `OrderedRequestContextFilter` (WebMvcAutoConfiguration); no `web.xml`, no
`@WebServlet`/`@WebFilter`, no servlet/filter bean, Spring Security absent (so no
`DelegatingFilterProxy`). WebGoat: **none** — the prepared input is `webgoat-container` plus
the vulnerable-components lesson (`scripts/prepare-webgoat.sh`); its launcher
`StartWebGoat` (`@SpringBootApplication`, `@ServletComponentScan`) lives in `webgoat-server`,
which is not part of the analyzed input, so nothing in the input enables Boot
auto-configuration and the model registers nothing (WebGoat's graph and verdicts are
identical to P4). This is the model working as specified, not a measurement of WebGoat's real
deployment.

**Edges (petclinic, 12 added by the model, identical in all four runs):**
`ApplicationFilterChain#internalDoFilter` → `HttpServlet#service(ServletRequest, …)` (the
body JVM selects for `DispatcherServlet`) and → `OncePerRequestFilter#doFilter` (the body
of the three filters); `HttpServlet#service(…)`/`#doHead` → `FrameworkServlet#service(Http…)`,
`#doGet/#doPost/#doPut/#doDelete/#doOptions/#doTrace`. Two more `Servlet#service` sites in
Spring MVC — `SimpleServletHandlerAdapter#handle` and
`ServletWrappingController#handleRequestInternal` — also get the registered set: the model
expands every reachable `Servlet#service` call to the registered servlets, so these are a
bounded over-approximation (they dispatch to a servlet wrapped as a handler, which petclinic
does not define). `DispatcherServlet#doDispatch` and `ResourceHttpRequestHandler#handleRequest`
are now reachable from `main` in both modes (BFS over the exported graphs; not reachable with
`a64edf7`).

**Verdict changes against P4 (same input, same mode; v1 per finding, `a64edf7` vs this
analyzer):** WebGoat none. petclinic — every change is `NOT_REACHABLE` → `FRAMEWORK`:

| Advisory | Vulnerable code | New path (from the servlet dispatch on) | Scenarios | Check |
|---|---|---|---|---|
| CVE-2024-38819 | `ResourceHttpRequestHandler#isInvalidPath` (mapped from the fix diff) | `internalDoFilter` → `HttpServlet#service` → `FrameworkServlet#service` → `DispatcherServlet#doDispatch` → `HttpRequestHandlerAdapter#handle` → `ResourceHttpRequestHandler#handleRequest` → `#getResource` → `#isInvalidPath` | all four | The method runs on every static-resource request (petclinic serves `/resources/**` and webjars through `ResourceHttpRequestHandler`). **Correct** for the mapped method; whether the advisory's functional-endpoint condition applies is outside reachability |
| CVE-2026-41844 | `DefaultRequestToViewNameTranslator#getViewName` | `… DispatcherServlet#doDispatch` → `#applyDefaultViewName` → `#getDefaultViewName` → `#getViewName` | all four | `applyDefaultViewName` runs after every handler; `getDefaultViewName` only when the handler set no view. A real path, **plausible** (not exercised here) |
| CVE-2023-20883 | `WelcomePageHandlerMapping#getHandlerInternal` and its welcome-page lookup | CHA: `internalDoFilter` → `OncePerRequestFilter#doFilter` → CHA on `doFilterInternal` to `CorsFilter` → `HandlerMappingIntrospector` → `AbstractHandlerMapping#getHandler` → `WelcomePageHandlerMapping#getHandlerInternal`; RTA: `… DispatcherServlet#doDispatch` → `#getHandler` → `AbstractHandlerMapping#getHandler` → `WelcomePageHandlerMapping#getHandlerInternal` | all four | Boot registers `WelcomePageHandlerMapping` and `DispatcherServlet` consults it for each request it has not matched earlier. **Correct** (RTA's path is the real one; CHA's shortest path goes through `CorsFilter`, which petclinic does not register — a CHA artifact on an otherwise reachable method) |
| CVE-2024-22259 | `UriComponentsBuilder#fromOriginHeader/#fromHttpUrl` | RTA: `… DispatcherServlet#doDispatch` → `HandlerExecutionChain#applyPreHandle` → `AbstractHandlerMapping$CorsInterceptor#preHandle` → `DefaultCorsProcessor#processRequest` → `CorsUtils#isCorsRequest` → `fromOriginHeader` | profile / rta only (CHA already `FRAMEWORK` through the WebFlux `CorsWebFilter` artifact) | P4's "undetermined": the servlet path is now visible. Spring MVC runs the CORS processor for a request with an `Origin` header when a CORS configuration applies or for a pre-flight request; the path is real, **plausible** for an attacker-sent pre-flight (not exercised here) |
| CVE-2024-38809 | `ServletWebRequest#validateIfNoneMatch` | profile / rta: reported path is still the shortest one, through the reactive `WriteBarrier` → `DeferredResultSubscriber#onNext` (that subscriber is now live: WebMvc.fn code reachable from `DispatcherServlet` creates it); the real path `… DispatcherServlet#doDispatch` → `HttpRequestHandlerAdapter#handle` → `ResourceHttpRequestHandler#handleRequest` → `ServletWebRequest#checkNotModified(J)` → `#checkNotModified(String, J)` → `#validateIfNoneMatch` is in the graph in both modes | profile / rta only (CHA already `FRAMEWORK`) | P4's **false negative is fixed**: `FRAMEWORK` in every petclinic scenario, and the method is reachable through the request path that really runs it. The report shows the shortest path, which is not the real one — read it as "reachable", not as the route |

So the model adds 3 `FRAMEWORK` results to every petclinic scenario and 2 more under `-roots
profile` with RTA, all through the request path; none is a known false positive, and no
`REACHABLE` result changes (the `FRAMEWORK`/`REACHABLE` split is unchanged: the new paths
run entirely through framework code). The petclinic RTA-vs-CHA changes drop from 6 to 4 (the
four Tomcat valve/realm findings, still correct).

## Service-provider dispatch (model `service-provider:dispatch`, `p1`)

Found by a runtime trace (RUNTIME_NATIVE N1): on petclinic the JVM ran
`HibernateValidator#buildValidatorFactory` (called by Hibernate's
`AbstractConfigurationImpl#buildValidatorFactory` through
`javax.validation.spi.ValidationProvider`, a skipped receiver) while the graph stopped there.
The model expands calls on a skipped interface that is a registered `META-INF/services` service
(to its registered providers) or a product of one (an interface of the same API family reached
through the service's return types; to implementations in the providers' jars) — see the README.

**Catalog.** petclinic: 8 skipped service interfaces with registered providers, 1 096 call
signatures, 5 127 model edges in total (servlet dispatch included); WebGoat: 11 services,
1 105 signatures, 1 843–1 848 model edges. Families on petclinic include Bean Validation
(`javax.validation`, Hibernate Validator), JDBC (`java.sql`, H2 and MySQL drivers) and JCache
(`javax.cache`, Ehcache). A first version that took every interface of the family (not only
products) also sent `javax.servlet` calls to spring-web's servlet classes — a family brought
in by `ServletContainerInitializer`, which hands nothing out — and turned CVE-2021-22118 into a
false `FRAMEWORK`; the return-type closure removed it.

**Verdict changes against the analyzer without the model** (same input, `-roots all`, CHA,
`-graph-compare`; identical under `-roots profile` and RTA), each checked against a JFR
recording of petclinic with exact entries (`jdk.MethodTrace`) of the methods involved:

| App | Advisory | Change | New edge on the path | Runtime check | Verdict |
|---|---|---|---|---|---|
| petclinic, WebGoat | CVE-2025-35036 (Hibernate Validator) | `NOT_REACHABLE` → `FRAMEWORK` | `MessageInterpolatorFactory#getObject` → `AbstractConfigurationImpl#getDefaultMessageInterpolator` (a `javax.validation.Configuration` product) | `HibernateValidator#buildValidatorFactory` entered 3 times at startup; the advisory's mapped methods entered (26 exact entries in the vulnerable-method recording) | **correct**: a static false negative removed |
| petclinic | CVE-2021-46877, CVE-2022-45868, CVE-2023-6378 | `FRAMEWORK` → `REACHABLE` | `CacheConfiguration#lambda$petclinicCacheConfigurationCustomizer$0` (application code) → `Eh107CacheManager#createCache` (a `javax.cache.CacheManager` product of Ehcache's `CachingProvider`) | `Eh107CacheManager#createCache` entered (the first hop is real); the rest of the path — Ehcache's disk store `FileBasedStateRepository#loadMaps` → deserialization → `NodeSerialization#readExternal` — is pre-existing CHA/deserialization over-approximation: 0 entries | first hop **correct**; the `REACHABLE` result is as plausible as the downstream path, i.e. not (petclinic's caches are heap-only) |

No `REACHABLE` result is lost; RTA-versus-CHA changes are unchanged (the 4 + 2 above).

## Every verdict change, checked by hand

`FRAMEWORK` → `NOT_REACHABLE` under RTA, `-roots profile` (`verdict-changes.tsv`). With the
servlet dispatch model the petclinic rows for CVE-2024-22259 and CVE-2024-38809 are gone
(both `FRAMEWORK` in both modes, see above); the WebGoat CVE-2024-38809 row stays, because
nothing in WebGoat's analyzed input registers `DispatcherServlet`. The table below keeps the
P4 analysis of all 8 for the record:

| App | Advisory | Vulnerable code | Why CHA reached it | Check | Verdict |
|---|---|---|---|---|---|
| petclinic | CVE-2022-45143 | `JsonErrorReportValve#report` | `ErrorReportValve#invoke` → CHA to the subclass | No class of the input references `JsonErrorReportValve` (bytecode scan of every jar, slashed and dotted names): Tomcat creates valves only from `server.xml`/`context.xml`, which embedded Spring Boot does not use | **correct** |
| petclinic | CVE-2025-55752, CVE-2026-53404 | `RewriteValve#invoke/#parse`, `RewriteRule#evaluate` | `LifecycleBase#start` → CHA to `RewriteValve#startInternal` | Same: `RewriteValve` is referenced by no other class; configuration-only | **correct** |
| petclinic | CVE-2026-55957 | `JNDIRealm#bindAsUser` | `BasicAuthenticator#doAuthenticate` → CHA to `JNDIRealm#authenticate` | Only `MBeanFactory#createJNDIRealm` (a JMX management operation, unreachable) does `new JNDIRealm`; realms otherwise come from `server.xml` | **correct** |
| WebGoat | CVE-2021-22118 | reactive `DefaultPartHttpMessageReader` | WebFlux `DefaultServerWebExchange#initFormData` → CHA into the reactive multipart readers | The reactive stack is present in `spring-web` but WebGoat is a servlet MVC application: no reachable code creates a WebFlux exchange or these readers | **correct** |
| petclinic | CVE-2024-22259 | `UriComponentsBuilder#fromHttpUrl/#fromOriginHeader` | WebFlux `DefaultWebFilterChain#invokeFilter` → CHA to `CorsWebFilter#filter` | The CHA path was an artifact (WebFlux, unused). Whether a servlet path exists cannot be told from either graph: servlet request dispatch is cut in both modes (see below) | artifact removed; real reachability **undetermined** |
| petclinic, WebGoat | CVE-2024-38809 | `ServletWebRequest#validateIfNoneMatch` (ETag parsing) | reactive `ChannelSendOperator$WriteBarrier#emitCachedSignals` → CHA on `Subscriber#onNext` to the servlet-functional `PublisherEntityResponse$DeferredResultSubscriber` | The CHA path was an artifact (that subscriber is created only by WebMvc.fn publisher responses, which neither app uses). **But the method really runs:** `ResourceHttpRequestHandler#handleRequest` → `ServletWebRequest#checkNotModified(J)` → `validateIfNoneMatch` on every static-resource request. Neither graph reaches it, because `ApplicationFilterChain#internalDoFilter` → `javax.servlet.Servlet#service` is a `javax.*` receiver the skip rule never expands, so `DispatcherServlet` is unreachable in both modes | **false negative** under RTA (CHA's `FRAMEWORK` was right only by accident) |

So of P4's 8 changes: 5 correct, 1 undetermined, 2 (one advisory, both apps) a false
negative. After the servlet dispatch model, 6 remain: the 5 correct ones and WebGoat's
CVE-2024-38809, still a false negative of this *input* (its launcher is not analyzed;
`-roots profile` has no root that reaches Undertow's dispatch registered to anything).
The false negative is not RTA's fixed point going wrong — RTA removed exactly the absurd
edge it should — but it shows that a narrower graph turns a gap that CHA's
over-approximation happened to paper over (the unexpanded servlet dispatch) into a
`NOT_REACHABLE`. `MODEL_INCOMPLETE` is attached, but its gap counts name reflection,
deserialization and the like, not the skipped servlet dispatch.

## Decision: the default stays CHA

RTA stays opt-in (`-callgraph-mode rta`); the default strategy is not changed. Re-evaluated
after the servlet dispatch model:

1. **The petclinic false negative is gone** (CVE-2024-38809 is `FRAMEWORK` in every
   scenario, through the real request path), and P4's undetermined CVE-2024-22259 now has a
   visible servlet path. What remains of RTA's verdict changes on petclinic are the four
   correct Tomcat removals.
2. **But a false negative of the same kind remains wherever the input does not show the
   registration** — WebGoat's analyzed input lacks its Boot launcher, and the model's own
   gaps (servlets/filters registered by a `ServletContainerInitializer`, Actuator, Spring
   Security's inner chain, `ServletContext#addServlet`, `FilterRegistrationBean` with an
   unknown class) are not all reported as gaps of a finding. RTA turns such a missing
   dispatch into `NOT_REACHABLE`; CHA's over-approximation often still papers over it.
3. **With the default `-roots all`, RTA still changes no verdict** on either app: the default
   user would pay the extra pass and gain nothing visible.
4. **The fallback is a heuristic:** a class created by reflection and named only in a
   configuration file is not live (it is the reason the four Tomcat removals are right for
   these apps — and would be wrong for an app whose `server.xml` declares those valves).

Points-to (GRAPH_PRECISION P3) is still not justified by these numbers: on both apps the
fallback, not the dispatch precision, decides most of what RTA keeps.

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
  does not expand — for servlet containers, check the `servlet dispatch` log line: a servlet
  or filter missing from it (registered outside the analyzed input, by a container
  initializer or programmatically) is dispatched to by nothing in the graph.
