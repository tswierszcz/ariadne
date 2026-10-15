# Runtime i native — fazy N0–N5

Status: plan. Wymagania: [RUNTIME_NATIVE.MD](RUNTIME_NATIVE.MD). Zasady: [ROADMAP.MD](ROADMAP.MD).

## N0. Kontrakt śladu i zakres dowodów

Zależności: B0–B1, wspólne pochodzenie CVE F1; snapshoty CVE F3 przed deklaracją odtwarzalności.

- Zaprojektować wersjonowany format: run id, scenariusz, runtime, narzędzie rejestrujące i jego wersja, zakres pomiaru, class/loader/method ids oraz hashe zawartości.
- Wyróżnić zdarzenia loaded class, method entry, observed call oraz przekazanie zadania. Pole call site/parent event musi wynikać z pomiaru, nie zgadywania po kolejności czasowej.
- Zapisać sampling, utratę zdarzeń, ograniczenia instrumentacji, zakres czasowy i przyczyny niepełności.
- Zdefiniować dane minimalne bez argumentów i payloadów. Ślady oraz opcjonalny bytecode są osobnymi artefaktami z hashami i kontrolą zgodności.

Gotowe: można opisać wiarygodność pojedynczego zdarzenia i całego przebiegu, także przy częściowym śladzie.

## N1. Importer offline i korelacja

- Utworzyć proponowany `internal/runtimeevents` z walidacją formatu, typów, odwołań i limitów danych wejściowych.
- Powiązać metodę z komponentem/bytecode/deskryptorem. Niezgodny hash, nieznany loader lub brak klasy daje brak dopasowania, a nie arbitralne przypięcie do znanego FQCN.
- Pokazać zaobserwowane metody i relacje w raporcie bez nadpisywania statycznego grafu. Załadowanie klasy nie oznacza wykonania wszystkich jej metod.
- Zachować id dowodu w wynikach i snapshotach, np. proponowane wejście `-runtime-trace`.

Gotowe: poprawny trace, inna wersja artefaktu, dwa loadery i niepełny trace mają odrębne wyniki. N0–N1 to pierwszy zakres runtime, bez konieczności tworzenia agenta.

## N2. Dodatkowe klasy i kontekst ładowania

Zależności: N1, JVM G1/G4, CVE F1/F6 dla pochodzenia komponentów.

- Wprowadzić jawne dołączenie plugin JAR-ów i deklaracji kontekstu loadera. Obecna mapa FQCN pozostaje adapterem scenariusza jednego kontekstu, nie uniwersalną tożsamością.
- Zidentyfikować relacje loaderów, wybrane reguły delegacji i widoczność klas. Nieznanego custom loadera nie normalizować po cichu do parent-first.
- Dla klas przechwyconych w runtime zachować hash i pochodzenie wersji przed/po transformacji. Nie zakładać dostępności bytecode każdej hidden class.
- Generowana klasa bez Maven identity może mieć osobne pochodzenie i powiązania z interfejsem/targetem; podobieństwo nazwy nie potwierdza przypisania CVE.
- Import nie pozwala dowodom z jednego kontekstu tworzyć ścieżek w innym bez jawnego powiązania.

Gotowe: dwie definicje tej samej klasy pozostają odrębne, a znany plugin ma warunkowe/zaobserwowane użycie zgodnie z dowodami.

## N3. Pierwszy opcjonalny agent Java

- Utworzyć osobny artefakt, np. katalog `agent/`, ze wskazaną wersją JDK i procesem budowania. Czysty analizator Go nie zależy od jego obecności.
- Wybrać mechanizm instrumentacji i ewentualną bibliotekę dopiero po sprawdzeniu możliwości, licencji i zakresu wspieranych klas/JDK. Samo java.lang.instrument nie daje automatycznie gotowego call trace.
- MVP uruchamia się przez świadomie wskazany przez użytkownika scenariusz startowy z agentem; dynamic attach jest osobnym rozszerzeniem, nie warunkiem wydania.
- Instrumentować ograniczony katalog metod z kontrolą reentrancy, błędów i wyjątków. Zdarzenia mają stabilne id w przebiegu oraz odwołanie do konkretnej wersji klasy.
- Zastosować buforowanie, limity i licznik utraty. Przy zakończeniu niepełny bufor pozostaje oznaczony; brak eventu nie jest dowodem braku wywołania.
- Zweryfikować wpływ transformacji na weryfikację bytecode, moduły i inicjalizację; jawnie wykluczyć obszary bez pokrycia. Native methods bez Java body wymagają innych mechanizmów.

Gotowe: agent w kontrolowanej aplikacji rejestruje znane metody/wywołania w swoim zakresie i nie ukrywa błędów instrumentacji. Integracje agenta są oddzielne od zwykłych testów Go.

## N4. Refleksja, proxy i asynchroniczna przyczynowość

Zależności: N3, modele odpowiednich API z REFLECTION/FRAMEWORKS/ASYNC.

- W wybranych miejscach rejestrować rzeczywiste cele refleksji i proxy z kontekstem loadera; wykorzystać je jako dowody dla tego przebiegu.
- Dla asynchronicznych zadań przekazywać token rejestracji do wykonania tylko tam, gdzie instrumentacja rzeczywiście zapewnia korelację. Sama bliskość timestampów nie wystarcza.
- Dopasować observed edges do modelowanych krawędzi i raportować różnicę: potwierdzony cel, brakująca statyczna krawędź lub niedopasowany kontekst.
- Nie usuwać statycznych możliwości na podstawie pojedynczego testowego przebiegu. Nie traktować method entry jako potwierdzonego wpływu danych atakującego.

Gotowe: dwa zadania i dwa proxy nie mieszają dowodów; brak korelacji skutkuje obserwacją metody bez wymyślonej relacji caller → callee.

## N5. Native/JNI i kwalifikacja

- W parserze i hierarchii rozpoznać native methods; rejestrować `load`/`loadLibrary` jako informacje o bibliotece, z rozróżnieniem próby i powodzenia, jeśli dane to umożliwiają.
- Zaprojektować jawne streszczenia Java ↔ native: sygnatury, identyfikator biblioteki/builda, przepływ argumentów i callbacki Java. Nie zakładać, że nazwa eksportu identyfikuje wszystkie dynamiczne `RegisterNatives`.
- Bez streszczenia wskazać nieprześledzoną granicę i wpływ na wynik. Analiza C/C++, JVMTI lub Foreign Function & Memory wymaga osobnego zakresu i nie jest ukrytym elementem agenta Java.
- Fixtures obejmują nieznaną native metodę, modelowany callback, ładowanie klasy, dwie definicje, retransformation/niezgodny hash, utratę zdarzeń i async bez korelacji.
- Wykonać B4 dla importera, osobne testy integracyjne agenta i pomiar narzutu. Dokumentować platformy i JDK, a nie deklarować uniwersalnego tracingu runtime’u.

Pierwszy tor JNI może zakończyć się na jawnej inwentaryzacji granic i jednym poprawnym streszczeniu. Rozbudowa native jest uzależniona od potrzeb i pomiarów.
