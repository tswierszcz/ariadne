# Control library for Micronaut bean definitions (LIBRARY_MODE LB2, measurement)

Written 2026-11-20 for the LB2 question "are Micronaut's generated bean
definitions already public API?". Contract: `docs/RESULTS_CONTRACT.MD` §8.

`lib.jar` (sources in `lib/`) is compiled with Micronaut 3.9.7's annotation
processor (`scripts/library-micronaut-oracle.sh`), which generates, next to
each bean, a package-private synthetic `$X$Definition` (its `build`,
`injectBean`, `initialize` call the bean's constructor, `@Inject` and
`@PostConstruct` methods; `$X$Definition$Exec#dispatch` calls executable
methods such as `@EventListener`) and a public synthetic
`$X$Definition$Reference` (`load()` only instantiates the definition), listed
under `META-INF/micronaut/io.micronaut.inject.BeanDefinitionReference/`. Its
public API (`mn.api`: `Banner`, `Store`, `Component`, `Startup`) reaches
nothing vulnerable. The only caller of commons-text 1.9's
`StringSubstitutor#replace` (CVE-2022-42889; CVE-2025-48924 behind it) is the
package-private `mn.internal.Expand#expand`, called from package-private
classes and methods Micronaut runs — the same cases as `testdata/library-cdi`:

| Case | Code | Runs under Micronaut? |
|---|---|---|
| (t) | `@Factory class Producers`, package-private `@Singleton Banner banner()` | yes, when a client asks for `Banner` |
| (u) | `@Singleton class AuditListener`, package-private `@EventListener onStartup(Startup)` | yes, when a client publishes `Startup` |
| (v) | `@Singleton class CachedStore implements Store`, package-private `@PostConstruct warm()` | yes, when a client uses a `Store` |
| (w) | `@Singleton class Wiring implements Component`, package-private `@Inject init(Banner)` | yes, when a client asks for a `Component` |
| (x) | `class Unmanaged` — no bean annotation — with `@PostConstruct never()` | no: no definition is generated |

`observed.txt` (JDI trace of `fx.Main`, which runs `ApplicationContext.run()`,
asks for `Banner`, publishes `Startup`, uses `Store` and `Component`): the JVM
ran `Producers#banner`, `AuditListener#onStartup`, `CachedStore#warm`,
`Wiring#init`, `Expand#expand` and `StringSubstitutor#replace`; never
`Unmanaged`.

## Measured (2026-11-20)

- The definitions are **not** public API: `$X$Definition` is package-private
  and synthetic; `$X$Definition$Reference` is public but synthetic (library
  mode leaves synthetic classes out) and only instantiates the definition.
  The framework calls `build`/`injectBean`/`initialize`/`dispatch` through
  `BeanDefinition`, from outside the library.
- `-roots api` (CHA, RTA; `-profile generic` or the detected `micronaut`;
  with or without the Micronaut jars in `-deps`): both advisories
  `NOT_REACHABLE`, `exposed_api` empty — **a false negative**: the JVM runs
  `replace`. `coverage.api`: 4 classes, 5 methods, 0 announced.
- `-roots all`: both `REACHABLE`, but every method is a root, `Unmanaged`'s
  too.

## Expected with Micronaut definitions as announcements (not implemented)

`-roots api`, CHA and RTA: both advisories `REACHABLE`; `exposed_api` is
exactly `Producers#banner`, `AuditListener#onStartup`, `CachedStore#warm`,
`Wiring#init` (4) — reached through each listed definition's
`build`/`injectBean`/`initialize`/`Exec#dispatch` as roots, with
`announced_by` naming the `META-INF/micronaut/…BeanDefinitionReference`
entry. Never in `exposed_api`: `Unmanaged#never`, `Expand#expand`. Micronaut
2.x lists the references in `META-INF/services/io.micronaut.inject.BeanDefinitionReference`
instead; 4.x is not measured here (its jars need Java 17+ class files, fine
for ariadne, but the layout has to be checked).
