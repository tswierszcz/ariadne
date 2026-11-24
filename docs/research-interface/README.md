# Ariadne research interface

Contract revision **1.0.0-draft.1**, documented 2026-11-24. The schemas, example bundle, conformance cases and documentation validator form the design package (RI0). A first implementation of `ariadne research index` and `ariadne research query` is available (RI1–RI3, 2026-11-24); its scope and known deviations are listed under [Implementation](#implementation) below.

This interface lets [Tekmerion](https://github.com/tswierszcz/tekmerion) discover JVM API entries and call sites, then ask whether selected methods can be reached in a frozen analysis. Targets may belong to the input library or its dependencies and can be assessed without an advisory.

- [Normative contract](CONTRACT.md) defines identity, requests, results, evidence, error handling and compatibility.
- [Integration and implementation guide](INTEGRATION.md) defines producer/consumer responsibilities, hand-off steps and acceptance tests.
- [JSON Schema](v1/contract.schema.json) contains the named definitions `Manifest`, `Identity`, `Request`, `Response` and `Record`.
- [Example manifest](v1/fixtures/bundle/manifest.json), [identity](v1/fixtures/bundle/identity.json) and [graph records](v1/fixtures/bundle/records.jsonl) form a hash-verified data-only bundle.
- [Conformance suite](v1/cases.json) identifies positive and deliberately invalid fixtures.
- [Validator](v1/validate.mjs) checks schema conformance, bundle hashes, graph references and the documented fixture semantics.

The validator uses Node.js with AJV 8 and `ajv-formats` available to CommonJS module resolution. Run it from any directory.

```sh
node docs/research-interface/v1/validate.mjs
```

The shipped fixture has three method nodes. Public `extract` calls private `writeEntry`. Private `unused` has no path from the API root. A reachable unresolved reflection site makes the model incomplete. The example source/JAR/class hashes are synthetic markers identified by `provenance.kind = fixture`. The directory provides contract data for client development.

| Example | Request | Response | Meaning |
|---|---|---|---|
| Positive path | [request](v1/fixtures/path-found-request.json) | [response](v1/fixtures/path-found-response.json) | Path found with bytecode evidence and a separately reported model gap |
| Negative graph search | [request](v1/fixtures/no-path-request.json) | [response](v1/fixtures/no-path-response.json) | Complete search finds no path, while the model remains incomplete |
| Missing target | [request](v1/fixtures/target-missing-request.json) | [response](v1/fixtures/target-missing-response.json) | Missing method is explicitly unevaluated |
| Work limit | [request](v1/fixtures/budget-exhausted-request.json) | [response](v1/fixtures/budget-exhausted-response.json) | Incomplete search retains a visible budget reason |
| API discovery | [request](v1/fixtures/entries-request.json) | [response](v1/fixtures/entries-response.json) | Public API root returned with count and pagination metadata |

The complete suite also includes an [error envelope](v1/fixtures/error-response.json) and rejected examples for unknown fields, conflicting selectors, contradictory target/path states, broken path shape and an analysis-identity mismatch.

The [implementation plan](../../RECOMMENDATION/RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD) records phases RI0–RI6. This design package delivers the specification portion of RI0. Consumer acceptance (TM0) remains work in Tekmerion.

## Implementation

```sh
go build -o bin/ariadne ./cmd/ariadne
bin/ariadne research index -app library.jar -deps deps/ -deps-list deps.txt -roots api -profile generic -out bundle
bin/ariadne research query -bundle bundle -request request.json -out response.json
node scripts/research-check.mjs bundle response.json   # schema check of a real bundle and responses
```

`research index` accepts the static flags of the CVE scan (`-app`, `-deps`, `-deps-list`, `-roots` — required —, `-profile`, `-profile-file`, `-target-java`, `-active-profiles`, `-jpms-scenario`, `-jpms-module-path`, `-api-exclude`, `-library-pom`, `-callgraph-mode`, `-embedded-components`, `-sbom`, `-cache-dir`) and rejects every other flag. It runs the same pipeline as the CVE scan up to the point where advisories would be read, always offline: Maven Central reference data (embedded components, the library's pom) comes only from `-cache-dir` and `ARIADNE_MAVEN_LOCAL`. Diagnostics go to stderr. The bundle is published atomically; an existing destination is refused.

`research query` reads one request and writes one response (or an error envelope) atomically. It reads and verifies in full — hash, record count, strict record shape, references — the datasets its operation needs, and checks the others for presence and size only.

Rules implemented: `research-paths-1` (app-only pass before the full pass, framework instances listed in the identity setting `framework_instance_ids`; within a pass a path without CHA-only steps first; shortest; ties by id), `research-relevance-1`, `research-work-1` (one shared forward search from the query roots, charged to the first matched target, then one reverse search per target and variant over the nodes the roots reach), `research-dispatch-1` (client dispatch by library mode's LB2 edge rule; supplied dispatch for caller-supplied field values).

Known deviations from the draft, also listed in each bundle's `coverage.limits`:

- Gap sites are anchored at their calling method (`pc` 0, kind `implicit`): Ariadne's gap rule records the method of an unresolved site, not its offset.
- `supplied_dispatch` is not set for a caller-supplied parameter of the entry method itself (lb2-7 marks those only on paths that start there).
- Source locations are LineNumberTable ranges with `path = null`; no source file is bound.
- Root `basis` is the selection mode (`public_api`, `inherited_api`, `announced_api`, `profile_entry`, `application_method`); web.xml and bean-lifecycle roots are not told apart.
- A bundle has one dataset per record type; there is no streaming reader, so a query holds the datasets it needs in memory.

Measured on 2026-11-24 (16 cores; see the plan's RI1–RI3 status for the full table): Petclinic `-roots profile` indexes in 61 s (peak 2.4 GB) into an 851 MB bundle of 257,640 nodes, 1,027,757 edges and 821,366 calls; WebGoat in 92 s (3.7 GB, 1.2 GB bundle), JPetStore in 15 s (199 MB). On Petclinic a discovery query takes about 3 s and a reachability query of 100 targets about 15 s, most of it loading.

