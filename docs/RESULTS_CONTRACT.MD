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
| Component identity | Which artifact and component instance is this, and how was it identified? | `identified` (GAV from `pom.properties`/`deps.txt`, with evidence), `unidentified`, `conflict` | GAV or "unidentified jar"; since CVE:F1a an artifact/instance inventory with SHA-256, stable location and identity evidence (report "Components"), not yet used by findings |
| Applicability | Does the advisory's affected range include this instance's version? | `applicability`: `affected`, `not_affected`, `undetermined` | OSV is queried with the exact GAV; an unidentified jar is never queried |
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
deprecated. CVE:F0 defines them; nothing sets them yet.

| Code | Axis | Meaning | Example |
|---|---|---|---|
| `VERSION_UNKNOWN` | identity/version | The component's version could not be established, so the affected range cannot be checked | a jar without `pom.properties` and no `deps.txt` line |
| `CLASSPATH_CONFLICT` | identity/version | Several copies of a class are on the input; the selected one follows the declared classpath order, which may not match the real deployment | the same FQCN in two jars (`Hierarchy.Duplicates()`) |
| `MAPPING_PARTIAL` | mapping | The method list is known to cover only part of the vulnerable code | a diff mapping that touched one of several affected methods |
| `MAPPING_ENTRYPOINT_ONLY` | mapping | The mapped methods are entry points into the vulnerable behavior (e.g. a deserialization call), not the vulnerable code itself | XStream `fromXML` for gadget-chain CVEs |
| `METHOD_NOT_FOUND` | mapping | No mapped method exists in the analyzed graph | a method renamed in the analyzed version |
| `SOURCE_UNAVAILABLE` | mapping | The evidence needed for a mapping (fix commit, diff, advisory detail) could not be obtained | an OSV record without a fix reference |
| `MODEL_INCOMPLETE` | model | An explicit gap in the call-graph model (unresolved reflection, ServiceLoader, callback, …) is relevant to this finding | an unresolved `Method#invoke` in the affected package |
| `RUNTIME_UNSUPPORTED` | runtime | The finding depends on runtime behavior ariadne does not model (JDK/JRE vulnerabilities, custom class loaders) | a CVE in the JDK itself |

## 4. Model gaps and verdicts (B0)

- A **global gap** is any incomplete-model fact of the run (today's "model analysis: …"
  warnings, unresolved reflection/callback/ServiceLoader sites).
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

## 5. JSON compatibility

- **v1** (today's `-json` array of `{cve, status, methods, source, …}`) stays the default
  and keeps its meaning. Additive fields may appear; no field changes type or meaning.
- **v2 (F1)** is opt-in (`-json-format v2`): an object with `schema_version`, components,
  findings (advisory × component instance), coverage and a data-snapshot reference.
- **v1 export from v2 (F1):** one row per advisory. Aggregation: any positive finding
  wins (`REACHABLE` over `FRAMEWORK`); without one, an unresolved instance is never
  hidden under `NOT_REACHABLE` — the row becomes `UNKNOWN`; `NOT_APPLICABLE` only when
  every instance is a WebJar. Counts of unique advisories and of per-component findings
  are reported separately.

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
