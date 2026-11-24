# Producer and consumer integration guide

Contract revision 1.0.0-draft.1. The contract package is implemented as documentation, schema and fixture data. Commands in this guide describe the planned CLI.

## 1. Files and source of truth

Ariadne owns `docs/research-interface/CONTRACT.md` and `v1/`. Tekmerion vendors the exact `v1/` directory under `contracts/ariadne/v1/` together with a `PIN.json` next to it. Consumer guidance belongs to Tekmerion's `docs/ARIADNE_CONTRACT.md`.

The pin records the upstream repository, upstream path, contract revision, package status and SHA-256 of every vendored file. Until a source commit contains the contract package, the pin names its state as `workspace-draft` and uses content hashes for integrity. An integration release MUST replace that state with a committed upstream revision. A source implementation commit alone cannot identify uncommitted contract content.

Any revision to authoritative schemas or fixtures updates the whole vendored package and pin in one consumer change. Tests MUST validate the copied schema locally. References MUST NOT fetch mutable remote schema content.

## 2. Minimal consumer workflow

1. Record the repository revision, build recipe, artifact/classpath hashes and threat model in the Tekmerion manifest.
2. Invoke `research index` once per artifact/configuration/root scenario.
3. Validate the bundle format, hashes, dataset references and declared capabilities.
4. Discover `entries`, `methods` and `calls`. Expand source context through `callers` and `callees`.
5. Generate and validate candidate hypotheses. Match cited method identities to the bundle.
6. Submit exact targets in a `reachability` batch. Save raw input and response plus invocation exit and diagnostics.
7. Bind method/source locations to any additional data-flow evidence. Preserve declarations and source-binding uncertainty.
8. Build an API-level experiment. Record target observation and the property assertion separately.
9. Review or refute the hypothesis, save the decision and make fresh-run reproduction possible.

A consumer MUST check exit status, output freshness, schema, operation, request id, analysis id and target order before accepting results. It MUST retain model incompleteness and work truncation when converting the response into its own ledger.

## 3. Fixture-backed development

The sample bundle describes one synthetic library and has no executable target artifacts. It supports schema/adapter development and example response playback. A fixture adapter identifies itself with `provenance.kind = fixture` and selects a recorded response for the exact recorded request. An unknown request receives an explicit fixture-adapter error.

Fixture playback MUST NOT invent responses for new targets or label the data as a real analysis. A report made entirely from fixture responses states that provenance. Later RI1–RI3 replace the adapter path with actual indexing and query execution for integration tests.

Run conformance checks in either repository, with Node.js, AJV 8 and `ajv-formats` available.

```sh
# Ariadne
node docs/research-interface/v1/validate.mjs

# Tekmerion
node contracts/ariadne/v1/validate.mjs
```

The validator checks expected schema and semantic outcomes from `cases.json`. Cases expected to fail must remain invalid. Positive examples demonstrate a found path with a model gap, no path with the same gap, a missing target, work-budget exhaustion and API enumeration.

## 4. First implementation increments

| Increment | Ariadne producer | Tekmerion consumer | Acceptance artifact |
|---|---|---|---|
| Contract hand-off | RI0 schemas, fixture bundle and state rules | TM0 schema validation, pin and fixture adapter | Same conformance package passes in both repositories |
| Target preparation | RI1 shared static pipeline, RI2 bundle export | TM1 manifest, checkout/build identity and bundle adapter | Fresh-process verified bundle for local control library |
| Discovery and targets | RI3 reachability, needed RI4 catalog/context operations | TM2 hypothesis generation, TM3 static triage | Candidate starts without a CVE and maps to an exact target |
| Controlled experiment | RI5 fixture/pilot integration | TM4 harness, TM5 review and report | Public API test with observed property and control variant |
| Pilot | Completed RI4–RI5 | TM6 real library | Reproducible manifest, all candidate outcomes and cost |
| Runtime extension | RI6 trace filter and correlation | Follow-up TM4–TM5 | Observations tied to target hashes and experiment identity |

The contract specification portion of RI0 is delivered by this package. Runtime CLI support, producer regression checks and TM0's executable client/ledger remain planned. A completed documentation package alone does not satisfy runtime milestones.

## 5. Producer implementation review

Extract a shared static preparation boundary from the existing CLI while retaining the current CVE behavior. Reuse selected class/component identities, model coordination, library API roots, specialized receivers, BFS/path evidence and the gap vocabulary. Model passes currently run across several portions of `cmd/ariadne/main.go`; an early graph snapshot is insufficient.

Preserve information per site. The existing comparison export records edge identities and PCs, while research consumers also need site conditions, roots, isolated methods and ownership. Verify original versus round-tripped graph evidence and query results.

Track named query-rule versions for path preference, gap relevance and work accounting. Record target/body absence explicitly. Reuse existing library-mode client/supplied-dispatch rules in query entry assessments.

## 6. Consumer implementation review

Hypothesis generation needs source, documentation and an explicit property. The call catalogue supplies locations for inspection. Reaching a sensitive operation is a candidate-selection fact. The hypothesis specifies why the implementation may violate the property under attacker control.

Source-level analysis adapters identify their database/query/model versions and how source methods are matched to bytecode. The first property model may use CodeQL rather than extending Ariadne's narrow taint pilot. Unmatched source references retain a diagnostic and can trigger context repair.

Worker execution records caller configuration, dependency versions and the public entry used. Harness behavior must be part of the threat-model review. JUnit assertions and targeted fuzzing produce experiment artifacts. Model opinion, target invocation, a runtime exception and a reproduced security property each retain their evidence category.

## 7. Compatibility and publication checklist

- The authoritative revision and all vendored hashes agree.
- Both schema and deliberately invalid semantic cases pass their expected outcomes.
- The adapter preserves every state axis, evidence character, condition and truncation field.
- Required capability checks fail clearly for an incompatible binary/bundle.
- An actual index/query invocation works with networking unavailable.
- Producer regressions preserve documented CVE behavior, graph facts and root selection under the same inputs/settings.
- Consumer fresh-run experiments preserve source/build/bundle identity and security assertions.
- README and SUPPORT identify exactly which commands are implemented at the time of release.
- The producer/consumer release notes identify counterpart commits and any migration.

## 8. Updating the contract

Change the authoritative spec, schema, semantic rules and relevant fixtures together. Regenerate example hashes after modifying bundle data, run the conformance package, then update the consumer mirror and pin. A fixture content change receives a fixture revision. A meaning or field change receives a reviewed contract revision. Runtime-rule changes also update their algorithm identities.

The first stable version requires review of semantic edge cases listed in the normative contract and a real producer/consumer pilot. The draft remains usable for parallel development through explicit pinning.
