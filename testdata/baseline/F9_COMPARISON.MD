# CVE:F9.7 — control apps, CVE:F0 baseline vs the F0–F7 increments

Every v1 verdict change between the CVE:F0 baseline build (`45045d6`, "Implement CVE:F0")
and `main` after CVE:F9a (`e8ee758`), on petclinic and WebGoat, with its cause. Run on
2026-11-06 with `scripts/cve-compare.sh 45045d6 main`: both builds offline, `-profile
spring`, on the same `.cache` (a data snapshot cannot be shared — `-snapshot-in` refuses
another analyzer version by design, CVE:F3). The F0 build on this cache reproduces the
committed baseline (`testdata/baseline/*/results.json`) byte for byte, so the old side is
the baseline itself.

A falling number of `REACHABLE` rows is not, by itself, an improvement; each change below
is attributed to a change of data, component identity, mapping or graph, and checked.

## Summary

| App | Advisories F0 → now | Status changes | Of which: data | reason codes (F1b) | mapping (F5) |
|---|---|---|---|---|---|
| petclinic | 148 → 153 | 9 + 5 new | 5 new | 7 | 2 |
| WebGoat | 160 → 164 | 11 + 4 new | 4 new | 6 | 5 (incl. 2 `REACHABLE` → `FRAMEWORK`) |

No change is unexplained. No change comes from the graph (JVM_GRAPH, FRAMEWORKS and the
other areas did not change between these builds in a way that moved a verdict here) or
from embedded components (none of the embedded versions found by CVE:F6/F7 has an OSV
advisory).

## 1. Data: advisories present only now (5 petclinic, 4 WebGoat)

| App | Advisory | Package | Status now |
|---|---|---|---|
| petclinic | CVE-2026-53404, CVE-2026-55276, CVE-2026-55956, CVE-2026-55957 | `org.apache.tomcat.embed:tomcat-embed-core` 9.0.45 | FRAMEWORK |
| both | CVE-2026-59889 | `com.fasterxml.jackson.core:jackson-databind` 2.11.4 | FRAMEWORK |
| WebGoat | CVE-2026-28367, CVE-2026-28368, CVE-2026-28369 | `io.undertow:undertow-core` 2.2.4.Final | FRAMEWORK |

Cause: **a different OSV answer**, not a code change. The OSV `querybatch` cache key is
derived from the queried GAV list, and the current build also queries the embedded
components confirmed by CVE:F6a/F6b/F7a (petclinic 89 GAVs instead of 87 — `sizeof`,
`terracotta-utilities-tools`; WebGoat 133 instead of 126 — jnr-ffi, jitescript, five ASM
modules). So each build reads its own cached answer, fetched at a different time
(petclinic 10:13:56 vs 10:16:55 UTC, WebGoat 10:14:18 vs 11:13:17 UTC). Verified from the
cached responses (`ARIADNE_CACHE_TRACE`): the newer answers contain exactly these advisories'
GHSA ids (petclinic `GHSA-h792-v28v-ppgr`, `-mqg3-r7h5-24x4`, `-5v3h-fjv2-54fg`,
`-7895-gffq-w6jq`, `-5gvw-p9qm-jgwh`; WebGoat `GHSA-3gv6-g396-9v4r`, `-8v4x-mgvp-p658`,
`-vqqj-9cmv-hx43`, `-5gvw-p9qm-jgwh`) in addition to everything in the older ones, and
nothing is missing. The affected packages are ordinary dependencies, not embedded ones.

## 2. Reason codes: `NOT_REACHABLE` → `UNKNOWN` + `METHOD_NOT_FOUND` (7 petclinic, 6 WebGoat)

CVE-2025-11226, CVE-2026-1225 (logback-core 1.2.3), CVE-2026-50193, CVE-2026-54514,
CVE-2026-54515 (jackson-databind 2.11.4), CVE-2026-68494 (jackson-core 2.11.4) on both apps,
and CVE-2026-34487 (tomcat-embed-core 9.0.45) on petclinic.

Cause: CVE:F1b's contract rule (`docs/RESULTS_CONTRACT.MD` §2) — a mapping whose methods
match no node of the graph is "method not found", not "no path". These are exactly the
7 and 6 rows predicted before any code change by
[VERDICT_CHANGE_ESTIMATE.MD](VERDICT_CHANGE_ESTIMATE.MD) §1. Some method lists are now
shorter (e.g. CVE-2026-1225 16 → 10): CVE:F5b keeps only fix-diff candidates confirmed
by the bytecode — the verdict stays `UNKNOWN` either way. CVE-2026-68494 also carries
`MAPPING_NEW_SIDE_ONLY` (CVE:F5a: the fix diff could only be read on its post-fix side).

## 3. Mapping (CVE:F5): both fix sides, nested classes, bytecode verification

| App | Advisory | F0 | Now | Cause |
|---|---|---|---|---|
| both | CVE-2024-12798 (logback) | UNKNOWN, no method | FRAMEWORK, 12 exact selectors | The fix deletes `JaninoEventEvaluator.java`; F0 read only the post-fix side, so nothing was left. CVE:F5a reads the pre-fix side (methods `removed`, `bytecode: exact`), and a path through `JaninoEventEvaluatorBase#evaluate` exists from framework code. |
| both | CVE-2023-20873 (Spring Boot actuator) | NOT_REACHABLE, 14 name-only selectors | FRAMEWORK (`MAPPING_NEW_SIDE_ONLY`), 14 selectors | CVE:F5b attributes changed methods to nested classes (`ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor#postProcessAfterInitialization`); F0 looked them up on the outer class, where they do not exist. The new path ends in that nested-class method. |
| WebGoat | CVE-2024-38821 (Spring Security) | REACHABLE, 227 name-only selectors | FRAMEWORK, 1 exact selector | F0 took every method of every touched file (e.g. `MapUserDetailsRepository#findByUsername`) and the REACHABLE path ended in an unrelated `AbstractServerWebExchangeMatcherRegistry#matchers`. CVE:F5b keeps the bytecode-verified changed method `WebFilterChainProxy#filter(ServerWebExchange, WebFilterChain)`; it is reached through framework code, not directly from the app. |
| WebGoat | CVE-2026-41848 (Spring `AntPathMatcher`) | REACHABLE, 7 name-only selectors | FRAMEWORK, 1 exact selector | F0's names included `AntPathMatcher#<init>`, `charAt`, `length`, … (members of the changed file, not the fix), and the REACHABLE path ended in `AntPathMatcher#<init>()V`. The fix is in the nested `AntPathMatcher$AntPathStringMatcher#matchStrings(String, Map)` (CVE:F5b nested classes); that method is reached through framework code. |

The two `REACHABLE` → `FRAMEWORK` changes are therefore the removal of false positives of
an over-broad mapping (a path to a constructor or an unrelated method of the fixed file),
not a lost finding: the fixed method itself is still reachable, via the framework.

## Reproducing

```bash
scripts/cve-compare.sh 45045d6 main
```

The output lists the same 9/11 changes and 5/4 new advisories as long as the cache holds
the same OSV answers; with a refreshed cache the data part (§1) can differ and has to be
re-explained.
