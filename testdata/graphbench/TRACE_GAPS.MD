# Static-graph gaps against a real run (petclinic)

What the call graph misses, measured against a recorded run rather than guessed: every
caller → callee relation a JFR recording of petclinic measured (RUNTIME_NATIVE N1,
`-runtime-trace`) is checked against the static graph, and each one the graph has no edge for
gets a cause decided from the caller's bytecode at the measured index (causes `g1`,
`runtimeevents.GapCauseVersion`; report section "Runtime evidence", v2
`runtime.calls_not_in_graph_by_cause`).

**One run, one scenario:** pages, then static resources with `If-None-Match`
(`scripts/runtime-trace-petclinic.sh`, JDK 25, 10 ms stack sampling plus exact entries of the
mapped vulnerable methods). A relation the run did not take is not counted — this measures
**missed targets the JVM really used**, not over-approximation. Analyzer at `main` after the
service-provider dispatch model (#26); same trace for every row.

## Totals

| Mode | Roots | Measured relations | In the static graph | Not in the static graph |
|---|---|---|---|---|
| CHA | all | 2 867 | 2 646 | **221 (7.7%)** |
| CHA | profile | 2 867 | 2 646 | 221 (7.7%) |
| RTA | all | 2 867 | 2 645 | 222 (7.7%) |
| RTA | profile | 2 867 | 2 541 | **326 (11.4%)** |

The roots scenario does not change what CHA misses (roots decide reachability, not edges).
RTA with `-roots all` removes one relation the run took; with `-roots profile` it removes
**105** — measured evidence for P4's decision to keep CHA as the default.

## Causes (CHA, 221 relations)

| Cause | Calls | Share | Most frequent receiver types / sites | Owner (RECOMMENDATION) |
|---|---|---|---|---|
| `skipped-receiver` | 87 | 39% | `java.sql.Connection` 7, `javax.servlet.http.HttpServletRequest` 7, `HttpServletResponse` 7, `javax.servlet.FilterChain` 6, `java.sql.Statement` 4, `java.io.Reader` 3, `java.lang.Class` 3, `java.lang.Object` 3, `java.util.Map` 3 | JVM_GRAPH skip rule; per API, a model like servlet/SPI dispatch |
| `sam-call` | 58 | 26% | `java.util.function.Supplier` 8, `Consumer` 5, Spring `ReflectionUtils$MethodCallback` 5, `java.util.stream.Stream` 4, `Runnable` 3, `UnaryOperator` 3, Spring `ObjectFactory` 3 | DYNAMIC_LINKAGE L1 (SAM-call materialization) |
| `reflection` | 33 | 15% | `Method#invoke` 30, `Constructor#newInstance` 3 | REFLECTION R3 |
| `inherited-impl` | 17 | 8% | Spring `ConfigurableListableBeanFactory` 9, Thymeleaf `ISpringTemplateEngine` 2, Tomcat `Container`/`Server` 2, `BeanWrapper`, `ListableBeanFactory`, `WebApplicationContext`, Hibernate `XProperty` | JVM_GRAPH G2 (a CHA defect, below) |
| `proxy` | 12 | 5% | JPA `EntityManager` 5, `EntityManagerFactory`, `Query`, Spring Data repositories 3, annotations 2 | FRAMEWORKS F3 |
| `class-init` | 8 | 4% | `<clinit>` of Jackson, Log4j, Tomcat, H2, Hibernate Validator, Spring classes | JVM_GRAPH G3 |
| `async` | 3 | 1% | `ExecutorService#submit` 2, `ThreadPoolExecutor#execute` 1 | ASYNC_CALLBACKS A3 |
| `no-call-at-site` | 3 | 1% | no call instruction at the measured index (inlined frames / line tables) | — |

What each cause means (checked on examples, and on the fixture `testdata/runtime` for
`reflection`, `inherited-impl` and `sam-call`, `TestGapClassifier`/`TestRuntimeTrace`):

- **`skipped-receiver`** — a virtual/interface call on a `java.`/`javax.`/`jdk.`/`sun.` type
  the CHA skip never expands, running a class's implementation. The servlet objects are the
  container's request/response wrappers and the filter chain (`FilterChain#doFilter` is not
  covered by the servlet dispatch model, only `Filter#doFilter` and `Servlet#service`); the
  JDBC ones are Hikari's pool proxies (`HikariProxyConnection`), outside the driver jars the
  service-provider model is bounded to.
- **`sam-call`** — a functional-interface call running a lambda body or a method-reference
  target created in **another** method (L1 links a lambda only to the method creating it).
  The largest single gap a model can close, and spread over JDK and library interfaces.
- **`inherited-impl`** — **a defect of ariadne's CHA**, not a missing model: for a call on
  type `T`, CHA adds the subtypes of `T` that *declare* the method; when a subtype `S` of `T`
  inherits the body from a superclass `B` that is not itself a subtype of `T` (Spring's
  `DefaultListableBeanFactory` implements `ConfigurableListableBeanFactory` but inherits
  `getBean` from `AbstractBeanFactory`; Tomcat's `StandardHost` inherits `start` from
  `LifecycleBase`), the JVM runs `B`'s body and CHA has no edge to it. Fix: for each concrete
  subtype, add the body `ResolveExact` selects.
- **`rta-pruned`** (RTA only) — a CHA edge RTA removed although the JVM took it. 35 of the 105
  are Thymeleaf/AttoParser markup handlers: classes whose `new` sits in code the static graph
  does not reach (behind the gaps above) or that are created reflectively by name — RTA's
  narrowing compounds every other gap.

## What this says about the next steps

Ranked by measured payoff per effort:

1. **Fix CHA's inherited implementations (`inherited-impl`, 17).** A correctness fix in one
   place (`callgraph.CHA.DispatchTargets`), JVM-exact, no heuristic; every graph changes, so
   it needs the usual verdict comparison.
2. **SAM-call materialization (`sam-call`, 58).** DYNAMIC_LINKAGE L1's open half: follow a
   lambda/method reference from its creation site to the functional-interface calls that
   receive it (DATA_FLOW summaries already track values across calls).
3. **Reflection across methods (`reflection`, 33).** REFLECTION R3 — most are framework
   internals (Spring's `ReflectionUtils`, Hibernate accessors).
4. **Servlet API wrappers and `FilterChain` (`skipped-receiver`, ~20 of 87).** An extension
   of the servlet dispatch model; the rest of `skipped-receiver` is spread thin over many JDK
   types (collections, I/O), where lifting the skip would fan out (the reason it exists).
5. **JPA/Spring Data proxies (`proxy`, 12)** — FRAMEWORKS F3/F4.

Not supported by this measurement: RTA as a default (it adds 105 missed relations with
`-roots profile`), and points-to (GRAPH_PRECISION P3) — what the graph misses is dominated by
unmodeled mechanisms, not by imprecise dispatch.
