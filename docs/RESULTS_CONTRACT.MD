# Results contract (CVE:F0)

Status: **contract** — CVE:F0 of [archive/2026-11-23/cve/CVE_IMPLEMENTATION_PLAN.MD](../archive/2026-11-23/cve/CVE_IMPLEMENTATION_PLAN.MD),
shared with ROADMAP B0. It fixes what results *mean* before CVE:F1 changes how they are
computed. Nothing here is a new user-visible feature: today's `results.json` (v1) and
report are described as they are, and every planned change is marked **(F1)**. What the
tool currently supports is in [SUPPORT.MD](SUPPORT.MD).

## 1. Axes of a finding

A finding answers several independent questions. They are never folded into one
confidence number (B0), and an answer on one axis never stands in for another.

| Axis | Question | Values (F1 vocabulary) | Today (v1) |
|---|---|---|---|
| Component identity | Which artifact and component instance is this, and how was it identified? | `identified` (GAV from `pom.properties`/`deps.txt`, with evidence), `unidentified`, `conflict`, `version_candidates` (CVE:F6b: an embedded instance whose classes fit several published versions) | GAV or "unidentified jar"; since CVE:F1a an artifact/instance inventory with SHA-256, stable location and identity evidence (report "Components"); since CVE:F1b/F1c findings are per instance (report, `-json-format v2`) |
| Applicability | Does the advisory's affected range include this instance's version? | `applicability`: `affected`, `not_affected`, `undetermined` | OSV is queried with the exact GAV; an unidentified jar is never queried. CVE:F6b: a `version_candidates` instance is queried once per candidate version; its finding is `affected` only when every candidate is affected, otherwise `undetermined` |
| Mapping precision | Which methods does the advisory concern, and how sure is that? | method `role` (`vulnerable`, `entrypoint`, `sink`); `coverage` (`complete_under_declared_assumptions`, `partial`, `unknown`); `review` (`reviewed`, `unreviewed`) | `methods` list + `source` (`manual`, `diff`, `llm`) + `evidence` for LLM; no role, no completeness |
| Evidence source | Where did each fact come from? | JVM rule, library model, heuristic, input declaration, observation (B0) | edge `basis` on the entry edge; `source` on the mapping |
| Model completeness | Could a gap in the call-graph model hide a path relevant to this finding? | `complete`, `incomplete` (+ reasons), split into **global** and **relevant** gaps | global warnings only (e.g. "model analysis: …"); never per CVE |
| Reachability | Is a mapped method reachable from the selected roots in the declared model? | `REACHABLE`, `FRAMEWORK`, `NOT_REACHABLE`, `UNKNOWN`, `NOT_APPLICABLE`, and v2-only `not_assessed` | the `status` field (no `not_assessed`) |

## 2. Reachability states

`NOT_REACHABLE` means **no path in the declared model**, never "safe". Each state below
has the v1 meaning first and the F1 refinement, if any, after it.

| State | Meaning | Example |
|---|---|---|
| `REACHABLE` | A mapped method is reachable from the selected roots using only the app's own code up to the dependency call | WebGoat: XStream `fromXML` reached from a lesson endpoint |
| `FRAMEWORK` | Reachable only through framework/dependency code (no app→dependency edge on the shortest path that ends in app code) | petclinic: a Spring internal reaching a vulnerable Jackson method |
| `NOT_REACHABLE` | Mapped methods exist in the analyzed graph and none is reachable in the declared model. **(F1)** Not allowed when a *relevant* model gap exists (see §4) or when the mapping is `entrypoint`-only and the entry point is absent for an unrelated reason | petclinic: a vulnerable method of a dependency no root calls |
| `UNKNOWN` | No verdict: no method mapping, or an LLM advisory-only mapping that is not reachable (would otherwise read as "safe"). **(F1)** also: relevant model gap, `METHOD_NOT_FOUND`, `VERSION_UNKNOWN` | a CVE with no fix diff and no manual entry |
| `NOT_APPLICABLE` | Every affected dependency is a WebJar (client-side assets, no bytecode), or (2026-11-16) the manual table declares the advisory concerns no JVM code (reason `NOT_CODE_MAPPABLE`); the call graph cannot answer. Unchanged in F1 — in particular it does **not** mean "version not affected" (that is `applicability: not_affected`) | WebGoat: a jQuery advisory on a WebJar |
| `not_assessed` (v2 only) | Reachability was not evaluated for this finding: the version is not affected or undetermined, or there is no reachability model for it (e.g. a JIT/VM flaw in the runtime). Never exported to v1 as `NOT_REACHABLE` | a JDK CVE before CVE:F8 |

### "Method not found" vs "no path" (F0 requirement)

These are different facts and must stay distinguishable:

- **No path** — the mapped method is a node of the analyzed graph and no root reaches it.
  Correct `NOT_REACHABLE` (subject to §4).
- **Method not found** — no mapped method matches any node: the mapping names a
  method absent from this version, the class is in a jar ariadne did not read, or the
  selector is wrong. **v1 reports this as `NOT_REACHABLE`** (the matched set is simply
  empty). **(F1)** It becomes `UNKNOWN` with reason `METHOD_NOT_FOUND`.

## 3. Reason codes

Names follow the proposal in CVE_IMPLEMENTATION_PLAN.MD ("Wynik v2"), now stabilized
by F0. Findings without an advisory (an unidentified runtime, a library with no
identifiable version) go to the coverage section, never under an invented advisory ID.

Stable, machine-readable codes (`internal/findings`). Once emitted they are part of the
output contract: never renamed, never reused for another meaning; a superseded code is
deprecated. CVE:F0 defines them. Since CVE:F1b, `METHOD_NOT_FOUND`, `CLASSPATH_CONFLICT`,
`MAPPING_PARTIAL`, `MAPPING_ENTRYPOINT_ONLY` and `MODEL_INCOMPLETE` are emitted (v1 `reasons`);
CVE:F4 adds `VERSION_UNKNOWN`, `MAPPING_VERSION_UNCOVERED`, `MAPPING_CONFLICT`; CVE:F5 adds
`MAPPING_NEW_SIDE_ONLY` and emits `SOURCE_UNAVAILABLE`; CVE:F6b adds `VERSION_CANDIDATES`; CVE:F10 adds `NOT_CODE_MAPPABLE` (on `NOT_APPLICABLE` findings); `MAPPED_CODE_ABSENT` (2026-11-16, on `NOT_REACHABLE` findings, and on the v1 row when every finding carries it); `MAPPING_INITIALIZER_ONLY` (2026-11-17, mapping qualifier); `SUPPLIED_DISPATCH_ONLY` (2026-11-19, `-roots api`);
`RUNTIME_UNSUPPORTED` is not set yet.
`MAPPING_NEW_SIDE_ONLY`, like `MAPPING_PARTIAL`, qualifies the advisory row and stays off
instance findings (v2 carries it as `mapping.new_side_only`); so does `MAPPING_INITIALIZER_ONLY`
(v2 `mapping.initializer_only`).
`MAPPING_PARTIAL` on v1 qualifies `NOT_REACHABLE` without changing it (F1b, "option B"): every
mapping today is partial, so making it `UNKNOWN` would erase the state; v2 carries coverage as
its own axis.

| Code | Axis | Meaning | Example |
|---|---|---|---|
| `VERSION_UNKNOWN` | identity/version | The component's version could not be established, or (CVE:F4) is irregular (SNAPSHOT, non-numeric), so a version range cannot be checked | a jar without `pom.properties` and no `deps.txt` line |
| `VERSION_CANDIDATES` | identity/version | (CVE:F6b) The instance is an embedded copy whose classes fit several published versions equally (e.g. no class changed between them); the advisory affects at least one candidate, not necessarily the one present. Qualifies any verdict; with applicability `undetermined` unless every candidate is affected | `org.ow2.asm:asm-util` in jruby-core fits 9.0-beta and 9.0 |
| `CLASSPATH_CONFLICT` | identity/version | Several copies of a class are on the input; the selected one follows the declared classpath order, which may not match the real deployment | the same FQCN in two jars (`Hierarchy.Duplicates()`) |
| `MAPPING_PARTIAL` | mapping | The method list is known to cover only part of the vulnerable code | a diff mapping that touched one of several affected methods |
| `MAPPING_ENTRYPOINT_ONLY` | mapping | The mapped methods are entry points into the vulnerable behavior (e.g. a deserialization call), not the vulnerable code itself | XStream `fromXML` for gadget-chain CVEs |
| `METHOD_NOT_FOUND` | mapping | No mapped method exists in the analyzed graph | a method renamed in the analyzed version |
| `MAPPING_VERSION_UNCOVERED` | mapping | (CVE:F4) The mapping is versioned and none of its variants covers this instance's version: there is no method list for it | a mapping for `[1.0,2.0)` and `[2.0,3.0)`; the input has 3.1 |
| `MAPPING_CONFLICT` | mapping | (CVE:F4) Several variants of the mapping cover this instance's version with different methods; none is chosen | overlapping `[1.0,2.0]` and `[2.0,3.0)` for 2.0 |
| `MAPPING_NEW_SIDE_ONLY` | mapping | (CVE:F5) A fix-diff mapping was read on the post-fix side only — commit metadata unavailable, a merge or root commit, or the pre-fix file unreadable — so methods the fix removed are missing and helpers it added may be included | a fix commit with two parents |
| `MAPPING_INITIALIZER_ONLY` | mapping | (2026-11-17) A fix-diff or release-diff mapping names only static initializers (`<clinit>`) — the fix changed field initializers or static blocks and no method. A static initializer runs the first time the class is used, so reaching it means "the class was used", not that the changed data matters on that path; the verdict is kept. A curated family whose initializer is a data table guarding configuration-gated code (jackson-databind's polymorphic-typing blocklist) is instead mapped to the entry points that enable that code (role `entrypoint`, v2 `mapping.initializer_redirect`, row note) and does not carry this code | jackson-databind CVE-2020-25649: the fix hardened `DOMDeserializer`'s static `DocumentBuilderFactory` setup |
| `SOURCE_UNAVAILABLE` | mapping | The evidence needed for a mapping (fix commit, diff, advisory detail) could not be obtained. (CVE:F5) Emitted when the advisory has fix-commit references and none could be read (finding `UNKNOWN`); an advisory with no fix reference at all is a plain unmapped `UNKNOWN` | every fix-commit diff answers HTTP 500 |
| `MAPPING_MISSING` | mapping | (ROADMAP B0, 2026-11-13) No source produced any mapping for the advisory — manual table, fix diff, `release-diff`, LLM — so no verdict is possible; the `UNKNOWN` row says why instead of carrying no reason (before this code 20 of petclinic's 27 `UNKNOWN` rows had none) | an advisory with no fix commit and no class name |
| `NOT_CODE_MAPPABLE` | mapping | (CVE:F10, 2026-11-16) The manual table declares (`not_code`, with the reason) that the advisory concerns no JVM code of the artifact — an installer, a deployment setting — so no method can be mapped; the finding is `NOT_APPLICABLE`, like a WebJar, and the note carries the reason. Not a verdict that the app is safe | Tomcat's Windows installer (CVE-2025-49124) filed against `tomcat-embed-core` |
| `MAPPED_CODE_ABSENT` | mapping | (2026-11-16) No class the selected mapping names is on the analyzed classpath — the instance's version or the application does not ship that code (e.g. a Spring advisory mapped to `spring-webmvc` in an application without it) — so nothing can run it: the finding is `NOT_REACHABLE`, the note names the classes. A class that is present without the mapped method stays `METHOD_NOT_FOUND`. Assumes, like every `NOT_REACHABLE`, that the inputs are the deployed classpath | JPetStore (no `spring-webmvc`) and CVE-2018-1271 |
| `MODEL_INCOMPLETE` | model | An explicit gap in the call-graph model (unresolved reflection, ServiceLoader, callback, …) is relevant to this finding | an unresolved `Method#invoke` in the affected package |
| `SUPPLIED_DISPATCH_ONLY` | model | (`lb2-7`, `-roots api`) Every API method that reaches the code does so only through a call on an object the library's caller or container supplies (see "Caller-supplied objects"); the implementation that runs there is theirs. The verdict is kept | mybatis-spring-boot-autoconfigure: `MybatisAutoConfiguration#<init>` → `ObjectProvider#getIfAvailable` → (CHA) `YamlMapFactoryBean` → `Yaml#loadAll` |
| `RUNTIME_UNSUPPORTED` | runtime | The finding depends on runtime behavior ariadne does not model (JDK/JRE vulnerabilities, custom class loaders) | a CVE in the JDK itself |

## 4. Model gaps and verdicts (B0)

- A **global gap** is any incomplete-model fact of the run (today's "model analysis: …"
  warnings, unresolved reflection/callback/ServiceLoader sites, and — since
  DESERIALIZATION Z1 — every deserialization entry, kind `deserialization`, unbounded:
  the JDK's own serialization callbacks are not modeled; with `-callgraph-mode rta`
  (GRAPH_PRECISION P2) also every `Unsafe.allocateInstance` and
  `MethodHandles.Lookup#findConstructor` site, kind `allocation`, unbounded: RTA keeps only
  the fallback classes it can justify for them). DESERIALIZATION Z3a's chain
  states (`candidate_chain`, `conditions_unresolved`, `candidate_blocked`,
  `classes_absent`, `no_entry`) are report-only evidence of kind `chain-candidate`: they
  neither close nor narrow that gap and never change a verdict — a blocked chain rules
  out only that chain, not the advisory.
- A gap is **relevant** to a finding when the unresolved site is in a method that is
  **reachable from the roots** *and* its possible targets could include the affected
  component (its candidate owners lie in the affected artifact, or the site has no
  bounded candidate set at all). Both conditions are required: the CVE:F0 estimate
  ([testdata/baseline/VERDICT_CHANGE_ESTIMATE.MD](../testdata/baseline/VERDICT_CHANGE_ESTIMATE.MD))
  measured that "any unresolved site inside the affected jar" alone would cover 25 of 26
  (petclinic) and 17 of 18 (WebGoat) `NOT_REACHABLE` results — too broad to inform.
- **Rule (F1):** a finding with a relevant gap and no positive path is `UNKNOWN` with
  `MODEL_INCOMPLETE`, never `NOT_REACHABLE`. A positive path is reported regardless of
  gaps. Global gaps are reported once per run; relevant gaps per finding.
- The relevance rule is deliberately conservative and will be measured; it is not
  assumed to be right without the baseline comparison (§7).
- **Measured (CVE:F1b, 2026-11-05) and amended to report-only.** Gap sites: unresolved
  reflection `Method#invoke`/`Constructor#newInstance` and `MethodHandle#invoke`
  (unbounded), unresolved callback arguments (bounded by the callback interface's subtypes),
  unresolved ServiceLoader load/consumption sites (bounded by the registered providers).
  Reachable sites: 424 on petclinic (376 unbounded), 526 on WebGoat (456 unbounded). The rule
  as written applies to **every** `NOT_REACHABLE` row (19/19, 12/12); with unbounded sites
  ignored it still covers 18/19 and 11/12. A rule that covers every case carries no
  information, so in F1 a relevant gap **does not change the verdict**: the finding keeps
  `NOT_REACHABLE` and adds `MODEL_INCOMPLETE` with the gap counts by kind (v1 `model_gaps`).
  The `UNKNOWN` form of the rule waits until gap resolution (REFLECTION, DYNAMIC_LINKAGE,
  ASYNC_CALLBACKS) makes it selective; control case ARIADNE-TEST-0004 records that target.
- **Probe of rule variants (ROADMAP B0 b0-4, 2026-11-14, report-only).** Every run measures,
  per `NOT_REACHABLE` finding, the gaps four variants of the rule would count (v2
  `findings[].gap_probe`: `unbounded`, `bounded` with `bounded_by_kind`, `lambda`,
  `skipped_receiver`, `skipped_receiver_object`; `coverage.gap_probe`: how many findings each
  variant marks; report "Gap-rule probe"). `current` = the rule above; `scoped` = bounded
  sites only; `+lambda` adds reachable functional-interface calls left unresolved by
  DYNAMIC_LINKAGE L1 whose interface (or a subinterface) has a lambda created by a class of
  the instance; `+skipped` adds reachable direct virtual/interface calls on a
  `java.`/`javax.`/`jdk.`/`sun.` type whose expansion — the one CHA skips — would select a
  body in the instance (`-object`: excluding `java.lang.Object` methods). It changes no
  verdict, reason code or `model_gaps`; v1 does not carry it. Measured on the default runs
  (petclinic / WebGoat / JPetStore, 18 / 18 / 15 findings): `current` 18 / 18 / 15,
  `scoped` 17 / 15 / 14, `scoped+lambda` 17 / 16 / 14, both `+skipped` variants 18 / 18 / 15.
  Adding lambda or skipped-receiver gaps to the rule therefore changes **no** mark (the
  unbounded sites — 611 / 681 / 161 reachable — already mark everything), and the skipped
  receivers land in **every** instance (hundreds to thousands of calls each, also without
  `Object` methods): as a rule term they carry no information. The rule stays as it is.
- **Where the unbounded sites are (ROADMAP B0 b0-5, 2026-11-15, report-only).** v2
  `coverage.unbounded_sites` (`sites`, `callers`, `by_kind`, `by_component`, `top_callers`
  with `top_n`/`top_caller_sites`; report "Unbounded gap sites") groups the reachable sites
  with no bounded candidate set by the calling class's component and calling method. A
  site can only be bounded by a model of what its caller dispatches to; measured on the
  default runs they are a long tail — petclinic 611 sites in 477 distinct methods (40
  components), WebGoat 681 in 549 (55), JPetStore 161 in 129 (14); the 15 busiest methods
  hold 82 / 83 / 44. The reflective dispatchers ariadne already models elsewhere (Spring
  bean creation/lifecycle, `RepositoryMethodInvoker`, Spring ORM handlers,
  `InvocableHandlerMethod`, Stripes) hold 25 / 25 / 19: bounding them would leave every
  `NOT_REACHABLE` finding marked, so they stay unbounded and the rule stays report-only.
  Making it selective needs a general bound on unresolved reflection (REFLECTION R3/R4,
  GRAPH_PRECISION P3), not a catalog.

## 5. JSON compatibility

- **v1** (today's `-json` array of `{cve, status, methods, source, …}`) stays the default
  and keeps its meaning. Additive fields may appear; no field changes type or meaning.
- **v2 (F1, implemented by F1c)** is opt-in (`-json-format v2`): an object with `schema_version`, components,
  findings (advisory × component instance), coverage and a data-snapshot reference.
- **v1 export from v2 (F1):** one row per advisory. Aggregation: any positive finding
  wins (`REACHABLE` over `FRAMEWORK`); without one, an unresolved instance is never
  hidden under `NOT_REACHABLE` — the row becomes `UNKNOWN`; `NOT_APPLICABLE` only when
  every finding is (a WebJar, or `NOT_CODE_MAPPABLE`); a `NOT_REACHABLE` row carries
  `MAPPED_CODE_ABSENT` only when every finding does. Counts of unique advisories and of per-component findings
  are reported separately.
- **Writing (CVE:F9a):** the report, the results file and a snapshot manifest are
  written together at the end of a run, all or nothing; a run that fails writes none of
  them (an earlier run's files are left as they were).

### v2 schema (implemented by CVE:F1c, `schema_version: "2"`)

| Field | Content |
|---|---|
| `analysis_id` | (CVE:F9a) `sha256:` over the document's own JSON with this field empty: identifies the result, not the moment — the same inputs, data and build give the same ID. The report header and a `-snapshot-out` manifest of the same run carry it; the report also names the results file's SHA-256 (for v1 as well) |
| `tool` | `name`, `version` (the build) |
| `snapshot` | `kind: "inputs"`, `id` = SHA-256 over every artifact's stable location and content hash (independent of temp paths), `offline`, `classpath_order`, and (CVE:F3) `data_set` when the run recorded or replayed a data snapshot: `id` (the manifest's `data_set_id`), `replayed`, `config_matches` |
| `components[]` | one per component instance, inventory order: `id`, `artifact_id`, `location`, `role`, `sha256`, `bytes`, `gav`, `identity` (`identified`, `unidentified`, `conflict` = a shadowed class of this instance differs in content from the selected copy), `detection` (`pom.properties`, `deps-list`, `webjar`, `unidentified`, `app`, `embedded` — CVE:F6a, a library confirmed inside another jar; its `location` is `<host location>!/META-INF/maven/<g>/<a>`, `artifact_id`/`sha256` are the host's), `evidence[]`, `classes`, `shadowed`, `shadowed_differing`, `embedded_in` (the host instance of an embedded instance), `relocations[]` (CVE:F7a: `from` original package, `to` package in the jar, `source` `build-config: …pom.xml (maven-shade-plugin)` or `declared (-embedded-components)`), `version_candidates` (CVE:F6b: identity `version_candidates`, `gav` written `group:artifact:{v1,v2}`; an embedded instance named by build configuration or an SBOM has `location` `<host>!/build-config/<g>/<a>` or `<host>!/sbom/<g>/<a>`) |
| `advisories[]` | sorted by id: `aliases`, `summary`, `severity`, `affected[]`, `mapping` (null when no source mapped it): `methods`, `match_precision` (CVE:F4: `name`, `params`, `exact` or `mixed`, from the selectors), `source` (evidence kind), `role`, `coverage`, `review`, `variants` (versioned manual mappings), `method_roles` (LLM), LLM `model`/`evidence`; for a diff mapping (CVE:F5) `new_side_only` and `fix_commits[]`: `repo`, `sha`, `parent`, `base` (`parent` / `new_side_only` with `base_reason`), `ref_type` (the OSV reference type; GHSA Maven records label commit links `WEB`), `release_fit` (`candidates_present`, `no_candidate_present` — another commit's candidates are present — `no_candidates`, or absent), `release_tag`, `release_line` (CVE:F5c: `descends_from_release`, `in_release`, `other_line`, or absent), `release_line_reason` (why the line is unknown, or the off-branch-tag note for `descends_from_release`), `excluded` (`release_line` / `release_fit`: the commit's methods are left out of `methods`), `error`, `files[]` (`old_path`, `new_path`, `change`, `old_sha256`/`new_sha256`, `new_side_only`, `unavailable`, `methods[]`: `selector`, `kind` `changed`/`removed`/`added`, `source_params`, `bytecode` `exact`/`ambiguous`/`params_mismatch`/`method_missing`/`class_missing`, `possible_rename`) |
| `findings[]` | one per advisory × affected instance: `advisory_id`, `component_instance_id`, `identity`, `applicability` (`affected`: OSV is queried with the exact version) with `applicability_check` (CVE:F4: the OSV ranges evaluated by `internal/mavenversion` — `in`/`out`/`undetermined`/`none` — and whether that agrees), `match_precision` (of the selector a positive path reached), `reachability`, `reason_codes` (instance axis only — mapping role/coverage stay on `mapping`), `model_gaps`, `path`, `entry`, `note` |
| `coverage` | `global_gaps` (unresolved sites, reachable, unbounded, model-analysis warnings), `unidentified_components`, `embedded_candidates` (CVE:F6a: every further `pom.properties` of a jar — `host`, `gav`, `path`, `kind` `module`/`library`, `status` `confirmed`/`content_differs`/`partial`/`classes_absent`/`reference_unavailable`, `reason`, `reference_url`, `reference_sha256`, `reference_classes`, `present`, `identical`, `instance` when confirmed, `declared`, `relocations[]` — CVE:F7a adds statuses `collision` and `relocation_conflict`; CVE:F6b adds `version_candidates` (an instance with candidate versions is created) and `version_unknown` (classes present, no checked version fits or none could be checked: no instance), and the fields `source` (`pom.properties`/`declared`/`build-config`/`sbom`), `source_detail`, `version` (the version the classes confirm when it is not the declared one), `version_candidates`, `versions_checked[]` (`version`, `reference_sha256`, `reference_classes`, `present`, `identical`, `fits`, `declared`, `error`; present when other versions were compared), `versions_published`, `version_search` (how the compared set was chosen: up to 8 published versions nearest the declared one, or all when none is declared and at most 8 are published); `gav` is `group:artifact` when no version was declared; a library candidate that is not confirmed and not `version_candidates` has no instance and its advisories were not looked up), `embedded_declaration_issues` (CVE:F7a: `-embedded-components` entries not applied — no analyzed jar with that hash, or a hash mismatch for a named jar), `sbom` (CVE:F6b, with `-sbom`: `path`, `sha256`, `format`, `spec_version`, `serial`, counts `components`, `linked` — tied to an analyzed jar by SHA-256 —, `candidates` — nested components added as embedded candidates —, `unlinked` — Maven components tied to nothing, not used —, and `issues`), `advisories_without_instance` (with their whole-graph verdict), `counts` (advisories, findings, advisories by v1 status, findings by reachability), `data_freshness` (CVE:F2: `state` `complete`/`stale`/`unknown`, `ttl`, `max_data_age`, per-kind counts and oldest validation time), `withdrawn_advisories` (every source withdrawn; no findings), `snapshot_changes` (with `-snapshot-compare`: per advisory, why its derived result was invalidated and how its verdict changed) |
| `deserialization` | (DESERIALIZATION Z4, additive; absent when not rendered) `model_catalog` and `chain_catalog` versions; `entries[]` — `id` (`caller#name+desc@pc`), `call`, `model`, `version`, `app`, `policy` (Z2 state), `policy_detail`, `candidate_classes`, `rejected_classes`; `chains[]` — one per catalog chain and copy of its classes (`copy`: "" for the original names, else the instance whose relocations name them): `id`, `component_instance_id`, `format`, `state`, `advisories`, `classes[]` (`name`, `found`, `component_instance_id`, `confirmed`), `missing`, `blocks`, `unresolved`, `unchecked_conditions`, `flow` (DESERIALIZATION Z3: the strongest taint state among the chain's sites where it is not blocked — all sites when blocked everywhere — ordered `flow_found` > `undetermined` > `no_flow_in_model` > `not_assessed`; `not_analyzed` without `-taint`), `flows_by_state` (sites per taint state), `sites[]` (`entry`, `state`, the site's own `reasons`, `flow` — the entry's taint state or `not_analyzed`), `sites_by_state`; `summary` (`entries_by_policy`, `chains_by_state`) |
| `runtime` | (RUNTIME_NATIVE N1, additive; only with `-runtime-trace`) the imported trace ([RUNTIME_TRACE.MD](RUNTIME_TRACE.MD)): `import_version` (`n1-1`), `format`, `trace_sha256`, `run_id`, `scenario`, `started`/`ended`, `runtime`, `recorder`, `event_kinds`, `method_filters`, `sampling`, `complete`, `lost_events`, `truncated_stacks`, `incomplete_reasons`, `classes_by_status` (`matched`, `matched_artifact`, `different_version`, `name_only`, `not_in_input`), `classes_with_several_definitions`, `loaded_by_status`, `executed_by_basis_usable` (`traced`, `sampled`), `executed_not_usable`, `calls_by_status` (`in_graph`, `not_in_graph`, `unusable`), `handoffs`, `gap_cause_version` (`g1`) and `calls_not_in_graph_by_cause[]` (`cause`, `calls`, `top_details[]` — `detail`, e.g. the call's receiver type, and `calls` — most frequent first; causes in [RUNTIME_TRACE.MD](RUNTIME_TRACE.MD)). Evidence only: never part of `reachability`, and the absence of an event is never a negative |
| `deserialization.entries[].data_depends_on` … | (DATA_FLOW D3, additive) what the data the entry reads depends on within its method, as dependency-summary sources: `param:<i>`, `this`, `field:<owner>.<name>`, `call:<owner>#<name><desc>` (an unmodeled call's result), `call:*` (many); `data_flow_incomplete` (why it may be incomplete), `data_flow_widened` (calls of the method answered conservatively). Not taint: no statement that the data comes from outside |
| `deserialization.taint_catalog`, `deserialization.entries[].taint`, `deserialization.summary.entries_by_taint` | (DATA_FLOW D4, additive, only with `-taint`; absent otherwise) taint catalog version (`taint-1`); per entry: `state` (`flow_found`, `no_flow_in_model`, `undetermined`, `not_assessed`), `catalog`, `rule` (sink rule: `sink:deser-java-stream`, `sink:xstream-fromxml`), `flows[]` (≤ 5: `rule` — source rule `src:spring-mvc-param`, `src:spring-mvc-implicit`, `src:jaxrs-param`, `src:servlet-request` —, `source`, `steps[]` from the source to the entry: `method`, `input`, `call_pc` (-1 at the entry's method)), `flow_count`, `guard_state` (with a flow: `effective`, `partial`, `unknown`, `unrelated`, `none`), `guards[]` (`kind`: `policy`, `unrelated_sanitizer`, `control_check`; `effective`, `detail`, `method`, `pcs`), `gaps` (≤ 20) and `gap_count`, `outside_inputs`, `reason` (`not_assessed`). A separate flow assessment: never part of `reachability`, the v1 row or LLM evidence |
| `callgraph` | (GRAPH_PRECISION P0, additive) the call graph's settings — `mode` (`cha`, or `rta` with `-callgraph-mode rta`), `builder` (builder version), `skip_rule`, `unknown_allocation_policy` (`all-subtypes` for CHA, `cha-fallback-with-gap` for RTA), `roots`, `target_java` (omitted when unset) — its `nodes` and `edges`, and with `-graph-compare` a `comparison`: `config_differences` (always disclosed), `edges_before`/`edges_after`, `added_by_basis`/`removed_by_basis`, `evidence_changed`, `findings_changed[]` (`id`, `old_status`, `new_status`, `path_changed`, `because` — the added/removed edges on the paths), `findings_only_before`/`findings_only_after`, `types_added`/`types_removed` (P1, when both files carry created types). `created_types` (GRAPH_PRECISION P1): `catalog` (`p1-2`; `p1-1` before registered servlets/filters counted as `container`), `types`, `types_by_origin` (`new`, `service`, `reflection`, `model`, `container`, `deserialization`, `entry`), `types_only_modeled`, `lambda_sites`, `gaps_by_kind` (`reflection-unresolved`, `dynamic-proxy`, `unsafe-allocation`, `constructor-handle`, `service-unresolved`, `container-proxy`, `external-input`), `external_inputs`, `root_reference_parameters`, `cha_edges`, `cha_edges_to_classes_not_created` (informational, counted on the CHA candidate set). `rta` (GRAPH_PRECISION P2, only with `-callgraph-mode rta`): `version` (`p2-1`), `reachable_methods`, `live_classes`, `live_only_by_fallback`, `cha_edges_before`, `cha_edges_removed`, `cha_edges_kept_by_fallback`, `active_gaps` (unknown-allocation gaps whose site is reachable, by kind). Never part of `reachability` |
| `frameworks` | (FRAMEWORKS F5 follow-up, 2026-11-23, additive) the run's framework facts, until now only in the log and the report: `profile` (the ariadne profile), `rules_version` (`profile.RulesVersion`), `active_profiles` (as given by `-active-profiles`; omitted when not given — every condition is then `unknown`); `beans[]` — `class`, `name` (the container's name: explicit, else Spring's default), `explicit`, `source` (`stereotype`, `bean_method`, `repository`), `primary`, `qualifier`; `conditions[]` — one per `@Profile` usage: `class`, `method` (omitted for a class), `names`, `active` (`true`, `false`, `unknown`); `injections[]` — `class`, `member`, `param_index` (-1 for a field), `type`, `qualifier`, `value` (a `@Value` expression: never a bean lookup), `resolved` (the bean class the container would wire), `candidates`, `reason` (why `resolved` is empty); `entries[]` — external entry points: `class`, `method`, `basis` (HTTP mapping, `web.xml` servlet/filter/listener), `active`; `roots[]` — methods the profile made roots because the framework calls them reflectively: `class`, `method` (name+descriptor), `kind` (bean lifecycle: `constructor`, `injection method`, `init annotation`, …; Spring MVC bodies: `response body getter`, `request body constructor`, `request body setter`), `origin` (the bean or handler); `summary` (`beans`, `conditions_by_state`, `injections`, `injections_resolved`, `injections_value`, `injections_unresolved`, `entries_by_basis`, `roots_by_kind`). Lists sorted, deterministic. Facts, not findings: never part of `reachability` or the v1 row |
| `findings[].deserialization_chains` | (Z4, additive) catalog chains that name the finding's advisory (ID or alias) and whose classes belong to its instance: `chain`, `copy`, `state`, `flow` (Z3: the chain's `flow`). Evidence of kind `chain-candidate` only — never part of `reachability` or the v1 row |
| `findings[].runtime_observed` | (RUNTIME_NATIVE N1, additive; only with `-runtime-trace`) the finding's mapped vulnerable methods seen executing in the trace, on a class of the finding's instance matched by hash: `evidence` (the trace's record id), `method`, `loader`, `basis` (`traced` exact entries, `sampled` stack samples), `count`, `class_status`, `in_graph`. Execution is not exploitation; an absent or empty list says nothing — never part of `reachability` or the v1 row |
| `findings[].runtime_traced` | (RUNTIME_NATIVE N1, additive; only with `-runtime-trace`) `true` when the trace's exact-entry filter (`scope.method_filters`) covered a mapped method of this finding. Without a `traced` entry in `runtime_observed` it means "traced, not entered in this run" — still not evidence of unreachability; never part of `reachability` or the v1 row |
| `findings[].path_edges`, `findings[].character` | (ROADMAP B0/B1, 2026-11-13, additive; on findings with a path) `path_edges[]`: one per step of `path` — `from`, `to`, `character` (the step's strongest evidence), `evidence[]` with every edge between the two methods (`kind`, `basis`: `direct`, `init`, `cha`, `synthetic`, `serviceloader`, `callback`, `methodhandle`, `reflection`, `deserialization`, `sam`, `model`; `character`; `pcs`: call-site offsets). `character`: the set of the steps' characters — `jvm_rule` (a bytecode call or class initialization), `library_model` (a modeled API), `input_declaration` (a profile's declared edge), `heuristic` (a CHA expansion only), `unclassified`. Kinds of evidence, not a confidence score; a step a bytecode call also backs is `jvm_rule`. Since engine version 3 (ROADMAP B0 b0-2) the reported path has no CHA-only step whenever such a path exists in the same pass (REACHABLE: app-only, FRAMEWORK: full), even a longer one, so `heuristic` means reachability depends on a CHA expansion; the verdict is decided by the plain passes as before. Not in v1 |

`not_assessed` is defined but not emitted yet (no finding today lacks a reachability
model). **Export limits of v1:** a v1 row shows one path (the shortest over every
instance's code) and drops the per-instance verdicts, identities and coverage; reasons on
a v1 row mix both axes (`MAPPING_*` and instance codes). Every run builds the v2 document
and fails if its v1 export disagrees with the v1 rows (`resultsv2.CheckV1`).

## 6. Legacy mappings

The embedded manual table (`internal/vulnmap/vuln_methods.yaml`) predates roles and
completeness. Until an entry is confirmed by its source, F1 treats it as **partial** and,
where the audit marks it so, **entrypoint-only** — never as a complete mapping of the
vulnerable code. The per-entry audit is in
[testdata/cve-control/MANUAL_MAPPING_AUDIT.MD](../testdata/cve-control/MANUAL_MAPPING_AUDIT.MD).
Diff- and LLM-derived mappings keep their existing source labels.

**Migration to v2 (CVE:F9b).** Three entries are versioned (schema v2, exact selectors):
CVE-2022-1471 (CVE:F4), CVE-2021-44228 and CVE-2022-42889. The migration of the latter two
is documented in the audit (bytecode of every listed release checked; variants follow
where the declared methods change). Migration changes the precision of the selectors and
adds the version span; it does **not** change their role or coverage (`sink`/`entrypoint`,
`partial`) — a v2 entry is not a complete mapping by its format. The other 36 entries
(the XStream anchor) stay legacy, name-only.

## 7. Reference data

- **Control set:** [testdata/cve-control/](../testdata/cve-control/README.MD) — fictional
  advisories for mechanics and a source-documented real-CVE list for quality, covering
  all five CVE extensions.
- **Baseline:** [testdata/baseline/](../testdata/baseline/README.MD) — the current
  version's results on petclinic and WebGoat with hashes of every input and of every
  network response used. A baseline is a reference for change, **not ground truth**.

## 8. Library mode (LIBRARY_MODE LB0 contract, implemented by LB1 — 2026-11-16)

A library jar answers a different question than an application: whether anyone using its
public API can reach mapped vulnerable code in its dependencies, and through which API
methods ([RECOMMENDATION/LIBRARY_MODE.MD](../archive/2026-11-23/library_mode/LIBRARY_MODE.MD)). This
section fixed the contract before the implementation (LB0); LB1 implements it (`-roots
api`, `internal/libapi` version `lb1-1`, `-api-exclude`); LB2 (`lb2-1`, `lb2-2`, 2026-11-17) the `client_code` gap; LB3 (`lb3-1`, 2026-11-18) the dependency declarations; `lb2-6` (2026-11-18) CDI bean archives; `lb2-7` (2026-11-19) caller-supplied objects; `lb2-8` (2026-11-19) Spring-announcing libraries are framework-mediated.
Member-class accessibility follows the class's own access flag and its enclosing classes
(a `protected` member class is compiled `public` and counts as accessible).

**Selection.** Library mode is a root set, `-roots api`, next to `all` and `profile` (not a
separate `-mode` flag: the graph, the models and every other output stay as they are; only
the roots and one additive field change). The library jar is the `-app` input, its
dependencies are `-deps`/`-deps-list` as for an application. `callgraph.roots` is `api`.

**The public API** — the roots — is every method a user of the library can call or
override from outside its packages:

- the methods and constructors with `public` or `protected` access of every *accessible*
  class: a `public` top-level class, or a `public`/`protected` member class whose
  enclosing classes are all accessible (`InnerClasses` flags, not only the class's own);
  `protected` counts only in a class that is not `final` (a subclass can call it);
- the public `default` and `static` methods of public interfaces; abstract methods (no
  body) are roots that reach nothing;
- the class initializer (`<clinit>`) of every accessible class;
- (`lb2-3`, 2026-11-17) the methods an accessible class inherits from a dependency
  supertype: `public` (or `protected`, unless the class is `final`) instance methods with a
  body, declared in a dependency superclass or superinterface, to which a call on the API
  class resolves (not overridden on the way) — `YAMLMapper#readValue` is
  `ObjectMapper#readValue`. (`lb2-4`) The root is a copy of the dependency's method per
  inheriting API class: in it, a virtual/interface call on the method's own unchanged
  `this` goes only to the bodies that class and its subtypes select, and a call on `this`
  to another method the class inherits from a proper supertype goes to that method's copy
  for the class (transitively). Calls on any other receiver dispatch as in the graph. A
  copy whose bytecode the local-value analysis cannot read keeps every edge;
- the providers the library registers in its own `META-INF/services` (their public
  no-argument constructor and the service methods they implement) — a consumer's
  `ServiceLoader` creates them;
- (`lb2-5`, 2026-11-17) the classes the library announces to Spring in its own jar, whatever
  their access — every class of every key of `META-INF/spring.factories`, every class of a
  `META-INF/spring/*.imports` file (Spring Boot 2.7+; key = file name without `.imports`).
  An `EnableAutoConfiguration`/`AutoConfiguration` entry, or one annotated `@Configuration`
  (`@AutoConfiguration`, `@SpringBootConfiguration`), is a configuration class: its
  constructors and `<clinit>`, its `@Bean` methods (any access, static too), its methods
  annotated `@Autowired`/`@Value`/`@Inject`/`@Resource`/`@PostConstruct`/`@PreDestroy`/
  `@EventListener`, the public methods of container callbacks it implements
  (`InitializingBean`, `DisposableBean`, `ImportSelector`, `ImportBeanDefinitionRegistrar`,
  bean (factory) post-processors), its member classes annotated as configuration and its
  `@Import` targets, transitively. Any other `spring.factories` entry is a factory Spring
  instantiates and calls through the key's type: constructors, `<clinit>`, public instance
  methods. Conditions (`@ConditionalOn…`, `@Profile`) are not evaluated — an entry "may
  run". `-api-exclude` drops announced classes as it drops providers;
- (`lb2-6`, 2026-11-18) the beans of a CDI bean archive, whatever their access: a library jar
  with `META-INF/beans.xml` (or only a Jandex index, `META-INF/jandex.idx`, which Quarkus
  discovers as `annotated`). The discovery mode is beans.xml's `bean-discovery-mode`; without one
  an empty or pre-4.0 file is `all` (CDI 1.1–3.0's default; 4.0 made it `annotated`, so `all` is
  what some container may do) and a 4.0+ one `annotated`; `none` discovers nothing. In `all` mode
  a bean is every concrete class of the jar with a no-argument or `@Inject` constructor; in
  `annotated` mode a class with a bean-defining annotation (`@ApplicationScoped`,
  `@RequestScoped`, `@SessionScoped`, `@ConversationScoped`, `@Dependent`, `@Singleton` — Quarkus
  treats it as one —, `@Interceptor`, `@Decorator`, `javax.*` and `jakarta.*`, or a library
  annotation meta-annotated `@Stereotype`/`@NormalScope`/`@Scope`). A bean's entries: its
  constructors and `<clinit>`, every method annotated `@Inject`, `@Produces`, `@PostConstruct`,
  `@PreDestroy`, `@AroundInvoke`, `@AroundConstruct` or with an `@Observes`/`@ObservesAsync`/
  `@Disposes` parameter (any access, static too), and its public instance methods. Provenance
  (`announced_by`): `<file> bean-discovery-mode=<mode>[ @<annotation>]: <class>`; the beans are
  `announcements` and count in `announced_classes`. Alternatives, specialization, vetoes and
  extensions are not evaluated (a bean "may run");
- not roots (unless announced): `private` and package-private methods, members of inaccessible classes,
  synthetic and bridge methods (their targets are roots in their own right), lambda bodies.

Two scopes narrow it, both reported: with `-jpms-scenario module-path` and a library
`module-info`, only the packages the module `exports` (unqualified) are API; `-api-exclude
p1,p2` removes package prefixes the user calls internal (`…internal…` is public to the
JVM, so it is API by default). Kotlin `internal`, Java access through reflection
(`setAccessible`) and `Unsafe` are not access rules here.

**Spring-announcing libraries (`lb2-8`, 2026-11-19).** When the library announces classes to Spring
(`spring.factories`, `META-INF/spring/*.imports`, see `lb2-5` below), the framework package
patterns of the embedded `spring` profile that the selected profile lacks are added to the
framework matcher, whatever `-profile` selects: such a library runs inside a Spring container, and
what the container does is framework code, as for an application under `-profile spring`. The
added patterns are `coverage.api.framework_packages_added`. A CDI bean archive adds nothing.

**Verdicts.** `reachability` keeps its meaning with the API as the roots: `REACHABLE`
means "reachable from the public API without passing through framework code", not "run
by an application". `NOT_REACHABLE` means no API method reaches the mapped code in the
model. All reason codes keep their meaning.

**New fields (additive, v2 only).**

| Field | Content |
|---|---|
| `findings[].exposed_api[]` | with `-roots api`, on a `REACHABLE`/`FRAMEWORK` finding: every API method from which a mapped method of the finding is reached, sorted by key — `method` (`owner#name+desc`), `pass` (`app_only` or `full`, the BFS pass that reaches it), `character` (the set of its example path's step characters, as `findings[].character`), `path` (one shortest example path from that method; one without a client dispatch when there is one), `client_dispatch` (LB2; `true` only when every path in that pass is one — see "Client code"), `supplied_dispatch` (`lb2-7`; `true` only when every path in that pass goes through a call on a caller-supplied object — see "Caller-supplied objects"), `announced_by` (`lb2-5`; for an entry of an announced class, how it was announced: `<file> <key>: <class>`, then ` > member <class>` / ` > @Import <class>` per step), `inherited_by` (`lb2-3`; for a dependency method API classes inherit, those classes — `lb2-4`: only the classes whose receiver-specialized copy reaches the code; the entry's `pass`/`path` are the best among them). `path`/`path_edges` of the finding stay the overall shortest path. |
| `findings[].supplied_by` | (`lb3-1`, 2026-11-18) with `-roots api`: who supplies the finding's instance to a user of the library — `library`, `user` or `undeclared` (see "Dependency declarations"); absent for an instance the declarations do not cover. |
| `components[].declaration` | (`lb3-1`) with `-roots api`: the declaration matched to that instance, as in `coverage.api.dependencies.declared[]`. |
| `coverage.api.alternate_versions` | (`lb3-2`, 2026-11-23, additive) with `-roots api -alt-versions …`: `strategy` (`newest`), `mode` (`approx`), `rules_version` (`lb3-2`), `dependencies[]` — one per declared dependency with an identified instance: `package`, `supplied_by`, `instance`, `instance_version`, `declared`, `note` (why no version was chosen: version list unavailable, nothing fits), `versions[]` — `version`, `chosen_by` (`range_max`, `line_latest`), `mode`, `basis` (what the assessment rests on; `approx`: the other version's advisories and mapping variant, reachability on the analyzed version's code), `not_assessed` (why the version was not assessed; `findings` is then empty and no evidence), `findings[]` (`advisory`, `status`, `reasons`, `path`, `note`). Never part of `findings[]`, the v1 rows or a verdict |
| `coverage.api` | with `-roots api`: `dependencies` (`lb3-1`: `pom`, `library`, `declared[]`, `notes[]` — see "Dependency declarations"), `classes` (accessible classes), `methods` (root methods), `inherited_methods` (`lb2-3`: of those, dependency methods API classes inherit), `receiver_specialization` (`lb2-4`: `clones` — method copies, `this_sites` — copied call sites on `this`, `edges_dropped` — their edges to a body the API class does not select, `unanalyzed_methods` — copies whose edges were kept unchanged), `supplied` (`lb2-7`: `sites`, `field_sites`, `edges`, `fields`, `unanalyzed_methods` — see "Caller-supplied objects"), `framework_packages_added` (`lb2-8`: see "Spring-announcing libraries"), `announced_classes` (`lb2-5`: classes made entries from framework announcements, with member classes and `@Import`s), `announcements[]` (`file`, `key`, `class` — what the jar announces), `providers` (`META-INF/services` providers counted as roots), `excluded_packages` (from `-api-exclude`), `jpms` (`exports` when the module's exports decided, `classpath` otherwise), `exports[]` (the exported packages used). |

v1 rows are unchanged in meaning; a v1 row has no `exposed_api`.

**Client code (LB2, `lb2-1`/`lb2-2`, 2026-11-17).** A call the library makes on a method a client
can implement or override runs code the analysis does not have. A *client-code site* is a
virtual or interface call in a library method whose resolved method is declared in an
accessible, in-scope library type the client can subtype (an interface, or a class that is
not `final` and has a `public`/`protected` constructor), with `public` or `protected`
access, not `static`, not `final`. Calls on dependency and JDK types are not client-code
sites (callbacks there have their own mechanisms). Each site is a gap of mechanism
`client_code` (site code `client`; `coverage.gaps` with `count` and `reachable`). Its
candidate set is empty — the client's classes are not in the input, and client code
reaches the library only through its API, which is already the roots — so it never makes
a finding `MODEL_INCOMPLETE`. No edge is invented: such sites stay ordinary CHA/RTA call
sites. A *client dispatch* is a graph edge whose every bytecode site is a client-code site
and whose target is neither the called method's own body nor a body of a class (or a
subtype of it) whose objects exist without client code — one the input creates
(created-type catalog, any origin but an assumed entry) or one a client can construct or
subclass without overriding (an accessible, in-scope interface, a class with a public
constructor, a non-`final` class with a protected one; `lb2-2`): that body stands in for
the client's own code. An `exposed_api`
entry is `client_dispatch: true` when every path from it in its pass contains a client
dispatch; otherwise its example path avoids them. Not yet: vulnerable code reached only
after a client call returns is not marked conditional.

**Caller-supplied objects (LB2, `lb2-7`, 2026-11-19).** A library is often handed objects of its
dependencies' types — a parameter of an API method, a Spring `ObjectProvider` or `BeanFactory`
the container passes to an announced class. Which implementation of that type runs is chosen by
whoever supplies it; CHA/RTA dispatch to every implementation in the dependencies. A value is
*caller-supplied* in a library method when it is a parameter of a dependency type (a class or
interface outside the library and the JDK that a caller can subtype: an interface or a
non-`final` class) of that method and the method is a root — caller-supplied only on paths that
start at that root, since the library may call the same method with its own object — or when it is
read from a library instance field of a dependency type every write of which in the library stores
a parameter of a root that no library method calls (on every path). A *caller-supplied site* is a
virtual or interface call on a dependency type whose receiver is caller-supplied. Each site is a gap of mechanism `client_code` (site code `supplied`),
with an empty candidate set — never `MODEL_INCOMPLETE`; no edge is added or removed. An edge whose
every bytecode site is a caller-supplied site is a *supplied dispatch* (for a parameter site, only
as the first step of a path from its own method); an `exposed_api` entry is `supplied_dispatch:
true` when every path from it in its pass contains one (otherwise its example path avoids them), and a positive finding all of whose `exposed_api` entries are marked carries
`SUPPLIED_DISPATCH_ONLY` (v1: the row too, when every positive finding does). The verdict is
kept: the mark says the code that runs there is the caller's choice, not that the path is false.
Not followed: a caller-supplied value handed to a dependency method or stored inside a dependency
object and called there, values passed on between library methods. Verified on
[testdata/library-supplied/](../testdata/library-supplied/EXPECTED.MD) and, as the motivating
measurement, the JVM oracle of [testdata/library-starter/](../testdata/library-starter/EXPECTED.MD).

**Dependency declarations (LB3, `lb3-1`, 2026-11-18).** The versions analyzed are the
input's; a user of the library may get others, and some dependencies only exist if the user
adds them. With `-roots api` the library's own pom is read — the one embedded in the jar
(`META-INF/maven/<group>/<artifact>/pom.xml`; with several, the one whose artifactId names the
jar file), or `-library-pom` — with its parents and their properties and
`dependencyManagement` (versions and scopes, including BOMs it imports with `scope import`; the
pom's own and inherited entries win over imports, the first import over later ones), fetched
from Maven Central through the cache (a local Maven repository first with `ARIADNE_MAVEN_LOCAL`). The
tree a *consumer* of the library gets is built as Maven builds it: the library's `compile` and
`runtime` declarations that are not `optional`, then, transitively, every `compile`/`runtime`,
non-`optional` declaration of their poms (each followed at the version in the input when it
satisfies the declaration, else at the declared exact version; a range with no instance inside
it is not followed), nearest first, a declaration's `exclusions` applying below it, scopes
combined as Maven does (`runtime` above `compile` gives `runtime`), at most 6 levels. Every
dependency of that tree has `supplied_by: library`. The library's `provided`, `system` and
`optional` declarations, and what reaches the tree only through them, have `supplied_by: user`
(`provided` is not transitive; `optional` exists only when the user adds it). `test` scope is
not published and is ignored. An input dependency instance (identified, not embedded) the tree
does not contain is `undeclared`. Each `declared[]` entry: `package` (`group:artifact`),
`supplied_by`, `scope`, `optional`, `via` (the dependencies it is declared through, nearest the
library first; empty when direct), `declared` (the version as written), `version` (resolved:
exact version or range; empty when unresolved), `instance`/`instance_version` (the matching input
instance) and `version_match`: `declared` (the declared exact version), `differs`, `in_range`,
`out_of_range`, `undetermined` (the instance's version cannot be ordered), `unresolved` (the
declared version is not known), `unidentified` (no identified instance, but an unidentified
input jar is named `<artifact>-<version>.jar` — a jar without `pom.properties`; `instance` names
it, `-deps-list` identifies it), `absent` (no instance in the input). For `absent` and
`unidentified` the dependency's advisories are not assessed; a warning names each such
dependency supplied by the library. What cannot be read
(no pom, unavailable parent or dependency pom, unresolved property) is a note, never an error.
Declarations change no verdict and no reason code. Not yet: advisories for other versions a
user can get (the declared range, the newest release of the line) and `MAPPING_VERSION_UNCOVERED`;
Gradle module metadata; profiles in the library's pom. Instances embedded in another jar
(CVE:F6a) come with their host and carry no `supplied_by`.

**Control library:** [testdata/library/](../testdata/library/EXPECTED.MD) — cases (a)–(f)
with the expected roots, `exposed_api` and verdicts written before LB1, checked against
the real JVM (`scripts/library-oracle.sh`). Dependency declarations:
[testdata/library-deps/](../testdata/library-deps/EXPECTED.MD) — `supplied_by: library` equals
what Maven 3.9 resolves for a consumer of the library (`scripts/library-deps-oracle.sh`). CDI bean
archives: [testdata/library-cdi/](../testdata/library-cdi/EXPECTED.MD) — every exposed entry ran
on Weld SE 5.1.2 (`scripts/library-cdi-oracle.sh`), the non-bean never.

## Gap vocabulary

(ROADMAP B0 b0-3, 2026-11-13, `internal/gaps` version `b0-3`.) Every analyzer that reports a model gap keeps its own codes — they are part of its output and are never renamed — and each code maps to exactly one **mechanism**, so the same gap reads the same way wherever it is reported. Sources: `site` — the unresolved dynamic sites of the model-gap rule (§4, `model_gaps` keys); `allocation` — the created-type catalog's unknown allocations (GRAPH_PRECISION P1/P2); `lambda` — SAM-call resolution gaps (DYNAMIC_LINKAGE L1); `runtime` — causes of measured calls missing from the graph (RUNTIME_NATIVE N1); `coordinator` — incomplete model facts of the B3 coordinator. The mapping is a reading aid: it changes no verdict and no reason code; only `site` gaps take part in the model-gap rule. v2: `coverage.gaps[]` (`mechanism`, `source`, `detail` — the source's own code —, `count`, and for `site` `reachable`) and `findings[].gaps[]` (the finding's `model_gaps` in this form). `TestVocabularyDocumented` keeps this table and the code identical.

| Mechanism | Meaning | Source codes |
|---|---|---|
| `reflection` | Method#invoke, Constructor#newInstance, Class.forName with an unresolved value | site/`reflection`, allocation/`reflection-unresolved`, runtime/`reflection` |
| `method_handle` | MethodHandle#invoke*, Lookup#find* with an unresolved value | site/`methodhandle`, allocation/`constructor-handle`, runtime/`method-handle` |
| `callback_async` | a task or callback handed to a thread, executor or framework | site/`callback`, runtime/`async` |
| `service_loader` | ServiceLoader use whose service or providers are unknown | site/`serviceloader`, allocation/`service-unresolved` |
| `deserialization` | a deserialization entry whose input classes are open | site/`deserialization` |
| `dynamic_proxy` | a JDK Proxy / generated InvocationHandler dispatch | allocation/`dynamic-proxy`, runtime/`proxy` |
| `container_proxy` | a container-generated implementation (repository, AOP proxy) | allocation/`container-proxy` |
| `unknown_allocation` | an object created where the type is unknown (Unsafe, constructor handles) | site/`allocation`, allocation/`unsafe-allocation` |
| `external_input` | `-roots all`: a root's parameters may be any subtype | allocation/`external-input` |
| `client_code` | `-roots api`: a library call a client can implement or override | site/`client`, site/`supplied` |
| `lambda` | a functional-interface call whose lambda / method reference is unresolved | lambda/`bootstrap`, lambda/`budget`, lambda/`call-result`, lambda/`depth`, lambda/`field`, lambda/`instance-call`, lambda/`parameter`, lambda/`sam-argument`, lambda/`unknown-value`, runtime/`invokedynamic`, runtime/`sam-call` |
| `skipped_receiver` | a call on `java.`/`javax.`/`jdk.`/`sun.` types, never CHA-expanded | runtime/`skipped-receiver` |
| `virtual_dispatch` | a virtual call target the graph has no edge for (RTA pruning, inherited bodies) | runtime/`dispatch-missing`, runtime/`inherited-impl`, runtime/`rta-pruned` |
| `class_init` | a class initializer the JVM runs implicitly | runtime/`class-init` |
| `model_incomplete` | a library model that could not resolve a site (B3 coordinator; detail: model id) | coordinator/*model id* |
| `trace_alignment` | a measured call not attributable to one site (dropped frames, inlining) | runtime/`indirect`, runtime/`no-call-at-site` |
| `unclassified` | no mechanism known | runtime/`unclassified` |

