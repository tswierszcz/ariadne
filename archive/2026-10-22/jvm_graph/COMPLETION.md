# JVM_GRAPH (G0–G5) — completion record

**Archived:** 2026-10-22. **Status: DONE**, with one explicitly open item (see
"Remaining limitations" below) — not hidden by this archiving.

This area's two live documents (`JVM_GRAPH.MD`, the functional spec, and
`JVM_GRAPH_IMPLEMENTATION_PLAN.MD`, the phase-by-phase plan) are archived
in this directory as `JVM_GRAPH.MD` and `JVM_GRAPH_IMPLEMENTATION_PLAN.MD`
— their content as it stood for this entire area's work, moved out of
`RECOMMENDATION/` because every phase it names (G0–G5) shipped. The
detailed per-phase completion notes (scope, files touched, exact
commands run, measured before/after numbers) live in
`RECOMMENDATION/IMPLEMENTATION_STATUS.MD`'s "Rejestr faz" section under
`JVM_GRAPH:G0` through `JVM_GRAPH:G5`, which stays in place as the
historical record — this file is a short pointer and summary, not a
duplicate of that detail.

## What shipped

| Phase | What it did | Commit |
|---|---|---|
| G0 | Call-site/edge-evidence contract: `Call.PC`, `Edge.Basis`/`Sites`, surfaced in the report's entry-call line (kind/resolution/offset). | `83d2580` |
| G1 | Deterministic duplicate-class selection: merge order now follows explicit input order, never worker-finish order; structured `Hierarchy.Duplicates()` with content hashes. | `f542ac3` |
| G2 | Exact method resolution: CHA matches by descriptor, not name only (fixed real overload conflation); a direct call to an inherited, non-overridden method resolves to the actual declaring class instead of a dead-end node. | `c8ff41a` |
| G3 | Class-initialization edges: `new`/`getstatic`/`putstatic`/`invokestatic` → the target class's own `<clinit>`; a class's `<clinit>` → its superclass's `<clinit>`. | `684b9c4` |
| G4 | `-target-java N`: multi-release JAR variant selection per the JAR File Specification, with a documented fallback to base classes when unset. | `a9b9f12` |
| G5 | Integration/qualification: before/after graph comparison across G0–G4, a real memory measurement of G0's evidence storage, two stale README passages corrected, fixture sources added for G2's open item. | `f10eecf` |

## User-visible behavior now available

- The report's REACHABLE/FRAMEWORK entry-call line names the call kind,
  the resolution basis (`direct`/`cha`/`synthetic`), and the bytecode
  offset.
- CHA no longer conflates two overloads of the same method name — on
  real petclinic this roughly halved the CHA-added edge count.
- A call to an inherited method correctly continues the analysis
  through the class that actually supplies the body.
- The call graph includes class-initialization edges (a new "class-init"
  figure in the report, alongside "direct"/"added by CHA").
- `-target-java N` resolves multi-release JAR variants; without it,
  behavior is unchanged from before this entire area's work.

## Verified

Every phase above shipped with `go build`/`go vet`/`go test ./...`
green, a real end-to-end `bin/ariadne -offline` run against petclinic
and WebGoat (not just unit tests), and `go test -race` on every touched
package — G5's own qualification pass ran `go test -race ./...` for the
**whole project**. Measured, not assumed, effects include:
- G2 roughly halved CHA-added edges on petclinic (630816 → 310418) and
  fixed one confirmed false positive (`CVE-2021-22118`,
  FRAMEWORK → NOT_REACHABLE — a CHA artifact through Spring WebFlux's
  reactive multipart reader, which petclinic's classic Spring MVC stack
  never uses).
- G4 activated on 15 genuinely multi-release dependency jars in
  petclinic and 21 in WebGoat, confirmed by reading their manifests
  directly, not assumed from file names.
- G5's memory measurement: G0's per-edge evidence (`Sites`) costs
  ~32.7 MiB on petclinic's full graph, ~18% of the ~177 MiB total heap
  after building it — real, not dominant.
- `results.json` stays byte-identical across repeated runs throughout,
  including with `-target-java` set.

## Remaining limitations (not resolved by this work, not hidden either)

- **G2's bridge-method and default-method resolution is verified only
  against a hand-built hierarchy** (`hierarchy.NewForTesting`), not real
  compiled bytecode: this working environment has no JDK (`javac` is
  unavailable; only `openjdk-21-jre`/`openjdk-25-jre` packages are
  installed). The exact fixture sources needed to close this gap are
  already written (`testdata/fixture/src/fixture/Box.java`,
  `StringBox.java`, `BridgeDemo.java`, and a default method added to
  `Greeter.java`) and documented in `scripts/build-fixtures.sh`'s own
  comment — a future session with a JDK should run that script, then
  add the cross-check `BridgeDemo.java`'s own doc comment describes.
- Default-method superinterface eager-initialization (JVMS 5.5's
  narrower rule) and reflection-driven initialization (`Class.forName`)
  are not modeled — out of this area's scope, not attempted.
- `docs/SUPPORT.MD`'s reflection/callback rows remain `Not implemented`,
  unchanged by any of this work (checked explicitly in G5, per this
  phase's own warning against overclaiming).

## Where to pick this back up

If the bridge/default-method gap above gets closed (a session with a
JDK available), that's a small, scoped follow-up — not a reason to
reopen this whole area. Otherwise, per `RECOMMENDATION/ROADMAP.MD`, this
area's completion unblocks **R1** (`DATA_FLOW:D0–D2`,
`SERVICE_LOADING:S0–S2`, `ASYNC_CALLBACKS:A0–A2`), which is the next
natural direction for this project.

## Addendum (2026-11-07): G2's bridge/default-method gap closed

Closed by GRAPH_PRECISION P0 with a JDK now available (OpenJDK 25): instead of the
`testdata/fixture` sources named above, a dedicated fixture `testdata/callgraph` is
compiled with the real javac and run on the real JVM (`scripts/callgraph-oracle.sh`),
recording which body actually runs at each call site for every receiver the fixture can
create — including a default method kept and overridden, a generic bridge and a
covariant-return bridge. `internal/callgraph` `TestKnownTargets` finds every body the JVM
ran, with three explained extras (the two compiler bridges, a super call's helper) and no CHA
edge to an undeclared method. Details: `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, entry
"GRAPH_PRECISION:P0".
