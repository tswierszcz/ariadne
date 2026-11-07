# Deserialization real-JVM oracle (DESERIALIZATION Z4, ROADMAP B4)

Harmless fixtures for the callbacks Java serialization runs, with the
real JVM's behaviour recorded:

| File | What |
|---|---|
| `src/sc/*` → `scenarios.jar` | entries `sc.Entries#plain` and `#filtered` (`createFilter("sc.PrivateCb;!*")`); scenario classes: `PrivateCb` (private `readObject`, `readResolve`), `PublicCb` (public `readObject` — not a callback), `Ext` (Externalizable), `Point` (record), `Queue` (serializable only through `LinkedBlockingQueue`, outside ariadne's JDK table), `Oracle` (the driver) |
| `shaded-src` → `lib.jar`, `shaded.jar` | `lib.Cb` and its relocated copy `shaded.lib.Cb`: two instances of one library class |
| `observed.tsv` | per entry and object: the callbacks the real JVM ran, in order, or REJECTED |
| `model-gaps.tsv` | callbacks the JVM ran that the model does not connect — each a declared gap |
| `model-extras.tsv` | callbacks the model connects that the JVM did not run — each with a reason |

Regenerate with `scripts/deserialize-oracle.sh` (needs a JDK with the
`jdk.compiler` module). `internal/deserialize` `TestDeserializationOracle`
compares ariadne's model with these files without a JDK; both lists must
match exactly. Recorded on OpenJDK 25: no extra edge, two declared gaps
(the record's canonical constructor, the unknown-serializability class).
A callback not observed here is not proven never to run: the oracle shows
the model's agreement on these scenarios, nothing beyond them.
