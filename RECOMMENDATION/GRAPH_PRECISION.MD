# Precyzja grafu: RTA i points-to — rekomendacja funkcjonalna

Status: plan. Fazy: [GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD](GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD). Wspólne zasady: [ROADMAP.MD](ROADMAP.MD).

## Cel i stan obecny

Ograniczać nieuzasadnione cele wywołań wirtualnych bez ukrywania braków modelu i bez utraty uzasadnionych ścieżek.

Obecny `internal/callgraph/build.go` stosuje CHA, z pomijaniem pewnych prefiksów przez `skip.go`. Zanim porówna się algorytmy, JVM G2 musi naprawić rozróżnienie sygnatur i rozwiązanie implementacji. Samo usunięcie przypadkowych overloads nie jest dowodem zalet RTA.

## Tryby docelowe

| Tryb | Zasada | Główne założenia |
|---|---|---|
| CHA | Dopuszczalne implementacje wynikają z hierarchii w zadanym kontekście klas. | Obecny zbiór klas i przyjęte reguły dispatch/skip. |
| RTA | Dodatkowo ogranicza implementacje do typów, które mogą być tworzone w analizowanym zakresie. | Alokacje jawne i modelowane pośrednio, właściwe korzenie, iteracja. |
| Points-to | Dla konkretnego receivera przybliża obiekty/typy, które mogą do niego trafić. | Model wartości, pól, kontekstu i kodu zewnętrznego. |

Są to tryby eksperymentalne do czasu kwalifikacji. Wybór i porównanie są jawne, np. proponowane `-callgraph-mode`; istniejące polecenia nie przełączają się automatycznie na najwęższy graf.

## Scenariusze

- Interfejs ma trzy implementacje, w osiągalnym kodzie powstaje jedna: RTA może wyeliminować pozostałe przy wystarczającym pokryciu alokacji.
- Instancję tworzy DI, ServiceLoader, refleksja albo deserializer: ten typ musi wejść do zbioru RTA, nawet bez zwykłego `new` w aplikacji.
- Typ może powstać przez nieznany mechanizm: raport ujawnia lukę i określoną politykę fallback, zamiast dowodzić nieosiągalności usuniętych implementacji.
- Dwie instancje interfejsu trafiają do różnych miejsc: points-to może dać precyzję większą niż globalna lista RTA.

## Raport i zasady interpretacji

Wynik zawiera tryb, korzenie, reguły typów tworzonych pośrednio, fallbacki, przerwania i dowody zmian celów. Porównanie pokazuje dodane/usunięte krawędzie i zmiany wyników dla tej samej konfiguracji i danych CVE.

RTA przy `roots=all` może tworzyć dużo typów, również z nieużywanych w runtime metod aplikacji; mniejsza poprawa jest wtedy oczekiwaną konsekwencją zakresu. Nie wolno potajemnie zawężać korzeni, aby wykazać lepszy wynik.

Pełne CHA byłoby szerokim przybliżeniem w określonym modelu, ale obecny skip JDK i inne braki powodują, że bieżący graf Ariadne nie jest automatycznie formalnie pełnym nadzbiorem wszystkich rzeczywistych wywołań. Porównania jakości muszą ten fakt uwzględniać.

## Gotowość i źródło

Gotowość oznacza zachowanie wszystkich oczekiwanych celów w zestawie kontrolnym, wyjaśnienie redukcji, jawne traktowanie nieznanych alokacji i pomiar czasu/pamięci. Mniej krawędzi i mniej `REACHABLE` nie są samodzielnym dowodem poprawy.

[SootUp: CHA i RTA](https://soot-oss.github.io/SootUp/develop/callgraphs/) jest punktem odniesienia; integracja z tym narzędziem nie jest warunkiem realizacji.
