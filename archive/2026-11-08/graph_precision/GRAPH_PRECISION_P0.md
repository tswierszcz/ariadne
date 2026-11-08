# GRAPH_PRECISION:P0 — Uczciwy baseline i tryby porównania — archiwum

**Data archiwizacji:** 2026-11-08.
**Stan:** DONE dla zatwierdzonego zakresu. Wiersze `docs/SUPPORT.MD` „Direct calls, CHA and
inheritance” (Partial) i „RTA and points-to” (Not implemented — P0 tylko przygotowuje
porównanie) opisują bieżące zachowanie.

Oryginalny tekst fazy (poniżej, z dopiskiem „Wykonano”) był w
`RECOMMENDATION/GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD`; zastąpiony tam wskaźnikiem do tego
pliku. Pełny wpis z weryfikacją: `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, „GRAPH_PRECISION:P0”.
Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Commity

- `d5422a6` — Implement GRAPH_PRECISION P0: strategy, graph settings, graph comparison;
  zmergowane jako PR #15 (`d9614c1`).

## Co zostało zrobione

- `internal/callgraph/strategy.go`: `Strategy` (wybór celów wywołań wirtualnych/interfejsowych)
  oddzielony od budowy grafu; `CHA` jako jedyna strategia, `Build` jako adapter — graf
  identyczny z wcześniejszym. `Config` grafu: tryb, wersja buildera, reguła skip, polityka
  nieznanych alokacji (`all-subtypes` dla CHA; `cha-fallback-with-gap` ustalona dla RTA),
  korzenie, target Java.
- `internal/callgraph/compare.go`: eksport `ariadne-callgraph-1` (tożsamość krawędzi:
  metoda źródłowa i docelowa, rodzaj, podstawa; PC; ścieżki ustaleń) i porównanie (różnice
  ustawień, krawędzie dodane/usunięte wg podstawy, ustalenia ze zmienionym werdyktem lub
  ścieżką z krawędziami wyjaśniającymi).
- CLI `-callgraph-mode cha`, `-graph-out`, `-graph-compare`; nagłówek raportu i sekcja
  porównania; v2 `callgraph`; snapshot config.
- `testdata/callgraph` + `scripts/callgraph-oracle.sh`: wyrocznia znanych celów na
  prawdziwym JVM (overloady, metoda default, bridge generyczny i kowariantny, ciało
  dziedziczone, wywołanie super) — domyka zaległą weryfikację JVM_GRAPH:G2.

## Dowody

- `TestKnownTargets`: 13/13 ciał uruchomionych przez JVM obecnych w grafie, 3 nadmiarowe z
  uzasadnieniem, brak krawędzi CHA do metody niezadeklarowanej; mutacja dopasowania po
  nazwie wykryta. `TestExportRoundTrip`, `TestCompareExplains`.
- petclinic/WebGoat offline vs `main`: werdykty bez zmian, v2 różni się tylko sekcją
  `callgraph`; v2 i `-graph-out` bajtowo identyczne w dwóch przebiegach; porównanie z samym
  sobą puste; porównanie z `-roots profile` ujawnia różnicę ustawień.
- `go test ./...`, `go vet ./...` — PASS.

## Co nie zostało zrobione

- Błędów deskryptorów do skorygowania nie znaleziono — nic do poprawienia przed RTA.
- Jedyny tryb to CHA; RTA to P2, katalog tworzonych typów — P1.
- Plik `-graph-out` jest duży (≈ 370 MB dla WebGoat); ścieżki ustaleń w eksporcie w wersji v1.

## Oryginalny tekst fazy

## P0. Uczciwy baseline i tryby porównania — DONE (2026-11-07)

Zależności: JVM G0–G4, B0–B1, CVE F1–F4 dla porównywalnych wyników.

- W `internal/callgraph` rozdzielić wybór strategii od internowania i przechowywania grafu. Zachować istniejący builder przez adapter do trybu legacy/CHA.
- Zapisać sygnatury, korzenie, target i konfigurację skip w metadanych analizy. Nie zmieniać tych parametrów między porównywanymi trybami bez ujawnienia różnicy.
- Przygotować przypadki, w których wiadomo, które cele są poprawne; skorygować błędy deskryptorów przed przypisaniem redukcji nowemu algorytmowi.
- Zdefiniować politykę nieznanej alokacji: np. konserwatywny fallback do uzasadnionego zbioru typów dla dotkniętego miejsca, a jeśli nie ma dostatecznej podstawy — niekompletność i ograniczenie wyniku negatywnego.

Gotowe: porównanie krawędzi ma stabilne identyfikatory i wyjaśnia zmiany, a nie tylko rozmiar grafu.

Wykonano (2026-11-07): `internal/callgraph/strategy.go` (`Strategy`, `CHA`, `BuildWith`; `Build` jako adapter — wynik identyczny z `main`), `-callgraph-mode cha`, `Config` grafu w nagłówku raportu, v2 `callgraph` i snapshotcie; `-graph-out`/`-graph-compare` (tożsamość krawędzi: metoda źródłowa i docelowa, rodzaj, podstawa; zmiany werdyktów/ścieżek z krawędziami wyjaśniającymi; różnice ustawień zawsze ujawnione); `testdata/callgraph` z wyrocznią prawdziwego JVM (`scripts/callgraph-oracle.sh`, `TestKnownTargets`) — bez błędów deskryptorów do skorygowania; polityka nieznanej alokacji: CHA `all-subtypes`, dla RTA ustalona `cha-fallback-with-gap`. Szczegóły: IMPLEMENTATION_STATUS, wpis GRAPH_PRECISION:P0.
