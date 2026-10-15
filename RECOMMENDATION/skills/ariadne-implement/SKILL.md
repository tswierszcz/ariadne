---
name: ariadne-implement
description: "Przedstaw do zatwierdzenia plan kolejnej fazy Ariadne z RECOMMENDATION, a po zgodzie użytkownika zrealizuj implementację, weryfikację i aktualizację opisu wsparcia. Używaj przy zleceniu wykonania lub wznowienia faz Ariadne; sam przegląd planów nie uruchamia implementacji."
---

# Ariadne — realizacja faz

## Wynik jednej sesji

Dostarcz jeden spójny przyrost z planów Ariadne: działający kod, adekwatną weryfikację, aktualny opis wsparcia i stan pozwalający wznowić pracę. Szersze polecenie użytkownika określa szerszy zakres; sam skill nie oznacza zgody na wykonanie całego backlogu.

Plany są wymaganiami i hipotezami projektowymi. Obecne zachowanie ustalaj z kodu i dowodów. Nie oznaczaj funkcji jako wspieranej tylko dlatego, że ma plan, wpis profilu albo interfejs.

## Obowiązkowe zatwierdzenie planu przed realizacją

Przy każdym nowym wywołaniu tego workflow lub wznowieniu w nowej sesji najpierw ponownie przedstaw plan. Użytkownik chce mieć możliwość zmiany kolejności i priorytetów przed rozpoczęciem pracy.

Przed zgodą wolno wykonać przegląd tylko do odczytu: ustalić stan repozytorium, przejrzeć dokumentację, kod i dowody oraz ocenić zależności. Nie zmieniaj plików, nie rozpoczynaj implementacji, nie uruchamiaj testów ani operacji instalacyjnych w ramach proponowanego przyrostu. Przygotowanie kontraktów i aktualizowanie rejestru faz również należą do realizacji wymagającej zgody.

Przedstaw użytkownikowi:

- proponowane fazy i kolejność, z pełnymi identyfikatorami;
- uzasadnienie priorytetu i konieczne zależności;
- konkretny zakres bieżącego przyrostu i oczekiwany rezultat;
- sposób weryfikacji oraz dokumenty wsparcia do aktualizacji;
- istotne decyzje, które użytkownik może zmienić przed startem.

**Po przedstawieniu planu zaczekaj na wyraźną zgodę użytkownika.** Poproś o zatwierdzenie lub zmianę kolejności/priorytetów i oddaj turę; nie kontynuuj realizacji w tle. Wyjaśnij krótko, że oczekiwanie wynika z tej instrukcji skilla ustawionej na życzenie użytkownika, podając odnośnik do aktualnego `SKILL.md` i przytoczoną powyżej zasadę.

Za zgodę uznaj odpowiedź odnoszącą się do przedstawionego planu, np. „zatwierdzam”, „start” lub „realizuj według tego planu”. Samo wywołanie skilla, polecenie kontynuacji, historyczna zgoda z innej sesji, brak odpowiedzi albo upływ czasu nie zastępują tej zgody. Gdy użytkownik proponuje zmiany, przedstaw zaktualizowany plan i zaczekaj na jego zatwierdzenie, chyba że w tej samej odpowiedzi jednoznacznie zatwierdził już konkretny zmieniony zakres.

Po uzyskaniu zgody wykonuj zaakceptowany zakres autonomicznie. Nie pytaj ponownie o rutynowe kroki mieszczące się w tym zakresie. Istotne rozszerzenie zakresu lub zmiana kolejności/priorytetów wymaga przedstawienia korekty i ponownego zatwierdzenia; istniejąca zgoda na konkretny zestaw faz obejmuje przejścia między nimi zgodnie z zaakceptowaną kolejnością.

## Wejście w repozytorium i wybór zakresu

`INSTRUCTION FOR USER.MD` jest poradnikiem wyłącznie dla człowieka. Pomijaj go przy starcie: nie odczytuj, nie analizuj ani nie streszczaj automatycznie, również przy zbiorczym wczytywaniu dokumentacji. Odczytaj go tylko na wyraźną prośbę użytkownika dotyczącą tego poradnika. Źródłami workflow są ten skill i wskazane poniżej dokumenty operacyjne.

1. Ustal root aktualnego checkoutu. Potwierdź projekt przez `go.mod`, `cmd/ariadne` i `RECOMMENDATION/README.MD`; nie używaj na stałe ścieżki z poprzedniej sesji. Przeczytaj obowiązujące instrukcje repozytorium i sprawdź stan zmian.
2. Odczytaj `RECOMMENDATION/README.MD`, `RECOMMENDATION/ROADMAP.MD`, `RECOMMENDATION/IMPLEMENTATION_STATUS.MD` i `docs/SUPPORT.MD`. Następnie wczytaj tylko właściwy plan funkcjonalny, plan faz i rzeczywiste zależności wybranego przyrostu.
3. Identyfikuj fazę jako `OBSZAR:FAZA`, np. `JVM_GRAPH:G2`, `CVE:F4`, `FRAMEWORKS:F2`; same `F2` lub `R0` są niejednoznaczne. Przy niejasnym poleceniu sprawdź kontekst; zapytaj tylko wtedy, gdy wybór nadal zmieniałby zakres pracy.
4. Gdy nie wskazano fazy, zaproponuj wznowienie rozpoczętej pracy, jeśli jej dalsza realizacja jest możliwa. W przeciwnym razie zaproponuj najwcześniejszy gotowy przyrost zgodny z roadmapą. Sprawdź kod i kryteria zależności — etykieta w rejestrze nie jest samodzielnym dowodem gotowości.
5. W pierwszej sesji zaproponuj fundamenty `ROADMAP:R0`: konieczne części B0–B1 i wykonalną fazę. Wspólne zadania JVM/CVE realizuj po zgodzie jeden raz. Minimalny kontrakt potrzebny do kodu nie oznacza ukończenia całej powiązanej fazy CVE.
6. Konieczne brakujące warunki uwzględnij w planie do zatwierdzenia. Po zgodzie realizuj je w ramach zaakceptowanego przyrostu i zapisz zakres. Przy sprzecznych wymaganiach lub koniecznej decyzji użytkownika podaj konkretne możliwości i pytanie; kontynuuj tylko niezależną część już zatwierdzonej pracy. Nie wybieraj arbitralnie dostawcy JDK lub niekompatybilnego formatu wyników.

Najpierw uzyskaj zatwierdzenie przedstawionego planu zgodnie z sekcją powyżej. Następnie zlecenie realizacji prowadzi do wykonania zaakceptowanej pracy, nie tylko kolejnego planu. Faza z definicji kontraktowa może zakończyć się konkretnym kontraktem i scenariuszami zamiast kodu; nie przedstawiaj jej jako działającej funkcji.

## Implementacja i dowody

- Korzystaj ze wspólnych kontraktów B0–B4 i właścicieli prac w ROADMAP. Proponowane pakiety/flagę wdrażaj tylko wtedy, gdy służą wybranemu zakresowi; nie twórz rusztowania pozostałych obszarów bez potrzeby.
- Zachowaj różnice: brak metody vs brak ścieżki, deklaracja/rejestracja vs wykonanie, model statyczny vs obserwacja, dokładna sygnatura vs dowód podatności. Nie zamieniaj braków modelu lub przerwanej analizy w dowód bezpieczeństwa.
- Przy zmianie zachowania przygotuj istotne przypadki dodatnie, ujemne i nierozstrzygnięte odpowiednie do fazy. Wybierz testy sprawdzające semantykę, a nie jedynie nazwy pól czy zgodność testu z implementacją.
- Uruchom właściwe testy zmienionych pakietów. Przed oznaczeniem implementacyjnej fazy jako ukończonej wykonaj wymagane w jej planie kontrole integracyjne oraz zwykle `go test ./...` i `go vet ./...`. Jeśli środowisko czegoś nie pozwala wykonać, zapisz dokładny brak weryfikacji i pozostaw adekwatny stan fazy/wsparcia.
- Analiza statyczna ma zachować pracę na gotowych artefaktach bez obowiązkowego JDK. Regeneracja Java fixtures i integracje opcjonalnego agenta mają osobne wymagania; zachowaj źródła i powtarzalny sposób przygotowania artefaktów.
- Przy porównaniu CVE lub grafu zachowaj wejście i dane źródłowe/snapshot. Nie interpretuj spadku liczby alarmów jako samodzielnego dowodu poprawy. Nie uruchamiaj płatnych wywołań LLM jako ubocznego efektu zwykłych testów.

## Dokumentacja jest częścią kryterium ukończenia

Przy każdej zmianie zakresu wsparcia wczytaj [kontrakt dokumentacji](references/documentation-contract.md) i wykonaj go w tym samym przyroście. Obejmuje to `docs/SUPPORT.MD`, odpowiednie sekcje README i — gdy zmieniają się twierdzenia raportu — jego szablon oraz testy. Sprawdź też odpowiednie FAQ i dokumenty danego mechanizmu.

Opisuj obsługiwane API/sygnatury, wersje, wejścia, flagi, domyślne zachowanie, warunki, brakujące przypadki i wpływ na wynik. Ukończenie części ServiceLoadera nie oznacza pełnego wsparcia ServiceLoader/JPMS. Jeśli faza zmienia tylko fundament, wpisz poprawę fundamentu, nie funkcję, której użytkownik nadal nie może użyć.

## Zakończenie i wznowienie

- Zaktualizuj wpis konkretnej fazy w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`: zakres wykonany, pozostałe kryteria, zmienione pliki, rzeczywiście uruchomione kontrole i wyniki, ograniczenia oraz dokładny następny krok. Oznacz w odpowiednim planie tylko rzeczywiście ukończone fazy, zachowując jego kryteria.
- `DONE` wymaga spełnienia kryteriów, adekwatnej weryfikacji i zgodnej dokumentacji. Brakujące kontrole oznacz jako `NEEDS_VERIFICATION`; częściową implementację jako `IN_PROGRESS`; rzeczywistą zależność zewnętrzną jako `BLOCKED` z powodem. To stany pliku, nie polecenia zarządzania celami aplikacji.
- Przy końcu sesji zapisz użyteczne przekazanie także dla pracy nieukończonej. Następna sesja ma móc kontynuować z repozytorium bez dostępu do historii czatu.
- Podsumuj użytkownikowi: co zrealizowano, co jest teraz wspierane i na jakich warunkach, co sprawdzono, co pozostaje poza zakresem oraz jaka faza jest następna. Podaj odnośniki do zmienionej dokumentacji.
