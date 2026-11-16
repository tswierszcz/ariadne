# Tryb biblioteki — rekomendacja funkcjonalna

Status: plan (2026-11-16), bez implementacji. Fazy: [LIBRARY_MODE_IMPLEMENTATION_PLAN.MD](LIBRARY_MODE_IMPLEMENTATION_PLAN.MD). Zasady wspólne: [ROADMAP.MD](ROADMAP.MD).

## Cel

Ariadne odpowiada dziś na pytanie aplikacji: „czy mój kod może dojść do podatnego kodu
w moich zależnościach?”. Dla biblioteki pytanie jest inne: **„czy ktokolwiek, kto użyje
mojego publicznego API, może przez nie dojść do podatnego kodu w moich zależnościach —
i przez które metody?”**. Odpowiedź jest potrzebna autorowi biblioteki (co naprawić,
czy podbić zależność, co napisać w advisory) i jej użytkownikom (czy metody, których
używają, są narażone).

## Stan obecny

- Jar biblioteki można podać jako `-app`, a jej zależności przez `-deps`. Z `-roots all`
  korzeniami są **wszystkie** metody jara wejściowego (`reach.AppRoots`): także prywatne,
  pakietowe i klasy wewnętrzne. To przybliżenie bezpieczne (nie przeoczy ścieżki z API),
  ale nadmiarowe: wskazuje ścieżki z wewnętrznych pomocników, do których użytkownik nie
  ma dostępu.
- Wynik podaje jedną najkrótszą ścieżkę na advisory × instancję. Nie mówi, **które
  metody API** prowadzą do podatności.
- Wywołania zwrotne do kodu użytkownika (interfejsy do zaimplementowania, metody do
  nadpisania) nie mają w wejściu implementacji i nie są oznaczane jako luka.
- Wersje zależności to te, które dostarczy `-deps`. Aplikacja używająca biblioteki może
  dostać inne (mediacja Maven/Gradle, zakresy, `provided`, `optional`).

## Wymagania

### Korzenie: publiczne API

Korzeniami są metody, które użytkownik biblioteki może wywołać albo nadpisać:

- publiczne i chronione metody oraz konstruktory publicznych klas (także publicznych
  klas zagnieżdżonych w publicznych klasach), metody domyślne i statyczne publicznych
  interfejsów;
- inicjalizacja tych klas (`<clinit>`), bo uruchamia ją pierwsze użycie;
- wejścia, które biblioteka sama ogłasza: dostawcy z `META-INF/services`, klasy z
  autokonfiguracji frameworka (np. `spring.factories`, `AutoConfiguration.imports`),
  klasy wskazane w plikach konfiguracyjnych biblioteki.

Metody prywatne i pakietowe nie są korzeniami. Gdy biblioteka ma `module-info`, w
scenariuszu JPMS korzeniami są tylko pakiety eksportowane (`exports`); w scenariuszu
classpath — wszystkie publiczne pakiety. Pakiety „wewnętrzne” z konwencji
(`.internal.`, `.impl.`) są publiczne dla JVM: domyślnie są API; wyłączenie ich jest
jawną opcją użytkownika, raportowaną w wyniku.

### Wynik: narażone metody API

Dla każdego advisory × instancja zależności wynik podaje **listę metod API, z których
podatny kod jest osiągalny**, a dla każdej z nich przykładową ścieżkę i jej charakter
(dowód krawędzi jak dziś). Status `REACHABLE` znaczy „osiągalny z publicznego API”, nie
„wykonywany przez aplikację”. Wynik v2 dostaje to jako pole addytywne; v1 i obecne pola
zachowują znaczenie.

### Kod klienta jako niewiadoma

Biblioteka woła kod użytkownika, którego nie ma w wejściu: implementacje jej interfejsów,
podklasy nadpisujące metody, lambdy przekazane do API. Takie miejsca są jawną luką
„kod klienta” (słownik luk B0), a nie cichym brakiem krawędzi. Ścieżka do podatnego kodu
**przez** kod klienta (biblioteka woła callback, callback woła z powrotem bibliotekę)
jest możliwa, ale zależy od użytkownika — raportowana jako warunkowa.

### Zależności i wersje

- Wersje rozwiązane z deklaracji biblioteki są punktem wyjścia; wynik mówi, którą wersję
  zbadano.
- Zależności `provided`/`optional` dostarcza użytkownik: są w analizie, ale oznaczone.
- Dalszy krok: sprawdzenie advisory dla zakresu wersji, które użytkownik może dostać
  (zakres deklarowany, najnowsze wydania linii), z jawnym oznaczeniem wersji, dla których
  mapowanie nie ma wariantu.
- Zależności zacienione (shading) obsługuje istniejące CVE F7.

## Granice

- Tryb biblioteki nie mówi, czy konkretna aplikacja jest narażona — to robi zwykła
  analiza aplikacji (z biblioteką jako zależnością).
- Osiągalność z API nie oznacza, że dane atakującego tam dotrą; taint (DATA_FLOW D4)
  pozostaje osobnym, opcjonalnym krokiem.
- Runtime (RUNTIME_NATIVE) nie jest częścią trybu: biblioteka nie ma czego uruchomić;
  testy biblioteki to tylko przykładowe użycie.
