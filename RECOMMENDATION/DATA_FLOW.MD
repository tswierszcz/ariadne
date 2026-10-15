# Przepływ wartości i danych — rekomendacja funkcjonalna

Status: plan. Fazy: [DATA_FLOW_IMPLEMENTATION_PLAN.MD](DATA_FLOW_IMPLEMENTATION_PLAN.MD). Zasady: [ROADMAP.MD](ROADMAP.MD).

## Cel i punkt wyjścia

Najpierw dostarczyć wspólne fakty o wartościach potrzebne innym mechanizmom, a następnie badać możliwość przepływu danych z określonego źródła do argumentu operacji związanej z CVE.

Obecne `internal/classfile` zachowuje surowy bytecode i wywołania, ale `code.go` pomija tabelę wyjątków. Obecny BFS w `internal/reach` analizuje wywołania metod, a `internal/exploit` daje ocenę LLM na podstawie ścieżek. Żaden z tych mechanizmów nie jest analizą przepływu danych.

## Dwa przyrosty funkcjonalne

### Wartości dla modeli JVM

- Rozpoznać, jaki napis, typ, obiekt albo zbiór możliwych wartości dochodzi do wywołania.
- Zachować wartości przez stos JVM, zmienne lokalne, gałęzie, argumenty i wyniki metod w zadeklarowanym zakresie.
- Dostarczyć te dane refleksji, ServiceLoaderowi, callbackom, uchwytom i modelom frameworka.
- Nie traktować nazw parametrów/debug symboli jako warunku podstawowego działania.

Przykład: nazwa klasy przypisana do zmiennej przed `Class.forName` pozostaje rozpoznawalna, jeśli jej przepływ mieści się w obsługiwanym zakresie.

### Dane kontrolowane przez użytkownika — taint

- Zdefiniować wersjonowane źródła danych, np. wskazany argument endpointu HTTP, i sinki, np. konkretny argument metody związanej z CVE.
- Śledzić propagację przez przypisania, wybrane operacje na obiektach i wywołania.
- Uwzględniać zabezpieczenia tylko w odpowiednim kontekście, zgodnie z regułą podatności; nie istnieje uniwersalna metoda „sanityzująca” wszystkie dane.
- Pokazać ścieżkę danych i założenia. Samo wywołanie metody nie wystarcza do stwierdzenia dopływu danych kontrolowanych przez użytkownika.

## Wyniki i granice

Proponowana osobna ocena: `flow_found`, `no_flow_in_model`, `undetermined` lub `not_assessed`, wraz z modelem źródła/sinka. Nie zastępuje to statusu reachability i nie dowodzi wykorzystania CVE.

Przykłady: dopływ do jednego przeciążenia nie oznacza dopływu do drugiego; stała przekazana do sinka pozostaje różna od danych żądania; nieznane native wywołanie na ścieżce daje niepewność, a nie dowód odcięcia przepływu.

MVP obejmuje ograniczony przepływ lokalny i jawne streszczenia prostych metod, a pierwszy taint — mały katalog źródeł i sinków dla wybranego scenariusza. Pełna analiza współbieżności, wszystkich pól, refleksji i bibliotek nie jest warunkiem pierwszego przyrostu. Każda luka lub limit jest widoczny.

## Relacje z innymi planami

Ten obszar jest właścicielem wspólnego IR, CFG i wartości B2. Późniejsze points-to w GRAPH_PRECISION rozszerza modele obiektów; nie buduje drugiego silnika stosu. FRAMEWORKS dostarcza informacje o źródłach i aktywnych konfiguracjach. CVE dostarcza zweryfikowane mapowanie sinków i ich wersje.

Gotowość oznacza poprawne przykłady dodatnie, ujemne i nieznane, rozróżnienie instancji i argumentów oraz wykaz modelowanych operacji. Brak znalezionego przepływu bez pełnego pokrycia modelu nie jest stwierdzeniem bezpieczeństwa.

## Źródło

[CodeQL: przepływ danych i taint w Java/Kotlin](https://codeql.github.com/docs/codeql-language-guides/analyzing-data-flow-in-java/). To odniesienie do rodzaju analizy, nie deklaracja integracji z CodeQL.
