# DESERIALIZATION Z4 — qualification

Run on 2026-11-07, offline, on the same `.cache` and inputs for both builds: `main` after
Z3a (`3243863`) and the Z4 branch. Every comparison below is on the same configuration
(ROADMAP B4). A change in the number of findings is not, by itself, evidence of an
improvement; none happened here.

## 1. Model vs the real JVM (callbacks)

`testdata/deserialize/oracle` ([README](oracle/README.md)): every scenario the plan names —
private callback, public look-alike, Externalizable, a special type without coverage (a
record), a class of unknown serializability, two instances of one class (plain and
relocated), and an evaluated filter — written and read back on OpenJDK 25.

| Result | Count | Detail |
|---|---|---|
| Model edge the JVM did not run | 0 | — |
| JVM callback the model does not connect | 2 | `sc.Point#<init>` (record canonical constructor, not modeled), `sc.Queue#readObject` (serializable only through `LinkedBlockingQueue`, outside the JDK table: unknown, counted) — both declared in `model-gaps.tsv` |
| Filter | agrees | `createFilter("sc.PrivateCb;!*")`: the JVM rejects every other class, the model connects only `PrivateCb`'s callbacks |
| Two instances | agrees | `lib.Cb#readObject` and `shaded.lib.Cb#readObject` both run and are both connected |

Checked by `TestDeserializationOracle` without a JDK.

## 2. Source-documented CVEs vs ariadne

Expectations are those of [cves.yaml](cves.yaml) (from the cited sources). "ariadne" is the
chain state of the copy belonging to the advisory's component, or the finding's
reachability for a callback case.

| Advisory | Component | Expected (source) | ariadne — vulnerable release | ariadne — fixed release | Agrees |
|---|---|---|---|---|---|
| CVE-2015-7501 | commons-collections 3.2.1 / 3.2.2 | candidate / blocked | `candidate_chain` at `Main#read`, `candidate_blocked` at the filtered entry | `candidate_blocked` (guard `FunctorUtils#checkUnsafeSerialization`) | yes |
| CVE-2015-7501 | commons-collections4 4.0 / 4.1 | candidate / blocked | `candidate_chain` | `candidate_blocked` (`InvokerTransformer` not Serializable) | yes |
| CVE-2015-4852 | commons-collections | candidate / blocked | not checked: the advisory's component is WebLogic; no WebLogic input | — | n/a |
| CVE-2015-3253 | groovy 2.4.3 / 2.4.4 | candidate only on JDK ≤ 8u71 / blocked | `conditions_unresolved` (no `-target-java`); `candidate_blocked` with `-target-java 17` | `candidate_blocked` (`MethodClosure#readResolve`) | yes |
| CVE-2013-7285 | XStream 1.4.6 / 1.4.7 | candidate / blocked | `candidate_chain` (entries `unrestricted`) | `candidate_blocked` (`ReflectionConverter.eventHandlerType`) | yes |
| CVE-2013-7285 | XStream 1.4.5 (WebGoat) | candidate | `candidate_chain` at `VulnerableComponentsLesson#completed` | — | yes |
| CVE-2023-6378 | logback-classic 1.2.3 | reachable only through a deserialization entry | `REACHABLE`: `Main#main → Main#read → LoggingEventVO#readObject` (edge basis `deserialization`) | — | yes |
| CVE-2021-46877 | jackson-databind 2.11.4 | reachable only through a deserialization entry | petclinic `FRAMEWORK` (Spring `DefaultDeserializer`), WebGoat `REACHABLE` (JRuby `JavaProxy#marshal_load`) | — | yes |

The real JVM confirms each guard (`scripts/prepare-deserialization-chains-demo.sh`: the
vulnerable release round-trips the guarded class, the fixed release refuses it). CommonsBeanutils1
has no fixed release; it stays `conditions_unresolved` on both builds without
`-target-java` and is `candidate_blocked` with `-target-java 17`, as its JDK note says.

## 3. Old vs new build (same inputs, same cache)

| Input | Verdicts | v2 apart from the Z4 fields | Two runs |
|---|---|---|---|
| petclinic (`-profile spring`) | identical (3/71/16/57/6) | identical | byte-identical |
| WebGoat (`-profile spring`) | identical (42/51/10/61/0) | identical | byte-identical |
| chains demo, vulnerable / fixed / vulnerable `-target-java 17` | identical | identical (tool version only) | — |
| deserialization demo | identical | identical (tool version only) | — |

Z4 fields: `deserialization` (model and chain catalog versions, entries with policy,
chains per copy with site states) and `findings[].deserialization_chains`. WebGoat links
CVE-2013-7285's finding to `xstream:eventhandler candidate_chain`; petclinic has no
linked finding (every Java chain `classes_absent`, the XStream chain `no_entry`). The v2
file grows by 3 % (petclinic 1.71 → 1.76 MB, WebGoat 1.95 → 2.01 MB).

## 4. What this does not show

No data-flow analysis (DATA_FLOW:D4): no result says the bytes read at an entry come from
an attacker. The chain catalog has five chains; an entry without a catalog chain is not
free of chains. The oracle shows agreement on its scenarios, not the absence of other
callbacks.
