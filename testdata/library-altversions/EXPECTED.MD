# Control case for other dependency versions (LIBRARY_MODE LB3 `lb3-2`)

Written 2026-11-23 with `lb3-2`. The library is `testdata/library-deps/lib.jar` (one API method,
`ld.Render#render`, reaching commons-text's interpolator and, behind it, commons-lang3's
`ClassUtils#getClass` — CVE-2022-42889 and CVE-2025-48924), its dependencies
`testdata/library/deps` (commons-text 1.9, commons-lang3 3.11), and its declarations
`pom.xml` here, read with `-library-pom`:

| Case | Declaration | Version in the input |
|---|---|---|
| (u) | `commons-text` `[1.6,1.10.0)`, `provided` | 1.9 |
| (v) | `commons-lang3` `[3.11,3.18.0)` (compile) | 3.11 |

`cache/` adds to `testdata/library/cache` and `testdata/library-deps/cache`: commons-lang3's
`maven-metadata.xml` (Central, fetched 2026-11-10: 3.11 … 3.19.0) and the OSV answer for
commons-lang3 3.17.0 (fetched 2026-11-23: CVE-2025-48924, fixed in 3.18.0). commons-text's version
list is deliberately absent.

## Expected with `-roots api -alt-versions newest` (mode `approx`)

- The run's own findings are those of the same run without `-alt-versions`; v1 is byte-identical.
- `coverage.api.alternate_versions`: `strategy: newest`, `mode: approx`, `rules_version: lb3-2`;
  - (v) commons-lang3: one version, **3.17.0**, `chosen_by: [range_max]` (the highest published
    version inside `[3.11,3.18.0)`; the 3.11 line has nothing newer, so no `line_latest`);
    finding **CVE-2025-48924 `REACHABLE`** — the advisory affects 3.17.0, its mapping has no
    variants, and reachability is computed on the 3.11 code (`basis` says so);
  - (u) commons-text: no version, a note that its published versions are unavailable (offline,
    not in the cache) — never "no other version".
- Without the OSV answer for 3.17.0 the version is chosen but `not_assessed` with the reason,
  and the run still succeeds.
- `-alt-versions` without `-roots api`, an unknown strategy, and `-alt-versions-mode rerun` (the
  future second graph run, option B) are rejected.
