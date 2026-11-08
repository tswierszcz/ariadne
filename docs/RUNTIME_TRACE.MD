# Runtime trace contract (`ariadne-runtime-trace-1`)

RUNTIME_NATIVE N0. A runtime trace is evidence from **one** recorded run of the application:
which classes were loaded, which methods were seen executing, and which caller → callee
relations were measured. `-runtime-trace FILE` imports it next to the static analysis
(N1, `internal/runtimeevents`). It never changes the call graph or a verdict.

Reading rules, everywhere a trace is shown:

- **Absence of an event never proves anything.** A method not seen in a trace may run under
  other traffic, configuration, timing, a longer run, or simply between two samples.
  Nothing becomes `NOT_REACHABLE`, or less reachable, because of a trace.
- **Execution is not exploitation.** A vulnerable method seen executing confirms that this
  code ran in this run, not that attacker-controlled data reached it.
- **A loaded class is not an executed method.** Loading says nothing about which of the
  class's methods ran.
- **Order in time is not causality.** A caller → callee relation exists only where the
  recorder measured it (a stack frame directly above another, or an instrumented entry
  with its caller frame), never inferred from timestamps.
- **A trace is about the artifacts it was recorded on.** Evidence is attached to an
  analyzed class only when the class content (or, failing that, its artifact) has the
  same SHA-256 as recorded; a name match alone is reported, not used.

## File format

UTF-8 JSON Lines: one JSON object per line, with a `record` field naming its kind. The
first line is the header. Every id is unique within its kind and only meaningful inside
this file (a loader id in particular says nothing about another run). Unknown record kinds
or fields, a dangling reference, a duplicate id or a value outside the enumerations below
reject the whole file.

Limits (the importer refuses larger input): file 1 GiB, line 1 MiB, 5 000 000 records,
strings 4 096 bytes.

### Header

```json
{"record":"header","format":"ariadne-runtime-trace-1","run_id":"…","scenario":"…",
 "started":"2026-11-08T20:00:00Z","ended":"2026-11-08T20:05:00Z",
 "runtime":{"name":"OpenJDK Runtime Environment","version":"25.0.5","vendor":"…"},
 "recorder":{"name":"ariadne-jfr-convert","version":"1","source":"JFR jdk.ExecutionSample, jdk.MethodTrace, jdk.ClassLoad"},
 "scope":{"event_kinds":["loaded","executed","call"],"classes":"artifacts","method_filters":["…"]},
 "sampling":{"kind":"periodic","interval_ms":10},
 "completeness":{"complete":false,"lost_events":0,"truncated_stacks":3,"reasons":["…"]},
 "data_policy":"identifiers-only"}
```

| Field | Meaning |
|---|---|
| `run_id`, `scenario` | who ran what: a free-form scenario description from the person recording (the importer does not interpret it) |
| `started`, `ended` | the recorded time range (RFC 3339) |
| `runtime` | the JVM that ran the application |
| `recorder` | the tool that produced the file and its version, and where its data came from |
| `scope.event_kinds` | which of `loaded`, `executed`, `call`, `handoff` the recorder could produce at all — a kind not listed is *not measured*, not "none happened" |
| `scope.classes` | `artifacts` (only classes found in the listed artifacts are recorded) or `all` |
| `scope.method_filters` | methods recorded exactly on entry (`executed` with basis `traced`), if any |
| `sampling` | `none` (every event in scope recorded) or `periodic` with `interval_ms` |
| `completeness` | `complete` is `true` only if the recorder knows it lost nothing in scope; `lost_events` and `truncated_stacks` count what it knows it lost; `reasons` say why the run is incomplete (sampling, truncated stacks, a filter, a crash before the end) |
| `data_policy` | always `identifiers-only`: no argument values, payloads, field contents or request data |

### Records

| `record` | Fields | Meaning |
|---|---|---|
| `artifact` | `id`, `name`, `sha256` | a file the run's classes came from (a jar or class directory archive) |
| `loader` | `id`, `name`, `type` | a class loader of this run |
| `class` | `id`, `name` (binary name, dotted), `loader`, optional `artifact`, optional `sha256` (of the class file bytes), `source_basis` | a class definition: the same name under two loaders is two `class` records. `source_basis` is `observed` (the recorder saw the bytes the JVM defined) or `classpath-lookup` (the recorder found a class of that name in the listed artifact; the JVM may have defined other bytes) |
| `method` | `id`, `class`, `name`, `desc` (JVM descriptor) | a method of a recorded class |
| `loaded` | `id`, `class` | the class was loaded in this run |
| `executed` | `id`, `method`, `basis`, `count` | the method was seen executing: `traced` = `count` exact entries were recorded by instrumentation; `sampled` = it was on the stack in `count` samples (any frame, not only the top) |
| `call` | `id`, `caller`, `callee`, `basis`, `source`, `count`, optional `bci`, optional `line` | a measured relation: `stack-adjacent` = the caller frame was directly above the callee frame; `stack-through-runtime` = only frames outside the recorded scope (JDK, reflection, proxies) were between them. `source` is `sampled` or `traced`. `bci`/`line` are the caller frame's bytecode index and line, when the recorder had them |
| `handoff` | `id`, `from`, `to`, `mechanism`, `basis` | a task handed from one method to another thread (executor, thread start); only where the recorder correlated both ends, never from timing |

Record ids (`id` of `loaded`/`executed`/`call`/`handoff`) are the **evidence ids** results
refer to.

## Correlation with the analyzed input

For every `class` record the importer decides one status:

| Status | When | Used as evidence |
|---|---|---|
| `matched` | an analyzed class of that name exists and its class-file SHA-256 equals the record's `sha256` | yes |
| `matched_artifact` | the record has no class hash, its artifact's SHA-256 equals that of the jar the analyzed class was read from | yes |
| `different_version` | the class (or its artifact) hash differs from the analyzed one | no: another build of the code |
| `name_only` | the record carries no hash at all | no: shown, not attached |
| `not_in_input` | no analyzed class of that name | no |

Two `class` records with the same name (two loaders) are correlated separately and reported
as separate definitions; neither is merged into the other.

A `method` of a usable class is matched to the call-graph node with the same owner, name
and descriptor. A `call` between two matched methods is reported as **in the static graph**
(an edge exists, with its basis) or **not in the static graph** — a relation the static
analysis missed for this run, worth investigating, never an error of the trace by itself.

## Producing a trace

`scripts/JfrToTrace.java` converts a Java Flight Recorder recording (JDK 11+ for
`jdk.ExecutionSample`/`jdk.ClassLoad`; JDK 25+ for exact entries with `jdk.MethodTrace`,
JEP 520) into this format, recording only classes found in the artifacts it is given. JFR
does not report where a class's bytes came from, so its classes have `source_basis`
`classpath-lookup`. See the README for the commands. No Java agent is needed; an agent
(RUNTIME_NATIVE N3) would be another recorder of the same format.
