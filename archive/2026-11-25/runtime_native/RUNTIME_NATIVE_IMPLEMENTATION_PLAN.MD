# Runtime i native — fazy N0–N5

Status: **Obszar DONE for this scope (2026-11-25), zarchiwizowany** — zob. [COMPLETION.md](COMPLETION.md). Stan faz przed zamknięciem: N0, N1 DONE (2026-11-08, zarchiwizowane: [archive/2026-11-08/runtime_native/RUNTIME_NATIVE_N0_N1.md](../../../archive/2026-11-08/runtime_native/RUNTIME_NATIVE_N0_N1.md); kontrakt `ariadne-runtime-trace-1`, importer `internal/runtimeevents` `n1-1`, `-runtime-trace`, rejestrator z JFR bez agenta); N2 odłożony (2026-11-23), N3 i N4 ocenione 2026-11-25 (agent niebudowany — patrz niżej), N5 wycinek `n5-1`; zob. IMPLEMENTATION_STATUS. Wymagania: [RUNTIME_NATIVE.MD](RUNTIME_NATIVE.MD). Zasady: [ROADMAP.MD](../../../RECOMMENDATION/ROADMAP.MD).

## N0. Kontrakt śladu i zakres dowodów — DONE (zarchiwizowane 2026-11-08)

## N1. Importer offline i korelacja — DONE (zarchiwizowane 2026-11-08)

Kontrakt `ariadne-runtime-trace-1` (`docs/RUNTIME_TRACE.MD`), importer `internal/runtimeevents`
(`-runtime-trace`), rejestrator z JFR bez agenta (`scripts/JfrToTrace.java`), a jako
rozszerzenie N1 — `-runtime-filter-out` (ustawienia JFR śledzące podatne metody ustaleń);
zobacz [archive/2026-11-08/runtime_native/RUNTIME_NATIVE_N0_N1.md](../../../archive/2026-11-08/runtime_native/RUNTIME_NATIVE_N0_N1.md)
po pełny zapis i `docs/SUPPORT.MD` po bieżący opis zachowania.

## N2. Dodatkowe klasy i kontekst ładowania

**Ocena 2026-11-23:** nie zaczęty; po stronie śladu kontrakt N0/N1 już rozróżnia loadery i definicje klas, po stronie statycznej jest jedna mapa FQCN. Żaden cel (aplikacje referencyjne, flota bibliotek) nie ma kilku kontekstów ładowania — odłożony do pierwszego takiego celu. Zob. IMPLEMENTATION_STATUS.

Zależności: N1, JVM G1/G4, CVE F1/F6 dla pochodzenia komponentów.

- Wprowadzić jawne dołączenie plugin JAR-ów i deklaracji kontekstu loadera. Obecna mapa FQCN pozostaje adapterem scenariusza jednego kontekstu, nie uniwersalną tożsamością.
- Zidentyfikować relacje loaderów, wybrane reguły delegacji i widoczność klas. Nieznanego custom loadera nie normalizować po cichu do parent-first.
- Dla klas przechwyconych w runtime zachować hash i pochodzenie wersji przed/po transformacji. Nie zakładać dostępności bytecode każdej hidden class.
- Generowana klasa bez Maven identity może mieć osobne pochodzenie i powiązania z interfejsem/targetem; podobieństwo nazwy nie potwierdza przypisania CVE.
- Import nie pozwala dowodom z jednego kontekstu tworzyć ścieżek w innym bez jawnego powiązania.

Gotowe: dwie definicje tej samej klasy pozostają odrębne, a znany plugin ma warunkowe/zaobserwowane użycie zgodnie z dowodami.

## N3. Pierwszy opcjonalny agent Java

**Ocena 2026-11-25: NOT STARTED — agent niebudowany, zastąpiony ścieżką JFR.** Cel MVP N3 (dokładne wejścia ograniczonego katalogu metod, ramka wywołującego, oznaczona utrata zdarzeń, brak zdarzenia ≠ brak wywołania) spełnia od N1 rejestrator bez agenta: JFR `jdk.MethodTrace` (JDK 25+, JEP 520) z filtrem metod, `scripts/JfrToTrace.java`, `-runtime-filter-out`. Pomiar [testdata/runtime-n4](../../../testdata/runtime-n4/RESULTS.MD): refleksja i proxy dają dokładne wejścia z relacją przez ramki runtime. Agent dałby ponad to tylko: (a) dokładne wejścia na JDK < 25, (b) korelację przez granicę wątku (token rejestracji — N4 async), (c) bajty klas przed/po transformacji i hidden classes (N2). Wznowić, gdy cel tego wymaga — badany przebieg na JDK < 25, eksperyment Tekmerion wymagający przypisania wykonania do rejestracji, cel z transformacją klas; wtedy agent jest kolejnym rejestratorem tego samego formatu `ariadne-runtime-trace-1`.

- Utworzyć osobny artefakt, np. katalog `agent/`, ze wskazaną wersją JDK i procesem budowania. Czysty analizator Go nie zależy od jego obecności.
- Wybrać mechanizm instrumentacji i ewentualną bibliotekę dopiero po sprawdzeniu możliwości, licencji i zakresu wspieranych klas/JDK. Samo java.lang.instrument nie daje automatycznie gotowego call trace.
- MVP uruchamia się przez świadomie wskazany przez użytkownika scenariusz startowy z agentem; dynamic attach jest osobnym rozszerzeniem, nie warunkiem wydania.
- Instrumentować ograniczony katalog metod z kontrolą reentrancy, błędów i wyjątków. Zdarzenia mają stabilne id w przebiegu oraz odwołanie do konkretnej wersji klasy.
- Zastosować buforowanie, limity i licznik utraty. Przy zakończeniu niepełny bufor pozostaje oznaczony; brak eventu nie jest dowodem braku wywołania.
- Zweryfikować wpływ transformacji na weryfikację bytecode, moduły i inicjalizację; jawnie wykluczyć obszary bez pokrycia. Native methods bez Java body wymagają innych mechanizmów.

Gotowe: agent w kontrolowanej aplikacji rejestruje znane metody/wywołania w swoim zakresie i nie ukrywa błędów instrumentacji. Integracje agenta są oddzielne od zwykłych testów Go.

## N4. Refleksja, proxy i asynchroniczna przyczynowość

**Ocena 2026-11-25: refleksja pokryta, proxy pokryte od `n4-1`, async odłożone.** Pomiar [testdata/runtime-n4](../../../testdata/runtime-n4/RESULTS.MD) na JDK 25:
- *Refleksja* — pokryta bez agenta: `main` (linia wywołania `invoke`) → rzeczywisty cel, `stack-through-runtime`, `traced`; przyczyna `reflection` w klasyfikacji luk `g2` już istnieje.
- *Proxy* — handler → cel pokryty; miejsce wywołania proxy → handler ginie w konwerterze (dla zdarzeń `traced` bierze tylko najbliższą ramkę z zakresu), choć JFR ma pełny stos. To poprawka konwertera, nie agent. **Wycinek `n4-1` DONE (2026-11-25):** `JfrToTrace` czyta cały stos zdarzenia `jdk.MethodTrace` (relacje `traced` między kolejnymi ramkami z zakresu, kluczowane bci/linią), a klasyfikator luk uznaje za `proxy` także handler-lambdę/referencję metody utworzoną dla `InvocationHandler#invoke`. Sonda: dwa wywołania proxy z linii 30 i 31 `main` to dwie relacje `proxy`; async bez wymyślonej relacji. `TestRuntimeN4Probe`.
- *Async* — wykonanie jest widoczne (`Task#run` → `work`), rejestracja nie: stos wątku roboczego jej nie zawiera, JFR nie przenosi tokenu. Wymaga agenta (N3 (b)); statycznie async jest modelowane (ASYNC_CALLBACKS A0–A4, wyrocznie JVM), a lambda z osobnego miejsca ma własne ciało, więc niejednoznaczne jest tylko to samo zadanie zgłoszone z kilku miejsc. Odłożone do konkretnej potrzeby.
- *Dopasowanie obserwacji do krawędzi* — istnieje od N1 (`not_in_graph` z przyczyną `g2`).

Zależności: N3, modele odpowiednich API z REFLECTION/FRAMEWORKS/ASYNC.

- W wybranych miejscach rejestrować rzeczywiste cele refleksji i proxy z kontekstem loadera; wykorzystać je jako dowody dla tego przebiegu.
- Dla asynchronicznych zadań przekazywać token rejestracji do wykonania tylko tam, gdzie instrumentacja rzeczywiście zapewnia korelację. Sama bliskość timestampów nie wystarcza.
- Dopasować observed edges do modelowanych krawędzi i raportować różnicę: potwierdzony cel, brakująca statyczna krawędź lub niedopasowany kontekst.
- Nie usuwać statycznych możliwości na podstawie pojedynczego testowego przebiegu. Nie traktować method entry jako potwierdzonego wpływu danych atakującego.

Gotowe: dwa zadania i dwa proxy nie mieszają dowodów; brak korelacji skutkuje obserwacją metody bez wymyślonej relacji caller → callee.

## N5. Native/JNI i kwalifikacja

**Wycinek `n5-1` DONE (2026-11-23):** inwentaryzacja granic — metody `native` z wejścia jako nieograniczone luki `native` (reguła B0), wywołania `loadLibrary`/`load` z nazwą biblioteki, gdy literał; report.md „Native code boundaries”, v2 `coverage.native`. Petclinic 72/361 osiągalnych, WebGoat 108/206, JPetStore 0/3; 0 zmian werdyktów. Streszczenia Java ↔ natywne — odłożone. Zob. IMPLEMENTATION_STATUS.

- W parserze i hierarchii rozpoznać native methods; rejestrować `load`/`loadLibrary` jako informacje o bibliotece, z rozróżnieniem próby i powodzenia, jeśli dane to umożliwiają.
- Zaprojektować jawne streszczenia Java ↔ native: sygnatury, identyfikator biblioteki/builda, przepływ argumentów i callbacki Java. Nie zakładać, że nazwa eksportu identyfikuje wszystkie dynamiczne `RegisterNatives`.
- Bez streszczenia wskazać nieprześledzoną granicę i wpływ na wynik. Analiza C/C++, JVMTI lub Foreign Function & Memory wymaga osobnego zakresu i nie jest ukrytym elementem agenta Java.
- Fixtures obejmują nieznaną native metodę, modelowany callback, ładowanie klasy, dwie definicje, retransformation/niezgodny hash, utratę zdarzeń i async bez korelacji.
- Wykonać B4 dla importera, osobne testy integracyjne agenta i pomiar narzutu. Dokumentować platformy i JDK, a nie deklarować uniwersalnego tracingu runtime’u.

Pierwszy tor JNI może zakończyć się na jawnej inwentaryzacji granic i jednym poprawnym streszczeniu. Rozbudowa native jest uzależniona od potrzeb i pomiarów.
