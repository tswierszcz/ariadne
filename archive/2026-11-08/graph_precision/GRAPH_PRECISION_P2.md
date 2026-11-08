# GRAPH_PRECISION:P2 — RTA i punkt stały — archiwum

**Data archiwizacji:** 2026-11-08.
**Stan:** DONE dla zatwierdzonego zakresu (opt-in `-callgraph-mode rta`; domyślny tryb CHA bez
zmian). Wiersz `docs/SUPPORT.MD` „RTA and points-to” (Partial — RTA jest, points-to nie)
opisuje bieżące zachowanie.

Oryginalny tekst fazy (poniżej, z dopiskiem „Wykonano”) był w
`RECOMMENDATION/GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD`; zastąpiony tam wskaźnikiem do tego
pliku. Pełny wpis z weryfikacją: `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, „GRAPH_PRECISION:P2”.
Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Commity

- `8591702` — Implement GRAPH_PRECISION P2: opt-in RTA with a fixed point; zmergowane jako
  PR #19 (`2082ff7`). Zawiera też commit archiwizacji P1 (`2beb4dc`, PR #18 zamknięty bez
  osobnego merge'a).

## Co zostało zrobione

- `internal/callgraph/rta.go` (`p2-1`): strategia `RTA` (kandydaci CHA przy budowie),
  `ApplyRTA` — punkt stały po osiągalnych metodach i klasach żywych z katalogu P1, cele CHA
  parkowane per klasa deklarująca, końcowa decyzja per cel, niezależność od kolejności.
- Polityka `cha-fallback-with-gap`: osiągalna luka ożywia klasy fallbacku (nazywalne w
  czasie działania — `classfile.NamedClasses`, adnotacje, typy członków; zarejestrowani
  providerzy; podtypy parametrów korzeni przy `-roots all`; proxy — brak), krawędzie
  zachowane przez fallback liczone osobno, `Unsafe`/uchwyty konstruktora jako luki
  `allocation`.
- `cmd/ariadne/rta.go`, raport „RTA”, v2 `callgraph.rta`.
- `testdata/rta` + `scripts/rta-oracle.sh`: wyrocznia prawdziwego JVM.

## Dowody

- `TestRTAOracle`: każde ciało uruchomione przez JVM osiągalne poza jedną oczekiwaną stratą
  pokrytą aktywną luką; 3 usunięte krawędzie wyjaśnione; `-roots all` nic nie usuwa; dwie
  mutacje wykryte. `TestRTAOrderIndependent`, `TestRTAFallbackAndGap`, `TestNameableClasses`.
- petclinic `-roots profile`: −52 561 z 310 418 krawędzi CHA, 6 FRAMEWORK → NOT_REACHABLE
  (komponenty Tomcata tylko z `server.xml`, nieużywany stos reaktywny), wszystkie
  `MODEL_INCOMPLETE`; WebGoat: −95 378 z 556 769, 2 zmiany (stos reaktywny); `-roots all`:
  bez zmian werdyktów; żaden REACHABLE nie stracony; przejście RTA 0,6–1,3 s; dwa przebiegi
  bajtowo identyczne. `go test ./...`, `go vet ./...` — PASS.

## Co nie zostało zrobione

- Fallback to uzasadniona heurystyka: klasa tworzona refleksyjnie i nazwana tylko w
  konfiguracji lub sygnaturze generycznej nie jest żywa (luka oznacza negatyw, nie przywraca
  krawędzi); fallback globalny, nie per miejsce; krawędzie bezpośrednie nie są zawężane.
- Domyślna strategia bez zmian — decyzja po P4. Points-to to P3.

## Oryginalny tekst fazy

## P2. RTA i punkt stały — DONE (2026-11-08)

Zależności: P1, B3; początkowo ograniczony katalog modeli z jawnymi lukami.

- Utworzyć kolejkę osiągalnych metod, odkrytych alokacji i nierozstrzygniętych wywołań. Nowy typ ponownie rozpatruje odpowiednie call sites.
- Wirtualny dispatch ograniczać dopiero po dokładnym dopasowaniu deskryptora i właściwej implementacji z JVM G2.
- Zapewnić niezależność od kolejności odkrycia typów, cykli i alokacji dopiero w metodzie odkrytej przez model.
- Stosować politykę unknown z P0; nie odrzucać typów tworzonych przez nieznaną fabrykę bez ostrzeżenia.
- Udostępnić opt-in RTA, raportować fallbacki i wpływ na werdykty. `roots=all` i korzenie profilowe porównywać jako oddzielne scenariusze.

Gotowe: fixtures z trzema implementacjami, pośrednią alokacją, cyklem i nieznaną fabryką zachowują wszystkie uzasadnione cele. P0–P2 to pierwsze wydanie precyzji grafu.

Wykonano (2026-11-08): `internal/callgraph/rta.go` — strategia `rta` (zbiór kandydatów CHA przy budowie), `ApplyRTA` po dodaniu wszystkich krawędzi modeli: kolejka osiągalnych metod, klasy żywe z pochodzeń katalogu P1 (kontener i klasy wejść od startu), krawędzie CHA czekające na typ (indeks po klasie deklarującej — nowy typ budzi tylko swoje cele), decyzja po dokładnym deskryptorze i `ResolveExact`; wynik niezależny od kolejności (`TestRTAOrderIndependent`). Polityka `cha-fallback-with-gap`: osiągalna luka ożywia klasy fallbacku (refleksja/`Unsafe`/uchwyt konstruktora: klasy nazywalne w czasie działania — literały klas i stałe-nazwy `classfile.NamedClasses`, klasy z adnotacją, typy pól/parametrów/wyników; nieznana usługa: zarejestrowani providerzy; `-roots all`: podtypy parametrów korzeni; proxy: brak klas wejścia), krawędzie zachowane tylko przez fallback liczone osobno, `Unsafe`/uchwyty jako luki modelu `allocation`. Raport „RTA”, v2 `callgraph.rta`. Wyrocznia JVM `testdata/rta` (`TestRTAOracle`), porównanie z CHA na petclinic/WebGoat w obu scenariuszach korzeni. Szczegóły: IMPLEMENTATION_STATUS, wpis GRAPH_PRECISION:P2.
