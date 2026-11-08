# GRAPH_PRECISION:P1 — Katalog typów tworzonych jawnie i pośrednio — archiwum

**Data archiwizacji:** 2026-11-08.
**Stan:** DONE dla zatwierdzonego zakresu (tylko raport — graf i werdykty bez zmian). Wiersz
`docs/SUPPORT.MD` „RTA and points-to” (Not implemented — P1 dostarcza katalog, z którego
skorzysta RTA w P2) opisuje bieżące zachowanie.

Oryginalny tekst fazy (poniżej, z dopiskiem „Wykonano”) był w
`RECOMMENDATION/GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD`; zastąpiony tam wskaźnikiem do tego
pliku. Pełny wpis z weryfikacją: `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, „GRAPH_PRECISION:P1”.
Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Commity

- `d646971` — Implement GRAPH_PRECISION P1: catalog of created types; zmergowane jako PR #17
  (`46076b5`).

## Co zostało zrobione

- `internal/alloc` (`CatalogVersion = "p1-1"`): klasy tworzone z pochodzeniem i miejscami —
  `new` (tylko klasy konkretne), `service`/`reflection`/`model` (krawędzie modeli do
  `<init>`), `container` (stereotypy Springa, konkretne typy zwracane przez `@Bean`),
  `deserialization` (kandydaci per format z wejściem tego formatu), `entry` (klasy
  metod-korzeni instancyjnych przy `-roots profile`); miejsca lambd osobno; luki
  `reflection-unresolved`, `dynamic-proxy`, `unsafe-allocation`, `constructor-handle`,
  `service-unresolved`, `container-proxy`, `external-input` (`-roots all`).
- `cmd/ariadne/alloc.go`, sekcja raportu „Created types”, v2 `callgraph.created_types`,
  `-graph-out` → `ariadne-callgraph-2` z listą klas; `-graph-compare` czyta też v1 i podaje
  dodane/usunięte typy.
- `testdata/alloc` + `scripts/alloc-oracle.sh`: wyrocznia prawdziwego JVM.

## Dowody

- `TestCreatedTypesOracle` (pełne CLI bez JDK): każda klasa utworzona przez JVM jest w
  katalogu z pochodzeniem swojego mechanizmu, proxy pokryte luką `dynamic-proxy`, jedyny
  nadmiar to `new` w nieosiąganej metodzie; mutacja (bez pochodzenia `reflection`) wykryta.
  `TestOriginsAndGaps`, `TestTypesInExport`.
- petclinic 17 763 klas / 329 luk, WebGoat 26 594 / 328; `-graph-compare` z grafem `main`:
  0 krawędzi dodanych/usuniętych, 0 zmienionych ustaleń; v2 różni się tylko
  `created_types`; dwa przebiegi bajtowo identyczne; czas w granicach szumu.
- `go test ./...`, `go vet ./...` — PASS.

## Co nie zostało zrobione

- Katalog nie zna osiągalności — aktywacja klas przez osiągalne miejsca to P2 (RTA).
- Deserializacja bez odrzuceń polityki (nadmiar, bezpieczny); `@Bean` tylko z konkretnym
  typem zwracanym; tablice nieskatalogowane; `clone()` nie tworzy nowego typu; eksport
  podaje liczbę miejsc per pochodzenie, nie same miejsca.

## Oryginalny tekst fazy

## P1. Katalog typów tworzonych jawnie i pośrednio — DONE (2026-11-08)

Zależności: DATA_FLOW D1–D2, rejestr B3.

- Rozpoznać alokacje `new`, tablice w odpowiednim zakresie oraz instancje dostarczane jako wejście scenariusza.
- Zbierać typy tworzone przez SERVICE S2, FRAMEWORK F2/F3, REFLECTION R2, DYNAMIC L1/L2 i DESERIALIZATION Z1, gdy te modele są aktywne.
- Brak modelu któregoś mechanizmu rejestrować jako potencjalną lukę; wdrożenie pełnego katalogu nie jest ukrytym wymogiem startu P1.
- Typy przekazywane do publicznych metod w trybie biblioteki nie są ograniczone wyłącznie do alokacji wewnątrz badanej biblioteki. Zadeklarować zewnętrzne wejścia/nieznane typy.

Gotowe: modelowana alokacja ma pochodzenie i może uruchomić rozwinięcie grafu tak samo jak jawne `new`.

Wykonano (2026-11-08): `internal/alloc` — klasy tworzone z pochodzeniem i miejscami (`new`, `service`/`reflection`/`model` z krawędzi modeli do konstruktora, `container`, `deserialization`, `entry`), miejsca lambd osobno, luki (nierozwiązane `newInstance`, `Proxy`, `Unsafe`, `findConstructor`, nieznane usługi, implementacje generowane przez kontener, wejścia zewnętrzne przy `-roots all`); każde pochodzenie ma metodę-miejsce, której osiągalność w P2 aktywuje klasę tak samo jak jawne `new`. Raport, v2 `callgraph.created_types`, lista w `-graph-out` (`ariadne-callgraph-2`). Wyrocznia JVM `testdata/alloc` (`TestCreatedTypesOracle`). Tablice nieskatalogowane (bez metod do dispatchu poza `Object`). Szczegóły: IMPLEMENTATION_STATUS, wpis GRAPH_PRECISION:P1.
