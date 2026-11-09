# RUNTIME_NATIVE:N0–N1 — kontrakt śladu i importer offline — archiwum

**Data archiwizacji:** 2026-11-08.
**Stan:** DONE dla zatwierdzonego zakresu (bez agenta Java). Bieżące zachowanie:
`docs/SUPPORT.MD`, wiersz „Runtime evidence (traces)”; kontrakt: `docs/RUNTIME_TRACE.MD`.

Oryginalny tekst faz (poniżej, z dopiskami „Wykonano”) był w
`RECOMMENDATION/RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD`; zastąpiony tam wskaźnikiem do tego
pliku. Pełny wpis z weryfikacją: `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, „RUNTIME_NATIVE:N0–N1”
(oraz rozszerzenie „RUNTIME_NATIVE:N1 — filtr dokładnego śledzenia z mapowań CVE”).
Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Commity

- `659dff3` — Implement RUNTIME_NATIVE N0–N1: runtime trace contract and offline importer;
  zmergowane jako PR #24.

## Co zostało zrobione

- `docs/RUNTIME_TRACE.MD`: format `ariadne-runtime-trace-1` (metadane, zakres, próbkowanie,
  kompletność, artefakty, loadery, definicje klas z hashami, metody, dowody `loaded`/
  `executed`/`call`/`handoff`), reguły czytania, statusy korelacji.
- `scripts/JfrToTrace.java`: rejestrator bez agenta z nagrania JFR.
- `internal/runtimeevents` (`n1-1`): ścisły parser z limitami, korelacja po SHA-256 klas,
  definicje per loader osobno, wywołania vs graf statyczny, podatne metody per ustalenie.
- `-runtime-trace`, raport „Runtime evidence”, v2 `runtime`/`findings[].runtime_observed`,
  snapshot (wejście, wersja `runtime.import`).

## Dowody

- `TestRuntimeTrace`, `TestRuntimeTraceSnapshot`, `TestParseRejects`, `TestCorrelateStatuses`
  (mutacja ignorująca hash wykryta) na fixture JFR `testdata/runtime`.
- petclinic pod JFR, dwa przebiegi: 10 527 klas dopasowanych po hashu; CVE-2024-38809
  śledzone ok. 1 400 razy; CVE-2025-35036 wykonane mimo statycznego NOT_REACHABLE (SPI
  `javax.validation`); werdykty bez zmian.

## Co nie zostało zrobione

- Agent Java (N3), kontekst loaderów poza rozdzieleniem definicji (N2), korelacja handoff
  (N4), native/JNI (N5). Źródło bajtów klasy z JFR nieznane (hash z classpath).

## Oryginalny tekst faz

## N0. Kontrakt śladu i zakres dowodów — DONE (2026-11-08)

Zależności: B0–B1, wspólne pochodzenie CVE F1; snapshoty CVE F3 przed deklaracją odtwarzalności.

- Zaprojektować wersjonowany format: run id, scenariusz, runtime, narzędzie rejestrujące i jego wersja, zakres pomiaru, class/loader/method ids oraz hashe zawartości.
- Wyróżnić zdarzenia loaded class, method entry, observed call oraz przekazanie zadania. Pole call site/parent event musi wynikać z pomiaru, nie zgadywania po kolejności czasowej.
- Zapisać sampling, utratę zdarzeń, ograniczenia instrumentacji, zakres czasowy i przyczyny niepełności.
- Zdefiniować dane minimalne bez argumentów i payloadów. Ślady oraz opcjonalny bytecode są osobnymi artefaktami z hashami i kontrolą zgodności.

Gotowe: można opisać wiarygodność pojedynczego zdarzenia i całego przebiegu, także przy częściowym śladzie.

Wykonano (2026-11-08): `docs/RUNTIME_TRACE.MD` — format JSON Lines `ariadne-runtime-trace-1`: nagłówek (run id, scenariusz, zakres czasu, runtime, rejestrator i wersja, zakres: rodzaje zdarzeń, klasy, filtry metod; próbkowanie; kompletność: utracone zdarzenia, ucięte stosy, przyczyny; polityka danych `identifiers-only`), rekordy tożsamości (artefakt z SHA-256, loader, definicja klasy per loader z hashem i `source_basis` `observed`/`classpath-lookup`, metoda z deskryptorem) i dowodów (`loaded`, `executed` `traced`/`sampled`, `call` `stack-adjacent`/`stack-through-runtime` z bci/linią wywołującego — tylko zmierzone, `handoff` tylko przy skorelowanych końcach); reguły czytania (brak zdarzenia nic nie dowodzi, wykonanie ≠ wykorzystanie, załadowanie ≠ wykonanie, czas ≠ przyczynowość) i statusy korelacji. Rejestrator bez agenta: `scripts/JfrToTrace.java` (JFR `jdk.ClassLoad`, `jdk.ExecutionSample`, na JDK 25+ `jdk.MethodTrace`).

## N1. Importer offline i korelacja — DONE (2026-11-08)

- Utworzyć proponowany `internal/runtimeevents` z walidacją formatu, typów, odwołań i limitów danych wejściowych.
- Powiązać metodę z komponentem/bytecode/deskryptorem. Niezgodny hash, nieznany loader lub brak klasy daje brak dopasowania, a nie arbitralne przypięcie do znanego FQCN.
- Pokazać zaobserwowane metody i relacje w raporcie bez nadpisywania statycznego grafu. Załadowanie klasy nie oznacza wykonania wszystkich jej metod.
- Zachować id dowodu w wynikach i snapshotach, np. proponowane wejście `-runtime-trace`.

Gotowe: poprawny trace, inna wersja artefaktu, dwa loadery i niepełny trace mają odrębne wyniki. N0–N1 to pierwszy zakres runtime, bez konieczności tworzenia agenta.

Wykonano (2026-11-08): `internal/runtimeevents` (`n1-1`): `Parse` — ścisła walidacja (nieznane rodzaje/pola, wiszące odwołania, duplikaty, wartości spoza enumeracji, rodzaj niezadeklarowany w nagłówku, próbkowany a „kompletny”, limity rozmiaru) odrzuca cały plik; `Correlate` — klasa po SHA-256 pliku klasy (albo artefaktu): `matched`/`matched_artifact` (dowód), `different_version`/`name_only`/`not_in_input` (pokazane, nieużyte), definicje per loader osobno; metody → węzły grafu; zmierzone wywołania → `in_graph` (z podstawami krawędzi) / `not_in_graph`; podatne metody zaobserwowane per ustalenie i instancja. CLI `-runtime-trace` (walidacja przed analizą), raport „Runtime evidence”, v2 `runtime` i `findings[].runtime_observed`, snapshot (wejście `runtime-trace`, wersja `runtime.import`). Graf i werdykty bez zmian. Fixture JFR `testdata/runtime` (`scripts/runtime-trace-fixture.sh`) i prawdziwy przebieg petclinic (`scripts/runtime-trace-petclinic.sh`). Szczegóły: IMPLEMENTATION_STATUS, wpis RUNTIME_NATIVE:N0–N1.

## Addendum (2026-11-09): przyczyny luk grafu względem śladu (`g1`)

Rozszerzenie N1 po archiwizacji: każde zmierzone wywołanie, dla którego graf statyczny nie ma
krawędzi, jest klasyfikowane z bytecode wywołującego pod zmierzonym indeksem (`sam-call`,
`skipped-receiver`, `reflection`, `inherited-impl`, `proxy`, `class-init`, `async`,
`rta-pruned`, …) i grupowane z typami odbiorców (raport, v2
`runtime.calls_not_in_graph_by_cause`; `TestGapClassifier`). Pomiar na petclinic:
`testdata/graphbench/TRACE_GAPS.MD`. Skutki: naprawa CHA dla implementacji dziedziczonych
(JVM_GRAPH G2, PR #29) i materializacja wywołań SAM (DYNAMIC_LINKAGE L1, PR #30). Commit
`16ff33a` (PR #28, wszedł do `main` razem z #29). Pełny wpis:
`RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, „RUNTIME_NATIVE:N1 — przyczyny luk grafu względem
śladu”.
