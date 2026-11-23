# CVE:F8 — zakończone (2026-11-23)

## Zapis ukończenia

- **Podfazy:** F8a DONE (2026-11-13, identyfikacja runtime'u: `-jdk-home`/`-runtime-file`, `internal/runtimeinfo`); F8b DONE (2026-11-13 mechanizm, 2026-11-20 pierwsza kuracja: dostawca upstream OpenJDK przez kuratorowany katalog, 8 advisories z 2026, 88 CVE, wersja katalogu 1; `internal/runtimevuln`, `cmd/openjdk-catalog`, PR #118); **F8c zamknięte 2026-11-23 bez budowy** — decyzja użytkownika po ocenie (PR #131).
- **F8c — decyzja:** werdykt metodowy dla CVE JDK świadomie niebudowany: katalog podaje tylko komponent JDK, publicznego powiązania CVE → błąd JDK → commit brak, diff wydań CPU dałby metody per komponent, nie per CVE, a wczytanie klas runtime'u do grafu narusza kryterium 2. Zmierzony sygnał ekspozycji API komponentu (API nieużywane dla 3/9/30 z 88 CVE na petclinic/WebGoat/JPetStore) też niebudowany. Ustalenia runtime'u zostają: `applicability` z katalogu, `reachability=not_assessed` z powodem.
- **Kryteria ukończenia F8:** ten sam poziom bytecode'u przy dwóch runtime'ach daje różne wyniki (ocena wydania z `-jdk-home`); wydanie podatne i poprawione rozróżniane z katalogu (w tym wydania awaryjne, JDK 8); nieznany vendor, niepełny build, brak historii (runtime starszy niż katalog) i backport (inne dystrybucje: `not_performed` z powodem) mają jawne rozstrzygnięcia; adapter udokumentowany ([docs/RUNTIME_SOURCES.MD](../../../docs/RUNTIME_SOURCES.MD)) i nie udaje całego ekosystemu OpenJDK (tylko Temurin).
- **Weryfikacja:** zob. wpisy CVE:F8a, F8b, kuracja katalogu i ocena F8c w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD` (petclinic z runtime'ami Temurin dla każdej reguły).
- **Pozostałe ograniczenia:** inne dystrybucje (Corretto, Zulu, pakiety Debiana, Oracle JDK) nieoceniane; katalog tylko od 2026; brak werdyktu metodowego dla CVE JDK; F9 (integracja i kwalifikacja wydania) dla przyrostu F8 nie została powtórzona — zostaje w planie.

## Tekst fazy z planu

## F8. Docelowy runtime i pierwszy dostawca JDK

### Cel

Rozpoznawać podatności konkretnej dystrybucji i wydania runtime’u bez wnioskowania z samego bytecode’u.

### F8a. Identyfikacja i kontrakt źródła — DONE (2026-11-13; zob. IMPLEMENTATION_STATUS, [studium źródeł](../docs/RUNTIME_SOURCES.MD); pierwszy dostawca F8b: upstream OpenJDK)

1. Dodać `internal/runtimeinfo` z proponowanymi wejściami `-jdk-home` i `-runtime-file`; w MVP wzajemnie wykluczającymi się.
2. Czytać dostępne metadane instalacji i deklarowany opis. Nie uruchamiać automatycznie docelowego `java`; nie wybierać domyślnie runtime’u hosta skanera.
3. Zachować pełne surowe wartości i ostrożnie znormalizować vendor, dystrybucję, linię wersji, update, build, system, architekturę oraz ewentualny pakiet systemowy.
4. Nie utożsamiać ogólnego `IMPLEMENTOR` z jednoznacznie rozpoznanym produktem, jeśli dane nie wystarczają. Brakujące wymagane pola oznaczają nierozstrzygniętą ocenę.
5. Wybrać pierwszego dostawcę na podstawie docelowych wdrożeń. Oracle jest możliwym pilotażem, ale nie jest narzuconym wyborem użytkownika.
6. Przed budową adaptera zbadać dostępność, strukturę, sposób pobierania, warunki użycia oraz historyczne pokrycie danych. Jeśli brak stabilnego formatu, zacząć od wersjonowanego katalogu wpisów z odnośnikami do oficjalnych advisory.

### F8b. Ocena wydania — DONE dla mechanizmu (2026-11-13; katalog kuratorowany, pusty do pierwszej kuracji; zob. IMPLEMENTATION_STATUS)

1. Dodać interfejs `RuntimeAdvisoryProvider` i jeden adapter, np. w `internal/runtimevuln`.
2. W rekordzie zachować produkt, dystrybucję, zakresy/wyliczone wydania, ograniczenia platformy, źródło, daty, komponent JDK i informacje o poprawce, jeśli źródło je podaje.
3. Wykorzystać historię advisory w zadeklarowanym zakresie, nie tylko najnowszy biuletyn. Lista wspieranych wersji w jednym advisory nie stanowi automatycznie pełnego przedziału wszystkich starszych wydań.
4. Obsługiwać backporty za pomocą danych właściwego dostawcy lub dystrybucji. Nie przenosić numeru poprawionego wydania jednego produktu na inne produkty.
5. Raportować dopasowanie wydania i osiągalność niezależnie. Dla znanego podatnego wydania bez modelu kodu: `applicability=affected`, `reachability=not_assessed` i powód; to nie jest `NOT_REACHABLE`.
6. Przy braku opisu runtime’u wykazać `runtime assessment: not performed`. Przy nieobsługiwanym dostawcy wykazać brak pokrycia zamiast pustej listy jako dowodu bezpieczeństwa.
7. Podać zakres adaptera: dostawca, obsługiwane linie, okres danych, platformy, braki i datę walidacji.
8. Stosować ten sam mechanizm cache, świeżości i snapshotów co dla pozostałych źródeł.

### F8c. Osiągalność kodu JDK — dalszy przyrost

**Ocena 2026-11-23:** werdykt metodowy niewykonalny — katalog podaje tylko komponent, publicznego powiązania CVE → błąd JDK → commit brak; diff wydań CPU dałby metody per komponent, nie per CVE. Zmierzona alternatywa: raportowa ekspozycja API komponentu (petclinic/WebGoat/JPetStore: API nieużywane przez osiągalny kod dla 3/9/30 z 88 CVE). Do decyzji: `f8c-1` (sygnał raportowy, `reachability` dalej `not_assessed`) albo zamknięcie F8c. Zob. IMPLEMENTATION_STATUS.

1. Tylko dla przypadków z wiarygodnym mapowaniem na kod zaprojektować odczyt właściwych klas runtime’u i model granic analizy.
2. Nie znosić globalnie ograniczeń CHA dla `java.*`; wymagałoby to osobnej oceny rozmiaru grafu i liczby fałszywych ścieżek.
3. Podatności JIT, VM i kodu natywnego pozostawić bez werdyktu metodowego, jeśli model analizy ich nie obejmuje.

F8c nie jest warunkiem pierwszego wydania oceny podatności JDK.

### Kryteria ukończenia

- Ten sam poziom bytecode’u przy dwóch różnych runtime’ach nie prowadzi automatycznie do tych samych wyników podatności.
- Wydanie podatne i poprawione są rozróżniane na podstawie właściwego źródła.
- Nieznany vendor, niepełny build, brak danych historycznych i backport mają jawne rozstrzygnięcia lub brak rozstrzygnięcia.
- Pierwszy adapter jest udokumentowany i nie udaje wsparcia całego ekosystemu OpenJDK.
