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
