# Precyzja grafu — fazy P0–P4

Status: plan. Wymagania: [GRAPH_PRECISION.MD](GRAPH_PRECISION.MD). Wspólna kolejność: [ROADMAP.MD](ROADMAP.MD).

## P0. Uczciwy baseline i tryby porównania

Zależności: JVM G0–G4, B0–B1, CVE F1–F4 dla porównywalnych wyników.

- W `internal/callgraph` rozdzielić wybór strategii od internowania i przechowywania grafu. Zachować istniejący builder przez adapter do trybu legacy/CHA.
- Zapisać sygnatury, korzenie, target i konfigurację skip w metadanych analizy. Nie zmieniać tych parametrów między porównywanymi trybami bez ujawnienia różnicy.
- Przygotować przypadki, w których wiadomo, które cele są poprawne; skorygować błędy deskryptorów przed przypisaniem redukcji nowemu algorytmowi.
- Zdefiniować politykę nieznanej alokacji: np. konserwatywny fallback do uzasadnionego zbioru typów dla dotkniętego miejsca, a jeśli nie ma dostatecznej podstawy — niekompletność i ograniczenie wyniku negatywnego.

Gotowe: porównanie krawędzi ma stabilne identyfikatory i wyjaśnia zmiany, a nie tylko rozmiar grafu.

## P1. Katalog typów tworzonych jawnie i pośrednio

Zależności: DATA_FLOW D1–D2, rejestr B3.

- Rozpoznać alokacje `new`, tablice w odpowiednim zakresie oraz instancje dostarczane jako wejście scenariusza.
- Zbierać typy tworzone przez SERVICE S2, FRAMEWORK F2/F3, REFLECTION R2, DYNAMIC L1/L2 i DESERIALIZATION Z1, gdy te modele są aktywne.
- Brak modelu któregoś mechanizmu rejestrować jako potencjalną lukę; wdrożenie pełnego katalogu nie jest ukrytym wymogiem startu P1.
- Typy przekazywane do publicznych metod w trybie biblioteki nie są ograniczone wyłącznie do alokacji wewnątrz badanej biblioteki. Zadeklarować zewnętrzne wejścia/nieznane typy.

Gotowe: modelowana alokacja ma pochodzenie i może uruchomić rozwinięcie grafu tak samo jak jawne `new`.

## P2. RTA i punkt stały

Zależności: P1, B3; początkowo ograniczony katalog modeli z jawnymi lukami.

- Utworzyć kolejkę osiągalnych metod, odkrytych alokacji i nierozstrzygniętych wywołań. Nowy typ ponownie rozpatruje odpowiednie call sites.
- Wirtualny dispatch ograniczać dopiero po dokładnym dopasowaniu deskryptora i właściwej implementacji z JVM G2.
- Zapewnić niezależność od kolejności odkrycia typów, cykli i alokacji dopiero w metodzie odkrytej przez model.
- Stosować politykę unknown z P0; nie odrzucać typów tworzonych przez nieznaną fabrykę bez ostrzeżenia.
- Udostępnić opt-in RTA, raportować fallbacki i wpływ na werdykty. `roots=all` i korzenie profilowe porównywać jako oddzielne scenariusze.

Gotowe: fixtures z trzema implementacjami, pośrednią alokacją, cyklem i nieznaną fabryką zachowują wszystkie uzasadnione cele. P0–P2 to pierwsze wydanie precyzji grafu.

## P3. Ograniczona analiza points-to

Zależności: DATA_FLOW D3; P2 jest przydatnym baseline, nie koniecznym fundamentem każdej implementacji points-to.

- Rozszerzyć wspólną domenę obiektów, nie budować drugiego interpreter stosu. Klucz alokacji i przyjęta czułość kontekstu muszą być jawne.
- Modelować przepływ przez parametry, wyniki, pola, tablice i interfejsy w określonym zakresie. Nieznane zewnętrzne metody mają konserwatywną politykę.
- Dla konkretnego receivera wyznaczać możliwe cele i aktualizować graf/streszczenia do kontrolowanego punktu stałego.
- Rozdzielić obiekty, gdy istnieje wystarczający dowód. Brak aliasingu nie wynika z różnicy nazw zmiennych.
- Wprowadzić limity pamięci/czasu, poszerzanie i flagi niekompletności. Fallback nie może podawać ograniczonej listy jako pełnej.

Gotowe: dwa receivery, alias pól, tablica i fabryka dają uzasadnione różnice względem RTA. Metryki wskazują koszt oraz jakość, zanim zwiększy się czułość analizy.

## P4. Benchmark i kwalifikacja

- Osobno mierzyć błędne cele, pominięte oczekiwane cele, nieznane miejsca, zmiany CVE, czas i pamięć na fixtures i reprezentatywnych aplikacjach.
- Jeśli używany jest zewnętrzny silnik referencyjny, przypiąć wersję, zakres i konfigurację; jego wynik także nie jest automatycznie prawdą.
- Porównanie musi uwzględniać wersję modeli frameworka/refleksji i reguły JDK skip. Ulepszenie potwierdzać ręcznie opisanymi przypadkami, nie opinią LLM o krótszej ścieżce.
- Wykonać B4, snapshoty i dokumentację migracji. Zmiana domyślnej strategii jest osobną decyzją po pomiarze.

Nie podaje się z góry gwarantowanej redukcji fałszywych alarmów ani kosztu analizy całego classpath.
