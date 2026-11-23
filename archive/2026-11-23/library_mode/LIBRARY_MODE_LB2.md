# LIBRARY_MODE:LB2 — DONE for this scope (2026-11-23)

## Zapis ukończenia

- **Decyzja użytkownika (2026-11-23)** po ocenie (PR #139): zamknąć.
- **Zrobione:** `lb2-1`…`lb2-9` — luka kodu klienta (`client_code`, `client_dispatch`, `supplied_dispatch`), metody odziedziczone jako API ze specjalizacją odbiorcy, korzenie z ogłoszeń Spring Boot, CDI/Jandex, frameworka i Micronauta (3.x) z pochodzeniem; wyrocznie JVM `testdata/library*`.
- **Odłożone:** ścieżki warunkowe po powrocie z kodu klienta (semantyka niezdefiniowana); wartości dostarczone przekazane do metod zależności i `@ConditionalOn…` (zmierzone bez zysku); Micronaut 4.x; zapis wykonania listenera klienta w wyroczni (d).
- **Weryfikacja:** wpisy LIBRARY_MODE:LB2 w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.

## Tekst fazy z planu

## LB2. Kod klienta i wejścia frameworka

**Ocena 2026-11-23:** kryterium gotowości spełnione (`lb2-1`…`lb2-9`); otwarte tylko ścieżki warunkowe po kodzie klienta (semantyka do zdefiniowania, brak zmierzonego przypadku). Rekomendacja: zamknąć jako DONE for this scope. Zob. IMPLEMENTATION_STATUS.

Postęp: **lb2-1 DONE (2026-11-17)** — miejsca kodu klienta (`libapi.ClientSites`), luka
`client_code` (pusty zbiór kandydatów, bez `MODEL_INCOMPLETE`), `exposed_api[].client_dispatch`
dla metod sięgających kodu tylko przez wysyłkę do klasy, której wejście nie tworzy; przypadek (d)
fixture. Zostaje: ścieżki warunkowe po powrocie z kodu klienta, wywołania na typach zależności,
autokonfiguracja jako korzenie, wyrocznia z przykładowym klientem dla (d) poza `observed.txt`.
Pomiar 2026-11-17 (3 biblioteki, `scripts/library-measure.sh`): werdykty jak `-roots all`;
poprawka `lb2-2` (klasy konstruowalne przez klienta nie są zastępstwem); `lb2-3` (decyzja
użytkownika): metody odziedziczone z nadtypów zależności są API (jackson-yaml: 1 → 17
`REACHABLE`); `lb2-4`: korzeń odziedziczony to kopia per klasa API, wywołania na `this`
rozwiązywane od tej klasy (jackson-yaml: −1007 krawędzi CHA, `exposed_api` CVE-2025-49128
101 → 87, werdykty bez zmian). Zostaje dla odziedziczonych: odbiorca w parametrze (`this`
przekazany jako argument); odbiorca w polu nie jest tu nadmiarem — `ObjectMapper#_jsonFactory`
to w JVM także `MappingJsonFactory` (`new YAMLMapper(null)`, sprawdzone 2026-11-18), więc reguła
typu pola z konstruktorów nie dałaby na jackson zysku (punkt odłożony). **`lb2-5` DONE (2026-11-17):**
klasy ogłaszane przez bibliotekę w `META-INF/spring.factories` i `META-INF/spring/*.imports`
(z klasami zagnieżdżonymi `@Configuration` i `@Import`) są korzeniami z pochodzeniem
(`exposed_api[].announced_by`), niezależnie od dostępu; fixture `testdata/library-autoconfig`
z wyrocznią Spring Boot 2.4.5. **`lb2-6` DONE (2026-11-18):** beany archiwum CDI
(`META-INF/beans.xml` albo indeks Jandex — Quarkus) są korzeniami z pochodzeniem, niezależnie od
dostępu (producenci, obserwatorzy, `@Inject`, `@PostConstruct`…); fixture `testdata/library-cdi`
z wyrocznią Weld SE 5.1.2. **`lb2-7` DONE (2026-11-19):** wywołania na obiektach typu zależności dostarczonych przez
wywołującego (parametr korzenia, pole ustawiane tylko przez korzenie spoza biblioteki) są luką
`client_code` ze znacznikiem `supplied_dispatch`; werdykty bez zmian. Zostaje z LB2: ścieżki warunkowe po powrocie z kodu klienta (do
zdefiniowania), wartości dostarczone przekazane do metod zależności — **zmierzone 2026-11-21, nie budowane** (5 bibliotek: 0–107 miejsc, żaden werdykt ani powód by się nie zmienił; zob. IMPLEMENTATION_STATUS), Micronaut — **`lb2-9` DONE (2026-11-20):** definicje beanów z listy `BeanDefinitionReference` są ogłoszeniami, a to, co ich wygenerowany kod woła wprost, korzeniami (fixture `testdata/library-micronaut` z wyrocznią JVM; prywatne metody beanów przez refleksję — nie znajdowane; 4.x niesprawdzony), warunki `@ConditionalOn…` (pomiar 2026-11-20 na starterze mybatis: 0 z 26 ustaleń zależy od warunku — modelu nie budujemy bez biblioteki, na której to coś zmienia).

Zależności: LB1, ASYNC_CALLBACKS (rejestracje callbacków), FRAMEWORKS (konfiguracja).

- Miejsca, gdzie biblioteka woła metodę, którą może dostarczyć klient: wywołania
  interfejsów i metod abstrakcyjnych bez implementacji w wejściu, wywołania nadpisywalnych
  metod (publiczne/chronione, nie `final`, w klasie nie `final`) na obiektach, które
  klient może podać. Luka „kod klienta” (relevant per B0), bez wymyślonych krawędzi.
- Ścieżki warunkowe: podatny kod osiągalny dopiero po powrocie z kodu klienta — oznaczone
  w dowodzie, nie mieszane ze ścieżkami bezwarunkowymi.
- Wejścia z autokonfiguracji frameworków ogłaszanej przez bibliotekę (Spring Boot
  `spring.factories`/`AutoConfiguration.imports`, inne katalogi profili) jako korzenie
  z pochodzeniem.
- Weryfikacja: przypadek (d) fixture; wyrocznia z przykładowym klientem.

Gotowe: callback klienta jest jawną luką, autokonfiguracja biblioteki — korzeniem.
