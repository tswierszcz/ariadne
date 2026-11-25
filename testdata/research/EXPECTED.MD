# Research interface control library (RESEARCH_INTERFACE RI1–RI3, RESEARCH_HARNESS RH-A2)

Written 2026-11-24 with the first `ariadne research index` / `research query`. The contract is
[docs/research-interface/CONTRACT.md](../../docs/research-interface/CONTRACT.md). Built by
`scripts/research-fixture.sh` (jars and `observed.txt` are committed; tests need no JDK).

`lib.jar` (sources in `lib/`) is the analysed library, indexed with `-roots api -profile generic`.
`deps/` holds two versions of one dependency class: `fx-dep-1.0.jar` (`depold/`) and
`fx-dep-2.0.jar` (`dep/`). Ariadne lists a dependency directory by file name and the first copy of
a class wins, so version 1.0's `fx.dep.Paths2` is selected and version 2.0's is a shadowed copy.
`client.jar` (`client/fx/client/Main.java`) is the JVM oracle: it calls the public API in a
temporary directory, on the same class-path order.

No case has an advisory; every target is an arbitrary method.

| Case | Target | JVM ran it | Expected query result |
|---|---|---|---|
| (a) private method behind public API | `Extractor#writeEntry(Path, byte[])` | yes | `matched`, `found`: entries `extract` and `save(String, byte[], Path)`, `jvm_rule` |
| (b) dead private method | `Extractor#unused(Path)` | no | `matched`, `not_found` |
| (c) model gap | `Extractor#plugin` calls `Class.forName(param).newInstance()` | — | one unbounded reflection gap, reachable: every matched target's model is `incomplete` |
| (d) overloads | `Extractor#save([B)` | yes | `found`, zero-edge path (the root is the target) |
| (e) template method | `Cleaner#wipe(Path)` through `Handler#process` → abstract `handle` → `DeletingHandler#handle` | yes | `found`, 3 entries; `process`'s path has a CHA (`heuristic`) step to a library class, `client_dispatch = false` |
| (f) dependency code | `fx.dep.Paths2#resolve` in the 1.0 instance | yes | `found` from `extract` |
| (g) selected copy, unused method | `Paths2#legacyResolve` (1.0 only) | no | `not_found` |
| (h) shadowed copy | `Paths2#safeResolve` (2.0 only) in the 2.0 instance | no | `shadowed`, `not_evaluated` |
| (i) absent method | `Extractor#absent()V` | — | `missing`, `not_evaluated` |
| (j) client dispatch | `Cleaner#purge`, called only by the package-private, never-created `PurgingSink#accept` (the library's one `Sink`) | no | `found` from `Drains#drain(Sink, Path)` and `Pipeline#run` with `client_dispatch = true`: every path needs a sink only a client can supply (LB2), and the client's sinks ran instead |
| (j) supplied dispatch | `fx.dep.Purger#call` (1.0), the dependency's own `Callback` | no | `found` from `Relay#send` with `supplied_dispatch = true` (the callback is a field only `Relay`'s caller sets); from `Drains#hand(Callback, Path)` with `supplied_dispatch = false`: a parameter of the entry itself (lb2-7) is not reflected, which the response states in `warnings` |
| (k) conditional callback | `Workers#recover`, registered with `exceptionally` next to `thenAccept(archive)` | no (the future completes normally) | `found` from `Workers#later`; `callees` of `later`: two callback edges, the `exceptional` condition only on `recover`'s site |
| JDK operations | `calls` with `declared_class_prefix` `java.nio.file.Files` | — | `write`, `delete`, `deleteIfExists`, each `body_in_input = false` |

`client_dispatch` is false on (e) because `DeletingHandler` is a library class the library itself
lets a client construct (LIBRARY_MODE LB2's rule); a class only a client could supply would make
that step client dispatch.

Cases (j) and (k) (2026-11-25) are checked by `TestResearchDispatchAndConditions`. Checked by `TestResearchIndexAndQuery` (`cmd/ariadne/research_test.go`): every method in
`observed.txt` that is a target is `found`, and an index built twice has the same `analysis_id`.
