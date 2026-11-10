# Static-graph gaps against a real run (petclinic, JPetStore)

What the call graph misses, measured against a recorded run rather than guessed: every
caller → callee relation a JFR recording of petclinic measured (RUNTIME_NATIVE N1,
`-runtime-trace`) is checked against the static graph, and each one the graph has no edge for
gets a cause decided from the caller's bytecode at the measured index (causes `g1`; `g2`
from 2026-11-09, see "A sharper `sam-call`" below;
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

## After fixing CHA's inherited implementations (builder `3`)

Step 1 below, done the same day: CHA now also targets, for each concrete subtype, the body JVM
method selection finds for it when it is declared above the subtype in a class outside the
call's type (`callgraph.CHA.DispatchTargets`, `TestCHAInheritedBody`). Same trace:

| Mode | Roots | Not in the static graph | Change |
|---|---|---|---|
| CHA | profile | 203 (7.1%) | −18: `inherited-impl` 17 → **0**, `no-call-at-site` 3 → 2 |
| RTA | profile | 234 (8.2%) | −92: `rta-pruned` 105 → **31** |

The RTA drop is the compounding effect described above: code that is now reachable creates
classes RTA used to consider dead. Verdicts: none changed on petclinic or WebGoat in either
roots mode (CHA, `-graph-compare` against the builder `2` graph); 43–44 petclinic and 1–3
WebGoat findings got a shorter path through a body the JVM really runs (e.g.
`OwnerController#processFindForm` → `AbstractErrors#rejectValue` behind `BindingResult`);
+11 713 / +16 439 edges.

## After SAM-call materialization (builder `4`)

Step 2 below, 2026-11-09 (DYNAMIC_LINKAGE L1, `internal/samcall`, rules `l1-1`): a lambda or
method reference created in one method and run through its functional interface in another
gets a `sam` edge from the calling site; the value is followed with DATA_FLOW (callers,
captured values, field writers, helper returns), a value handed to a JDK method is linked from
that call, and an edge needs a reachable creator of the value (README "SAM calls").

**A new recording.** The trace above was not kept, so petclinic was recorded again with the
same script and scenario (`scripts/runtime-trace-petclinic.sh`, JDK 25.0.5): 2 702 measured
relations. Before/after are both on this new trace — the builder `3` binary (`651b3b5`) and
builder `4` — CHA, `-roots profile`:

| Cause | Builder `3` | Builder `4` |
|---|---|---|
| `skipped-receiver` | 93 | 92 |
| **`sam-call`** | **50** | **22** |
| `reflection` | 37 | 37 |
| `proxy` | 9 | 9 |
| `class-init` | 8 | 8 |
| `async`, `no-call-at-site`, `indirect` | 3, 3, 1 | 3, 3, 1 |
| **Not in the static graph** | **204 of 2 702 (7.5%)** | **175 (6.5%)** |

(The first trace had 58 `sam-call` of 203; sampling makes each recording differ.) The
`skipped-receiver` relation that went away is `MergedAnnotationCollectors#lambda$toMultiValueMap$4`
→ `MultiValueMapAdapter#add`, now in the graph.

What the 22 left are, checked one by one (the measured callee, and the reason the resolver recorded
for the site):

- **12 are not lambdas at all.** The classifier counts as `sam-call` any virtual call whose
  measured callee has another name than the called method; through JDK frames that also
  catches a collection calling `equals`/`hashCode` (`Map` 3, `Set` 2), a `Stream` pulling from
  a Spring `Spliterator` (3), `List#sort` running a `Comparator` object, `ClassLoader` delegation,
  a Tomcat `FutureTask` (`Runnable` → `StartChild#call`) and a JDK stream sink in front of a
  lambda (`Consumer`). A sharper classifier would move them to `skipped-receiver`.
- **7 are real, with the value lost on the way** (each counted as a gap in the run's log):
  `ServletContextInitializer` ×2 (an array element, `TomcatStarter`), `Supplier` ×2 (a local
  live across an exception handler, `SpringIterableConfigurationPropertySource$Mappings`),
  `BiPredicate` (the result of an instance call, `getAncestorOfCheck()`), Spring Boot's
  `JCacheManagerCustomizer` (itself the argument of a `Consumer` run over an `ObjectProvider`
  stream), ByteBuddy's `TypeCache#findOrInsert` (`Callable`; its class extends the JDK's
  `ReferenceQueue`, which might declare a method of that signature, so its callers are not
  closed).
- **3 are left out by the reachability gate:** the values are resolved, but no method creating
  them is reachable from the profile roots — Spring Data's `RepositoryMethodInvoker`
  (`Invokable` → `AbstractJpaQuery#execute`: the creator runs behind the repository proxy), the
  actuator's `DiscoveredOperationsFactory` (`MethodIntrospector$MetadataLookup`) and
  `SpringConfigurationPropertySources#adapt` (`Function`). Without the gate they are edges, but
  so are spurious ones: a first version without it changed CVE-2021-23463 (H2 `JdbcSQLXML`)
  from `NOT_REACHABLE` to `FRAMEWORK` through `JdbcTemplate#execute(ConnectionCallback)` "running"
  `AbstractJdbcInsert`'s lambda, whose creating method nothing reachable calls (a
  context-insensitive union: the helper's parameter got every lambda any caller passes).

### A sharper `sam-call` (causes `g2`)

`g1` counted as `sam-call` any virtual call whose measured callee had another name than the
called method — which, through JDK frames, also caught the 12 non-lambdas above. Since `g2`
(2026-11-09) a call is `sam-call` only when the callee implements a lambda or method reference
the input creates (`invokedynamic` with a functional interface; an override of a
method-reference target included) whose interface is the call's type, or a subtype of it, and
whose method is the called one. Same trace, builder `4`:

| Cause | `g1` | `g2` |
|---|---|---|
| `sam-call` | 22 | **12** |
| `skipped-receiver` | 92 | **102** |
| Not in the static graph | 175 | 175 |

The 10 that moved are the JDK collection, spliterator, `Comparator`, `ClassLoader` and
`FutureTask` relations listed above (one, a stream sink in front of a lambda of another
interface, moved because the lambda's interface is not the call's). The 12 left are the 7 real
ones with the value lost, the 3 left out by the reachability gate, and 2 whose callee is also
the target of a method reference the input creates for that interface
(`EndpointIdTimeToLivePropertyFunction#apply` through `Function`,
`SpringConfigurationPropertySources#iterator` through `Iterable`) — counted as `sam-call`,
though the receiver may be an ordinary object of that class (the classifier cannot tell).
`TestGapClassifier` covers a callee no lambda implements at a SAM call site.

## After the servlet request/response objects (model `s2`)

Step 4 below, 2026-11-09 (`servlet-container:dispatch` `s2`, `servletdispatch.ObjectEdges`): a
virtual/interface call on `javax.servlet` `ServletRequest`/`ServletResponse`, their `Http*` and
`*Wrapper` types or `FilterChain` gets a `model` edge to the body each concrete implementation
in the input runs, once a method creating that implementation is reachable from the roots
(the SAM calls' creator gate); Thymeleaf's `ResourceUrlEncodingFilter` joins the Boot catalog.
Same trace, builder `4`, causes `g2`, CHA `-roots profile`:

| Cause | `s1` | `s2` |
|---|---|---|
| **`skipped-receiver`** | **102** | **69** |
| `no-call-at-site` | 3 | 0 |
| `reflection`, `sam-call`, `proxy`, `class-init`, `async`, `indirect` | 37, 12, 9, 8, 3, 1 | unchanged |
| **Not in the static graph** | **175 of 2 702 (6.5%)** | **139 (5.1%)** |

Of the 35 servlet relations under `skipped-receiver`, 33 are edges now: the filters' `chain.doFilter`
into Tomcat's `ApplicationFilterChain`, Spring/Thymeleaf code into `RequestFacade`/`ResponseFacade`
and the servlet wrappers, and `ResourceUrlEncodingFilter` with its wrappers. Left: Tomcat's
WebSocket `WsFilter` (registered by a `ServletContainerInitializer`) and that initializer's own
`onStartup` — not modeled. The 3 `no-call-at-site` relations (`UrlPathHelper#getContextPath`,
`ResourceHttpRequestHandler#getMediaType`, measured at an index with no call instruction) are
request accessors the graph now has as edges from the same caller.

The gate leaves out the wrappers only an unreachable method creates (Tomcat's
`CsrfPreventionFilter`/`ExpiresFilter`/`RemoteIpFilter`), but not those of filters that are
reachable only through CHA (Spring's `ForwardedHeaderFilter`, `ShallowEtagHeaderFilter`: their
`doFilterInternal` is a CHA target of `OncePerRequestFilter#doFilter`).

## After the JDBC objects (model `j1`)

2026-11-09 (`jdbc:objects` `j1`, `internal/jdbcdispatch`): a virtual/interface call on a
`java.sql` connection, statement, result set, metadata, LOB, `Savepoint`, `Wrapper` or
`javax.sql.DataSource` gets a `model` edge to the body each concrete implementation in the input
runs, once a method creating that implementation is reachable from the roots — the servlet
objects' creator gate (`callgraph.ObjectEdges`, now shared by both models; README "JDBC objects").

**A new recording.** The `s2` trace was not kept (the machine was restarted), so petclinic was
recorded again with the same script and scenario (`scripts/runtime-trace-petclinic.sh`, JDK
25.0.5): 2 946 measured relations. Before/after are both on this new trace — `main` at `fac65dc`
(`s2`, `u1`) and `j1`, CHA, `-roots profile`, causes `g2`:

| Cause | `main` | `j1` |
|---|---|---|
| **`skipped-receiver`** | **85** | **62** |
| `reflection`, `sam-call`, `proxy`, `class-init`, `async`, `no-call-at-site` | 43, 15, 12, 7, 1, 1 | unchanged |
| **Not in the static graph** | **164 of 2 946 (5.6%)** | **141 (4.8%)** |

The 23 that are edges now: all 22 relations on a `java.sql`/`javax.sql` receiver — Hibernate
and Spring JDBC into Hikari's proxies (`HikariProxyConnection#commit`/`setAutoCommit`/
`setReadOnly`, `ProxyConnection#close`, `HikariProxyPreparedStatement#executeQuery`/
`getQueryTimeout`, `HikariProxyResultSet#getInt`/`getDate`, `HikariProxyStatement#execute`,
`HikariProxyDatabaseMetaData#storesLowerCaseIdentifiers`, `HikariDataSource#getConnection`),
Hikari's proxies into H2 (`JdbcConnection`, `JdbcStatement`, `JdbcPreparedStatement`,
`JdbcResultSet`, `JdbcDatabaseMetaData`), Hikari's pool into its `DriverDataSource` — and the
`java.sql.Driver` relation behind them (`DriverDataSource#getConnection` →
`org.h2.Driver#connect`): the service-provider model already expands that call, but its caller
was not reachable before. No new relation is missing. No `java.sql`/`javax.sql` relation is left.

## What the `proxy` relations are (FRAMEWORKS F3, measured, not built)

2026-11-09, before building JDK-proxy dispatch (a call on interface `I` → `H#invoke` when a
reachable `Proxy.newProxyInstance` site has handler `H` and `I` in its interface set): the 12
`proxy` relations of the `j1` recording above (CHA, `-roots profile`), each traced to the
`newProxyInstance` site that creates its proxy, and that site's two arguments read with
DATA_FLOW's local values (`dataflow.Analyze`, `ArgumentAt`; a throwaway probe, not committed).

| Handler the JVM ran | Relations | Creation site | Handler argument | Interface array |
|---|---|---|---|---|
| Spring AOP `JdkDynamicAopProxy` (Spring Data repositories, from `OwnerController`/`VetController`/`VisitController`) | 5 | `JdkDynamicAopProxy#getProxy(ClassLoader)` | `this` (exact class) | field `proxiedInterfaces`, filled from the `AdvisedSupport` configuration |
| `SharedEntityManagerCreator$SharedEntityManagerInvocationHandler` (Spring Data JPA query creation) | 2 | `SharedEntityManagerCreator#createSharedEntityManager` | `new` | a `Class[]` built from the `interfaces` parameter plus `EntityManagerProxy` (`System.arraycopy`) |
| `ExtendedEntityManagerCreator$ExtendedEntityManagerInvocationHandler` | 2 | `ExtendedEntityManagerCreator#createProxy` | `new` | the result of a call (`ClassUtils.toClassArray` over a set of the raw EntityManager's interfaces) |
| `AbstractEntityManagerFactoryBean$ManagedEntityManagerFactoryInvocationHandler` | 2 | `AbstractEntityManagerFactoryBean#createEntityManagerFactoryProxy` | `new` | `ClassUtils.toClassArray(Collection)` over interfaces collected from the native factory's class |
| `SharedEntityManagerCreator$DeferredQueryInvocationHandler` | 1 | `SharedEntityManagerInvocationHandler#invoke` | `new` | the result of a call (`ClassUtils.getAllInterfacesForClass` of the real query) |

The handler's class is known at all 6 creation sites; **the interface set is bounded at none of
them** — each comes from configuration, a parameter or the class of an object created at run
time. Over the whole input: petclinic 59 `newProxyInstance` sites, 24 with a literal interface
array (21 also with a `new` handler); WebGoat 52 / 21 (16) — annotation proxies, Spring JDBC's
close-suppressing and transaction-aware connection proxies, Hibernate's `Session` proxy, MySQL
Connector/J wrappers: none on a measured relation. So the planned rule would add **0 of the 12**;
the stop rule set before the measurement applies (as for REFLECTION R3), nothing was built.

A model that could close them would not infer the interface set but take it from a catalog: the
proxies `SharedEntityManagerCreator`/`ExtendedEntityManagerCreator` create implement
`javax.persistence.EntityManager` (and `Query`/`EntityManagerFactory` for the other two handlers),
and Spring Data wraps each repository bean (FRAMEWORKS F2's `RepositoryBeans`) in a
`JdkDynamicAopProxy`. Even then the edge stops at the handler's `invoke`: from there the JVM
calls `Method#invoke` on the target (on the builder `4` recording, 6 of the 37 `reflection` relations
were exactly these handlers passing their `Method` on), and Spring Data's repository path already reaches
`SimpleJpaRepository` through the profile's synthetic edges. Expected effect on reachability:
small; not started.

**Built afterwards as a catalog (2026-11-10, `framework:proxies` `x1`):** the interface sets are
taken from what the frameworks document — MyBatis mappers, `SqlSessionTemplate`'s `SqlSession`,
Spring Data repositories, Spring ORM's `EntityManager`/`Query`/`EntityManagerFactory` — and each
edge waits for the handler's creator. On a new petclinic recording all 13 `proxy` relations are
graph edges (relations not in the graph 148 → 127); on JPetStore 6 of 8 (the 2 left are
`SqlSessionTemplate` → `SqlSessionInterceptor`: Spring creates the template only reflectively, so
the gate keeps its interceptor out). README, "Framework proxies".

**Spring AOP's interceptor chain (measured 2026-11-10, not built).** Behind
`JdkDynamicAopProxy#invoke` CHA expands `MethodInterceptor#invoke` and `TargetSource#getTarget` to
every implementation (petclinic: 60 interceptors, 25 of them never created with `new` anywhere in
the input — `EventPublicationInterceptor`, the remoting/EJB/JMX interceptors —, and 17 target
sources). Before building a model that keeps only configured interceptors, an upper bound: in the
`-roots all` CHA graphs, **every** CHA edge out of `ReflectiveMethodInvocation#proceed`,
`JdkDynamicAopProxy#invoke` and `CglibAopProxy` removed (petclinic 274, WebGoat 281, JPetStore 244 —
real ones included, more than any model would remove), then each positive finding's mapped
methods searched from the application roots: **no finding loses its last path** on any of the
three. The over-approximation only shortens paths (the 63/23 changed paths above); the stop rule
applies, nothing built. A model would also need further gates: `LazyInitTargetSource`, on the
observed path, is created with `new` by `MBeanExporter` and `LazyInitTargetSourceCreator`, which
CHA reaches too.

## A second application: JPetStore 6.0.2 (WAR)

2026-11-09: the same measurement on MyBatis JPetStore 6.0.2 (`scripts/prepare-jpetstore.sh`), a
WAR with `web.xml`-registered Stripes `StripesFilter`/`DispatcherServlet` and Spring's
`ContextLoaderListener`, Spring 4.3 XML configuration, MyBatis 3.4 over an embedded HSQLDB.
Recorded with `scripts/runtime-trace-jpetstore.sh`: the WAR in Jetty 9.4 (`jetty-runner`; the
container is not part of the analyzed input), JDK 25.0.5 with `--add-opens
java.base/java.lang=ALL-UNNAMED` (Spring 4.3's CGLIB), 50 × catalog, categories, product, item,
search, cart, sign-on with the application's seeded demo account, new order form, sign-off.
Analyzer `j1`, CHA, `-roots profile`, causes `g2`.

**722 measured relations, 37 (5.1%) not in the static graph:**

| Cause | Calls | What they are |
|---|---|---|
| `skipped-receiver` | 17 | `RequestDispatcher#forward` into JSP tag handlers 9 (the JSPs are compiled by the container, not in the input); `FilterChain#doFilter` → Stripes `DispatcherServlet#service` 1 and `HttpServletRequest` → `StripesRequestWrapper` 1 (the chain and request classes are the container's: `s2`'s creator gate has no creator in a WAR); `DataSource#getConnection` → Spring's `SimpleDriverDataSource` 1 and `Driver#connect` → HSQLDB's `JDBCDriver` 1 (below); `Map`, `Object#hashCode`, `FutureTask`, `DocumentBuilder` 1 each |
| `reflection` | 11 | Stripes calling the action beans' event handlers through `Method#invoke` 4 (`CatalogActionBean#viewCategory`, `#viewProduct`, `#searchProducts`, `CartActionBean#addItemToCart`); MyBatis-Spring's `SqlSessionInterceptor` → `DefaultSqlSession` 2; Spring XML property injection 1; `Constructor#newInstance` from bean definitions and logging/HSQLDB factories 4 |
| `proxy` | 7 | MyBatis mapper proxies (`MapperProxy#invoke` from `CatalogService` 5) and `SqlSessionTemplate`'s `SqlSession` proxy 2 |
| `class-init` | 2 | `FormatNumberSupport`, `CglibAopProxy` |

What this says about the models built on petclinic:

- **JDBC objects (`j1`)** reach nothing here on the measured path: the `DataSource` is Spring's
  `SimpleDriverDataSource`, created by an XML bean definition (reflection), so it has no creator
  and the gate keeps `DataSourceUtils#doGetConnection` → `getConnection` out — the documented
  limitation, measured. The driver's `connect` relation behind it follows from that (its caller is
  not reachable). `j1` still adds 811 edges (`-roots profile`) on JPetStore, none on a measured
  relation.
- **Servlet objects (`s2`)** add nothing: in a WAR the request, response and filter-chain classes
  belong to the container, which is not in the input, so no creator exists.
- **JDK proxies (FRAMEWORKS F3)**: unlike petclinic's 12, **2 of these 7** have a bounded interface
  set — `SqlSessionTemplate`'s constructor creates `Proxy.newProxyInstance(…, new Class[]
  {SqlSession.class}, new SqlSessionInterceptor())`; the 5 mapper proxies take their interface
  from `MapperProxyFactory`'s `mapperInterface` field (configuration: the mapper scanner). The
  petclinic stop decision stands (2 of 19 over both applications), but the bounded case exists.
- **Roots.** The same trace shows mapped vulnerable methods of CVE-2022-22965/CVE-2022-22970
  (`CachedIntrospectionResults#<init>`, 17 traced entries), CVE-2026-41848
  (`AntPathMatcher$AntPathStringMatcher#matchStrings`, 238) and CVE-2026-41851
  (`SpelParserConfiguration#<init>`) running. With `-roots profile` all four are `FRAMEWORK`
  (paths from `ContextLoaderListener#contextInitialized`/`contextDestroyed`, a web.xml root);
  with the default `-roots all` the first three are **`NOT_REACHABLE`**: `all` roots every
  *application* method, and JPetStore's container entry points are library classes named in
  `web.xml` (web.xml roots were `-roots profile` entry points only). Fixed the same day: the
  web.xml lifecycle methods are roots of `-roots all` too, and both scenarios agree. See
  RESULTS.MD, "JPetStore".

## What this says about the next steps

Ranked by measured payoff per effort:

1. **Fix CHA's inherited implementations (`inherited-impl`, 17) — done** (section above).
2. **SAM-call materialization (`sam-call`, 58) — done** (section above: 50 → 22 on a new
   recording, 12 of the rest not lambdas).
3. **Reflection across methods (`reflection`, 33).** REFLECTION R3 — most are framework
   internals (Spring's `ReflectionUtils`, Hibernate accessors).
4. **Servlet API wrappers and `FilterChain` (`skipped-receiver`, ~20 of 87) — done** (section
   above: 33 of the 35 servlet relations). Originally: an extension
   of the servlet dispatch model; the rest of `skipped-receiver` is spread thin over many JDK
   types (collections, I/O), where lifting the skip would fan out (the reason it exists).
5. **JPA/Spring Data proxies (`proxy`, 12)** — FRAMEWORKS F3/F4. Measured 2026-11-09 (section "What the `proxy` relations
   are"): 0 of 12 have a bounded interface set; only a catalog-based model could close them.
6. **JDBC objects (`skipped-receiver`, 22 on the 2026-11-09 recording) — done** (section
   above, `j1`: all 22, plus the `Driver#connect` relation behind them).

Not supported by this measurement: RTA as a default (it adds 105 missed relations with
`-roots profile`), and points-to (GRAPH_PRECISION P3) — what the graph misses is dominated by
unmodeled mechanisms, not by imprecise dispatch.

## What the `reflection` relations are (REFLECTION R3, measured, not built)

2026-11-09, before starting REFLECTION R3 (reflection across methods): the 37 `reflection`
relations of the builder `4` trace above (same recording, CHA, `-roots profile`), each checked
in the caller's bytecode for where the `Method`/`Constructor`/`Class` comes from. This corrects
item 3 of the list above: they are not `ReflectionUtils`/Hibernate-accessor calls on a
`Method` looked up elsewhere.

| Where the reflective object comes from | Relations | Sites |
|---|---|---|
| Spring `@Bean` factory methods: the `factoryMethod` parameter, picked by `ConstructorResolver` from `getDeclaredMethods()` of the configuration class by a name taken from the bean definition | 18 | `SimpleInstantiationStrategy#instantiate` |
| JDK proxy `InvocationHandler` passing on its `Method` parameter (the interface method called on the proxy) | 6 | `SharedEntityManagerCreator$SharedEntityManagerInvocationHandler`/`$DeferredQueryInvocationHandler`, `ExtendedEntityManagerCreator$ExtendedEntityManagerInvocationHandler`, `AbstractEntityManagerFactoryBean#invokeProxyMethod` ×2 |
| Spring MVC handler methods (`getBridgedMethod()`, found by `MethodIntrospector` scanning for `@RequestMapping`) | 3 | `InvocableHandlerMethod#doInvoke` |
| Classes named in bean definitions or `spring.factories` | 3 | `BeanUtils#instantiateClass` ×2, Boot `Instantiator#instantiate` |
| Hibernate defaults: a class from an interface call (`PersisterClassResolver`) or a `Map` lookup by mode | 2 | `PersisterFactoryImpl#createEntityPersister`, `EntityTuplizerFactory#constructTuplizer` |
| SpEL: a name from a template expression | 2 | `ReflectiveMethodExecutor#execute`, `ReflectivePropertyAccessor$OptimalPropertyAccessor#read` |
| Lookup and call in one method, but the `Class` comes from a facade or a field | 2 | `BeanValidationIntegrator#integrate` (`ClassLoaderService#classForName`), devtools `RestartLauncher#run` (`Class.forName(this.mainClassName)`) |

**0 of 37** have the shape R3 targets — a `Method`/`Constructor` found by constant names (R1/R2)
in one method and called in another — so building it would change nothing here; the
2026-11-05 probe over every `Method#invoke` site of petclinic and WebGoat found the same (0
sites). `getDeclaredMethods()`/`getMethods()` filtered by name is not a bounded set either: every
filter here takes its name from annotations or bean definitions. The bounded sets are
annotation-defined (all `@Bean` methods, all request-mapping methods), which the spring profile
already roots — an edge there would shorten paths, not change a verdict. Configuration-sourced
names (`spring.factories`, 3 relations) and `setAccessible` were not measured as worth a model.
Not done.
