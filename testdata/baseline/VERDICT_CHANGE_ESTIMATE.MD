# Which v1 verdicts the F1 semantics would change (CVE:F0 estimate)

Measured on the CVE:F0 baseline ([README.MD](README.MD)) with a temporary probe (a Go test
using `reach.matchVulnerableMethod` on the same hierarchy/graph, removed after use). Only
`NOT_REACHABLE` results are affected by the contract's new rules; `REACHABLE`,
`FRAMEWORK`, `NOT_APPLICABLE` and existing `UNKNOWN` keep their v1 verdicts.

## 1. "Method not found" reported as `NOT_REACHABLE` (contract §2)

A `NOT_REACHABLE` row whose mapped methods match **no** node of the graph. Under F1 these
become `UNKNOWN` + `METHOD_NOT_FOUND`.

| App | `NOT_REACHABLE` today | Method not found | Advisories |
|---|---|---|---|
| petclinic | 26 | **7** | CVE-2025-11226, CVE-2026-1225 (logback-core 1.2.3); CVE-2026-34487 (tomcat-embed-core 9.0.45); CVE-2026-50193, CVE-2026-54514, CVE-2026-54515 (jackson-databind 2.11.4); CVE-2026-68494 (jackson-core 2.11.4) |
| WebGoat | 18 | **6** | the same logback, jackson-databind and jackson-core advisories |

All of these are diff-derived mappings (the methods were taken from the fix commit and do
not exist, under those names, in the analyzed version) — the case CVE:F4/F5 (version and
pre-fix mapping) address. This is a real semantic flaw of v1 output, stated now in
README/SUPPORT rather than left for F1.

## 2. Model gaps (contract §4)

Upper bound with the *broad* relevance rule "any unresolved reflection / ServiceLoader /
callback site inside the affected jar", ignoring whether the site is reachable:

| App | `NOT_REACHABLE` | With a gap in the affected jar | Also method-not-found |
|---|---|---|---|
| petclinic | 26 | 25 | 7 |
| WebGoat | 18 | 17 | 6 |

Conclusion recorded in the contract: the broad rule would turn almost every
`NOT_REACHABLE` into `UNKNOWN` and carry no information, so F1's rule also requires the
unresolved site to be reachable from the roots. The tighter count is measured when F1b
implements the rule; it is not estimated here.

## 3. What should stay compatible

`REACHABLE`/`FRAMEWORK` paths, `NOT_APPLICABLE` (WebJars) and the existing `UNKNOWN`
rows (no mapping, LLM advisory-only) keep their verdicts. Counts in the v1 summary change
only by moving rows from `NOT_REACHABLE` to `UNKNOWN`.
