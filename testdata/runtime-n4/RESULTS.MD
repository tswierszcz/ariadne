# RUNTIME_NATIVE N3–N4 assessment probe (2026-11-25)

This probe checks what the agent-free recorder (N1: JFR on JDK 25 with `jdk.MethodTrace`, JEP 520, converted by `scripts/JfrToTrace.java`) measures for the three N4 mechanisms. `src/n4/Main.java` contains:

- `Method#invoke` of `target` in a loop on one line,
- two JDK proxies created by `proxy("p1")` and `proxy("p2")`, whose handler forwards to `Real#op`,
- one task class `Task`, submitted to an executor from two methods (`submitA`, `submitB`), whose `run` calls `work`.

The probe is not part of the tests; it is the measurement behind the assessment in `RECOMMENDATION/RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD`. To reproduce it on JDK 25+:

```sh
java scripts/JavaBuild.java testdata/runtime-n4/src /tmp/n4.jar 21
java -XX:StartFlightRecording=filename=/tmp/rec.jfr,settings=profile,jdk.MethodTrace#enabled=true,'jdk.MethodTrace#filter=n4.Main::target;n4.Main$Real::op;n4.Main::work' -cp /tmp/n4.jar n4.Main
java scripts/JfrToTrace.java --jfr /tmp/rec.jfr --artifact /tmp/n4.jar --scenario probe --out /tmp/run.trace
```

Measured on OpenJDK 25.0.5: 7 method traces, 0 truncated stacks.

| Mechanism | What the trace has | What it does not have |
|---|---|---|
| Reflection | `main` (line 29, the `invoke` call) → `target`, `stack-through-runtime`, `traced` ×3: the actual reflective target together with its call site | — |
| Proxy | the handler `lambda$proxy$0` (line 21) → `Real#op`, `stack-through-runtime`, `traced` ×2 | which proxy call site (`main` line 31 or 32) led to the handler. JFR has the frames, but the converter keeps only the nearest in-scope frame of a traced event. |
| Async | `Task#run` (line 14) → `work`, `stack-adjacent`, `traced` ×2 | which registration (`submitA` or `submitB`) the execution came from: the worker thread's stack does not contain the submitter, and JFR carries no token across threads |
