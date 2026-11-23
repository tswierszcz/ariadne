# Tryb biblioteki — fazy LB0–LB4

Status: LB0 DONE (2026-11-16: kontrakt `docs/RESULTS_CONTRACT.MD` §8, biblioteka kontrolna `testdata/library` z wyrocznią JVM); LB1 DONE (2026-11-16: `-roots api`, `internal/libapi`, `exposed_api`); LB2 w toku (lb2-1 2026-11-17: luka `client_code`, `client_dispatch`; lb2-3/lb2-4: metody odziedziczone jako API, specjalizowane do odbiorcy; lb2-5: ogłoszenia Spring jako korzenie; lb2-6 2026-11-18: archiwa beanów CDI jako korzenie; lb2-7 2026-11-19: obiekty dostarczane przez wywołującego jako luka; lb2-8 2026-11-19: biblioteki ogłaszające klasy Springowi — pakiety Springa jako framework); LB3 w toku (lb3-1 2026-11-18: deklaracje zależności z pomu biblioteki, `supplied_by`); LB4 w toku (lb4-1 2026-11-19: zestaw kwalifikacyjny 5 bibliotek, wyrocznia JVM jetty-client, użycie w CI; lb4-2: wyrocznie JVM pozostałych bibliotek); zob. IMPLEMENTATION_STATUS. Wymagania: [LIBRARY_MODE.MD](LIBRARY_MODE.MD). Zasady: [ROADMAP.MD](ROADMAP.MD) (kontrakt wyniku B0, dowody krawędzi B1, weryfikacja B4).

Oznaczenia faz `LB` (litera `L` należy do DYNAMIC_LINKAGE).

## LB0. Kontrakt trybu — DONE (2026-11-16)

Decyzje: tryb to zestaw korzeni `-roots api` (nie osobna flaga `-mode`), zawężenia
`-api-exclude` i scenariusz JPMS `exports`; pola `findings[].exposed_api[]` i `coverage.api`;
mechanizm luki `client_code` zarezerwowany dla LB2. Oczekiwania i pomiar „przed”:
`testdata/library/EXPECTED.MD`.


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

## LB1. Korzenie z publicznego API i narażone metody — DONE (2026-11-16)

Zrobione bez pomiaru na prawdziwych bibliotekach — ten punkt przechodzi do LB4 (wybór
bibliotek razem z flotą).


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

## LB2. Kod klienta i wejścia frameworka — DONE for this scope (2026-11-23; [zarchiwizowane](../archive/2026-11-23/library_mode/LIBRARY_MODE_LB2.md))

## LB3. Wersje zależności

Postęp: **lb3-1 DONE (2026-11-18)** — pom biblioteki (wbudowany albo `-library-pom`) z rodzicami
i importowanymi BOM-ami; drzewo konsumenta jak w Mavenie (`compile`/`runtime` bez `optional`,
przechodnio, najbliższy wygrywa, `exclusions`); `supplied_by` (`library`/`user`/`undeclared`) i
`version_match` per instancja (`coverage.api.dependencies`, `components[].declaration`,
`findings[].supplied_by`); fixture `testdata/library-deps` z wyrocznią Maven 3.9. Zostaje z LB3:
ocena advisory dla innych wersji z zakresu, `MAPPING_VERSION_UNCOVERED`. Przy okazji: dostęp do
Maven Central przez lokalne repozytorium (`ARIADNE_MAVEN_LOCAL`), bez ponawiania po 429.

Zależności: LB1, CVE F4 (wersjonowane mapowania).

- Odczyt deklaracji zależności biblioteki (pom: zakresy wersji, `provided`, `optional`)
  i oznaczenie instancji dostarczanych przez użytkownika.
- Opcjonalnie: ocena advisory dla wersji z zakresu, które użytkownik może dostać
  (np. najnowsze w deklarowanej linii), jako osobne ustalenia z jawną wersją;
  `MAPPING_VERSION_UNCOVERED`, gdy mapowanie nie ma wariantu.
- Weryfikacja: fixture z zakresem wersji i zależnością `provided`.

Gotowe: wynik mówi, której wersji zależności dotyczy każde ustalenie i kto ją dostarcza.

## LB4. Kwalifikacja i użycie w CI — DONE for this scope (2026-11-23; powiązanie z CI floty → `FLEET_CI_PLAN.MD`; [zarchiwizowane](../archive/2026-11-23/library_mode/LIBRARY_MODE_LB4.md))

## Kolejność i szacunek

LB0 i LB1 to pierwsze wydanie (użyteczne samo w sobie); LB2 i LB3 po pomiarze LB1; LB4
na końcu. LB1 dotyka głównie wyboru korzeni i nowego pola wyniku — bez zmian w budowie
grafu i modelach.
