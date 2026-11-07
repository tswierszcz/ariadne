# DESERIALIZATION (Z0–Z4) — completion record

**Archived:** 2026-11-07. **Status: DONE** for every phase's declared scope. The
`docs/SUPPORT.MD` row "Deserialization callback/gadget analysis" stays **Partial**:
the remaining limitations below are real and stated there.

This directory holds the area's two former live documents (`DESERIALIZATION.MD`,
`DESERIALIZATION_IMPLEMENTATION_PLAN.MD`). Detailed per-increment notes (scope, files,
commands, numbers) stay in `RECOMMENDATION/IMPLEMENTATION_STATUS.MD` (entries
"DESERIALIZATION:Z0+Z1", "Z2", "Z3a", "Z4", "Z3"). Current behavior: `docs/SUPPORT.MD`
(the row above and "Data flow and taint"), README ("Deserialization" paragraphs in the
call-graph section, "Request data reaching the entries"), FAQ, `docs/RESULTS_CONTRACT.MD`
(`deserialization` in v2).

## What shipped

| Phase | What it did | Commits / PR |
|---|---|---|
| Z0 | Contract: per-deserializer model catalog (component, exact entry signatures, type rule, callbacks with access rules, protections, source); five evidence kinds kept apart | `f16ea8f` (#4) |
| Z1 | Java serialization entries (`readObject`/`readUnshared`, `ObjectInput#readObject`) with "may run" edges to the callbacks of the input's concrete serializable classes; serializability through the input and a JDK base table; every entry an unbounded model gap | `f16ea8f` (#4) |
| Z2 | Policy per entry (`evaluated`/`unrestricted`/`default_allowlist`/`unknown`/`none_seen`/`version_unknown`): literal JEP 290 filters on the same stream; the XStream model by release (1.4.5/1.4.7/1.4.18 boundaries) and literal permissions; rejected classes get no edge | `87baaa6` (#5) |
| Z3a | Versioned catalog of five documented gadget chains with required classes, bytecode guards of fixed releases, JDK limits and unchecked conditions; identity through the component inventory; per-entry policy; states `candidate_chain`/`conditions_unresolved`/`candidate_blocked`/`classes_absent`/`no_entry` | `7c93784` (#6) |
| Z4 | Chains per copy of their classes (relocated embedded instances); v2 `deserialization` and `findings[].deserialization_chains`; snapshot catalog versions; real-JVM oracle and a source-documented CVE set | `8750728` (#7) |
| Z3 (data flow) | Each chain × entry carries the entry's DATA_FLOW D4 taint state as `flow`; chain-level `flow`/`flows_by_state`; report Flow column and CVE row note; catalog `z3-1` | `380b6d4` (#11), on top of DATA_FLOW D3 (#8) and D4 (#9) |

## Evidence

- **Unit and fixture tests:** `internal/deserialize` (models, sites, policies, chains,
  `TestChainFlow`), `testdata/deserialize` (`scripts/build-deserialize-fixture.sh`: XStream
  API stub + one entry per policy under four releases), `internal/resultsv2`,
  `internal/report`; `go test ./...`, `go vet ./...` pass at `e818150`.
- **Real-JVM oracle:** `testdata/deserialize/oracle` (`scripts/deserialize-oracle.sh`,
  `TestDeserializationOracle`): 0 model edges the JVM did not run; 2 declared gaps (record
  canonical constructor, a class of unknown serializability).
- **CVE set:** `testdata/deserialize/cves.yaml` with `QUALIFICATION.MD` — every checkable
  case agrees.
- **Control apps (run on OpenJDK 25):** `scripts/prepare-deserialization-demo.sh` (callbacks,
  filter), `scripts/prepare-xstream-demo.sh` (1.4.5/1.4.10/1.4.20 policies),
  `scripts/prepare-deserialization-chains-demo.sh` (vulnerable vs fixed builds, guards
  observed by round-tripping the guarded class; request-fed `demo.Web` entries: with
  `-taint` CC6/CC2-4/XStream `candidate_chain` + `flow_found` on the vulnerable build,
  blocked on the fixed one), `scripts/prepare-taint-demo.sh` (taint states per entry).
- **Real apps (offline):** petclinic and WebGoat verdicts changed only where Z1 made a
  callback reachable (CVE-2021-46877, CVE-2023-6378, CVE-2022-45868 — documented in the Z1
  entry); later phases verdict-neutral; v2 byte-identical over two runs at every phase.
  WebGoat's vulnerable-components lesson: XStream 1.4.5 `unrestricted`,
  `xstream:eventhandler` `candidate_chain`, with `-taint` `flow_found` from
  `@RequestParam payload`; the CVE-2013-7285 row links it.

## Remaining limitations (also in `docs/SUPPORT.MD`)

- The JDK's own callbacks and what they call on the objects read (`HashMap` → `hashCode`)
  and XStream's converters for JDK types are not modeled — every entry stays an unbounded
  model gap.
- Only literal policies in the entry's own method are evaluated; `resolveClass` overrides,
  `jdk.serialFilter`, filters or XStream instances built elsewhere, custom permissions stay
  `unknown` (edges kept).
- `validateObject`, record canonical constructors, `readObjectOverride` and other
  libraries' deserializers (Jackson polymorphic typing, SnakeYAML, Kryo…) are not modeled.
- Chains: a five-chain catalog, not a search; a blocked chain never clears the advisory.
- Data flow: only request data (DATA_FLOW D4 sources, `-taint` opt-in); fields, lambda
  callers and callers outside the application are `undetermined`; `no_flow_in_model` is not
  proof the bytes are trusted; `candidate_chain` + `flow_found` is not an exploit.
- The JDK serializability table is incomplete (classes beyond it are counted as unknown).

## Follow-up outside this area

DATA_FLOW:D5 (measurement and qualification of the taint mode, a precondition for turning
`-taint` on by default) is the natural next step for the data-flow side of these results.
