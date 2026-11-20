# Library mode qualification set (LIBRARY_MODE LB4, slice 1, 2026-11-19)

Five real libraries, one per API style, with their real (vulnerable)
dependencies. `scripts/library-measure.sh` fetches them into `work/libs/`
(once, from Maven Central; afterwards `-offline`) and runs `-roots all` and
`-roots api` under CHA and RTA. Only jetty-client is committed here (its jars
and the 25 cache entries its run reads), because it is the one whose verdicts
were checked on a JVM in this slice (`TestLibraryJetty`,
`scripts/library-lb4-oracle.sh`); the other four were checked by hand in the
LB2/LB3 measurements (RECOMMENDATION/IMPLEMENTATION_STATUS.MD) and are
re-measured here on `lb2-6`.

| Library | API style | Dependencies analyzed |
|---|---|---|
| commons-configuration2 2.7 | facade / builder | commons-text 1.8, commons-lang3 3.9 |
| jackson-dataformat-yaml 2.12.3 | factory, mapper subclass (`YAMLMapper extends ObjectMapper`) | jackson-databind/core 2.12.3, snakeyaml 1.27 |
| jjwt-jackson 0.11.2 | SPI (`META-INF/services`) | jackson-databind 2.9.10.4 |
| mybatis-spring-boot-autoconfigure 2.1.2 | Spring Boot autoconfiguration | the starter's 25 jars (mybatis 3.5.4, Spring Boot 2.2.5, Spring 5.2, logback, snakeyaml) |
| jetty-client 9.4.36 | callbacks / async (`Response.Listener`, NIO selector) | jetty-http/io/util 9.4.36 |

CDI bean archives (smallrye-context-propagation 2.1.0, mutiny 2.5.8) were
measured in `lb2-6`: no new roots, verdicts unchanged.

## Results (`-roots api`, `lb2-6` analyzer, offline)

| Library | API methods | `REACHABLE` CHA / RTA (`all`: CHA / RTA) | other | `exposed_api` per `REACHABLE` (max) |
|---|---|---|---|---|
| commons-configuration2 | 2156 | 2 / 2 (2 / 2) | — | 778 |
| jackson-dataformat-yaml | 888 (640 inherited) | 17 / 17 (1 / 1) | 9 `NOT_REACHABLE` | 97 |
| jjwt-jackson | 10 | 15 / 14 (15 / 15) | 21 / 22 `NOT_REACHABLE` | 4 |
| mybatis-spring-boot-autoconfigure | 53 (+9 announced classes) | 26 / 24 (26 / 26) | 12 / 14 `NOT_REACHABLE`, 1 `UNKNOWN` | 13 |
| jetty-client | 1175 (129 inherited) | 2 / 2 (2 / 2) | 1 `NOT_REACHABLE`, 2 `UNKNOWN` | 200 |

`-roots api` never loses a `REACHABLE` that `-roots all` has (one RTA
difference in jjwt, CVE-2026-68494: the async parser jjwt never creates —
RTA is right). It gains 16 on jackson-dataformat-yaml (methods inherited from
`ObjectMapper`/`JsonFactory`, `lb2-3`): the only place where `api` reports
more than `all`.

## Hand-checked verdicts

- **commons-configuration2:** both `REACHABLE` real — every getter interpolates
  through `ConfigurationInterpolator#interpolate` → `StringSubstitutor#replace`
  (CVE-2022-42889); `ConstantLookup#fetchClass` → `ClassUtils#getClass`
  (CVE-2025-48924). Exposure is wide (≈775 methods) because it is.
- **jackson-dataformat-yaml:** 13 of the 16 inherited `REACHABLE` are databind
  deserialization through `readValue`/`readValueAs` — real for a `YAMLMapper`
  user; CVE-2025-49128 through `ObjectMapper#_jsonFactory` — real (JVM probe,
  `lb2-4` follow-up); CVE-2026-68494 and CVE-2026-89425 come from a client
  `JsonParser` parameter of `JsonGenerator#_copyCurrentContents` (any parser
  the client passes — the LB1 parameter approximation). snakeyaml's six CVEs
  `NOT_REACHABLE` — right, Jackson uses the event parser, not `Yaml#load`.
- **jjwt-jackson:** 14 of 15 through `JacksonDeserializer#readValue` into
  databind (jjwt deserializes every JWT header and body with it; the
  deserializer reached depends on the claim types — not JVM-checked), one of
  them, CVE-2020-25649, mapped only to `DOMDeserializer#<clinit>`
  (`MAPPING_INITIALIZER_ONLY`); CVE-2025-49128 through
  `JacksonSerializer#serialize` (an error message's source description). The
  gadget-blocklist family is gone since the initializer redirect
  (2026-11-17).
- **mybatis-spring-boot-autoconfigure:** 22 of 26 `REACHABLE` paths are
  `heuristic` and run through the Spring container, not mybatis: e.g.
  `MybatisAutoConfiguration#<init>` → `ObjectProvider#getIfAvailable` →
  (CHA, every `BeanFactory`) → `YamlMapFactoryBean#getObject` → `Yaml#loadAll`
  (CVE-2022-1471; only if the application defines such a bean), or →
  `doResolveDependency` → SpEL (only with `#{…}` in bean definitions).
  The logback ones (`MybatisAutoConfiguration#<clinit>` → slf4j → logback
  autoconfiguration) do run, but logback is the application's
  (`supplied_by: undeclared`, 12 of 26). For a starter, `REACHABLE` mostly
  means "the container the starter plugs into reaches it".
- **jetty-client (JVM-checked):** a user calling only `HttpClient#start`,
  `#GET`, `#stop` against a local socket answering `Content-Length: +5` gets
  `status=200 body=hello` — the vulnerable parser accepts it — and the JVM
  enters `HttpURI#parse(State,…)` (CVE-2022-2047) and
  `HttpParser#convertContentLength` (CVE-2023-40167) (`observed.txt`). Both
  `REACHABLE`, `HttpClient#GET` exposed (both overloads); 85 of the 200
  exposed jetty-client methods ran in this one scenario (the rest — POST,
  other listeners, connection pools — were not exercised: no claim either
  way). Both advisories needed manual mappings (OSV gives no fix commit;
  bytecode of 9.4.36 vs 9.4.57). CVE-2024-6763 `NOT_REACHABLE` — right by
  its advisory (Jetty's own use of `HttpURI` is not vulnerable), though the
  reason (`MAPPED_CODE_ABSENT`: the mapping comes from the Jetty 12 fix) is
  incidental. CVE-2025-11143 and CVE-2026-2332 `UNKNOWN`
  (`MAPPING_MISSING`): no fix commit, and 9.4 is end-of-life, so no fixed
  9.4 release to diff.

## JVM oracles for the other three libraries (2026-11-19)

`scripts/library-lb4-jvm-oracle.sh` builds, per library, a client that uses only the library's
public API (`<lib>/client/fx`) and runs it under `scripts/MethodBreakTrace.java` (one-shot JDI
breakpoints) on every class of the library, every class of a mapped method of any finding
(`<lib>/mapped.txt`, the `-roots api` CHA verdicts at `lb2-8`) and, for jackson-dataformat-yaml,
`ObjectMapper` (`trace-classes.txt`: the inherited API). Output: `<lib>/observed.txt` (stable over
two runs). Jars from `work/libs` (`scripts/library-measure.sh`); no network.
`TestLibraryLB4Oracles` checks the recorded files.

| Library | Client (public API only) | `REACHABLE` whose mapped code ran | `NOT_REACHABLE` that ran |
|---|---|---|---|
| commons-configuration2 | `FileHandler#load` of a properties text with `${sys:…}` and `${const:…}`, `getString` | 2 of 2 | — (none) |
| jackson-dataformat-yaml | `YAMLMapper#readValue` (inherited) into a map, POJOs with enum/`Date`/`Path`/`UUID`/`InetSocketAddress`, `@JsonUnwrapped`, `@JsonTypeInfo` name and class ids, object-id forward references, `UNWRAP_SINGLE_VALUE_ARRAYS`, `activateDefaultTyping` with `BasicPolymorphicTypeValidator` / `DefaultBaseTypeLimitingValidator`, a type mismatch | 15 of 17 | 0 of 9 |
| jjwt-jackson | `JacksonSerializer#serialize`, `JacksonDeserializer#deserialize` with and without claim types (`Path`, `UUID`, `InetSocketAddress`, `Locale`, `@JsonUnwrapped`, polymorphic, object ids, `org.w3c.dom.Document`), malformed JSON | 14 of 15 | 0 of 21 |

- For every finding whose mapped code ran, the API method the client called
  (`AbstractConfiguration#getString`/`FileHandler#load`, `ObjectMapper#readValue` inherited by
  `YAMLMapper`, `JacksonDeserializer#deserialize`/`JacksonSerializer#serialize`) is in its
  `exposed_api`.
- **Not shown (3):** CVE-2026-68494 in both Jackson libraries (`NonBlockingJsonParser`, the async
  parser — neither library creates one; in jjwt RTA already says `NOT_REACHABLE`) and
  CVE-2026-89425 in jackson-dataformat-yaml (`_reportInvalidToken` of the JSON parsers —
  `YAMLMapper` parses with `YAMLParser`). Both are reached in the model only through a `JsonParser`
  the caller would pass to an inherited method or through CHA over `JsonParser#nextToken`: a
  caller-supplied parser (the `lb2-7` idea, which does not cover inherited roots) — over-approximate
  for ordinary use.
- No `NOT_REACHABLE` finding's mapped code ran in any scenario: no false negative was found.
  Together with jetty-client (above) every library of the set now has a JVM oracle; the starter has
  its own (`testdata/library-starter`).

## Limits found (not fixed in this slice)

1. **Async hand-off is not in the graph.** In jetty-client the real path is
   `HttpClient#GET` → … → the NIO selector thread → `ReadCallback#succeeded`
   → `HttpReceiverOverHTTP#parse`. The graph has no edge for the selector
   wake-up; the verdict comes out right only because CHA finds another route
   (`HttpRequest#abort` → `ResponseNotifier#notifyComplete` →
   `AuthenticationListener#onComplete` → `HttpConnection#send`, …) whose
   middle never ran. Example paths of callback-style libraries are therefore
   not execution traces. *(2026-11-19, ASYNC_CALLBACKS `a3-2`: a JVM stack
   oracle, `scripts/MethodStackTrace.java`, showed the one real edge the graph
   lacks is the pool's `runJob` → `task.run()`; tasks handed through a
   method's parameter to `Executor#execute` are now resolved at its callers,
   adding `SelectorManager#execute` → `ManagedSelector$CreateEndPoint#run`.
   Fully executed reported paths 40 → 66 of 402, verdicts unchanged; the path
   from `HttpClient#GET` still is not real — the connect request crosses
   Jetty's own selector queue, `ManagedSelector#submit`.)*
2. **Public is not intended API.** jetty-client's shortest finding path starts
   at `HttpReceiverOverHTTP#receive`, public in a public package, which no user
   calls. `-api-exclude` helps only when internals live in separate packages.
3. **`heuristic` is not a noise filter.** It marks wide CHA dispatch: real on
   jackson/jjwt (databind's deserializer dispatch), container
   over-approximation on the starter.
4. **Exposure lists are large** where the library is a facade (775) or a
   client (200): useful as evidence, not as a list for a person to read.
