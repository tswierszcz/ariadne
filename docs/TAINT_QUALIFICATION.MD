# Taint pilot — qualification (DATA_FLOW D5)

What the `-taint` mode (DATA_FLOW D4, `internal/taint`, catalog `taint-1`) was measured
against, and what it got right and wrong, on 2026-11-07. It applies to **this pilot only**:
request data (the source catalog in `docs/SUPPORT.MD`, row "Data flow and taint") reaching
the data a deserialization entry reads (Java serialization `readObject`/`readUnshared`,
XStream `fromXML`). No global accuracy figure is derived from it (ROADMAP B4): the fixture
is small and was written to exercise known shapes.

**Decision:** `-taint` stays **opt-in** (default off). The pilot has one sink family and a
narrow source catalog; turning it on would change every user's v2 document and report for
a result most runs do not use.

## 1. Real-JVM oracle (Java serialization)

`scripts/taint-oracle.sh` (OpenJDK 25) builds `testdata/taint/app.jar` and
`testdata/taint/oracle/oracle.jar`, then calls every stub-`@PostMapping` handler of the
fixture with a serialized `oracle.Canary` as request data (a Base64 `String`, a
`ServletRequest` returning it, or bytes). A process-wide `ObjectInputFilter` factory wraps
every stream's filter and records each entry (the method calling `ObjectInputStream`'s
read) that read a Canary — `read`, or `rejected` when the entry's own filter refused it.
`testdata/taint/oracle/observed.tsv` is committed; `internal/taint` `TestTaintOracle`
compares the model with it without a JDK.

| Model \ JVM | read the Canary | did not (or not exercised) |
|---|---|---|
| `flow_found` | **10** (one `rejected` by its own filter: guard `effective`) | **0** — no unconfirmed flow |
| `undetermined` | **4** — measured gaps (below) | 1 (`Endpoints#unknown`: a literal behind a two-implementation interface) |
| `no_flow_in_model` | **0** — no miss | 6 (constant, one helper called with a literal, two streams where the literal one is read, the `byte[]` overload fed a literal, two entries called only from `main`) |

21 entries in all. Observed flows the model cannot follow, each reported `undetermined` with
the reason as a gap, never as "no flow":

| Entry | Shape | Gap |
|---|---|---|
| `Codec#readDeep` | 9 helper levels | depth limit (8 calls) |
| `Codec#readOther` | handler stores the token in an instance field, another method reads it | flows through fields are not traced |
| `Loader#inCallback` | read inside a lambda's body | caller is a lambda: its arguments are not traced |
| `Shapes#exceptionMessage` | token carried in an exception's message | `Throwable#getMessage` is not modeled (its result is an unmodeled call) |

Shapes the plan names and how they come out (`TestShapes`): aliasing through a second
reference to a builder — `flow_found`; array elements — `flow_found`; locals into an
exception handler — `flow_found`; the exception object — `undetermined`; two instances (two
streams in one method, the literal one read) — `no_flow_in_model`; overload (`load(String)`
fed request data, `load(byte[])` a literal) — `flow_found` / `no_flow_in_model`; callback
(lambda) — `undetermined`; no sources (a batch class) — `no_flow_in_model`, its input listed
as outside the catalog.

A mutation check (source parameters not recognized) makes `TestTaintOracle` report five
misses. An entry the oracle did not reach is not proven safe: the oracle covers these calls
only.

## 2. XStream

No runtime oracle: XStream does not consult `ObjectInputFilter`. Checked against
independently verified examples instead:

- WebGoat `VulnerableComponentsLesson#completed` (XStream 1.4.5): the source passes
  `@RequestParam String payload` through `String#replace` to `xstream.fromXML(payload)` —
  the model reports `flow_found`, guard `unrelated`, and also for XStream's own inner
  `fromXML(Reader)` call.
- `scripts/prepare-deserialization-chains-demo.sh` (XStream 1.4.6 / 1.4.7): `Web#uploadXml`
  passes its `@RequestParam` to `fromXML` — `flow_found`; `Main#readXml`, called only from
  `main` with a literal — `no_flow_in_model`.

## 3. Real applications (offline, `-profile spring` / `generic`)

| Input | Entries | `-taint` result | Searches / inputs visited / limit stops | Search time | Wall time without / with | Peak RSS without / with |
|---|---|---|---|---|---|---|
| taint demo (`scripts/prepare-taint-demo.sh`) | 9 | 5 flow_found, 3 no_flow, 1 undetermined | 9 / 10 / 0 | 66 µs | 0.1 s / 0.1 s | — |
| chains demo, vulnerable | 98 | 3 flow_found, 2 no_flow, 2 undetermined, 91 not_assessed | 7 / 194 / 0 | 0.6 ms | 2.3 s / 2.4 s | — |
| chains demo, fixed | 104 | 3 flow_found, 2 no_flow, 2 undetermined, 97 not_assessed | 7 / 181 / 0 | 0.3 ms | 2.6 s / 2.5 s | — |
| petclinic | 130 | 130 not_assessed (library entries no application code calls) | 0 / 0 / 0 | — | 12.2–12.3 s / 12.3–13.3 s | 1348–1399 MB / 1367–1368 MB |
| WebGoat | 143 | 2 flow_found, 141 not_assessed | 2 / 81 / 0 | 0.3 ms | 23.9–24.0 s / 24.3–29.5 s | 1736–1769 MB / 1780–1784 MB |

Wall time and RSS are two runs each on one machine; the spread between runs is larger than
the difference the flag makes. The search itself costs well under a millisecond; the
additional dependency summaries of the callers (D3) are within the noise on these inputs
(WebGoat: same 2420 summarized methods). Verdicts are identical with and without `-taint`,
and two `-taint` runs give byte-identical v2 documents.

## 4. What this qualifies, and what it does not

Qualified for this pilot: the five D4 outcomes are told apart (request data, constant,
effective guard, unrelated guard, unknown fragment); no observed flow is reported as "no
flow"; per-call-site substitution keeps a helper's request and literal callers apart;
overloads are kept apart; cost is negligible next to the rest of the run.

Not qualified: other sinks (any CVE method other than deserialization entries), other sources
(files, sockets, message queues, `@ModelAttribute` beans), flows through fields (instance or
static), lambda/method-reference callers, callers outside application classes, control
dependence, heap or context sensitivity beyond per-call-site substitution, and thread
interleavings. `no_flow_in_model` is relative to the source catalog; `flow_found` is not an
exploit.

Regenerate: `scripts/taint-oracle.sh` (JDK 17+ with `jdk.compiler`), then `go test ./internal/taint`.
