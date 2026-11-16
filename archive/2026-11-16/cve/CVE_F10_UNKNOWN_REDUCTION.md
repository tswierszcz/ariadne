# CVE:F10 — redukcja `UNKNOWN` (kontynuacja) — archiwum

**Data archiwizacji:** 2026-11-16.
**Stan:** DONE dla zadeklarowanego zakresu. Faza CVE:F10 była zarchiwizowana 2026-11-13
([CVE_F10.md](../../2026-11-13/cve/CVE_F10.md)); to jest zapis jej drugiej kontynuacji.
Bieżące zachowanie: `docs/SUPPORT.MD` (wiersze „Manual and fix-diff method mapping” i
„Reachability statuses”), `docs/RESULTS_CONTRACT.MD` §2–3, README (źródła mapowania, tabela
statusów). Pełne wpisy z weryfikacją: `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, cztery wpisy
„CVE:F10 …” z 2026-11-16. Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Zakres (zatwierdzony 2026-11-16)

Przegląd 41 unikalnych advisory `UNKNOWN` na petclinic/WebGoat/JPetStore (27/26/5 ustaleń):

- grupa A: brak commita poprawki i nazwy klasy (~27);
- grupa B: mapowanie z commita wskazuje metody nieobecne w badanej linii wydań (9);
- grupa C: advisory, które nie dotyczą kodu (~5).

Plan: ręczne mapowania z diffów wydań w trzech PR-ach, kod powodu dla grupy C, pomiar
przed/po. Po części 3 dodano decyzję o kodzie nieobecnym na classpath.

## Commity

| Część | Commit | PR (merge) | Zakres |
|---|---|---|---|
| 1 | `5189cb9` | #84 (`fc26be1`) | 15 wpisów Spring Framework/Boot/Data/Security (68 → 83) |
| 2 | `bdcce87` | #85 (`4b27cbd`) | 14 wpisów Tomcat, Jackson, logback, Undertow, XNIO (83 → 97) |
| 3 | `50d2163` | #86 (`c6b897b`) | 9 wpisów H2, MySQL, pgjdbc, jsoup, hibernate-validator, Spring Boot CloudFoundry + wpis `not_code` (97 → 107); kod `NOT_CODE_MAPPABLE` |
| kod nieobecny | `1a42298` | #87 (`534e20f`) | kod `MAPPED_CODE_ABSENT` |

## Co zostało zrobione

- **Metoda mapowania:** diff znormalizowanego bytecode'u ostatniego wydania podatnego i
  pierwszego naprawionego (narzędzie robocze `work/jardiff`: `-has`, `-list`, `-calls`),
  przypisanie zmian po treści advisory, sprawdzenie każdej metody z nazwy w wydaniach
  wariantu (komentarz `checked:` przy każdym wpisie). Gdy commit poprawki nazywa klasę,
  której badana linia nie ma (parser async Jacksona od 2.14, procesory modelu logbacka od
  1.3, pakiety SCRAM pgjdbc), wariant starszej linii wskazuje metodę z tym samym kodem.
  Gdy poprawka zmienia współdzielone narzędzie wykonywane przy starcie, wpis nazywa punkt
  wejścia z treści advisory (`role: entrypoint`).
- **`NOT_CODE_MAPPABLE`:** wpis `not_code: "<powód>"` (bez metod, wariantów, roli, pokrycia)
  → `NOT_APPLICABLE`, powód w notatce; `ResolverVersion` 6 → 7. Jedyny wpis: CVE-2025-49124
  (instalator Tomcata dla Windows). Grupa C okazała się mniejsza niż szacunek:
  CVE-2026-22733 (CloudFoundry) dotyczy kodu autokonfiguracji i dostał mapowanie.
- **`MAPPED_CODE_ABSENT`:** żadna klasa wybranego wariantu (także po relokacji) nie istnieje
  na classpath → `NOT_REACHABLE`, klasy w notatce; reguła luk B0 pominięta; wiersz v1 niesie
  kod, gdy niosą go wszystkie ustalenia. Klasa obecna bez metody zostaje `METHOD_NOT_FOUND`.

## Dowody

`UNKNOWN` (v2, ustalenia, `-roots all` CHA, offline):

| Aplikacja | Start | Po cz. 1 | Po cz. 2 | Po cz. 3 | Po #87 |
|---|---|---|---|---|---|
| petclinic | 27 | 19 | 10 | 3 | 1 |
| WebGoat | 26 | 17 | 6 | 1 | 0 |
| JPetStore | 5 | 3 | 3 | 3 | 0 |

- Każda zmiana werdyktu omówiona we wpisach IMPLEMENTATION_STATUS. Nowe `REACHABLE`: 2
  (część 1), 4 (część 2), 4 (część 3). Wiarygodne: CVE-2026-22732 (nagłówki Spring Security),
  CVE-2026-3260 i CVE-2020-14340 (`getParameterNames` w WebGoat). Pozostałe to nadmiarowe
  przybliżenia grafu (CHA/interfejsy, HTTP/2 wyłączony, sterowniki JDBC), także przy RTA.
  Trzy z nich (pgjdbc w WebGoat) usunął potem model `jdbc:drivers` (PR #88).
- Testy: `TestLoadManualMeta_embeddedTable` (107 wpisów, role 55/8/43 + 1 `not_code`),
  `TestServerManualMappings` (wybór wariantu według wersji, wydanie naprawione nic nie
  wybiera), `TestManualNotCode`, `TestFindingsForUnmatchedVerdicts`,
  `TestAnalyze_unmatchedMethodIsMethodNotFound`, `TestReasonCodesDocumented`;
  `go test ./...`, `go vet ./...` po każdej części i na połączonym `main` po #87–#89.
- `MAPPED_CODE_ABSENT`: porównanie v1 z `main` w 12 konfiguracjach (3 aplikacje ×
  `-roots all`/`profile` × CHA/RTA) — zmieniło się wyłącznie 6 oczekiwanych advisory.

## Ograniczenia, które zostają

- CVE-2022-21363 (MySQL 8.0.28): advisory Oracle nie nazywa komponentu, diff wydania nie ma
  zmiany do przypisania — `MAPPING_MISSING`.
- Mapowania na wspólne metody (np. `FromStringDeserializer$Std#_deserialize`,
  `ResultSetImpl#getObject`, `BeanDeserializerBase#createContextual`) są nadmiarowe z
  założenia: ścieżka oznacza użycie funkcji, nie warunku ataku.
- `MAPPED_CODE_ABSENT` zakłada, że wejście to wdrażany classpath; niepełne `-deps` może dać
  fałszywe `NOT_REACHABLE`.
- Tabela pokrywa wersje sprawdzone z nazwy; inne wersje dostają
  `MAPPING_VERSION_UNCOVERED`.
