# Control library for caller-supplied receivers (LIBRARY_MODE LB2, `lb2-7`)

Written 2026-11-19, before the implementation. Contract: `docs/RESULTS_CONTRACT.MD` §8.

Motivation: `testdata/library-starter` (JVM oracle) — 14 of the mybatis starter's 26
`REACHABLE` findings never ran in a real application, and the reported paths leave mybatis
through calls on objects the container hands to the library (`ObjectProvider` constructor
parameters, a `BeanFactory` set through `BeanFactoryAware`), which CHA dispatches to every
implementation in the dependencies. The implementation that runs there is chosen by whoever
supplies the object, not by the library.

`lib.jar` (sources in `lib/`) depends on commons-text 1.9 and commons-lang3 3.11
(`testdata/library/deps`). commons-text's `ConstantStringLookup#lookup` reaches commons-lang3's
`ClassUtils#getClass` (CVE-2025-48924); `StringSubstitutor#replace` is CVE-2022-42889.
`StringLookup` is a commons-text (dependency) interface.

| Case | Code | Receiver of `StringLookup#lookup` | `supplied_dispatch` |
|---|---|---|---|
| (y) | `sp.supplied.Resolver#resolve(StringLookup, String)` | a parameter of the API method | yes |
| (z) | `sp.supplied.Holder#get(String)`, field set only by `setLookup(StringLookup)`, which the library never calls | a field written only from the parameter of an API method no library code calls | yes |
| (aa) | `sp.api.Constants#constant(String)` | `StringLookupFactory.INSTANCE.constantStringLookup()` — the library's own | no |
| (bb) | `sp.api.Mixed#get(String)`, field set by `Mixed(StringLookup)` and by `Mixed()` to the library's default | a field also written with the library's own value | no |
| (cc) | `sp.api.Registrar#expand(StringLookup, String)` → `new StringSubstitutor(lookup).replace(…)` | the dependency's own field, holding the caller's value | no — **limit**: a value stored inside a dependency object is not followed |
| (dd) | `sp.api.Wrapper#wrap(String)` → `new Resolver().resolve(<the library's constant lookup>, key)` | (y)'s parameter, but filled by the library, not a caller | no — (y)'s parameter is caller-supplied only on a path that enters at (y) |

(dd) was added after the first implementation (and fields now count only when the roots writing
them are never called by the library itself — on jetty-client `AbstractConnectionPool`'s `Callback`
field is set by its public constructor, which the library calls with its own destination): on jackson-dataformat-yaml it marked
`JsonFactory#createJsonParser`, whose path runs through the public constructor `YAMLParser#<init>`
with the library's own `IOContext` — a root's parameter had been taken as caller-supplied on every
path through the root, not only on paths that start there.

## Before (measured 2026-11-19, `lb2-6`)

`-roots api` (CHA): CVE-2022-42889 `REACHABLE` (exposed: `Registrar#expand`), CVE-2025-48924
`REACHABLE` (exposed: `Constants#constant`, `Holder#get`, `Mixed#get`, `Registrar#expand`,
`Resolver#resolve`), every entry alike — nothing tells (y)/(z) from (aa).

## Expected (`lb2-7`), CHA and RTA

- Verdicts unchanged: both `REACHABLE`, the same exposed methods.
- CVE-2025-48924: `Resolver#resolve` and `Holder#get` carry `supplied_dispatch: true`; the others
  (`Wrapper#wrap` too) do not. The finding has no `SUPPLIED_DISPATCH_ONLY` (three entries reach it without one).
- CVE-2022-42889: `Registrar#expand` not marked.
- With `-api-exclude sp.api`: CVE-2022-42889 `NOT_REACHABLE`; CVE-2025-48924 `REACHABLE`,
  exposed exactly (y) and (z), both marked, and the finding carries `SUPPLIED_DISPATCH_ONLY`.
  *Measured:* so under CHA; under RTA CVE-2025-48924 is `NOT_REACHABLE` there — with `sp.api`
  excluded nothing in the analyzed code creates a `ConstantStringLookup` (RTA's created-class rule
  covers the library's accessible classes, not a dependency's factory a client may call). That is
  RTA's existing approximation, not this rule.
- Gap sites (mechanism `client_code`, kind `supplied`): the two `StringLookup#lookup` calls of (y)
  and (z); none for (aa), (bb), (cc), (dd). Never `MODEL_INCOMPLETE`. *(A first implementation
  also counted calls passing a supplied value to a dependency method — (cc)'s
  `StringSubstitutor#<init>` — and marked the whole edge; on the mybatis starter that marked
  `sqlSessionTemplate(SqlSessionFactory)`, whose path never touches the factory again. Removed:
  only calls on the supplied object are sites.)*

The marking says "the code that runs here is the one the caller supplies"; it is not a claim that
the path is false — a caller who passes `StringLookupFactory.INSTANCE.constantStringLookup()` to
(y) does run `ClassUtils#getClass`.
