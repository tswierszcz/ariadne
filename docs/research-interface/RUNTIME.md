# Runtime evidence for research targets (RI6, experimental)

Stage RI6 of the [research interface plan](../../RECOMMENDATION/RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD) and RH-A4 of the [harness plan](../../RECOMMENDATION/RESEARCH_HARNESS_IMPLEMENTATION_PLAN.MD), first written 2026-11-25.

These are two commands, **outside contract `1.0.0-draft.1`**. CONTRACT.md §13 names `trace-config` and `correlate` as later operations, so they are not in the request schema, not in a bundle's `capabilities`, and not in the pinned `v1/` package. Their documents carry their own schema identifiers. A later contract revision may adopt them after the consumer has used them.

The trace format is the existing [`ariadne-runtime-trace-1`](../RUNTIME_TRACE.MD), produced from a JFR recording by `scripts/JfrToTrace.java`.

## 1. `research trace-config`

```sh
ariadne research trace-config -bundle BUNDLE -request TARGETS.json -out CONFIG.json -jfc SETTINGS.jfc
```

**Input.** The request is `ariadne-research-trace-request-1`:

```json
{"schema_version": "ariadne-research-trace-request-1", "analysis_id": "sha256:…", "request_id": "run-17",
 "targets": [{"id": "write", "selector": {"instance_id": "c-…", "class": "px.archive.internal.Store", "name": "write", "descriptor": "(Ljava/nio/file/Path;[B)V"}}]}
```

Targets have the shape of reachability targets: a `method_id` or a selector, 1–256 of them. They are checked and resolved by the same rules, so `matched`, `missing`, `shadowed` and `ambiguous` mean the same as in `reachability`.

**Output.** Two files:

- **`CONFIG.json`** (`ariadne-research-trace-config-1`) records the analysis id, the request id, the rule version `research-runtime-1`, the sorted list of filters and `minimum_jdk` 25. Each target has its `resolution`, `matched_node_ids`, its `filter` (`Class::method`, or null when the target is not resolved) and `also_traced_node_ids`.
- **`SETTINGS.jfc`** holds JFR settings that enable `jdk.MethodTrace` with a stack trace and threshold 0 for those filters. Record with `-XX:StartFlightRecording=settings=profile,settings=SETTINGS.jfc` on JDK 25 or later, then convert the recording with `scripts/JfrToTrace.java`.

**Overloads.** A JFR filter names a method, not a descriptor, so every overload of a target is traced as well. `also_traced_node_ids` lists those overloads; their entries count for them, not for the target.

## 2. `research correlate`

```sh
ariadne research correlate -bundle BUNDLE -trace RUN.trace [-config CONFIG.json] -out CORRELATION.json
```

The output is `ariadne-research-correlation-1`. The whole bundle is loaded and verified. A configuration made for another analysis is refused.

### Identity

A class of the trace attaches to the bundle **only by content**: it needs the same name and the same SHA-256 as a bundle class record. That class record's instance then selects the method nodes.

- A class whose name is in the bundle but whose bytes differ (another build, or a transformed class) attaches nothing. It is counted in `classes.hash_differs` and named in `warnings`.
- A class not in the bundle is counted in `classes.not_in_bundle`; harness and client classes are typical examples.
- Evidence attaches to the canonical method node. A receiver specialization is the same method at run time: the run cannot tell specializations apart, so an edge from or to a specialization counts as the bundle's edge.

### Output members

| Member | Content |
|---|---|
| `trace` | the trace's SHA-256, `run_id`, scenario, runtime, recorder, sampling, completeness (lost events, truncated stacks, reasons) and the method filters the run was recorded with |
| `classes` | `matched`, `hash_differs`, `not_in_bundle` |
| `calls` | the trace's caller → callee relations: `in_bundle` (both ends matched and the bundle has an edge), `not_in_bundle` (both ends matched, no edge: a gap of the static graph for this run), `unmatched` (an end is not a matched class) |
| `executed` | per node: `traced` entries and `sampled` stack counts |
| `targets` | with `-config`, per target: `class_identity` (`matched`, `differs`, `not_seen`), `traced` (a filter of the run covers it), `observation`, `entries`, and `callers` |

Each entry of `callers` gives the caller node (null when its class is not matched), the recorded caller method, the bundle's `edge_id` for the step (or null), the basis, source, count, bci and line.

### Observation values

| `observation` | Meaning |
|---|---|
| `entered` | exact entries recorded |
| `sampled` | only seen on sampled stacks |
| `traced_not_entered` | the run traced the method and it was not entered **in this run** |
| `not_seen` | not traced and not seen |
| `class_differs` | the run's class has other bytes, so nothing is attached |
| `not_evaluated` | the target did not resolve |

**No observation is evidence of unreachability.** Another scenario may run the method. An entry shows that the code ran, not that attacker-controlled data reached it. The experiment's assertion decides the property (CONTRACT.md §13).

## 3. Measured on the pilot target

`scripts/pilot-runtime.sh` runs the whole path on [testdata/pilot](../../testdata/pilot/EXPECTED.MD) for each variant: index, trace-config, the oracle client under JFR with that filter, conversion and correlation. The traces are committed as `testdata/pilot/runtime-*.trace`, and `TestPilotRuntime` repeats the configuration and the correlation without a JDK.

| Target | defective | fixed |
|---|---|---|
| `Store#write` | `entered` 3 (`unpack` ×2, `writeFile` ×1; every step a bundle edge) | `entered` 2 (`unpack` ×1: the traversal entry is refused) |
| `Entries#target` | `entered` 2 | `entered` 2 |
| `Store#overwrite` (dead) | `traced_not_entered` | `traced_not_entered` |
| `Unpacker#filter` (reflection) | `traced_not_entered` | `traced_not_entered` |

The defective variant's trace correlated against the fixed bundle gives `Entries#target` `class_differs`, 1 class with other bytes, and a warning. The classes the two variants share stay matched.

## 4. Limits

- Exact entries need JDK 25 or later (`jdk.MethodTrace`, JEP 520). On older JDKs only sampled evidence exists.
- Sampled relations vary from run to run, by about 40% on petclinic ([TRACE_GAPS.MD](../../testdata/graphbench/TRACE_GAPS.MD)). Compare runs on traced evidence.
- JFR does not report where class bytes came from. `JfrToTrace` looks them up in the artifacts it is given, so identity is exactly as good as those artifacts.
- The registration site of an asynchronous task is not recorded; see RUNTIME_NATIVE N4 in [the archive](../../archive/2026-11-25/runtime_native/COMPLETION.md).
- The trace's scenario and run ids are the recorder's. Tekmerion's experiment id is its own record that refers to them; `correlate` does not invent a link between them.
