# DYNAMIC_LINKAGE:L1 — lambdy i method references — archiwum

**Data archiwizacji:** 2026-11-09.
**Stan:** DONE dla zadeklarowanego zakresu. Bieżące zachowanie: `docs/SUPPORT.MD` (wiersz
„Lambdas and method references”, `Partial`); wyniki: `testdata/graphbench/RESULTS.MD`,
`testdata/graphbench/TRACE_GAPS.MD`.

Oryginalny tekst fazy (poniżej) był w `RECOMMENDATION/DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD`;
zastąpiony tam wskaźnikiem do tego pliku. Pełny wpis z weryfikacją:
`RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, „DYNAMIC_LINKAGE:L1 — materializacja wywołań SAM”.
Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Commity

- Wcześniejsza połowa (miejsce utworzenia): wartość `valueflow.KindFunctional` w
  `internal/dataflow`, cel implementacji rozwiązany z bootstrapu, rodzaj uchwytu — razem z L0.
- `a19eab5` — DYNAMIC_LINKAGE L1: SAM-call materialization; zmergowane jako PR #30
  (`2aa71d9`), na gałęzi od `cha-inherited` (PR #29, `36b361f`).

## Co zostało zrobione

- `internal/samcall` (reguły `l1-1`): wywołanie metody SAM interfejsu funkcyjnego (i wywołanie
  metody spoza wejścia z argumentem takiego typu — `List#forEach`, `Map#computeIfAbsent`)
  dostaje krawędź `sam` do implementacji lambdy/referencji, której wartość dociera do odbiorcy
  przez DATA_FLOW: bezpośredni wywołujący (także przez nadpisaną deklarację nadtypu), wartości
  przechwycone przez lambdę od twórcy, zapisy pola w zagnieżdżeniu klas, zwroty metod
  static/private. Ogniwa zapamiętane, cykle cięte, typy, które nie mogą trzymać lambdy, nie są
  śledzone; budżet na miejsce.
- Referencja metody instancyjnej rozwijana jak wywołanie wirtualne; konstruktor jako `<init>`.
- `callgraph.BasisSAM`, `callgraph.SAMGate` — krawędź dopiero, gdy metoda tworząca wartość
  jest osiągalna z korzeni (punkt stały z modelami bibliotek). `BuilderVersion` 3 → 4.
- Nierozwiązany odbiorca nie dodaje krawędzi; luka z kategorią w logu i nagłówku raportu.
- Przy okazji: parser klas bez bufora 32 KiB na atrybut, `dataflow.CallerIndex` trzyma
  wskaźnik do metody zamiast kopii (~300 MB mniej na petclinic).

## Dowody

- `testdata/samcall` (prawdziwy javac + JVM, `scripts/samcall-oracle.sh`), `TestSAMFixture`:
  wszystkie 8 zmierzonych relacji między metodami to krawędzie `sam` pod zmierzonym bci;
  nieużyta lambda i lambda z nigdy niewywoływanej metody nie dostają żadnej.
  `TestSAMGapsDisclosed`, `TestMayDeclareOutside`.
- `TestRuntimeTrace`: luka `sam-call` w fixture zniknęła.
- petclinic, ślad JFR: poza grafem 204 → 175 z 2 702, `sam-call` 50 → 22 (12 z nich to nie
  lambdy).
- `scripts/graph-benchmark.sh --repeat 3`: żadnej zmiany werdyktu w 8 konfiguracjach;
  +1 626 / +1 557 krawędzi `sam` (petclinic / WebGoat), 0 usuniętych; analiza szybsza niż z
  builderem 3 (petclinic 14,6 → 10,4–11,1 s, WebGoat 23,2–24,0 → 16,2–16,3 s), pamięć bez zmian.
- Bramka twórcy usunęła fałszywe `FRAMEWORK` dla CVE-2021-23463, które dawała wersja bez niej.
- `go test ./...`, `go vet ./...`, `go test -race ./internal/samcall ./internal/dataflow`.

## Co nie zostało zrobione

- Wartość nie jest śledzona przez wynik wywołania instancyjnego, element tablicy, kolekcję ani
  zmienną żywą przez handler wyjątku.
- Analiza niewrażliwa na kontekst (suma po wywołujących, złagodzona bramką twórcy).
- Wartość przekazana do metody JDK traktowana jako przez nią uruchamiana, także gdy JDK ją tylko
  przechowuje (`Map#put`).
- Krawędzie `sam` nie są zawężane przez RTA.
- Klasyfikator luk śladu liczy jako `sam-call` także wywołania nie-lambd przez ramki JDK.

## Oryginalny tekst fazy

## L1. Lambdy i method references — DONE (2026-11-09)

- Model LambdaMetafactory tworzy wartość obiektu funkcyjnego i przechwytuje receiver/argumenty w zakresie lokalnej analizy.
- Wywołanie SAM albo zarejestrowany callback materializuje cel z prawidłowym przekazaniem argumentów. Współpracuje z ASYNC A1–A2, nie wymaga jego pełnego katalogu.
- Rozróżnić statyczny cel, wirtualny method reference i konstruktor. Wirtualny cel korzysta z JVM G2 i znanych typów.
- Oddzielić możliwy kod bootstrapu od ciała lambdy. W miarę migracji zachować jawnie oznaczony konserwatywny wariant dla nieznanego użycia obiektu zamiast cichego utracenia dotychczasowych celów.

Gotowe: nieużyta lambda, natychmiastowe wywołanie, zapis w zmiennej i reference do override mają oczekiwane różnice. To pierwszy przyrost poprawy istniejącego wsparcia.
