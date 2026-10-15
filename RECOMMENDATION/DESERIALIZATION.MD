# Deserializacja — rekomendacja funkcjonalna

Status: plan. Fazy: [DESERIALIZATION_IMPLEMENTATION_PLAN.MD](DESERIALIZATION_IMPLEMENTATION_PLAN.MD). Zasady: [ROADMAP.MD](ROADMAP.MD).

## Cel i stan obecny

Rozpoznawać wybrane wejścia deserializacji, wykonywane przez nią callbacki i warunki powstawania kandydatów niebezpiecznych łańcuchów.

Obecna ręczna tabela `internal/vulnmap/vuln_methods.yaml` wiąże wiele CVE XStream ze wspólnymi punktami `fromXML`/`unmarshal`. To dowód użycia powierzchni ataku w modelu, nie odtworzenie konkretnego łańcucha gadgetów ani jego warunków. Role entry point i vulnerable method muszą pozostać odrębne zgodnie z planem CVE.

## Funkcje i poziomy ustalenia

| Poziom | Co ustala analiza |
|---|---|
| Wejście | Kod może wywołać wskazaną operację deserializacji danego komponentu i wersji. |
| Model obiektów | Obsługiwany deserializer może tworzyć wskazane typy i wywoływać modelowane callbacki. |
| Konfiguracja | Znane filtry i ograniczenia typów dopuszczają, odrzucają lub nie rozstrzygają wybranego kandydata. |
| Dane | Istnieje modelowana ścieżka danych z zewnętrznego źródła do wejścia deserializacji. |
| Łańcuch | Istnieje kandydat łańcucha z jawnymi założeniami; potwierdzenie exploitability wymaga dodatkowych dowodów. |

Pierwszy zakres obejmuje natywną serializację Java i jej wybrane callbacki, np. `readObject`/`readResolve`. Następny obejmuje jeden model biblioteki, np. wersje XStream użyte w aplikacji kontrolnej. Nie istnieje jeden wspólny model wszystkich deserializerów.

## Scenariusze

- `ObjectInputStream.readObject` może prowadzić do prywatnego callbacka klasy, mimo braku zwykłego wywołania w bytecode aplikacji.
- Filtr odrzucający konkretny typ może wykluczyć dany kandydat łańcucha w danym scenariuszu; nie usuwa automatycznie wszystkich CVE biblioteki.
- Nieznany lub instalowany dynamicznie filtr pozostawia niepewność. Brak wykrytego filtra nie oznacza, że środowisko nie ma ograniczeń.
- Dwa obiekty deserializatora z inną konfiguracją nie dzielą automatycznie jednego werdyktu.

## Dane, raport i granice

Raport podaje deserializer i wersję, punkt wejścia, typy/kandydatów, callbacki, konfigurację, dowody i brakujące przesłanki. Łańcuch może zależeć od pól obiektów, kształtu grafu, ustawień oraz kolejności; same zgodne sygnatury metod nie wystarczą do potwierdzenia.

Pełna analiza dowolnego grafu obiektów, wszystkich gadget chains i niestandardowych deserializerów nie jest obietnicą tego planu. Ograniczenia typów należy modelować według konkretnego API; konfiguracja jednej biblioteki nie jest filtrem innej.

Gotowość MVP: punkt wejścia i callbacki są rozpoznawane, filtr/nieznany filtr rozróżnione, a raport nie nazywa samego `fromXML` potwierdzeniem exploitability konkretnego CVE.

## Źródła

- [Serializacja Java: reguły odtwarzania i callbacki](https://docs.oracle.com/en/java/javase/25/docs/specs/serialization/input.html).
- [ObjectInputFilter](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/io/ObjectInputFilter.html).

Wersje i zachowanie bibliotek zewnętrznych wymagają oddzielnych źródeł/fixtures przy wdrażaniu ich modelu.
