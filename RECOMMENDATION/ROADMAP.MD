# Wspólna kolejność realizacji i kontrakty analiz

Status: propozycja architektury i kolejności prac. Indeks: [README.MD](README.MD).

## 1. Cel

Rozszerzać pokrycie mechanizmów JVM, zachowując pochodzenie każdego ustalenia, precyzję tożsamości komponentu i jawną informację o brakach. Poszczególne mechanizmy korzystają ze wspólnego grafu i wspólnej analizy wartości; nie tworzą jedenastu niezależnych silników.

Tor [CVE](CVE_IMPLEMENTATION_PLAN.MD) odpowiada za to, co jest podatne. Pozostałe tory opisują, jak kod może zostać wywołany, jakie dane mogą do niego dotrzeć i co rzeczywiście zaobserwowano.

## 2. Wspólne kontrakty — B0–B4

### B0. Wyniki, dowody i pokrycie

- Wykorzystać i uzgodnić model z CVE F0–F1. Jeden schemat wersjonowanego wyniku, instancji komponentu i dowodu.
- Oddzielić: tożsamość artefaktu, jakość mapowania CVE, możliwość wywołania, pochodzenie danych i obserwację wykonania.
- Rejestrować kod powodu, miejsce w bytecode/konfiguracji, wersję modelu, warunki, źródła i granice zastosowania.
- Ustaleniom nadawać charakter: wynik reguły JVM, model biblioteki, heurystyka, deklaracja wejściowa albo obserwacja. To nie jest wspólny procent pewności.
- Ustalić wpływ luk na wynik. Nieznany cel w obszarze relewantnym dla ocenianego CVE może uniemożliwić negatywne rozstrzygnięcie. Luki globalne i istotne dla konkretnego wyniku raportować osobno; nie zakładać ich nieistotności bez uzasadnienia.
- Wszystkie nowe stany, flagi i pakiety są propozycjami. `NOT_REACHABLE` oznacza brak ścieżki w zadeklarowanym modelu; `NOT_APPLICABLE` zachowuje dotychczasowe znaczenie WebJar-ów.

Gotowe, gdy ta sama luka nie jest różnie interpretowana przez dwa analizatory, a eksport legacy nie ukrywa niepewności.

### B1. Identyfikatory i dowody krawędzi

- Tożsamość metody: pochodzenie klasy/wybrany kontekst ładowania, nazwa klasy, nazwa metody, deskryptor.
- Miejsce wywołania: stabilna metoda + offset bytecode + rodzaj operacji. Numery linii służą prezentacji, nie identyfikacji.
- Rozdzielić graf wywołań od relacji takich jak rejestracja callbacka, utworzenie obiektu, inicjalizacja czy przepływ danych.
- W grafie wywołań zachować rodzaj/podstawę krawędzi i jej warunki. Wirtualny węzeł zdarzenia jest jawnie oznaczony; nie udaje metody znalezionej w bytecode.
- Zachować związek kroku ścieżki z krawędzią. Obecna lista nazw metod nie wystarcza do wyjaśnienia przejścia przez proxy lub wątek.
- Współdzielić deterministyczny wybór klas i instancji z JVM G1/CVE F1. Pełne izolowane classloadery należą do RUNTIME N1–N2, nie do niejawnego rozszerzenia jednej mapy FQCN.

Gotowe, gdy dwa powody utworzenia tej samej krawędzi są zachowane i można odtworzyć je z raportowanej ścieżki.

### B2. Wspólna reprezentacja kodu i wartości

- Implementację wykonują DATA_FLOW D0–D2: instrukcje, CFG, obsługa wyjątków, stos, zmienne lokalne, identyfikatory obiektów oraz wartości abstrakcyjne.
- Obecny `parseCodeAttribute` pomija tabelę wyjątków; przyszła analiza musi ją zachować. Uzupełnienie wykonuje DATA_FLOW D1 wspólnie z parserem, bez równoległego drugiego formatu.
- Domena wartości obejmuje m.in. stałe tekstowe i klasowe, skończone zbiory typów/celów, alokacje, `null` i wartość nieznaną.
- Każdy limit zbioru, głębokości lub czasu powoduje poszerzenie do nieznanej wartości z diagnostyką. Nie wolno po prostu odrzucać nadmiarowych kandydatów.
- Interproceduralne streszczenia wprowadzane są przyrostowo. Analiza refleksji nie wymaga ukończenia taint analysis D4.

Gotowe, gdy dwa konsumenty, np. refleksja i callbacki, mogą użyć tych samych faktów o argumencie bez ponownego dekodowania semantyki stosu.

### B3. Modele bibliotek i iteracja analizy

- Wprowadzić rejestr modeli z identyfikatorem, obsługiwanymi sygnaturami, zakresem wersji i źródłem semantyki. Nowy profil może dostarczać dane, nie wykonywalny kod z YAML.
- Obsługę znanego API przypinać do konkretnego wywołania/obiektu, nie do wszystkich metod o tej samej nazwie w całej aplikacji.
- Graf, odkryte alokacje i streszczenia wartości wpływają na siebie. Używać kolejki pracy i kontrolowanego punktu stałego; pojedyncze uruchomienie modelu przed BFS nie wystarczy.
- Wprowadzić budżety i stan `incomplete` z przyczyną. RTA i points-to mają własne reguły aktualizacji opisane w GRAPH_PRECISION.
- Przy poprawianiu modelu unieważniać zależne wyniki, wykorzystując zasady CVE F2–F3. Kolejność uruchomienia modeli nie może losowo zmieniać wyniku.

Proponowany podział implementacyjny: `internal/analysis` koordynuje iterację, a `internal/models` zawiera kontrakt i rejestr modeli. Parser, hierarchia i IR dostarczają dane wejściowe; model zwraca fakty o celach wywołań, alokacjach, wejściach i lukach. Koordynator aktualizuje graf oraz kolejkę. `internal/profile` pozostaje dostawcą konfiguracji i wyboru modeli, nie drugim koordynatorem. Kierunek zależności pakietów trzeba ustalić przed refaktorem, aby uniknąć cyklu profile → callgraph → models → profile.

Gotowe, gdy model usługi może wprowadzić typ, który uruchamia kolejne poprawne rozwinięcie grafu, a limit zatrzymania jest widoczny.

### B4. Weryfikacja i kwalifikacja przyrostów

- Każdy obszar ma małe przypadki dodatnie, ujemne i nierozstrzygnięte. W miarę możliwości założenia modelu potwierdzać oddzielnym kontrolowanym uruchomieniem JVM; brak obserwacji nie dowodzi nieosiągalności.
- Stosować źródła i gotowe JAR-y fixtures, aby zwykłe testy silnika Go nie wymagały JDK ani sieci. Regeneracja fixtures i integracje agenta wymagają jawnie przygotowanego JDK.
- Porównywać grafy, przyczyny i werdykty na tej samej konfiguracji, snapshotcie oraz wejściu; wynik starego programu nie jest sam w sobie wzorcem poprawności.
- Mierzyć pominięte oczekiwane cele, błędne krawędzie, liczbę nierozstrzygniętych miejsc, czas i pamięć. Mniej alarmów lub większy graf nie są samodzielnym kryterium jakości.
- Przy implementacji uruchamiać właściwe testy pakietów, a przed wydaniem `go test ./...` i `go vet ./...`. Płatne LLM i testy live są osobne.
- Dokumentacja ma wskazywać dokładnie obsługiwane wersje i warunki, bez deklaracji pełnego wsparcia JVM.

## 3. Kolejność i pierwsze wydania

| Przyrost | Zakres | Warunek wejścia | Efekt |
|---|---|---|---|
| R0 | B0–B1, JVM G0–G4; wspólne części CVE F0–F1/F4 | Uzgodnione pochodzenie klas i wyników. | Stabilny graf, dziedziczenie, inicjalizacja i wybór właściwych klas. |
| R1 | DATA_FLOW D0–D2; SERVICE S0–S2; ASYNC A0–A2 | R0 i lokalne wartości. | Pierwsze usługi i callbacki z dowodami przy miejscu wywołania. |
| R2 | REFLECTION R0–R2; DYNAMIC L0–L2; FRAMEWORK F0–F2 | B2–B3 oraz modele wejść. | Podstawowa refleksja, uchwyty i ograniczony model frameworka. |
| R3 | SERVICE S3, FRAMEWORK F3–F4; DATA_FLOW D3; GRAPH_PRECISION P0–P2 | Dostępne modele typów tworzonych pośrednio. | JPMS, więcej wejść, streszczenia między metodami, eksperymentalne RTA. |
| R4 | DATA_FLOW D4–D5; DESERIALIZATION Z0–Z3 | Zweryfikowane źródła/sinki i przepływ wartości. | Ścieżki danych i modele wejść deserializacji. |
| R5 | RUNTIME N0–N4 | B0–B1 i CVE F3 dla powtarzalnych dowodów. | Import śladów, opcjonalny agent, jawne granice native. |
| R6 | GRAPH_PRECISION P3–P4; dalsze modele pozostałych torów | Wyniki pomiarów wcześniejszych wydań. | Analiza możliwych obiektów i szersze modele, jeśli uzasadnia je benchmark. |
| Osobny pilot | ECOSYSTEMS E0–E4 | B0 i uzgodniony wynik v2. | Import/integracja Go bez przepisywania całego silnika JVM. |

To kolejność dostarczania, nie szacowanie czasu. RUNTIME N0–N1 (kontrakt i import) można wdrożyć wcześniej do pomiarów modeli statycznych; agent nie jest wymagany do pierwszego wydania.

CVE F2–F5 powinny być dostarczane jako osobny przyrost po wspólnej podstawie i najpóźniej przed deklarowaniem poprawy jakości rozpoznawania CVE. Shading CVE F6–F7 i runtime CVE F8 pozostają oddzielnymi etapami tego istniejącego planu.

Wzmianka o wspólnych częściach CVE F4 w R0 dotyczy dokładnego rozwiązywania metod i grafu, a nie ukończenia całej fazy mapowań przed jej zależnościami F2–F3. Zadania wspólne otrzymują jedną implementację, a pełne fazy CVE są kwalifikowane według własnego planu.

## 4. Właściciele prac wspólnych — bez duplikacji

| Praca | Główny plan | Współpraca |
|---|---|---|
| Wynik v2, komponenty, dowody CVE | CVE F0–F1 | Wszystkie tory korzystają z kontraktu B0. |
| Determinizm klas i rozwiązywanie metod | JVM G1–G2 | Realizuje odpowiednie zadania CVE F1/F4 jeden raz. |
| Inicjalizacja i multi-release JAR-y | JVM G3–G4 | Runtime CVE F8 może dostarczyć target; sama wersja językowa nie ustala podatności JDK. |
| CFG, stos, wartości, streszczenia | DATA_FLOW D0–D3 | Refleksja, callbacki, uchwyty, RTA i frameworki korzystają z tego samego IR. |
| Metadane i scenariusze konfiguracji | FRAMEWORK F1–F2 | Usługi, refleksja i deserializacja otrzymują dane z pochodzeniem. |
| Typy tworzone przez modele | Rejestr B3 | SERVICE, FRAMEWORK, REFLECTION i DESERIALIZATION przekazują je do GRAPH_PRECISION. |
| Cache, snapshoty, wersje modeli | CVE F2–F3 | Każdy tor dopisuje swoje wejścia i wersje. |
| Ślady wykonania i loader identity | RUNTIME N0–N4 | Pozostałe tory mogą importować dowody; brak śladu nie usuwa statycznej możliwości. |

Żaden plan nie wymaga ukończenia pełnego points-to przed pierwszym modelem usług lub refleksji. MVP korzystają z ograniczonych wartości lokalnych, jawnie oznaczając nierozstrzygnięte przypadki. Taint D4 jest konsumentem fundamentu D0–D3, a nie jego warunkiem.

## 5. Spójność wyników i kompatybilność

- Tryb `roots=all` pozostaje analizą od wszystkich metod aplikacji; nie dowodzi istnienia zewnętrznego źródła atakującego. Źródła HTTP i przepływ danych są osobną informacją.
- Rejestracja zadania, dostawcy lub listenera nie oznacza, że został wykonany. Graf możliwości może zawierać warunkową krawędź z wyjaśnieniem.
- Dotychczasowy podział dwóch BFS i etykieta `FRAMEWORK` pozostają punktem odniesienia. Wariant z zewnętrznymi korzeniami i modelami frameworka wymaga jawnej konfiguracji i opisanej agregacji, zamiast cichego przemianowania wszystkich wyników.
- Wyniki runtime’u są przypięte do przebiegu, hasha klasy i kontekstu loadera. Podobna nazwa metody nie wystarcza do połączenia dowodów.
- Obserwacja wywołania, znaleziony przepływ danych i potwierdzenie wykorzystania podatności to różne rodzaje ustaleń.
- Statystyki grafu rozdzielają zwykłe wywołania, rozwinięcia dispatch i relacje modelowane/inicjalizacyjne. Nie dopisują wszystkich nowych krawędzi do licznika CHA; suma i zasady deduplikacji są opisane w formacie raportu.
- Przy limicie analizy lub nieznanym modelu zachować pozytywne dowody i ograniczyć moc wyniku negatywnego. Nowe flagi są opt-in do czasu pomiarów i opisania migracji.

## 6. Decyzje i granice

Pierwszym frameworkiem pilotażowym jest Spring, bo ma już profil i aplikacje kontrolne. Poszczególne wersje muszą mieć własne fixtures; bieżąca dokumentacja Spring nie dowodzi zachowania starych wersji użytych w repozytorium.

Pierwszy silnik przepływu wartości jest ograniczony i napisany w Go, zgodnie z obecną ścieżką uruchomienia. SootUp/WALA/CodeQL mogą służyć jako referencja lub osobny eksperyment porównawczy; integracja któregoś z nich wymaga oddzielnej oceny licencji, zależności, wersji i możliwości pracy offline.

Agent Java jest osobnym opcjonalnym artefaktem. Nie zmienia warunku, że analiza statyczna istniejących JAR-ów działa bez JDK. Kod natywny, arbitrary custom classloaders i pełne badanie exploitability pozostają zakresem częściowym, z lukami w raporcie.

Nie zmienia się istniejącego `PLAN.md` w celu oznaczenia tych prac jako wykonanych. Po rzeczywistym ukończeniu fazy zapisuje się wersję, przykłady, polecenia weryfikacji i znane ograniczenia.
