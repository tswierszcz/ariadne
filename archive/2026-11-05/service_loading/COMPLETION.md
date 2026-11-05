# SERVICE_LOADING (S0–S4) — completion record

**Archived:** 2026-11-05. **Status: DONE** for every phase's declared scope. The
`docs/SUPPORT.MD` row "ServiceLoader and JPMS service models" stays **Partial**:
the remaining limitations below are real and stated there.

This directory holds the area's two former live documents (`SERVICE_LOADING.MD`,
`SERVICE_LOADING_IMPLEMENTATION_PLAN.MD`) plus the per-phase records
`SERVICE_LOADING_S2.md` and `SERVICE_LOADING_S3.md`. Detailed per-increment notes
(scope, files, commands, numbers) stay in `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.
Current behavior: `docs/SUPPORT.MD` (the row above and the section "ServiceLoader:
exact supported signatures and scenarios"), README, FAQ.

## What shipped

| Phase | What it did | Commits |
|---|---|---|
| S0 | `META-INF/services` inventory with diagnostics (duplicates, missing classes); registrations are never edges | `61f7ec8` |
| S1 | `ServiceLoader.load*` sites, service `Class` argument via DATA_FLOW local values, loader kinds | `df546f7` |
| S2 | Lazy materialization at `Iterator.next()`/`Provider.get()` linked to the exact load site, constructor/`<clinit>` and descriptor-exact dispatch edges through the B3 coordinator; provenance through helpers, D3 fields, linear `stream().findFirst()` | `30859e5`, `2e3ab2f`, `877c788` |
| S3 | `Module` attribute and module inventory; `-jpms-scenario classpath\|module-path\|mixed` with `provides`, `provider()`, automatic modules, `uses`, `requires` warnings, `ModuleLayer.boot()` | `749c380`, `ef9d2f0`, `5f71000`, `fa297ef` |
| Follow-ups | Recognized `ClassLoader` shapes; DATA_FLOW D3 parameters from callers and final fields from constructor parameters; `AccessController.doPrivileged` | `26eee7e`, `76f39ba`, `7f9a7b3`, `0b20992`, `709b7f4` |
| S4 | Real-JVM oracle (`scripts/serviceloader-oracle.sh`, `testdata/oracle/`, `TestServiceLoaderOracle`); compiled-fixture duplicate and two-loader cases; historical measurement; consolidated supported-signature list | this commit |

## S4 evidence

- **Oracle:** 62 entry/scenario pairs across `testdata/serviceloader` and
  `testdata/jpms`, class-path / module-path / mixed launches on JDK 25: the model
  connects every provider the JVM instantiated except one declared gap
  (`stream().map(Provider::get)`, reported Incomplete) and connects **no** provider the
  JVM did not instantiate. Mutation checks (boot-layer filter disabled; an explicit
  module's `META-INF/services` not ignored) fail the test.
- **Fixture cases:** a provider listed twice in one file and shipped in two jars is
  materialized once with a warning; two loaders for two services in one method link
  only their own providers (`TestRegression_duplicatesAndTwoLoaders`).
- **Stream API:** not extended — 0 `ServiceLoader.stream()` uses on petclinic/WebGoat;
  callback pipelines remain an explicit gap, as S4 itself requires ("dopiero z modelem
  operacji").
- **Historical measurement** (offline, `-profile spring`, fixed input, one build per
  commit):

| Build | petclinic resolved / linked / model edges / callback edges / time | WebGoat resolved / linked / model edges / callback edges / time |
|---|---|---|
| `30859e5` (S2 start) | 13 / – / – / – / 6.6 s | 31 / – / – / – / 10.5 s |
| `2e3ab2f` (B3 + dispatch) | 13 / 39 / 4 / 62 / 8.8 s | 31 / 87 / 75 / 42 / 14.1 s |
| `877c788` (D3 fields) … `26eee7e` | 13 / 39 / 4 / 62 / ~8.4 s | 31 / 90 / 75 / 42 / ~13.2 s |
| `7f9a7b3` (D3 parameters) | 20 / 45 / 7 / 62 / 9.3 s | 38 / 99 / 79 / 42 / 14.5 s |
| `0b20992` (D3 ctor fields) | 25 / 46 / 7 / 62 / 9.3 s | 43 / 100 / 79 / 42 / 14.5 s |
| `709b7f4` (doPrivileged), S4 | 25 / 52 / 9 / 288 / 9.5 s | 43 / 106 / 79 / 284 / 14.7 s |

  `results.json` is **byte-identical across all eleven builds** on both apps
  (`3/62/26/51/6`, `42/44/18/56/0`): the area raised model coverage and removed gaps
  without changing any CVE verdict on these two inputs. Memory was not measured (no
  `/usr/bin/time` on the host).

## Remaining limitations (explicit, not hidden work)

- Mutable/arbitrary fields, virtual helper returns, unknown parameters, stream
  callback pipelines (`map`, `filter`, …), `doPrivilegedWithCombiner`.
- Single application class loader assumption for recognized loaders; containers and
  nested-jar launchers are not modeled.
- JPMS: full module resolution/readability (assumes `ALL-MODULE-PATH`), qualified
  exports, non-boot layers, exact dispatch after `provider()`, module variant in the
  model key and CVE F3 snapshots.
- Incomplete model facts are warnings; they do not yet downgrade CVE verdicts (B0).
