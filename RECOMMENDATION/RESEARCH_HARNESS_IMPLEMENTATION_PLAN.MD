# Wsparcie procesu Recon → Hunt → Validate w Ariadne

Data 2026-11-25. Stan PLANOWANE. Plan rozszerza [RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD](RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD) o potrzeby procesu badawczego Tekmerion. Dotychczasowe RI0–RI6 zachowują swoje znaczenie.

## 1. Punkt wyjścia i źródła

Podstawa kodowa przeglądu to `3c999a8` (RI4), scalona do `main` przez `0c6f4f6`. Działają indeksowanie bez CVE, zapytania badawcze, zakres `reachable` i lokalizacje źródeł. RI1–RI4 mają nadal opisane ograniczenia. Szczegóły produktu pozostają w [dokumentacji interfejsu](../docs/research-interface/README.md).

Materiał do adaptacji obejmuje [artykuł Cloudflare](https://blog.cloudflare.com/build-your-own-vulnerability-harness/) oraz [security-audit-skill przy commicie c1c8a8c1471069fb0e188eeaff69b8e8db6564a8](https://github.com/cloudflare/security-audit-skill/tree/c1c8a8c1471069fb0e188eeaff69b8e8db6564a8). Lokalna kopia była przeglądana w `/opt/src/security-audit-skill`. Materiał służy jako wzorzec zadań i dowodów. Implementacja Ariadne zachowuje własne wersjonowane kontrakty JVM.

Plan konsumenta znajduje się w `docs/RESEARCH_HARNESS_IMPLEMENTATION_PLAN.md` w [Tekmerion](https://github.com/tswierszcz/tekmerion).

## 2. Odpowiedzialność

| Ariadne | Tekmerion |
|---|---|
| Tożsamość klas, instancji, metod i scenariusza | Model zagrożenia, hipotezy i decyzje badawcze |
| Katalog API, wywołań i faktów grafu | Rozpoznanie dokumentacji i właściwości bezpieczeństwa |
| Ścieżki, warunki krawędzi i luki modelu | Zadania Huntera i niezależnego Validatora |
| Powtarzalne paczki odpowiedzi i dowody pochodzenia | Trwały rejestr pokrycia, kolejka, koszty i wznowienia |
| Korelacja wykonania z analizowanym kodem w RI6 | Harnessy, asercje skutku i raporty |

Tekmerion wywołuje Ariadne przez kod adaptera i przekazuje modelowi ograniczony kontekst. Decyzja o użyciu faktów grafu ma jawny ślad w rejestrze zadania.

Początkowy dostęp do LLM w Tekmerion będzie realizowany przez [OpenRouter](https://openrouter.ai), z osobnym wyborem modelu dla Recon, Hunt i Validate. Plan RH-T1 konsumenta obejmuje estymację kosztów przed badaniem, opcjonalną kalibrację oraz trwałe rezerwacje i rozliczanie kosztu żądań. Tekmerion zapisuje konfigurację modeli i metadane użycia, a klucz dostępu wyklucza z eksportowanych dowodów. Ariadne dostarcza pomiary czasu, pamięci i rozmiaru kontekstu potrzebne do planowania lokalnej analizy; koszty LLM rozlicza Tekmerion.

Plan wejścia Tekmerion obejmuje dostarczoną binarkę JVM, źródła przypięte do rewizji oraz opcjonalny pakiet CVE: identyfikator, advisory, poprawkę i reproducer. Sam identyfikator wystarcza do rozpoczęcia zbierania materiałów. Tekmerion utrwala ich pochodzenie i hashe oraz oddziela deklarowaną zgodność binarki ze źródłami od zgodności potwierdzonej. Oryginalny artefakt i przebudowana biblioteka zachowują osobne tożsamości. CVE nie jest wymagane do indeksowania ani zapytań badawczych Ariadne.

Przebieg może obejmować potwierdzenie znanego CVE, szukanie jego wariantów i szersze poszukiwanie nowych podatności, z osobnym zakresem i budżetem każdego etapu. Ariadne pozostaje początkowym backendem dowodów statycznych, wymaganym we wspólnym rzeczywistym pilocie JVM. Tożsamości metod, pochodzenie krawędzi i powtarzalne ścieżki wspierają selekcję i sprawdzanie hipotez. Nierozstrzygnięta osiągalność pozostaje luką dowodową; kontrolowany eksperyment może wykazać skutek przy nadal nierozstrzygniętej osi statycznej.

Tekmerion eksportuje planowane `report.md`, wersjonowane `report.json` i katalog dowodów z reproducerami. Raport rozdziela ocenę defektu, osiągalności, odtworzenia skutku i nowości, opisuje związek z CVE oraz różnice nowego wariantu. Zawiera kontrole, niezależną weryfikację, zakres pokrycia, odrzucenia, blokady i rozliczenie kosztów. Brak publicznego dopasowania nie potwierdza statusu zero-day. Brak znalezisk dotyczy wyłącznie zbadanego zakresu.

## 3. Etapy

| Etap | Powiązanie RI | Zakres | Kryterium odbioru |
|---|---|---|---|
| RH-A0 | RI0, RI3 | Akceptacja kontraktu na rzeczywistych wynikach | Wyniki poprawne i celowo uszkodzone mają zgodne rozstrzygnięcia po obu stronach |
| RH-A1 | RI4 | Powtarzalny kontekst Recon/Hunt | Tekmerion potrafi składać małą paczkę API/wywołań/źródeł bez utraty pochodzenia |
| RH-A2 | RI3, RI5 | Dowody do niezależnej weryfikacji | Validator otrzymuje odtwarzalne zapytanie i pełne wyjaśnienie ścieżki oraz luk |
| RH-A3 | RI5 | Kwalifikacja wspólnego przebiegu | Hipoteza bez CVE przechodzi przez analizę i kontrolowany test właściwości |
| RH-A4 | RI6 | Dowody wykonania dla własnych celów | Obserwacje mają zgodne hashe, zakres pomiaru i identyfikator eksperymentu |

### RH-A0 — akceptacja producenta i konsumenta

**Status 2026-11-25: IN_PROGRESS (strona producenta zrobiona).** Punkty 1–5 po stronie Ariadne: pakiet [acceptance/](../docs/research-interface/acceptance/) (biblioteka kontrolna, bundle `all` i `reachable`, 17 prawdziwych odpowiedzi wszystkich operacji, 132 mutacje z regułą), `ariadne research check` (werdykt producenta, reguły w [ACCEPTANCE.md](../docs/research-interface/ACCEPTANCE.md#3-rules)), `research query -verify full` i opis weryfikacji na żądanie, macierz producent/konsument. Te same reguły dają werdykty przypadków `v1/` (w tym odrzucenie `invalid-missing-with-path` i `invalid-path-shape`). Pozostaje strona konsumenta (RH-T0: pin pakietu i porównanie werdyktów) (każda reguła poza `replay` ma od 2026-11-25 mutację w pakiecie; `path-dispatch` dostał przypadek i poprawkę w RH-A2).

1. Zebrać rzeczywiste odpowiedzi z kontrolnej biblioteki dla wszystkich operacji. Zachować binarkę, scenariusz, `analysis_id` i żądania.
2. Sprawdzić schemat oraz semantykę odpowiedzi, w tym rozwiązanie celu, ciągłość krawędzi, korzenie, warunki i limity.
3. Udostępnić Tekmerion zestaw przypadków poprawnych oraz mutacji odpowiedzi do testów adaptera. W szczególności `missing` ze ścieżką i nieciągła ścieżka mają być odrzucane.
4. Opisać reguły weryfikacji integralności zbiorów ładowanych na żądanie. Aktualny czytnik w pełni weryfikuje zbiory użyte przez operację, a pozostałe sprawdza pod kątem istnienia i rozmiaru. Uzgodnić jawne ograniczenie lub tryb pełnej weryfikacji dla eksperymentów akceptacyjnych.
5. Utrwalić przypięty pakiet kontraktu i macierz producent/konsument. Zmiana pól lub semantyki wymaga aktualizacji kontraktu, przykładów i testów klienta.

Testy obejmują także brak/duplikaty pól, cel poza zakresem `reachable`, przesłoniętą klasę, kopie specjalizowane, budżet przed i po znalezieniu ścieżki, skróconą prezentację oraz niezgodne identyfikatory.

### RH-A1 — kontekst z istniejących operacji

1. Zmierzyć potrzeby pierwszych zadań Recon/Hunt na API plikowym. Zacząć od `entries`, `methods`, `calls`, `callers` i `callees`.
2. Uzgodnić z klientem paginację, wielkość paczek i ograniczanie sąsiedztwa. Pierwsza wersja klienta może składać kontekst z kilku istniejących zapytań.
3. Dopiero przy wykazanym koszcie kolejnych procesów rozważyć paczkowe zapytania kontekstowe lub ograniczone sąsiedztwo po stronie Ariadne. Zmianę poprzedzić pomiarem i przykładem klienta.
4. Zachować przypisanie fragmentu źródeł do instancji i hasha klasy, poziom `declared/build_verified/unavailable` oraz konflikty dopasowania. Tekmerion może dodać własny zweryfikowany indeks bez zmiany oryginalnego rekordu Ariadne.
5. Dostarczać luki i pominięte dane razem z kontekstem. Lista wejść lub wywołań nie daje sama oceny pokrycia audytu.

Odbiór obejmuje bibliotekę z przeciążeniami, klasą zagnieżdżoną, dwiema kopiami zależności, brakiem symboli i niejednoznacznym plikiem źródła. Powtórny odczyt tego samego indeksu daje te same fakty.

### RH-A2 — dowody do weryfikacji kandydata

**Status 2026-11-25: IN_PROGRESS (strona producenta zrobiona).** Biblioteka kontrolna dostała przypadki (j) client/supplied dispatch i (k) callback z warunkiem, z wyrocznią JVM. Reguła `path-dispatch` w `research check` przelicza flagi wpisu po wszystkich ścieżkach (CONTRACT.md §9) — wersja z RH-A0 wnioskowała je z pokazanej ścieżki. Ograniczenie lb2-7 jest w `warnings` każdej odpowiedzi `reachability` bundle'a `-roots api`, podzbiór korzeni na grafie RTA/PTA dostaje ostrzeżenie o przycięciu. Mapowanie pięciu punktów i czterech przypadków odbioru: [ACCEPTANCE.md §5](../docs/research-interface/ACCEPTANCE.md#5-evidence-for-an-independent-validator-rh-a2). Pozostaje: weryfikacja przez niezależny proces Validatora po stronie Tekmerion; usunięcie ograniczenia lb2-7 wymaga nowej rewizji kontraktu (miejsca parametrów korzenia w bundle'u).

1. Sprawdzić, czy odpowiedź dla konkretnej hipotezy zawiera wszystkie węzły, krawędzie i luki wymagane do weryfikacji wskazanej drogi.
2. Zachować rozróżnienie zwykłego wywołania, modelu, CHA, deklaracji wejściowej i warunku callbacka. Warunki różnych miejsc wywołania pozostają oddzielne.
3. Zweryfikować oznaczenia client/supplied dispatch. Znane ograniczenie parametrów wejściowych klienta ma pozostać widoczne, aż zostanie usunięte osobnym przyrostem.
4. Dla podzbioru korzeni ujawniać podstawę grafu oraz wpływ pierwotnego scenariusza RTA/PTA i modeli.
5. Umożliwić ponowienie żądania przez niezależny proces Validatora. Kolejność prezentacji i limity mają udokumentowane wersje reguł.

Odbiór obejmuje hipotezę potwierdzoną drogą bez CHA, drogę zależną od heurystyki, brak ścieżki przy istotnej luce oraz cel nieobecny w ograniczonym bundle'u. Każdy przypadek zachowuje wszystkie osie oceny.

### RH-A3 — wspólny pilot

**Status 2026-11-25: IN_PROGRESS (cel kontrolny po stronie Ariadne gotowy).** [testdata/pilot](../testdata/pilot/EXPECTED.MD): `px.archive.Unpacker#unpack` w wariancie wadliwym (`resolve` nazwy wpisu) i poprawnym (normalizacja i odmowa), te same API i ta sama metoda zapisu `Store#write`; klient-wyrocznia przez publiczne API na plikach lokalnych (scenariusze normalny, traversal, dozwolone API) pod `MethodTrace` — te same metody w obu wariantach, różny skutek (plik poza katalogiem tylko w wadliwym). Kontrole: martwa metoda (`not_found`), nierozwiązana refleksja (model `incomplete`), API dozwolone przez kontrakt (`found`, nie znalezisko). `TestPilotTarget`: wszystko, co wykonał JVM, jest `found`; odpowiedzi dla `Store#write` identyczne w obu wariantach; strażnik widoczny w `calls` `Entries#target`. `TestPilotConflictingSources`: binarka wadliwa ze źródłami poprawnymi dostaje hash podanego pliku z wiązaniem `declared` — zgodność binarki i źródeł rozstrzyga konsument. Pozostaje: eksperyment i asercja własności po stronie Tekmerion (RH-T4–T5), potem rzeczywista biblioteka.

Kontrolny target zawiera wadliwą implementację własności ograniczenia zapisu do katalogu oraz wariant zachowujący tę własność. Oba mogą osiągać tę samą metodę zapisu. Klient testowy korzysta z publicznego API, lokalnych plików i zarejestrowanego scenariusza.

Ariadne dostarcza dowód drogi. Tekmerion sprawdza wynik operacji i wiąże go ze źródłami, wejściem i asercją. Odbiór wymaga zgodności z niezależną obserwacją JVM oraz uwzględnienia martwej metody, nierozwiązanej refleksji i jawnie dozwolonej niebezpiecznej funkcji API jako przypadków kontrolnych.

Potem pilot obejmuje jedną rzeczywistą bibliotekę oraz jedną własność. Zmiany Ariadne wynikające z pilota otrzymują własne minimalne przykłady, testy regresji i pomiar wpływu na dotychczasowy pipeline CVE.

Rozszerzenie odbioru obejmuje dostarczoną binarkę z potwierdzonym i sprzecznym dopasowaniem źródeł, znany CVE i wariant poprawiony oraz raport z nierozstrzygniętą osiągalnością. Tekmerion porównuje także koszty i jakość przebiegów z podanym CVE oraz bez niego na tych samych przypadkach kontrolnych; wynik historycznego testu nie jest deklaracją znalezienia nowej podatności.

### RH-A4 — obserwacje runtime

Realizować w ramach RI6 po działającym pilocie, gdy obserwacja wywołania rozstrzyga konkretny problem. Filtr metod obejmuje własne cele Tekmerion. Korelacja zachowuje loader, hashe klas, rodzaj pomiaru, jego kompletność oraz identyfikator eksperymentu. Wynik asercji bezpieczeństwa pozostaje po stronie Tekmerion.

## 4. Wspólna kolejność

| Milestone | Ariadne | Tekmerion | Artefakt |
|---|---|---|---|
| H0 | RH-A0 | RH-T0 | Adapter odrzuca nieprawidłowe dowody kontraktu |
| H1 | Potrzebne operacje RH-A1 | RH-T1–RH-T2 | Trwałe jednostki pokrycia i mapa wejść/granic |
| H2 | RH-A1–RH-A2 | RH-T3 | Hipotezy z cytowaniami i drogami, możliwe do wznowienia |
| H3 | RH-A2–RH-A3 | RH-T4–RH-T5 | Niezależna weryfikacja z eksperymentem na oryginalnym kodzie |
| H4 | RH-A3 | RH-T6 | Raport pilota, przyczyny odrzuceń, koszt i luki pokrycia |
| H5 | RH-A4 według potrzeby | RH-T7 | Dalsze przyrosty uzasadnione wynikami pilota |

Pierwszym zadaniem sesji Ariadne jest RH-A0. Sesja Tekmerion może równolegle realizować RH-T0 i projekt rejestru RH-T1. Właścicielem kontraktu jest Ariadne, właścicielem promptów i rejestru badania jest Tekmerion.

## 5. Weryfikacja i status

Każdy PR określa etap, zmienione reguły, ograniczenia i dowody odbioru. Zmiany wspólnego pipeline'u wymagają `go test ./...`, `go vet ./...` oraz porównania istotnych grafów i wyników CVE na tych samych wejściach. Zmiany samego kontraktu wymagają walidacji przykładów i konsumenta. Pomiary obejmują czas, pamięć, rozmiar paczki i ponowne użycie indeksu.

Przygotowanie tego planu nie zamyka faz RI ani RH. Aktualizować stan konkretnego przyrostu w `IMPLEMENTATION_STATUS.MD`, a faktyczne możliwości i odstępstwa w README interfejsu oraz SUPPORT. Plan nie wprowadza nowych wymagań do wykonanych wcześniej analiz.
