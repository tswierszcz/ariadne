# GRAPH_PRECISION:P4 — Benchmark i kwalifikacja — archiwum

**Data archiwizacji:** 2026-11-08.
**Stan:** DONE. Decyzja: domyślna strategia pozostaje CHA, RTA (`-callgraph-mode rta`) jest
opt-in. Bieżące wyniki: `testdata/graphbench/RESULTS.MD`; bieżące zachowanie: `docs/SUPPORT.MD`
(wiersze „RTA and points-to” i „Servlet container dispatch”).

Oryginalny tekst fazy (poniżej, z dopiskami „Wykonano” i „Ponowna ocena”) był w
`RECOMMENDATION/GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD`; zastąpiony tam wskaźnikiem do tego
pliku. Pełny wpis z weryfikacją: `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, „GRAPH_PRECISION:P4”.
Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Commity

- `a64edf7` — Implement GRAPH_PRECISION P4: CHA vs RTA benchmark and qualification;
  zmergowane jako PR #21 (`95ae6ec`).
- `dcbaa44` — Model servlet container dispatch to registered servlets and filters (wycinek
  FRAMEWORKS F2, ponowna ocena P4); PR #22 zmergowany do gałęzi `graph-precision-p4` już po
  merge'u #21, do `main` trafił razem z tą archiwizacją.

## Co zostało zrobione

- `scripts/graph-benchmark.sh`: CHA vs RTA na petclinic/WebGoat w obu scenariuszach korzeni;
  mierzone przebiegi bez pliku grafu (czas, szczytowe RSS) oddzielone od porównania.
- `TestGraphBenchmarkFixtures`: pominięte/nadmiarowe ciała wobec wyroczni JVM.
- Wersje katalogu i RTA w snapshotach (tryb `rta`).
- `testdata/graphbench/RESULTS.MD`: wyniki, ręczna kontrola każdej zmiany werdyktu, decyzja,
  migracja; po modelu dispatchu servletów — ponowny pomiar.

## Dowody

- RTA kosztuje poniżej sekundy i niemierzalną pamięć; `-roots all` bez zmian werdyktów.
- Pierwszy pomiar: 8 zmian przy `-roots profile` — 5 poprawnych, 1 nierozstrzygnięta,
  2 fałszywie negatywne (CVE-2024-38809, niemodelowany dispatch kontener → servlet).
- Po modelu `servlet-container:dispatch`: na petclinic CVE-2024-38809 i CVE-2024-22259 są
  FRAMEWORK przez prawdziwą ścieżkę żądania; zostają 4 poprawne zmiany RTA. WebGoat bez
  zmian (launcher Spring Boot poza analizowanym wejściem, 0 rejestracji).

## Co nie zostało zrobione

- Zewnętrzny silnik referencyjny — żaden nie jest przypięty; wzorcem są wyrocznie JVM i
  ręcznie sprawdzony bytecode.
- Fałszywy negatyw CVE-2024-38809 na WebGoat (wejście bez launchera Boot).
- P3 (points-to) — nieuzasadnione wynikami; obszar GRAPH_PRECISION pozostaje otwarty tylko
  dla P3.

## Oryginalny tekst fazy

## P4. Benchmark i kwalifikacja — DONE (2026-11-08)

- Osobno mierzyć błędne cele, pominięte oczekiwane cele, nieznane miejsca, zmiany CVE, czas i pamięć na fixtures i reprezentatywnych aplikacjach.
- Jeśli używany jest zewnętrzny silnik referencyjny, przypiąć wersję, zakres i konfigurację; jego wynik także nie jest automatycznie prawdą.
- Porównanie musi uwzględniać wersję modeli frameworka/refleksji i reguły JDK skip. Ulepszenie potwierdzać ręcznie opisanymi przypadkami, nie opinią LLM o krótszej ścieżce.
- Wykonać B4, snapshoty i dokumentację migracji. Zmiana domyślnej strategii jest osobną decyzją po pomiarze.

Nie podaje się z góry gwarantowanej redukcji fałszywych alarmów ani kosztu analizy całego classpath.

Wykonano (2026-11-08): `scripts/graph-benchmark.sh` (petclinic/WebGoat × `-roots profile`/`all` × CHA/RTA; 3 mierzone przebiegi bez pliku grafu — czas i szczytowe RSS przez `wait4` — plus przebieg z `-graph-out`/`-graph-compare`; wszystkie liczby poza czasem/pamięcią muszą być identyczne) → `testdata/graphbench/{results.tsv,verdict-changes.tsv,versions.txt}`; `TestGraphBenchmarkFixtures` (pominięte/nadmiarowe ciała wobec wyroczni JVM `testdata/callgraph` i `testdata/rta`); wersje katalogu i RTA w snapshotach (tylko tryb `rta`). Bez zewnętrznego silnika referencyjnego (żaden nie jest przypięty; wzorcem są wyrocznie JVM i ręcznie sprawdzony bytecode). Kwalifikacja (`testdata/graphbench/RESULTS.MD`): RTA kosztuje 0,4–0,7 s i niemierzalną pamięć; `-roots all` bez zmian werdyktów; 8 zmian przy `-roots profile` — 5 poprawnych, 1 nierozstrzygnięta, 2 fałszywie negatywne (CVE-2024-38809: niemodelowany dispatch kontener → servlet przez pomijany odbiorca `javax.*` w obu trybach). Decyzja: domyślnie CHA, RTA opt-in; ponowna ocena po zamodelowaniu dispatchu servletów/filtrów. Szczegóły: IMPLEMENTATION_STATUS, wpis GRAPH_PRECISION:P4.

Ponowna ocena (2026-11-08) po modelu dispatchu kontenera servletów (FRAMEWORKS F2, `servlet-container:dispatch`): benchmark powtórzony (`--repeat 3`, wersje: builder `2`, katalog `p1-2`, `servlet.dispatch` `s1`), fixture `testdata/servletdispatch` w `TestGraphBenchmarkFixtures`. petclinic: 4 zarejestrowane servlety/filtry (autokonfiguracja Boot), `DispatcherServlet` osiągalny w obu trybach; CVE-2024-38809 i CVE-2024-22259 są FRAMEWORK przez rzeczywistą ścieżkę żądania, więc zostają 4 poprawne zmiany RTA (zawory/realm Tomcata); 3 nowe FRAMEWORK w każdym scenariuszu (CVE-2024-38819, CVE-2026-41844, CVE-2023-20883) — opisane w `testdata/graphbench/RESULTS.MD`. WebGoat bez zmian: analizowane wejście nie zawiera launchera Boot (`webgoat-server`), więc nic nie rejestruje `DispatcherServlet`, a fałszywie negatywne CVE-2024-38809 przy RTA pozostaje. Decyzja bez zmian: domyślnie CHA; P3 nadal nieuzasadnione.
