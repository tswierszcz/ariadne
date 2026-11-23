# LIBRARY_MODE:LB4 — DONE for this scope (2026-11-23)

## Zapis ukończenia

- **Decyzja użytkownika (2026-11-23)** po ocenie (PR #139): zamknąć.
- **Zrobione:** `lb4-1` (zestaw 5 bibliotek z ręcznymi oczekiwaniami, wyrocznia jetty-client, README „Library mode in a library's CI”, SUPPORT.MD `Partial`), `lb4-2` (wyrocznie JVM pozostałych bibliotek, 0 `NOT_REACHABLE` wykonanych), zamknięcie krawędzi przekazania asynchronicznego (2026-11-21).
- **Przeniesione:** powiązanie z CI floty → `RECOMMENDATION/FLEET_CI_PLAN.MD` (decyzja w połowie grudnia 2026).
- **Weryfikacja:** wpisy LIBRARY_MODE:LB4 w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.

## Tekst fazy z planu

## LB4. Kwalifikacja i użycie w CI

**Ocena 2026-11-23:** kryterium gotowości spełnione (`lb4-1`, `lb4-2`); powiązanie z CI floty należy do `FLEET_CI_PLAN.MD` (grudzień). Rekomendacja: zamknąć jako DONE for this scope. Zob. IMPLEMENTATION_STATUS.

Zależności: LB1–LB3, ROADMAP B4.

- Zestaw prawdziwych bibliotek (różne style API: fasada, builder, SPI, callbacki) z
  ręcznie sprawdzonymi oczekiwaniami; pomiar nadmiarowych `REACHABLE` vs `-roots all`.
- Dokumentacja użycia: uruchomienie na artefakcie biblioteki w jej CI, interpretacja
  `exposed_api`, granice (tryb nie ocenia konkretnej aplikacji).
- Powiązanie z planem CI floty (`FLEET_CI_PLAN.MD`), jeśli flota obejmie biblioteki.

Gotowe: SUPPORT.MD opisuje tryb ze stanem `Partial` i zmierzonymi ograniczeniami.

**`lb4-1` DONE (2026-11-19):** zestaw pięciu bibliotek (fasada, fabryka/podklasa, SPI, starter,
callbacki — jetty-client 9.4.36) z ręcznie sprawdzonymi werdyktami
(`testdata/library-lb4/EXPECTED.MD`), wyrocznia JVM dla jetty-client
(`scripts/library-lb4-oracle.sh`, `TestLibraryJetty`), `-roots api` vs `-roots all` (żaden
`REACHABLE` nie stracony), README „Library mode in a library's CI”, SUPPORT.MD ze zmierzonymi
ograniczeniami. Wyrocznia startera mybatis (2026-11-19, `testdata/library-starter`): z 26 `REACHABLE` 5 wykonuje się w najmniejszej aplikacji, 7 więcej (snakeyaml) przez sam Boot przy `application.yml`, 14 nigdy; niewykonane wychodzą z biblioteki wywołaniem na obiekcie z kontenera (`ObjectProvider`, `BeanFactory`) — argument za LB2 „wywołania na typach zależności”. **`lb4-2` DONE (2026-11-19):** wyrocznie JVM dla commons-configuration2, jackson-dataformat-yaml,
jjwt-jackson (`scripts/library-lb4-jvm-oracle.sh`): 2/2, 15/17, 14/15 `REACHABLE` wykonane, 0
`NOT_REACHABLE` wykonanych. **Krawędź przekazania asynchronicznego — zamknięta (pomiar
2026-11-21):** na `main` po `lb2-9` graf jetty-client ma prawdziwą ścieżkę `HttpClient#GET` → selektor
NIO → `ReadCallback#succeeded` → parser (krawędzie `a3-2`–`a3-4`), CHA i RTA; zob. IMPLEMENTATION_STATUS.
Zostaje z LB4: powiązanie z planem CI floty (przesunięte na połowę grudnia 2026).
