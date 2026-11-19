# Control library for CDI bean archives (LIBRARY_MODE LB2)

Written 2026-11-18, before the implementation. Contract: `docs/RESULTS_CONTRACT.MD` §8.

`lib.jar` (sources in `lib/`) is a CDI bean archive: `META-INF/beans.xml` with
`bean-discovery-mode="annotated"` (CDI 4.0). Its public API (`cd.api`: `Banner`, `Startup`,
`Store`, `Component`) reaches nothing vulnerable. The only caller of commons-text 1.9's
`StringSubstitutor#replace` (CVE-2022-42889; commons-lang3 3.11's `ClassUtils#getClass` behind
the interpolator's lookups, CVE-2025-48924) is the package-private `cd.internal.Expand#expand`,
called from package-private classes and methods a CDI container runs:

| Case | Code | Runs under CDI? |
|---|---|---|
| (t) | `@ApplicationScoped class Producers`, package-private `@Produces Banner banner()` | yes, when a client injects `Banner` |
| (u) | `@Dependent class AuditObserver`, package-private `onStartup(@Observes Startup)` | yes, when a client fires `Startup` |
| (v) | `@ApplicationScoped class CachedStore implements Store`, package-private `@PostConstruct warm()` | yes, when a client uses a `Store` |
| (w) | `@Dependent class Wiring implements Component`, package-private `@Inject init(Banner)` | yes, when a client injects `Component` |
| (x) | `class Unmanaged` — no bean-defining annotation — with `@PostConstruct never()` and `@Produces neverProduced()` | no: in an `annotated` archive it is not a bean |

`observed.txt` (`scripts/library-cdi-oracle.sh`: Weld SE 5.1.2.Final, CDI 4.0, JDI trace of
`fx.Main`, which boots the container and injects `Banner`, fires `Startup`, uses `Store` and
`Component`): the JVM ran `Producers#banner`, `AuditObserver#onStartup`, `CachedStore#warm`,
`Wiring#init`, `Expand#expand` and `StringSubstitutor#replace`; never `Unmanaged`.

## Before (measured 2026-11-18, `lb3-1`)

- `-roots all` (CHA, RTA): both advisories `REACHABLE` — every method is a root, `Unmanaged`'s
  too.
- `-roots api` (CHA, RTA): both `NOT_REACHABLE`, `exposed_api` empty — although the JVM runs
  `replace`.

## Expected with CDI bean archives as announcements

`-roots api`, CHA and RTA: CVE-2022-42889 and CVE-2025-48924 `REACHABLE`; `exposed_api` is
exactly `Producers#banner`, `AuditObserver#onStartup`, `CachedStore#warm`, `Wiring#init` (4),
each with `announced_by` naming `META-INF/beans.xml`, the discovery mode and the bean-defining
annotation. Never in `exposed_api`: `Unmanaged#never`, `Unmanaged#neverProduced`,
`Expand#expand` (not an entry). `coverage.api.announced_classes` = 4 and the announcements list
the four beans. Every `exposed_api` method ran on the JVM.

`-roots all` and `-roots profile`: unchanged.
