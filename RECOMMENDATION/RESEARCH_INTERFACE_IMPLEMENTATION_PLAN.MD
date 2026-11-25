# Interfejs badawczy Ariadne dla Tekmerion

Data planu 2026-11-23. Aktualizacja 2026-11-24. Stan RI0 — specyfikacja i pakiet danych kontraktu przygotowane, akceptacja konsumenta (TM0) pozostaje do wykonania. RI1–RI3 IN_PROGRESS — pierwszy działający przekrój `research index`/`research query` (status poniżej, w sekcji 6). RI4–RI6 PLANOWANE. Kontrakt opisuje [dokumentacja interfejsu badawczego](../docs/research-interface/README.md).

## 1. Cel i podział odpowiedzialności

Udostępnić zewnętrznemu projektowi powtarzalne zapytania o kod JVM, wejścia API i osiągalność dowolnych metod, również gdy badany problem nie ma advisory ani numeru CVE. Zapytania mają działać na metodach samej biblioteki i jej zależności.

Projekt zewnętrzny to **[Tekmerion](https://github.com/tswierszcz/tekmerion)**. Plan po jego stronie znajduje się w `IMPLEMENTATION_PLAN.md`. Aktualizacja koordynacji obu repozytoriów pochodzi z 2026-11-24.

| Ariadne | Tekmerion, osobne repozytorium |
|---|---|
| Normalizacja zbudowanych artefaktów, wybór klas i instancji komponentów | Pobranie źródeł, wybór commita, build i konfiguracja eksperymentu |
| Graf JVM z modelami i pochodzeniem krawędzi | Lektura dokumentacji, model zagrożenia i generowanie hipotez przez LLM |
| Katalog API, instrukcji wywołań i luk modelu | Klasyfikacja operacji według CWE i własności bezpieczeństwa |
| Zapytania o osiągalność, ścieżki i kod pośredniczący | CodeQL, generowanie testów JUnit i harnessów Jazzer |
| Dopasowanie obserwacji JVM do analizowanych klas | Uruchamianie eksperymentów, asercje skutku, minimalizacja wejść |
| Wersjonowany kontrakt dowodów statycznych | Ocena hipotez, budżety LLM i raporty badawcze |

Publiczną granicą integracji będzie CLI i pliki JSON. Tekmerion przypina wersję binarki Ariadne. Ariadne dostarcza fakty o możliwości wywołania i ich ograniczenia. Ocena błędu, kontroli danych i skutku pozostaje osobnymi ustaleniami Tekmerion.

## 2. Stan kodu i elementy do wykorzystania

Punkt odniesienia to commit `c43ad7d`.

| Istniejący element | Wykorzystanie |
|---|---|
| `cmd/ariadne/main.go`, `analyze` | Obecne złożenie pipeline'u, również etapów modeli uruchamianych obecnie po pobraniu CVE |
| `internal/input`, `classfile`, `hierarchy`, `component` | Artefakty, klasy, deskryptory, hashe, pochodzenie, konflikty i wybór wersji |
| `internal/callgraph`, `analysis`, `models` | Jeden graf, koordynacja modeli, warunki krawędzi, limity i precyzja CHA/RTA/PTA |
| `internal/libapi` | Korzenie API, metody odziedziczone, specjalizacja odbiornika i wejścia ogłaszane frameworkom |
| `internal/reach/bfs.go`, `exposure.go`, `pathevidence.go` | Wyszukiwanie ścieżek, katalog narażonych wejść i dowody kroków |
| `internal/gaps`, `nativebound`, `reflectinvoke` | Wspólne mechanizmy luk oraz diagnostyka |
| `internal/runtimeevents` | Korelacja klas i wywołań po hashach oraz tożsamości loadera |
| `internal/snapshot`, `atomicfile` | Weryfikacja wejść, wersje reguł, hashe i bezpieczna publikacja artefaktów |

Obecne `reach.Engine.Analyze` przyjmuje listę CVE i mapowania. Potrzebna jest niższa warstwa zapytań o cele. Obecny eksport `ariadne-callgraph-2` służy porównywaniu grafów i zachowuje identyfikatory krawędzi oraz PC. Eksport badawczy musi dodatkowo zachować m.in. warunki miejsc wywołań, samodzielne węzły, korzenie i specjalizacje odbiornika.

Taint w `internal/taint` rozpoznaje aktualnie źródła HTTP i obsługuje pilotaż deserializacji. Parametr publicznej metody biblioteki wymaga jawnej deklaracji kontroli danych w scenariuszu. Pierwsze wydanie interfejsu umożliwi zewnętrznej analizie danych korzystanie z identyfikatorów Ariadne.

## 3. Proponowany interfejs CLI

Poniższe polecenia są projektem kontraktu. Pierwszy token `research` kieruje do nowej obsługi poleceń. Dotychczasowy parser wywołania `ariadne -app ...` zachowuje swój kontrakt.

```sh
ariadne research index \
  -app library.jar -deps dependencies -deps-list deps.txt \
  -roots api -profile auto -target-java 17 \
  -out analysis-bundle

ariadne research query \
  -bundle analysis-bundle -request request.json -out response.json
```

`index` korzysta z istniejących flag wejścia i konfiguracji, o ile wpływają na statyczny model. Pierwszy zakres przyjmuje wszystkie obecnie wspierane zbudowane wejścia, z biblioteką JAR jako przypadkiem pilotażowym. Wymagane są jawne korzenie `all`, `profile` albo `api`. CLI odrzuca nieobsługiwane flagi i opisuje dostępne możliwości w `research --help`.

`index` pracuje na dostarczonych artefaktach i konfiguracji, bez żądań sieciowych. Tekmerion dostarcza wcześniej zbudowany classpath. Pomiary CVE, wybór innych wersji z Maven Central oraz wywołania LLM należą do odrębnych operacji. `query` czyta gotowy bundle bez ponownego parsowania JAR-ów i budowania modeli.

Kody wyjścia mają następujące znaczenie.

| Kod | Znaczenie |
|---|---|
| 0 | Poprawnie wykonane zapytanie, również wynik `incomplete` albo `target_missing` |
| 1 | Błąd odczytu, zapisu lub wykonania |
| 2 | Błędne argumenty, nieobsługiwany schemat lub nieprawidłowe żądanie |

Diagnostyka procesu trafia na stderr. Wskazany plik JSON jest głównym wynikiem maszynowym. Przerwanie procesu przed publikacją wyniku pozostawia wykrywalny brak kompletnego artefaktu.

## 4. Kontrakt bundle'a i identyfikatorów

Bundle jest katalogiem z wersjonowanym manifestem i plikami danych. Proponowany format manifestu to `ariadne-research-bundle-1`.

Wymagane dane obejmują następujące elementy.

- Wersję narzędzia i reguł każdego uczestnika analizy, ustawienia grafu, scenariusz JPMS, target Java i aktywne profile.
- Hashe wejść, logiczną kolejność classpath, warianty multi-release i deklaracje komponentów osadzonych.
- Inwentarz instancji komponentów, pochodzenie klas, hashe wybranych klas, konflikty oraz klasy przesłonięte.
- Wszystkie węzły grafu, w tym metody izolowane, i odwzorowanie specjalizacji odbiornika na oryginalne metody.
- Wszystkie krawędzie z podstawą, rodzajem, miejscami wywołań i warunkami każdego miejsca.
- Korzenie i powód wyboru każdego korzenia, katalog API oraz oznaczenia kodu klienta i obiektów dostarczanych przez klienta.
- Metody z deskryptorami, flagami dostępu i dostępnymi informacjami o liniach oraz parametrach.
- Miejsca wywołań z deklarowanym celem, PC i rozpoznanymi celami grafu, także gdy deklarowany cel leży poza wejściem.
- Luki modelu, znany zakres celów i informację o osiągalności miejsca luki, limity oraz stan punktu stałego.

Tożsamość metody obejmuje instancję/wybrane pochodzenie klasy, nazwę klasy, nazwę metody i pełny deskryptor JVM. Węzeł specjalizowany ma własny identyfikator oraz odnośnik do metody i typu odbiornika. Miejsce wywołania identyfikuje metoda, PC i rodzaj operacji. Numery linii pozostają informacją prezentacyjną.

Pliki danych mają deterministyczną kolejność i SHA-256 w manifeście. `analysis_id` wynika z kanonicznych danych wejściowych, ustawień, wersji reguł oraz zawartości analizy. Czas utworzenia jest metadanymi przebiegu poza tym hashem.

Publikacja odbywa się przez zapis nowego katalogu roboczego, obliczenie hashy i opublikowanie kompletnego katalogu na tym samym systemie plików. Istniejący poprawny bundle pozostaje dostępny do zakończenia tej operacji. Czytnik sprawdza format, hashe, referencje i kompletność. Dwa równoległe zapisy pod tę samą nazwę kończą się jawnym konfliktem. Query może pracować na czytanym tylko bundle'u równolegle z innymi zapytaniami.

Query korzysta z zamrożonego bundle'a. Zmieniony później lokalny JAR wymaga nowego indeksowania. Powiązanie z zewnętrznym źródłem lub trace'em wymaga pasujących identyfikatorów i hashy.

## 5. Żądania i odpowiedzi

Schemat żądania w kontrakcie draft.1 to `ariadne-research-query-1`, odpowiedzi `ariadne-research-response-1`. [Specyfikacja, JSON Schema i przykłady](../docs/research-interface/README.md) są dokumentacyjną częścią RI0. Pole `analysis_id` wiąże żądanie z konkretnym bundle'em, `request_id` łączy odpowiedź z eksperymentem Tekmerion. Poniższy opis stanowi skrót, szczegóły formatu definiuje kontrakt.

### 5.1. Odkrywanie kandydatów

Operacje `entries`, `methods` i `calls` umożliwiają pobranie API, metod oraz miejsc wywołań z filtrami komponentu, klasy, nazwy i deskryptora. `callers` i `callees` dają sąsiedztwo wskazanej metody. Wyniki zachowują pochodzenie, podstawę krawędzi i warunki. Wywołanie API JDK może być widocznym miejscem wywołania również przy braku analizowanego ciała JDK.

Tekmerion używa tego katalogu przed wygenerowaniem hipotez. Może zestawić wejścia API z wywołaniami plikowymi lub sieciowymi, a następnie przekazać wybrane fragmenty źródeł LLM. Klasyfikację wrażliwych operacji i własności bezpieczeństwa wersjonuje Tekmerion.

Listy mają stabilną paginację i kolejność. Cursor jest przypięty do `analysis_id`, operacji i filtrów. Każda odpowiedź ujawnia liczbę zwróconych rekordów, następny cursor oraz powód ograniczenia.

### 5.2. Osiągalność własnych celów

Operacja `reachability` przyjmuje paczkę celów, jawne korzenie albo podzbiór korzeni indeksu oraz limity prezentacji. Cel jest selektorem istniejącej metody z pełnym deskryptorem. Pole instancji rozstrzyga kopie klas. Cel w przesłoniętej kopii otrzymuje informację o konflikcie.

Korzenie żądania muszą być podzbiorem korzeni indeksu. RTA/PTA i bramki modeli zależą od korzeni, dlatego szerszy albo inny scenariusz wymaga nowego indeksu. Zawężenie korzeni używa grafu zbudowanego dla pełnego zestawu i ujawnia tę podstawę. Nazwa bez deskryptora jest filtrem odkrywania metod. Zapytanie o osiągalność wymaga rozstrzygniętego celu.

Przykład żądania po odkryciu metod i identyfikatorów instancji.

```json
{
  "schema_version": "ariadne-research-query-1",
  "analysis_id": "sha256:...",
  "request_id": "experiment-001",
  "operation": "reachability",
  "targets": [{
    "id": "candidate-001",
    "selector": {
      "instance_id": "c-...",
      "class": "example.archive.Extractor",
      "name": "extract",
      "descriptor": "(Ljava/io/InputStream;Ljava/nio/file/Path;)V"
    }
  }],
  "roots": {"selection": "indexed"},
  "limits": {"max_entries": 20, "max_path_nodes": 100}
}
```

Odpowiedź opisuje osobno trzy osie.

| Oś | Proponowane wartości |
|---|---|
| Rozwiązanie celu | `matched`, `missing`, `ambiguous`, `shadowed` |
| Poszukiwanie ścieżki | `found`, `not_found`, `not_evaluated`, `incomplete` |
| Kompletność modelu | `complete_under_declared_model`, `incomplete`, `undetermined` |

`found` oznacza dowód ścieżki w wybranym grafie. Może współistnieć z lukami. `not_found` wolno zwrócić po pełnym przeszukaniu właściwego grafu. Istotne luki nadal są osobną, widoczną osią. Limit pracy, który zatrzymał przeszukiwanie, daje `incomplete` i zachowuje znalezione ścieżki. Brak celu daje `missing` i `not_evaluated`. Wartość `complete_under_declared_model` odnosi się wyłącznie do zadeklarowanego modelu i zarejestrowanych ograniczeń.

Wynik zawiera znalezione cele, identyfikatory wejść API, ścieżkę i dowody każdego kroku, rodzaj przejścia przez framework, oznaczenia client/supplied dispatch, warunki callbacków, źródła reguł oraz luki globalne i przypisane do celu. Reguła istotności luk ma własną wersję i opis założeń. Niepewna istotność jest widoczna. Kod badanego komponentu i kod pokrewnego artefaktu pozostają rozróżnialne.

Limity liczby pokazanych wejść i długości ścieżki ograniczają prezentację. Wynik ujawnia `truncated` i odróżnia całkowitą liczbę znalezionych wejść od pokazanych, gdy policzono całość. Skrócona prezentacja dodatniej ścieżki zachowuje `found`.

### 5.3. Informacje do weryfikacji eksperymentalnej

Każdy wynik niesie identyfikatory metod i wywołań nadające się do związania ze źródłami, zapytaniem CodeQL i obserwacją JVM. Lokalizacja źródłowa obejmuje klasę, plik i linię, gdy metadane istnieją. Dostawca źródeł podaje hash i powiązanie z buildem. Brak symboli debugowania pozostaje jawny.

Późniejszy przyrost `trace-config` przyjmie dowolne rozstrzygnięte cele i wygeneruje istniejący format ustawień JFR. `correlate` skojarzy trace z bundle'em na obecnych zasadach hashy i loaderów. Wynik oddzielnie opisze wejście do metody, zaobserwowaną relację wywołań i zakres pomiaru. Eksperyment Tekmerion dostarczy asercję skutku oraz kontrolowanych danych.

## 6. Fazy implementacji

| Faza | Zakres | Warunek odbioru |
|---|---|---|
| RI0 | Kontrakt, JSON Schema, przykłady wymiany, wersje i limity | Tekmerion może walidować przykładowe żądania i odpowiedzi bez silnika |
| RI1 | Wspólny pipeline statyczny i indeksowanie bez advisory | Pełne modele grafu działają przy braku sieci i pustym cache |
| RI2 | Bundle oraz katalogi metod, API i wywołań | Odczyt po restarcie procesu zachowuje tożsamości i dowody |
| RI3 | Paczkowe zapytania o cele i narażone wejścia API | Cele samej biblioteki i zależności rozstrzygane przez wspólną logikę grafu |
| RI4 | Kontekst do LLM i analizy źródłowej | Callers/callees, lokacje, luki i limity tworzą sprawdzalny pakiet kontekstu |
| RI5 | Eksperyment z Tekmerion i kwalifikacja pierwszego wydania | Hipoteza bez CVE prowadzi do sprawdzalnego celu i kontrolowanego testu JVM |
| RI6 | JFR dla dowolnych celów i korelacja z bundle'em | Obserwacje mają zgodne hashe, zakres i identyfikator eksperymentu |

### RI0 — kontrakt przed rozbudową CLI

- Przygotować schematy manifestu, żądań, odpowiedzi oraz przykładowy bundle.
- Rozstrzygnąć identyfikatory klas/węzłów, specjalizacje i warunki per miejsce wywołania.
- Zdefiniować obsługę nieznanych wersji, pól, operacji i limitów.
- Ustalić reguły `missing`, `ambiguous`, `shadowed` i relację limitu do wyniku.
- Zapisać interoperacyjne przykłady w Ariadne jako pliki danych do wykorzystania przez testy klienta Tekmerion.

### RI1 — wspólne przygotowanie grafu

**Status 2026-11-24: IN_PROGRESS.** Etap CVE (OSV, mapowanie) przeniesiony za statyczny pipeline w `analyze`; `research index` publikuje stan w tym punkcie (`researchState` w [research.go](../cmd/ariadne/research.go)). Uzupełnienie 2026-11-25: granica wydzielona — `prepare` zwraca `staticRun` (stan statycznego pipeline'u z funkcją cleanup, wykonywaną też po błędzie), `analyze` = `prepare` + `scanCVEs`, a `research index` woła `prepare` bezpośrednio. Pusty cache bez sieci: graf identyczny, ale komponenty osadzone wymagające danych referencyjnych Central (Petclinic: 2 w `ehcache`) nie są rozpoznawane, bundle zgłasza ich w `coverage.limits` ze statusem (`reference_unavailable`). Skan CVE bez zmian na Petclinic (all/profile/RTA), WebGoat, JPetStore i fixture biblioteki (`scripts/pipeline-compare.sh`). `proxycatalog.Find`: 2m12s → 0,96 s na Petclinic przy identycznych krawędziach.

- Wydzielić z `cmd/ariadne/main.go` granicę pipeline'u zwracającą inwentarz, hierarchię, finalny graf, korzenie, fakty frameworków, diagnostykę oraz funkcję cleanup.
- Zastosować tę samą ścieżkę przygotowania w dotychczasowej analizie CVE i `research index`.
- Zachować wszystkie modele, bramki twórców, źródła korzeni, warianty JVM i strategie grafu wymagane przez konfigurację.
- Oddzielić przejściowe pliki przygotowania od trwałych rekordów. Cleanup wykonuje się także po błędzie.
- Utrzymać `internal/analysis` jako koordynator punktu stałego. Granica całego pipeline'u ma oddzielną odpowiedzialność.
- Zmierzyć i sprofilować etap `proxycatalog.Find`, w szczególności powtarzane przeglądanie podtypów w `transactionalEdges`. W przeglądzie 2026-11-23 etap proxy trwał ok. 117 s na Petclinic i 187 s na WebGoat przy równoległych skanach. Wyniki stanowią punkt do pomiaru izolowanego. Optymalizacja wymaga potwierdzenia identycznych krawędzi i dowodów.

### RI2 — trwały indeks

**Status 2026-11-24: IN_PROGRESS.** Bundle `ariadne-research-bundle-1` z siedmioma zbiorami (po jednym na typ rekordu), atomowa publikacja, weryfikujący czytnik, deterministyczny `analysis_id`. Brak macierzy kompatybilności poza dokładną rewizją draftu. Uzupełnienie 2026-11-24: `-scope reachable` (Petclinic 525 MB zamiast 851 MB, te same odpowiedzi osiągalności) i strumieniowy czytnik (zapytanie 5,0 s / 1,6 GB zamiast 7,2 s / 2,8 GB).

- Wprowadzić rekordy wymiany dla danych wymaganych przez query, bez serializacji funkcji ani prywatnego układu struktur Go.
- Zapisać i odtworzyć pełne miejsca krawędzi z warunkami i specjalizacjami. Czytnik sprawdza odnośniki do metod, korzeni i instancji.
- Wprowadzić katalog wywołań obejmujący również API zewnętrzne, którego ciała nie weszły do grafu.
- Ustalić jawną macierz kompatybilności czytnika i wersji bundle'a. Zmiana semantyki danych podnosi wersję właściwego kontraktu/reguły.
- Zapewnić deterministyczny zapis, integralność i atomową publikację.

### RI3 — dowolne cele

**Status 2026-11-24: IN_PROGRESS.** Zapytania o cele w [internal/research](../internal/research/query.go) na rekordach bundle'a (nie na `reach.Engine`): te same przebiegi app_only/full, preferencja ścieżek bez kroków CHA, reguła client dispatch silnika zapisana na krawędziach. Zgodność z silnikiem sprawdzona na 256 wywołujących luk (Petclinic). Odstępstwa: [README](../docs/research-interface/README.md#implementation).

- Wydzielić z `internal/reach` rozwiązywanie celów i ścieżek niezależne od advisory. Analiza CVE nadal korzysta ze swoich reguł mapowania i agregacji.
- Wykorzystać istniejące BFS, specjalizacje, ekspozycję API i dowody ścieżek.
- Zachować wszystkie podstawy krawędzi. Preferencja ścieżki bez kroków wyłącznie CHA pozostaje jawnie wersjonowana.
- Udostępnić zapytania paczkowe i ponowne użycie indeksu odwrotnych krawędzi. Ścieżkę i dowody generować na żądanie, z limitami wyniku.
- Dodać reguły luk dla dowolnego celu, także należącego do analizowanej biblioteki. Istniejące przypisanie luk do komponentów dotkniętych CVE wymaga tu uogólnienia.

### RI4 — kontekst do badania

**Status 2026-11-24: IN_PROGRESS.** Lokalizacje źródeł: `source.path` z atrybutu `SourceFile` i pakietu, linie z LineNumberTable, `-sources DIR|JAR` dodaje `source_sha256` (powiązanie `declared`). Filtrowanie i paginacja działają od RI3. Pozostają: ograniczone sąsiedztwo z głębokością i opublikowany opis adaptera (adapter Tekmerion istnieje: tswierszcz/tekmerion#1).

- Udostępnić stabilne filtrowanie i paginację API, metod i miejsc wywołań.
- Umożliwić związanie metody z liniami źródła i pochodzeniem klasy. Przy braku linii zwracać deskryptor i PC.
- Zwracać ograniczone sąsiedztwo wywołań i diagnostykę, z jawnym kierunkiem oraz głębokością.
- Opublikować opis CLI i formatów umożliwiający przygotowanie cienkiego adaptera w Tekmerion.

### RI5 — pierwszy działający przebieg

- Użyć małej biblioteki kontrolnej z publicznym API, martwą metodą prywatną, dziedziczeniem, kodem klienta i jednym zapisem plikowym.
- Tekmerion generuje hipotezę dla wskazanej własności, zapisuje własny identyfikator i pyta Ariadne o wybrany cel bez CVE.
- Test JVM używa publicznego API i lokalnego katalogu testowego. Oddzielne asercje weryfikują wykonanie celu i własność operacji.
- Wariant kontrolny zachowujący własność może nadal osiągać ten sam cel. Test integracyjny utrwala ten przypadek.
- Następnie przeprowadzić pilot na jednej rzeczywistej bibliotece i jednej klasie błędów, z budżetem zapisanym po stronie Tekmerion.

### RI6 — dowody runtime

- Uogólnić wybór metod dla filtra JFR z celów mapowania CVE na cele zapytania.
- Reużyć obecny kontrakt i importer trace, uwzględniając wersję JDK oraz dostępne rodzaje zdarzeń.
- Korelować oryginalne metody i odpowiednie specjalizacje, zachowując pochodzenie klasy i loadera.
- Zachować identyfikatory scenariusza, eksperymentu oraz kompletność pomiaru.

## 7. Testy i pomiary

Testy mają sprawdzać zachowanie przez normalne wejścia CLI oraz zgodność obu konsumentów wspólnego silnika.

| Przypadek | Sprawdzana własność |
|---|---|
| Cel prywatny osiągany z publicznej metody | Ścieżka przez prawdziwe wejście API |
| Prywatna metoda bez wejść | `not_found` po kompletnym przeszukaniu |
| Metoda nieobecna i niejednoznaczny selektor | `missing`/`ambiguous` bez negatywnego wniosku |
| Dwie kopie klasy, różne wersje | Instancja, wybór klasy i `shadowed` |
| Przeciążenia i deskryptory | Rozstrzygnięcie dokładnej sygnatury |
| Metody odziedziczone i specjalizacje | Identyczne cele i dowody przed/po serializacji |
| Callback sukcesu/porażki, AOP i wątki | Zachowane warunki i podstawa krawędzi |
| Nierozwiązana refleksja lub JNI | Widoczna niekompletność modelu |
| Cel w bibliotece i w zależności | Ocena obu właścicieli kodu |
| Wywołanie API JDK spoza wejścia | Dostępne miejsce wywołania z oznaczeniem braku ciała |
| Limit pracy i limit prezentacji | Rozróżnienie nieukończonej analizy i obciętej odpowiedzi |
| Błędny hash, brak pliku, nieobsługiwany schemat | Jawny błąd wejścia bundle'a |
| Przerwany i równoległy zapis | Dostępny wyłącznie spójny opublikowany indeks |
| Zapytanie po restarcie procesu | Odpowiedź z bundle'a bez ponownego builda |
| Kod z luką i wariant kontrolny | Osiągalność oddzielona od wyniku asercji bezpieczeństwa |

Na Petclinic, WebGoat i JPetStore porównać graf, korzenie, warunki krawędzi, ustawienia i wyniki CVE przed wydzieleniem pipeline'u oraz po nim na tych samych wejściach i danych. Rozdzielić oczekiwaną zmianę metadanych wersji od zmiany semantyki. Powtórzyć istotne konfiguracje CHA/RTA/PTA i `roots all/profile/api` w odpowiednim zestawie kwalifikacji.

Zmierzony koszt obejmuje indeksowanie, rozmiar bundle'a, czas odczytu, pamięć szczytową oraz zapytania dla 1/10/100 celów na gotowym indeksie. Zapisać sprzęt, ustawienia i hashe wejść. Kryterium funkcjonalne to pojedyncze przygotowanie grafu na scenariusz i kolejne paczki zapytań bez ponownego parsowania JAR-ów. Budżety wydajnościowe ustalić po pomiarze bazowym.

Przed scaleniem fazy uruchomić właściwe testy pakietów, a dla wspólnego pipeline'u i wydania pełne `go test ./...` oraz `go vet ./...` zgodnie z B4. Weryfikacja funkcjonalna pierwszego wydania używa lokalnych fixture i deterministycznych żądań. Eksperyment z LLM zapisuje model, wersję instrukcji i odpowiedź jako osobne dane Tekmerion.

## 8. Kolejność dostarczania i dokumentacja

Pierwszy działający interfejs obejmuje RI0–RI3. RI4–RI5 domykają pierwsze wydanie używane przez Tekmerion. RI6 jest kolejnym przyrostem dowodów runtime.

Każda faza ma osobny PR z opisem zakresu, poleceniami weryfikacji i znanymi ograniczeniami. Po implementacji zaktualizować README i SUPPORT oraz dopisać osobny kontrakt w `docs/`. Wersje dotychczasowych algorytmów podnosić przy rzeczywistej zmianie ich semantyki. Wersjonowany interfejs badawczy ma własną historię zmian.

Generowanie hipotez, katalogi własności, ogólny taint, interpretacja warunków ataku, uruchamianie kodu generowanego przez LLM oraz publikowanie zgłoszeń są zadaniami Tekmerion. Ewentualne nowe fakty silnika JVM wracają do Ariadne jako osobne, mierzone przyrosty. Nowy klient powinien korzystać z opublikowanej wersji kontraktu i zachowywać surowe odpowiedzi Ariadne przy eksperymencie.

## 9. Kryterium ukończenia

Tekmerion potrafi zindeksować dostarczoną bibliotekę i jej classpath bez advisory, odkryć wejścia oraz operacje wywołań, wskazać własny cel, otrzymać drogę od API z dowodami i lukami, a następnie związać ją z kontrolowanym testem. Bundle można przenieść i odczytać w nowym procesie, zachowując integralność i deklarowaną kompatybilność. Regresje dotychczasowego skanowania CVE są rozstrzygnięte przed wydaniem wspólnego pipeline'u.

## 10. Wspólna kolejność pracy z Tekmerion

| Kamień milowy | Ariadne | Tekmerion | Wynik do przeglądu |
|---|---|---|---|
| M0 — umowa wymiany | RI0 | TM0 | Schematy, przykładowe komunikaty i semantyka niepewności |
| M1 — obraz biblioteki | RI1–RI2 | TM1 | Odtwarzalny build, indeks bez CVE, katalog API i wywołań |
| M2 — pierwsze hipotezy | RI3 i potrzebna część RI4 | TM2–TM3 | Hipotezy wygenerowane z kodu, dowody statyczne i powody odrzucenia |
| M3 — eksperyment | RI5 | TM4–TM5 | Test przez API, niezależna asercja skutku i wariant kontrolny |
| M4 — pierwszy pilot | Ukończone RI4–RI5 | TM6 | Raport jednej rzeczywistej biblioteki, odtwarzalny przebieg i pomiar kosztu |
| M5 — dowody wykonania | RI6 | Rozszerzenie TM4–TM5 | Trace powiązany z celami eksperymentu i właściwymi hashami |

TM0–TM6 są etapami planu Tekmerion. TM0 i TM1 mogą być realizowane na uzgodnionych fixture kontraktu podczas prac RI1–RI2. Adapter ujawnia wtedy pochodzenie danych jako fixture. Pilot M4 korzysta z rzeczywistej binarki Ariadne.

Pierwszy zakres implementacyjny po stronie Ariadne to RI0. Jego PR dostarcza schematy JSON i fixture dodatniej ścieżki, braku ścieżki, brakującego celu oraz luki modelu. Pierwszy zakres Tekmerion to TM0, który waliduje te same przykłady i zapisuje identyfikator kontraktu. Kolejnym przyrostem jest RI1 z porównaniem dotychczasowych grafów oraz TM1 z manifestem wejścia i kontrolowanym buildem.

Ariadne jest właścicielem schematów bundle'a, query i response. Tekmerion jest właścicielem schematów scenariusza, hipotezy, eksperymentu i raportu. Wydanie klienta przypina wersję kontraktu i identyfikator zestawu fixture. Zmiana wspólnego kontraktu otrzymuje przykład migracji oraz test konsumenta przed aktualizacją przypiętej wersji. Dane modelu zagrożenia przekazywane jako kontekst zachowują pochodzenie jako deklaracje użytkownika lub LLM.

Każdy plan zapisuje wyłącznie postęp swojego repozytorium i wskazuje odpowiadający mu PR lub commit drugiej strony, gdy integracja jest gotowa. Wpis DONE wymaga artefaktów odbioru z tabeli. Niniejsza aktualizacja ustanawia plany i kolejność prac.
