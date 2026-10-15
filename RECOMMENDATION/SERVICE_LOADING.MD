# ServiceLoader i JPMS — rekomendacja funkcjonalna

Status: plan. Fazy: [SERVICE_LOADING_IMPLEMENTATION_PLAN.MD](SERVICE_LOADING_IMPLEMENTATION_PLAN.MD). Kontrakty: [ROADMAP.MD](ROADMAP.MD).

## Cel i stan obecny

Rozpoznawać implementacje usług wybierane przez `ServiceLoader`, zachowując informację, kiedy samo wykrycie dostawcy przechodzi w możliwość utworzenia obiektu i wywołania jego metod.

Obecny model syntetycznych krawędzi w `internal/callgraph/synthetic.go` dotyczy reguł profilu. Nie interpretuje rejestrów usług; `hierarchy/jarread.go` pomija `module-info.class`. Potrzebne są dane zasobów i modułów, a nie wyłącznie kolejne wpisy YAML.

## Zachowanie

- Dla classpath zbierać dostawców z `META-INF/services/<Interface>`, z pochodzeniem każdego wpisu i obsługą powtórzeń.
- Przy znanym argumencie `ServiceLoader.load(Service.class)` powiązać konkretne miejsce z daną usługą i właściwym kontekstem ładowania.
- Rozróżnić utworzenie loadera, iterację/odkrycie dostawcy, `Provider.type()`, `Provider.get()` i użycie obiektu. Lazy loading nie uzasadnia natychmiastowego wywołania każdej metody każdego dostawcy.
- Dla znanego użycia utworzyć warunkowe cele konstruktora/fabryki i właściwy dispatch metod usługi, a typy dostawców przekazać do RTA.
- Dla module path odczytać `uses`/`provides`, widoczność, kontekst modułów i dopuszczalne metody `provider()`. Uwzględnić wybrany wariant deskryptora modułu.
- Dla nieznanej usługi lub warstwy modułów pokazać nierozstrzygnięte miejsce; nie dodawać wszystkich usług z całej aplikacji jako pewnych celów.

### Przykład

Aplikacja znajduje `Codec` przez ServiceLoader i woła `decode` na dostawcy. Ariadne pokazuje: miejsce użycia loadera → model utworzenia dostawcy → jego `decode` → podatna metoda. Plik usług bez użycia loadera pozostaje informacją inwentaryzacyjną.

## Wejście i raport

Zasoby mogą pochodzić z aplikacji, zwykłych zależności i dostępnych modułów. Lista classpath/module path i opcjonalne warstwy muszą być zadeklarowane lub opisane jako założenie. Normalizacja archiwum ma zachować pochodzenie zasobu, również w kilku modułach.

Raport pokazuje nazwę usługi, dostawcę, rejestrację, miejsce materializacji, model i ograniczenia. Deklaracja dostawcy, możliwy cel i zaobserwowane wykonanie są różnymi informacjami.

## Zakres i gotowość

MVP: classpath, znany literał klasy, iterator i prosty `Provider.get`. Rozszerzenie: wartości przekazywane między metodami, JPMS i jawne scenariusze warstw. Niestandardowy classloader z runtime’u wymaga informacji z [RUNTIME_NATIVE](RUNTIME_NATIVE.MD).

Gotowość: dostawca używany jest uwzględniony, samo zadeklarowanie nie tworzy ścieżki do wszystkich jego metod, a brak klasy/fabryki i nieznana usługa mają jawne powody.

## Źródło

[Dokumentacja ServiceLoader](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/ServiceLoader.html). Model ma być zweryfikowany dla wersji Javy obsługiwanych w danym wydaniu Ariadne.
