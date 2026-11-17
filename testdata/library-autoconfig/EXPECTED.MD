# Control starter for library mode: autoconfiguration roots (LIBRARY_MODE LB2)

Written 2026-11-17, before the implementation. Contract: `docs/RESULTS_CONTRACT.MD` §8.

`lib.jar` (sources in `lib/`) is a small Spring Boot starter. Its vulnerable dependency is
commons-text 1.9 (`StringSubstitutor#replace`, CVE-2022-42889; commons-lang3 3.11 behind it,
CVE-2025-48924), the same jars as `testdata/library` (`../library/deps`, offline cache
`../library/cache`). Spring is `provided`: the library compiles against Spring Boot 2.4.5 /
Spring Framework 5.3.6, but they are not analysis dependencies (as a starter's own pom would
declare them) — the analyzer sees the annotation names in the bytecode only.

Every class but `ac.api.Greeter` is package-private or has package-private members: none of
the vulnerable code is reachable from the public API. Spring reaches it because the library
announces its entries:

- `META-INF/spring.factories`:
  `org.springframework.boot.autoconfigure.EnableAutoConfiguration=ac.autoconfigure.GreeterAutoConfiguration`,
  `org.springframework.context.ApplicationListener=ac.autoconfigure.StartupListener`;
- `META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports`
  (Spring Boot 2.7+ format): `ac.autoconfigure.ModernAutoConfiguration`.

## Cases

| Case | Code | Root under `-roots api`? | Runs on Spring Boot 2.4.5? |
|---|---|---|---|
| (j) | `GreeterAutoConfiguration#banner` — package-private `@Bean` method of the announced autoconfiguration → `replace` | yes (announced: `@Bean` of an `EnableAutoConfiguration` entry) | yes |
| (k) | `GreeterAutoConfiguration$NestedConfiguration#nestedBanner` — package-private `@Bean` in a package-private nested `@Configuration` class | yes (member configuration class of an announced one) | yes |
| (l) | `ImportedConfiguration#init` — `@PostConstruct` of a package-private `@Configuration` class the announced one `@Import`s | yes (`@Import` of an announced one) | yes |
| (m) | `StartupListener#onApplicationEvent` — package-private `ApplicationListener` in `spring.factories` | yes (a `spring.factories` entry of any other key: constructor and public methods) | yes |
| (n) | `UnusedConfiguration#unusedBanner` — package-private `@Configuration`, never announced or imported | no | no |
| (o) | `ModernAutoConfiguration#modernBanner` — only in `AutoConfiguration.imports` | yes | no — Spring Boot 2.4 does not read that file (2.7+ does); an expected difference from the oracle |

`ac.api.Greeter#greet` is public API and a `@Bean` product; it reaches nothing.

## Expected with `-roots api` (CHA and RTA)

Before this change (`lb2-4`): CVE-2022-42889 and CVE-2025-48924 `NOT_REACHABLE` — no public
API method reaches `replace`.

After: both `REACHABLE`; `exposed_api` (library methods) = `GreeterAutoConfiguration#banner`,
`GreeterAutoConfiguration$NestedConfiguration#nestedBanner`, `ImportedConfiguration#init`,
`StartupListener#onApplicationEvent`, `ModernAutoConfiguration#modernBanner`, each with
`announced_by` naming the file, the key and how it was reached (member class, `@Import`).
Never: `UnusedConfiguration#unusedBanner`, `Greeter#greet`. `coverage.api` counts the
announced classes. With `-roots all` nothing changes (every method is a root already).

`observed.txt` (`scripts/library-autoconfig-oracle.sh`, JDI trace of `fx.Main`, a Spring
Boot 2.4.5 application with `@EnableAutoConfiguration` and the starter on the class path):
the bodies of (j)–(m) ran and `StringSubstitutor#replace` with them; (n) and (o) never ran.

## Measured (2026-11-17, `lb2-5`)

As expected, under CHA and RTA. Before (`lb2-4`, same input and cache): both advisories
`NOT_REACHABLE`, `coverage.api` 2 classes, 3 methods. After: both `REACHABLE`, 13 root
methods (10 from 5 announced classes in 3 announcements), `exposed_api` exactly the five
entries above with `announced_by`, e.g. `ImportedConfiguration#init`:
`META-INF/spring.factories org.springframework.boot.autoconfigure.EnableAutoConfiguration:
ac.autoconfigure.GreeterAutoConfiguration > @Import ac.autoconfigure.ImportedConfiguration`.
The oracle (`observed.txt`) ran (j)–(m) and `replace`, and neither (n) nor (o).
