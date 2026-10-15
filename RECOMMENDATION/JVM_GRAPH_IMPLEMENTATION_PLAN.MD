# Podstawy grafu JVM — fazy implementacji G0–G5

Status: plan. Wymagania: [JVM_GRAPH.MD](JVM_GRAPH.MD). Kontrakty B0–B4: [ROADMAP.MD](ROADMAP.MD).

## G0. Kontrakt miejsc i krawędzi

Zależności: B0–B1 uzgodnione z CVE F0–F1.

Zadania:

- Rozszerzyć dane `Call` o offset instrukcji i pochodzenie; zachować symboliczny cel z bytecode oraz osobno cel rozwiązany.
- W `internal/callgraph/graph.go` zachować odwołania do dowodów krawędzi i rozróżnić `calls`/`initializes`; zapewnić adapter do dotychczasowego BFS.
- Wspólnie z `internal/reach` określić zapis rodzaju krawędzi w rekonstruowanej ścieżce. Deduplikacja nie usuwa alternatywnych dowodów.

Gotowe: krok raportu odsyła do konkretnego miejsca wywołania i reguły rozwiązania; kolejność wyników jest stabilna.

## G1. Pochodzenie i deterministyczny wybór klas

Zależności: G0, model komponentu CVE F1. To jest jedna realizacja wspólnych zadań, nie drugi refaktor tych samych struktur.

Zadania:

- W `internal/hierarchy/build.go` indeksować wyniki odczytu według jawnej kolejności wejścia i scalać po odczycie, nie według szybkości workerów.
- Zachować listę duplikatów, ich hashe i przypisanie do artefaktów. W `input`/CLI przyjąć możliwą do odtworzenia konfigurację classpath.
- Oznaczać przyjętą kolejność jako założenie, jeśli nie pochodzi z faktycznego uruchomienia. Dla nieobsługiwanej izolacji loaderów zachować konflikt w pokryciu.

Gotowe: permutacja czasu workerów nie zmienia grafu; zmiana jawnej kolejności zmienia wybraną implementację w przewidywalny sposób.

## G2. Dokładne rozwiązywanie wywołań

Zależności: G1, wspólny selektor metody CVE F4.

Zadania:

- W `internal/hierarchy/resolve.go` dodać lookup z nazwą, deskryptorem, rodzajem wywołania i informacjami o klasie wywołującej tam, gdzie wymaga tego JVM.
- Rozróżnić rozwiązywanie symbolu od możliwego wyboru implementacji przy dispatch. Uwzględnić interfejsy, nadklasy, konstruktory, `super`, private/static i bridge methods.
- W `callgraph/build.go` prowadzić wywołanie do dostępnego ciała implementacji; rozszerzać CHA zgodnie z deskryptorem.
- Zaktualizować `reach/match.go` i grounding LLM z CVE F4, zachowując jawny tryb przybliżonego dopasowania legacy.

Gotowe: fixtures obejmują odziedziczoną metodę wołającą bibliotekę, przeciążenia, override, default method, wywołanie super oraz bridge. Nieprawidłowy lub brakujący cel daje diagnostykę.

## G3. Relacje inicjalizacji

Zależności: G2; część metadanych instrukcji i pól współdzielona z DATA_FLOW D1.

Zadania:

- W `classfile` zachować odwołania `new`, `getstatic`, `putstatic`, `invokestatic` i dane pozwalające rozwiązać klasę deklarującą pole/metodę.
- Opracować model inicjalizacji według odpowiednich reguł JVM, łącznie z wymaganymi nadklasami i przypadkami interfejsów.
- Dodawać krawędzie do istniejącego `<clinit>` tylko dla właściwych operacji. Samo posiadanie literału klasy nie staje się regułą inicjalizacji.
- Zapisać w korzeniach analizowanego scenariusza wymagane inicjalizacje entry class; tryb `roots=all` zachowuje konserwatywne założenia.
- Model inicjalizacji udostępnić refleksji i uchwytom jako wspólną funkcję, zamiast kopiować logikę w każdym analizatorze.

Gotowe: ścieżka do wywołania w zależnościowym `<clinit>` istnieje tam, gdzie jest uzasadniona; testy obejmują odziedziczone pole statyczne, brak `<clinit>`, cykl i przypadki bez inicjalizacji. Graf nie obiecuje dokładnej kolejności runtime’u.

## G4. Target Java i multi-release JAR

Zależności: G1; wspólny opis targetu z CVE F8a, jeśli jest już dostępny. Pełny adapter podatności JDK nie jest wymagany.

Zadania:

- Dodać jawny target feature version, np. proponowane `-target-java`; uzgodnić go z metadanymi runtime’u, jeśli podano oba wejścia. Sprzeczność jest błędem.
- Odczytać manifest `Multi-Release` i wybrać właściwe wpisy według reguł JAR. Nie traktować każdego `META-INF/versions/` jako aktywnego wariantu.
- Zachować faktyczną ścieżkę wybranej klasy oraz wersję wariantu w pochodzeniu. Obsłużyć wyłącznie poprawne lokalizacje i raportować niezgodne wejścia.
- Dla niepodanego targetu zachować udokumentowany fallback do bazowych klas z widoczną niepewnością, zamiast odgadywać runtime z najwyższej wersji bytecode’u.

Gotowe: targety bazowy/11/17 wybierają oczekiwane klasy; brak flagi manifestu, brak wariantu i sprzeczne wejście mają oczekiwane wyniki. Snapshot zachowuje wybór.

## G5. Integracja i kwalifikacja

Zależności: G0–G4 w wydawanym zakresie.

- Wprowadzić fixtures do `testdata`, wraz ze źródłami i instrukcją regeneracji. Testy parsera/grafu działają na zapisanych JAR-ach.
- Porównać graf przed/po, uzasadnić zmienione ścieżki i sprawdzić koszt pamięci przechowywania dowodów.
- Sprawdzić właściwe pakiety i komplet wymaganych kontroli według B4; uaktualnić raport, README i ograniczenia.
- Nie ogłaszać, że rozwiązano refleksję, callbacki i JNI przez samo poprawienie dispatch.

Pierwszy zakres: G0–G3. G4 warto dostarczyć przed analizą nowych aplikacji z multi-release zależnościami; do tego czasu ograniczenie musi pozostać widoczne.
