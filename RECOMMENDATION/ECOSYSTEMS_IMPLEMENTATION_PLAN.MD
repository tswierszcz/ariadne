# Inne ekosystemy — fazy E0–E4

Status: plan eksploracyjny. Wymagania: [ECOSYSTEMS.MD](ECOSYSTEMS.MD). Zależności: [ROADMAP.MD](ROADMAP.MD).

## E0. Ocena kontraktu i zakresu pilota

Zależności: B0, wynik v2 i komponenty CVE F1.

- Sprawdzić konkretną wersję `govulncheck`, format ustrukturyzowany, znaczenie findingów, tryby źródła/binarium, zasady dostępu do bazy i warunki użycia.
- Przygotować nagrane/syntetyczne wyniki źródłowe i binarne. Nie uznać, że każdy ma taką samą informację o call paths.
- Zdefiniować tabelę mapowania pól na model wspólny wraz z polami, których nie można uczciwie znormalizować.
- Wybrać importer jako pierwszy rezultat. Własny frontend SSA lub analiza Go machine code wymagają odrębnego uzasadnienia, nie są domyślną pracą E0.

Gotowe: kontrakt wskazuje poziom dowodów dla każdego typu findingu i jawne różnice względem JVM.

## E1. Odseparowanie raportowania od JVM

- Wydzielić wspólny model `Finding`, komponentów, referencji i pokrycia z zależności od GAV/hierarchii klas. Adaptować dotychczasowe wyniki JVM bez zmiany ich znaczenia.
- Wprowadzić ekosystem i dostawcę analizy jako pola, zachować oryginalne advisory ids i aliasy. Nie scalać różnych komponentów/ekosystemów tylko po CVE.
- Rozszerzyć `internal/report` o formaty i poziomy dowodów niezależne od metod JVM. Sekcje FRAMEWORK/CHA pokazują się tylko dla odpowiedniego rodzaju analizy.
- `internal/reach` nie musi być przepisywane dla importera. Gdy naprawdę będzie potrzebny wspólny graf, oddzielić abstrakcyjny rdzeń od frontendów jako osobny przyrost.

Gotowe: raport JVM zachowuje dane, a Go może pokazać symbol lub oryginalny trace bez fikcyjnego deskryptora JVM.

## E2. Import govulncheck i snapshot

Zależności: E1, CVE F3 dla deklaracji odtwarzania.

- Utworzyć proponowane `internal/adapters/govulncheck` i jawne wejście importu, np. osobny subcommand. Szczegóły CLI ustalić bez zmiany istniejącej komendy JVM.
- Walidować format/wersję, obsłużyć strumień komunikatów, config/progress/error i finding zgodnie z wybraną wersją. Nie ignorować błędu narzędzia tylko dlatego, że wcześniej wystąpił finding.
- Zachować surowy raport i metadane; różne finding levels tłumaczyć zgodnie z E0. Brak call trace nie jest domyślnie `NOT_REACHABLE`.
- Zapisać hash wejścia, konfigurację, wersję dostawcy i jego dane o bazie w snapshotcie. Jeśli oryginalny raport nie zawiera czasu/stanu bazy, oznaczyć to jako nieznane.
- Import offline nie wysyła nowych zapytań i nie uruchamia Go. Aktualność zaimportowanego raportu pozostaje oddzielna od daty jego importu.

Gotowe: prawidłowe wyniki obu trybów są zachowane; błąd/nieznany format i niepełne dane mają czytelną obsługę. E0–E2 to MVP innych ekosystemów.

## E3. Jawne uruchomienie lokalnego narzędzia

- Dodać opcjonalny adapter uruchamiający wskazaną wersję `govulncheck` z argumentami przekazanymi bez powłoki, z timeoutem i zachowaniem exit code/stderr.
- Obsłużyć osobno źródła i binarium. W scenariuszu źródłowym zapisać workspace/moduły/replace, build tags, platformę, toolchain i warunki zależności; nie zgadywać brakującej wersji lokalnego replacement.
- Określić pełny kontrakt offline dostawcy, także zależności i bazy. Jeśli adapter nie potrafi zagwarantować offline, zgłasza brak tej możliwości zamiast wykonywać sieć poza kontrolą Ariadne.
- Nie używać daty uruchomienia jako dowodu świeżości bazy. Zmiana bazy lub narzędzia jest widoczna w metadanych i snapshotcie.
- Uruchomienie korzysta z istniejącego narzędzia; instalacja/aktualizacja to osobna jawna operacja użytkownika.

Gotowe: źródła/binarium, brak narzędzia, timeout, błąd kompilacji, zależność lokalna i niekompletne offline mają poprawne rozstrzygnięcia.

## E4. Kwalifikacja i decyzja o dalszym rozwoju

- Porównać wyniki adaptera z surowym wyjściem tej samej wersji narzędzia: bez utraty modułów, symboli, przesłanek i błędów.
- Sprawdzić źródła i binarium o znanej zawartości, aliasy CVE, lokalny replace, warunkowe pliki i brak danych wersji.
- Wykonać B4 dla kodu Go; integracje z realnym narzędziem są osobne i używają zapisanych danych/konfiguracji bez niejawnej instalacji.
- Udokumentować różnice precyzji, wspierane wersje i koszty utrzymania. Na podstawie pilota zdecydować, czy rozszerzać integrację, dodać drugi ekosystem czy uzasadnić własny frontend.

Sam importer nie obiecuje pełnej reprodukcji zewnętrznej analizy po zmianie toolchainu. Odtwarzanie raportu i ponowne wykonanie analizy źródłowej są odrębnymi możliwościami.
