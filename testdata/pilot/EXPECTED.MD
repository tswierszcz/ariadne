# Pilot control target (RESEARCH_HARNESS RH-A3)

Written 2026-11-25 for stage RH-A3 of the [Recon → Hunt → Validate plan](../../RECOMMENDATION/RESEARCH_HARNESS_IMPLEMENTATION_PLAN.MD). The target is built by `scripts/pilot-fixture.sh`. Its jars, `observed-*.txt` and `effects-*.txt` are committed, so the tests need no JDK.

**Property:** `px.archive.Unpacker#unpack` writes only inside the destination directory the `Unpacker` was constructed with.

The library comes in two variants with the same public API. Both are built from `common/` plus one version of `px.archive.Entries`:

| Jar | `Entries#target` | Property |
|---|---|---|
| `defective.jar` (`defective/`) | `dest.resolve(name)`, so `../x` leaves the directory | violated |
| `fixed.jar` (`fixed/`) | normalizes the name and refuses it if it resolves outside `dest` | kept |

The oracle is `client.jar` (`client/px/client/Main.java`). It uses only the public API and local files in a temporary directory, and runs three scenarios:

- **normal:** one entry inside the directory,
- **traversal:** the entry `../escaped.txt`,
- **control:** `writeFile` to a path the client chooses.

Each variant is run under `scripts/MethodTrace.java`, which produces two files:

- `observed-VARIANT.txt`: the `px.archive.*` method bodies the JVM entered. They are the same for both variants.
- `effects-VARIANT.txt`: the client's file-system observations, independent of any static claim:

| Scenario | defective | fixed |
|---|---|---|
| normal | unpacked 1, nothing outside | unpacked 1, nothing outside |
| traversal | unpacked 1, **file outside dest** | refused, nothing outside |
| allowed API | wrote | wrote |

## What Ariadne answers

Each variant is indexed with `-roots api -profile generic`, offline. The answers below are checked by `TestPilotTarget` in `cmd/ariadne/pilot_test.go`.

| Target | Answer (both variants) | Meaning for the experiment |
|---|---|---|
| Every method the JVM ran | `found` | The static graph covers the observed execution |
| `Store#write` (the write method) | `found`, the same entries and characters in both variants: `Unpacker#unpack`, `Unpacker#writeFile`, and `Store#write` itself (a public class in `internal`, so part of the API unless `-api-exclude px.archive.internal`) | Reachability does not decide the property; only the experiment's assertion does |
| `Store#overwrite` (control: dead method) | `not_found` | — |
| Model of every target | `incomplete`: `Unpacker#filter` loads a class named by the caller (control: unresolved reflection) | A gap is a gap in the evidence, not a refutation |
| `Unpacker#writeFile` → `Store#write` (control: allowed API) | `found` | Allowed by the API contract. It is a path, not a finding. |
| `calls` in `Entries#target` on `java.nio.file.Path` | defective: `resolve`; fixed: `normalize`, `resolve`, `startsWith`, … | Where the variants differ: the guard |

`TestPilotConflictingSources` covers the case where `defective.jar` is indexed with `-sources` from the fixed variant. `Entries#target` then gets the fixed file's hash under the binding `declared`. Ariadne matches sources by path, so it cannot see that they were not the sources of the build. Deciding whether binary and sources correspond (rebuild, compare) is the consumer's job.
