# Control library for library mode (LIBRARY_MODE LB0)

Written 2026-11-16, before LB1. The contract is `docs/RESULTS_CONTRACT.MD` §8.

`lib.jar` (sources in `lib/`) uses commons-text 1.9's `StringSubstitutor#replace`
(CVE-2022-42889, a manual-table entry; commons-lang3 3.11's `ClassUtils#getClass`, behind
it, is CVE-2025-48924). `deps/` holds both jars from Maven Central (checked by
`scripts/library-oracle.sh`). `client.jar` (`client/fx/Client.java`) is the oracle's
client: it calls each public entry once and passes its own `Listener`.

## Cases

| Case | Code | Public API? | Reaches `replace` on the JVM? |
|---|---|---|---|
| (a) | `Templates#render` → package-private `Engine#expand` → `replace` | yes | yes |
| (b) | `Formatter#unsafe` is `private` and never called (`Formatter#format` is the API) | `format` yes, `unsafe` no | no |
| (c) | package-private class `Hidden`, `public run` → `replace`; nothing calls it | no (inaccessible class) | no — no client can call it |
| (d) | `Pipeline#run(Listener, …)` calls `listener.onEvent`; the library's package-private `ExpandingListener` (→ `replace`) is never created | `run` yes | no — the client's listener runs there |
| (e) | `Codecs#decodeAll` runs every `ServiceLoader` `Codec`; the library registers `lib.internal.ExpandCodec` (→ `replace`) | `decodeAll` yes; `ExpandCodec` (public, and a registered provider) yes | yes |
| (f) | `lib.internal.Util#expand` → `replace` | yes (public), unless `-api-exclude lib.internal` | yes |
| (g) | `Expander extends StringSubstitutor`, nothing declared (added 2026-11-17, `lb2-3`) | the inherited `StringSubstitutor` methods, `replace` among them | yes |
| (h) | `SafeExpander extends StringSubstitutor`, overrides `replace(String)` to return its input (`lb2-3`) | its own `replace(String)` (reaches nothing); the other inherited overloads | its own `replace` runs, not the inherited one |
| (i) | `StrictExpander extends StringSubstitutor`, overrides the protected `substitute(TextStringBuilder, int, int)` — which every inherited `replace` calls on `this` — to return `false` (`lb2-4`) | every inherited `replace`; its own `substitute` | the inherited `replace` runs, then its own `substitute`; never `StringSubstitutor#substitute` or `resolveVariable` |

`observed.txt` (`scripts/library-oracle.sh`, JDI trace of `fx.Client`): the JVM ran
`Templates#render`, `Engine#expand`, `Formatter#format`, `Pipeline#run`,
`Codecs#decodeAll`, `ExpandCodec#decode`, `Util#expand` and `StringSubstitutor#replace`;
it never ran `Formatter#unsafe`, `Hidden#run` or `ExpandingListener#onEvent`.

## Today (before LB1): `-roots all`

Every method of `lib.jar` is a root, so the finding is `REACHABLE` with the path
`Engine#expand → StringSubstitutor#replace` (a package-private class's method), and nine
library methods reach `replace` — the six above plus `Formatter#unsafe`, `Hidden#run` and
`ExpandingListener#onEvent` — under CHA and under RTA (with every method a root,
`ExpandingListener`'s constructor is one too, so RTA keeps it). Measured 2026-11-16.

## Expected with `-roots api` (LB1)

CVE-2022-42889 and CVE-2025-48924: `REACHABLE`, with `exposed_api` (method names; every
overload of the listed name):

| Graph | `exposed_api` |
|---|---|
| CHA | `Codecs#decodeAll`, `ExpandCodec#decode`, `Pipeline#run`, `Templates#render`, `Util#expand` (5) — `Pipeline#run` through the CHA step `Listener#onEvent` → `ExpandingListener#onEvent`, character `heuristic`: an over-approximation the JVM does not confirm |
| RTA | `Codecs#decodeAll`, `ExpandCodec#decode`, `Templates#render`, `Util#expand` (4) — nothing creates `ExpandingListener` from the API roots |
| CHA, `-api-exclude lib.internal` | `Codecs#decodeAll`, `Pipeline#run`, `Templates#render` (3) — internal code reached through the API still counts |

Never in `exposed_api` (not roots): `Engine#expand`, `Formatter#unsafe`, `Hidden#run`,
`ExpandingListener#onEvent`. `Formatter#format` is a root that reaches nothing.

Every `exposed_api` method the JVM ran reaches `replace` on the JVM too (the oracle), and
every method the JVM never entered is absent from `exposed_api` under RTA.

Case (d) is the client-code case LB2 makes explicit: under CHA the library's own
`ExpandingListener` stands in for client listeners; LB2 reports the call as a
`client_code` gap.

## Inherited methods (`lb2-3`, written with cases (g) and (h), 2026-11-17)

`StringSubstitutor#replace(String)` is exposed with `inherited_by` = `Expander` only:
`SafeExpander` overrides it. The other `replace` overloads are inherited by both. The JVM
ran `Expander#<init>`, `SafeExpander#replace` and `StringSubstitutor#replace(String)`
(`observed.txt`). The `exposed_api` sets of the tables above count only the library's own
methods; inherited entries come on top of them. `SafeExpander#replace` is never exposed.

## Receiver specialization (`lb2-4`, case (i) written before the change, 2026-11-17)

Expected and measured (CHA and RTA): `StrictExpander` is in `inherited_by` of every
inherited `replace` overload for CVE-2022-42889 — the overloads are its mapped code — but
of none for CVE-2025-48924: on a `StrictExpander` they call its own `substitute`, which
reaches no lookup. It stays in `inherited_by` of `resolveVariable` (inherited directly).
Before `lb2-4` (the inherited root was the shared `StringSubstitutor` node, so the call
on `this` went to every `substitute`) it was listed for both advisories.
`observed-strict.txt` (`fx.Client strict`, case (i) alone) is the oracle:
`StringSubstitutor#replace(String)` and `StrictExpander#substitute` ran,
`StringSubstitutor#substitute` and `resolveVariable` did not.

## LB2 (`lb2-1`, `lb2-2`, measured 2026-11-17)

Client-code sites (a `client_code` gap, 3 sites, all reachable; never `MODEL_INCOMPLETE`):
`Pipeline#run` → `Listener#onEvent` (abstract), `Codecs#decodeAll` → `Codec#decode`
(abstract), `Templates#render` → `Templates#prepare` (protected, overridable). With
`-api-exclude lib.api` there are none.

`client_dispatch`: under CHA (with or without `-api-exclude lib.internal`) only
`Pipeline#run` — its one path dispatches `Listener#onEvent` to `ExpandingListener`, which
nothing creates; that is the one exposed method the JVM does not confirm. `Codecs#decodeAll`
is not marked: its dispatch reaches `ExpandCodec`, which the library's `ServiceLoader`
creates. Under RTA nothing is marked (the edge is pruned). The `exposed_api` sets above are
unchanged.
