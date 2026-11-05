# SERVICE_LOADING:S3 — JPMS i warstwy — archiwum

**Data archiwizacji:** 2026-11-05.
**Stan:** DONE dla zadeklarowanego zakresu. Decyzja użytkownika: pozostałe
punkty (lista niżej) są jawnymi ograniczeniami, nie ukrytą pracą — tak samo
jak przy S2. Obszar SERVICE_LOADING pozostaje aktywny: S4 (regresja i
rozszerzenia API) jest otwarte.

Oryginalny tekst fazy jest na końcu pliku; w
`RECOMMENDATION/SERVICE_LOADING_IMPLEMENTATION_PLAN.MD` zastąpiono go
wskaźnikiem. Bieżące zachowanie opisują `docs/SUPPORT.MD` (wiersz
"ServiceLoader and JPMS service models"), README (flagi `-jpms-scenario`,
`-jpms-module-path`, ograniczenia ServiceLoader/JPMS) i FAQ. Szczegółowe wpisy
trzech wycinków zostają w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`. Ten plik
jest zapisem historycznym, nie źródłem prawdy o bieżącym stanie.

## Commity

1. `749c380` — wycinek 1: dekodowanie atrybutu `Module`
   (`internal/classfile/module.go`), inwentarz deskryptorów i `provides`
   (`internal/serviceload/modules.go`), niezgodności
   module-info↔META-INF/services, sekcja raportu "JPMS module descriptors";
   fixture `testdata/jpms`.
2. `ef9d2f0` — wycinek 2: `-jpms-scenario module-path` (`provides` jawnych
   modułów, `provider()` jako cel, automatic modules przez
   `META-INF/services`, kontrola `uses`, ostrzeżenie o brakującym
   `requires`); aplikacja kontrolna `scripts/prepare-jpms-demo.sh`.
3. `5f71000` (+ `c17cdfd`, gofmt) — wycinek 3: `-jpms-scenario mixed` z
   `-jpms-module-path`, `load(ModuleLayer.boot(), S)`; nieznana warstwa jako
   luka.

## Kryteria „Gotowe” i dowody

| Kryterium | Dowód |
|---|---|
| classpath, named module, automatic module i niedostępny provider są rozróżnione | `TestInventoryModules_compiledFixture`, `TestInventoryModules_classpathSemanticsUnchanged`, `TestScenario_providerSetsDiffer`, `TestScenario_mixed`; demo: CVE-2022-42889 NOT_REACHABLE (classpath) vs REACHABLE (module-path, mixed), zgodnie z realnym JVM (0 vs 1 provider) |
| Nieznana warstwa nie daje pewnego negatywnego wyniku | `TestScenario_unknownLayerKeepsGap` (kandydaci zostają + luka); `ModuleLayer.boot()` modelowane dokładnie (`TestScenario_bootLayer`) |
| Dane modułu poza indeksem klas; atrybut `Module`, `uses`, `provides` | `TestParse_moduleInfoFromCompiledFixture`; hierarchia nadal pomija `module-info.class` |
| Reguła `provider()` nie stosowana automatycznie do classpath | `TestScenario_providerSetsDiffer` (FactoryCodec tylko na module path), automatic modules tylko konstruktor |
| Ostrzeżenia o nieuruchamialnej konfiguracji | `TestApplyScenario_missingRequiresWarns`, `TestApplyScenario_mixedWarnings`; demo: `FindException` w JVM ↔ ostrzeżenie ariadne |

Weryfikacja na realnych aplikacjach (offline, `-profile spring`): petclinic i
WebGoat w scenariuszu domyślnym — `results.json` bajtowo identyczny z
baseline'em po każdym wycinku; `module-path` — statusy bez zmian (krawędzie
modelu 4→2, 75→73), prawdziwe ostrzeżenia o `requires` HikariCP 3.4.5;
żadna z aplikacji nie wywołuje `load(ModuleLayer, …)`.

## Co nie zostało zrobione (jawne ograniczenia)

- Pełne rozwiązywanie modułów i czytelność: tylko ostrzeżenia o brakującym
  `requires`; nie odtwarzamy, że main z classpath bez `--add-modules` nie
  rozwiązuje modułów z module path (założenie `ALL-MODULE-PATH`).
- Kwalifikowane eksporty/`opens`, split packages.
- Warstwy inne niż `ModuleLayer.boot()` — kandydaci + luka widoczności.
- Dokładny dispatch po obiekcie z `provider()` — luka, zostają krawędzie CHA.
- Wariant modułu w kluczu modelu i snapshot CVE F3 — nietknięte (CVE F3 nie
  istnieje).
- Słowa kluczowe Javy w nazwach automatycznych nie są sprawdzane; jary
  automatyczne w raporcie tylko liczone.

---

## Oryginalny tekst fazy

## S3. JPMS i warstwy


Zależności: S2, JVM G4; zaawansowane wartości mogą korzystać z DATA_FLOW D3.

- Rozszerzyć `internal/classfile` o dane atrybutu `Module`, `uses` i `provides`. Zachować deskryptor modułu poza zwykłym indeksem klas.
- Utworzyć jawny model module path i czytelności w analizowanym scenariuszu. Brak scenariusza nie oznacza poprawnego odtworzenia pełnej konfiguracji runtime’u.
- Rozwiązywać właściwe reguły dostawcy, m.in. metodę `provider()` dla odpowiednich rodzajów modułów; nie stosować jej automatycznie do każdego wpisu classpath.
- Dodać ograniczoną deklarację warstw i rodziców. Dynamicznie budowane nieznane warstwy oznaczać jako lukę do uzupełnienia śladem runtime’u.
- Zachować wszystkie dane w snapshotcie CVE F3; uwzględnić wariant modułu w kluczu modelu.

Gotowe: classpath, named module, automatic module i niedostępny provider są rozróżnione. Nieznana warstwa nie daje pewnego negatywnego wyniku.

