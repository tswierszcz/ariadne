# Refleksja — rekomendacja funkcjonalna

Status: plan. Fazy: [REFLECTION_IMPLEMENTATION_PLAN.MD](REFLECTION_IMPLEMENTATION_PLAN.MD). Fundament wartości: [DATA_FLOW.MD](DATA_FLOW.MD).

## Cel i stan obecny

Odtwarzać cele refleksji na podstawie dostępnych wartości nazw klas i metod, zachowując nieznane przypadki. Obecny ekstraktor widzi wywołania API refleksji, ale nie odtwarza ich docelowych metod jako krawędzi.

Nie wystarczy dodać krawędzi od każdego `Class.forName` do wszystkich metod znalezionej klasy. Ładowanie, inicjalizacja, wyszukanie obiektu `Method` i jego wywołanie są oddzielnymi zdarzeniami.

## Zachowanie

- Rozpoznawać ograniczony zbiór API: `Class.forName`, znane `ClassLoader.loadClass`, `getMethod`, `getDeclaredMethod`, konstruktory i `Method.invoke`/`Constructor.newInstance`.
- Przenosić znane napisy, literały klas, typy parametrów i obiekty refleksji przez zmienne; później przez streszczenia metod i konfigurację.
- Zachować reguły wyszukiwania deklarowanego/odziedziczonego członka, zakres dostępu i właściwy kontekst loadera. Sukcesu wyszukania nie zakładać przy sprzecznych lub niepełnych danych.
- Uwzględniać argument `initialize` w odpowiednich wariantach `forName`; `loadClass` samo w sobie nie jest ogólną regułą wywołania `<clinit>`.
- Dopiero użycie obiektu refleksji tworzy odpowiednie możliwe wywołanie. Dla wirtualnej metody zachować możliwe typy receivera, a dla konstruktora przekazać alokację do RTA.
- Dla nazwy z kilku gałęzi zachować wszystkie uzasadnione cele; nieznana nazwa lub przekroczony limit pozostają luką.

### Scenariusze

1. `Class.forName("demo.Parser")` → wyszukanie `parse(String)` → `invoke`: modelowany cel i dowód w raporcie.
2. Znana klasa została tylko znaleziona: brak automatycznej ścieżki do `parse`.
3. Nazwa z requestu: miejsce refleksji o nieznanym celu; nie arbitralna lista wszystkich klas.
4. Konfiguracja wskazuje klasę, ale aktywny profil jest nieznany: cele warunkowe z opisem scenariusza.

## Wynik i granice

Raport pokazuje miejsce refleksji, wartości/źródło nazwy, rozwiązane cele, regułę, dostęp i nieznane warunki. Deklaracje użytkownika służą jako oznaczone założenia, nie jako obserwacja runtime’u.

MVP obejmuje publiczne metody i konstruktory przy jednoznacznych stałych lokalnych. Dalej: odczyt konfiguracji i przepływ między metodami, reguły dostępu modułowego oraz większe zbiory typów. `setAccessible` nie jest automatycznie skuteczne dla wszystkich wersji Javy i modułów.

Brak rozstrzygnięcia refleksji relewantnej dla CVE ogranicza negatywny wynik. Pozytywna krawędź jest statyczną możliwością, nie dowodem wykonania.

## Gotowość i źródła

Wymagane są przykłady literału, zmiennej, dwóch overloads, dziedziczenia, braku metody, nieznanej nazwy i kontroli inicjalizacji. Modele muszą zachować różnicę między wyszukaniem a użyciem.

- [Class: API wyszukiwania i ładowania](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/lang/Class.html).
- [WALA: istniejące podejścia do modelowania refleksji](https://wala.github.io/javadoc/com/ibm/wala/ipa/callgraph/AnalysisOptions.html).

WALA jest punktem odniesienia, nie nową zależnością projektu.
