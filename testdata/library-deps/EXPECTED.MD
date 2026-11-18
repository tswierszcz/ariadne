# Control library for dependency declarations (LIBRARY_MODE LB3)

Written 2026-11-18, before `lb3-1`. Contract: `docs/RESULTS_CONTRACT.MD` §8.

`lib.jar` (sources in `lib/`) has one API method, `ld.Render#render`, which calls
`StringSubstitutor.createInterpolator().replace` from commons-text (CVE-2022-42889; commons-lang3's
`ClassUtils#getClass` behind the interpolator's lookups is CVE-2025-48924), as `testdata/library`
does. Its embedded pom `META-INF/maven/fx/library-deps/pom.xml` declares:

| Case | Declaration | Who supplies it to a user of the library | In the analysis input (`testdata/library/deps`)? |
|---|---|---|---|
| (p) | `commons-text` `${text.range}` = `[1.6,1.10.0)`, `provided` | the user | yes, 1.9 — inside the range |
| (q) | `commons-lang3` `3.11` (compile) | the library | yes, 3.11 — the declared version |
| (r) | `commons-io` `2.6`, `runtime` | the library | no |
| (s) | `spring-core` `5.3.0`, `optional` | the user (only if added) | no |
| — | `junit` `4.13.2`, `test` | nobody: not part of the published dependencies | no |

commons-text 1.9's own pom declares commons-lang3 3.11 (compile). For a consumer that edge does not
matter: provided dependencies are not transitive, and commons-lang3 is a direct compile
dependency of the library anyway (nearest wins).

## Oracle

`scripts/library-deps-oracle.sh` puts the jar and its pom into a file repository and asks Maven
3.9 what a consumer of `fx:library-deps:1.0` gets on its runtime classpath
(`maven-consumer.txt`): `commons-io:commons-io:2.6:runtime`, `org.apache.commons:commons-lang3:3.11:compile`
— exactly cases (q) and (r). Neither the provided commons-text nor the optional spring-core.

## Expected with `-roots api` (`lb3-1`, CHA and RTA)

- Verdicts unchanged from the same run without declarations: CVE-2022-42889 and CVE-2025-48924
  `REACHABLE`, `exposed_api` = `Render#render`.
- `coverage.api.dependencies`: `pom` = `lib.jar!/META-INF/maven/fx/library-deps/pom.xml`,
  `library` = `fx:library-deps:1.0`; `declared` (by package):
  - `commons-io:commons-io` — `supplied_by: library`, scope `runtime`, version `2.6`,
    `version_match: absent`;
  - `org.apache.commons:commons-lang3` — `library`, `compile`, `3.11`, `declared`, with the
    instance of `commons-lang3-3.11.jar`;
  - `org.apache.commons:commons-text` — `user`, `provided`, declared `${text.range}`, version
    `[1.6,1.10.0)`, `in_range` (1.9);
  - `org.springframework:spring-core` — `user`, `compile`, `optional`, `5.3.0`, `absent`;
  - `org.springframework:spring-jcl` — `user`, `compile`, `optional`, via `spring-core`, `5.3.0`,
    `absent` (*added after the first run, 2026-11-18*: spring-core 5.3.0's pom declares it
    compile, so a user who adds the optional spring-core gets it too; the list above missed it);
  - no junit (test scope), nothing `undeclared`.
- The set with `supplied_by: library` equals `maven-consumer.txt`.
- `components[]`: the commons-text and commons-lang3 instances carry their `declaration`;
  `findings[].supplied_by`: CVE-2022-42889 `user`, CVE-2025-48924 `library`.
- A warning names commons-io: a consumer gets it through the library, but it is not in the
  input, so its advisories are not assessed.
- `-roots all` and `-roots profile`: no `dependencies`, no `supplied_by`, nothing read.
