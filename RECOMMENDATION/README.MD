# Rekomendacje rozwoju Ariadne — indeks

Status wszystkich nowych obszarów: plan, bez implementacji w ramach przygotowania tych dokumentów.

Punkt wyjścia: [oryginalne rekomendacje](MD). Wspólne zależności, zasady i kolejność: [ROADMAP.MD](ROADMAP.MD). Istniejący zakres rozpoznawania CVE pozostaje w osobnej parze dokumentów.

Do realizacji w nowej sesji użyj [promptu startowego i skilla](START_IMPLEMENTATION.MD).
Postęp zapisuje [IMPLEMENTATION_STATUS.MD](IMPLEMENTATION_STATUS.MD), a faktycznie dostępne
możliwości i ograniczenia opisuje [SUPPORT.MD](../docs/SUPPORT.MD).

Poradnik dla człowieka: [INSTRUCTION FOR USER.MD](../INSTRUCTION%20FOR%20USER.MD).
AI ma pomijać ten poradnik przy starcie i odczytywać go wyłącznie na wyraźną
prośbę użytkownika dotyczącą jego treści.

## Dokumenty według obszaru

| Obszar | Plan funkcjonalny | Plan implementacyjny |
|---|---|---|
| Podstawy grafu JVM: inicjalizacja, dziedziczenie, wybór klas | [JVM_GRAPH.MD](JVM_GRAPH.MD) | [JVM_GRAPH_IMPLEMENTATION_PLAN.MD](JVM_GRAPH_IMPLEMENTATION_PLAN.MD) |
| ServiceLoader i moduły JPMS | [SERVICE_LOADING.MD](SERVICE_LOADING.MD) | [SERVICE_LOADING_IMPLEMENTATION_PLAN.MD](SERVICE_LOADING_IMPLEMENTATION_PLAN.MD) |
| Refleksja i rozpoznawanie nazw klas/metod | [REFLECTION.MD](REFLECTION.MD) | [REFLECTION_IMPLEMENTATION_PLAN.MD](REFLECTION_IMPLEMENTATION_PLAN.MD) |
| Frameworki: proxy, AOP, DI, konfiguracja i zewnętrzne wejścia | [FRAMEWORKS.MD](FRAMEWORKS.MD) | [FRAMEWORKS_IMPLEMENTATION_PLAN.MD](FRAMEWORKS_IMPLEMENTATION_PLAN.MD) |
| Callbacki i wykonywanie asynchroniczne | [ASYNC_CALLBACKS.MD](ASYNC_CALLBACKS.MD) | [ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD](ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD) |
| MethodHandle, invokedynamic i dynamiczne stałe | [DYNAMIC_LINKAGE.MD](DYNAMIC_LINKAGE.MD) | [DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD](DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD) |
| Deserializacja i kandydaci łańcuchów gadgetów | [DESERIALIZATION.MD](DESERIALIZATION.MD) | [DESERIALIZATION_IMPLEMENTATION_PLAN.MD](DESERIALIZATION_IMPLEMENTATION_PLAN.MD) |
| Precyzja grafu: RTA i analiza możliwych obiektów | [GRAPH_PRECISION.MD](GRAPH_PRECISION.MD) | [GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD](GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD) |
| Przepływ wartości i danych kontrolowanych przez użytkownika | [DATA_FLOW.MD](DATA_FLOW.MD) | [DATA_FLOW_IMPLEMENTATION_PLAN.MD](DATA_FLOW_IMPLEMENTATION_PLAN.MD) |
| Obserwacja runtime’u, dynamiczne klasy i JNI | [RUNTIME_NATIVE.MD](RUNTIME_NATIVE.MD) | [RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD](RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD) |
| Inne języki: pilotaż integracji Go | [ECOSYSTEMS.MD](ECOSYSTEMS.MD) | [ECOSYSTEMS_IMPLEMENTATION_PLAN.MD](ECOSYSTEMS_IMPLEMENTATION_PLAN.MD) |
| Rozpoznawanie podatności: wersje, poprawki, shading, cache, JDK | [CVE.MD](CVE.MD) | [CVE_IMPLEMENTATION_PLAN.MD](CVE_IMPLEMENTATION_PLAN.MD) |

## Pokrycie pierwotnych rekomendacji

| Punkt dokumentu źródłowego | Dokument odpowiedzialny |
|---|---|
| Inicjalizacja klas `<clinit>` | JVM_GRAPH |
| Wywołania odziedziczonych metod | JVM_GRAPH; wspólna realizacja z CVE F4 |
| ServiceLoader i JPMS | SERVICE_LOADING |
| Refleksja przez literały, zmienne i konfigurację | REFLECTION; wartości dostarcza DATA_FLOW, konfigurację FRAMEWORKS |
| Proxy, AOP i dependency injection | FRAMEWORKS |
| Callbacki i kod asynchroniczny | ASYNC_CALLBACKS |
| MethodHandle i pozostałe invokedynamic | DYNAMIC_LINKAGE |
| Deserializacja | DESERIALIZATION |
| Konfiguracja i wejścia frameworka | FRAMEWORKS |
| Dynamiczne ładowanie klas i JNI | RUNTIME_NATIVE; statyczne nazwy klas obsługuje również REFLECTION |
| RTA i analiza możliwych obiektów | GRAPH_PRECISION |
| Taint analysis | DATA_FLOW |
| Obserwacja działającej aplikacji | RUNTIME_NATIVE |
| Rozpoznawanie samych podatności — pięć punktów | Istniejące CVE i CVE_IMPLEMENTATION_PLAN |
| Inne języki, w szczególności Go | ECOSYSTEMS |
| Pomiar skuteczności i kolejność realizacji | ROADMAP i kryteria poszczególnych faz |

Multi-release JAR-y i stabilny wybór klas są uzupełnieniem niezbędnym do wiarygodnego grafu. Nie są nową deklaracją, że oryginalny zakres został już zrealizowany.

## Jak korzystać z planów

1. Zacząć od ROADMAP i wspólnych kontraktów, aby nie budować odrębnych formatów dowodów, klas i metod dla każdego mechanizmu.
2. Dla obszaru przeczytać dokument funkcjonalny, następnie jego fazy i wskazane zależności.
3. Realizować pierwszy zakres danego mechanizmu, a rozszerzenia eksperymentalne dopuścić dopiero po pomiarze.
4. Po wdrożeniu aktualizować stan konkretnej fazy i zapisywać dowody weryfikacji. Samo utworzenie dokumentu nie oznacza ukończenia fazy.

Plany obejmują dokumentację wymagań i prac. Proponowane flagi, pakiety, stany i formaty nie są obecnie dostępnymi funkcjami. Nie podano terminów ani procentu gwarantowanego pokrycia.
