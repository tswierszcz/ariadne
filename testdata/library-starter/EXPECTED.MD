# Starter oracle: mybatis-spring-boot-autoconfigure 2.1.2 (LIBRARY_MODE, after LB4)

Question: of the 26 `REACHABLE` findings `-roots api` gives the mybatis
Spring Boot starter (`testdata/library-lb4/EXPECTED.MD`; CHA, `lb2-6`), which
mapped vulnerable methods actually run in a real application of the starter?

**Application** (`app/fx`): Spring Boot 2.2.5 with the starter's 25 jars,
embedded HSQLDB 2.5.1, one `@Mapper` with one `@Select`, no web server, no
SpEL, no `logback.xml`, no dynamic SQL. Two scenarios: `plain` (no
configuration file) and `yml` (`yml/application.yml` with one `mybatis.*`
property — what most Boot users have). `scripts/library-starter-oracle.sh`
runs it under `scripts/MethodBreakTrace.java` (JDI, one-shot breakpoints on
every method of each mapped class; 9 s, where tracing every method entry
did not finish in 8 min) and writes `observed-plain.txt`/`observed-yml.txt`.
`mapped.txt` lists the mapped methods of the 26 findings. Jars come from
`work/libs/mybatis-starter/all` (`scripts/library-measure.sh`) and
`~/.m2`; no network.

## Result (2026-11-19)

| Ran | Advisories | First call out of mybatis on the reported path |
|---|---|---|
| in both scenarios (5) | CVE-2022-22965, CVE-2022-22970 (`CachedIntrospectionResults`, every Spring app); CVE-2026-41848 (`AntPathMatcher` from `SpringBootVFS#list`); CVE-2021-42550 (only the helper `OptionHelper#isEmpty` of an 88-method diff mapping); CVE-2026-41851 (only `SpelParserConfiguration#<init>` of an 81-method diff mapping) | `BeanWrapperImpl`, `PathMatchingResourcePatternResolver`, slf4j `<clinit>`, `ObjectProvider#getIfAvailable` |
| only with `application.yml` (7) | CVE-2022-1471, CVE-2022-25857, CVE-2022-38749/38750/38751/38752, CVE-2022-41854 (snakeyaml, `supplied_by: undeclared`) | `ObjectProvider#getIfAvailable`, `AutoConfigurationPackages#get` |
| never (14) | 7 SpEL (CVE-2022-22950, CVE-2023-20861, CVE-2023-20863, CVE-2024-38808, CVE-2026-41849/41850/41852); 4 logback configuration (CVE-2023-6378, CVE-2024-12801, CVE-2025-11226, CVE-2026-1225); CVE-2022-27772 (embedded web server temp dir); CVE-2024-38820 (`MimeType`, 50-method mapping); CVE-2026-113730 (OGNL) | `ObjectProvider#getIfAvailable` (8), `AutoConfigurationPackages#get` (2), slf4j `<clinit>` (3), `XMLConfigBuilder#parse` (1) |

Reading it:

- **The snakeyaml seven are real for a Boot user with YAML, but not through
  mybatis.** On the JVM `Yaml#loadAll` runs from Boot's own configuration
  loading; the reported paths (`MybatisAutoConfiguration#<init>` →
  `ObjectProvider#getIfAvailable` → CHA → `YamlMapFactoryBean` →
  `Yaml#loadAll`) never ran. The verdict would be the same without the starter.
- **The 14 that never ran need the application to do something:** use SpEL
  (`#{…}`), configure logback with XML, run an embedded web server, write
  dynamic SQL (OGNL). One scenario cannot show they never run; it shows the
  starter alone does not run them.
- **Where the reported paths leave mybatis:** 10 of the 14 and all 7 snakeyaml
  paths leave through a call on an object the **container hands to the
  library** — an `ObjectProvider` constructor parameter (`getIfAvailable`) or
  the `BeanFactory` given to `AutoConfigurationPackages#get` — which CHA
  dispatches to every Spring implementation (`StaticListableBeanFactory`,
  `YamlMapFactoryBean`, bean-definition SpEL…). The library does not choose
  that implementation; the application's container does. This is the LB2
  open item "calls on dependency types", seen on a real library.
- **`heuristic` does not separate them:** 12 of the 14 that never ran have
  `heuristic` paths, but so do 9 of the 12 that ran.
- **Three of the five that run are mapping coarseness, not library mode:**
  the run method is a helper the fix touched (`OptionHelper#isEmpty`,
  `SpelParserConfiguration#<init>`) or a class every Spring app uses
  (`CachedIntrospectionResults`, where the exploit needs data binding on a
  WAR deployment).

So for this starter, 5 of 26 `REACHABLE` findings are reached through the
starter in its smallest application, 7 more are reached by Boot itself when
the application has YAML, and 14 depend on what the application does —
reached in the graph through calls on container-supplied objects.

## After `lb2-7` (caller-supplied objects, 2026-11-19)

`lb2-7` marks calls on objects the container hands the library
(`testdata/library-supplied/EXPECTED.MD`). On this starter (CHA): 25 caller-supplied sites (one on a
field), `MybatisAutoConfiguration#<init>` — its `ObjectProvider` parameters — is
`supplied_dispatch` in all 26 findings; verdicts and exposed methods unchanged; no finding is
`SUPPLIED_DISPATCH_ONLY`. In every finding 7–11 of the 13 exposed methods reach the code through
**logging initialization** instead: `MybatisAutoConfiguration#<clinit>` (or anything that creates a
logger) → `LoggerFactory#getLogger` → `StaticLoggerBinder#getLoggerFactory` → (CHA over
`ContextSelector`) `ContextJNDISelector` → logback's `GenericConfigurator` → (CHA over joran
`Action`s) `SpringPropertyAction` → `Environment#getProperty` → conversion service → SpEL /
`BeanFactory` → `YamlMapFactoryBean`… — interface calls inside the dependencies whose
implementation is chosen by configuration (a system property, a `logback-spring.xml`). That, not
the caller-supplied objects, is what keeps the 14 never-run findings `REACHABLE`.

## Where the paths leave the real execution (2026-11-19)

Method: the classes on all 364 reported paths of the 26 `REACHABLE` findings (finding paths and
`exposed_api` paths, `lb2-7`, CHA; `scripts/path-divergence.py classes`) traced in both scenarios
with `scripts/MethodBreakTrace.java`; per path, the first method that never ran
(`scripts/path-divergence.py diverge`). For each divergence at a CHA dispatch site, the CHA
targets of that call site were listed from the graph (a scratch tool) and traced too.

- **9 of 364 paths have every method run** (a set, not an ordered trace): CVE-2022-22965/22970 from
  `MybatisAutoConfiguration#sqlSessionFactory` and `…$AutoConfiguredMapperScannerRegistrar` through
  `BeanWrapperImpl` to `CachedIntrospectionResults#<init>`, CVE-2021-42550 from
  `MybatisAutoConfiguration#<clinit>` through slf4j and logback's initialization to
  `OptionHelper#isEmpty` (a helper of an 88-method mapping). Both scenarios give the same
  divergences.
- **Divergence edges** (18 distinct): `StaticLoggerBinder#getLoggerFactory` →
  `ContextJNDISelector#getLoggerContext` (93 paths, interface/CHA); lambdas and private helpers
  whose branch did not run (`resolveMapperLocations`, `MyBatisExceptionTranslator`,
  `SqlSessionFactoryBean#scanClasses` — value-dependent, 70); roots this application never calls
  (`SpringBootVFS#list`, `VFS#list`, `SpringBootVFS#<init>`, 79); and further CHA sites in the
  container (`AbstractResource#exists`, SpEL `getValue`, `BeanPostProcessor`, `ObjectProvider`,
  `BeanFactory#getBean`, `InitializingBean`).
- **At 11 of the 13 diverging CHA sites the call did happen — the JVM chose another
  implementation**: `DefaultContextSelector` (not `ContextJNDISelector`, which logback uses only
  with `-Dlogback.ContextSelector=JNDI`), `LiteralExpression` (not SpEL: no `#{…}`),
  `DependencyObjectProvider` (not `StaticListableBeanFactory`), `ClassPathResource#exists`, the
  bean factory's own `getBean`… The implementation is chosen by configuration or by which beans
  exist. The other 2 sites never ran.
- **Experiment E1** (a scratch build dropping the one edge to `ContextJNDISelector`, i.e. modeling
  logback's default selector): no verdict changes, on the starter (CHA/RTA) or on petclinic
  (`-roots all`, CHA: its 3 logback-configuration advisories stay `FRAMEWORK` through other
  routes). One model does not help: there are many redundant over-approximate routes.
- **Experiment E2** (upper bound, oracle-guided: at every diverging CHA site where the JVM ran
  another target, keep only the targets that ran; then trace the new paths and repeat): round 1
  drops 135 edges at 11 sites, rounds 2–4 another 119 at 7 more sites, round 5 finds none left —
  **all 26 stay `REACHABLE`** (CHA), the exposed methods fall from 338 to 175. *(A first run of this
  experiment reported "all 26 `NOT_REACHABLE`": its rule separator was `;`, which JVM descriptors
  contain, so one empty rule dropped every edge. Corrected and re-run.)* What holds them after that
  (201 paths, 16 with every method run): **branches that depend on configuration or data**, called
  directly — `DataSourceInitializerInvoker#initialize` (only with `schema.sql`/`data.sql`), logback's
  `ContextInitializer#configureByResource` (only with a `logback.xml`), the exception-translation
  lambda (only on an error), `SqlSessionFactoryBean#scanClasses` (only with
  `type-aliases-package`), error paths (`raiseNoMatchingBeanFound`) — plus 4 CHA sites that never
  ran and API methods this application never calls.
- **`-profile spring`** (Spring packages are framework code, as for the reference applications):
  25 of 26 are `FRAMEWORK` (CHA and RTA alike) — "reachable only through framework code" — and one
  stays `REACHABLE`: CVE-2026-113730, mybatis → its embedded OGNL through `XMLConfigBuilder`, which
  runs only with an XML configuration using `${…}` (this application has none). The measurement
  runs (`scripts/library-measure.sh`) use `-profile generic`, where every path is "app only".

## After `lb2-8` (Spring-announcing libraries are framework-mediated, 2026-11-19)

With `-roots api` this library's `spring.factories` makes the spring profile's framework packages
framework code under any `-profile`. `scripts/library-measure.sh -offline` (`-profile generic`):
CHA 25 `REACHABLE` → `FRAMEWORK`, 1 `REACHABLE` (CVE-2026-113730, OGNL); RTA 23 → `FRAMEWORK`, 1
`REACHABLE`. Against the oracle: all 12 whose mapped methods run (5 plain, 7 with YAML) are
`FRAMEWORK` — still reported, as reached through the container — and the one `REACHABLE` runs only
with an XML configuration using `${…}`. The other four measured libraries and every `-roots all`
run are unchanged.
