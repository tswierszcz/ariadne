# Ariadne research interface contract

Revision **1.0.0-draft.1**. Date 2026-11-24. Status **DRAFT — FIRST IMPLEMENTATION AVAILABLE** (`research index` / `research query`, RI1–RI3; deviations in [README](README.md#implementation)).

The words MUST, SHOULD and MAY express obligations of future conforming implementations. The schema and fixtures accompany this draft. Current product capabilities remain described in [SUPPORT.MD](../SUPPORT.MD).

## 1. Purpose and ownership

Ariadne owns the static JVM analysis, selected class definitions, graph, model diagnostics, API roots and path evidence. Tekmerion owns threat models, LLM hypotheses, source-level investigations, experiments and security-effect assertions. The interface exchanges facts tied to exact artifacts and configuration.

The first version accepts built artifacts. Indexing MUST use the same supported JVM analysis pipeline and rules as the CVE CLI. The research interface addresses any selected method independently of advisory availability. An artifact supplied as the input library retains its own method targets and origin.

A path represents a possibility in a declared graph. The existence of a bug, attacker control of an argument, satisfaction of branch predicates and observation of a security effect are separate evidence supplied by Tekmerion. Conditions attached to modeled edges MUST remain available to that assessment.

## 2. Wire documents and schema

The schema namespace is `urn:ariadne:research:contract:1`. It identifies a local JSON Schema resource. Implementations MUST resolve its internal references from the shipped schema, without network access.

| Document | `schema_version` | Schema definition |
|---|---|---|
| Bundle manifest | `ariadne-research-bundle-1` | `Manifest` |
| Frozen analysis identity | `ariadne-research-identity-1` | `Identity` |
| Query | `ariadne-research-query-1` | `Request` |
| Success or error | `ariadne-research-response-1` | `Response` |
| JSONL record | Discriminated by `record_type` | `Record` |

Every top-level document except the request carries `contract_revision`. The request's schema version and bundle select its revision. Draft revision 1 accepts exactly `1.0.0-draft.1`.

Documents use UTF-8 JSON without a BOM. Duplicate object members MUST be rejected before schema validation. Numeric values MUST be integers in the supported schema ranges, with exact representation up to `2^53 - 1`. Strings MUST contain valid Unicode scalar values. Boolean and integer settings remain distinct. JSONL has one nonempty record per line and ends with LF. The producer MUST apply deterministic ordering and encoding across repeated identical analyses.

The named definitions in [contract.schema.json](v1/contract.schema.json) define the exhaustive fields, types and bounds. Objects reject unknown fields. Empty arrays explicitly represent known empty collections. `null` represents an unavailable or undetermined value as specified for that field. Required nullable fields MUST remain present.

Schema validation checks shape. A conforming implementation MUST also enforce the semantic invariants in this document, including cross-references, hash binding, result-state consistency and path continuity.

## 3. CLI lifecycle

These command names are reserved by the draft.

```sh
ariadne research index -app library.jar -deps dependencies -deps-list deps.txt \
  -roots api -profile auto -target-java 17 -out bundle

ariadne research query -bundle bundle -request request.json -out response.json
```

`index` consumes the provided built classpath and explicit configuration, runs all applicable models, and publishes a frozen bundle. It MUST work with networking unavailable. Options requiring advisory/LLM lookup or remote artifact resolution MUST be rejected. Existing static configuration with a selected classpath, profiles and Java target is recorded in the identity. Unknown options MUST be errors.

`query` validates and reads that bundle without rebuilding the graph or reading mutable external JARs. Each invocation accepts one request document. Reachability requests support batches of targets. A new root scenario, target Java, artifact or model configuration requires a new bundle.

Output files are published through a complete temporary output and atomic replacement on the same filesystem. Bundle publication fails on an existing destination or concurrent publisher. A failed publication MUST leave a previously valid destination readable. Interrupted temporary directories MUST be distinguishable from published bundles.

Human-readable diagnostics go to stderr. Machine-readable results are written to the requested output path. A validated domain result, including `missing`, `not_found` and `incomplete`, exits with code 0. Invalid arguments, schema, operation, filters, cursor, analysis binding or root selection exit with code 2. Bundle corruption/incompatibility, I/O failure and internal execution failure exit with code 1.

When the output path can be determined and written, a query failure MUST publish an error envelope. If parsing CLI arguments, interruption or the filesystem prevents publication, stderr and a nonzero exit remain the process evidence. A client MUST require an output document for the current invocation before consuming a result left at the same path by an earlier attempt.

## 4. Bundle layout and identity

The bundle contains `manifest.json`, its referenced identity document and one or more referenced JSONL datasets. A producer MAY partition records into multiple datasets. Dataset enumeration is complete in the identity.

`Manifest` requires these fields.

| Field | Meaning |
|---|---|
| `schema_version`, `contract_revision` | Format and contract revision |
| `analysis_id` | `sha256:` followed by the lowercase SHA-256 of the exact identity file bytes |
| `identity_path` | Bundle-relative identity file path |
| `created_at` | RFC 3339 timestamp for publication metadata |

`Identity` requires these groups.

| Field | Meaning |
|---|---|
| `provenance` | `kind` of `analysis` or `fixture`, producer and producer revision |
| `tool` | Ariadne version, nullable binary hash and model/algorithm version map |
| `configuration` | Root mode, graph mode, Java target, profile, active profiles, JPMS scenario, exclusions and additional effective settings |
| `inputs` | Input ids, roles, logical locations and content hashes |
| `classpath_order` | Ordered input ids that determine class selection |
| `datasets` | Relative path, exact byte hash, byte size, record count and encoding for every dataset |
| `capabilities` | Supported query operations for this bundle |
| `fixed_point` | `converged` or `incomplete`, with reasons |
| `coverage` | Declared limitations and disabled participants |

The Java target value 0 means the current base-variant fallback. A positive value names the effective Java feature target. Profile selection MUST be resolved to its effective identity. Additional settings include every analysis-affecting option not represented by the explicit fields. Output paths, time of publication and secrets are excluded from the identity.

Producers MUST write identical identity and dataset bytes for identical inputs, effective settings, versions and analysis content. Consumers verify hashes against the original bytes and MUST NOT reserialize JSON to calculate `analysis_id`. This avoids a cross-language canonical-JSON dependency. A changed publication timestamp alone leaves `analysis_id` unchanged.

Readers MUST reject missing files, hash/size/count mismatches, unsupported encoding, dangling references, duplicate record identifiers, duplicate input ids and repeated dataset paths. Bundle-relative paths MUST stay below the bundle directory. Readers MUST reject symlinks in the member path, traversal through `..` and absolute paths. Readers impose resource limits before allocating according to declared sizes or decoding records.

The analysis is frozen. An optional source checkout or later runtime trace requires explicit identity matching. Source bindings based only on a declaration are labeled `declared`.

## 5. Record catalogue

Every record uses a `record_type` discriminator and the schema definition in this table.

| Type / definition | Required facts |
|---|---|
| `component` / `Component` | Instance id, input role, identity status, nullable GAV, artifact hash and logical location |
| `class` / `Class` | Class id, instance, binary class name, class hash, selected flag, selected-copy reference and Java variant |
| `node` / `Node` | Node id/kind, nullable method identity, canonical node, receiver type, body kind, access flags and source location |
| `edge` / `Edge` | Edge id, endpoints, kind, basis, evidence character, rule version, sites and client/supplied-dispatch flags |
| `root` / `Root` | Root id, node, selection basis, inherited receiver classes and framework announcements |
| `gap` / `Gap` | Gap id, mechanism, reason, nullable site, indexed-root reachability, candidate methods/instances and rule version |
| `call` / `Call` | Call id/site, declared class/name/descriptor, target nodes and whether the declared body is in the input |

Ids are opaque ASCII tokens of at most 160 characters, stable inside and across identical analyses. Consumers MUST retain their associated `analysis_id`. Method identity includes instance, selected class hash, binary class name, JVM name and full descriptor. Multiple node specializations can share one method identity. For an external declaration with unavailable class bytes, instance, class hash and access flags may be null. Its opaque method id is bound to the external declaration in this analysis. Consumers can target that id directly. Origins and flags MUST NOT be fabricated to fill unavailable values.

A regular method node references itself as `canonical_node_id` and has a null receiver type. A receiver-specialized node references its canonical method node and names the specialized receiver. A synthetic event has a null method identity and `body = synthetic`. A synthetic event MUST NOT resolve as a real JVM method selector.

`body` distinguishes bytecode, native, abstract, external and synthetic nodes. An external node represents a known call target with an unavailable body. Model gaps MUST account for unmodeled external behavior relevant to downstream calls. Call-site discovery also lists declared external/JDK operations when their bodies are absent from the input.

A selected class references itself. A shadowed class references the selected copy of the same binary name. Origin and hash MUST follow the selected definition. A single flat selected class universe is the current model. Explicit JVM classloader isolation remains an assumption stated in coverage.

`Source` carries nullable file path, first/last line and source content hash, plus `binding` of `build_verified`, `declared` or `unavailable`. A filename or line table alone supports a location hint. `build_verified` requires evidence connecting those source bytes with the analyzed build. Line numbers are presentation metadata.

`Site` identifies caller node, bytecode PC, operation kind and nullable condition. Implicit/model sites use a documented anchor PC in the caller or an empty edge site list when no bytecode site applies. Consumers MUST retain every site and its condition. Edges with different rules or conditions retain their separate evidence. A representative path selects edge ids, with the rest of the bundle available for inspection.

Edge characters are `jvm_rule`, `library_model`, `input_declaration`, `heuristic` and `unclassified`. They express evidence origin. Rule versions, kind, basis and site conditions explain each step. Unclassified evidence MUST generate a warning.

## 6. Discovery requests

All requests require `schema_version`, `analysis_id`, `request_id` and `operation`. `request_id` is a caller-assigned correlation id and MUST be echoed. A response must match its request and bundle analysis id. A result cache is keyed by analysis, semantic request and query-rule versions. Request ids can change between equivalent requests.

| Operation | Result records | Filter interpretation |
|---|---|---|
| `entries` | Root records | Filters apply to the referenced node's method |
| `methods` | Real method/specialization nodes | Filters apply to that node |
| `calls` | Call records | Method filters apply to the caller, declared-class filter applies to the operation called |
| `callers` | Incoming edges of `node_id` | One graph hop |
| `callees` | Outgoing edges of `node_id` | One graph hop |

`entries`, `methods` and `calls` accept optional `filters`. `instance_id` matches an exact instance. `class_prefix` is a literal binary-name prefix, `method_name` and `descriptor` are exact matches. `node_ids` is an explicit node filter. `reachable_only` uses indexed-root graph reachability. `declared_class_prefix` is a literal prefix of the class named at a call site.

| Operation | Accepted filters |
|---|---|
| `entries` | `instance_id`, `class_prefix`, `method_name`, `descriptor`, `node_ids` |
| `methods` | Entry filters plus `reachable_only` |
| `calls` | Method filters on caller plus `declared_class_prefix` |

Filters outside the operation's accepted set MUST yield `UNSUPPORTED_FILTER`. An unknown node supplied for neighborhood lookup yields `INVALID_REQUEST` with its field location. Empty discovery results are successful.

Discovery results have `records` and `page`. Records are sorted by their opaque ids lexicographically. `returned_count` equals the array length. `total_count` is nullable when a complete count was not performed. `next_cursor` is nullable. `truncated` identifies omitted remaining results or incomplete work, and `work_complete` distinguishes a completed page from a work-budget cutoff.

A cursor binds the analysis, schema/rule revision, operation, filters, effective record limit and continuation position. Clients treat it as opaque and send the same query parameters on subsequent pages. A changed binding yields `INVALID_CURSOR`. A bounded query halted before a resumable continuation exists may return a null cursor with `truncated = true` and `work_complete = false`.

## 7. Reachability request and target resolution

`reachability` requires `targets` and `roots`. It accepts 1–256 targets with distinct caller-assigned target ids. A target contains exactly one of `method_id` or `selector`. A selector contains `instance_id`, binary `class`, JVM `name` and full `descriptor` including the return type. The parser MUST validate a real JVM descriptor in addition to schema shape validation.

Root selection is `indexed` or `subset` with distinct `node_ids` drawn from indexed roots. Arbitrary new roots yield `ROOTS_OUTSIDE_INDEX`. Root-sensitive pruning and model gates are frozen in the bundle. A subset query uses that existing graph and reports `root_basis = subset_on_indexed_graph`. It is a conservative search basis for the subset, subject to the recorded model gaps. Recomputing a model for another scenario requires indexing it.

| Resolution | Meaning and search behavior |
|---|---|
| `matched` | At least one selected node matches, including its receiver specializations |
| `missing` | No matching method is represented in the bundle, search is `not_evaluated` |
| `ambiguous` | Identity cannot be resolved uniquely under the declared model, search is `not_evaluated` |
| `shadowed` | Requested instance's class copy lost selection, search is `not_evaluated` |

`matched_node_ids` may include multiple specializations of the same resolved method. Ambiguity may expose candidate nodes for diagnosis. Missing and shadowed results have no matched nodes. Missing method lookup differs from an unknown request operation and is a successful domain result.

A represented external target can be reached at its invocation node. Its `body = external` remains visible, and no claim of body analysis follows. A declaration whose target has no represented node can still be inspected through `calls`; its method query is missing.

## 8. Result-state invariants

A reachability response contains one result per target in request order. `target_id` echoes the corresponding target id. The result has independent `resolution`, `search` and `model` assessments.

| Search state | Required facts |
|---|---|
| `found` | Matched target and at least one entry/path witness, even when further work or the model is incomplete |
| `not_found` | Matched target, complete search, no entries and `total_entry_count = 0` |
| `not_evaluated` | Unresolved target identity or an explicitly explained evaluation decision |
| `incomplete` | Search stopped before finding a witness, `work_complete = false`, null total count and a reason |

If a budget is exhausted after a witness was found, the state remains `found`, `work_complete = false` and the omitted/unknown entry coverage remains visible. An interrupted search MUST preserve valid witnesses already obtained. Non-found states have empty entry lists.

| Model state | Meaning |
|---|---|
| `complete_under_declared_model` | No recorded relevant gap, complete applicability of the declared analysis rules and `scope_uncertain = false` |
| `incomplete` | At least one recorded relevant limitation or model budget affects the target assessment |
| `undetermined` | Relevance/applicability cannot be assessed sufficiently, including an unresolved target |

The model object requires `relevant_gap_ids`, `global_gap_count`, `relevance_rule_version` and `scope_uncertain`. Search completeness concerns traversal of the available graph. Model completeness concerns the rules and missing behavior represented by that graph.

The first research relevance rule, `research-relevance-1`, counts a site when it is reachable from the query's selected roots and the target lies in the union of its supplied candidate method and component sets. A null set means that no bound on that axis was supplied. When either set is non-null, the supplied sets define the candidate union. Empty supplied sets mean a known empty union. When both sets are null, relevance is unbounded. Whole-analysis incomplete facts without a site apply globally. A case whose scope cannot be bounded sets `scope_uncertain = true`.

This rule MUST be evaluated against the query roots. Indexed-root reachability stored on a gap is a hint, not a substitute for subset traversal. `global_gap_count` counts bundle gaps once, while the per-target list counts unique relevant gaps. The relevance rule is separately versioned and its limitations MUST be disclosed. Recognized model restrictions remain present in bundle coverage even when the rule marks every target.

No-path plus an incomplete model is encoded as `search.state = not_found` and `model.state = incomplete`. Consumers preserve both. They may prioritize or report such candidates according to their own policy, with that policy separately identified.

## 9. Entry and path evidence

Each entry record contains `root_node_id`, traversal `pass`, `path`, evidence characters, `client_dispatch` and `supplied_dispatch`. `app_only` and `full` preserve Ariadne's framework traversal distinction. The character list is the union of the selected path edge characters. A zero-edge path starts at a root that is itself the target and has an empty character list.

The first query rule, `research-paths-1`, evaluates all selected indexed roots under the existing app-only/full semantics, preferring a path without CHA-only steps when one exists in the same pass. Within that preference it selects a shortest path. Ties use lexicographic node/edge ids. Returned entries sort by root node id. The producer MUST version changes to these rules. The character list summarizes the complete selected witness before display truncation.

`client_dispatch` and `supplied_dispatch` describe path dependence using the existing library-mode rules, including whether every relevant path from that entry requires the dispatch. They MUST be recomputed for the chosen entry/pass semantics and MUST NOT be inferred from one incidental edge flag alone. Edge flags expose site facts for review.

`path.node_ids` begins at the entry and, when complete, ends at a matched target. `edge_ids` has exactly one fewer element, in traversal order. Every referenced edge joins the corresponding adjacent nodes. Every shown node, edge and relevant gap MUST appear once in the response's `evidence` records. Supporting component and selected-class records MUST be included for the shown method nodes.

The response MUST include every relevant gap record in full. Evidence for distinct paths is deduplicated by record type and id. The full bundle remains the authoritative context. A response MUST NOT alter records copied from it.

A path longer than `max_path_nodes` may be returned as a prefix with `truncated = true`; the witness's search state remains `found`. `full_node_count` gives the complete count when known. A full path has its exact count. `total_entry_count` is the complete count of reaching entries when computed, otherwise null. `entries_truncated` signals an incomplete displayed or searched entry set.

## 10. Limits and deterministic work accounting

The server returns all effective limits, including defaults, in every successful response.

| Limit | Default | Allowed range | Scope |
|---|---|---|---|
| `max_records` | 100 | 1–1000 | Records in one discovery page |
| `max_entries` | 20 | 1–1000 | Shown entries per target |
| `max_path_nodes` | 100 | 1–10000 | Shown nodes per entry path |
| `max_work_units` | 1000000 | 1–1000000000 | Entire query |

A work unit is one graph node visitation or one adjacency-edge inspection during a traversal. The sum of per-target `work_units` MUST fit the query limit. Shared graph work is charged once to the earliest target using it, in request order. Discovery work uses the same bound while scanning relevant records. Cache hits consume the work needed for the current operation and preserve the same logical budget behavior as the uncached algorithm. Work accounting rules are versioned.

Targets are processed in request order. If the remaining budget is exhausted, later matched targets return `incomplete` with zero additional work and an explanation. Identity resolution and schema validation have their own implementation resource limits and MUST fail visibly if those limits are exceeded.

Record count, path length and entry display limits MUST NOT turn a found witness into a negative result. A work cutoff MUST NOT produce `not_found`. Hard process/memory cancellation can prevent a valid response and is then recorded by the caller as a failed/inconclusive attempt.

## 11. Error envelope

An error response has `status = error`, nullable correlation fields and `error` with `code`, human-readable `message`, field `details` and `retryable`. Locations are JSON Pointers or CLI option names. Messages are diagnostic text; clients branch on codes.

| Code | Meaning | Exit |
|---|---|---|
| `INVALID_REQUEST` | Malformed JSON, field/descriptor error, duplicate target id or invalid node reference | 2 |
| `UNSUPPORTED_SCHEMA` | Unsupported request version/revision | 2 |
| `UNSUPPORTED_OPERATION` | Operation unavailable in the bundle or implementation | 2 |
| `UNSUPPORTED_FILTER` | Filter unsupported for that operation | 2 |
| `ANALYSIS_MISMATCH` | Request/response binding differs from the bundle | 2 |
| `INVALID_CURSOR` | Invalid or differently bound continuation | 2 |
| `ROOTS_OUTSIDE_INDEX` | Query asks for a root outside the indexed set | 2 |
| `BUNDLE_CORRUPT` | Hash, record count, references or file layout fail validation | 1 |
| `INCOMPATIBLE_BUNDLE` | Reader cannot interpret the bundle version or required rule semantics | 1 |
| `IO_ERROR` | Filesystem operation failed | 1 |
| `INTERNAL_ERROR` | Unexpected producer/query failure | 1 |
| `CANCELLED` | A controlled cancellation produced a final error envelope | 1 |

Retryability is explicit and specific to the failure. A repeated schema error is terminal for the same request. A target-level missing result stays inside an `ok` response. A valid batch can include missing and matched targets. An invalid target declaration invalidates the request as a whole.

## 12. Versioning and compatibility

Draft consumers pin the exact contract revision and the contract-set file hashes. A `schema_version` suffix identifies a wire major. A stable major MAY add functionality through a negotiated capability/revision update. Strict readers MUST reject unsupported revisions and unknown fields. A client update and its contract tests precede changes to the pinned revision.

Changes to required fields, meanings, identity, state combinations or work accounting require a reviewed revision. Breaking stable changes require a new wire major. A producer lists its model and query algorithms, and clients retain those identities with evidence. A bundle reader must preserve the original semantics or reject the input.

The existing CVE result schemas, graph comparison exports and research documents have separate versions. The research bundle includes facts that the current graph-comparison export does not preserve. Converting between formats MUST retain required research evidence or report incompatibility.

Every coordinated release ships schema files, conformance cases, a fixture revision, checksums and a producer/consumer compatibility statement. The fixture provenance remains visible in tests and demonstrations.

## 13. Sources, hypothesis linkage and runtime observations

Tekmerion owns `run_id`, `hypothesis_id`, target descriptions, threat models, harness identities, assertion outcomes and reports. It maps each `request_id` and target id to these records and saves the unmodified Ariadne request/response.

The same method can be reached in a vulnerable implementation and a property-preserving control. Both path results remain valid. The experiment records which security assertion changed, the exact configuration and any code supplied by the harness.

Runtime correlation reuses [RUNTIME_TRACE.MD](../RUNTIME_TRACE.MD). `trace-config` and `correlate` are later RI6 operations and are absent from this revision's request schema/capabilities. The future extension must preserve scenario identity, content matching, loader context, sampling and measurement scope. An unobserved method or a trace entry alone leaves separate questions for the experiment's security assertion.

## 14. Conformance and status

A conforming producer and consumer MUST pass schema tests and semantic/integrity tests. The shipped [validator](v1/validate.mjs) exercises the data-only fixture and its declared positive/negative cases. Its scope is documentation conformance. It does not implement the research CLI, model query algorithms or a general verifier of arbitrary graph reachability.

Required implementation tests also include parser rejection of duplicate members, JVM descriptor validation, class conflicts, receiver specializations, conditional callbacks, isolated nodes, multiple roots, root subsets, multi-page discovery, changed cursor bindings, budget exhaustion after a witness, zero-edge paths, hash corruption, symlink/traversal rejection, interrupted publication and fresh-process bundle loading. Producer/consumer acceptance remains tracked in RI0–RI6 and TM0–TM6.
