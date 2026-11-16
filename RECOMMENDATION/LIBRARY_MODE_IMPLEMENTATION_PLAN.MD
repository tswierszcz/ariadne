# Tryb biblioteki — fazy LB0–LB4

Status: plan (2026-11-16), żadna faza nierozpoczęta; zob. IMPLEMENTATION_STATUS. Wymagania: [LIBRARY_MODE.MD](LIBRARY_MODE.MD). Zasady: [ROADMAP.MD](ROADMAP.MD) (kontrakt wyniku B0, dowody krawędzi B1, weryfikacja B4).

Oznaczenia faz `LB` (litera `L` należy do DYNAMIC_LINKAGE).

## LB0. Kontrakt trybu

Zależności: CVE F0–F1 (wynik v2), ROADMAP B0.

- Flaga trybu (propozycja: `-mode library`; `-mode application` jako dotychczasowe zachowanie,
  domyślne). Wejście jak dziś: jar biblioteki w `-app`, zależności w `-deps`/`-deps-list`.
- Definicja publicznego API (LIBRARY_MODE.MD, „Korzenie”), w tym: klasy zagnieżdżone,
  metody mostkowe i syntetyczne (nie są API, ale ich cele tak), `<clinit>`, scenariusz
  JPMS (`exports`) vs classpath, opcja wyłączenia pakietów wzorcem.
- Kontrakt wyniku v2 (addytywny): na ustaleniu lista `exposed_api` — metody API, z których
  osiągalny jest mapowany kod, każda z charakterem i przykładową ścieżką; licznik metod
  API w `coverage`; `analysis.mode` w nagłówku. Znaczenie statusów w trybie biblioteki
  opisane w `docs/RESULTS_CONTRACT.MD`. v1 bez zmian znaczenia.
- Nazwa luki „kod klienta” w słowniku `internal/gaps` (używana od LB2).
- Biblioteka kontrolna (fixture): podatny kod z commons-text (`StringSubstitutor#replace`,
  CVE-2022-42889) osiągalny (a) z metody publicznej, (b) tylko z prywatnej, (c) tylko z
  pakietowej klasy, (d) tylko przez callback do kodu klienta, (e) z dostawcy
  `META-INF/services`. Oczekiwania zapisane przed implementacją.

Gotowe: kontrakt w `docs/RESULTS_CONTRACT.MD`, fixture z oczekiwaniami, decyzje o nazwach.

## LB1. Korzenie z publicznego API i narażone metody

Zależności: LB0.

- Wybór korzeni z hierarchii (modyfikatory klasy i metody, zagnieżdżenie), wejścia
  `META-INF/services` biblioteki. Log linia: liczba metod API, wyłączone pakiety.
- Narażone metody API: dla każdego mapowanego węzła wszystkie korzenie API, z których jest
  osiągalny (wsteczne przejście grafu od mapowanych metod, przecięte z korzeniami), z
  deterministyczną kolejnością i ścieżką przykładową z istniejącego BFS.
- Raport Markdown: sekcja „Exposed API” per advisory; v2 `exposed_api`.
- Weryfikacja: fixture LB0 — (a) i (e) `REACHABLE` z właściwymi metodami API, (b) i (c)
  `NOT_REACHABLE`; wyrocznia JVM: mały sterownik wołający każdą metodę API po kolei
  (refleksyjnie, z pliku listy) i śledzenie JDI (`scripts/MethodTrace.java`) potwierdza,
  że z każdej zgłoszonej metody API podatny kod się wykonuje, gdy argumenty na to
  pozwalają. Pomiar na 2–3 prawdziwych bibliotekach z podatnymi zależnościami (wybór w
  ramach fazy) — liczba ustaleń i narażonych metod vs `-roots all`.

Gotowe: tryb działa end-to-end na fixture i prawdziwej bibliotece; v1 aplikacji bajtowo
bez zmian (12 konfiguracji referencyjnych).

## LB2. Kod klienta i wejścia frameworka

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

## LB3. Wersje zależności

Zależności: LB1, CVE F4 (wersjonowane mapowania).

- Odczyt deklaracji zależności biblioteki (pom: zakresy wersji, `provided`, `optional`)
  i oznaczenie instancji dostarczanych przez użytkownika.
- Opcjonalnie: ocena advisory dla wersji z zakresu, które użytkownik może dostać
  (np. najnowsze w deklarowanej linii), jako osobne ustalenia z jawną wersją;
  `MAPPING_VERSION_UNCOVERED`, gdy mapowanie nie ma wariantu.
- Weryfikacja: fixture z zakresem wersji i zależnością `provided`.

Gotowe: wynik mówi, której wersji zależności dotyczy każde ustalenie i kto ją dostarcza.

## LB4. Kwalifikacja i użycie w CI

Zależności: LB1–LB3, ROADMAP B4.

- Zestaw prawdziwych bibliotek (różne style API: fasada, builder, SPI, callbacki) z
  ręcznie sprawdzonymi oczekiwaniami; pomiar nadmiarowych `REACHABLE` vs `-roots all`.
- Dokumentacja użycia: uruchomienie na artefakcie biblioteki w jej CI, interpretacja
  `exposed_api`, granice (tryb nie ocenia konkretnej aplikacji).
- Powiązanie z planem CI floty (`FLEET_CI_PLAN.MD`), jeśli flota obejmie biblioteki.

Gotowe: SUPPORT.MD opisuje tryb ze stanem `Partial` i zmierzonymi ograniczeniami.

## Kolejność i szacunek

LB0 i LB1 to pierwsze wydanie (użyteczne samo w sobie); LB2 i LB3 po pomiarze LB1; LB4
na końcu. LB1 dotyka głównie wyboru korzeni i nowego pola wyniku — bez zmian w budowie
grafu i modelach.
