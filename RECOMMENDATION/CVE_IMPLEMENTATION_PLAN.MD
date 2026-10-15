# Rozpoznawanie CVE — plan implementacyjny w fazach

Status: plan, implementacja nie została rozpoczęta w ramach tego dokumentu.

Wymagania funkcjonalne: [CVE.MD](CVE.MD). Szersze rekomendacje rozwoju JVM: [MD](MD).

## 1. Zasady realizacji

- Każda faza kończy się działającym przyrostem, sprawdzalnymi scenariuszami i aktualizacją dokumentacji.
- Nazwy nowych pakietów, plików, pól oraz flag poniżej są propozycjami. Nie są opisem dostępnego API.
- Brak danych, błędy źródeł i konflikty nie mogą być interpretowane jako brak podatności.
- Dokładność sygnatury, pochodzenie dowodu, kompletność mapowania i osiągalność są osobnymi informacjami.
- Wynik negatywny dotyczy zakresu modelowanego grafu; nie dowodzi bezpieczeństwa w obecności niewspieranych mechanizmów JVM.
- Nie wprowadzamy automatycznych zapytań LLM ani kosztów jako efektu zwykłego odświeżania CVE.
- Zachowujemy możliwość pracy offline i czysto go-ową ścieżkę analizy istniejących artefaktów. Nowy parser lub porównywarka wersji nie może niejawnie wymagać JDK/cgo.
- Zmiany w istniejącym `PLAN.md` i plikach użytkownika nie są nadpisywane przy realizacji tego planu.

## 2. Stan wyjściowy w repozytorium

| Obszar | Obecna implementacja | Konsekwencja dla planu |
|---|---|---|
| Identyfikacja zależności | `internal/input/resolve.go`, `gav.go`: jeden wybrany GAV na `DepJar`. | Oddzielić fizyczny artefakt od komponentów i ich kopii. |
| Hierarchia | `internal/hierarchy`: jedna klasa na FQCN; odczyt JAR-ów równoległy. | Ustalić stabilny wybór klas i jawnie reprezentować konflikty. |
| Graf | `internal/callgraph`: klucz zawiera FQCN, nazwę i deskryptor; CHA wybiera metody po nazwie. | Wykorzystać deskryptory również w rozwiązywaniu wywołań i dopasowaniu. |
| Dane CVE | `internal/cves`: OSV i deduplikacja aliasów; wynikowy `Affected` zawiera pakiet i wersję. | Zachować także dane źródłowe zakresów, dat i wycofania. |
| Mapowania | `internal/vulnmap`: `map[string]Entry`, metody `[]string`, priorytet manual → diff → LLM. | Wprowadzić mapowania dla komponentu, wersji i instancji oraz ich dowody. |
| Analiza poprawki | `internal/javadiff` i `vulnmap/diffsource.go`: nowe linie i źródło po commicie. | Obsłużyć oba stany pliku i usunięcia. |
| Dopasowanie metod | `internal/reach/match.go`: nazwa bez deskryptora. | Wprowadzić rozróżnienie dokładne/przybliżone. |
| Cache | `internal/httpcache`: pliki z treścią, bez czasu ważności. | Dodać metadane, klasy zasobów i politykę odświeżania. |
| Wyniki | `internal/report`: lista `reach.Result`, jeden wynik na CVE. | Dodać wersjonowany wynik z wieloma ustaleniami na CVE. |
| Runtime JDK | Brak opisu docelowego runtime’u i adaptera advisory. | Zbudować osobny tor identyfikacji i oceny podatności runtime’u. |

## 3. Zależności i punkty wydania

```text
F0: kontrakty i przypadki odniesienia
  → F1: komponenty, dowody, wynik v2
  → F2: świeżość cache
  → F3: snapshoty i odtwarzanie
  → F4: wersje, sygnatury, dopasowanie
  → F5: analiza obu stron poprawki
  → F6: biblioteki osadzone bez relokacji
  → F7: jawna relokacja
  → F8: docelowy runtime i pierwszy dostawca JDK
  → F9: integracja, migracja i kwalifikacja wydania
```

Kolejność jest zalecaną sekwencją wdrażania. F8 zależy przede wszystkim od F1–F3; nie wymaga ukończenia heurystyk relokacji. Nie oznacza to obowiązku równoległej pracy.

- **Pierwszy przyrost:** F0–F5 oraz odpowiednie sprawdzenia wydaniowe z F9. Aktualne dane i lepsze mapowania dla zwykłych JAR-ów.
- **Drugi przyrost:** F6–F7 oraz sprawdzenia z F9. Komponenty osadzone i jawna relokacja.
- **Trzeci przyrost:** F8 oraz sprawdzenia z F9. Podatności runtime’u dla pierwszego dostawcy.

## F0. Kontrakty, zakres i przypadki odniesienia

### Cel

Ustalić znaczenie nowych wyników oraz mierzalną bazę przed zmianą algorytmów.

### Zadania

1. Opisać słowniki: identyfikacja komponentu, podatność wersji, precyzja dopasowania, źródło dowodu, kompletność i osiągalność.
2. Zdefiniować stabilne kody powodów, np. `VERSION_UNKNOWN`, `METHOD_NOT_FOUND`, `MAPPING_PARTIAL`, `SOURCE_UNAVAILABLE`, `CLASSPATH_CONFLICT`, `RUNTIME_UNSUPPORTED`.
3. Przeanalizować istniejące ręczne mapowania, w tym rolę XStream `fromXML` jako punktu wejścia. Nie oznaczać wszystkich wpisów jako kompletnych mapowań kodu podatnego.
4. Zdefiniować kontrolny zestaw przypadków opisanych w sekcji 5. Użyć małych fikcyjnych advisory do testowania mechaniki i osobnego, źródłowo udokumentowanego zestawu prawdziwych CVE do oceny jakości.
5. Utrwalić baseline bieżącej wersji z zapisanymi danymi sieciowymi. Wersje, źródła i oczekiwane rozstrzygnięcia muszą być zapisane; sam wynik aktualnego narzędzia nie jest prawdą odniesienia.
6. Zapisać decyzje o kompatybilności JSON, obsłudze starych mapowań i zasadach agregowania wyników.

### Moduły

`internal/vulnmap`, `internal/reach`, `internal/report`, `cmd/ariadne`, dokumentacja i istniejące testdata.

### Kryteria ukończenia

- Każdy stan ma opis oraz przykład; brak metody i brak ścieżki są rozróżnione.
- Zestaw kontrolny obejmuje wszystkie pięć rozszerzeń.
- Ustalono, które stare wyniki zmienią się ze względu na ostrożniejszą semantykę, a które powinny pozostać zgodne.

## F1. Model komponentów, dowodów i wyników v2

### Cel

Przestać przypisywać jedną listę metod do CVE niezależnie od artefaktu i wersji.

### Zadania

1. Wprowadzić model komponentów, np. `internal/component`: `Artifact`, `ComponentInstance`, `IdentityEvidence`.
2. Dla fizycznego artefaktu zapisywać SHA-256. Dla instancji komponentu zapisywać tożsamość logiczną, wersję, artefakt nadrzędny, sposób wykrycia i zakres przypisanych klas.
3. Identyfikatory tworzyć na podstawie zawartości i stabilnej lokalizacji wewnątrz wejścia. Losowe ścieżki plików tymczasowych nie mogą zmieniać identyfikatorów.
4. Dodać strukturalne `MethodSelector`, `MappingEvidence`, `Finding`. Deskryptor w selektorze jest opcjonalny, a rola metody i kompletność mapowania jawne.
5. Zmienić wynik resolvera z klucza CVE na klucz obejmujący advisory i instancję komponentu. Nie usuwać informacji o źródłowych rekordach po scaleniu aliasów.
6. Powiązać `ClassInfo` z pochodzeniem i wybraną instancją komponentu; początkowo zwykły JAR zawiera jeden rozpoznany komponent.
7. Ustalić deterministyczny wybór klasy przy konflikcie, niezależny od kolejności kończenia workerów. Raportować przyjęte założenie kolejności classpath. Jeśli kolejność rzeczywistego wdrożenia jest nieznana, oznaczyć zależne ustalenia jako niepewne.
8. Zachować inwentaryzację kopii niewybranych do grafu. Nie używać ścieżki do wybranej klasy jako dowodu osiągalności innej kopii.
9. Dodać opt-in `-json-format v2`: obiekt z `schema_version`, komponentami, ustaleniami, pokryciem i odwołaniem do snapshotu. Format v1 pozostaje dostępny na czas migracji.
10. Określić agregację v1: pozytywny dowód jest widoczny; bez pozytywnego dowodu nierozstrzygnięta kopia nie może zniknąć pod `NOT_REACHABLE`. `NOT_APPLICABLE` zachowuje znaczenie dotyczące WebJar-ów. Liczyć osobno unikalne advisory i ustalenia per komponent.

### Kryteria ukończenia

- Dwie instancje tego samego komponentu mają osobne wyniki i dowody.
- Losowa kolejność odczytu nie zmienia wybranej klasy.
- Wyjście v2 nie miesza niepewności mapowania z osiągalnością; dokumentacja opisuje ograniczenia eksportu v1.
- Dotychczasowe polecenia pozostają poprawne składniowo; uzasadnione zmiany werdyktów są opisane.

## F2. Cache z metadanymi i polityką świeżości

### Cel

Uniemożliwić bezterminowe używanie starych odpowiedzi jako aktualnych danych CVE.

### Zadania

1. Rozszerzyć `internal/httpcache` o zapis treści i metadanych jako jednej spójnej jednostki: czas pobrania, czas ostatniej walidacji, hash, typ zasobu, identyfikator żądania i ewentualne walidatory HTTP.
2. Nie zapisywać sekretów ani nagłówków autoryzacji w metadanych. Zachować bezpieczny zapis atomowy przy równoległych żądaniach.
3. Wyróżnić zasoby zmienne OSV, źródła przypięte pełnym SHA, wyniki pochodne i odpowiedzi LLM.
4. Dodać konfigurowalny TTL dla danych OSV; proponowany domyślny okres to 24 godziny. Obejmuje to odpowiedzi querybatch/query bez podatności.
5. Odświeżać zarówno zapytania o listę CVE, jak i szczegóły rekordów. Zachować `modified`, `published`, `withdrawn`, zakresy i oryginalne źródła w modelu `internal/cves`.
6. Przeliczać aliasy po odświeżeniu. Wycofanie jednego rekordu nie wycofuje automatycznie całego CVE, jeśli pozostaje inne aktywne źródło.
7. Dodać proponowane `-refresh`, `-max-data-age`, `-allow-stale`. Połączenie `-refresh` i `-offline` jest błędem walidacji.
8. Zwykłe `-offline` może użyć starego kompletnego cache z widocznym ostrzeżeniem. Jawne `-max-data-age` stanowi twarde wymaganie i w offline może zakończyć analizę błędem. Brak wymaganego wpisu offline nadal jest błędem.
9. Online nieudane wymagane odświeżenie kończy analizę błędem; użycie starego wpisu wymaga `-allow-stale`. Oznaczać wtedy niepełną świeżość całego raportu.
10. Stare wpisy cache bez metadanych traktować jako dane o nieznanym wieku: online wymagają walidacji; offline są oznaczane. Nie przypisywać im daty migracji jako daty pobrania.
11. Przekazywać zegar do logiki TTL, aby sprawdzać granice ważności bez oczekiwania. Zachować rozdzielenie odświeżania CVE i ponawiania LLM.

### Moduły

`internal/httpcache/cache.go`, `fetch.go`, `internal/cves/osv.go`, `cves.go`, `cmd/ariadne/main.go`, nagłówek i sekcja pokrycia raportu.

### Kryteria ukończenia

- Ta sama aplikacja może otrzymać nowe CVE po odświeżeniu wcześniej pustej odpowiedzi.
- Wiek danych i awaria odświeżenia są widoczne i mają jednoznaczną semantykę.
- Wycofane źródło i aktywny alias są poprawnie rozpatrywane.
- Odświeżenie CVE nie powoduje automatycznych płatnych zapytań LLM.

## F3. Snapshoty, odtwarzanie i unieważnianie wyników pochodnych

### Cel

Pozwolić odtworzyć analizę oraz ustalić, dlaczego wynik się zmienił.

### Zadania

1. Wprowadzić `internal/snapshot` z manifestem wersji formatu i magazynem treści identyfikowanych hashem.
2. Manifest zawiera dokładne odpowiedzi i metadane źródeł użyte w analizie, mapowania, profile, konfigurację, wersje algorytmów, wersję narzędzia i hashe wejściowych artefaktów.
3. Zapisywać także obserwacje negatywne i błędy, jeżeli wpłynęły na wynik. Nie wymagać skopiowania całego współdzielonego cache.
4. Zdefiniować klucz mapowania pochodnego: instancja komponentu + zakres wersji + hash advisory + hashe źródeł poprawki + wersja resolvera. Dla LLM również model, prompty i odpowiedź.
5. Dodać proponowane `-snapshot-out` oraz `-snapshot-in`. Odtwarzanie nie odświeża źródeł ani nie wysyła żądań sieciowych; brak danych w snapshotcie jest widocznym błędem.
6. Przy odtwarzaniu sprawdzać hashe dostarczonych artefaktów. Domyślnie nie archiwizować automatycznie cudzych binariów; manifest wskazuje wymagane wejścia.
7. Nie obiecywać zgodności dla zmienionego kodu analizatora: wymagać zgodnej wersji i wersji algorytmów. Dla lokalnego dirty build zapisywać brak pełnej odtwarzalności, chyba że zachowano odpowiedni build lub źródła.
8. Oddzielić identyfikator logicznego zestawu danych od bieżącej daty wygenerowania raportu. Wyniki v2 dla tego samego snapshotu, konfiguracji i analizatora mają stabilną kolejność oraz treść pól merytorycznych.

### Kryteria ukończenia

- Snapshot przechodzi analizę bez sieci przy zgodnych artefaktach.
- Zmieniony artefakt jest odrzucany jako niezgodny z manifestem.
- Zmiana advisory, mapowania lub wersji algorytmu unieważnia właściwe wyniki pochodne.
- Dokumentacja uczciwie rozróżnia odtworzenie danych od odtworzenia całego środowiska wykonania narzędzia.

## F4. Mapowania wersji i dokładnych sygnatur

### Cel

Stosować właściwy zakres wersji, komponent i przeciążenie, nie wprowadzając fałszywej precyzji grafu.

### Zadania

1. Dodać schemat ręcznych mapowań v2 z walidacją nieznanych pól, identyfikatorów, zakresów i deskryptorów.
2. Stary YAML ładować przez adapter jako mapowanie bez jawnego zakresu, dopasowanie po nazwie i niepotwierdzona kompletność. Nie podnosić automatycznie jakości starych dowodów.
3. Dodać moduł porównywania wersji Maven zgodny z uzgodnionymi regułami i zestawem przypadków porównawczych. Obsłużyć granice przedziałów i kwalifikatory; nie używać standardowej porównywarki SemVer.
4. Dla nietypowych wersji, snapshotów lub nieobsługiwanych zakresów zwracać stan nierozstrzygnięty. Nie zastępować błędu porównaniem tekstowym.
5. Wprowadzić dokładny lookup `fqcn + name + descriptor` w `internal/hierarchy`. Obsłużyć konstruktory, przeciążenia, metody dziedziczone, interfejsy i metody bridge zgodnie z ich rolą w bytecode.
6. Uzupełnić `internal/callgraph/build.go`: wywołanie dziedziczone ma prowadzić do właściwej implementacji; rozwinięcie CHA musi respektować deskryptor. Dopasowanie po nazwie nie może tworzyć krawędzi między niepowiązanymi przeciążeniami.
7. Jeśli odmienny wariant metody jest związany z covariant return/generics, wykorzystywać rzeczywiste bridge methods, gdy są obecne, zamiast arbitralnie uznawać wszystkie sygnatury za równoważne.
8. Zmienić `internal/reach/match.go`, by rozdzielał dopasowanie dokładne i przybliżone oraz weryfikował instancję komponentu.
9. Resolver wybiera ręczne mapowanie tylko dla pasującego komponentu i zakresu. Przechowuje konflikty i zakres dowodów; samo wcześniejsze źródło nie ukrywa niezgodności z bytecode.
10. Zachować role `vulnerable` i `entry_point` z odpowiedzi LLM w modelu v2. Pomyślne sprawdzenie obecności metody nie potwierdza jej podatności ani kompletności mapowania.
11. Przy braku pasujących metod, częściowym mapowaniu lub konflikcie nie wydawać mocnego werdyktu negatywnego. Pozytywną ścieżkę opisywać z jakością jej dowodów.

### Moduły

`internal/vulnmap/manual.go`, `resolve.go`, `llmanswer.go`, nowa logika wersji, `internal/hierarchy`, `internal/callgraph`, `internal/reach`, `internal/report`.

### Kryteria ukończenia

- Dwa przeciążenia i dwa zakresy wersji dają rozłączne dopasowania.
- Graf nie dokleja podatnego przeciążenia wyłącznie ze względu na wspólną nazwę.
- Wywołania odziedziczone zachowują dostęp do ciała właściwej metody.
- Przypadki brakującej metody, nieznanej wersji i konfliktu pozostają nierozstrzygnięte z kodem powodu.

## F5. Obie strony poprawki i powiązanie z badanym wydaniem

### Cel

Wskazywać kod występujący w podatnym wydaniu, także gdy poprawka go usuwa lub przenosi.

### Zadania

1. Rozszerzyć `internal/javadiff.FileDiff` o stare i nowe ścieżki, stare i nowe linie zmian oraz typ zmiany. Zachować dodania, usunięcia i renames.
2. Zbudować dostawcę źródeł commita w `internal/vulnmap`: pełne SHA, rodzice, obie rewizje, metadane plików. Nie zakładać, że patch zwrócony przez API jest zawsze kompletny.
3. Dla pojedynczego rodzica używać właściwej pary rewizji. Dla merge commita wymagać jednoznacznej podstawy porównania; w przeciwnym razie zapisywać niepewność.
4. Pobierać starą treść dla usuniętych i przemianowanych plików. Zachować dowody: repozytorium, SHA obu rewizji, ścieżka, strona diffu, linie i hash źródła.
5. Rozszerzyć skaner metod o dane potrzebne do rozróżnienia kandydatów. Gdy nie da się jednoznacznie odtworzyć deskryptora z Java source, np. z powodu erasure, importów lub klas wewnętrznych, rozstrzygać przy pomocy bytecode lub pozostawiać dopasowanie przybliżone. Nie zgadywać sygnatury.
6. Rozróżnić metodę usuniętą, zmienioną oraz dodane zabezpieczenie. Automatyczny diff generuje dowód-kandydata, nie certyfikat kompletnego zestawu metod podatnych.
7. Oceniać powiązanie commita z CVE, komponentem i gałęzią wydania. Referencje `FIX` oraz inne linki do commitów zachować jako odrębne rodzaje dowodów. Nie łączyć ślepo mapowań dla różnych wydań.
8. Zweryfikować kandydatów w klasach konkretnego artefaktu. Opcjonalny sources JAR może być dodatkowym źródłem, jeżeli został powiązany z badaną wersją.
9. Wnioskowanie o przemianowaniu metody na podstawie podobieństwa pozostawić jako heurystykę; jawnie wskazywać niejednoznaczność.
10. Błędy źródeł mapowania dotyczące pojedynczego CVE raportować jako brak/niepełność mapowania i kontynuować pozostałe CVE. Złamanie globalnego kontraktu offline, snapshotu lub wymaganej świeżości nadal kończy uruchomienie błędem.
11. Aktualizować klucze cache pochodnego i opcjonalne prompty LLM o obie strony poprawki oraz zakres komponentu. Zmiana promptu nie uruchamia LLM bez odpowiedniej opcji użytkownika.

### Kryteria ukończenia

- Usunięta podatna metoda jest kandydatem na podstawie starego kodu.
- Nowy helper zabezpieczający nie jest automatycznie uznany za istniejący podatny kod starej wersji.
- Zmiana nazwy pliku, dwie gałęzie, merge commit i brak źródeł mają odrębne oczekiwane wyniki.
- Przypadki nierozstrzygnięte pozostają widoczne, zamiast poprawiać statystyki kosztem błędnego `NOT_REACHABLE`.

## F6. Komponenty osadzone bez relokacji

### Cel

Rozpoznawać wiele komponentów wewnątrz jednego JAR-a i oceniać je osobno.

### Zadania

1. Rozszerzyć `internal/input` o zachowanie wszystkich kandydatów `pom.properties`, niezależnie od wyboru głównego artefaktu.
2. Zbudować inwentaryzację komponentów: osobny fizyczny JAR i lista potencjalnie osadzonych bibliotek. Wiele plików POM może opisywać moduły projektu, a nie niezależne opublikowane biblioteki.
3. Dodać opcjonalny import SBOM w jednym wybranym formacie, np. CycloneDX JSON. Weryfikować związek SBOM z artefaktem; sam wykaz zależności nie dowodzi osadzenia ich kodu.
4. Potwierdzać obecność na podstawie klas i, gdy potrzebne, porównania z referencyjnym artefaktem. Pobieranie referencji kierować przez cache z ograniczonym i udokumentowanym zestawem źródeł.
5. Modelować wersję potwierdzoną, zbiór kandydatów lub wersję nieznaną. Nie wykonywać zapytania o przypadkowo wybraną wersję jako rozstrzygającego.
6. Przypisywać klasy do komponentów z informacją o niejednoznaczności. Nie nadawać całemu JAR-owi wielu GAV-ów bez mapy klas.
7. Zapytania OSV wykonywać dla rozpoznanych wersji. Przy ograniczonym, uzasadnionym zbiorze wersji łączyć możliwe CVE jako kandydatów. Przy wersji zupełnie nieznanej wykazać lukę pokrycia i nie obiecywać kompletnej listy CVE.
8. Wybrane do grafu klasy wiązać z odpowiednimi komponentami. Dla kopii przesłoniętych lub niejednoznacznych raportować brak rozstrzygającej oceny osiągalności.
9. Dodać lokalizację komponentu, dowody identyfikacji, braki pokrycia i osobne wyniki do raportu oraz snapshotu.

### Kryteria ukończenia

- Osadzona biblioteka z zachowanymi pakietami otrzymuje odrębny wynik.
- Stare metadane bez odpowiadających im klas nie dają identyfikacji potwierdzonej.
- Dwie kopie w różnych wersjach nie są scalane w jedno mapowanie.
- Pominięte lub nieznane wersje są widoczne w pokryciu.

## F7. Jawna relokacja pakietów

### Cel

Dopasowywać znany komponent po zmianie nazw jego klas i odwołań.

### Zadania

1. Dodać opcjonalny plik opisu komponentów osadzonych i relokacji, np. `-embedded-components`, przypięty do hashy artefaktów. Może deklarować kandydatów; deklaracja użytkownika jest oznaczona jako taka i podlega weryfikacji.
2. Przyjmować konfigurację builda jako dodatkowy dowód. Nie zakładać, że wynikowy JAR zawsze zawiera uniwersalne metadane relokacji.
3. Dla jawnej relokacji przechowywać obie tożsamości: oryginalną klasę komponentu i klasę występującą w artefakcie.
4. Przepisywać nazwy klas również wewnątrz deskryptorów parametrów i typów wynikowych, w tym tablic i klas wewnętrznych. Unikać kolizji i częściowych dopasowań prefiksów.
5. Weryfikować, czy przekształcone selektory odpowiadają faktycznym klasom i metodom przypisanym do komponentu.
6. Dodać przypadek biblioteki po minimalizacji: część klas została usunięta. Brak klasy nie oznacza automatycznie braku całej biblioteki ani bezpieczeństwa pozostałej zawartości.
7. Automatyczne odkrywanie nieznanej relokacji na podstawie znormalizowanych fingerprintów realizować dopiero jako osobny etap eksperymentalny, z reprezentatywnym benchmarkiem. Podobieństwo kodu bez rozstrzygnięcia wersji pozostaje heurystyką.

### Kryteria ukończenia

- Relokowane klasy i deskryptory prowadzą do właściwych metod w grafie.
- Raport pokazuje nazwy oryginalne i rzeczywiste oraz podstawę relokacji.
- Błędna mapa, kolizja pakietów, niezgodny hash i niejednoznaczna wersja nie dają potwierdzonego mapowania.
- Zakres wydania wyraźnie obejmuje jawną relokację, a nie wszystkie możliwe shaded JAR-y.

## F8. Docelowy runtime i pierwszy dostawca JDK

### Cel

Rozpoznawać podatności konkretnej dystrybucji i wydania runtime’u bez wnioskowania z samego bytecode’u.

### F8a. Identyfikacja i kontrakt źródła

1. Dodać `internal/runtimeinfo` z proponowanymi wejściami `-jdk-home` i `-runtime-file`; w MVP wzajemnie wykluczającymi się.
2. Czytać dostępne metadane instalacji i deklarowany opis. Nie uruchamiać automatycznie docelowego `java`; nie wybierać domyślnie runtime’u hosta skanera.
3. Zachować pełne surowe wartości i ostrożnie znormalizować vendor, dystrybucję, linię wersji, update, build, system, architekturę oraz ewentualny pakiet systemowy.
4. Nie utożsamiać ogólnego `IMPLEMENTOR` z jednoznacznie rozpoznanym produktem, jeśli dane nie wystarczają. Brakujące wymagane pola oznaczają nierozstrzygniętą ocenę.
5. Wybrać pierwszego dostawcę na podstawie docelowych wdrożeń. Oracle jest możliwym pilotażem, ale nie jest narzuconym wyborem użytkownika.
6. Przed budową adaptera zbadać dostępność, strukturę, sposób pobierania, warunki użycia oraz historyczne pokrycie danych. Jeśli brak stabilnego formatu, zacząć od wersjonowanego katalogu wpisów z odnośnikami do oficjalnych advisory.

### F8b. Ocena wydania

1. Dodać interfejs `RuntimeAdvisoryProvider` i jeden adapter, np. w `internal/runtimevuln`.
2. W rekordzie zachować produkt, dystrybucję, zakresy/wyliczone wydania, ograniczenia platformy, źródło, daty, komponent JDK i informacje o poprawce, jeśli źródło je podaje.
3. Wykorzystać historię advisory w zadeklarowanym zakresie, nie tylko najnowszy biuletyn. Lista wspieranych wersji w jednym advisory nie stanowi automatycznie pełnego przedziału wszystkich starszych wydań.
4. Obsługiwać backporty za pomocą danych właściwego dostawcy lub dystrybucji. Nie przenosić numeru poprawionego wydania jednego produktu na inne produkty.
5. Raportować dopasowanie wydania i osiągalność niezależnie. Dla znanego podatnego wydania bez modelu kodu: `applicability=affected`, `reachability=not_assessed` i powód; to nie jest `NOT_REACHABLE`.
6. Przy braku opisu runtime’u wykazać `runtime assessment: not performed`. Przy nieobsługiwanym dostawcy wykazać brak pokrycia zamiast pustej listy jako dowodu bezpieczeństwa.
7. Podać zakres adaptera: dostawca, obsługiwane linie, okres danych, platformy, braki i datę walidacji.
8. Stosować ten sam mechanizm cache, świeżości i snapshotów co dla pozostałych źródeł.

### F8c. Osiągalność kodu JDK — dalszy przyrost

1. Tylko dla przypadków z wiarygodnym mapowaniem na kod zaprojektować odczyt właściwych klas runtime’u i model granic analizy.
2. Nie znosić globalnie ograniczeń CHA dla `java.*`; wymagałoby to osobnej oceny rozmiaru grafu i liczby fałszywych ścieżek.
3. Podatności JIT, VM i kodu natywnego pozostawić bez werdyktu metodowego, jeśli model analizy ich nie obejmuje.

F8c nie jest warunkiem pierwszego wydania oceny podatności JDK.

### Kryteria ukończenia

- Ten sam poziom bytecode’u przy dwóch różnych runtime’ach nie prowadzi automatycznie do tych samych wyników podatności.
- Wydanie podatne i poprawione są rozróżniane na podstawie właściwego źródła.
- Nieznany vendor, niepełny build, brak danych historycznych i backport mają jawne rozstrzygnięcia lub brak rozstrzygnięcia.
- Pierwszy adapter jest udokumentowany i nie udaje wsparcia całego ekosystemu OpenJDK.

## F9. Integracja, migracja i kwalifikacja wydania

### Cel

Udostępnić spójny przepływ użytkownika i sprawdzić jakość całego przyrostu. Część zadań wykonywana jest przy każdym punkcie wydania.

### Zadania

1. Zaktualizować CLI, pomoc, README, przykłady mapowań, format JSON, sekcje raportu i opis ograniczeń.
2. Raportować osobno: tożsamość komponentów, zastosowanie advisory do wersji, jakość mapowania, osiągalność, świeżość danych oraz zakres nieoceniony.
3. Przeprowadzić migrację części osadzonych ręcznych mapowań do v2 z udokumentowaną weryfikacją. Pozostałe zachować jako legacy; nie deklarować ich kompletności przez samą zmianę formatu.
4. Sprawdzić eksport v1 dla wielu instancji i niepewności. Nie zmieniać domyślnego formatu JSON bez ogłoszonej polityki migracji.
5. Dla tego samego snapshotu sprawdzić stabilność identyfikatorów, kolejności wyników i wybranych ścieżek. Pola czasu prezentacji nie są podstawą porównania merytorycznej deterministyczności.
6. Uruchomić odpowiednie testy pakietów, następnie `go test ./...` i `go vet ./...` dla wydawanego przyrostu. Integracje sieciowe oprzeć na nagranych lub syntetycznych odpowiedziach; testy live wykonywać osobno, bez domyślnego LLM.
7. Porównać rzeczywiste aplikacje kontrolne na identycznym snapshotcie. Każdą zmianę statusu uzasadnić zmianą danych, rozpoznania komponentu, mapowania lub grafu. Nie wymagać zachowania znanego błędnego wyniku baseline.
8. Sprawdzić błędy danych: uszkodzony snapshot, niezgodny hash, nieobsługiwany schemat, niedostępne źródło, rate limit i częściowe mapowanie. Globalny błąd nie może zostawić nowego raportu wyglądającego na kompletny.
9. Zapisywać artefakty wyniku atomowo i spinać je wspólnym identyfikatorem analizy, aby nie mieszać raportu Markdown z JSON-em z innego uruchomienia.

### Kryteria wydania

- Każdy wydawany scenariusz kontrolny ma poprawny i wyjaśniony wynik.
- Niepewność nie jest ukrywana jako bezpieczeństwo ani jako brak CVE.
- Brak niezamierzonych zapytań sieciowych offline i niezamierzonych wywołań LLM.
- Można wskazać dowody i dane wejściowe każdego nowego rozstrzygnięcia.
- Dokumentacja wyraźnie wymienia funkcje działające, eksperymentalne i nieobsługiwane.

## 4. Proponowany kontrakt schematów

### Mapowanie v2

Minimalne pola rekordu:

```text
advisory_id, aliases
package: ecosystem, name
version_constraints
methods: class, name, descriptor?, role
evidence: kind, references, revision?, path?, side?, lines?, content_hash?
coverage: complete_under_declared_assumptions | partial | unknown
review: reviewed | unreviewed
```

Kompletność wymaga jawnego uzasadnienia i zakresu. Automatyczne znalezienie wszystkich metod zmienionych w commicie nie wystarcza do ustawienia `complete_under_declared_assumptions`.

### Wynik v2

Minimalne pola ustalenia:

```text
advisory_id, component_instance_id
identity: state, evidence_ids
applicability: affected | not_affected | undetermined
mapping: match_precision, evidence_kind, coverage, review
reachability: REACHABLE | FRAMEWORK | NOT_REACHABLE | UNKNOWN |
              NOT_APPLICABLE | not_assessed
reason_codes, evidence_ids, path?, entry?
```

Nazwy stanów należy ustabilizować w F0. `not_assessed` jest proponowaną nową wartością v2, a nie istniejącym statusem CLI. `NOT_APPLICABLE` zachowuje dotychczasowe znaczenie WebJar-ów; poprawione wydanie JDK jest opisywane przez `applicability=not_affected`, a nie przez ten status.

Wyniki bez advisory, np. nierozpoznany runtime lub nieznana wersja biblioteki, zapisuje się w sekcji pokrycia. Nie tworzy się dla nich fikcyjnych identyfikatorów CVE.

## 5. Minimalna macierz scenariuszy weryfikacyjnych

| Obszar | Przypadek | Oczekiwany rezultat |
|---|---|---|
| Wersje | CVE ma różne mapowania dla dwóch gałęzi biblioteki. | Wybór mapowania właściwego dla instancji. |
| Wersje | Granice przedziału, kwalifikatory i nieznana wersja. | Poprawne granice albo jawny brak rozstrzygnięcia. |
| Sygnatury | Podatne `parse(String)`, aplikacja wywołuje `parse(InputStream)`. | Brak fałszywej krawędzi wyłącznie od wspólnej nazwy. |
| Dziedziczenie | Wywołana odziedziczona metoda prowadzi dalej do podatnego kodu. | Graf kontynuuje właściwe ciało metody. |
| Pochodzenie | Ta sama klasa w dwóch JAR-ach, różne wersje. | Stabilny wybór i jawny konflikt; brak pożyczania ścieżek między kopiami. |
| Mapowanie | Metoda istnieje, ale dowód jest wyłącznie sugestią LLM. | Obecność bytecode nie podnosi jakości dowodu do potwierdzonego. |
| Mapowanie | Brak metody, częściowa lista lub konflikt źródeł. | Brak automatycznego `NOT_REACHABLE`. |
| Diff | Usunięcie metody lub całego pliku. | Kandydat pochodzi ze starej strony zmiany. |
| Diff | Dodany helper zabezpieczający. | Brak automatycznego utożsamienia z podatnym kodem starej wersji. |
| Diff | Rename, dwie gałęzie, merge lub niekompletny patch. | Właściwe rewizje albo jawna niepewność. |
| Shading | Metadane biblioteki bez jej klas. | Brak potwierdzonej obecności. |
| Shading | Osadzone dwie wersje, jedna podatna. | Osobne ustalenia i poprawne pochodzenie kodu. |
| Relokacja | Nazwy klas w parametrach, tablicach i wyniku. | Spójna transformacja deskryptorów. |
| Relokacja | Kolizja prefiksów, zła mapa albo minimalizacja. | Brak arbitralnego potwierdzenia mapowania. |
| Cache | Nowe CVE po wcześniej pustym querybatch. | Nowy rekord po odświeżeniu. |
| Cache | Wycofany rekord i aktywny alias. | Ponowna ocena źródeł, bez usunięcia aktywnego dowodu. |
| Cache | Awaria odświeżenia i stare dane. | Błąd albo jawny tryb stale, zgodnie z opcjami. |
| Snapshot | Offline, niezgodny hash lub brak treści. | Odtworzenie zgodnego wejścia albo jednoznaczny błąd. |
| Runtime | Wydanie podatne i poprawione tego samego dostawcy. | Różne applicability zgodnie ze źródłem. |
| Runtime | Backport, nieznany dostawca lub niepełny build. | Rozstrzygnięcie tylko przy wystarczających danych. |
| Runtime | Podatność w JIT/VM bez modelu osiągalności. | Podatne wydanie z `not_assessed`, bez `NOT_REACHABLE`. |

## 6. Decyzje do rozstrzygnięcia w odpowiedniej fazie

| Decyzja | Domyślna propozycja | Kiedy potrzebna |
|---|---|---|
| Nowy JSON | Format v2 opt-in; v1 z opisaną agregacją. | F0–F1. |
| Ważność danych CVE | 24 godziny, konfigurowalne. | F2. |
| Komparator wersji Maven | Go, zgodność sprawdzona na przypadkach referencyjnych; decyzja o bibliotece po przeglądzie dostępności i licencji. | F4. |
| Pierwszy SBOM | Jeden format i jawny obsługiwany zakres. | F6. |
| Pierwszy dostawca JDK | Wybrać według rzeczywistych wdrożeń; Oracle możliwym pilotażem. | F8a. |
| Heurystyczna relokacja | Osobny eksperyment po obsłudze jawnej mapy. | Po F7. |
| Osiągalność kodu JDK | Osobny przyrost; nie blokuje oceny wydań. | Po F8b. |

Decyzje te nie blokują rozpoczęcia F0 i projektu wspólnego modelu. Wybór dostawcy JDK wymaga informacji o faktycznym celu użycia przed implementacją jego adaptera.

## 7. Źródła i granice zobowiązania

- [Semantyka wersji Maven](https://maven.apache.org/pom.html#dependency-version-requirement-specification).
- [Schemat OSV: zakresy, daty, wycofanie](https://ossf.github.io/osv-schema/).
- [GitHub: dane commitów i plików](https://docs.github.com/en/rest/commits/commits#get-a-commit).
- [Maven Shade: relokacja klas](https://maven.apache.org/plugins/maven-shade-plugin/examples/class-relocation.html).
- [Przykładowy advisory Oracle](https://www.oracle.com/security-alerts/cpujul2026.html).
- [Red Hat: znaczenie backportów](https://www.redhat.com/en/blog/what-backporting-and-how-does-it-apply-rhel-and-other-red-hat-products).

Dokument nie obiecuje pełnego rozpoznawania dowolnych shaded JAR-ów, dowolnego runtime’u ani wszystkich metod związanych z każdym CVE. Terminy, koszty i procent pokrycia wymagają pomiarów. W tej zmianie zapisano dokumentację; implementacja i wykonanie opisanych testów należą do przyszłych faz.
