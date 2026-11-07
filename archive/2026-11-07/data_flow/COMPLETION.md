# DATA_FLOW (D0–D5) — completion record

**Archived:** 2026-11-07. **Status: DONE** for every phase's declared scope. The
`docs/SUPPORT.MD` row "Data flow and taint" stays **Partial**: the remaining limitations
below are real and stated there.

This directory holds the area's two former live documents (`DATA_FLOW.MD`,
`DATA_FLOW_IMPLEMENTATION_PLAN.MD`) and the per-phase record `DATA_FLOW_D4.md`. Detailed
per-increment notes (scope, files, commands, numbers) stay in
`RECOMMENDATION/IMPLEMENTATION_STATUS.MD` (entries "DATA_FLOW:D0" … "DATA_FLOW:D5").
Current behavior: `docs/SUPPORT.MD` (the row above), `docs/TAINT_QUALIFICATION.MD`,
README (dependency summaries, "Request data reaching the entries"), FAQ,
`docs/RESULTS_CONTRACT.MD`.

## What shipped

| Phase | What it did | Commits / PR |
|---|---|---|
| D0 | IR and value-domain contract (`internal/valueflow`) | `3f0195d` |
| D1 | Control-flow graph and method-code semantics (`classfile.BuildCFG`, local ops) | `43267a9` |
| D2 | Width-correct abstract interpreter producing local value facts (`internal/dataflow.Analyze`) — the engine behind SERVICE_LOADING, ASYNC_CALLBACKS, DYNAMIC_LINKAGE, REFLECTION | `85cec3f` |
| D3 (values) | Five interprocedural slices: static fields, final instance fields, call results with a shared field/call chase, parameters from callers, final fields from constructor parameters; `doPrivileged` result summary | `98d1114`, `42a1649`, `1e9eda5`, `76f39ba`, `0b20992`, `709b7f4` |
| D3 (dependencies) | Dependency summaries (`internal/flowsum`, `d3-1`): per-call-site substitution, conservative policy for unknown calls, SCC fixed point; first consumer: what a deserialization entry's data depends on | `d683dd7` (#8) |
| D4 | Taint pilot (`internal/taint`, `taint-1`, opt-in `-taint`): request sources → data read by deserialization entries; states, paths, model-specific guards, v2 and report | `c97724f` (#9); phase record `DATA_FLOW_D4.md` (#10) |
| D5 | Qualification: fixture shapes, real-JVM canary oracle, measurement, `docs/TAINT_QUALIFICATION.MD`; `-taint` stays opt-in | `69d06e4` (#13) |

## Evidence

- **Tests:** `internal/dataflow`, `internal/valueflow`, `internal/flowsum`
  (`testdata/flowsum`, `scripts/build-flowsum-fixture.sh`), `internal/taint`
  (`testdata/taint`, `scripts/build-taint-fixture.sh`; `TestTaintOracle`, `TestShapes`),
  with mutation checks recorded per slice; `go test ./...`, `go vet ./...` pass at the
  D5 merge.
- **Real-JVM oracle (D5):** `scripts/taint-oracle.sh`, `testdata/taint/oracle` — 21
  entries, no observed flow reported as "no flow", no unconfirmed `flow_found`, four
  observed flows reported `undetermined` (depth limit, instance field, lambda caller,
  `Throwable#getMessage`).
- **Control apps:** `scripts/prepare-dataflow-demo.sh` (D3 dependency shapes),
  `scripts/prepare-taint-demo.sh` (taint states), `scripts/prepare-deserialization-chains-demo.sh`
  (request-fed entries with gadget chains).
- **Real apps (offline):** petclinic and WebGoat — each slice's measured effect is in its
  status entry (several value slices honestly zero new edges; parameters-from-callers and
  final-from-constructor raised resolved ServiceLoader sites); dependency summaries
  1252 / 2420 methods in ~1.7 s / ~6.7 s; `-taint` costs under 1 ms of search; verdicts
  unchanged by D3 (dependencies) to D5; v2 byte-identical over two runs.

## Remaining limitations (also in `docs/SUPPORT.MD`)

- Values: arithmetic/conversion results are unknown; non-final instance fields need real
  points-to; virtual-call return values are not summarized.
- Dependencies: no control/implicit flow; fields per field, not per object; a small JDK
  catalog (`ByteArrayOutputStream#toByteArray`, `Throwable#getMessage` and most of the JDK are
  conservative).
- Taint: one sink family (deserialization entries), a small source catalog, callers in
  application classes only (8 levels), fields/lambdas/unmodeled calls are `undetermined`,
  method-level paths without in-method transformation steps; opt-in.
- Heap/context sensitivity beyond per-call-site substitution and thread interleavings are
  later stages that need their own measurement (GRAPH_PRECISION owns points-to).
