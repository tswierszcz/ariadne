# LIBRARY_MODE — obszar DONE (2026-11-23)

Plan i opis: [LIBRARY_MODE.MD](LIBRARY_MODE.MD), [LIBRARY_MODE_IMPLEMENTATION_PLAN.MD](LIBRARY_MODE_IMPLEMENTATION_PLAN.MD); archiwa faz: [LB2](LIBRARY_MODE_LB2.md), [LB4](LIBRARY_MODE_LB4.md).

| Faza | Stan | Najważniejsze |
|---|---|---|
| LB0 | DONE (2026-11-16) | kontrakt trybu (`docs/RESULTS_CONTRACT.MD` §8), fixture `testdata/library` |
| LB1 | DONE (2026-11-16) | `-roots api`: korzenie z publicznego API, `exposed_api` per ustalenie |
| LB2 | DONE for this scope (2026-11-23) | luka kodu klienta, metody odziedziczone, ogłoszenia Spring/CDI/Micronaut jako korzenie |
| LB3 | DONE for this scope (2026-11-23) | `lb3-1`: kto dostarcza zależność (`supplied_by`, `version_match`); `lb3-2`: `-alt-versions newest`, `-alt-versions-mode approx` — inne wersje, które użytkownik może dostać |
| LB4 | DONE for this scope (2026-11-23) | zestaw 5 bibliotek z wyroczniami JVM, README „Library mode in a library's CI” |

**Weryfikacja:** fixtures `testdata/library*`, wyrocznie JVM (`library-oracle.sh`, `library-lb4-*`, `library-deps-oracle.sh` z Maven 3.9), pomiary `scripts/library-measure.sh` — wpisy LIBRARY_MODE w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.

**Odłożone (przyszłość):** tryb `rerun` dla innych wersji zależności (flaga zarezerwowana), inne strategie wyboru wersji, ścieżki warunkowe po kodzie klienta, Micronaut 4.x; powiązanie z CI floty — `RECOMMENDATION/FLEET_CI_PLAN.MD`. Wiersz `docs/SUPPORT.MD` „Library mode” pozostaje `Partial`.
