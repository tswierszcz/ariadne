# RUNTIME_NATIVE — obszar DONE for this scope (2026-11-25)

Plan i opis: [RUNTIME_NATIVE.MD](RUNTIME_NATIVE.MD), [RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD](RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD). Archiwum N0–N1: [archive/2026-11-08/runtime_native](../../2026-11-08/runtime_native/RUNTIME_NATIVE_N0_N1.md).

Obszar zamknięty 2026-11-25 z polecenia użytkownika („rób wszystko”). Wszystko, co zostało, czeka na konkretny cel. Poniżej zapisane jest, co uzasadnia wznowienie każdej części.

| Faza | Stan | Najważniejsze |
|---|---|---|
| N0 | DONE (2026-11-08) | kontrakt śladu `ariadne-runtime-trace-1` ([docs/RUNTIME_TRACE.MD](../../../docs/RUNTIME_TRACE.MD)): brak zdarzenia niczego nie dowodzi, kolejność w czasie nie jest przyczynowością |
| N1 | DONE (2026-11-08) | importer `-runtime-trace` (`n1-1`), rejestrator JFR bez agenta (`scripts/JfrToTrace.java`), `-runtime-filter-out`; przyczyny luk grafu `g2` ([TRACE_GAPS.MD](../../../testdata/graphbench/TRACE_GAPS.MD)) |
| N2 | odłożony (2026-11-23) | kontrakt śladu rozróżnia loadery, strona statyczna ma jedną mapę FQCN; żaden cel nie ma kilku kontekstów ładowania |
| N3 | NOT STARTED, świadomie niebudowany (2026-11-25) | MVP agenta daje JFR `jdk.MethodTrace` (JDK 25+) |
| N4 | DONE for this scope (2026-11-25) | refleksja pokryta; proxy pokryte od `n4-1` (pełne stosy `traced`, handler-lambda jako `proxy`); async: wykonanie widoczne, rejestracja nie |
| N5 | DONE for this scope (`n5-1`, 2026-11-23) | granice natywne jako luki `native`, inwentarz `load`/`loadLibrary`; streszczenia Java ↔ native nie budowane |

## Pomiary

- petclinic: luki grafu względem śladu 221 → 121 → 105/144 (stary/nowy konwerter, to samo nagranie). Po `n4-1`: 3 445 relacji, 95% nowych w grafie.
- Głębokość stosu JFR: 64 ramki wystarczają (256 ramek dodaje 5 relacji `traced`). Relacje `sampled` różnią się między przebiegami o ok. 40%.
- Sonda N4: [testdata/runtime-n4](../../../testdata/runtime-n4/RESULTS.MD).

## Wznowić, gdy

- **N2:** cel z kilkoma kontekstami ładowania (pluginy, serwer aplikacyjny z izolacją) albo z klasami transformowanymi lub ukrytymi, gdzie liczy się hash wersji.
- **N3/N4 async:** badany przebieg na JDK < 25; eksperyment (np. Tekmerion), który musi przypisać wykonanie zadania do miejsca rejestracji; potrzeba bajtów klas po transformacji. Agent będzie wtedy kolejnym rejestratorem formatu `ariadne-runtime-trace-1`.
- **N5, streszczenia:** podatność albo cel badania, w którym przepływ przez metodę `native` decyduje o werdykcie.

Dowody runtime dla dowolnych celów interfejsu badawczego to osobny tor: RI6 / RH-A4 w [RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD](../../../RECOMMENDATION/RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD). Używa tego samego formatu śladu.
