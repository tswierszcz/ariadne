# Kontrakt dokumentacji wspieranych funkcji Ariadne

Wszystkie ścieżki w tym dokumencie są względem root repozytorium Ariadne. Czytaj ten plik przy implementacji zmieniającej zachowanie lub zakres wsparcia.

## 1. Dokumenty pełnią różne role

| Dokument | Rola i oczekiwana aktualizacja |
|---|---|
| `docs/SUPPORT.MD` | Utrzymywany opis aktualnych możliwości, ich granic, warunków i dowodów. Zmieniaj odpowiednie wiersze po rzeczywistej zmianie zachowania. |
| `README.md` | Publiczny opis działania i użycia. Aktualizuj pipeline, flagi, przykłady, wymagania i ograniczenia w zakresie zmiany. |
| `internal/report/templates/report.md.tmpl` | Twierdzenia trafiające do wygenerowanego raportu. Aktualizuj, gdy nowy model zmienia ich prawdziwość; sprawdź powiązane testy renderowania. |
| `FAQ.md`, `FAQ_PL.MD` | Wyjaśnienia i przykłady wsparcia. Popraw oba warianty językowe, jeśli zawierają twierdzenie zmienione przez daną fazę. Nie przepisuj niepowiązanych rozdziałów. |
| `RECOMMENDATION/<OBSZAR>.MD` | Cel i wymagania. Dodaj odnośnik do zakresu dostarczonego; reszta pozostaje planem, dopóki nie została wykonana. |
| `RECOMMENDATION/<OBSZAR>_IMPLEMENTATION_PLAN.MD` | Kryteria i stan konkretnych faz z dowodami. Nie usuwaj niewykonanych wymagań, aby pozornie ukończyć fazę. |
| `RECOMMENDATION/IMPLEMENTATION_STATUS.MD` | Aktualny punkt kontynuacji między sesjami i rejestr faktycznie wykonanych kontroli. |
| `RECOMMENDATION/README.MD`, `ROADMAP.MD` | Nawigacja, zależności i kolejność. Aktualizuj, jeśli rzeczywista realizacja wymaga korekty tych informacji. |
| `docs/phase4-llm.md` i inne dokumenty szczegółowe | Aktualizuj przy zmianie danego kontraktu, evidence, promptu lub sposobu interpretacji wyniku. |
| `PLAN.md` | Historia i wcześniejsze założenia. Dopisuj wynik wdrożenia, jeśli potrzebne; zachowaj istniejące wpisy i zmiany użytkownika. Nie traktuj starego planu jako aktualnej macierzy wsparcia. |

Zachowaj język istniejącego dokumentu: README i FAQ angielskie, FAQ_PL polski; macierz SUPPORT jest angielska. Nowy wpis ma być czytelny dla użytkownika produktu, nie tylko autora kodu.

## 2. Minimalny opis nowej możliwości

Dla każdej dodanej lub zmienionej funkcji odpowiedz w dokumentacji:

1. Co użytkownik może teraz przeanalizować? Wskaż konkretny mechanizm i obsługiwany podzbiór.
2. Dla jakich wejść, wersji Javy/frameworka/biblioteki i kontekstów to sprawdzono? Rozróżnij zakres modelu od wersji potwierdzonych przykładami.
3. Jak to włączyć? Używaj wyłącznie wdrożonych flag i poprawnych przykładów; podaj domyślne zachowanie oraz wymagane zasoby.
4. Jakiego rodzaju dowód powstaje: statyczna możliwość, heurystyka, przepływ danych czy obserwacja wykonania?
5. Co pozostaje nieobsługiwane lub nierozstrzygnięte i jak to wpływa na status, luki pokrycia i wynik negatywny?
6. Jak sprawdzono zachowanie? Wskaż test/scenariusz i wynik rzeczywiście wykonanej weryfikacji.

Przykład właściwej deklaracji: „Classpath service declarations with a known service class and the implemented iterator API are modeled; custom module layers remain unsupported.” Nie pisz „pełna obsługa ServiceLoader” po wdrożeniu tylko tego zakresu.

## 3. Status wsparcia i podstawa weryfikacji

Używaj oddzielnych pól:

- `Implemented`: dostępny opisany zakres; nie oznacza wszystkich wariantów mechanizmu.
- `Partial`: działa konkretny podzbiór, a luki są wymienione.
- `Experimental`: dostępne, lecz nadal oceniane; wskaż warunki i opt-in.
- `Not implemented`: tego zachowania brak w aktualnej ścieżce wykonania.
- `Unassessed`: brak dostatecznej podstawy do ustalenia aktualnego zakresu.

Obok podaj podstawę: przegląd kodu, istniejące lecz nieuruchomione fixtures, wykonane testy, kontrolowane uruchomienie lub udokumentowane wyniki historyczne. Wpis w kodzie nie dowodzi przejścia testów, a istniejący raport historyczny nie jest nowym live run.

Stan fazy w rejestrze jest innym wymiarem: możliwe jest `DONE` dla fazy projektu kontraktu przy nadal `Not implemented` dla funkcji użytkowej. Częściowe wsparcie całego obszaru nie oznacza nieukończenia poprawnie ograniczonej fazy MVP.

## 4. Sprawdzenie spójności

- Sprawdź, czy README lub raport nie mówi „niewidoczne”, gdy wdrożono już część modelu. Zachowaj ograniczenie dla pozostałych przypadków zamiast usuwać je w całości.
- Sprawdź zgodność pomocy CLI, wartości domyślnych i przykładów z kodem. Proponowane w planach flagi nie trafiają do opisu dostępnego CLI przed implementacją.
- Rozróżnij wybrany target Java od wydania i producenta JDK; obecność klasy od jej wykonania; entry point od podatnej metody; LLM od zweryfikowanego mapowania.
- Nie nazywaj `NOT_REACHABLE` dowodem bezpieczeństwa ani braku śladu runtime dowodem nieosiągalności.
- Nie zapewniaj pełnej deterministyczności/aktualności/offline, jeśli nie zweryfikowano wszystkich wejść i zależności danego trybu.
- Nie utrzymuj dwóch sprzecznych opisów tego samego zachowania. Jeśli usunięcie sprzeczności wykracza poza bieżącą fazę, zapisz ją jako jawną lukę i ogranicz deklarację.

## 5. Dowód zamknięcia

W rejestrze fazy zapisz nazwy i wyniki wykonanych poleceń, zakres sprawdzonych scenariuszy, ograniczenia, stan dokumentacji i kolejny krok. Polecenie niewykonane wpisz wprost jako niewykonane wraz z powodem; nie oznaczaj go „PASS”.

W podsumowaniu dla użytkownika najpierw podaj nowy zakres wsparcia i jego warunki. Pliki implementacji i szczegóły refaktoru są dowodami, nie zastępują tego opisu.
