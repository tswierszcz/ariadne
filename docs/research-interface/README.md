# Ariadne research interface

Contract revision **1.0.0-draft.1**, documented 2026-11-24. The CLI described here is planned. The schemas, example bundle, conformance cases and documentation validator are available as a design package.

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

The [implementation plan](../../RECOMMENDATION/RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD) records phases RI0–RI6. This design package delivers the specification portion of RI0. Runtime producer and consumer acceptance remain implementation work in the respective repositories.
