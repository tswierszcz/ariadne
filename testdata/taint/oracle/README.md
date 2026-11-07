# Taint real-JVM oracle (DATA_FLOW D5, ROADMAP B4)

| File | What |
|---|---|
| `src/oracle/Canary.java` | the marker sent as request data (a harmless `Serializable` class) |
| `src/oracle/TaintOracle.java` | the driver: calls every stub-`@PostMapping` handler of `demo.Endpoints`, `demo.Extra`, `demo.Shapes` (`../src`) with a serialized Canary (Base64 `String`, a `FakeRequest` returning it, or bytes), then `Extra#fromField`; a process-wide `ObjectInputFilter` factory records each entry that read a Canary |
| `oracle.jar` | the compiled driver (built against `../app.jar`) |
| `observed.tsv` | entry (`owner#name+desc`) → `read`, or `rejected` when the entry's own filter refused the Canary |

Regenerate with `scripts/taint-oracle.sh` (needs a JDK 17+ with the `jdk.compiler`
module; also rebuilds `../app.jar`). `internal/taint` `TestTaintOracle` compares the
model with `observed.tsv` without a JDK: no observed entry may be `no_flow_in_model`, every
`flow_found` must be observed, and the observed-but-`undetermined` entries are pinned.
Recorded on OpenJDK 25. An entry not observed here is not proven unreachable by request
data: the oracle covers these calls only. Results: `docs/TAINT_QUALIFICATION.MD`.
