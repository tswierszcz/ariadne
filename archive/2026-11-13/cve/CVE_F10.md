# CVE:F10 — pokrycie mapowań bez commita poprawki — archiwum

**Data archiwizacji:** 2026-11-13.
**Stan:** DONE dla zadeklarowanego zakresu (wraz z kontynuacją: ręczne mapowania z diffów
wydań). Bieżące zachowanie: `docs/SUPPORT.MD` (wiersz „Manual and fix-diff method mapping”,
`Partial`, oraz akapit o kontrakcie wyników — liczba wpisów tabeli ręcznej); README, sekcja
źródeł mapowania; wyniki: `testdata/graphbench/RESULTS.MD`.

Oryginalny tekst fazy (poniżej) był w `RECOMMENDATION/CVE_IMPLEMENTATION_PLAN.MD`; zastąpiony
tam wskaźnikiem do tego pliku. Pełne wpisy z weryfikacją: `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`,
„CVE:F10 — pokrycie mapowań bez commita poprawki” i „CVE:F10 (kontynuacja) — ręczne mapowania z
diffów wydań”. Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Commity

- `fe252a6` — CVE:F10 mapping coverage: pull requests, field initializers, release diff;
  zmergowane jako PR #48 (`4b88699`).
- `2feaa7f` — Manual mappings from hand-read release diffs (CVE:F10 follow-up); zmergowane jako
  PR #49 (`8d9ff45`).

## Co zostało zrobione

- **Pull requesty:** odnośnik `…/pull/N` → commit merge (API GitHub), merge dwurodzicowy
  względem pierwszego rodzica; niezmergowane PR-y i PR-y w `github/advisory-database` pomijane;
  nieczytelny PR to błąd odnośnika, nie przebiegu (`internal/vulnmap/diffsource.go`).
- **Inicjalizatory pól:** `internal/javadiff` przypisuje zmienione inicjalizatory i bloki
  inicjalizacyjne do `<clinit>`/wszystkich konstruktorów, tylko gdy poprawka nie zmieniła żadnej
  metody.
- **`release-diff`** (`internal/vulnmap/releasediff.go`, domyślnie włączone,
  `-release-diff=false`): metody zmienione między wydaniem przed poprawką a wydaniem patch z
  poprawką, w klasach nazwanych w advisory (`classfile.MethodFingerprint`); szeregowany ruch do
  Maven Central z User-Agentem, offline zapas z `~/.m2/repository`.
- **Ręczne mapowania (kontynuacja):** 22 wpisy schematu v2 w `vuln_methods.yaml` (46 → 68) dla
  advisories bez commita i bez nazwy klasy — SnakeYAML DoS, Spring MVC (zasoby statyczne, SSE,
  szablony skryptowe, multipart), SpEL, wstrzykiwanie do logów Springa, Thymeleaf — z ręcznie
  czytanych diffów wydań, sprawdzone z nazwy na zakresie wydań każdego wariantu
  (`testdata/cve-control/MANUAL_MAPPING_AUDIT.MD`).

## Dowody

- Testy: `TestFixDiff_pullRequest*`, `TestFixDiff_fieldInitializers`,
  `TestFixDiff_constructorNextToFieldInitializer`, `TestInitializers`, `TestReleaseDiff_*`,
  `TestReleaseBefore`, `TestReleaseDiffManualMappings`, `TestLoadManualMeta_embeddedTable`;
  `go test ./...`, `go vet ./...`.
- Walidacja `release-diff` na 130 mapowaniach z commitów: precyzja metod 95 %, pokrycie w
  nazwanych klasach 85 % (`testdata/cve-control/RELEASE_DIFF_VALIDATION.MD`).
- `UNKNOWN` (`-roots all`, CHA): petclinic 57 → 49 → 27, WebGoat 60 → 45 → 23, JPetStore
  12 → 8 → 5; zmieniają się tylko dotychczasowe `UNKNOWN` (plus CVE-2024-6162 na WebGoat,
  backport). `scripts/graph-benchmark.sh` — grafy bez zmian w obu przyrostach.

## Co nie zostało zrobione

- Advisory bez nazwy klasy albo z poprawką tylko w nowym wydaniu minor nie jest mapowane
  automatycznie; `release-diff` to heurystyka (`unreviewed`).
- PR z rebase wielu commitów czytany przez ostatni commit.
- Bez mapowania: CVE-2025-41242 (6.2.9 → 6.2.10 nie zmienia metod webmvc), CVE-2024-38828
  (poprawka tylko komercyjna); pozostałe `UNKNOWN` to głównie advisories bez kodu JVM albo bez
  publicznego kodu poprawki (MySQL Connector/J).
- Na WebGoat werdykty zasobów statycznych i widoków Thymeleaf to prawdopodobnie fałszywie
  negatywne `NOT_REACHABLE` (brak rejestracji servletów — znana luka, `MODEL_INCOMPLETE`).

## Oryginalny tekst fazy

## F10. Pokrycie mapowań bez commita poprawki — DONE (2026-11-10; zob. IMPLEMENTATION_STATUS)

### Cel

Zmniejszyć liczbę `UNKNOWN` wynikających z braku mapowania. Pomiar 2026-11-10 (petclinic,
WebGoat, JPetStore): z 69 niezmapowanych advisories 52 nie mają w OSV commita poprawki, 10 mają
tylko pull request, 5 commit z diffem bez metod.

### Zadania

1. Odnośnik do pull requestu → commit, który go zmergował (merge dwurodzicowy względem
   pierwszego rodzica); PR niezmergowany jako błąd odnośnika.
2. Zmienione inicjalizatory pól i bloki inicjalizacyjne → `<clinit>`/wszystkie konstruktory,
   tylko gdy poprawka nie zmieniła żadnej metody.
3. Źródło `release-diff`: metody zmienione między wydaniem przed poprawką a wydaniem patch z
   poprawką, w klasach nazwanych w advisory; walidacja względem mapowań z commitów i reguła
   włączenia domyślnego zależna od precyzji.

### Kryteria ukończenia

- Testy jednostkowe dla PR, inicjalizatorów i porównania wydań; walidacja precyzji źródła 3
  ([testdata/cve-control/RELEASE_DIFF_VALIDATION.MD](../testdata/cve-control/RELEASE_DIFF_VALIDATION.MD)).
- Pomiar `UNKNOWN` i zmian werdyktów na aplikacjach referencyjnych, ręczna kontrola próby.
- Odpowiedzialny ruch do Maven Central (szeregowanie, cache, User-Agent).
