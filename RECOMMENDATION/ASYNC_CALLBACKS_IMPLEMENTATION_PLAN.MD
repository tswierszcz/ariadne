# Callbacki i asynchroniczność — fazy A0–A4

Status: plan. Wymagania: [ASYNC_CALLBACKS.MD](ASYNC_CALLBACKS.MD). Kolejność: [ROADMAP.MD](ROADMAP.MD).

## A0. Katalog modeli callbacków

Zależności: B0–B3, JVM G0–G2, DATA_FLOW D0–D2.

- Zdefiniować model: sygnatura rejestracji, pozycja argumentu callbacka, jego metoda, mapowanie parametrów, warunek oraz charakter wykonania.
- Wybrać początkowy katalog JDK i konkretne overloads; rejestr YAML jest deklaratywny i walidowany.
- Zaprojektować węzeł/relację zdarzenia z pochodzeniem, która nie udaje wywołania synchronicznego.
- Dodać diagnostykę brakującego celu, unknown receivera i nieobsługiwanego API.

Gotowe: katalog ma dodatni i ujemny przykład każdego pierwszego API oraz określony sposób rekonstruowania ścieżki.

## A1. Tożsamość zadania i funkcji

- Odczytać argumenty z DATA_FLOW, powiązać instancje klas `Runnable`/`Callable` oraz wartości funkcyjne z DYNAMIC_LINKAGE L1.
- Dla `Thread` rozdzielić przekazane zadanie i override `run` według właściwego modelu. Nie wyprowadzać związku z samej wspólnej nazwy metody.
- Wspólne captured values są częścią obiektu funkcyjnego; dwa miejsca utworzenia mogą nieść różne dane.
- Nieznany obiekt może dać konserwatywny zbiór celów z oznaczeniem podstawy; nie wolno uznać go za pełny, jeśli analiza go przycięła.

Gotowe: dwa zadania i niewykorzystana lambda nie tworzą nieuzasadnionych połączeń.

## A2. Podstawowe wykonanie i ścieżki

- Modele wybranych `start`/`execute`/`submit` tworzą możliwe przejścia do właściwych celów i znanych argumentów.
- Obsłużyć fakt, że implementacja Executor może wykonać zadanie bez nowego wątku; raport nie wymyśla thread identity w analizie statycznej.
- Zachować przejścia błędu/rejection w przyjętym zakresie, a ewentualne wykonanie oznaczać jako możliwe i warunkowe.
- Rozszerzyć dane ścieżki w `internal/reach`/`report`, tak aby model callbacka miał miejsce i dowód. Do wyników trafiają też nieobsługiwane miejsca.

Gotowe: ścieżka do wywołania w `run`/`call`, rejestracja bez wykonania i nieznany target mają oczekiwane wyniki. A0–A2 to MVP.

## A3. Kontynuacje, listenery i reactive

Zależności: DATA_FLOW D3, FRAMEWORK F2/F4 dla framework listeners.

- Dodać wybrane CompletionStage/future operacje: rezultat poprzednika, funkcję, warunek normalnego/błędnego zakończenia i zakres propagacji wyjątków.
- Dla listenerów wiązać rejestrację z modelem źródła zdarzenia. Sama metoda `addListener` bez znajomości kontraktu nie tworzy dowolnego zdarzenia.
- Wybrać jedną bibliotekę reactive i wersję; opisać warunki subskrypcji i wykonywania wybranego katalogu operatorów. Nie utożsamiać wszystkich streamów z jednym kontraktem.
- Przekazać mapowanie argumentów do taint D4; brak modelu kolejki lub deserializacji nie jest dowodem zachowania danych.
- Budżet liczby kontynuacji i wielkości pipeline ma jawny stan niekompletności.

Gotowe: dwie kontynuacje, exceptional branch, nieaktywny pipeline i znana subskrypcja są rozróżnione na wersjonowanych fixtures.

**Wycinek 2 (`a3-2`, 2026-11-19):** callback będący parametrem metody rejestrującej rozwiązywany u jej
wywołujących (3 poziomy, 32 miejsca), krawędź „może wykonać” z miejsca rejestracji; fixture
`testdata/async-forward` z wyrocznią JVM; motywacja i pomiar: jetty-client (LIBRARY_MODE LB4).

**Wycinek 3 (`a3-3`, 2026-11-19):** przekazania przez kolejkę w polu (magazyn → metoda opróżniająca,
krawędź od miejsca, gdzie obiekt wchodzi, do metody wołanej na nim przez drenaż; warunek `queued`) i
przez załącznik klucza NIO (`register(…, att)`/`attach` → pętla selektora; warunek `selected`); fixture
`testdata/async-handoff` z wyrocznią JVM; wyrocznia krawędzi `scripts/MethodEdgeTrace.java`; pomiar:
jetty-client.

**Wycinek 4 (`a3-4`, 2026-11-20):** zadania zwracane z metody — miejsce `Runnable#run`/`Callable#call` albo
rejestracji A2, którego zadanie pochodzi z wartości zwracanej (wywołanie dokładne albo implementacje
wywołania wirtualnego z typem odbiorcy zawężonym przez pole/parametr, pola final/static, `new`/lambda);
krawędź „może wykonać” z warunkiem `produced`, bramkowana (creator gate) najgłębszą metodą zwracającą
zadanie; fixture `testdata/async-returned` z wyrocznią JVM; pomiar: jetty-client (`EatWhatYouKill` →
`ChannelEndPoint$1#run`).

**Wycinek 5 (`a3-5`, 2026-11-20):** listenery — rejestracja powiązana z modelem źródła zdarzenia: lista
listenerów w polu (magazyn a3-3 + *rozgłoszenie*: obejście pola bez zdejmowania i czyszczenia) oraz
`PropertyChangeSupport` w polu (`addPropertyChangeListener` + `fire*` na tym samym polu); krawędź z wywołania
powiadamiającego do metody listenera (warunek `notified`), bramkowana metodą rejestrującą; sama rejestracja
bez powiadomienia nie daje krawędzi; fixture `testdata/async-listeners` z wyrocznią JVM.

**Wycinek 6 (`a3-6`, 2026-11-20):** reactive — jedna biblioteka i wersja: JDK `Flow` z `SubmissionPublisher`
(Java 9+, sprawdzane na JDK 25; Reactor/RxJava nie ma w lokalnym repozytorium, a Central blokuje pobieranie).
Katalog: `subscribe` → `onSubscribe`; `submit`/`offer` → `onNext` + `onError`; `close` → `onComplete`;
`closeExceptionally` → `onError`; `consume` → `Consumer#accept`; parowanie po obiekcie wydawcy (miejsce `new`,
`this` w procesorze); bramka: metoda subskrybująca. Zbudowany pipeline bez emisji i znana subskrypcja
rozróżnione na fixture `testdata/async-reactive` z wyrocznią JVM.

**Wycinek 7 (`a3-7`, 2026-11-20):** zakończenie future — dla `new CompletableFuture<>()` kończonego ręcznie
kontynuacje są bramkowane wywołaniami kończącymi właściwego rodzaju na tym future (parowanie po miejscu
`new`), a future, którego nic nie kończy (albo kończy tylko błędem dla `then*`), nie daje krawędzi; inne etapy
bez zmian; fixture `testdata/async-completion` z wyrocznią JVM.

## A4. Weryfikacja i kwalifikacja

- Fixtures: direct executor, thread override, task object, lambda, constructor, dwa tasks, rejection, future success/failure, listener i wybrany reactive pipeline.
- Przy kontrolowanych uruchomieniach porównać cele i zdarzenia, bez wnioskowania o niemożliwości niezaobserwowanych scenariuszy.
- Zmierzyć liczbę dodanych krawędzi i błędnych skojarzeń. Nie znosić globalnie skip dla `java.*` tylko po to, by fixtures stały się dodatnie.
- Wykonać B4, zapisać modele/scenariusz w snapshotach i udokumentować ograniczenia współbieżności.
