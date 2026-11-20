# ASYNC_CALLBACKS (A0–A4) — completion record

**Archived:** 2026-11-20. **Status: DONE** for every phase's declared scope. The
`docs/SUPPORT.MD` row "General callbacks and asynchronous execution models" stays
**Partial**: the remaining limitations below are real and stated there.

This directory holds the area's two former live documents (`ASYNC_CALLBACKS.MD`,
`ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD`). Detailed per-increment notes (scope, rule,
files, commands, numbers) stay in `RECOMMENDATION/IMPLEMENTATION_STATUS.MD` (entries
"ASYNC_CALLBACKS:A3 …" and "ASYNC_CALLBACKS:A4 …"). Current behavior: `docs/SUPPORT.MD`
(the row above), README ("Callbacks and async execution", "What an asynchronous edge
means"), `testdata/async-qualify/RESULTS.MD`.

## What shipped

| Phase | What it did | Commits / PR |
|---|---|---|
| A0–A2 | Registration-site catalog (`Thread`, `Executor`/`ExecutorService`, `doPrivileged`), callback identity via DATA_FLOW, "may run" `callback` edges (`internal/callback` catalog, sites, identity, edges) | before 2026-11-14; JVM oracle `testdata/async` — #61 |
| A3 slice 1 | `CompletionStage`/`CompletableFuture` continuations, `exceptional` condition | `022147d` (#80) |
| `a3-2` | Callback parameters resolved at the registering method's callers (`forward.go`) | `25cbe22` (#109) |
| `a3-3` | Hand-offs through a queue field and an NIO key attachment (`handoff.go`); edge oracle `scripts/MethodEdgeTrace.java`, `scripts/real-edge-paths.py` | `1252cf7` (#110) |
| `a3-4` | Tasks a method returns to the code running them (`returned.go`; gate `async-callbacks:returned-tasks`) | `3539ef8` (#111) |
| `a3-5` | Listeners: field lists, `PropertyChangeSupport` (`listeners.go`; gate `async-callbacks:listeners`) | `77e55c6` (#112) |
| `a3-6` | Reactive streams: JDK `Flow`/`SubmissionPublisher` (`reactive.go`; gate `async-callbacks:reactive`) | `c393d9a` (#113) |
| `a3-7` | Completion of hand-completed `CompletableFuture`s (`completion.go`; gate `async-callbacks:completion`) | `6d9e308` (#114) |
| `a4-1` | `Thread` subclasses' `start()`, `RejectedExecutionHandler` (`threads.go`; gate `async-callbacks:rejection`) | `712a549` (#115) |
| `a4-2` | Qualification: whole pipeline on all async fixtures vs JVM oracles | `b8e6af4` (#116) |

## Evidence

- **JVM oracles per slice:** `testdata/{async,continuations,async-forward,async-handoff,
  async-returned,async-listeners,async-reactive,async-completion,async-threads}` with
  `scripts/*-oracle.sh`; each Go test checks the model's edges and reachability after
  its gate against `observed.txt`, with a mutation check.
- **Qualification (`TestAsyncQualification`, `testdata/async-qualify/RESULTS.MD`):** 166
  bodies ran across nine fixtures, 1 not reached (a `MethodHandle` name chosen at run
  time — DYNAMIC_LINKAGE's gap); 60 `callback` edges from reached methods, 6 to bodies
  the oracles did not record (2 ran unmarked, 4 documented may-run cases), no wrong
  association. `TestAsyncTwoTasks`: two tasks at two sites never mix.
- **Measurement per slice** (`a3-3` … `a4-1`): five libraries × `-roots all/api` × CHA/RTA
  (20 runs) and petclinic/WebGoat/JPetStore × `-roots profile/all` × CHA/RTA (12
  configurations, `-graph-compare`): no verdict or `exposed_api` change in any slice; no
  reference-app example path change. jetty-client: `HttpParser#convertContentLength`
  reached over JVM-taken edges from 36 of 103 executed API entries (3 before `a3-4`).

## Remaining limitations (stated in `docs/SUPPORT.MD`)

- An asynchronous edge is "may run": thread identity, interleavings, scheduler and
  cancellation policy, reactive demand and memory visibility are not modeled.
- Completion is decided only for hand-completed `new CompletableFuture`s; derived stages
  (`supplyAsync`, continuations) keep their may-run edge.
- Reactive: JDK `Flow` only (Reactor/RxJava need access to their artifacts); input
  `Flow.Publisher` implementations not modeled.
- Listeners: arrays, lists walked through a parameter or a getter (Jetty's
  `ResponseNotifier`), `forEach` with a lambda, `Observable`, removal.
- Hand-offs: `AtomicReference`, drains through a helper, non-JDK collections, a returned
  parameter.
- Example paths are the shortest route (BFS), not the executed one: 18 of 408 jetty-client
  paths are real edge by edge.
- Spring's `ListenableFuture` not modeled.
