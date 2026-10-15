# Możliwości rozszerzenia Ariadne

## Zakres i charakter rekomendacji

Ariadne można znacząco rozszerzyć. Większość obecnych ograniczeń da się zmniejszyć, choć pełne i precyzyjne pokrycie wszystkich zachowań JVM wymagałoby połączenia analizy statycznej z obserwacją działającej aplikacji. Nawet wtedy nie można zagwarantować wykrycia każdej możliwej ścieżki.

Rekomendacje powstały na podstawie przeglądu implementacji, przyszłych etapów w `PLAN.md` oraz dokumentacji wskazanej poniżej. Są oceną możliwości rozwoju; nie zostały jeszcze potwierdzone eksperymentami na tym projekcie. Główny zakres to JVM, z osobną uwagą dotyczącą innych języków.

## Rozszerzenia obsługi mechanizmów JVM

| Mechanizm | Jak rozszerzyć Ariadne | Wpływ i ograniczenia | Względna trudność |
|---|---|---|---|
| **Inicjalizacja klas — `<clinit>`** | Dodawać krawędzie wywołujące inicjalizatory przy odpowiednich instrukcjach JVM. | Wykrywanie ścieżek uruchamianych przez statyczne pola i bloki inicjalizacyjne. | Średnia |
| **Wywołania metod odziedziczonych** | Rozwiązywać właściciela implementacji podczas budowania grafu, z uwzględnieniem deskryptora metody. | Pozwala kontynuować analizę wewnątrz rzeczywistej implementacji zamiast kończyć na pustym węźle podklasy. | Średnia |
| **ServiceLoader i moduły JPMS** | Czytać `META-INF/services`, deklaracje `uses`/`provides` i modelować tworzenie oraz użycie dostawców usług. | Obejmuje wtyczki i implementacje wybierane przez mechanizm usług. | Średnia |
| **Refleksja** | Śledzić wartości nazw klas i metod, także przez zmienne, argumenty i dostępne pliki konfiguracyjne. | Rozpoznaje więcej przypadków niż same literały. Dowolne nazwy przychodzące z zewnątrz nadal pozostają niepewne. | Duża |
| **Proxy, AOP i dependency injection** | Dodawać modele wiązania interfejsów, obiektów, interceptorów i metod docelowych. | Rozszerza obecne wsparcie Spring Data na kolejne mechanizmy. Modele muszą odpowiadać frameworkowi i konfiguracji. | Duża |
| **Callbacki i kod asynchroniczny** | Modelować np. `Executor.submit`, `Thread.start`, listenery, timery i wybrane operatory reaktywne. | Odtwarza przejście od rejestracji zadania do jego wykonania. | Średnia–duża |
| **MethodHandle i pozostałe `invokedynamic`** | Śledzić tworzenie uchwytów, ich przekształcenia oraz znane funkcje bootstrap. | Poszerza wsparcie dynamicznego wywoływania metod i języków JVM. Nie ma jednej reguły obsługującej każdy bootstrap. | Duża |
| **Deserializacja** | Modelować callbacki deserializacji, dozwolone typy, filtry i wybrane łańcuchy gadgetów. | Pozwala szukać ścieżek ukrytych przed zwykłym grafem wywołań. Wynik zależy również od konfiguracji. | Bardzo duża |
| **Konfiguracja i wejścia frameworka** | Czytać np. `web.xml`, konfigurację beanów, endpointów i konsumentów wiadomości. | Obejmuje kod wywoływany przez serwer lub kontener, zanim sterowanie dotrze do aplikacji. | Duża |
| **Dynamiczne ładowanie klas i JNI** | Przyjmować dodatkowe artefakty, modele kodu natywnego oraz ślady z uruchomienia. | Uzupełnia informacje niedostępne w wejściowym zestawie JAR-ów; zakres będzie częściowy. | Bardzo duża |

### Podstawy grafu wymagające uzupełnienia

Dwa pierwsze punkty warto potraktować jako uzupełnienie podstaw grafu. Obecny ekstraktor wywołań (`internal/classfile/calls.go`, funkcja `ExtractCalls`) analizuje instrukcje `invoke*`, ale nie dodaje zależności inicjalizacji klas. JVM może uruchomić inicjalizator m.in. przy `new`, `getstatic`, `putstatic` i `invokestatic`. [Specyfikacja JVM, sekcja 5.5](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-5.html#jvms-5.5)

Obsługa dziedziczenia w dopasowaniu podatnej metody (`internal/reach/match.go`) pomaga znaleźć sam cel, ale nie zastępuje rozwiązania odziedziczonej implementacji podczas budowania grafu (`internal/callgraph/build.go`). To istotne, gdy dopiero ta implementacja wywołuje kolejną, podatną metodę.

### Doprecyzowanie obecnego planu

- **ServiceLoader wymaga więcej niż parsowania plików usług.** W modułach dostawcy mogą być deklarowani przez `provides` i udostępniani przez metodę `provider()`. Samo wykrycie dostawcy nie oznacza jeszcze jego wykonania. [Dokumentacja ServiceLoader](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/util/ServiceLoader.html)
- **Refleksję można analizować szerzej niż wyłącznie dla literałów.** Analiza przepływu wartości może odtworzyć część nazw przekazywanych przez zmienne i metody. Przykładowo WALA udostępnia modele refleksji wykorzystujące stałe tekstowe i informacje z rzutowań. [Opcje analizy WALA](https://wala.github.io/javadoc/com/ibm/wala/ipa/callgraph/AnalysisOptions.html)

## Większe możliwości analityczne

### Dokładniejszy graf: RTA i analiza możliwych obiektów

Obecne CHA dopuszcza wiele implementacji tego samego wywołania. RTA ogranicza kandydatów do tworzonych typów, a analiza obiektów może dodatkowo ustalać, jakie implementacje trafiają do konkretnej zmiennej. Powinno to ograniczać fałszywe alarmy, ale wymaga modeli obiektów tworzonych przez refleksję i frameworki — inaczej można usunąć prawdziwe ścieżki. Takie algorytmy istnieją m.in. w [SootUp](https://soot-oss.github.io/SootUp/develop/callgraphs/).

### Śledzenie danych kontrolowanych przez użytkownika — taint analysis

Pozwoliłoby badać przepływ:

```text
parametr HTTP → przekształcenia danych → argument podatnej metody
```

To dostarcza mocniejszego uzasadnienia niż sama osiągalność metody lub obecna ocena LLM. Potrzebne byłyby modele źródeł danych, propagacji, zabezpieczeń i podatnych operacji. Taki rodzaj analizy opisuje [dokumentacja CodeQL dla Java/Kotlin](https://codeql.github.com/docs/codeql-language-guides/analyzing-data-flow-in-java/). Nadal nie byłby to automatyczny dowód wykorzystania CVE.

### Opcjonalna obserwacja działającej aplikacji

Agent Java mógłby zbierać wykonane wywołania, załadowane klasy i informacje o dynamicznych implementacjach. Ariadne importowałaby te dane jako dodatkowe dowody. JVM udostępnia [mechanizm agentów i instrumentacji](https://docs.oracle.com/en/java/javase/25/docs/api/java.instrument/java/lang/instrument/package-summary.html).

Zaobserwowane wykonanie metody byłoby wartościowym potwierdzeniem. **Brak wykonania w zebranym scenariuszu nie oznaczałby nieosiągalności** — dany endpoint, konfiguracja lub zadanie mogły po prostu nie zostać uruchomione.

## Rozpoznawanie samych podatności

Większy graf nie pomoże, jeżeli mapowanie CVE wskazuje niewłaściwą metodę. Przydatne byłyby:

- mapowania zależne od wersji biblioteki i dokładnej sygnatury metody;
- analiza kodu przed poprawką, w tym metod usuniętych lub przemianowanych;
- rozpoznawanie bibliotek osadzonych w shaded JAR-ach;
- jawna świeżość danych CVE i możliwość odświeżania cache;
- analiza podatności JDK na podstawie faktycznego producenta i wydania runtime’u — wersja bytecode’u nie wystarczy do ustalenia poziomu poprawek.

## Proponowana kolejność prac i ocena efektów

1. Poprawność podstaw grafu.
2. ServiceLoader i callbacki.
3. Modele frameworków i refleksji.
4. Dokładniejszy graf.
5. Analiza danych.
6. Obserwacja runtime’u.

Każdy etap powinien mieć małe przykłady z oczekiwaną ścieżką oraz pomiar zarówno pominiętych podatności, jak i fałszywych alarmów. Bez tego większa liczba znalezionych ścieżek nie dowodzi poprawy jakości.

## Inne języki

Wsparcie innych języków również jest możliwe, ale wymagałoby osobnych warstw analizy. Dla Go warto najpierw ocenić integrację z istniejącym `govulncheck`, który obsługuje analizę źródeł i plików binarnych. [Dokumentacja Go](https://go.dev/doc/security/vuln/)

## Granice oszacowania

Nie określono procentu możliwego pokrycia ani czasu realizacji: projekt nie ma jeszcze zestawu pomiarowego obejmującego te mechanizmy. Trudność podana w tabeli jest oceną względną, a nie oszacowaniem nakładu pracy.
