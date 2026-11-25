# Producer acceptance of the research interface (RH-A0)

Contract revision **1.0.0-draft.1**. Stage RH-A0 of the [Recon → Hunt → Validate plan](../../RECOMMENDATION/RESEARCH_HARNESS_IMPLEMENTATION_PLAN.MD), 2026-11-25.

This page gives a consumer three things:

- real producer output to test against,
- deliberately broken variants of that output, each with the rule it breaks,
- the producer's own verdict on any response.

The plan accepts the contract when the producer and the consumer reach the same verdict on every case in this package.

The [v1 conformance package](v1/) is unchanged: the files and hashes the consumer pins stay the same. The acceptance package lives next to it, in [acceptance/](acceptance/), with its own `SHA256SUMS`.

## 1. The package

The generator is [scripts/research-acceptance.sh](../../scripts/research-acceptance.sh), which runs [cmd/research-acceptance](../../cmd/research-acceptance/main.go). It builds `bin/ariadne` from a committed tree and indexes the control library [testdata/research](../../testdata/research/EXPECTED.MD) offline, with an empty cache, using `-roots api -profile generic`. There is one bundle per scope:

| Directory | Scope | Records |
|---|---|---|
| `bundle-all` | `-scope all` | 184 |
| `bundle-reachable` | `-scope reachable` | 172 |

It then answers every request in [requests.json](../../testdata/research/acceptance/requests.json) with that binary.

| Request | Bundles | What it covers |
|---|---|---|
| `reach-cases` | all, reachable | The EXPECTED.MD cases a–i: a private method behind the API, a dead method, a model gap, overloads, a template method with a CHA step, dependency code, a selected copy's unused method, a shadowed copy and an absent method. Targets are given by selector and by method id. In the reachable bundle, the dependency method nothing reaches (`g-legacy`) resolves as `missing`. |
| `reach-subset` | all | Roots limited to `Extractor#extract` (`subset_on_indexed_graph`). |
| `reach-truncated-entries` | all | `max_entries 1`: one entry of several is shown and the total is kept. |
| `reach-truncated-paths` | all | `max_path_nodes 2`: a path is cut and `full_node_count` is kept. |
| `reach-budget-before` | all | The budget ends before any path: `incomplete`, `total_entry_count` null. |
| `reach-budget-after` | all | The smallest budget that yields a witness and then ends: `found` with an incomplete search, total null, entries truncated. |
| `entries`, `entries-filtered` | all (+ reachable) | Every root, and a `class_prefix` filter. |
| `methods-page1`, `methods-page2` | all | A truncated page and the next one through its cursor. |
| `calls-files` | all | JDK calls (`body_in_input = false`). |
| `callers-write`, `callees-extract` | all | One-hop neighborhood. |
| `reach-dispatch` | all, reachable | Cases j and k: client dispatch (`Drains#drain`, `Pipeline#run`), a caller-supplied field (`Relay#send`), a parameter of the entry itself (`Drains#hand`, lb2-7: false, stated in `warnings`), a conditional callback (`Workers#recover`). |
| `callees-later` | all | Case k: `Workers#later`'s two callback edges; the `exceptional` condition is only on `recover`'s site. |
| `error-unknown-node`, `error-analysis` | all | Error envelopes `INVALID_REQUEST` and `ANALYSIS_MISMATCH`. |

Every real response (`kind: response` in `cases.json`) satisfies three checks:

- it passes `ariadne research check` with no violation,
- it is byte-identical to what the producer answers on replay,
- it is valid against the v1 schema (`node scripts/research-check.mjs`).

Each mutation (`kind: mutation`) is a copy of one real response with a single defect. It names the rule that defect breaks (`expected_rules`), and the generator refuses to write a mutation that the check does not reject under that rule.

Envelope mutations are generated once per kind of response (reachability, discovery, error): a duplicate or unknown member, another analysis, request id, producer revision or contract revision. The remaining mutations target the content.

| Mutation | Rule | Defect |
|---|---|---|
| `missing-with-path` | `resolution` | A found target is declared `missing` and keeps its path |
| `found-without-entries`, `not-found-to-found` | `search-state` | `found` without path evidence |
| `unknown-count-zeroed` | `entry-count` | An unknown `total_entry_count` (null) is turned into 0 |
| `budget-claimed-complete` | `entry-count` | A search the budget ended is reported complete |
| `work-over-budget` | `limits` | Work units exceed `max_work_units` |
| `model-complete-with-gaps` | `model` | `complete_under_declared_model` with relevant gaps |
| `path-gap`, `path-edge-swapped` | `path-continuity` | A step is removed, or an edge joins other nodes |
| `path-cut-silently` | `path-end` | The last step is removed and the path is still marked complete |
| `path-wrong-root` | `path-root` | The entry's root is not the first node of the path |
| `character-dropped` | `path-character` | An edge character is missing from the entry |
| `pass-unknown` | `path-pass` | A pass other than `app_only`/`full` |
| `client-dispatch-flipped`, `supplied-dispatch-flipped` | `path-dispatch` | An entry's dispatch flag inverted |
| `evidence-modified`, `evidence-dropped`, `evidence-duplicated` | `evidence-original`, `evidence-missing`, `evidence-extra` | An evidence record is changed, missing or repeated |
| `targets-reordered`, `root-basis` | `target-order`, `root-basis` | Results are out of request order, or the root basis is wrong |
| `record-modified`, `record-dropped`, `cursor-on-last-page`, `truncated-without-cursor` | `record-original`, `page` | Discovery records or pages are inconsistent |
| `unrelated-edge` | `neighbor` | A `callers`/`callees` edge does not touch the node |

`go test ./internal/research` (`TestAcceptancePackage`) checks the hashes in `SHA256SUMS`, every verdict and every replay. A failing replay means a query now answers differently. In that case, regenerate the package with `scripts/research-acceptance.sh` and state the change in the PR, because the consumer compares verdicts on this package.

## 2. `ariadne research check`

```sh
ariadne research check -bundle BUNDLE -request REQUEST.json -response RESPONSE.json [-replay]
```

The command loads the whole bundle with every dataset verified. It writes the verdict to stdout:

```json
{"valid": false, "violations": [{"rule": "path-continuity", "location": "/results/0/entries/0/path/edge_ids/0", "detail": "…"}], "replay": "skipped"}
```

| Exit | Meaning |
|---|---|
| 0 | Valid. With `-replay`, also byte-identical to this producer's answer. |
| 1 | A rule is broken, or `-replay` gives `differs`. |
| 2 | An invalid argument or an unreadable bundle. |

`valid` uses only rules a consumer holding the same bundle can apply itself (§3), so a consumer's adapter can compare its verdict with this one. `replay` uses this producer's query code, which the consumer does not have, so it stays a separate field.

The check is semantic. JSON Schema is checked separately, with `scripts/research-check.mjs` or the consumer's own validator.

The same rules also reproduce the semantic verdicts of the v1 conformance cases on the v1 fixture bundle (`TestCheckContractCases`): the rejected `invalid-missing-with-path`, `invalid-path-shape` and `invalid-analysis-binding`, and the accepted positive cases. The fixture responses carry evidence records no result refers to. The contract allows this, so unreferenced evidence is not a violation, and only replay pins the producer's exact evidence set.

## 3. Rules

| Rule | Checks |
|---|---|
| `json` | No duplicate member at any depth, valid UTF-8, no BOM, no trailing data |
| `shape` | Exactly the contract's members at every level of the envelope, results, paths, pages and errors: nothing unknown, nothing missing, `null` where the contract writes null. Records are compared whole under the `*-original` rules. |
| `version` | `schema_version` and `contract_revision` |
| `binding` | `analysis_id` is the bundle's, `provenance` is the bundle identity's, and `request_id`, `operation` and the request's analysis match |
| `error-envelope` | An error has a code and a message, and echoes the request id |
| `page` | `returned_count` equals the number of records, a cursor exists exactly when the page is truncated, `total_count` ≥ returned, and records are in ascending id order |
| `record-type` | Discovery record types per operation: `entries` root, `methods` node, `calls` call, `callers`/`callees` edge |
| `record-original`, `evidence-original` | Each record is the bundle's record with that id, member for member |
| `evidence-missing` | Every matched or path node, its component and class records, every path edge and every relevant gap is in the evidence |
| `evidence-extra` | No evidence record appears twice |
| `neighbor` | A `callers` edge enters the node; a `callees` edge leaves it |
| `target-order` | One result per target, in request order |
| `resolution` | The resolution and matched nodes are what the bundle gives: by method id or selector, the selected copy, plus the receiver specializations; `shadowed` when only a losing copy has the class; `missing` outside a reachable scope |
| `search-state` | `not_evaluated` only for unresolved targets; `found` needs entries; `not_found` is complete with no entries; `incomplete` has no entries, an incomplete search and a reason |
| `entry-count` | `not_found` has total 0. A complete `found` has total ≥ entries shown, with `entries_truncated` exactly when fewer are shown. An incomplete search has total **null**, `entries_truncated` true and a reason. |
| `model` | `global_gap_count` equals the bundle's gap count, the relevance rule version is right, relevant gaps exist, `complete_under_declared_model` has no gap and a certain scope, `incomplete` has a gap, and an unresolved target is `undetermined` |
| `limits` | The effective limits are the requested ones; entries, path nodes and the sum of work units stay within them |
| `root-basis` | `indexed_graph`, or `subset_on_indexed_graph` for a root subset |
| `path-root` | The entry's root is an indexed root (inside the requested subset), the first node of its path, and has one entry per root |
| `path-continuity` | Nodes and edges exist in the bundle, there is one edge fewer than nodes, and each edge runs from its node to the next |
| `path-end` | An untruncated path ends at a matched node, passes none before its end, and has `full_node_count` equal to its length. A truncated path has exactly `max_path_nodes` nodes, fewer than `full_node_count`. |
| `path-character` | `character` is the ordered set (`jvm_rule`, `library_model`, `input_declaration`, `heuristic`, `unclassified`) of the path's edge characters. A truncated path's shown edges are among them. |
| `path-pass` | `app_only` or `full`; an `app_only` path does not expand a framework instance's node (identity setting `framework_instance_ids`) |
| `path-dispatch` | After a complete search, an entry's `client_dispatch` (`supplied_dispatch`) is true exactly when no path from its root to a matched node avoids client (supplied) dispatch edges (CONTRACT.md §9: every path decides it, not the shown one) |
| `replay` | `-replay` only: byte-identical to this producer's answer |

The control library cannot break two of these rules, so the package exercises them only partly:

- `record-type`: a changed type is reported as `record-original`.
- `error-envelope`: the error mutations break `shape`.

Until 2026-11-25 (RH-A2), `path-dispatch` required the flag whenever the *shown* path took a dispatch edge. That reading contradicts CONTRACT.md §9 and would have rejected a correct entry whose other path avoids the edge. Nothing exercised it then, because the control library had no dispatch edge. It now recomputes the flag over every path.

## 4. Verification on demand

`research query` verifies a bundle at two depths:

- Every dataset is checked for presence and size.
- The datasets the operation reads are verified in full: hash, record count, the exact members of every record and nested object, valid UTF-8, and every cross-reference among them.

| Operation | Datasets read |
|---|---|
| `entries` | components, nodes, roots |
| `methods`, `callers`, `callees` | components, nodes, edges, roots |
| `calls` | the above + calls |
| `reachability` | components, classes, nodes, edges, roots, gaps |

As a result, a damaged dataset an operation does not read goes unnoticed by that operation. `-verify full` verifies every dataset in full before answering. `research check` always does. Use `-verify full` for acceptance runs and for the first query on a bundle that was copied or stored. The cost is reading every dataset once (Petclinic `-scope all`: the calls dataset, 821,366 records, is the largest addition). The test `TestResearchCheckAndVerify` damages `calls.jsonl`: `entries` still answers by default, and `-verify full` fails with a bundle error.

## 5. Evidence for an independent validator (RH-A2)

A validator can verify a hypothesis's path on its own from three things:

- the bundle,
- the request (rerun it with `research query`; `research check -replay` confirms the answer is the producer's),
- the response's evidence.

| Plan item | Where it stands |
|---|---|
| 1. Everything needed to verify the path | `evidence-missing` requires every matched and path node with its component and class, every path edge and every relevant gap. `path-continuity` and `path-end` tie the path to the bundle. |
| 2. Kinds of step kept apart | Each edge has `basis` (`direct`, `cha`, `model`, `callback`, …), `character` (`jvm_rule`, `library_model`, `input_declaration`, `heuristic`) and per-site `kind` and `condition`. Case k: two callbacks of one method are separate edges, and only `recover`'s site has `exceptional`. |
| 3. Client/supplied dispatch | Entry flags are recomputed over every path (rule `path-dispatch`). Case j covers client dispatch, a supplied field and a supplied parameter. The parameter case (lb2-7) is not in the bundle, so every reachability response of a `-roots api` bundle states it in `warnings`. |
| 4. Graph basis of a root subset | The identity records `graph_mode`, the participants with their versions, and the fixed point. A subset query on an RTA/PTA graph warns that the graph was pruned for all indexed roots (conservative for the subset). |
| 5. Reproducible request | Replay is byte-identical. The ordering and limit rules are versioned (`research-paths-1`, `research-relevance-1`, `research-work-1`, `research-dispatch-1`). |

The four acceptance cases in the plan:

| Case | In the package |
|---|---|
| Path confirmed without CHA | `reach-cases`, `a-write`: `extract` → `writeEntry`, `jvm_rule` only |
| Path that depends on a heuristic | `reach-cases`, `e-wipe`: a CHA step to `DeletingHandler#handle`, `heuristic`; `reach-dispatch`, `j-purge`: client dispatch |
| No path while a relevant gap exists | `reach-cases`, `b-unused`: `not_found` with the model `incomplete` (reachable unresolved reflection) |
| Target absent from a limited bundle | `reach-cases.reachable`, `g-legacy`: `missing`, the reason points to `-scope all` |

Every case keeps all of its axes: resolution, search state, model state, entry count, characters and dispatch flags.

## 6. Pilot control target (RH-A3)

[testdata/pilot](../../testdata/pilot/EXPECTED.MD) holds an archive unpacker in two variants:

- **defective**: writes outside its directory on `../` entries,
- **fixed**: refuses such entries.

Both have the same API and the same write method. Run under a JVM oracle, they execute the same methods and differ only in the file-system effect. Ariadne gives identical path evidence for the write method in both variants. Only the experiment's property assertion tells them apart, and the guard shows up in `calls`. The controls are a dead method, unresolved reflection and an API that is allowed to write anywhere. A consumer indexes `defective.jar` and `fixed.jar` itself with `research index -roots api -profile generic`.

## 7. Producer/consumer matrix

| Producer (Ariadne) | Contract | Conformance package `v1/` | Acceptance package | Consumer (Tekmerion) |
|---|---|---|---|---|
| `research index`/`query` since #151 (RI1–RI3); `-scope reachable` #152; source paths #153; `prepare` #154 | `1.0.0-draft.1` | Unchanged since `d08678c` (#150) | — | `contracts/ariadne/PIN.json`: upstream `d08678c`, `v1/` files pinned by hash |
| RH-A0 (#155): `research check`, `-verify full`, `acceptance/` | `1.0.0-draft.1` | Unchanged | `acceptance/cases.json`; its producer revision and bundles' `analysis_id` are recorded in it, and file hashes in `SHA256SUMS` | To pin `acceptance/` by `SHA256SUMS` hash next to `v1/` (RH-T0) |
| RH-A2 (this change): cases j/k, `path-dispatch` recomputed, lb2-7 and pruned-subset `warnings` | `1.0.0-draft.1` | Unchanged | Regenerated: new control library, new `analysis_id`s, `warnings` on library-mode reachability responses | Re-pin `acceptance/`; `warnings` are new strings, no new members |

Change rules:

- A change to members or semantics is a new contract revision. It needs new `v1/` files (or `v2/`), a migration example and consumer tests before the consumer moves its pin.
- A producer change that alters an answer without changing the contract is visible as a replay failure. It regenerates `acceptance/`. The consumer re-pins and re-runs its verdict comparison.
- A new rule or mutation that only adds cases is backward compatible. The consumer must still reject every mutation and accept every response.
