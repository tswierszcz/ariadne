# Results contract (CVE:F0)

Status: **contract** — CVE:F0 of [RECOMMENDATION/CVE_IMPLEMENTATION_PLAN.MD](../RECOMMENDATION/CVE_IMPLEMENTATION_PLAN.MD),
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
| `NOT_APPLICABLE` | Every affected dependency is a WebJar (client-side assets, no bytecode); the call graph cannot answer. Unchanged in F1 — in particular it does **not** mean "version not affected" (that is `applicability: not_affected`) | WebGoat: a jQuery advisory on a WebJar |
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
`MAPPING_NEW_SIDE_ONLY` and emits `SOURCE_UNAVAILABLE`; CVE:F6b adds `VERSION_CANDIDATES`;
`RUNTIME_UNSUPPORTED` is not set yet.
`MAPPING_NEW_SIDE_ONLY`, like `MAPPING_PARTIAL`, qualifies the advisory row and stays off
instance findings (v2 carries it as `mapping.new_side_only`).
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
| `SOURCE_UNAVAILABLE` | mapping | The evidence needed for a mapping (fix commit, diff, advisory detail) could not be obtained. (CVE:F5) Emitted when the advisory has fix-commit references and none could be read (finding `UNKNOWN`); an advisory with no fix reference at all is a plain unmapped `UNKNOWN` | every fix-commit diff answers HTTP 500 |
| `MODEL_INCOMPLETE` | model | An explicit gap in the call-graph model (unresolved reflection, ServiceLoader, callback, …) is relevant to this finding | an unresolved `Method#invoke` in the affected package |
| `RUNTIME_UNSUPPORTED` | runtime | The finding depends on runtime behavior ariadne does not model (JDK/JRE vulnerabilities, custom class loaders) | a CVE in the JDK itself |

## 4. Model gaps and verdicts (B0)

- A **global gap** is any incomplete-model fact of the run (today's "model analysis: …"
  warnings, unresolved reflection/callback/ServiceLoader sites, and — since
  DESERIALIZATION Z1 — every deserialization entry, kind `deserialization`, unbounded:
  the JDK's own serialization callbacks are not modeled).
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

## 5. JSON compatibility

- **v1** (today's `-json` array of `{cve, status, methods, source, …}`) stays the default
  and keeps its meaning. Additive fields may appear; no field changes type or meaning.
- **v2 (F1, implemented by F1c)** is opt-in (`-json-format v2`): an object with `schema_version`, components,
  findings (advisory × component instance), coverage and a data-snapshot reference.
- **v1 export from v2 (F1):** one row per advisory. Aggregation: any positive finding
  wins (`REACHABLE` over `FRAMEWORK`); without one, an unresolved instance is never
  hidden under `NOT_REACHABLE` — the row becomes `UNKNOWN`; `NOT_APPLICABLE` only when
  every instance is a WebJar. Counts of unique advisories and of per-component findings
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

## 7. Reference data

- **Control set:** [testdata/cve-control/](../testdata/cve-control/README.MD) — fictional
  advisories for mechanics and a source-documented real-CVE list for quality, covering
  all five CVE extensions.
- **Baseline:** [testdata/baseline/](../testdata/baseline/README.MD) — the current
  version's results on petclinic and WebGoat with hashes of every input and of every
  network response used. A baseline is a reference for change, **not ground truth**.
