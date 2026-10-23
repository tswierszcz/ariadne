# Stan realizacji i przekazanie między sesjami

Ten plik jest rejestrem pracy nad nowymi planami, a nie deklaracją wsparcia produktu. Aktualne możliwości opisuje [SUPPORT.MD](../docs/SUPPORT.MD), zakres prac [ROADMAP.MD](ROADMAP.MD), a sposób prowadzenia sesji [skill](skills/ariadne-implement/SKILL.md).

## Stan początkowy

Przygotowano plany, prompt i skill. W ramach ich tworzenia nie rozpoczęto implementacji nowych faz. Istniejący kod może spełniać część kryteriów; każda kolejna sesja ma to zweryfikować zamiast zakładać, że wszystko trzeba napisać od początku.

Stany faz: `UNASSESSED`, `READY`, `IN_PROGRESS`, `NEEDS_VERIFICATION`, `DONE`, `BLOCKED`. `DONE` oznacza spełnienie wszystkich kryteriów konkretnej fazy oraz wymaganej weryfikacji i dokumentacji. `BLOCKED` wymaga konkretnego powodu i sposobu odblokowania. To etykiety tego pliku, nie stany celów w aplikacji Codex.

## Przegląd obszarów

| Obszar | Plan | Ocena nowych faz | Wznowienie |
|---|---|---|---|
| ROADMAP | [Kontrakty B0–B4 i przyrosty](ROADMAP.MD) | B3: sama decyzja architektoniczna + kontrakt (`internal/models`) zrobione; koordynator iteracji (`internal/analysis`) NIE; B0–B2, B4: UNASSESSED | Uzgodnić B0–B1 z CVE/JVM; nie traktować całego R0 jako ukończonego po samym projekcie. B3's własny kontrakt (`Model`/`Registry`/`Fact`, dopasowanie po dokładnym `Signature`+GAV, nie po nazwie) jest gotowy i udokumentowany jako wyłącznie kontrakt — zob. własny wpis fazy. |
| CVE | [F0–F9](CVE_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Sprawdzić wspólne wymagania F0–F1 i aktualny model wyników. |
| JVM_GRAPH | [G0–G5, zarchiwizowane](../archive/2026-10-22/jvm_graph/COMPLETION.md) | G0, G1, G3, G4, G5: DONE; G2: DONE for overloads/inheritance, bridge/default-method real-bytecode verification still pending (needs a JDK, absent here) | Cały JVM_GRAPH's R0-owy zakres (G0–G5) zamknięty, z jednym odnotowanym, czekającym na JDK ograniczeniem w G2. Następny krok to inny obszar R0/R1 z ROADMAP.MD, nie JVM_GRAPH. |
| DATA_FLOW | [D0–D5](DATA_FLOW_IMPLEMENTATION_PLAN.MD) | D0–D2: DONE; D3 (streszczenia pól statycznych): częściowe; D3 (reszta), D4–D5: UNASSESSED | D3's pierwszy wycinek (streszczenia zapisów pól statycznych między metodami) jest realny i przetestowany, ale **zmierzony efekt na petclinic/WebGoat jest zerowy** — rzeczywiste miejsca `MethodHandle.invoke` na obu aplikacjach używają pól instancyjnych albo wyniku innego wywołania, nie pól statycznych. Odnotowane szczerze, nie skryte. |
| SERVICE_LOADING | [S0–S4](SERVICE_LOADING_IMPLEMENTATION_PLAN.MD) | S0, S1, S2: DONE; S3–S4: UNASSESSED | S2 dodaje realną krawędź do konstruktora providera dla każdego rozwiązanego interfejsu (8/147 providerów, 10/158 krawędzi na petclinic/WebGoat), ale tylko sam konstruktor — iterator/`Provider.get` nietknięte, JPMS (S3) nietknięte. |
| REFLECTION | [R0–R4](REFLECTION_IMPLEMENTATION_PLAN.MD) | R0+R1+R2: DONE dla całego zadeklarowanego zakresu ("R0–R2 stanowią pierwsze wydanie"); R3–R4: UNASSESSED | `Class.forName("literał")` rozwiązuje się do `KindConst`/`java.lang.Class` (reużycie, nie nowy typ). `getMethod`/`getDeclaredMethod` z literalną nazwą i literalną tablicą `Class[]` rozwiązuje się do nowego `KindReflectiveMethod` — wymagało nowej maszynerii: `KindArray` (wskaźnikowy, mutowalny przez `aastore`, prawdziwa semantyka referencji JVM) i minimalnego śledzenia literałów `int` (wcześniej całkowicie poza zakresem D2). Zmierzony efekt R0+R1 na petclinic/WebGoat: **zero** (żadna z aplikacji nie używa tych API), uczciwie odnotowane; brak crasha na pełnej hierarchii (29k/40k klas), bajt-identyczny `results.json`. R2's pierwsza połowa: `internal/reflectinvoke` łączy każde miejsce `Method#invoke`, którego receiver rozwiązuje się do powyższego faktu, realną krawędzią grafu (`BasisReflection`) do realnego celu dyspozycji (typ alokacji, nie tylko `Class` z lookupu), z uczciwym `Gap` przy niezgodności liczby argumentów/nierozpoznanym typie parametru/braku deklaracji — **efekt niezerowy**, inaczej niż R0/R1: petclinic 3/59/29/51/6→3/62/26/51/6 (5 krawędzi), WebGoat 40/43/21/56/0→42/44/18/56/0 (3 krawędzie). R2's druga połowa (nowa): analogiczna maszyneria dla `getConstructor`/`getDeclaredConstructor` (`KindReflectiveConstructor`, reużywa R1's `KindArray`, bez nowego modelu tablicy) i `Constructor#newInstance` → realna krawędź do **dokładnie jednego** `<init>` (konstruktor nigdy nie jest dziedziczony — JVMS 5.4.3.3, więc żadnego overload-insensitive doboru jak przy `Method#invoke`), plus powiązana krawędź `<clinit>` tą samą parą Kind/Basis, jaką `ApplyServiceLoaderEdges` już stosuje dla SERVICE_LOADING S2. Zmierzone: realne krawędzie istnieją (petclinic 264 miejsc/6 rozwiązań/7 krawędzi, WebGoat 264/5/5), ale **zero dodatkowego efektu na reachability** poza tym, co dała już pierwsza połowa — liczby REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN identyczne przed/po na obu aplikacjach, uczciwie odnotowane (nowo-konstruowalne klasy nie leżą akurat na ścieżkach podatnych metod w tych dwóch aplikacjach). Dwa przebiegi offline na każdej dają bajt-identyczny `results.json`. |
| FRAMEWORKS | [F0–F5, F1 zarchiwizowane](../archive/2026-10-23/frameworks/FRAMEWORKS_F1.md) | F0, F1: DONE (F1 dla swojego faktycznego zakresu); F2: sześć wycinków (DI, oznaczone korzenie web.xml, `@Bean` factory methods, katalog zewnętrznych punktów wejścia, Spring Data repository beans, `@Value`-owe właściwości) + korekta false positives zrobione; F3: pierwszy wycinek (tożsamość proxy/target/interceptor, `internal/aop`) zrobiony, reszta UNASSESSED; F4–F5: UNASSESSED | F0 oddziela wykrycie frameworka od aktywnego profilu. F1 dostarcza realne wartości adnotacji (Spring MVC routes/bean names/`@Profile`), ograniczony `web.xml`/XML bean definitions odczyt, i wykrywanie konfliktów nazw beanów. F2's pierwszy wycinek: `@Autowired` injection points rozwiązane po typie/qualifier/primary wobec F1's `BeanNames` — zmierzone pozytywnie na WebGoat (7/74), poprawnie 0/8 na petclinic (Spring Data repository, inny mechanizm). F2's drugi wycinek: `web.xml`-owe servlety/filtry/listenery dostają realne korzenie pod `-roots profile` — zweryfikowane tylko na testach jednostkowych (petclinic/WebGoat nie mają `web.xml`). F2's trzeci wycinek: `@Bean` factory methods na `@Configuration` klasach dodane do tego samego katalogu beanów — WebGoat skacze z 7/74 do 40/74 rozwiązanych (21 `@Bean` metod znalezionych), petclinic bez zmiany (0/8, niezależne typy). F2's czwarty wycinek: `ExternalEntries` scala `Routes`+`WebXMLRoots` w jeden katalog z `Basis`/`Active` (F0's `ExternalEntry` kontrakt, teraz realny) — czysto raportowy, zero nowych korzeni/krawędzi; petclinic 17, WebGoat 19 wpisów. F2's piąty wycinek: `RepositoryBeans` (interfejsy aplikacji — podtypy `Repository`, ten sam sprawdzian co `synthetic_edges`) — petclinic skacze z 0/8 do **8/8**. **Korekta (nie wycinek):** liczby WebGoat (74/7/40/50) były zawyżone fałszywymi trafieniami niejawnej reguły "jedyny konstruktor" Springa 4.3 (Lombok, enumy, ręczne `new`) — po poprawce w `InjectionPoints` prawdziwe liczby to **43 punkty, 37 rozwiązanych, 6 w pełni wyjaśnionych** (1 `@Value`, 4 `List<Bean>`, 1 framework-provided bean). F2's szósty wycinek: `@Value`-owe punkty wstrzyknięcia (pole lub parametr konstruktora z `@Value(...)`) rozpoznane jako **własna, odrębna kategoria** — nigdy nie przechodzą przez wyszukiwanie beanów (Spring rozwiązuje je przez property sources, nie katalog beanów), więc `Reason` mówi teraz "a @Value config value (...), not a bean lookup", nie mylące "no known bean". Ujawniło też, że pole-samo-`@Value` (bez `@Autowired`) **nigdy wcześniej nie było skanowane** — WebGoat skacze z 43 do **45** punktów (2 nowo znalezione pola `LessonMenuService#excludeCategories`/`excludeLessons`), z czego **8** nierozwiązanych teraz w pełni wyjaśnionych: **3** realne `@Value` (było 1 — pozostałe 2 to właśnie te nowo znalezione pola, które poprzednia sesja błędnie zaliczyła do "4 `List<Bean>`"), **4** realne `List<Bean>` (parametry konstruktora, odrębne od pól), **1** framework-provided bean. petclinic bez zmian (8/8, brak `@Value` w kodzie). Zero wpływu na reachability na obu aplikacjach (`@Value` to fakt konfiguracyjny, nigdy krawędź grafu). Rejestr B3 — nietknięty. |
| ASYNC_CALLBACKS | [A0–A4](ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD) | A0–A2 (MVP): DONE; A3–A4: UNASSESSED | A2 dodaje realną krawędź "może wykonać" do `run`/`call` albo (po DYNAMIC_LINKAGE L0/L1) prosto do implementacji lambdy — 67/44 z 69/46 z 135/134 rozwiązanych na petclinic/WebGoat, z 2 szczerymi lukami dla klas JDK poza analizowanym uniwersum. A0–A2 domykają zadeklarowane MVP tego obszaru; continuations/listenery/reactive (A3) nietknięte. |
| DYNAMIC_LINKAGE | [L0–L4](DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD) | L0, L1 (creation-site case), L2 (same-method + static-field case): DONE; L1 (SAM-call materialization), L2 (instance-field/return-value tracking), L3–L4: UNASSESSED | Po DATA_FLOW D3's pierwszym wycinku L2 wciąż daje **0 z 10/46** realnych krawędzi na petclinic/WebGoat — rzeczywiste miejsca używają pól instancyjnych albo wyniku innego wywołania, nie pól statycznych, które D3 teraz obsługuje. Szczery, zmierzony wynik po dwóch przyrostach. |
| DESERIALIZATION | [Z0–Z4](DESERIALIZATION_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Role entry point/mapowanie CVE i wspólny model obiektów. |
| GRAPH_PRECISION | [P0–P4](GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Najpierw poprawność dispatch i alokacje, potem RTA. |
| RUNTIME_NATIVE | [N0–N5](RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Importer i kontrakt dowodów przed opcjonalnym agentem. |
| ECOSYSTEMS | [E0–E4](ECOSYSTEMS_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Osobny pilot; nie jest warunkiem ulepszania JVM. |

Po pierwszej sesji wpisz poniżej konkretne identyfikatory faz i ich stan. Ogólne `UNASSESSED` w tabeli obszarów zastąp krótkim podsumowaniem rzeczywistego postępu; nie oznaczaj całego obszaru DONE po jednej fazie.

## Rejestr faz

### JVM_GRAPH:G0 — Kontrakt miejsc i krawędzi — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** minimalne B0–B1 potrzebne do G0 (format
  dowodu krawędzi); żadna inna faza nie była wymagana wcześniej.
- **Zmienione pliki:** `internal/classfile/calls.go` (`Call` zyskuje `PC int`
  — offset bytecode instrukcji invoke*, wypełniany w `ExtractCalls` i
  `resolveInvokeDynamic`); `internal/callgraph/graph.go` (nowy `EdgeBasis`
  z wartościami `BasisDirect`/`BasisCHA`/`BasisSynthetic` — miejsce
  zarezerwowane też na przyszłe `BasisInit` z G3 — oraz `CallSite`/`Sites
  []CallSite` na `Edge`, plus `EdgeBetween` do odczytu konkretnej
  krawędzi); `internal/callgraph/build.go` (`seen` zmienia typ z
  `map[Edge]bool` na `map[edgeKey]int`, żeby duplikat tego samego
  `(to,kind,basis)` z innego call site dopisywał dowód do istniejącej
  krawędzi, a nie go gubił); `internal/callgraph/synthetic.go`
  (`hasEdge`/`addEdge` zaktualizowane do nowej sygnatury);
  `internal/reach/reach.go` (`EntryEdge` zyskuje `Kind`/`Basis`/`PC` i
  metodę `PCString()`); `internal/report/templates/report.md.tmpl`
  (linia "Entry call" renderuje teraz rodzaj wywołania, regułę
  rozwiązania i offset bytecode).
- **Zachowanie dostępne użytkownikowi:** raport (`report.md` i
  `results.json`'s `entry.kind`/`entry.basis`/`entry.pc`) pokazuje dla
  wejściowej krawędzi REACHABLE/FRAMEWORK, czy to wywołanie bezpośrednie
  czy rozwinięte przez CHA (a więc mniej wiarygodne), jaki to rodzaj
  wywołania (virtual/interface/static/special/dynamic/synthetic), i
  konkretny offset bytecode (`pc=N`) — zamiast tylko dwóch nazw metod.
- **Spełnione kryteria (z G0's "Gotowe"):**
  - "krok raportu odsyła do konkretnego miejsca wywołania i reguły
    rozwiązania" — potwierdzone na realnym przebiegu petclinic
    (`OwnerController#setAllowedFields → WebDataBinder#setDisallowedFields`,
    `virtual call, resolution: direct, pc=10`) i WebGoat
    (`VulnerableComponentsLesson#completed → XStream#fromXML`, `virtual
    call, resolution: direct, pc=77`).
  - "deduplikacja nie usuwa alternatywnych dowodów" — test
    `TestEvidenceIsPreservedAcrossDuplicateCallSites`
    (`internal/callgraph/build_test.go`): cztery wywołania
    `StringBuilder#append` z czterech różnych offsetów w jednej metodzie
    kolapsują do jednej krawędzi, ale wszystkie cztery offsety
    przechowane jako `Sites`.
  - "kolejność wyników jest stabilna" — niezmienione (nie dotyczy tej
    fazy bezpośrednio, ale sprawdzone ponownie: dwa przebiegi offline na
    petclinic dają bajt-identyczny `results.json`, włącznie z nowymi
    polami `entry.*`).
- **Niespełnione/odłożone:** `EdgeBetween` bierze pierwszą krawędź
  między parą węzłów, jeśli istnieje więcej niż jedna (różny `Kind`) —
  wystarczające dla raportu, nie dla pełnej precyzji; to G1–G2's własny
  zakres (deterministyczny wybór klas, dokładne rozwiązywanie wywołań z
  deskryptorem) nie został tu dotknięty — overload-insensitive
  dopasowanie w `internal/reach/match.go` działa jak wcześniej.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (wszystkie pakiety, zielone); `bin/ariadne` dwa pełne
  przebiegi `-offline` na petclinic (`-profile spring`:
  REACHABLE=3/FRAMEWORK=59/NOT_REACHABLE=29/UNKNOWN=51/NOT_APPLICABLE=6,
  identyczne z istniejącą bazową wartością) i jeden na WebGoat
  (REACHABLE=40/FRAMEWORK=43/NOT_REACHABLE=21/UNKNOWN=56/NOT_APPLICABLE=0,
  również identyczne) — potwierdzające brak regresji liczb przy tym
  refaktorze; `diff` dwóch przebiegów petclinic na bajt identyczny
  `results.json`.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Direct calls, CHA
  and inheritance" i "Markdown and JSON reports" rozszerzone o nowy
  zakres); `README.md` (sekcja "8. The report" opisuje nową linię "Entry
  call").
- **Następny krok:** `JVM_GRAPH:G1` (deterministyczny wybór klas, jawna
  kolejność wejścia w `internal/hierarchy/build.go`) — wymaga zgody
  użytkownika przed startem, jak każda kolejna faza.

### JVM_GRAPH:G1 — Pochodzenie i deterministyczny wybór klas — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** G0 (gotowe wyżej). Model komponentu
  CVE F1 nie istnieje jeszcze — zamiast czekać na całą fazę F1 (poza
  zakresem tego przyrostu), z jej opisu wzięto tylko wspólny element,
  który G1 faktycznie potrzebuje: SHA-256 jako dowód tożsamości
  zawartości pliku klasy. To nie jest realizacja F1; gdy F1 faktycznie
  powstanie, ma użyć tego samego `hashClassEntry`, nie duplikować go.
- **Zmienione pliki:** `internal/hierarchy/build.go` (merge wyników
  czytania jarów odbywa się teraz po indeksie zadania w jawnej liście
  wejściowej — `results := make([]result, len(jobs))` — a nie w
  kolejności, w jakiej kanał `resCh` dostarczył wyniki, co wcześniej
  zależało od tego, który worker skończył pierwszy); nowy
  `internal/hierarchy/duplicates.go` (`SourceRef`, `Duplicate`,
  `hashClassEntry` — leniwe liczenie SHA-256 tylko dla faktycznych
  kolizji, nie dla każdej klasy, żeby nie dotykać gorącej ścieżki
  odczytu 29–40 tysięcy klas); `internal/hierarchy/hierarchy.go`
  (`Hierarchy.duplicates`/`Duplicates()`; poprawiony nieaktualny
  komentarz przy `SourceJar`, który twierdził, że jest "" dla klas
  aplikacji — kod zawsze ustawia rzeczywistą ścieżkę jara).
- **Zachowanie dostępne użytkownikowi:** które wystąpienie zduplikowanej
  klasy "wygrywa" zależy teraz wyłącznie od jawnej kolejności wejścia
  (aplikacja, potem zależności w kolejności podanej przez
  `ResolveDepJars`/CLI — `os.ReadDir` już sortuje alfabetycznie, co
  samo w sobie już daje odtwarzalną kolejność classpath, więc ten
  element zadania był już spełniony), nigdy od szybkości workera.
  `Hierarchy.Duplicates()` daje też strukturalny dowód (ścieżka +
  SHA-256) obok dotychczasowego ostrzeżenia tekstowego.
- **Spełnione kryteria (z G1's "Gotowe"):**
  - "permutacja czasu workerów nie zmienia grafu" — mechanizm wyklucza
    to z definicji (merge po indeksie zadania ustalonym przed
    uruchomieniem jakiegokolwiek workera), potwierdzone też testem pod
    `-race`.
  - "zmiana jawnej kolejności zmienia wybraną implementację w
    przewidywalny sposób" — test
    `TestBuild_duplicateWinnerFollowsExplicitOrderNotReadOrder`
    (`internal/hierarchy/hierarchy_test.go`): dwie kopie tego samego
    jara pod różnymi ścieżkami, zamiana ich kolejności w `depJars`
    przewidywalnie zamienia, która ścieżka "wygrywa".
  - Dowody nie są tylko konstrukcją testową: realny przebieg na
    petclinic ujawnił **31 prawdziwych kolizji klas** (np.
    `jakarta.activation-1.2.2.jar` vs
    `jakarta.activation-api-1.2.2.jar`), a na WebGoat **242** — czyli to
    wcześniej latentne niezdeterminowanie było realne, nie
    teoretyczne, na obu rzeczywistych aplikacjach kontrolnych.
- **Niespełnione/odłożone:** "oznaczać przyjętą kolejność jako
  założenie" i "zachować konflikt w pokryciu dla nieobsługiwanej
  izolacji loaderów" z G1's zadań nie dostały nowego, dedykowanego pola
  — pokrywa je nadal istniejący mechanizm ostrzeżeń plus nowy
  `Duplicates()`; pełny model izolacji loaderów to RUNTIME_NATIVE N1–N2,
  nie ta faza (zgodnie z ROADMAP.MD B1's własną uwagą).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone); `go test -race ./internal/hierarchy/...`
  (zielone); `bin/ariadne -offline` na petclinic (`-profile spring`:
  3/59/29/51/6, identyczne jak przed zmianą, 273180 węzłów/1229725
  krawędzi bez zmian) i WebGoat (40/43/21/56/0, 339786/2591563 bez
  zmian) — brak regresji; dwa przebiegi petclinic dają bajt-identyczny
  `results.json`.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Duplicate class
  selection" rozszerzony o nowy zakres i dowody z realnych przebiegów).
  `README.md`/`FAQ.md` nie opisywały wcześniej duplikatów klas jako
  osobnego tematu, więc nie wymagały zmiany.
- **Następny krok:** `JVM_GRAPH:G2` (dokładne rozwiązywanie wywołań —
  lookup `fqcn+name+descriptor+rodzaj wywołania`, współdzielone z CVE
  F4) — wymaga zgody użytkownika przed startem.

### JVM_GRAPH:G2 — Dokładne rozwiązywanie wywołań — DONE for overloads/inheritance; bridge/default-method real-bytecode verification pending (2026-10-22)

- **Zakres zlecenia i zależności:** G1 (gotowe wyżej). Współdzielony
  selektor metody z CVE F4 nie istnieje (F4 nieodnotowane) — z opisu F4
  wzięto tylko to, co G2 faktycznie potrzebuje: dopasowanie po
  `fqcn+name+descriptor`, nie całą fazę F4 (role `vulnerable`/
  `entry_point` z odpowiedzi LLM, wersjonowanie zakresów itd. — to
  zostaje nienaruszone, czeka na F4 samą).
- **Ograniczenie środowiska, odnotowane szczerze:** to środowisko ma
  tylko JRE (`openjdk-21-jre`/`openjdk-25-jre`), nie JDK — `javac` nie
  istnieje, więc nie dało się skompilować nowych klas fixture z
  przeciążeniami/bridge methods/default methods, jak proponowało G2's
  własne "Gotowe". Zamiast tego: (1) `hierarchy.NewForTesting` (nowy,
  mały, wyraźnie opisany jako testowy konstruktor `*Hierarchy` z
  gotowych `*ClassInfo`, bez I/O na jarach) pozwala testować dokładnie
  te kształty (przeciążenia, diamond interfaces) bez JDK; (2) scenariusz
  "odziedziczona metoda wołająca bibliotekę" jest pokryty **realnym,
  skompilowanym bytecode** — `testdata/reachfixture` już miało tę
  dokładną sytuację od wcześniejszej fazy (Phase 5 w PLAN.md), więc
  istniejący test `TestAnalyze_inheritedMethodMatchedThroughSubclassCall`
  służy jako prawdziwa, bytecode'owa weryfikacja tego przypadku, nie
  tylko konstrukcja testowa. Bridge methods i default methods **nie
  zostały** osobno zweryfikowane na realnym skompilowanym bytecode w tym
  przyroście — logika rozwiązywania jest zaprojektowana tak, by je
  obsłużyć bez specjalnych przypadków (patrz komentarz w
  `resolve.go`/`build.go`), ale to pozostaje nieprzetestowane na
  prawdziwych `.class`; zarejestrowane jako ograniczenie, nie przemilczane.
- **Zmienione pliki:** `internal/hierarchy/hierarchy.go`
  (`ClassInfo.MethodByNameAndDesc`, `NewForTesting`);
  `internal/hierarchy/resolve.go` (refaktor na współdzielone
  `resolveBy`, nowy `ResolveExact` — dokładny odpowiednik `ResolveMethod`,
  który zostaje jako jawny tryb legacy/przybliżony, bo LLM-grounding i
  vulnmap nigdy nie mają deskryptora); `internal/callgraph/build.go`
  (`resolveTarget` przekierowuje krawędź bezpośrednią na rzeczywistą
  klasę deklarującą metodę; CHA filtruje teraz po dokładnym deskryptorze,
  nie tylko nazwie); `internal/callgraph/graph.go`
  (`UnresolvedOwnerCount`); `internal/reach/match.go` (komentarz
  wyjaśniający, czemu dopasowanie podatnych metod zostaje
  overload-insensitive, i jak to teraz współdziała z G2); `cmd/ariadne/main.go`,
  `internal/report/report.go`, `internal/report/templates/report.md.tmpl`
  (nowa, jawnie nie-alarmująca linia diagnostyczna).
- **Zachowanie dostępne użytkownikowi:** (1) ścieżki w raporcie teraz
  pokazują rzeczywistą klasę deklarującą metodę, nie martwy węzeł pod
  nazwą podklasy; (2) CHA nie łączy już różnych przeciążeń tej samej
  nazwy; (3) nowa linia w raporcie "Calls resolved outside the analyzed
  input" — **jawnie opisana jako informacyjna, nie alarm**, po tym jak
  pierwsza wersja tego liczika (nazwana "unresolved call owners") okazała
  się być zdominowana przez zwykłe, oczekiwane dziedziczenie z JDK
  (`java.lang.Thread#start`, `javax.swing.JFrame#setSize`, ...) —
  sprawdzone wprost próbką 20 przykładów z realnego przebiegu na
  petclinic, nie założone.
- **Spełnione kryteria (z G2's "Gotowe", częściowo — patrz wyżej):**
  - "fixtures obejmują odziedziczoną metodę wołającą bibliotekę" —
    tak, na realnym bytecode (`testdata/reachfixture`).
  - "przeciążenia" — tak, ale na ręcznie zbudowanej hierarchii
    (`TestCHA_overloadsAreNotConflated`,
    `TestResolveExact_overloadsAndDiamondInterface`), nie na
    skompilowanym bytecode (brak JDK).
  - "override" — tak, pokryte już istniejącym
    `TestCHA_interfaceCallReachesBothImplementations` (bytecode) plus
    nowymi testami.
  - "default method" i "bridge" — **nie zweryfikowane na realnym
    bytecode**; logika jest zaprojektowana żeby to obsłużyć (patrz
    komentarz), ale to twierdzenie projektowe, nie zmierzony wynik.
    Odnotowane jako otwarte ograniczenie tej fazy.
  - "nieprawidłowy lub brakujący cel daje diagnostykę" — tak
    (`UnresolvedOwnerCount`), ale po korekcie: pierwsza wersja nazwy i
    opisu była myląca (patrz wyżej), poprawiona po realnej weryfikacji,
    nie przed nią.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone); `go test -race
  ./internal/callgraph/... ./internal/hierarchy/... ./internal/reach/...`
  (zielone); `bin/ariadne -offline` na petclinic (`-profile spring`:
  REACHABLE 3→3, FRAMEWORK 59→62, NOT_REACHABLE 29→26, węzły 273180→257640,
  krawędzie 1229725→908326 — CHA-dodane krawędzie 630816→310418, spadek
  o połowę, zgodny z oczekiwaniem po naprawieniu konfliktu przeciążeń)
  i WebGoat (REACHABLE 40→42, FRAMEWORK 43→44, NOT_REACHABLE 21→18);
  jednorazowy skrypt diagnostyczny (uruchomiony i usunięty) próbkujący
  8868 przypadków "unresolved owner" na petclinic, potwierdzający że
  dominują wywołania dziedziczone z JDK (`antlr.*`/`java.lang.Thread`/
  `javax.swing.JFrame` w przykładach); dwa przebiegi petclinic dają
  bajt-identyczny `results.json`; ręczne porównanie `results.json` przed/po
  wykazało **jeden rzeczywisty, potwierdzony false-positive naprawiony**:
  `CVE-2021-22118` FRAMEWORK → NOT_REACHABLE — stara (nazw-only) CHA
  tworzyła sztuczną ścieżkę przez Spring WebFlux'owy reaktywny
  multipart reader, którego klasyczny Spring MVC/Servlet stack petclinic
  w ogóle nie używa; to jest dokładnie ten rodzaj artefaktu CHA, który
  `-framework-check`'s LLM (Phase 8) był zbudowany żeby wyławiać —
  teraz usunięty strukturalnie, nie tylko oflagowany.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Direct calls, CHA
  and inheritance" rozszerzony o nowy zakres, ograniczenia i zmierzony
  efekt). `README.md`/`FAQ.md` nie opisywały wcześniej tego mechanizmu
  na tyle szczegółowo, żeby wymagały zmiany.
- **Niespełnione/odłożone:** bridge/default-method weryfikacja na
  realnym bytecode (potrzebuje JDK w środowisku lub przeniesienia tej
  weryfikacji do sesji, która go ma); CVE F4's właściwy zakres (wersje,
  role LLM, model v2) nienaruszony.
- **Następny krok:** `JVM_GRAPH:G3` (relacje inicjalizacji — krawędzie do
  `<clinit>` przy `new`/`getstatic`/`putstatic`/`invokestatic`) — wymaga
  zgody użytkownika przed startem.

### JVM_GRAPH:G3 — Relacje inicjalizacji — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** G2 (gotowe wyżej, częściowo —
  overloads/inheritance). Część metadanych współdzielona z DATA_FLOW D1
  nie została naruszona (D1 nieodnotowane) — `classfile.ExtractCalls`
  już wcześniej zachowywał tabelę wyjątków; G3 nie musiało tego dotykać.
- **Zmienione pliki:** `internal/classfile/opcodes.go` (`opNew`,
  `opGetStatic`, `opPutStatic`); `internal/classfile/pool.go`
  (`fieldref` — odpowiednik `methodref` dla Fieldref); `internal/classfile/calls.go`
  (`InitKind`/`InitRef`; `ExtractCalls` zyskuje drugą wartość zwracaną —
  zmiana sygnatury, zaktualizowano 4 miejsca wywołania);
  `internal/hierarchy/hierarchy.go` (`ClassInfo.Fields`,
  `FieldByNameAndDesc`, `Method.InitRefs`); `internal/hierarchy/jarread.go`
  (przekazuje `cf.Fields`/`InitRefs` dalej); `internal/hierarchy/resolve.go`
  (`ResolveField` — **inny porządek przeszukiwania niż `ResolveExact`**:
  JVMS 5.4.3.2 sprawdza superinterfejsy przed nadklasą, odwrotnie niż
  rozwiązywanie metod; z ochroną przed cyklem przez `seen` map);
  `internal/callgraph/graph.go` (`BasisInit`, `CallInit`, `InitEdgeCount`);
  `internal/callgraph/build.go` (`addInitEdge`; Pass 3 — krawędzie
  `<clinit>` → nadklasy `<clinit>`; poprawka liczenia `chaEdges`, żeby
  krawędzie inicjalizacyjne nie zawyżały liczby "added by CHA" — błąd
  złapany i naprawiony w trakcie, nie przed weryfikacją);
  `internal/profile/profile.go` (`roots.main` dodaje też `<clinit>` klasy
  wejściowej jako root); `cmd/ariadne/main.go`, `internal/report/report.go`,
  `internal/report/templates/report.md.tmpl` (nowa pozycja "class-init").
- **Ograniczenie środowiska (to samo co w G2):** brak JDK — scenariusze
  testowane przez (1) ręcznie zmontowany bytecode na poziomie
  `classfile.ExtractCalls` (`internal/classfile/initrefs_test.go`:
  `new`/`getstatic`/`putstatic`/`ldc` jako gotowe bajty, nie skompilowany
  plik — dla tej warstwy to nawet precyzyjniejszy test niż jedna
  konkretna wersja javac), i (2) ręcznie zbudowana hierarchia
  (`hierarchy.NewForTesting`) na poziomie `internal/callgraph`.
- **Spełnione kryteria (z G3's "Gotowe"):**
  - "ścieżka do wywołania w zależnościowym `<clinit>` istnieje tam, gdzie
    jest uzasadniona" — potwierdzone na realnych petclinic/WebGoat:
    66491/89483 nowych krawędzi `class-init`.
  - "testy obejmują odziedziczone pole statyczne" —
    `TestInit_inheritedStaticField`.
  - "brak `<clinit>`" — `TestInit_noClinitProducesNoEdge`.
  - "cykl" — `TestInit_cycleDoesNotHang`.
  - "przypadki bez inicjalizacji" — `TestExtractCalls_initRefs` (ldc
    `Foo.class` nie tworzy `InitRef`).
  - "Graf nie obiecuje dokładnej kolejności runtime'u" — prawda z
    konstrukcji: to pojedyncze krawędzie, nie model kolejności/czasu.
- **Błąd złapany i naprawiony w trakcie pracy, nie przemilczany:**
  pierwsza wersja liczyła nowe krawędzie inicjalizacyjne jako "added by
  CHA" (bo `chaEdges` liczyło się jako `EdgeCount()-PreCHAEdges` bez
  odjęcia nowej kategorii) — realny przebieg na petclinic pokazał CHA
  skaczące z 310418 (koniec G2) na 376381 mimo że logika CHA się nie
  zmieniła, co było sygnałem błędu w księgowaniu, nie w samej analizie.
  Naprawione dodaniem `Graph.initEdges`/`InitEdgeCount()` i odjęciem go
  od `chaEdges`; po naprawie CHA wróciło do identycznych 310418/556769
  jak w G2, a nowa kategoria "class-init" (66491/89483) jest liczona
  osobno.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone); `go test -race
  ./internal/callgraph/... ./internal/classfile/... ./internal/hierarchy/...
  ./internal/profile/... ./internal/reach/...` (zielone); `bin/ariadne
  -offline` na petclinic i WebGoat — węzły i direct/CHA bez zmian
  względem G2, 66491/89483 nowych krawędzi class-init, **żadna z CVE nie
  zmieniła statusu w żadnej z dwóch aplikacji** (porównanie `results.json`
  G2 vs G3, zero różnic) — szczery, zmierzony wynik: funkcja jest realna
  i przetestowana, ale na obecnej bazowej analizie petclinic/WebGoat
  jeszcze nie odkrywa nowych REACHABLE/FRAMEWORK ścieżek; dwa przebiegi
  petclinic dają bajt-identyczny `results.json`.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Implicit class
  initialization": Not implemented → Partial, z pełnym opisem zakresu,
  ograniczeń i zmierzonego efektu).
- **Niespełnione/odłożone:** domyślne metody i eager-init superinterfejsu
  je deklarującego (JVMS 5.5's węższa reguła) — nie zaimplementowane,
  rzadki przypadek; inicjalizacja przez refleksję (`Class.forName`) —
  poza zakresem tej fazy, należy do REFLECTION.
- **Następny krok:** `JVM_GRAPH:G4` (target Java i multi-release JAR) —
  wymaga zgody użytkownika przed startem. Alternatywnie: dokończyć G2's
  odłożoną weryfikację bridge/default methods, jeśli w przyszłej sesji
  będzie dostępny JDK.

### JVM_GRAPH:G4 — Target Java i multi-release JAR — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** G1 (gotowe wyżej). Wspólny opis
  targetu z CVE F8a nie istnieje (F8 nieodnotowane) — zgodnie z G4's
  własnym zapisem "pełny adapter podatności JDK nie jest wymagany",
  zaimplementowano tylko to, co G4 faktycznie potrzebuje (flaga
  `-target-java` + wybór wariantu), bez czekania na F8a.
- **Zmienione pliki:** `internal/hierarchy/jarread.go` (przebudowane:
  `selectClassEntries`/`isMultiReleaseJar`/`parseVersionDir` — logika
  wyboru wariantu; `readJarClasses` zyskuje `targetJava int` i drugą
  wartość zwraną `[]string` ostrzeżeń); `internal/hierarchy/hierarchy.go`
  (`ClassInfo.SourceEntry`/`VariantVersion`); `internal/hierarchy/build.go`
  (`Build` zostaje niezmienione jako wrapper; nowe `BuildWithTarget` z
  parametrem `targetJava` — **żadny z 11 istniejących wywołań `Build`
  nie wymagał zmiany**, dzięki temu podziałowi); `cmd/ariadne/main.go`
  (`-target-java` flaga + walidacja + log z liczbą wybranych wariantów).
- **Zachowanie dostępne użytkownikowi:** `-target-java 17` wybiera, dla
  jara z `Multi-Release: true` w manifeście, najwyższy
  `META-INF/versions/<V>/...` z `V <= 17` per klasa, z fallbackiem do
  wpisu bazowego. Bez flagi — zawsze wpis bazowy, **nigdy zgadywanie z
  najwyższej widzianej wersji bytecode** (to był jawny wymóg G4's
  "Gotowe"). `ClassInfo.SourceEntry`/`VariantVersion` zachowują
  faktyczną wybraną ścieżkę i wersję.
- **Spełnione kryteria (z G4's "Gotowe"):**
  - "targety bazowy/11/17 wybierają oczekiwane klasy" —
    `TestSelectClassEntries_picksHighestQualifyingVariant` (7 celów:
    8/9/10/11/16/17/21 — potwierdza też, że cel 21 wybiera wariant 17,
    nie zgaduje wariantu 21, którego nie ma).
  - "brak flagi" — `TestSelectClassEntries_unsetTargetFallsBackToBase`.
  - "brak wariantu" — pokryte przez przypadek target=8 w teście powyżej
    (żaden wariant nie kwalifikuje się, wpis bazowy wygrywa).
  - "sprzeczne wejście" — nie istnieje drugie źródło metadanych runtime'u
    do konfliktu z (F8a nieodnotowane), więc walidacja ograniczona do
    samej flagi (`-target-java` musi być dodatnie); odnotowane jako
    świadome ograniczenie, nie przemilczane.
  - "nie traktować każdego META-INF/versions/ jako aktywnego wariantu" —
    `TestSelectClassEntries_notMultiReleaseIgnoresVersionsDir` (jar bez
    `Multi-Release: true` w manifeście nigdy nie aktywuje wariantów,
    nawet z podanym celem).
  - "raportować niezgodne wejścia" —
    `TestSelectClassEntries_malformedVersionDirWarnsAndFallsBackToBase`
    (nienumeryczny katalog wersji i wersja <9 — oba dają ostrzeżenie,
    nie błąd).
  - "Snapshot zachowuje wybór" — `results.json` bajt-identyczny między
    dwoma przebiegami petclinic z `-target-java 17` (patrz dowody
    poniżej).
- **Dowód na realnym bytecode, nie tylko metadanych:**
  `TestBuildWithTarget_selectsRealVariantBytecode` — syntetyczny
  multi-release jar, którego wpis bazowy to prawdziwe, skompilowane
  bytecode `fixture.Greeter`, a wariant wersji 17 to prawdziwe
  `fixture.FormalGreeter` (dwie różne, rzeczywiste klasy jako
  rozróżnialna "treść wariantu" — bez JDK nie można było skompilować
  jednej klasy w dwóch wersjach celowo różniących się czymś
  obserwowalnym). Potwierdza, że `readJarClasses` faktycznie parsuje
  wybrane bajty, nie tylko że `selectClassEntries` nazwał właściwy
  `zip.File`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone); `go test -race ./internal/hierarchy/...`
  (zielone); `bin/ariadne -offline` na petclinic i WebGoat, z i bez
  `-target-java 17`:
  - **petclinic ma 8 realnych zależności multi-release**
    (`log4j-api-2.13.3`, `byte-buddy-1.10.22`, `classgraph-4.8.69`,
    `h2-1.4.200`, `HikariCP-3.4.5`, `istack-commons-runtime-3.0.12`,
    `jakarta.xml.bind-api-2.3.3`, `jaxb-runtime-2.3.4`) — potwierdzone
    bezpośrednim sprawdzeniem manifestów, nie założone. `-target-java 17`
    aktywuje **15 wariantów** (log: "29125 classes ... target Java 17,
    15 multi-release variants selected").
  - **WebGoat: 21 aktywowanych wariantów.**
  - **Zero zmian statusu CVE w obu aplikacjach** (porównanie `results.json`
    bez/z `-target-java 17`, zero różnic) — szczery wynik: funkcja
    realna i aktywowana na prawdziwych danych, ale na tej bazie nie
    odkrywa nowych ścieżek.
  - Dwa przebiegi petclinic z `-target-java 17` dają bajt-identyczny
    `results.json`.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Multi-release
  target selection": Not implemented → Partial, z pełnym opisem i
  dowodami z realnych przebiegów); `README.md` (nowy wiersz `-target-java`
  w tabeli flag).
- **Niespełnione/odłożone:** "uzgodnić z metadanymi runtime'u, jeśli
  podano oba wejścia" — nie zaimplementowane, bo drugie źródło metadanych
  nie istnieje jeszcze (CVE F8a); nic do uzgadniania na razie. Pełny
  adapter podatności JDK (reszta F8) — poza zakresem tej fazy z
  definicji.
- **Następny krok:** `JVM_GRAPH:G5` (integracja i kwalifikacja całego
  G0–G4 — fixtures do `testdata`, porównanie grafu przed/po, aktualizacja
  README/raportu/ograniczeń) — to domyka R0's JVM_GRAPH w całości.
  Wymaga zgody użytkownika przed startem.

### JVM_GRAPH:G5 — Integracja i kwalifikacja — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** G0–G4 (gotowe wyżej, z jednym
  odnotowanym ograniczeniem w G2). To faza integracyjna, nie nowa
  funkcja — zgodnie z jej własnym zapisem "nie ogłaszać, że rozwiązano
  refleksję, callbacki i JNI przez samo poprawienie dispatch", poniższe
  nie zmienia żadnego z tych trzech obszarów (wciąż `Not implemented` w
  `docs/SUPPORT.MD`, sprawdzone wprost poniżej).
- **Fixtures do testdata, ze źródłami i instrukcją regeneracji:**
  `testdata/fixture/src/fixture/Box.java` (generyczny interfejs),
  `StringBox.java` (implementacja `Box<String>` — wymusza syntetyczną
  metodę bridge), `BridgeDemo.java` (wywołuje obie nowe sceny: bridge i
  odziedziczoną metodę domyślną), oraz dodana metoda domyślna
  `Greeter#defaultGreeting`. **Status: źródła dodane, nieskompilowane**
  — to środowisko nie ma `javac` (sprawdzone wprost w G2/G3), więc
  `fixture-8.jar`/`fixture-21.jar` nie zawierają jeszcze tych klas.
  `scripts/build-fixtures.sh` dostał komentarz z instrukcją: uruchomić z
  prawdziwym javac, potem dodać cross-check opisany w `BridgeDemo.java`'s
  własnym komentarzu. Nie dodano analogicznej sztuki dla G4 (multi-release
  fixture wymagający dwóch kompilacji) — G4 ma już **realny dowód
  bytecode** inną metodą (patrz jego własny wpis: dwie prawdziwe klasy z
  istniejącego fixture użyte jako rozróżnialna "treść wariantu"), więc
  dodatkowy, niedostarczalny teraz fixture miałby niską wartość
  krańcową względem G2's rzeczywistej dziury.
- **Porównanie grafu przed/po (petclinic, `-profile spring -offline`,
  ten sam cache przez cały ciąg G0→G4):**

  | Etap | Status (R/F/NR/U/NA) | Węzły | Krawędzie (direct+CHA+init) |
  |---|---|---|---|
  | Przed G0 (Phase 7 baseline) | 3/59/29/51/6 | 273180 | 1229725 (598909+630816+0) |
  | Po G0 (dowody krawędzi, bez zmiany logiki) | 3/59/29/51/6 | 273180 | 1229725 (598909+630816+0) |
  | Po G1 (deterministyczny wybór duplikatów) | 3/59/29/51/6 | 273180 | 1229725 (598909+630816+0) |
  | Po G2 (dokładne rozwiązywanie — **realna poprawka**) | 3/62/26/51/6 | 257640 | 908326 (597908+310418+0) |
  | Po G3 (krawędzie inicjalizacji) | 3/62/26/51/6 | 257640 | 974817 (597908+310418+66491) |
  | Po G4 (`-target-java 17`) | 3/62/26/51/6 | 257688 | 974788 (597848+310414+66526) |

  **Uzasadnienie zmienionych ścieżek:** jedyna zmiana statusu CVE w
  całym ciągu G0–G4 nastąpiła przy G2 i ma w pełni zrozumiane przyczyny
  (zapisane w G2's własnym wpisie): 6 CVE NOT_REACHABLE→FRAMEWORK (CHA
  bez konfliktu przeciążeń odkrywa więcej prawdziwych połączeń) i
  dokładnie 1 FRAMEWORK→NOT_REACHABLE (`CVE-2021-22118` — potwierdzony,
  rzeczywisty artefakt CHA przez reaktywny WebFlux, którego petclinic
  nie używa). G3/G4 nie zmieniły żadnego statusu na tej bazie — spadek
  liczby węzłów po G2 (273180→257640) to martwe węzły-ślepe-uliczki,
  które przestały być tworzone po naprawie dziedziczenia, nie utracone
  informacje.
- **Koszt pamięci przechowywania dowodów (G0's `Edge.Sites`), zmierzony
  wprost na petclinic, nie oszacowany:** jednorazowy instrumentowany
  przebieg (`runtime.MemStats`, usunięty po pomiarze) na grafie po G4:
  974817 krawędzi, z czego 974289 ma niepuste `Sites` (łącznie 1359423
  wpisów `CallSite`). Szacowana pamięć samych dowodów: **~32.7 MiB**
  (1359423×8B danych + 974289×24B nagłówków slice'ów) z **~177 MiB**
  całkowitego RSS Go po zbudowaniu grafu — czyli **dowody krawędzi to
  ok. 18% pamięci grafu**, realny, ale nie dominujący koszt. Nie
  zmierzono analogicznie dla WebGoat (większy graf, oczekiwany
  proporcjonalnie podobny udział — nieprzetestowane, nie założone).
- **Sprawdzone pakiety i kontrole według ROADMAP.MD B4:** `go build
  ./...`, `go vet ./...`, `go test ./...` **i `go test -race ./...`
  dla całego projektu** (nie tylko dotknięte pakiety) — wszystko
  zielone. Testy LLM/live pozostają osobne i opt-in (`-llm`/`ARIADNE_LLM_E2E`),
  nieuruchamiane tutaj.
- **Sprawdzone: brak nadinterpretacji zakresu.** `docs/SUPPORT.MD`
  bezpośrednio sprawdzone: "General reflection, MethodHandle invocation
  and custom bootstrap models" i "General callbacks and asynchronous
  execution models" wciąż `Not implemented` — żadna z poprawek G0–G4 nie
  zmieniła tych wierszy, zgodnie z G5's własnym ostrzeżeniem.
- **Zmiany w dokumentacji:** `README.md` — dwa realnie nieaktualne akapity
  poprawione (nie tylko dopisane): sekcja "4. The call graph and CHA
  expansion" opisywała dawne zachowanie ("Methodref owner exactly as the
  bytecode names it... internal/reach's method matching accounts for
  this") jako wciąż aktualne, mimo że G2 to naprawiło u źródła; podobnie
  punkt "CHA over-approximates" w "Known limitations" nie wspominał
  dokładności deskryptora. Oba poprawione, plus nowy akapit o
  krawędziach inicjalizacji (G3). `scripts/build-fixtures.sh` — komentarz
  z instrukcją regeneracji dla nowych klas fixture.
- **Niespełnione/odłożone:** G2's bridge/default-method weryfikacja na
  realnym bytecode wciąż czeka na JDK w środowisku — źródła są już
  gotowe (ten wpis), tylko kompilacja i cross-check pozostają.
- **Następny krok:** JVM_GRAPH's R0-owy zakres jest zamknięty. Zgodnie z
  ROADMAP.MD, R1 (DATA_FLOW D0–D2, SERVICE S0–S2, ASYNC A0–A2) jest
  teraz odblokowane — albo najpierw wspólne B0–B1 z innych obszarów R0
  (CVE F0–F1 współdzielone części), jeśli jeszcze nieuzgodnione. Wymaga
  propozycji i zgody użytkownika przed startem, jak każda kolejna faza.

Przy dodawaniu kolejnego wpisu zapisz:

- Pełny identyfikator fazy i stan.
- Zakres zlecenia i niezbędne wspólne zależności, które faktycznie wykonano.
- Zmienione pliki i zachowanie dostępne użytkownikowi albo rezultat kontraktowy.
- Spełnione i niespełnione kryteria, z odnośnikami do dowodów.
- Polecenia rzeczywiście wykonane, ich wyniki i zakres; osobno kontrole niewykonane z powodem.
- Zmiany w SUPPORT, README, raporcie i innych dotkniętych opisach, albo uzasadnienie braku zmiany zakresu wsparcia.
- Pozostałe ograniczenia, konieczne decyzje oraz następny krok.

### SERVICE_LOADING:S0 — Inwentaryzacja zasobów i kontrakt zdarzeń — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** B0–B1 (wspólne), `JVM_GRAPH:G1`
  (gotowe, zarchiwizowane) — żadna zależność na DATA_FLOW, zgodnie z
  tym, co S0's własny plan deklaruje (S1 jest pierwszą fazą tego
  obszaru, która faktycznie potrzebuje DATA_FLOW D0–D2).
- **Zmienione pliki:** `internal/input/serviceresources.go`
  (`ListServiceFiles` — czyta surowe wpisy `META-INF/services/*` z jara,
  z hashem SHA-256 treści; żadnej semantyki parsowania, czysty dostęp do
  zasobu, zgodnie z planu rozkazem umieszczenia "indeksu zasobów" w
  `internal/input`); nowy pakiet `internal/serviceload`
  (`parse.go`: `Parse`/`Entry`/`ParseResult`, implementujące dokładnie
  reguły parsowania z javadoc `java.util.ServiceLoader` — komentarz `#`
  do końca linii, białe znaki ucinane, puste linie ignorowane,
  duplikaty tolerowane ale raportowane, nazwa musi być poprawną
  kwalifikowaną nazwą binarną; `inventory.go`: `ServiceRegistration`
  (fakt deklaratywny) + stubowe `ServiceLoadSite`/`ProviderMaterialization`
  (kontrakty dla S1/S2, jeszcze nigdy nie konstruowane w tej fazie) +
  `Inventory` — orkiestracja: dla app i każdego dep jara, listuje pliki
  serwisowe, parsuje, sprawdza każdą nazwę providera względem
  `*hierarchy.Hierarchy`); `internal/report/report.go`/`report.md.tmpl`
  (nowa sekcja "ServiceLoader provider registrations" — tabela
  interfejs/provider/found/source, jawnie opisana jako "not call-graph
  edges"); `cmd/ariadne/main.go` (wywołanie `Inventory` po zbudowaniu
  hierarchii, log z liczbą rejestracji).
- **Zachowanie dostępne użytkownikowi:** raport ma teraz tabelę
  wszystkich deklaracji dostawców usług (`META-INF/services/*`) w
  aplikacji i zależnościach, z informacją czy klasa providera istnieje
  w analizowanym świecie klas — **bez żadnej fałszywej krawędzi w
  grafie wywołań** (S0's własny wymóg: "Rejestracja nie jest krawędzią
  wykonania").
- **Spełnione kryteria (z S0's "Gotowe"):** "poprawna lista kandydatów
  i diagnostyka bez tworzenia fikcyjnych wywołań" — potwierdzone: (1)
  żadna nowa krawędź/węzeł w `internal/callgraph`/`internal/reach` nie
  powstaje z tej fazy (kod jej nie dotyka w ogóle); (2) realny przebieg
  na petclinic/WebGoat: 44/158 prawdziwych rejestracji, 0 fałszywych
  "not found" (sprawdzone wprost, nie założone — np. wszystkie wpisy
  `org.ehcache.core.spi.service.ServiceFactory` z `ehcache-3.9.2.jar`
  poprawnie znalezione, bo ehcache sam dostarcza swoje klasy providerów
  w tym samym jarze).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, w tym nowe testy `internal/input`:
  `TestListServiceFiles_*`; `internal/serviceload`: `TestParse_*`,
  `TestIsBinaryClassName`, `TestInventory_*`); `go test -race ./...`
  dla całego projektu (zielone); `bin/ariadne -offline` na petclinic
  (REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN bez zmian: 3/62/26/51/6 —
  potwierdzone, że S0 nie wpływa na reachability) i WebGoat
  (42/44/18/56/0, bez zmian); dwa przebiegi petclinic dają
  bajt-identyczny `results.json` i `report.md` (poza samym znacznikiem
  czasu) — sekcja ServiceLoader też jest deterministyczna.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "ServiceLoader
  and JPMS service models": Not implemented → Partial, z pełnym opisem
  zakresu i dowodami); `README.md` ("Known limitations" — poprawiony
  bullet o ServiceLoader, już nieaktualny po tej fazie).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy, nie
  przez przeoczenie):** `ServiceLoadSite`/`ProviderMaterialization` są
  zdefiniowane, ale nigdy nie konstruowane — to S1 (potrzebuje
  DATA_FLOW D0–D2 do rozwiązania argumentu `Class` w wywołaniu
  `ServiceLoader.load`) i S2 (modelowanie konstruktora providera jako
  celu w grafie). JPMS (`module-info`, `uses`/`provides`) to S3,
  nietknięte.
- **Następny krok:** `SERVICE_LOADING:S1` wymaga `DATA_FLOW:D0–D2`
  jako zależności (nieodnotowane) — więc naturalny następny krok to
  albo zacząć `DATA_FLOW:D0` (kontrakt IR/wartości), żeby odblokować
  S1 i resztę R1, albo pozostać przy mechanicznych, niezależnych
  krokach w innych obszarach, jeśli takie się znajdą. Wymaga propozycji
  i zgody użytkownika przed startem.

### DATA_FLOW:D0 — Kontrakt IR, wartości i zakresów — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** B0–B1 (wspólne), `JVM_GRAPH:G0–G2`
  (gotowe, zarchiwizowane). To wyłącznie faza kontraktowa — **żaden
  realny CFG czy analiza wartości nie istnieje po tej fazie**; D1
  (CFG) i D2 (lokalne wartości) są osobnymi, nienaruszonymi fazami.
- **Zmienione pliki (wyłącznie nowe pakiety, zero zmian w istniejącym
  kodzie produkcyjnym — D0 niczego jeszcze nie konsumuje):**
  `internal/bytecodeir/site.go` (`Site`, `OpKind` — tożsamość jednej
  instrukcji bytecode: metoda, deskryptor, offset, rodzaj operacji;
  mirror filozofii dowodowej z `JVM_GRAPH:G0`'s `Call.PC`/`Edge.Basis`,
  jeden poziom niżej — na poziomie instrukcji, nie krawędzi grafu
  wywołań); `internal/valueflow/value.go` (`Value`/`Kind`: `KindConst`
  — dokładnie znana stała, rozdzielona od `KindTyped` — tylko górne
  ograniczenie typu, bez znanej wartości — zgodnie z explicit
  wymogiem planu "rozdzielić dokładnie znaną stałą od jedynie górnego
  ograniczenia typu"; plus `KindAlloc`, `KindNull`, `KindSet`,
  `KindUnknown` z obowiązkowym `Reason`); `internal/valueflow/merge.go`
  (`Merge`/`MergeWithBudget`, `DefaultBudget=4` — kolejność
  poszerzania: identyczne wartości kolapsują do jednej, różne łączą się
  w `KindSet` do budżetu, przekroczenie budżetu **zawsze** poszerza do
  `KindUnknown` z powodem, nigdy nie przycina `Candidates` tak, żeby
  wyglądało na kompletne); `internal/valueflow/facts.go` (`Facts` —
  API faktów dla modeli B3, niezależne od CVE i konkretnego frameworka,
  zgodnie z wymogiem planu).
- **Zachowanie dostępne użytkownikowi: żadne jeszcze.** To jest
  świadomie kontrakt bez implementacji — `cmd/ariadne` nic z tego nie
  wywołuje, `go build -o bin/ariadne` i pełne przebiegi na
  petclinic/WebGoat są **niezmienione i nie wymagały ponownej
  weryfikacji** (sprawdzone: brak zmian w plikach, które wpływałyby na
  pipeline).
- **Spełnione kryteria (z D0's "Gotowe"):** "kontrakt ma przykłady
  gałęzi, pętli, wywołania i nieobsługiwanej operacji" —
  `internal/valueflow/merge_test.go`'s `TestExample_branch` (dwie
  różne stałe z dwóch gałęzi `if` łączą się w `KindSet`),
  `TestExample_loop` (pięć kolejnych scaleń w pętli przekracza budżet
  4 i poszerza się do `KindUnknown` z powodem), `TestExample_call`
  (wynik wywołania zewnętrznego — `KindUnknown`, bo streszczenia
  międzymetodowe to D3, nie D0), `TestExample_unsupportedOperation`
  (`jsr`/`ret`, przykład nazwany explicite w planie, jako
  `bytecodeir.OpUnsupported` + `KindUnknown`). "Nie wymaga ukończenia
  taint" — prawda z konstrukcji: `internal/taint` (D4) nie istnieje i
  nic z D0 go wymaga.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/valueflow`:
  `TestExample_*` [4], `TestMerge_*` [4], `TestUnknown_requiresAReason`,
  `TestFacts_contractIsImplementable`; `internal/bytecodeir` ma zero
  testów — czysty typ/enum bez logiki, D1 doda testy, gdy zacznie
  faktycznie dekodować bytecode do `Site`); `go test -race ./...` dla
  całego projektu (zielone). **Brak przebiegu `bin/ariadne` na
  petclinic/WebGoat — zamierzone, nie przeoczone**: ta faza nie dotyka
  żadnego pliku na drodze CLI→raport, więc taki przebieg
  potwierdziłby tylko to, co `go test ./...` już potwierdza (brak
  regresji), bez żadnego nowego sygnału.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Data flow and
  taint" — **status pozostaje `Not implemented`**, nie podniesiony do
  `Partial`: żaden realny mechanizm pipeline'u się nie zmienił, tylko
  dodano opis istniejących teraz pakietów kontraktowych w kolumnie
  "Available scope" — świadoma decyzja, żeby nie przeszacować zakresu,
  zgodnie z ostrzeżeniem `JVM_GRAPH:G5`).
- **Niespełnione/odłożone:** cała resztę D1–D5 (CFG, lokalne wartości,
  streszczenia międzymetodowe, taint, pomiar/wydanie) — to osobne fazy
  z własnymi zależnościami, nie przeoczenia tej.
- **Następny krok:** `DATA_FLOW:D1` (CFG i semantyka kodu metody —
  rozszerzenie `internal/classfile/code.go` o tabelę wyjątków,
  podstawowe bloki, normalne/wyjątkowe krawędzie) jest naturalnym
  kolejnym krokiem — to pierwsza faza, która faktycznie dotyka
  istniejącego kodu bytecode'owego i dekoduje prawdziwe instrukcje do
  `bytecodeir.Site`. Wymaga propozycji i zgody użytkownika przed
  startem.

### DATA_FLOW:D1 — CFG i semantyka kodu metody — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** D0 (gotowe wyżej). Nie wdrożono
  pełnej symulacji stosu (dup/dup2/swap, obiekty niezainicjalizowane) —
  **odłożone do D2 z jasnym uzasadnieniem**: D1's własny CFG (bloki,
  krawędzie) nie wymaga znajomości zawartości stosu, tylko miejsc
  skoków; D2 ("lokalne wartości") jest fazą, która faktycznie potrzebuje
  symulować stos, żeby wiedzieć co dana instrukcja czyta/zapisuje.
  Uzgodnienie współdzielenia z `JVM_GRAPH:G3` sprawdzone wprost: G3
  nigdy nie dotknęło tabeli wyjątków, więc nie ma nic do współdzielenia
  poza samym faktem, że obie fazy czytają ten sam Code attribute.
- **Zmienione pliki:** `internal/classfile/code.go` (przepisane:
  `parseCodeAttribute` zyskuje drugi parametr `pool` i zwraca strukturalny
  `CodeAttr` — `MaxStack`/`MaxLocals`/`ExceptionTable`/`LineNumbers` —
  zamiast tylko `[]byte`; nowy `parseLineNumberTable`);
  `internal/classfile/classfile.go` (`Method` zyskuje
  `MaxStack`/`MaxLocals`/`ExceptionTable`/`LineNumbers`; `readMethod`
  przekazuje `p` do `parseCodeAttribute`); `internal/classfile/opcodes.go`
  (nowe stałe opcode'ów dla całej rodziny branch/return: `if*`,
  `goto`/`goto_w`, `jsr`/`jsr_w`/`ret`, `*return`, `athrow`,
  `ifnull`/`ifnonnull`); nowy `internal/classfile/cfg.go`
  (`BasicBlock`/`EdgeKind`/`CFGEdge`/`CFG`/`BuildCFG` — prawdziwy
  budowniczy CFG: rozwiązywanie celów skoków dla całej rodziny
  if*/goto/goto_w/tableswitch/lookupswitch zgodnie z JVMS 6.5, z
  poprawnym wyliczeniem paddingu tableswitch/lookupswitch; `jsr`/`ret`
  jawnie oznaczone jako `Unsupported`, nigdy modelowane); nowy
  `internal/classfile/locals.go` (`ArgumentSlots`/`Argument` —
  szerokość slotu zmiennej lokalnej z deskryptora, `long`/`double` = 2
  sloty, zgodnie z explicit wymogiem D1).
- **Zachowanie dostępne użytkownikowi: żadne jeszcze** — `BuildCFG`
  istnieje i jest poprawny, ale nic w `cmd/ariadne`/`internal/reach` go
  nie wywołuje. Potwierdzone bezpośrednio: pełny przebieg `bin/ariadne
  -offline` na petclinic po tej fazie daje **identyczne** liczby
  REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN (3/62/26/51/6) jak przed D1
  — żadnej regresji, bo żadnego nowego zachowania w pipeline.
- **Spełnione kryteria (z D1's "Gotowe"):** "CFG i stany poprawne dla
  gałęzi, pętli, wyjątków i szerokich typów" — cztery dedykowane testy
  (`cfg_test.go`: `TestBuildCFG_branch`, `TestBuildCFG_loop`,
  `TestBuildCFG_exception`; `locals_test.go`:
  `TestArgumentSlots_wideTypes`) na ręcznie zmontowanym bytecode (brak
  JDK, jak w poprzednich fazach tej sesji). "Nie mylić rzuconego
  wyjątku z normalnym wynikiem metody" — `TestBuildCFG_exception`
  explicite sprawdza, że blok kończący się `return` ma *wyłącznie*
  krawędź wyjątkową do handlera, zero krawędzi normalnego przepływu.
  "Stare jsr/ret... oznaczać jako nieobsługiwane" —
  `TestBuildCFG_jsrRetIsUnsupportedNotCrashing`. "Niepoprawne wejścia
  nie powodują paniki ani fikcyjnego braku przepływu" —
  `TestBuildCFG_invalidOpcodeErrorsNotPanics` (błąd, nie panika) +
  **realny test na skalę**: `BuildCFG` uruchomione na **wszystkich
  metodach z kodem w petclinic (217965) i WebGoat (278394)** — **0
  błędów w obu**, w tym poprawnie wykryte 2 (petclinic) i 96 (WebGoat)
  prawdziwych metod z `jsr`/`ret` (stary styl kompilacji `finally`).
  "Zachować opcjonalne numery linii... brak debug data nie zatrzymuje
  analizy" — `TestParseCodeAttribute_noLineNumberTableIsNilNotError`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone — w tym nowe testy w `internal/classfile`:
  `TestParseCodeAttribute_*`, `TestBuildCFG_*` [7], `TestArgumentSlots_*`
  [5]); `go test -race ./internal/classfile/...` (zielone); **jednorazowy
  skrypt diagnostyczny** (napisany, uruchomiony, usunięty) wołający
  `BuildCFG` na każdej metodzie z kodem w realnym petclinic i WebGoat —
  wynik: 0 błędów na ~496k metod łącznie, maks. 451/1210 bloków i
  823/1612 krawędzi w jednej metodzie (odpowiednio); `bin/ariadne
  -offline` na petclinic — identyczne liczby jak przed fazą (różni się
  tylko wersja w nagłówku raportu, bo to inny commit).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Data flow and
  taint": Not implemented → **Partial** — tym razem uzasadnione, bo D1
  dostarcza realny, przetestowany na dużą skalę mechanizm, nie tylko
  kontrakt jak D0; jasno odnotowane, że nic w CLI go jeszcze nie używa).
- **Niespełnione/odłożone:** symulacja stosu (dup/dup2/swap, obiekty
  niezainicjalizowane) — odłożone do D2 z uzasadnieniem wyżej; D2–D5
  nietknięte.
- **Następny krok:** `DATA_FLOW:D2` (lokalne wartości i udostępnienie
  modelom — pierwsza faza, która faktycznie produkuje `valueflow.Value`
  z prawdziwego bytecode, korzystając z `BuildCFG`) odblokowuje
  `SERVICE_LOADING:S1`. Wymaga propozycji i zgody użytkownika przed
  startem.

### DATA_FLOW:D2 — Lokalne wartości i udostępnienie modelom — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** D1 (gotowe wyżej). D2's własny plan
  odkładał symulację stosu do tej fazy — wykonana w pełni, łącznie z
  całą rodziną `dup`/`dup2`/`pop`/`pop2`/`swap` świadomą kategorii
  (JVMS 6.5), nie tylko uproszczonym modelem "jedna wartość = jedno
  słowo" (co dałoby błędne przesunięcia głębokości stosu dla
  `long`/`double`).
- **Zmienione pliki:** nowy `internal/classfile/localops.go`
  (`LocalOp`/`LocalOpKind`/`ConstKind`, `DecodeLocalOps` — dekoduje
  *każdą* instrukcję na jej efekt stosu/lokalnych, rozwiązując
  referencje do stałej tabeli tam, gdzie to potrzebne — `ldc`/`ldc_w`
  stałych tekstowych/klasowych, `getfield`/`getstatic`, `invoke*`,
  `new`/`anewarray`/`multianewarray`, `checkcast`/`instanceof` — bez
  ujawniania `pool` poza pakietem, ta sama zasada co `ExtractCalls`;
  obsługuje `wide` i cały zakres opkodów 0x00–0xc9); nowy
  `internal/classfile/fortesting.go` (`PoolEntry`/`NewClassFileForTesting`
  — jawny, eksportowany odpowiednik `hierarchy.NewForTesting` dla
  `*ClassFile`, żeby testy w `internal/dataflow` mogły budować
  realistyczną tablicę stałych bez pisania binarnego `.class` ręcznie);
  nowy pakiet `internal/dataflow` (`analyze.go`: `Analyze` — worklist
  o punkcie stałym nad blokami D1's CFG, z `valueflow.Merge` na złączeniach
  i jawnym, nigdy zgadywanym, przerwaniem (`Aborted`) przy niezgodności
  głębokości/szerokości stosu albo `jsr`/`ret`; `step.go` — efekt każdej
  z ok. 40 kategorii instrukcji; `dup.go` — cztery formy `dup_x2`/`dup2_x1`/
  `dup2_x2` dokładnie według JVMS 6.5; `facts.go` — `MethodFacts` jako
  realna implementacja `valueflow.Facts`).
- **Zachowanie dostępne użytkownikowi: żadne jeszcze** — `dataflow.Analyze`
  istnieje i jest poprawny, ale `cmd/ariadne`/`internal/reach`/
  `internal/serviceload` go jeszcze nie wywołują (sprawdzone wprost:
  `bin/ariadne -offline` na petclinic/WebGoat po tej fazie nie wymagało
  ponownej weryfikacji liczb, bo żaden plik na ścieżce CLI→raport się
  nie zmienił).
- **Spełnione kryteria (z D2's "Gotowe"):** "literał przez zmienną" —
  `TestAnalyze_literalThroughVariable`: `String s = "a"; take(s);` daje
  `Const("java.lang.String","a")` dokładnie w miejscu wywołania, nie
  rozmyte przez przejście przez zmienną. "Dwa warianty gałęzi" —
  `TestAnalyze_twoBranchVariantsJoin`: `if/else` przypisujący `"a"`/`"b"`
  i wywołujący w punkcie złączenia daje `KindSet{"a","b"}`, nie jedną
  wartość i nie fikcyjne `Unknown`. "Nieznany parametr" —
  `TestAnalyze_unknownParameterStaysUnknown`: parametr `String p` przekazany
  dalej daje `KindTyped("Ljava/lang/String;")`, nigdy zgadaną konkretną
  wartość. "Pierwszy przyrost B2 można wydać bez D4" — prawda z
  konstrukcji: `internal/taint` (D4) nie istnieje i nic z D2 go wymaga.
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  jednorazowy skrypt diagnostyczny (napisany w `cmd/diagtmp`, uruchomiony,
  usunięty) wywołujący `dataflow.Analyze` na *każdej* metodzie z kodem w
  realnym petclinic i WebGoat (łącznie z zależnościami) — **petclinic:
  219077 metod, 0 błędów, 0 panik, 2 przerwane analizy** (te same dwie
  metody z `jsr`/`ret`, które D1's `CFG.Unsupported` już wcześniej
  znalazło); **WebGoat: 280172 metody, 0 błędów, 0 panik, 96 przerwanych
  analiz** — dokładnie 96, identyczna liczba jak D1's własny pomiar
  `jsr`/`ret` na WebGoat, potwierdzająca że jedyny powód przerwania to
  rzeczywiście ta sama, już znana, odnotowana granica, nie nowy błąd.
- **Świadoma decyzja o zakresie:** wynik arytmetyki/konwersji/porównań
  nie jest śledzony (`KindUnknown` z powodem) — D2's własny zapis
  "stałe napisowe/klasowe" ogranicza zakres do stałych tekstowych i
  klasowych, nie liczbowych; `iinc` i każdy binarny/unarny operator
  liczbowy wciąż poprawnie zachowuje kształt stosu (szerokość 1 albo 2
  słowa), tylko nie podaje konkretnej wartości. Wynik wywołania
  (`invoke*`) też zawsze `KindUnknown` — streszczenia międzymetodowe to
  D3, nie D2.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/classfile`:
  `TestDecodeLocalOps_*` [7, w tym cała rodzina dup/pop/swap, `wide`,
  szerokości arytmetyki]; `internal/dataflow`: `TestAnalyze_*` [6]);
  `go test -race ./internal/classfile/... ./internal/dataflow/...` i
  `go test -race ./...` dla całego projektu (zielone); jednorazowy
  skrypt diagnostyczny opisany wyżej (napisany, uruchomiony, usunięty —
  `cmd/diagtmp` nie istnieje w repozytorium); `go build -o bin/ariadne
  ./cmd/ariadne` (zielone, bez zmiany zachowania).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Data flow and
  taint" rozszerzony o D2's pełny zakres, ograniczenia i zmierzony
  dowód na petclinic/WebGoat).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy):** D3
  (streszczenia międzymetodowe), D4 (taint), D5 (pomiar/wydanie) —
  nietknięte. Żaden konsument (`SERVICE_LOADING:S1`, REFLECTION,
  ASYNC_CALLBACKS) jeszcze nie wywołuje `dataflow.Analyze`.
- **Następny krok:** `SERVICE_LOADING:S1` jest teraz odblokowane
  (DATA_FLOW D0–D2: DONE) — rozwiązanie argumentu `Class` przy
  `ServiceLoader.load(...)` przez `Facts.ArgumentAt`, budujące
  `ServiceLoadSite`/pierwszą realną krawędź "może być wywołany przez
  ServiceLoader" (dowody, nie fikcyjna krawędź grafu wywołań, zgodnie z
  S0's własnym zastrzeżeniem). Wymaga propozycji i zgody użytkownika
  przed startem.

### SERVICE_LOADING:S1 — Argument usługi i lokalne użycie loadera — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** S0 (gotowe wyżej), DATA_FLOW D0–D2
  (gotowe wyżej), JVM G2 (gotowe, zarchiwizowane). Rejestr modeli B3 nie
  istnieje jeszcze (nieodnotowane) — zamiast czekać na całą fazę,
  zaimplementowano tylko to, co S1 faktycznie potrzebuje (cztery
  konkretne przeciążenia `ServiceLoader.load`/`loadInstalled`
  rozpoznane po nazwie+deskryptorze), bez budowania ogólnego rejestru.
- **Zmienione pliki:** `internal/serviceload/inventory.go`
  (`ServiceLoadSite` rozszerzone o `LoaderKind`, `PC`,
  `UnresolvedReason` — poprzednio sam stub kontraktu z S0); nowy
  `internal/serviceload/loadsites.go` (`FindLoadSites` — dla każdej
  klasy w hierarchii, której `Method.Calls` (już wyekstrahowane przy
  budowie hierarchii) zawiera wywołanie jednego z czterech znanych
  przeciążeń `ServiceLoader`, plik klasy jest **ponownie odczytany** z
  jara (bo `hierarchy.Build` celowo nie zachowuje `pool`-a potrzebnego
  `DecodeLocalOps`-owi), a `internal/dataflow.Analyze` rozwiązuje
  argument `Class` przez `Facts.ArgumentAt`; `loadSitesInMethod` i
  `anyServiceLoaderCall` wydzielone do bezpośredniego testowania bez
  I/O na jarach); `cmd/ariadne/main.go` (wywołanie `FindLoadSites` po
  S0's `Inventory`, log z liczbą znalezionych/rozwiązanych wywołań);
  `internal/report/report.go`/`report.md.tmpl` (nowa sekcja
  "ServiceLoader.load call sites" — tabela caller/loader/interfejs,
  jawnie opisana jako "nie krawędź grafu wywołań", zgodnie z S0's
  własnym wymogiem).
- **Zachowanie dostępne użytkownikowi:** raport ma teraz tabelę
  wszystkich wywołań `ServiceLoader.load`/`loadInstalled` w aplikacji i
  zależnościach, z rozwiązanym interfejsem usługi tam, gdzie argument
  był dowodliwie stałą `Class`-ową, i jawnym powodem braku rozwiązania
  (typ znany, ale nie konkretna wartość; `KindUnknown`; albo cała
  metoda przerwana przez `dataflow.Analyze`) gdzie nie — **bez żadnej
  nowej krawędzi w grafie wywołań** (S1's własny wymóg zachowany: zero
  zmian w `internal/callgraph`/`internal/reach`, potwierdzone wprost —
  identyczne liczby REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN na
  petclinic i WebGoat przed/po tej fazie). `ServiceLoadSite.LoaderKind`
  odróżnia classpath/jawny loader/`ModuleLayer`/`loadInstalled`, zgodnie
  z S1's własnym wymogiem "nie przyjmować bez dowodu, że dowolny loader
  widzi wszystkie artefakty".
- **Spełnione kryteria (z S1's "Gotowe"):** "dwa loadery dla różnych
  usług nie mieszają dostawców" — każdy `ServiceLoadSite` jest
  samodzielnym faktem przy konkretnym `(metoda, pc)`, nigdy scalanym
  między wywołaniami; potwierdzone na realnym przebiegu petclinic:
  `com.fasterxml.jackson.databind.ObjectMapper#secureGetServiceLoader`
  ma **dwa** wywołania w jednej metodzie (pc=13 `classpath` 1-argumentowe,
  pc=21 `explicit` 2-argumentowe) poprawnie rozróżnione jako dwa
  niezależne site'y. "Niewykorzystany loader nie uruchamia
  konstruktora" — prawda z konstrukcji: S1 nie tworzy żadnej krawędzi
  do konstruktora providera w ogóle (to S2, nietknięte).
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  `bin/ariadne -offline` na petclinic: **29 wywołań `load()`, 13
  rozwiązanych** do prawdziwego interfejsu usługi — np.
  `java.nio.file.spi.FileSystemProvider`,
  `javax.persistence.spi.PersistenceProvider`,
  `javax.validation.spi.ValidationProvider`,
  `javax.websocket.ContainerProvider`, `org.jboss.logging.LoggerProvider`
  — reszta poprawnie nierozwiązana (parametr `Class<T>` przekazywany
  dalej przez bibliotekę, np. Jacksona/Logbacka własne
  `secureGetServiceLoader`, nigdy zgadany). WebGoat: **48 wywołań, 31
  rozwiązanych**. Zero zmian w licznikach REACHABLE/FRAMEWORK/
  NOT_REACHABLE/UNKNOWN na obu aplikacjach względem bazowej wartości
  przed tą fazą (3/62/26/51/6 i 42/44/18/56/0) — szczery wynik: S1 nie
  dotyka reachability, tylko dodaje nową, osobną informację do raportu.
  Dwa przebiegi petclinic dają bajt-identyczny `results.json` i
  `report.md` (poza znacznikiem czasu).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy w `internal/serviceload`:
  `TestLoadSitesInMethod_literalClassResolves`,
  `TestLoadSitesInMethod_unknownArgumentStaysUnresolved`,
  `TestAnyServiceLoaderCall`); `go test -race ./...` dla całego
  projektu (zielone); `go build -o bin/ariadne ./cmd/ariadne` i dwa
  pełne przebiegi `-offline` na petclinic i jeden na WebGoat (opisane
  wyżej).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "ServiceLoader
  and JPMS service models" rozszerzony o S1's zakres i zmierzone
  dowody). `README.md` ("Known limitations" i sekcja 8 "The report"
  rozszerzone o nową sekcję raportu).
- **Niespełnione/odłożone:** S2 (materializacja konstruktora providera
  jako cel grafu), S3 (JPMS `module-info`/`uses`/`provides`), S4 —
  nietknięte. Rejestr modeli B3 (ogólny mechanizm dla wielu torów, nie
  tylko ServiceLoader) wciąż nieodnotowany; S1 go nie zakłada, tylko
  rozpoznaje cztery konkretne przeciążenia po nazwie wprost.
- **Następny krok:** `SERVICE_LOADING:S2` (materializacja i krawędzie
  dispatch — cele konstruktora/fabryki providera jako rzeczywisty cel
  w grafie wywołań, zależny od `JVM_GRAPH:G3` i rejestru B3) albo inny
  obszar R1 (ASYNC_CALLBACKS A0–A2, teraz również odblokowany przez
  DATA_FLOW D0–D2). Wymaga propozycji i zgody użytkownika przed
  startem.

### SERVICE_LOADING:S2 — Materializacja i krawędzie dispatch — DONE for the constructor target; iterator/`Provider.get` API out of scope (2026-10-22)

- **Zakres zlecenia i zależności:** S1 (gotowe wyżej), `JVM_GRAPH:G3`
  (gotowe, zarchiwizowane — wzorzec krawędzi `<clinit>`). Rejestr
  modeli B3 nie istnieje jeszcze (nieodnotowane) — zamiast czekać na
  całą fazę, S2 rozpoznaje tylko jeden, konkretny, jawnie zakodowany
  kształt (konstruktor bezargumentowy publiczny, zgodnie z
  `META-INF/services`'s własnym kontraktem), bez budowania ogólnego
  rejestru modeli.
- **Zmienione pliki:** `internal/serviceload/inventory.go`
  (`ProviderMaterialization` zyskuje `Gap string`; dokumentacja
  poprawiona, żeby nie wspominać "provider() factory" — S2 celowo
  obsługuje wyłącznie klasyczny mechanizm `META-INF/services`'s
  konstruktora, nie Java 9+'s modułowego `provider()`, co jest
  zakresem S3); nowy `internal/serviceload/materialize.go`
  (`FindMaterializations` — łączy `ServiceRegistration` z
  rozwiązanymi `ServiceLoadSite`'ami po interfejsie, materializuje
  wyłącznie zarejestrowanych i *używanych* providerów, po jednym razie
  na parę (interfejs, klasa) nawet przy duplikacie wpisu; sprawdza
  publiczny bezargumentowy konstruktor przez
  `hierarchy.ClassInfo.MethodByNameAndDesc`, z jawną luką (`Gap`) dla
  klasy abstrakcyjnej/interfejsu, braku konstruktora albo konstruktora
  niepublicznego; `EdgeTargets` — łączy site'y z materializacjami w
  konkretne pary (caller, provider) do dodania jako krawędzie);
  `internal/callgraph/graph.go` (`BasisServiceLoader` — nowa, odrębna
  od `BasisSynthetic`, żeby czytelnik raportu odróżnił zgadywany proxy
  profilu od rozwiązanej rejestracji ServiceLoader); nowy
  `internal/callgraph/serviceloader_edges.go`
  (`ApplyServiceLoaderEdges` — dla każdej pary dodaje krawędź do
  `<init>()V` providera (`CallSynthetic`/`BasisServiceLoader`, z
  prawdziwym PC wywołania `ServiceLoader.load` jako dowodem) oraz,
  jeśli istnieje, powiązaną krawędź do `<clinit>()V`
  (`CallInit`/`BasisServiceLoader`) — "związane inicjalizacje" z S2's
  własnego zapisu); `cmd/ariadne/main.go` (wywołanie
  `FindMaterializations`/`EdgeTargets`/`ApplyServiceLoaderEdges` po
  zbudowaniu grafu i profilowych `synthetic_edges`, log z liczbą
  zamodelowanych providerów/dodanych krawędzi);
  `internal/report/report.go`/`report.md.tmpl` (nowa sekcja
  "ServiceLoader provider constructors modeled").
- **Zachowanie dostępne użytkownikowi:** graf wywołań ma teraz
  prawdziwą krawędź od miejsca wywołania `ServiceLoader.load(X.class)`
  do konstruktora każdego zarejestrowanego providera `X`, więc
  reachability BFS (`internal/reach`) może przejść przez nią jak przez
  każdą inną krawędź — bez żadnej specjalnej listy instancji tylko dla
  ServiceLoader (S2's własny wymóg: "nie tworzyć oddzielnej prywatnej
  listy instancji"). Raport ma nową tabelę interfejs/provider/cel
  konstruktora, z jawną luką tam, gdzie konstruktor nie istnieje albo
  nie jest publiczny.
- **Spełnione kryteria (z S2's "Gotowe", częściowo — patrz "Niespełnione"):**
  "pozytywna ścieżka" — `TestFindMaterializations_resolvesPublicNoArgConstructor`
  i `TestApplyServiceLoaderEdges` (realna krawędź w grafie,
  potwierdzona `g.EdgeBetween`, z poprawnym `Kind`/`Basis`/`Sites`).
  "Ujemny przypadek deklaracji bez użycia" —
  `TestFindMaterializations_unusedInterfaceIsNotAGap`: rejestracja,
  której interfejsu żaden rozwiązany site nie nazwał, nie dostaje
  żadnej materializacji. "Brak celu" —
  `TestFindMaterializations_gapCases` (klasa abstrakcyjna, brak
  konstruktora, konstruktor niepubliczny — trzy podtesty, każdy z
  jawną `Gap`, nigdy fikcyjnym celem). "Powtórzony wpis" —
  `TestFindMaterializations_duplicateRegistrationMaterializedOnce`:
  ten sam provider zarejestrowany w dwóch różnych jarach dla tego
  samego interfejsu materializuje się raz. Ponowne zastosowanie tych
  samych krawędzi nie dubluje ich — sprawdzone wprost w
  `TestApplyServiceLoaderEdges`'s ostatnim kroku (zero nowych krawędzi
  przy powtórnym wywołaniu).
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  `bin/ariadne -offline` na petclinic: **8 providerów
  zamodelowanych, 10 krawędzi dodanych** — np.
  `javax.persistence.spi.PersistenceProvider` →
  `org.hibernate.jpa.HibernatePersistenceProvider#<init>()V`,
  `javax.validation.spi.ValidationProvider` →
  `org.hibernate.validator.HibernateValidator#<init>()V`. WebGoat:
  **147 providerów, 158 krawędzi** (różnica 147→158 to krawędzie
  `<clinit>` dla klas, które go mają). **Zero luk (`Gap`) na obu
  aplikacjach** — każdy znaleziony, używany provider miał publiczny
  konstruktor bezargumentowy; mechanizm luk pozostaje więc zweryfikowany
  tylko testami jednostkowymi na tych dwóch kontrolnych aplikacjach,
  nie realnym przykładem — odnotowane szczerze, nie przemilczane.
  **Zero zmian w licznikach REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN**
  na obu aplikacjach (3/62/26/51/6 i 42/44/18/56/0, identyczne jak
  przed S2) — nowe krawędzie do konstruktorów providerów nie prowadzą
  na tej bazie do żadnej metody podatnej na obecnie mapowane CVE,
  szczery wynik, nie oczekiwany z góry. Dwa przebiegi petclinic dają
  bajt-identyczny `results.json` i `report.md` (poza znacznikiem czasu).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/serviceload`:
  `TestFindMaterializations_*` [4], `TestEdgeTargets_*`;
  `internal/callgraph`: `TestApplyServiceLoaderEdges`); `go test -race
  ./...` dla całego projektu (zielone); `go build -o bin/ariadne
  ./cmd/ariadne` i dwa pełne przebiegi `-offline` na petclinic plus
  jeden na WebGoat (opisane wyżej).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "ServiceLoader
  and JPMS service models" rozszerzony o S2's pełny zakres, nowy
  `BasisServiceLoader` i zmierzone dowody). `README.md` ("Known
  limitations", sekcja 8 "The report", i opis `resolution rule" przy
  entry call rozszerzone o `serviceloader`).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy, nie przez
  przeoczenie):** "Obsłużyć część iterator API i `Provider.get`" —
  **nie zaimplementowane w tym przyroście**; S2 modeluje wyłącznie
  konstruktor providera jako cel, nie żadne wywołanie metody na
  zwróconym `ServiceLoader`/`Iterator`/`Provider` — odnotowane jako
  otwarty zakres tej samej fazy, nie nowa faza. "Typy przekazać do
  późniejszego RTA" — nie dotknięte (RTA samo nie istnieje jeszcze,
  GRAPH_PRECISION). Java 9+'s modułowy `provider()` (statyczna
  fabryka) — poza zakresem, to S3 (JPMS). Weryfikacja loadera
  (czy jawnie podany `ClassLoader`/`ModuleLayer` faktycznie widzi dany
  provider) — nie sprawdzana; S2 łączy niezależnie od `LoaderKind`,
  co jest świadomą, odnotowaną niedokładnością (nadmiarowa, nie
  brakująca, ścieżka), nie przemilczanym założeniem.
- **Następny krok:** `SERVICE_LOADING:S3` (JPMS i warstwy —
  `module-info`'s `uses`/`provides`, zależne od `JVM_GRAPH:G4` i
  ewentualnie `DATA_FLOW:D3`) domknęłoby cały obszar SERVICE_LOADING.
  Alternatywnie: `ASYNC_CALLBACKS:A0–A2`, teraz również odblokowane
  przez `DATA_FLOW:D0–D2`. Wymaga propozycji i zgody użytkownika przed
  startem.

### ASYNC_CALLBACKS:A0 — Katalog modeli callbacków — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** B0–B3 (wspólne — rejestr modeli B3
  nie istnieje jeszcze, nieodnotowane; zamiast czekać, A0 koduje
  bezpośrednio mały, jawny katalog, bez budowania ogólnego rejestru,
  ta sama decyzja co SERVICE_LOADING S1 dla `ServiceLoader`'s
  przeciążeń), JVM G0–G2 (gotowe, zarchiwizowane), DATA_FLOW D0–D2
  (gotowe wyżej — choć A0 samo jeszcze nie rozwiązuje żadnego
  argumentu przez `internal/dataflow`; to jest A1's własny zakres).
- **Zmienione pliki:** nowy pakiet `internal/callback`
  (`catalog.go`: `RegistrationModel`/`ExecutionKind`/`DefaultCatalog` —
  `Thread`'s konstruktory przyjmujące `Runnable` i
  `Executor`/`ExecutorService#execute`/`#submit`, z `Lookup` po
  dokładnym `(owner,name,desc)`; `LooksLikeUncatalogedRegistration` —
  diagnostyka "znana nazwa API, nieskatalogowany wariant", sprawdzająca
  też, czy deskryptor faktycznie wspomina `Runnable`/`Callable` jako
  parametr, nie tylko właściciela+nazwę — bez tego każdy konstruktor
  podklasy `Thread` (np. `super(String)`, zero związku z callbackiem)
  dostałby fałszywie dodatnią diagnostykę, co faktycznie wystąpiło i
  zostało złapane przed, nie po, zmierzonym przebiegu (34 fałszywych
  trafień na petclinic przy pierwszej, naiwnej wersji
  owner+name-only — poprawione, zmierzone ponownie: 2); `sites.go`:
  `FindRegistrationSites`/`RegistrationSite`/`Outcome` — inwentaryzacja
  po już wyekstrahowanych `hierarchy.Method.Calls`, bez ponownego
  odczytu bytecode); `internal/report/report.go`/`report.md.tmpl`
  (nowa sekcja "Callback registration call sites"); `cmd/ariadne/main.go`
  (wywołanie `FindRegistrationSites` po S1's load-site'ach, log z
  liczbą znalezionych/niejednoznacznych wywołań).
- **Zachowanie dostępne użytkownikowi:** raport ma nową tabelę każdego
  wywołania rejestrującego callback przez znane API, z interfejsem
  callbacku i charakterem wykonania (`async-maybe`/`immediate-or-async`)
  dla skatalogowanych przeciążeń, i jawną diagnostyką
  `unsupported-api` dla realnych, ale nieskatalogowanych wariantów —
  **bez żadnej krawędzi grafu wywołań i bez rozwiązywania, czym
  faktycznie jest callback** (to A1/A2, jawnie odnotowane jako
  nieukończone w tej samej fazie, nie przemilczane).
- **Spełnione kryteria (z A0's "Gotowe", częściowo — patrz
  "Niespełnione"):** "katalog ma dodatni i ujemny przykład każdego
  pierwszego API" —
  `TestFindRegistrationSites_threadConstructor`/`_executorSubmitCallable`
  (dodatnie, dwa różne interfejsy callbacku) i
  `_uncatalogedOverloadIsFlagged`/`_unrelatedCallIsIgnored`/
  `_threadSubclassSuperCallIsNotFlagged` (ujemne — w tym realny,
  złapany false-positive). "Diagnostyka nieobsługiwanego API" —
  `unsupported-api` zaimplementowane i zweryfikowane na realnym
  bytecode. "Diagnostyka brakującego celu i unknown receivera" —
  **nie zaimplementowane w A0** — wymagają rozwiązanej tożsamości
  callbacku (A1), więc nie są tej fazy własnym zakresem, zgodnie z
  `RegistrationSite`'s/`Outcome`'s własnym komentarzem.
- **Błąd złapany i naprawiony w trakcie pracy, nie przemilczany:**
  pierwsza wersja diagnostyki (`IsKnownAPI(owner,name)`, ignorująca
  deskryptor) oznaczała **34 fałszywe trafienia na petclinic** —
  głównie podklasy `Thread` wołające `super(String)`/`super()`/
  `super(ThreadGroup,String)`, zero związku z `Runnable`. Naprawione
  przez `LooksLikeUncatalogedRegistration`, które dodatkowo sprawdza,
  czy deskryptor wspomina `Runnable;`/`Callable;`; po naprawie: 2
  prawdziwe, bliskie niedopasowania na petclinic (`ExecutorService#submit
  (Runnable, Object)`, `TaskThread`'s 4-argumentowy konstruktor), 6 na
  WebGoat — żadnego fałszywego trafienia w próbce sprawdzonej ręcznie.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy w `internal/callback`:
  `TestFindRegistrationSites_*` [5], `TestLookup_andLooksLikeUncatalogedRegistration`);
  `go test -race ./...` dla całego projektu (zielone); `bin/ariadne
  -offline` na petclinic (138 wywołań, 2 niejednoznaczne) i WebGoat
  (140, 6) — **zero zmian w REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN**
  na obu aplikacjach (ta faza nie dotyka grafu wywołań w ogóle);
  dwa przebiegi petclinic dają bajt-identyczny `results.json` i
  `report.md` (poza znacznikiem czasu).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "General
  callbacks and asynchronous execution models": Not implemented →
  Partial, z pełnym opisem zakresu i zmierzonymi dowodami).
  `README.md` ("Known limitations" — nowy bullet o callbackach).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy):** A1
  (tożsamość zadania — rozwiązanie konkretnego typu `Runnable`/`Callable`
  przez `internal/dataflow` dla `new X()` przekazanego bezpośrednio;
  lambdy/referencje metod zależą od DYNAMIC_LINKAGE L1, nieodnotowane),
  A2 (realna krawędź grafu wywołań do `run`/`call`), A3 (CompletionStage,
  listenery, reactive), A4 — nietknięte.
- **Następny krok:** `ASYNC_CALLBACKS:A1` (tożsamość zadania — zacząć
  od przypadku `new ConcreteRunnable()` przekazanego bezpośrednio do
  skatalogowanego wywołania, z użyciem `internal/dataflow`'s już
  istniejącego `Alloc`; świadomie odłożyć lambdy/referencje metod do
  DYNAMIC_LINKAGE L1) albo dokończenie `SERVICE_LOADING:S3` (JPMS).
  Wymaga propozycji i zgody użytkownika przed startem.

### ASYNC_CALLBACKS:A1 — Tożsamość zadania i funkcji — DONE for direct `new` allocation; lambdy/referencje metod odłożone do DYNAMIC_LINKAGE L1 (2026-10-22)

- **Zakres zlecenia i zależności:** A0 (gotowe wyżej), DATA_FLOW D0–D2
  (gotowe wyżej). DYNAMIC_LINKAGE L1 (wartości funkcyjne — lambdy,
  referencje metod) nie istnieje jeszcze (nieodnotowane) — A1 celowo
  ogranicza się do przypadku "bezpośrednia alokacja `new X(...)`
  przekazana do skatalogowanego wywołania", zgodnie z A1's własnym
  zapisem ("powiązać... wartości funkcyjne z DYNAMIC_LINKAGE L1" —
  czytane jako zależność tej *części* zakresu, nie całej fazy).
- **Zmienione pliki:** `internal/valueflow/value.go` (`Value.Type`
  rozszerzone, by obejmować też `KindAlloc` — konkretny typ alokacji,
  nie tylko miejsce; `Alloc` zyskuje drugi parametr `typ string`);
  `internal/dataflow/step.go` (trzy miejsca produkujące `KindAlloc`
  przekazują teraz `op.TypeName`); nowy `internal/classfile/jarclass.go`
  (`JarClassReader` — wydzielony z `serviceload`'s własnego
  `parseClassFromJar`, bo ta sama potrzeba — ponowny odczyt klasy z
  jara dla jej prawdziwej tablicy stałych — wystąpiła teraz drugi raz;
  `internal/serviceload/loadsites.go` zrefaktoryzowany, by z niego
  korzystać, usuwając duplikat); nowy `internal/callback/identity.go`
  (`ResolveCallbacks`/`ResolvedCallback`/`resolveOne` — dla każdego
  skatalogowanego `RegistrationSite` (A0) odczytuje klasę z jara przez
  `JarClassReader`, woła `internal/dataflow.Analyze`, i odczytuje
  wartość argumentu callbacku przez `Facts.ArgumentAt`);
  `internal/report/report.go`/`report.md.tmpl` (nowa sekcja "Callback
  identities resolved"); `cmd/ariadne/main.go` (wywołanie
  `ResolveCallbacks` po A0's inwentaryzacji, log z liczbą rozwiązanych).
- **Zachowanie dostępne użytkownikowi:** raport ma teraz tabelę
  każdego skatalogowanego wywołania rejestrującego callback z
  rozwiązaną konkretną klasą (gdy argument był bezpośrednią alokacją
  `new`) albo jawnym powodem braku rozwiązania (parametr, pole, wynik
  wywołania, lambda/referencja metody) — **wciąż bez żadnej krawędzi
  grafu wywołań** (to A2).
- **Spełnione kryteria (z A1's "Gotowe", częściowo — patrz
  "Niespełnione"):** "dwa zadania i niewykorzystana lambda nie tworzą
  nieuzasadnionych połączeń" —
  `TestResolveOne_directAllocationResolves` (dodatnie: `new Thread(new
  MyRunnable())` rozwiązuje się do `test.MyRunnable`, nie do samego
  `Thread`a ani do `java.lang.Runnable`) i
  `TestResolveOne_parameterStaysUnresolved` (ujemne: parametr `Runnable`
  przekazany dalej pozostaje nierozwiązany, z jawnym powodem). Lambda
  jako wartość funkcyjna — jawnie odłożona (patrz "Niespełnione"), nie
  fałszywie rozwiązana ani fałszywie oznaczona jako "nieznana" bez
  powodu.
- **Błąd złapany i naprawiony w trakcie pracy, nie przemilczany:**
  pierwszy przebieg na realnym petclinic **spanikował** —
  `ArgumentAt: argument index 0 out of range (call has 0 arguments)`
  dla `LiveReloadServer#<init>(I)V`. Przyczyna ustalona wprost (nie
  zgadywana): `classfile.ExtractCalls` dla `invokedynamic` zapisuje w
  `Call.Owner/Name/Desc` **rozwiązaną implementację referencji metody**
  (`Thread::new` użyte jako `ThreadFactory`), nie kształt samego miejsca
  wywołania — a ta implementacja przypadkiem ma identyczną sygnaturę co
  katalogowy wpis `Thread(Runnable)`. `FindRegistrationSites` (A0)
  błędnie dopasowywał też `classfile.CallDynamic`, tworząc fikcyjny
  `RegistrationSite`, którego PC w rzeczywistości wskazywało na
  wywołanie zero-argumentowe. Naprawione: A0 teraz **pomija każdy wpis
  `CallDynamic`** przy dopasowywaniu do katalogu (dodany test
  regresyjny `TestFindRegistrationSites_methodReferenceImplIsNotACallSite`);
  po naprawie: 137 (nie 138) miejsc na petclinic, zero paniki na obu
  aplikacjach kontrolnych.
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  `bin/ariadne -offline` na petclinic: **53 z 135 skatalogowanych
  miejsc rozwiązanych** do prawdziwej klasy — np.
  `ch.qos.logback.core.net.AbstractSocketAppender$1#run`,
  `com.mysql.cj.jdbc.ConnectionImpl$NetworkTimeoutSetter#run`,
  `io.github.classgraph.ClassGraph$1#run`. WebGoat: **35 z 134**.
  **Zero zmian w REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN** na obu
  aplikacjach (ta faza nie dotyka grafu wywołań). Dwa przebiegi
  petclinic dają bajt-identyczny `results.json` i `report.md` (poza
  znacznikiem czasu).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/callback`:
  `TestResolveOne_*` [2], `TestFindRegistrationSites_methodReferenceImplIsNotACallSite`;
  `internal/valueflow`/`internal/dataflow` bez regresji po zmianie
  sygnatury `Alloc`); `go test -race ./...` dla całego projektu
  (zielone); `bin/ariadne -offline` na petclinic (dwa przebiegi) i
  WebGoat (opisane wyżej, w tym pierwszy przebieg, który spanikował,
  i drugi po naprawie, który nie).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "General
  callbacks and asynchronous execution models" rozszerzony o A1's
  zakres, złapany błąd i zmierzone dowody). `README.md` ("Known
  limitations" — bullet o callbackach zaktualizowany).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy):**
  lambdy/referencje metod (DYNAMIC_LINKAGE L1, nieodnotowane) — wynik
  wywołania `invokedynamic` jest `KindUnknown` z D2's własnego powodu
  ("call return value is not tracked"), nigdy fałszywie rozwiązany;
  wzorzec "podklasa Thread z własnym override `run()`, bez żadnego
  obiektu Runnable" — nie wykryty w ogóle (katalog A0 dopasowuje tylko
  konstruktory przyjmujące `Runnable`); A2 (realna krawędź grafu
  wywołań), A3, A4 — nietknięte.
- **Następny krok:** `ASYNC_CALLBACKS:A2` (podstawowe wykonanie i
  ścieżki — realna krawędź z miejsca rejestracji do `run`/`call` na
  rozwiązanej klasie, zależna od `JVM_GRAPH:G3` dla powiązanej
  inicjalizacji, tym samym wzorcem co `SERVICE_LOADING:S2`) albo
  dokończenie `SERVICE_LOADING:S3` (JPMS). Wymaga propozycji i zgody
  użytkownika przed startem.

### ASYNC_CALLBACKS:A2 — Podstawowe wykonanie i ścieżki — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** A1 (gotowe wyżej). Ten sam wzorzec
  co `SERVICE_LOADING:S2` (realna krawędź "może zostać wywołane", bez
  prywatnej listy instancji, z jawną luką przy nieudanym rozwiązaniu
  celu) — zastosowany tu do pary (rejestracja callbacku, metoda
  `run`/`call`) zamiast (wywołanie `ServiceLoader.load`, konstruktor
  providera).
- **Zmienione pliki:** nowy `internal/callback/edges.go`
  (`FindCallbackEdges` — dla każdego `ResolvedCallback` (A1) z
  rozwiązanym `TargetFQCN` woła `hierarchy.ResolveExact` (zwykłe
  rozwiązywanie metod JVM po łańcuchu nadklas — klasa docelowa może
  tylko *dziedziczyć* `run`, nie nadpisywać go wprost), z jawną `Gap`
  przy niepowodzeniu; `EdgeTargets` — para (caller, deklarująca
  klasa#metoda) gotowa dla `internal/callgraph`); `internal/callgraph/graph.go`
  (nowa `BasisCallback`, odrębna od S2's własnej `BasisServiceLoader` —
  opisana explicite jako "może, nie na pewno" wykonanie, żeby nigdy nie
  sugerować wymyślonej tożsamości wątku); nowy
  `internal/callgraph/callback_edges.go` (`ApplyCallbackEdges`,
  ta sama struktura co `ApplyServiceLoaderEdges`); `cmd/ariadne/main.go`
  (wywołanie `FindCallbackEdges`/`EdgeTargets`/`ApplyCallbackEdges` po
  S2's `ApplyServiceLoaderEdges`); `internal/report/report.go`/
  `report.md.tmpl` (nowa sekcja "Callback call-graph edges").
- **Zachowanie dostępne użytkownikowi:** graf wywołań ma teraz realną
  krawędź od miejsca rejestracji callbacku do konkretnej metody
  `run`/`call`, którą reachability BFS (`internal/reach`) przechodzi
  jak każdą inną — bez wymyślania identyczności wątku (A2's własny
  wymóg "raport nie wymyśla thread identity"), bo krawędź po prostu
  łączy dwie metody, nie niesie żadnej informacji o konkretnym
  wykonaniu w czasie. Raport ma nową tabelę z jawną luką tam, gdzie
  rozwiązana klasa nie implementuje metody callbacku nigdzie w
  analizowanej hierarchii.
- **Spełnione kryteria (z A2's "Gotowe"):** "ścieżka do wywołania w
  run/call" — `TestApplyCallbackEdges` (realna krawędź w grafie,
  potwierdzona `g.EdgeBetween`, poprawny `Kind`/`Basis`/`Sites` z
  prawdziwym PC) i `TestFindCallbackEdges_resolvesDeclaringClass`
  (dziedziczona, nie nadpisana metoda `run` poprawnie rozwiązana do
  rzeczywistej deklarującej klasy, nie do klasy docelowej samej).
  "Rejestracja bez wykonania" — pokryte już przez A1's własny filtr
  (tylko rozwiązane `TargetFQCN` trafiają do A2 w ogóle). "Nieznany
  target" — `TestFindCallbackEdges_unresolvedIdentityContributesNothing`
  i `_unresolvableTargetIsAGap` (jawna luka, nie fikcyjna krawędź).
  Ponowne zastosowanie tych samych krawędzi nie dubluje ich
  (`TestApplyCallbackEdges`'s ostatni krok).
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  `bin/ariadne -offline` na petclinic: **51 z 53 rozwiązanych
  tożsamości zamodelowanych jako krawędzie (47 krawędzi — część
  callerów ma po kilka miejsc do tego samego celu, kolapsujące),
  2 szczere luki**: `org.springframework.util.concurrent.ListenableFutureTask`
  dziedziczy `run()` po `java.util.concurrent.FutureTask` — klasie
  JDK, którą ariadne z zasady nie czyta (poza analizowanym
  uniwersum), nie po błędzie analizy. WebGoat: **33 z 35, 32 krawędzie,
  0 luk.** **Zero zmian w REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN**
  na obu aplikacjach — nowe krawędzie "może wykonać" nie prowadzą na
  tej bazie do żadnej metody podatnej na obecnie mapowane CVE. Dwa
  przebiegi petclinic dają bajt-identyczny `results.json` i
  `report.md` (poza znacznikiem czasu).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/callback`:
  `TestFindCallbackEdges_*` [3]; `internal/callgraph`:
  `TestApplyCallbackEdges`); `go test -race ./...` dla całego
  projektu (zielone); `bin/ariadne -offline` na petclinic (dwa
  przebiegi) i WebGoat (opisane wyżej).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "General
  callbacks and asynchronous execution models" rozszerzony o A2's
  zakres i zmierzone dowody, w tym realny przypadek luki).
  `README.md` ("Known limitations" i opis `resolution rule` przy entry
  call rozszerzone o `callback`).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy):** A0–A2
  to zadeklarowane MVP tego obszaru — **domknięte**. Lambdy/referencje
  metod (DYNAMIC_LINKAGE L1), wzorzec "podklasa Thread z własnym
  override `run()`" (A0's katalog go nie wykrywa), A3 (CompletionStage,
  listenery, reactive), A4 — nietknięte, osobne, nieuzgodnione fazy.
- **Następny krok:** `ASYNC_CALLBACKS:A3` (kontynuacje, listenery,
  reactive — zależne od DATA_FLOW D3 i FRAMEWORK F2/F4, obie
  nieodnotowane) albo `SERVICE_LOADING:S3` (JPMS) albo inny obszar
  R1/R2 z ROADMAP.MD. Wymaga propozycji i zgody użytkownika przed
  startem.

### DYNAMIC_LINKAGE:L0–L1 — Reprezentacja bootstrapów i lambdy/method references — DONE for the creation-site case; SAM-call materialization and bound/virtual references odłożone (2026-10-22)

- **Zakres zlecenia i zależności:** JVM G0–G3 (gotowe, zarchiwizowane),
  DATA_FLOW D0–D2 (gotowe wyżej). Rejestr modeli B3 nie istnieje
  jeszcze (nieodnotowane) — zamiast czekać, L0/L1 rozszerzają
  bezpośrednio istniejące struktury `classfile`/`valueflow`/`dataflow`,
  bez budowania ogólnego rejestru modeli bibliotek. Sesja połączyła L0
  (kontrakt: ekspozycja metadanych bootstrapu) z praktycznym wycinkiem
  L1 (realna wartość `FunctionalValue` z miejsca tworzenia lambdy) w
  jeden przyrost, bo L0 sam w sobie (tylko metadane, bez konsumenta)
  nie dawałby żadnego mierzalnego efektu na realnym kodzie — ten sam
  powód, dla którego wcześniejsze fazy D1/S0/A0 też łączyły "kontrakt"
  z pierwszym realnym zastosowaniem, gdy to miało sens.
- **Zmienione pliki:** `internal/classfile/pool.go` (dopełniona
  tabela `Ref*` z JVMS 5.4.3.5 — `RefGetField`..`RefInvokeVirtual`,
  wcześniej tylko 6–9 były nazwane); `internal/classfile/calls.go`
  (`Call` zyskuje `RefKind byte` — rodzaj uchwytu rozwiązanej
  implementacji; `resolveInvokeDynamic` go wypełnia);
  `internal/classfile/localops.go` (`LocalOp` zyskuje
  `CallRefKind`/`SamInterface`/`SamMethod`/`CallImplDesc` — SAM
  interfejs i metoda czytane prosto z `NameAndType` niezależnie od
  tego, czy bootstrap się rozwiązał; `CallImplDesc` oddzielony od
  `CallDesc`, bo dla `invokedynamic` to dwa różne deskryptory — miejsca
  wywołania vs. rzeczywistej implementacji); `internal/valueflow/value.go`
  (nowy `KindFunctional` + `FunctionalValue{Site, SamInterface,
  SamMethod, ImplOwner/Name/Desc, RefKind, Captured []Value}` —
  "odrębny od gotowej krawędzi wywołania" model L0's własnego zapisu;
  `Equal` dla `KindFunctional` po identyczności `Site`, jak `KindAlloc`);
  `internal/dataflow/step.go` (`stepCall` dla `CallDynamic` z
  rozwiązanym `CallOwner` push'uje `valueflow.Functional(...)` z
  faktycznie przechwyconymi argumentami jako `Captured`, zamiast
  ogólnego `Unknown`; nierozwiązany bootstrap wciąż `Unknown`, ale z
  konkretnym powodem wymieniającym SAM interfejs);
  `internal/callback/identity.go` (`ResolvedCallback` zyskuje
  `ImplOwner/ImplName/ImplDesc`; `resolveOne` rozpoznaje
  `KindFunctional` i rozwiązuje prosto do implementacji, bez
  przechodzenia przez `TargetFQCN`); `internal/callback/edges.go`
  (`FindCallbackEdges` dla `ImplOwner`-owej tożsamości łączy się
  bezpośrednio, **bez** `hierarchy.ResolveExact` — to już dokładna
  metoda deklarująca, nigdy luka); `internal/report/report.md.tmpl`
  (sekcje "Callback identities resolved"/"Callback call-graph edges"
  rozszerzone o przypadek lambda/method reference).
- **Zachowanie dostępne użytkownikowi:** lambda/referencja metody
  przekazana jako argument `Thread`/`Executor` rozwiązuje się teraz do
  swojej rzeczywistej implementacji (statyczna metoda, konstruktor,
  prywatna metoda ciała lambdy) i dostaje realną krawędź w grafie
  wywołań — zamiast zatrzymywać się na "unresolved, needs
  DYNAMIC_LINKAGE L1" (poprzedni stan A1). Przechwycone zmienne wolne
  (captured) są dostępne jako osobne, już rozwiązane wartości, nie
  ponownie spłaszczone do `Unknown`.
- **Spełnione kryteria (z L0's "Gotowe"):** "rodzaj
  REF_invokeVirtual/REF_invokeStatic/REF_newInvokeSpecial jest
  dostępny dla modelu" — pełna tabela `Ref*` + `LocalOp.CallRefKind` +
  `FunctionalValue.RefKind`, potwierdzone testem
  `TestAnalyze_lambdaProducesFunctionalValue` (REF_invokeStatic na
  realnym kształcie metody statycznej jako celu lambdy). "Nieznany
  bootstrap ma diagnostykę przy konkretnym miejscu" — `SamInterface`/
  `SamMethod` dostępne niezależnie od rozwiązania bootstrapu; powód
  `Unknown` dla nierozwiązanego bootstrapu wymienia konkretny SAM
  interfejs/metodę, nie ogólny tekst.
- **Spełnione kryteria (z L1's "Gotowe", częściowo):** "natychmiastowe
  wywołanie [...] i zapis w zmiennej [...] mają oczekiwane różnice" —
  `TestAnalyze_lambdaProducesFunctionalValue`
  (`FunctionalValue` z poprawnym `ImplOwner/Name/Desc`) i
  `TestAnalyze_lambdaCapturesFreeVariable` (przechwycona stała
  dostępna jako osobna wartość `KindConst`, nie zgubiona). "Reference
  do override" — nietknięte w tym przyroście (patrz "Niespełnione").
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  jednorazowy skrypt diagnostyczny (napisany, uruchomiony, usunięty)
  wywołujący `dataflow.Analyze` na każdej metodzie z kodem w realnym
  petclinic (219077) i WebGoat (280172) — **0 błędów, 0 panik** na
  obu, potwierdzające że nowa ścieżka tworzenia `FunctionalValue` jest
  bezpieczna na pełnej skali. Realny efekt na ASYNC_CALLBACKS:
  **petclinic 53→69 z 135 rozwiązanych tożsamości, 51→67 krawędzi
  grafu; WebGoat 35→46 z 134, 33→44 krawędzi** — np.
  `CommonsObjectPool2Metrics#lambda$registerNotificationListener$2`
  rozwiązuje się prosto do
  `#lambda$registerNotificationListener$1` (realny, zagnieżdżony
  łańcuch lambd). **Zero zmian w REACHABLE/FRAMEWORK/NOT_REACHABLE/
  UNKNOWN** na obu aplikacjach — nowe krawędzie nie prowadzą na tej
  bazie do żadnej metody podatnej na obecnie mapowane CVE. Dwa
  przebiegi petclinic dają bajt-identyczny `results.json` i
  `report.md` (poza znacznikiem czasu).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone; poprawiony jeden istniejący test w
  `internal/classfile/classfile_test.go`, który porównywał `Call` przez
  pełną równość struktur i wymagał dopisania nowego pola `RefKind` do
  oczekiwanej wartości — nie regresja, tylko test musiał nadrobić nowe
  pole); nowe testy: `internal/dataflow`:
  `TestAnalyze_lambdaProducesFunctionalValue`,
  `TestAnalyze_lambdaCapturesFreeVariable`; `internal/callback`:
  `TestResolveOne_lambdaResolvesToImplementationMethod`,
  `TestFindCallbackEdges_lambdaSkipsHierarchyResolution`; `go test -race
  ./...` dla całego projektu (zielone); jednorazowy skrypt
  diagnostyczny opisany wyżej; `bin/ariadne -offline` na petclinic (dwa
  przebiegi) i WebGoat.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersze "Lambdas and
  method references" i "General callbacks and asynchronous execution
  models" rozszerzone o nowy zakres i zmierzony efekt). `README.md`
  ("Known limitations" — bullet o callbackach zaktualizowany, lambda
  nie jest już opisywana jako "unresolved").
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy, nie przez
  przeoczenie):** materializacja SAM-call'a jako samodzielny krok L1
  (ogólny konsument wywołujący `samMethod` na dowolnym obiekcie
  funkcyjnym, nie tylko przez ASYNC_CALLBACKS's katalog) — nietknięte;
  wirtualne/bound method references (`obj::method`, REF_invokeVirtual
  z przechwyconym receiverem) — rozpoznawane przez `RefKind`, ale nie
  specjalnie interpretowane (traktowane identycznie jak statyczne);
  L2 (MethodHandle.invoke/invokeExact, Lookup.findStatic/findVirtual) —
  nietknięte; L3 (bindTo/asType, StringConcatFactory, CONSTANT_Dynamic),
  L4 (fixtures wymagające JDK) — nietknięte.
- **Następny krok:** `ASYNC_CALLBACKS:A3` (teraz z lepszym pokryciem
  tożsamości callbacków) albo `DYNAMIC_LINKAGE:L2` (podstawowe
  MethodHandle) albo `SERVICE_LOADING:S3` (JPMS). Wymaga propozycji i
  zgody użytkownika przed startem.

### DYNAMIC_LINKAGE:L2 — Podstawowe MethodHandle (wartość, bez konsumenta) — DONE for the HandleValue model; invoke/invokeExact edge wiring explicitly deferred (2026-10-22)

- **Zakres zlecenia i zależności:** L0–L1 (gotowe wyżej). Zanim
  zaczęto, sprawdzono realne użycie na petclinic/WebGoat
  (jednorazowy skrypt diagnostyczny, uruchomiony i usunięty): oba mają
  prawdziwe wywołania `findStatic`/`findVirtual`/`findConstructor`/
  `invoke`/`invokeExact`/`bindTo` w zależnościach (petclinic: 12
  odrębnych par owner#name; WebGoat: 25, w tym 37×`invokeExact`,
  22×`findStatic`) — to nie jest praca bez realnego pokrycia.
- **Zmienione pliki:** `internal/valueflow/value.go` (nowy
  `KindHandle` + `HandleValue{Site, RefKind, RequestedOwner, Name}` —
  świadomie bez pola na deskryptor/dokładny przeciążenie, zgodnie z
  tym, że żaden inny konsument w projekcie [`hierarchy.ResolveMethod`]
  też nie wymaga dokładnego deskryptora; `RequestedOwner` to dosłowna
  klasa podana do `find*`, nie klasa deklarująca — rozwiązanie tego
  wymaga `hierarchy`, do którego ten pakiet świadomie nie ma dostępu);
  nowy `internal/dataflow/methodhandle.go` (`lookupHandleValue` —
  mały, jawny katalog czterech przeciążeń `find*`, rozpoznawany
  bezpośrednio w `internal/dataflow` zamiast w osobnym pakiecie
  konsumenckim, bo `java.lang.invoke.MethodHandles$Lookup` to część
  rdzennej maszynerii JVM-owej, nie dowolna biblioteka — ten sam
  poziom co `invokedynamic`/LambdaMetafactory już ma w `stepCall`);
  `internal/dataflow/step.go` (`stepCall` push'uje `valueflow.Handle(...)`
  gdy rozpoznany `find*` ma dowodliwie literalne argumenty `Class`+
  `String`, inaczej `Unknown` z konkretnym powodem wymieniającym metodę
  `find*`).
- **Zachowanie dostępne użytkownikowi: żadne jeszcze** — świadomie,
  tak jak DATA_FLOW D0: `HandleValue` istnieje i jest poprawny
  (potwierdzone testami i realnym bytecode), ale nic go nie
  konsumuje. `bin/ariadne -offline` na petclinic po tej fazie daje
  **identyczne** liczby jak przed nią — brak zmiany zachowania, bo
  brak nowego konsumenta w pipeline.
- **Spełnione kryteria (z L2's "Gotowe", częściowo — patrz
  "Niespełnione"):** "stały cel" —
  `TestAnalyze_lookupFindStaticProducesHandleValue` (literalna klasa +
  nazwa metody → poprawny `HandleValue` z właściwym `RefKind`).
  "Nieznany lookup" — `TestAnalyze_lookupFindVirtualWithUnknownNameStaysUnknown`
  (nazwa metody jako parametr, nie literał → `KindUnknown`, nie
  zgadane). "Niezgodny typ" — nie dotyczy tego wycinka (wymaga
  `MethodType`'owego rozwiązania, które `L2` samo odkłada — patrz
  "Niespełnione"). "Sam lookup nie wywołuje ciała celu" — prawda z
  konstrukcji: `HandleValue` nie tworzy żadnej krawędzi grafu, tylko
  wartość.
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  jednorazowy skrypt diagnostyczny (napisany, uruchomiony, usunięty)
  wołający `dataflow.Analyze` na każdej metodzie z kodem w realnym
  petclinic (219077) i WebGoat (280172) — **0 błędów, 0 panik** na
  obu, potwierdzające że nowa ścieżka rozpoznawania `Lookup.find*` jest
  bezpieczna na pełnej skali, nie tylko na syntetycznych testach.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy w `internal/dataflow`:
  `TestAnalyze_lookupFindStaticProducesHandleValue`,
  `TestAnalyze_lookupFindVirtualWithUnknownNameStaysUnknown`); `go test
  -race ./...` dla całego projektu (zielone); jednorazowy skrypt
  diagnostyczny opisany wyżej; `bin/ariadne -offline` na petclinic —
  identyczne liczby jak przed fazą (potwierdzone wprost, nie
  założone).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "General
  reflection, MethodHandle invocation and custom bootstrap models":
  Not implemented → Partial, z pełnym opisem zakresu, ograniczeń i
  zmierzonych dowodów użycia na realnych aplikacjach).
- **Niespełnione/odłożone — wymaga jawnej decyzji przed kontynuacją,
  nie przeoczenie:** połączenie `handle.invoke(...)`/`invokeExact(...)`
  z `HandleValue` jako krawędź grafu wywołań (L2's własny,
  pozostający bez realizacji fragment "Gotowe") wymaga, żeby `Facts`
  (kontrakt D0) ujawniał **receiver** wywołania, nie tylko argumenty —
  `ArgumentAt` z definicji wyklucza receiver. To realna zmiana
  kontraktu D0, nie szczegół implementacyjny, i świadomie nie została
  wykonana bez uzgodnienia: albo nowa metoda `Facts.ReceiverAt(site)`,
  albo inny mechanizm. `MethodType.methodType(...)`'s rozwiązywanie
  (potrzebne, by `invokeExact`'s "niezgodny typ" miał sens) — też
  nietknięte. L3 (`bindTo`/`asType`, `StringConcatFactory`,
  `CONSTANT_Dynamic`), L4 (fixtures wymagające JDK) — nietknięte.
- **Następny krok:** Zdecydować, czy/jak rozszerzyć `Facts` o dostęp
  do receivera (dokończenie L2's własnego zakresu — realne połączenie
  `invoke`/`invokeExact` z grafem), albo przejść do innego obszaru
  (`ASYNC_CALLBACKS:A3`, `SERVICE_LOADING:S3`). Wymaga propozycji i
  zgody użytkownika przed startem.

### DYNAMIC_LINKAGE:L2 — invoke/invokeExact → krawędź grafu (zasięg jednej metody) — DONE, z uzgodnionym rozszerzeniem kontraktu Facts (2026-10-22)

- **Zakres zlecenia i zależności:** L2's własna wartość `HandleValue`
  (gotowa wyżej). Ta faza wymagała **jawnej decyzji użytkownika**
  przed startem — rozszerzenie `valueflow.Facts` (kontrakt DATA_FLOW
  D0) o nową metodę, pierwszej od samego D0 — zgoda uzyskana przed
  implementacją, zgodnie z ustalonym wzorcem sesji.
- **Zmienione pliki:** `internal/valueflow/facts.go` (`Facts` zyskuje
  `ReceiverAt(at bytecodeir.Site) Value` — receiver wywołania
  instancyjnego, ten sam reżim panik/`Aborted` co `ArgumentAt`;
  `facts_test.go`'s `fakeFacts` zaktualizowane, żeby demonstrować
  implementowalność rozszerzonego kontraktu);
  `internal/dataflow/facts.go`/`analyze.go`/`step.go`
  (`MethodFacts.receivers` — nowa mapa, wypełniana w preregistration
  (`Unknown("receiver not yet analyzed")` dla każdego wywołania z
  receiverem) i w `stepCall` (prawdziwa wartość ze stosu, ta sama
  `entries[0]` która wcześniej była odrzucana)); nowy pakiet
  `internal/methodhandle` (`FindInvokeSites` — skan po już
  wyekstrahowanych `Calls` dla `MethodHandle#invoke`/`invokeExact`,
  dopasowanie tylko po owner+name, bo to metody polimorficzne
  sygnaturowo, bez ustalonego deskryptora; `ResolveInvokes` — ponowny
  odczyt klasy przez `classfile.JarClassReader`, `dataflow.Analyze`,
  `Facts.ReceiverAt`; `FindHandleEdges`/`EdgeTargets` — rozwiązanie
  deklarującej klasy przez `hierarchy.ResolveMethod` [po samej nazwie,
  bez deskryptora — ten sam poziom precyzji co CVE's własne
  dopasowanie podatnych metod w całym projekcie]); `internal/callgraph/graph.go`
  (`BasisMethodHandle`); nowy `internal/callgraph/methodhandle_edges.go`
  (`ApplyMethodHandleEdges` — łączy się z **każdym** przeciążeniem
  nazwy celu przez `g.IDsByName`, bo `find*` nigdy nie ustala
  dokładnego deskryptora); `cmd/ariadne/main.go` (wywołanie całego
  łańcucha po `ApplyCallbackEdges`); `internal/report/report.go`/
  `report.md.tmpl` (dwie nowe sekcje: "MethodHandle invoke call sites"
  / "MethodHandle call-graph edges").
- **Zachowanie dostępne użytkownikowi:** raport ma teraz tabelę
  każdego `invoke`/`invokeExact` z rozwiązanym (lub nie, z jawnym
  powodem) uchwytem, i drugą tabelę z realną krawędzią grafu tam,
  gdzie się udało. Krawędź jest oznaczona `methodhandle`
  (`BasisMethodHandle`) w linii "Entry call", odróżniona od
  `callback`/`serviceloader`.
- **Spełnione kryteria (z L2's "Gotowe", w zakresie jednej metody):**
  "różnica posiadanie/invocation" — `HandleValue` sam nie tworzy
  krawędzi (L1-owa faza), krawędź powstaje tylko przy faktycznym
  `invoke`/`invokeExact` (ta faza) —
  `TestAnalyze_receiverAtReadsBackHandleValue`,
  `TestResolveOne_handleReceiverResolves`,
  `TestApplyMethodHandleEdges` (w tym: dwa przeciążenia tej samej
  nazwy obie dostają krawędź, ponowne zastosowanie nie duplikuje).
  "Nieznany lookup" — pokryte już przez L1-ową `HandleValue`'s
  własny `Unknown`, teraz widoczne też na poziomie receivera
  (`TestFindHandleEdges_unresolvedInvokeContributesNothing`).
  "Niezgodny typ" — **nie zaimplementowane w tym przyroście** (patrz
  "Niespełnione").
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  `bin/ariadne -offline` na petclinic: **10 realnych wywołań
  `invoke`/`invokeExact` znalezionych, 0 rozwiązanych do krawędzi** —
  **szczery, zmierzony wynik, nie błąd**: każde z 10 miejsc ładuje
  swój receiver (uchwyt) z pola (`getfield`/`getstatic`, dające
  `KindTyped`, nie `KindHandle`) — dominujący, realny idiom "znajdź
  uchwyt raz, trzymaj go w polu, wywołuj później z innej metody",
  który jest poza zasięgiem analizy ograniczonej do jednej metody
  (potrzebuje DATA_FLOW D3). WebGoat: **46 wywołań, też 0 krawędzi**,
  identyczny wzorzec. **Zero zmian w REACHABLE/FRAMEWORK/
  NOT_REACHABLE/UNKNOWN** na obu aplikacjach. Dwa przebiegi petclinic
  dają bajt-identyczny `results.json`/`report.md`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/dataflow`:
  `TestAnalyze_receiverAtReadsBackHandleValue`; `internal/methodhandle`
  [nowy pakiet]: `TestFindInvokeSites`, `TestFindHandleEdges_*` [3],
  `TestResolveOne_handleReceiverResolves`; `internal/callgraph`:
  `TestApplyMethodHandleEdges`); `go test -race ./...` dla całego
  projektu (zielone); `bin/ariadne -offline` na petclinic (dwa
  przebiegi) i WebGoat (opisane wyżej).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "General
  reflection, MethodHandle invocation and custom bootstrap models"
  rozszerzony o L2's pełny, szczery zakres i zmierzony "zero krawędzi,
  ale dlaczego"). `README.md` ("Known limitations" i opis
  `resolution rule` rozszerzone o `methodhandle`).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy, nie przez
  przeoczenie):** śledzenie uchwytu między metodami (przez pole
  statyczne/instancyjne) — wymaga DATA_FLOW D3 (streszczenia
  międzymetodowe/pól), nieodnotowane; to jest **dominujący** realny
  wzorzec, więc realny efekt tej fazy na dwóch aplikacjach kontrolnych
  jest zerowy, odnotowane szczerze, nie przemilczane. "Niezgodny typ"
  dla `invokeExact` (wymaga rozwiązania `MethodType.methodType(...)`,
  nietknięte). L3 (`bindTo`/`asType`, `StringConcatFactory`,
  `CONSTANT_Dynamic`), L4 (fixtures wymagające JDK) — nietknięte.
- **Następny krok:** `DATA_FLOW:D3` (streszczenia międzymetodowe) —
  jeden przyrost odblokowałby jednocześnie ten dominujący wzorzec
  MethodHandle, ASYNC_CALLBACKS A3's własną zależność, i ogólnie
  poprawiłby precyzję całego projektu. Alternatywnie `SERVICE_LOADING:S3`
  (JPMS) albo `REFLECTION:R0`. Wymaga propozycji i zgody użytkownika
  przed startem.

### FRAMEWORKS:F0 — Model frameworka i scenariusza — DONE (2026-10-22)

- **Zakres zlecenia i zależności:** B0–B1 (wspólne, gotowe wcześniej),
  JVM G0–G2 (gotowe, zarchiwizowane). Rejestr modeli B3 (ogólny
  mechanizm dla wielu torów) nie istnieje jeszcze (nieodnotowane) —
  F0 celowo **nie** budował rejestru; zamiast tego rozszerzył
  bezpośrednio `internal/profile` i zapisał proponowane typy kontraktu
  jako gotowe DO UŻYCIA przez F1/F2, nie jako działający rejestr —
  ta sama decyzja co poprzednie fazy kontraktowe (D0, A0, S0, L0).
- **Zmienione pliki:** nowy `internal/profile/scenario.go`
  (`Condition` — trójstanowy fakt true/false/unknown, zgodnie z F0's
  własnym wymogiem; `Detection`/`DetectAll` — **jedyna część tej fazy
  faktycznie podłączona do pipeline'u**: dla każdego profilu w
  katalogu (poza `generic`) osobno sprawdza, czy jego własne markery
  `detect` są obecne, niezależnie od tego, które profile `Select`
  faktycznie scaliło w aktywny; `FrameworkScenario`/`BeanDefinition`/
  `ConditionalBinding`/`ExternalEntry` — proponowane kontrakty dla
  F1/F2, jawnie nieskonstruowane w tej fazie, z komentarzem
  wyjaśniającym dlaczego); `cmd/ariadne/main.go` (wywołanie
  `DetectAll` przed `Select`, log z listą wykrytych frameworków,
  `header.FrameworksDetected`); `internal/report/report.go`
  (`Header.FrameworksDetected string`); `internal/report/templates/report.md.tmpl`
  (nowa linia nagłówka "Frameworks detected" obok istniejącego
  "Profile").
- **Zachowanie dostępne użytkownikowi:** raport pokazuje teraz
  **dwie** osobne informacje tam, gdzie wcześniej była jedna: "Profile"
  (aktywny, ewentualnie scalony profil) i "Frameworks detected"
  (każdy profil, którego markery faktycznie znaleziono, niezależnie
  od `-profile`). To znaczy: `-profile spring` na petclinic/WebGoat
  teraz jawnie pokazuje "Frameworks detected: spring, servlet" — fakt
  wcześniej niewidoczny, bo wybór `-profile spring` *wymuszał* dokładnie
  jeden aktywny profil, chowając wykrycie `servlet`. Zero zmiany
  zachowania analizy samej (reachability, graf) — F0's własny wymóg
  "obecny profil zachowuje działanie" potwierdzony wprost.
- **Spełnione kryteria (z F0's "Gotowe"):** "obecny profil zachowuje
  działanie" — `go test ./...` bez regresji, `bin/ariadne -offline`
  na petclinic daje identyczne REACHABLE/FRAMEWORK/NOT_REACHABLE/
  UNKNOWN przed/po (3/62/26/51/6), bajt-identyczny `results.json`
  między dwoma przebiegami. "Dostępność biblioteki i aktywna
  konfiguracja są osobnymi informacjami" —
  `TestDetectAll_separatesPresenceFromActiveSelection` (dwa profile
  wykryte i scalone w "a+b" przez `Select`, ale `DetectAll` zwraca
  obie jako niezależne fakty, plus trzeci, nieobecny profil jako
  `Present: false`, nie pominięty); potwierdzone na realnym
  petclinic/WebGoat: `-profile spring` (wymuszony) i `-profile auto`
  (scalający `spring+servlet`) dają **identyczną** listę wykrytych
  frameworków ("spring, servlet"), potwierdzając że wykrycie jest
  rzeczywiście niezależne od wyboru aktywnego profilu.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy w `internal/profile`:
  `TestDetectAll_separatesPresenceFromActiveSelection`,
  `TestDetectAll_excludesGeneric`); `go test -race ./...` dla całego
  projektu (zielone); `bin/ariadne -offline` na petclinic (dwa
  przebiegi, plus trzeci z `-profile auto` dla porównania) i WebGoat.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Profiles and
  selected roots" rozszerzony o F0's zakres i zmierzony dowód).
  `README.md` (nowy akapit w sekcji o profilach, wyjaślniający
  rozdzielenie detekcji od aktywacji).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy, nie przez
  przeoczenie):** `FrameworkScenario`/`BeanDefinition`/
  `ConditionalBinding`/`ExternalEntry` są zdefiniowane, ale **nigdy
  konstruowane** — to F1 (adnotacje/zasoby/jawna konfiguracja) i F2
  (DI/zewnętrzne wejścia), które same zależą od DATA_FLOW D0–D2
  (gotowe) i rejestru B3 (nieodnotowany) odpowiednio. Migracja
  istniejących profili YAML/syntetycznych Spring Data edges do
  nowego kontraktu — nietknięta, YAML-owe profile nadal są
  jedynym mechanizmem niosącym realne zachowanie.
- **Następny krok:** `FRAMEWORKS:F1` (adnotacje z wartościami, zasoby,
  ograniczony odczyt `web.xml`/XML bean definitions) albo
  `DATA_FLOW:D3` (streszczenia międzymetodowe — odblokowałoby
  jednocześnie `DYNAMIC_LINKAGE:L2`'s dominujący wzorzec MethodHandle
  i `FRAMEWORKS:F2`'s własną zależność). Wymaga propozycji i zgody
  użytkownika przed startem.

### DATA_FLOW:D3 — Przepływ między metodami i streszczenia (wycinek: pola statyczne) — DONE for static-field summaries; measured effect on real apps is zero (2026-10-22)

- **Zakres zlecenia i zależności:** D2 (gotowe wyżej), `JVM_GRAPH:G2`
  (gotowe, zarchiwizowane). Rejestr B3 (ogólny mechanizm streszczeń)
  nie istnieje jeszcze (nieodnotowany) — D3 celowo **nie** zbudowało
  rejestru; zamiast tego rozszerzyło bezpośrednio
  `internal/valueflow`/`internal/dataflow` o konkretny, wąski wycinek:
  "zapis do pola statycznego w jednej metodzie, odczyt w innej metodzie
  tej samej klasy" — dokładnie wzorzec zmierzony jako brakujący przy
  weryfikacji `DYNAMIC_LINKAGE:L2`.
- **Zmienione pliki:** `internal/valueflow/value.go` (nowy `KindField` +
  `FieldRef{Owner,Name,Desc}` — `getstatic` identyfikuje teraz *które*
  pole odczytał, zamiast spłaszczać do samego typu deklarowanego;
  `Equal` dla `KindField` po identyczności referencji pola);
  `internal/dataflow/facts.go` (`MethodFacts.fieldWrites`,
  `FieldValueAt` — **metoda `MethodFacts`, nie część interfejsu
  `valueflow.Facts`**, bo nic w projekcie nie konsumuje `Facts` przez
  typ interfejsowy — ten sam poziom co istniejące `Aborted()`, żeby nie
  rozszerzać kontraktu bez realnej potrzeby); `internal/dataflow/step.go`
  (`LocalOpFieldGet` dla pola statycznego push'uje `valueflow.Field(...)`
  zamiast `Typed`; `LocalOpFieldPut` dla pola statycznego zapisuje
  wartość do `fieldWrites` przed odrzuceniem ze stosu — pola
  instancyjne (`getfield`/`putfield`) **celowo nietknięte**, zostają
  `Typed` jak wcześniej); `internal/dataflow/analyze.go`
  (preregistration dla `fieldWrites`); nowy `internal/dataflow/fields.go`
  (`ResolveStaticField` — dla danej `FieldRef`, `hierarchy.ResolveField`
  znajduje deklarującą klasę, ponownie odczytuje ją przez
  `JarClassReader`, i `mergeFieldWriters` [wydzielona, testowalna bez
  I/O na jarach, ten sam wzorzec co `resolveOne`/`ResolveCallbacks`]
  skanuje **tylko własne metody deklarującej klasy** po
  `InitRefs`'owych wpisach `InitPutStatic` [już wyekstrahowane przez
  `JVM_GRAPH:G3`, zero nowego skanu bytecode'u], analizuje każdego
  pisarza, i scala wszystkie zapisane wartości przez `valueflow.Merge`);
  `internal/methodhandle/resolve.go` (`resolveOne` rozpoznaje
  `KindField` i woła `dataflow.ResolveStaticField` jako drugą szansę,
  zanim ostatecznie uzna receiver za nierozwiązany).
- **Zachowanie dostępne użytkownikowi:** mechanizm istnieje i działa
  (potwierdzone testami z realnym scalaniem wielu pisarzy), ale **nie
  zmienia niczego widocznego na petclinic/WebGoat** — patrz "Dowód".
- **Spełnione kryteria (z D3's "Gotowe", częściowo — tylko wycinek pól
  statycznych):** "fabryka zwraca właściwą wartość" (tu: pole, nie
  fabryka) — `TestMergeFieldWriters_singleWriterResolves` (jeden
  pisarz, wartość poprawnie scalona i dostępna). "Rekurencja zbiega
  lub ujawnia limit" — nie dotyczy tego wycinka (pola statyczne nie
  rekursją; `valueflow.Merge` i tak ma własny budżet z poprzednich
  faz). "Dwa call sites nie tworzą nieuzasadnionej ścieżki" —
  `TestMergeFieldWriters_noWriterIsHonestUnknown` (brak pisarza →
  `KindUnknown`, nigdy fikcyjna wartość) i
  `TestResolveOne_fieldReceiverFallsThroughToResolveStaticField`
  (nierozwiązywalny właściciel → `UnresolvedReason`, bez paniki, bez
  dotykania jara).
- **Dowód na realnym bytecode, nie tylko testach jednostkowych —
  szczery, zmierzony wynik, nie taki jak oczekiwano:** `bin/ariadne
  -offline` na petclinic i WebGoat po tej fazie daje **identyczne 0 z
  10/46** rozwiązanych tożsamości `MethodHandle.invoke` jak przed D3
  (`DYNAMIC_LINKAGE:L2`'s własny pomiar). Sprawdzone wprost, które z
  10/46 realnych receiverów to teraz: **6/22 pole instancyjne**
  (`getfield`, celowo poza zakresem tego wycinka), **3/24 wynik innego
  wywołania** (zwykła propagacja międzymetodowa wyniku, D3's własny
  nieukończony punkt "parametr → wynik"), **1 nierozpoznany bootstrap**
  — **żaden z 10/46 nie używał pola statycznego**, czyli wzorzec, który
  ta faza faktycznie obsługuje. Mechanizm jest realny i przetestowany
  (ogólny wzorzec "`private static final X = ...;` ustawione raz,
  odczytane gdzie indziej" jest bardzo częsty w Javie w ogóle — po
  prostu nie w TYCH konkretnych 10/46 miejscach na tych dwóch
  aplikacjach kontrolnych dla MethodHandle). **Zero zmian w
  REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN** na obu aplikacjach,
  bajt-identyczny `results.json`/`report.md` między dwoma przebiegami
  petclinic. Czas przebiegu petclinic: ~9.6s (bez regresji
  wydajnościowej odczuwalnej z poziomu CLI).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/dataflow`:
  `TestAnalyze_staticFieldReadProducesKindField`,
  `TestAnalyze_staticFieldWriteRecordsFieldValueAt`,
  `TestMergeFieldWriters_*` [2], `TestResolveStaticField_*`;
  `internal/methodhandle`:
  `TestResolveOne_fieldReceiverFallsThroughToResolveStaticField`); `go
  test -race ./...` dla całego projektu (zielone); `bin/ariadne
  -offline` na petclinic (dwa przebiegi) i WebGoat (opisane wyżej).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersze "Data flow and
  taint" i "General reflection, MethodHandle invocation..." oba
  rozszerzone o D3's zakres **i** szczery, zerowy zmierzony efekt na
  realnych aplikacjach — żadne z dwóch nie przedstawia tego jako
  sukcesu większego niż jest).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy, nie przez
  przeoczenie):** pola instancyjne (`getfield`/`putfield`) — wymagają
  wiązania konkretnej instancji z miejscem jej utworzenia (prawdziwe
  points-to, znacznie trudniejsze niż jedno niewątpliwe miejsce dla
  pola statycznego), nietknięte. Propagacja wyniku przez zwykłe
  wywołanie metody ("parametr → wynik", D3's własny drugi punkt) —
  nietknięta, zwykły wynik wywołania to wciąż zawsze `KindUnknown`.
  Klucz kontekstu/instancji, wersjonowanie/unieważnianie streszczeń po
  zmianie kodu (CVE F3's snapshoty) — nietknięte, nie potrzebne przy
  tak wąskim wycinku. Rejestr B3 — wciąż nieodnotowany.
- **Następny krok:** Rozszerzenie D3 na pola instancyjne (bezpośrednio
  adresuje 6/10 i 22/46 realnych przypadków zmierzonych tutaj) albo
  propagacja wyniku zwykłego wywołania (3/10 i 24/46) — które z tych
  dwóch faktycznie zamknęłoby zmierzony brak na petclinic/WebGoat.
  Alternatywnie `FRAMEWORKS:F1` albo `ASYNC_CALLBACKS:A3`. Wymaga
  propozycji i zgody użytkownika przed startem.

### DATA_FLOW:D3 — Przepływ między metodami i streszczenia (wycinek: pola instancyjne, final, konstruktor) — DONE for this narrow shape; measured effect on real apps is zero (2026-10-23)

- **Zakres zlecenia i zależności:** D3's poprzedni wpis (pola statyczne,
  gotowy wyżej). Zgoda użytkownika na konkretny, wąski wycinek: tylko
  pole **final**, ustawiane wyłącznie przez `this.field = ...` w
  konstruktorze (`<init>`) własnej deklarującej klasy — nie pełne
  points-to, nie każdy `putfield` w klasie (druga, szersza opcja była
  przedstawiona i odrzucona przez użytkownika na rzecz tej węższej).
- **Zmienione pliki:** `internal/valueflow/value.go` (doc comments
  `KindField`/`FieldRef` rozszerzone: "field" obejmuje teraz zarówno
  static jak i instance, nic w samej wartości nie mówi które — to
  decyduje konsument z dostępem do hierarchii); `internal/dataflow/step.go`
  (`LocalOpFieldGet` ujednolicony: `getfield` produkuje teraz
  `valueflow.Field(...)` tak jak `getstatic`, bez rozróżnienia na
  static/instance w tym miejscu; `LocalOpFieldPut` dla instancji
  zapisuje wartość do `fieldWrites` **tylko** gdy metoda to `<init>` i
  objectref to dokładnie `this` — `entries[0].v` równe
  `valueflow.Typed(it.fqcn)`, jedynej wartości jaką slot 0 może mieć w
  konstruktorze, nigdy nieprzypisywanej ponownie przez poprawny
  bytecode); nowy `internal/dataflow/fields.go`'s `ResolveInstanceField`
  (sprawdza `AccessFlags&AccFinal` przez `ClassInfo.FieldByNameAndDesc`
  przed czymkolwiek innym, potem skanuje wyłącznie metody `<init>`),
  `mergeConstructorFieldWriters` (ten sam merge co pola statyczne, ale
  czyta `facts.fieldWrites` **bezpośrednio** z pakietu, nie przez
  eksportowane, panikujące `FieldValueAt` — miejsce nieznalezione tu
  oznacza "nie self-write", nie błąd wołającego), `fieldPutPCs`
  (dekoduje `LocalOps` konstruktora, bo `putfield`/`getfield` nigdy nie
  wyzwalają inicjalizacji klasy — JVM_GRAPH G3's `InitRef` nigdy ich nie
  wyodrębnił, więc nie ma gotowej listy kandydatów do ponownego użycia,
  w odróżnieniu od pól statycznych); nowy `ResolveField` — **jedyny
  publiczny punkt wejścia** dla odbiornika `KindField`: rozwiązuje
  deklarującą klasę i realne flagi dostępu **raz**, potem woła
  `ResolveStaticField` albo `ResolveInstanceField` — nigdy obu po
  kolei; `internal/methodhandle/resolve.go` (`resolveOne` woła teraz
  `dataflow.ResolveField` zamiast wcześniejszego samego
  `ResolveStaticField`).
- **Błąd złapany i naprawiony w trakcie pracy, nie przemilczany:**
  pierwsza wersja `resolveOne` próbowała `ResolveStaticField`, a przy
  `KindUnknown` **zawsze** próbowała `ResolveInstanceField` jako "drugą
  szansę" — na realnym przebiegu petclinic to ujawniło się jako
  realny, mylący wynik: pole `io.micrometer...HazelcastIMapAdapter#GET_NAME`
  jest w rzeczywistości **statyczne i final**, ale jego `invoke`'owy
  odbiornik i tak zostaje nierozwiązany (bo wartość uchwytu gubi się w
  nieśledzonym wywołaniu między `getstatic` a `invoke`) — przez "próbuj
  oba" raport pokazywał błędny powód "nie znaleziono zapisu w
  konstruktorze", sugerujący pole instancyjne, dla pola które w ogóle
  nie jest instancyjne. Naprawione jednym dispatcherem (`ResolveField`)
  decydującym raz, po realnych `AccessFlags`, zamiast prób/błędów —
  teraz ten sam przypadek poprawnie zgłasza "the receiver is unknown:
  call return value is not tracked by this analysis increment".
- **Zachowanie dostępne użytkownikowi:** mechanizm istnieje i działa
  (testy z realnym scalaniem, w tym wielu konstruktorów), ale **nie
  zmienia niczego widocznego na petclinic/WebGoat** w sensie nowych
  krawędzi — patrz "Dowód". Zmienia natomiast **dokładność powodu**
  zgłaszanego w raporcie dla nierozwiązanych odbiorników (patrz błąd
  wyżej) — to jest realna, pozytywna zmiana zachowania, nawet bez
  nowych krawędzi.
- **Spełnione kryteria:** "fabryka/pole zwraca właściwą wartość" —
  `TestMergeConstructorFieldWriters_singleWriterResolves`. "Dwa call
  sites nie tworzą nieuzasadnionej ścieżki" —
  `TestMergeConstructorFieldWriters_noWriterIsHonestUnknown`,
  `TestMergeConstructorFieldWriters_nonConstructorWriterIgnored`
  (zapis poza konstruktorem jawnie ignorowany, nie scalany).
  "Nie zgadywać nie-final" — `TestResolveInstanceField_nonFinalFieldIsUnknown`.
  "Dispatcher nie maskuje prawdziwego powodu" —
  `TestResolveField_dispatchesByAccessFlagsNotTrialAndError` (pole
  statyczne i instancyjne dają **różne** powody `Unknown`, dowodząc że
  faktycznie różne gałęzie się wykonały, nie jedna zamaskowana drugą).
  "Self-receiver, nie dowolny obiekt" —
  `TestAnalyze_instanceFieldSelfWriteInConstructorRecordsFieldValueAt`
  (zapis przez `this` w konstruktorze) i
  `TestAnalyze_instanceFieldWriteOutsideConstructorNotRecorded` (ten
  sam zapis, ale poza konstruktorem — nie zapisany).
- **Dowód na realnym bytecode, nie tylko testach jednostkowych —
  szczery, zmierzony wynik:** `bin/ariadne -offline` na petclinic i
  WebGoat po tej fazie daje **identyczne 0 z 10/46** rozwiązanych
  tożsamości `MethodHandle.invoke` jak przed tym przyrostem. Ale
  dzięki naprawie dispatchera, powód jest teraz precyzyjny dla każdego
  z 10/46 miejsc:
  - **petclinic (10):** 6 "call return value not tracked" (zwykła
    propagacja wyniku wywołania — D3's własny, wciąż nieukończony drugi
    punkt), 1 "typed", 1 pole statyczne (`GET_NAME`) którego uchwyt
    gubi się w nieśledzonym wywołaniu między odczytem pola i `invoke`
    (poprzednio błędnie zgłaszane jako problem pola instancyjnego —
    patrz "błąd naprawiony" wyżej), 2 inne nierozwiązane lookupy.
    **Zero realnych pól instancyjnych final+konstruktor w tym zbiorze.**
  - **WebGoat (46):** 30 "call return value not tracked", **10
    prawdziwych pól instancyjnych** (`org.jruby.internal.runtime.methods.CompiledIRMethod#specific`,
    `DynamicMethod#handle`) — poprawnie zidentyfikowanych jako pola
    instancyjne, ale **nie-final** (JRuby cache'uje uchwyty leniwie,
    nadpisując pole) — ten przyrost **poprawnie odmawia** ich
    rozwiązania, zamiast zgadywać; to jest właściwość bezpieczeństwa
    tej fazy działająca jak zaprojektowano, nie niedociągnięcie. 6
    "typed".
  - **Zero zmian w REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN** na obu
    aplikacjach (petclinic 3/62/26/51/6, WebGoat 42/44/18/56/0 —
    identyczne z bazową wartością przed tą fazą). Dwa przebiegi
    petclinic i dwa przebiegi WebGoat dają bajt-identyczny
    `results.json`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy w `internal/dataflow`:
  `instance_fields_test.go` — `TestAnalyze_instanceFieldReadProducesKindField`,
  `TestAnalyze_instanceFieldSelfWriteInConstructorRecordsFieldValueAt`,
  `TestAnalyze_instanceFieldWriteOutsideConstructorNotRecorded`,
  `TestMergeConstructorFieldWriters_*` [3], `TestResolveInstanceField_*` [2],
  `TestResolveField_dispatchesByAccessFlagsNotTrialAndError`); `go test
  -race ./...` dla całego projektu (zielone); `bin/ariadne -offline` na
  petclinic (dwa przebiegi) i WebGoat (dwa przebiegi), opisane wyżej;
  jednorazowy diagnostyczny program (`tmp_diag`, w katalogu repo bo
  pakiety `internal/*` tego wymagają — uruchomiony i usunięty)
  potwierdzający realne flagi dostępu i miejsca zapisu pola
  `HazelcastIMapAdapter#GET_NAME` przed naprawą dispatchera.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersze "Data flow and
  taint" i "General reflection, MethodHandle invocation..." oba
  rozszerzone o ten wycinek, naprawiony błąd dispatchera, i pełny,
  precyzyjny rozkład 10/46 miejsc). `README.md` (akapit o `MethodHandle`
  w "Known limitations" rozszerzony o pola instancyjne final+konstruktor
  i zmierzony wynik).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy, nie przez
  przeoczenie):** pola instancyjne nie-final (prawdziwy points-to,
  potrzebny dla WebGoat's JRuby 10/46) — nietknięte, zgodnie z tym co
  faza explicite wyłączyła. Propagacja wyniku zwykłego wywołania
  ("parametr → wynik") — wciąż **dominujący** pozostały brak na obu
  aplikacjach (6/10, 30/46) — nietknięta. Klucz kontekstu/instancji,
  wersjonowanie streszczeń (CVE F3) — nietknięte. Rejestr B3 — wciąż
  nieodnotowany.
- **Następny krok:** Propagacja wyniku zwykłego wywołania metody
  (adresuje bezpośrednio największy zmierzony pozostały brak: 6/10 i
  30/46) — naturalny kolejny krok tego samego wątku DATA_FLOW/DYNAMIC_LINKAGE.
  Alternatywnie `FRAMEWORKS:F1` albo `ASYNC_CALLBACKS:A3`. Wymaga
  propozycji i zgody użytkownika przed startem.

### DATA_FLOW:D3 — Przepływ między metodami i streszczenia (wycinek: wynik wywołania invokestatic/invokespecial + wspólne łączenie pole↔wywołanie) — DONE for this shape; measured effect on real apps is zero, diagnoza znacznie głębsza (2026-10-23)

- **Zakres zlecenia i zależności:** D3's poprzednie dwa wpisy (pola
  statyczne i instancyjne, gotowe wyżej). Zgoda użytkownika: kontynuacja
  tego samego wątku, adresująca zmierzony największy pozostały brak
  (propagacja wyniku zwykłego wywołania). Celowo wyłączone: wywołania
  virtual/interface (cel zależy od typu w czasie działania — CHA — więc
  streszczenie tutaj byłoby zgadywaniem jednego override'u; tylko
  `invokestatic`/`invokespecial`, dokładnie rozwiązane już samym
  bytecode).
- **Zmienione pliki:** `internal/valueflow/value.go` (`KindCallResult` +
  `CallResultRef{Owner,Name,Desc}`, konstruktor `CallResult`, przypadek w
  `Equal`); `internal/dataflow/facts.go` (`MethodFacts.returns` +
  `returnOrder` — lista Site'ów w kolejności PC, bo iteracja po mapie w
  Go jest losowa, a projekt wymaga bajt-identycznych przebiegów);
  `internal/dataflow/analyze.go` (preregistracja każdego zwrotu z
  wartością, tak jak już robiono dla wywołań/zapisów pól);
  `internal/dataflow/step.go` (`LocalOpReturn` zapisuje faktycznie
  zwróconą wartość; `stepCall` dla `CallStatic`/`CallSpecial` z
  `op.Width > 0` produkuje teraz `valueflow.CallResult(...)` zamiast
  gołego `Unknown`); nowy `internal/dataflow/callresults.go`
  (`ResolveCallResult` — rozwiązuje dokładną deklarującą klasę przez
  `hierarchy.ResolveExact` [to samo G2 współdzieli z CVE F4], czyta jar,
  `mergeReturns` scala wszystkie zwroty w deterministycznej kolejności
  PC; `chaseFieldOrCall` — **wspólne, bezpieczne pod cyklami** łączenie
  pól i wyników wywołań: jeśli rozwiązana wartość pola jest sama
  `KindCallResult` [`private static final MethodHandle H = findIt();`]
  albo rozwiązany wynik wywołania jest sam `KindField` [getter zwracający
  scache'owane pole], chase idzie o jeden krok dalej, z jednym wspólnym
  `visited` dla obu rodzajów kluczy, żeby łańcuch przeplatający pola i
  wywołania nie zapętlił się w nieskończoność); `internal/dataflow/fields.go`
  (`ResolveField` i `ResolveCallResult`'s publiczne funkcje teraz
  wywołują `chaseFieldOrCall` po pierwszym rozwiązaniu, zamiast zwracać
  surowy, częściowo rozwiązany wynik; `resolveFieldOnce`/
  `resolveCallResultOnce` to nowe, nieeksportowane "jeden krok bez
  chase'owania" rdzenie, do których chase wraca); `internal/methodhandle/resolve.go`
  (`resolveOne` rozszerzone o `case valueflow.KindCallResult` wołające
  `dataflow.ResolveCallResult`, obok istniejącego `KindField`).
- **Zachowanie dostępne użytkownikowi:** mechanizm istnieje, działa, i —
  w przeciwieństwie do obu poprzednich wycinków D3 — **widocznie
  pogłębia realną diagnozę na petclinic**: `HazelcastIMapAdapter#GET_NAME`
  (pole statyczne, którego powód niepowodzenia wcześniej brzmiał "call
  return value is not tracked") teraz rozwiązuje się w pełni przez
  pole→wywołanie→`findVirtual`, docierając do **prawdziwej przyczyny
  źródłowej**: "MethodHandles.Lookup#findVirtual target is not
  statically known (Class/name argument not a literal)" — realne,
  różne ograniczenie, nie poprzednia, powierzchowna wiadomość.
- **Spełnione kryteria:** "fabryka/wywołanie zwraca właściwą wartość" —
  `TestMergeReturns_singleReturnResolves`. "Metoda void/bez zwrotu
  wartości daje szczere Unknown" — `TestMergeReturns_voidMethodIsHonestUnknown`.
  "Rekurencja zbiega lub ujawnia limit" —
  `TestChaseFieldOrCall_cycleDoesNotHang` (cykl wykryty, Unknown, bez
  zapętlenia). "Łączenie pole↔wywołanie faktycznie się odbywa, nie jest
  tylko zadeklarowane" — `TestChaseFieldOrCall_fieldToCallResultChases`
  (dowodzi, że `chaseFieldOrCall` dysponuje `KindCallResult`, nie
  zwraca go bez zmian). "Tylko invokestatic/invokespecial, nie
  virtual/interface" — `TestAnalyze_staticCallResultProducesKindCallResult`
  (bezpośrednio na realnym bytecode `invokestatic`).
- **Dowód na realnym bytecode, nie tylko testach jednostkowych —
  szczery, zmierzony wynik:** `bin/ariadne -offline` na petclinic i
  WebGoat po tej fazie daje **identyczne 0 z 10/46** rozwiązanych
  tożsamości `MethodHandle.invoke` jak przed tym przyrostem — ale
  diagnoza jest teraz znacznie głębsza:
  - **petclinic:** jedyne pole-oparte miejsce (`GET_NAME`) teraz
    rozwiązuje się w pełni przez 2 skoki (pole→wywołanie→lookup) do
    swojej prawdziwej przyczyny (opisanej wyżej) — wcześniej kończyło
    się na płytszym "no writer found"/"call return value not tracked".
    Jedno miejsce (`JSONUtils$1#run`) zmieniło powód z "call return
    value not tracked" na "the receiver is set" (bogatszy, bardziej
    precyzyjny `KindSet` zamiast gołego `Unknown`) — wciąż
    nierozwiązane, ale z lepszą informacją.
  - **WebGoat:** dwa nowe, realne wzorce ujawnione: wywołania do
    `java.lang.invoke.MethodHandles#foldArguments/insertArguments/guardWithTest`
    (kombinatory MethodHandle) poprawnie zgłaszane jako nierozwiązane,
    bo `java.lang.invoke.MethodHandles` to klasa JDK nigdy nie czytana
    (poza analizowanym uniwersum) — nie defekt tej fazy. 10 prawdziwych
    pól instancyjnych JRuby (`CompiledIRMethod#specific`,
    `DynamicMethod#handle`, `HandleMethod#target0-4`) wciąż poprawnie
    odrzuconych jako nie-final (bez zmian względem poprzedniej fazy).
    6 miejsc ujawniło nowy, szczery powód: "local variable value before
    the exception that entered this handler is not tracked" (D1's
    własne, wcześniej istniejące ograniczenie obsługi wyjątków, teraz
    widoczne jako finalna przyczyna po głębszym chase'u, nie nowy błąd).
  - **Zero zmian w REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN** na obu
    aplikacjach (petclinic 3/62/26/51/6, WebGoat 42/44/18/56/0 —
    identyczne z bazową wartością). Dwa przebiegi petclinic i dwa
    przebiegi WebGoat dają bajt-identyczny `results.json`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy w `internal/dataflow`:
  `callresults_test.go` — `TestAnalyze_staticCallResultProducesKindCallResult`,
  `TestMergeReturns_singleReturnResolves`,
  `TestMergeReturns_voidMethodIsHonestUnknown`,
  `TestResolveCallResult_unresolvableOwnerIsUnknown`,
  `TestChaseFieldOrCall_cycleDoesNotHang`,
  `TestChaseFieldOrCall_fieldToCallResultChases`); `go test -race
  ./...` dla całego projektu (zielone); `bin/ariadne -offline` na
  petclinic (dwa przebiegi) i WebGoat (dwa przebiegi), opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersze "Data flow and
  taint" i "General reflection, MethodHandle invocation..." oba
  rozszerzone o ten wycinek i pełny, głębszy rozkład diagnoz).
  `README.md` (akapit o `MethodHandle` w "Known limitations"
  rozszerzony o wywołania bezpośrednie i wspólny chase).
- **Niespełnione/odłożone (zgodnie z własnym zakresem fazy, nie przez
  przeoczenie):** wywołania virtual/interface — nietknięte z definicji
  (CHA-owa niejednoznaczność celu). Pola instancyjne nie-final —
  nietknięte (prawdziwy points-to). Klasy JDK poza analizowanym
  uniwersum (np. `java.lang.invoke.MethodHandles` samo) — strukturalnie
  nieosiągalne bez ich wczytania, nie w zakresie tej fazy. Obsługa
  wyjątków na poziomie D1 (stan lokalny przed handlerem) —
  nietknięta, osobne ograniczenie. Klucz kontekstu/instancji,
  wersjonowanie streszczeń (CVE F3), rejestr B3 — nietknięte.
- **Następny krok:** Zmierzony brak na obu aplikacjach kontrolnych jest
  teraz wyłącznie: (1) pola instancyjne nie-final (potrzebuje
  points-to — większy, osobny temat), (2) klasy JDK poza uniwersum
  (nie do naprawienia bez wczytania JDK), (3) D1's obsługa wyjątków.
  Żadne z tych trzech nie jest małym, mechanicznym kolejnym krokiem w
  tym samym stylu co D3's trzy dotychczasowe wycinki — to dobry
  moment, żeby rozważyć inny obszar: `FRAMEWORKS:F1` albo
  `ASYNC_CALLBACKS:A3`. Wymaga propozycji i zgody użytkownika przed
  startem.

### FRAMEWORKS:F1 — Adnotacje, zasoby i jawna konfiguracja (wycinek: wartości adnotacji + realne wartości Spring MVC) — DONE for this scope; web.xml/XML bean definitions/config priority deferred (2026-10-23)

- **Zakres zlecenia i zależności:** F0 (gotowe wyżej), JVM G1/G4
  (gotowe, zarchiwizowane). Zgoda użytkownika: "robmy całe" F1 — w
  praktyce zrealizowano jako trzy skoncentrowane, osobno zweryfikowane
  przyrosty w jednej sesji (ten sam wzorzec co każda inna "duża" faza w
  tym repo — JVM_GRAPH G0-G5, SERVICE_LOADING S0-S2, ASYNC_CALLBACKS
  A0-A2), nie jeden niezweryfikowany diff. **Nie wszystko z F1's
  własnego planu zostało zrobione** — patrz "Niespełnione" poniżej;
  to jest szczerze DONE dla zrealizowanego wycinka, nie dla całej
  fazy F1 jak opisana w planie.
- **Część 1 — model wartości adnotacji (`internal/classfile`):**
  `parseAnnotations`/`parseElementValue` (przepisane z
  `skipAnnotation`/`skipElementValue`, które tylko pomijały bajty)
  dekodują teraz każdy kształt `element_value` (JVM spec 4.7.16.1):
  stałe proste, enum, literał klasy, tablicę, zagnieżdżoną adnotację.
  `cpEntry` zyskał `raw32`/`raw64` do przechowania bajtów
  Integer/Float/Long/Double (wcześniej odczytywane i odrzucane).
  `RuntimeInvisibleAnnotations` czytane i scalane z `RuntimeVisible...`
  (wcześniej całkowicie ignorowane). Pola (`classfile.Field`) zyskały
  własne adnotacje (wcześniej żadnych). `AnnotationDefault` (JVM spec
  4.7.22) przechowywane jako `Method.DefaultValue`. Starą listę samych
  FQCN (`Annotations []string`) zachowano przez adapter
  (`annotationNames`) — żaden istniejący konsument (`internal/profile`'s
  own dopasowanie po nazwie) nie wymagał zmiany, potwierdzone pełnym
  `go test ./...` bez regresji i identycznym przebiegiem na petclinic
  (3/62/26/51/6, bez zmian).
- **Część 2 — rozwiązywanie między klasami (`internal/annotation`,
  nowy pakiet):** `ResolveDefault`/`EffectiveValue` (wartość domyślna
  elementu, gdy miejsce użycia jej nie podało — czytane z
  `hierarchy.Method.DefaultValue`, bez I/O na jarze, bo wartość jest
  już przechowana przy metodzie); `ClassAnnotations` (ponowny odczyt
  klasy przez `JarClassReader`, ten sam wzorzec co
  `ResolveStaticField`/`ResolveCallResult`); `RepeatableContainer`/
  `ExpandRepeatable` — rozpakowanie kontenera repeatable adnotacji
  (`@Foo` użyte dwa razy kompiluje się do `@Foos({@Foo,@Foo})`), **z
  weryfikacją, że zagnieżdżony typ faktycznie deklaruje
  `@Repeatable(ten-kontener)`, nie tylko po kształcie** (tablica
  jednotypowych zagnieżdżonych adnotacji w elemencie "value") — zwykła
  adnotacja z niepowiązanym elementem tablicowym adnotacji (np.
  Springowe `@ComponentScan#excludeFilters`) ma ten sam kształt i nie
  może być pomylona z kontenerem.
- **Część 3 — realne wartości Spring MVC (`internal/profile/springvalues.go`):**
  `Routes` (prawdziwa ścieżka/ścieżki i czasownik(i) HTTP każdego
  mapowania — `@RequestMapping`/`@GetMapping`/.../`@PatchMapping`,
  łączenie prefiksu z poziomu klasy z poziomem metody, czasownik
  nieznany = `nil`, nigdy zgadywany); `BeanNames` (deklarowana nazwa
  beana z `@Component`/`@Service`/`@Repository`/`@Controller`/
  `@RestController` + `@Primary`); `ProfileConditions` (deklarowane
  nazwy `@Profile`, włącznie z negacją `!x`) — **nowa flaga CLI
  `-active-profiles`**: brak flagi → `ConditionUnknown` dla każdego
  warunku (F1's own explicit requirement: "brak pliku... daje unknown,
  nie false"), flaga podana → `ConditionTrue`/`ConditionFalse`
  rozstrzygane naprawdę. Trzy nowe sekcje raportu ("Spring MVC routes" /
  "Spring bean names" / "Spring @Profile conditions").
- **Zachowanie dostępne użytkownikowi:** raport pokazuje teraz
  prawdziwe trasy HTTP (ścieżka + czasownik), nazwy beanów i warunki
  `@Profile` — wcześniej widoczna była tylko goła obecność adnotacji
  (przez `roots.annotations`), nigdy jej wartość.
- **Dowód na realnym bytecode:** `bin/ariadne -offline` na petclinic:
  **17 prawdziwych tras** (np.
  `OwnerController#showOwner` → `/owners/{ownerId}` GET,
  `VetController#showVetList` → `/vets.html` GET — zgodne z
  faktycznym zachowaniem tej aplikacji) i **7 stereotypów** (wszystkie
  bez jawnej nazwy — petclinic nigdy jej nie podaje, poprawnie
  zgłoszone jako "(default)", nie zgadywane). WebGoat: **19 tras**, **20
  stereotypów** (również wszystkie bez jawnej nazwy). **Żadna z dwóch
  aplikacji nie używa `@Profile` w swoim własnym kodzie** (potwierdzone
  bezpośrednio: zero wystąpień FQCN tej adnotacji w obu app-classes
  jar) — `ProfileConditions`' pozytywna ścieżka (w tym negacja `!x`,
  rozstrzygnięcie true/false z `-active-profiles`) zweryfikowana tylko
  na poziomie testów jednostkowych (`TestProfileConditions_*`), nie na
  tych dwóch aplikacjach kontrolnych — szczerze odnotowany brak
  pokrycia na realnej aplikacji, nie przemilczany. **Zero zmian w
  REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN** na obu aplikacjach
  (identyczne z bazową wartością); dwa przebiegi na każdej dają
  bajt-identyczny `results.json` (i `report.md` poza samym
  znacznikiem czasu).
- **Spełnione kryteria (z F1's własnego "Gotowe"):** "wartości
  qualifiera, profilu i trasy mają źródło" — tak, dla Spring MVC's
  pilotowego zakresu (`Routes`/`BeanNames`/`ProfileConditions`).
  "Brak pliku lub nieobsługiwany warunek daje unknown, nie false" —
  tak, `TestProfileConditions_noActiveListIsUnknown` i realny domyślny
  tryb CLI (brak `-active-profiles`).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/classfile`:
  `TestParseAnnotations_fullElementValueShapes`, `TestAnnotationNames`,
  `TestParseAnnotationDefault`, `TestMergeParameterAnnotations`,
  `TestClassLiteralFQCN_*`; `internal/annotation` [nowy pakiet]:
  `TestResolveDefault_*` [3], `TestEffectiveValue_*` [2],
  `TestRepeatableContainer_*`, `TestExpandRepeatable_*` [3];
  `internal/profile`: `TestRoutes_*` [3], `TestBeanNames_*` [2],
  `TestProfileConditions_*` [3]); `go test -race ./...` dla całego
  projektu (zielone); `bin/ariadne -offline` na petclinic (dwa
  przebiegi) i WebGoat (dwa przebiegi), opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Profiles and
  selected roots" znacząco rozszerzony o cały F1's zakres i dowody).
  `README.md` (nowa flaga `-active-profiles` w tabeli, nowy akapit w
  sekcji 8 o trzech nowych sekcjach raportu).
- **Niespełnione/odłożone (zgodnie z własnym zakresem tego przyrostu,
  nie przez przeoczenie — F1's plan ma więcej punktów niż
  zrealizowano):** `internal/input`'s "zachować zasoby i pochodzenie
  modułu" + "jawne wejście plików i właściwości scenariusza" —
  nietknięte (poza nową flagą `-active-profiles`, która adresuje
  dokładnie jeden z tych punktów: właściwość scenariusza, nie pliki).
  Ograniczony odczyt `web.xml` — nietknięty; `PrepareAppInput` nawet
  nie zachowuje `web.xml`'s własnych bajtów dziś (sprawdzone wprost:
  `extractModule`'s `classesPrefix` filtr nigdy nie obejmuje
  `WEB-INF/web.xml`, tylko `WEB-INF/classes/`). XML bean definitions —
  nietknięte. Priorytet/konflikt wielu źródeł konfiguracji —
  nietknięty (nie ma jeszcze drugiego źródła do konfliktu z
  adnotacjami). DI/wiring (`@Autowired`/`@Qualifier` na punkcie
  wstrzyknięcia, konstruktor/pole) — to F2's własny zakres, zależny od
  DATA_FLOW D0-D2 (gotowe) i rejestru B3 (nieodnotowany), nietknięty
  tutaj celowo.
- **Następny krok:** Albo dokończenie F1's pozostałych punktów
  (`web.xml`, XML bean definitions, jawne wejście plików/właściwości,
  priorytet konfiguracji) w kolejnym przyroście tej samej fazy, albo
  przejście do `FRAMEWORKS:F2` (DI, `@Bean`, qualifier/primary
  wiązanie) korzystając już z tego, co Część 3 tutaj dostarczyła
  (`BeanNames`). Wymaga propozycji i zgody użytkownika przed startem.

### FRAMEWORKS:F1 — Adnotacje, zasoby i jawna konfiguracja (wycinek: ograniczony odczyt web.xml) — DONE for this scope; XML bean definitions/priorytet konfiguracji wciąż odłożone (2026-10-23)

- **Zakres zlecenia i zależności:** F1's poprzednie trzy części
  (wartości adnotacji, rozwiązywanie między klasami, realne wartości
  Spring MVC — gotowe wyżej). Kontynuacja tej samej fazy na życzenie
  użytkownika ("Kontynuuj" po pytaniu czy robić dalej F1 czy przejść do
  F2).
- **Nowy pakiet `internal/webxml`:** `Parse` dekoduje ograniczony,
  bezpieczny podzbiór `web.xml` — `<servlet>`/`<servlet-mapping>`/
  `<filter>`/`<filter-mapping>`/`<listener>`/`<context-param>`, z
  każdym innym elementem obecnym w dokumencie zgłaszanym w `Skipped`
  (zdeduplikowane, w kolejności pierwszego wystąpienia) — nigdy
  po prostu odrzucanym bez śladu, zgodnie z F1's own "odnotować
  pominięte elementy". **Bezpieczne z konstrukcji, nie przez flagę**:
  `encoding/xml` z pakietu standardowego Go nigdy nie pobiera
  zewnętrznego DTD ani nie rozwiązuje zewnętrznej encji — potwierdzone
  testem (`TestParse_undefinedEntityIsError`), że dokument z
  niezdefiniowaną encją (`<!ENTITY xxe SYSTEM "file:///etc/passwd">`)
  daje błąd dekodowania, nie cichą podmianę.
- **`internal/input`:** `PreparedApp` zyskał `WebXMLs map[string][]byte`
  (klucz `""` dla samodzielnego WAR, nazwa zagnieżdżonego `*.war` dla
  modułu EAR) — **wcześniej `web.xml` nie był w ogóle zachowywany**
  poza `PrepareAppInput` (sprawdzone wprost: filtr `classesPrefix` w
  `extractModule` nigdy nie obejmował `WEB-INF/web.xml`, tylko
  `WEB-INF/classes/`). Nowy `readZipEntryBytes` czyta bajty jednego
  wpisu zip po nazwie; `prepareFromFatArchive` wywołuje go tylko dla
  prawdziwego WAR (`classesPrefix == "WEB-INF/classes/"`, nigdy dla
  Boot fat-jara); `prepareFromEar` czyta `web.xml` z każdego
  zagnieżdżonego modułu WAR osobno, więc EAR z wieloma modułami
  zachowuje każdy swój własny `web.xml`, nie tylko pierwszy znaleziony.
- **`cmd/ariadne`/`internal/report`:** nowy typ `report.WebXMLInput`
  (jeden parametr `Build`, nie sześć kolejnych pozycyjnych) scala
  wyniki z każdego modułu, w deterministycznej kolejności (sortowanie
  kluczy mapy `WebXMLs` przed iteracją — iteracja po mapie w Go jest
  losowa). Nowa sekcja raportu "web.xml".
- **Zachowanie dostępne użytkownikowi:** raport pokazuje teraz
  zawartość `web.xml`'a dla wejścia WAR/EAR (wcześniej: nic, bo bajty
  nawet nie docierały poza `PrepareAppInput`). Wciąż tylko
  inwentarz/deklaracje — **żadna nowa krawędź grafu wywołań**,
  świadomie: serwlet nazwany w `web.xml` nie staje się rootem (to
  osobny krok, nie tej fazy, ten sam rygor co SERVICE_LOADING S0's
  własne "rejestracja nie jest wywołaniem").
- **Dowód na realnym bytecode, nie tylko testach jednostkowych:**
  środowisko nie ma JDK (bez `javac`), więc nie da się skompilować
  świeżej aplikacji WAR do testu — ten sam ograniczenie co G2/G4.
  Zamiast tego: realny, już skompilowany `app-classes.jar` petclinic
  przepakowany pod `WEB-INF/classes/` plus ręcznie napisany `web.xml`
  (ten sam wzorzec co G4's własny fixture multi-release — "użyj
  realnego skompilowanego bytecode, dopisz ręcznie tylko to, co nie
  wymaga kompilacji"). `bin/ariadne -offline` na tym WAR: **poprawnie
  wykryty jako WAR, `web.xml` przechwycony i poprawnie sparsowany** —
  1 serwlet, 1 mapowanie, 1 filtr, 1 mapowanie filtra, 1 listener, 1
  context-param, 2 nieznane elementy (`<display-name>`,
  `<session-config>`) poprawnie nazwane w "Skipped", nie ukryte. Dwa
  przebiegi dają bajt-identyczny `results.json`. Dodatkowo:
  petclinic/WebGoat jako zwykłe jary (nie WAR) — zero zmian w
  zachowaniu, `WebXMLs` poprawnie puste.
- **Spełnione kryteria:** "XML nie pobiera zewnętrznych definicji ani
  nie uruchamia kodu" — `TestParse_undefinedEntityIsError`. "Odnotować
  pominięte elementy" — `TestParse_fullShape`'s `Skipped` assertion.
  "Dokument zepsuty to błąd, element nierozpoznany to nie błąd" —
  `TestParse_malformedDocumentIsError` vs. resztę testów.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/webxml`:
  `TestParse_fullShape`, `TestParse_filterMappingByServletName`,
  `TestParse_malformedDocumentIsError`,
  `TestParse_undefinedEntityIsError`, `TestParse_emptyWebApp`;
  `internal/input`: `TestPrepareAppInput_war_capturesWebXML`,
  `TestPrepareAppInput_war_noWebXMLIsEmpty`,
  `TestPrepareAppInput_springBootFatJar_noWebXML`,
  `TestPrepareAppInput_ear_capturesPerModuleWebXML`); `go test -race
  ./...` dla całego projektu (zielone); realny WAR zbudowany z
  istniejącego skompilowanego `app-classes.jar` petclinic +
  ręcznie napisany `web.xml`, dwa przebiegi `bin/ariadne -offline`
  (opisane wyżej); petclinic/WebGoat jako zwykłe jary bez zmian.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (nowy wiersz
  "Configuration (web.xml, XML bean definitions)": Not implemented →
  Partial, z pełnym opisem i dowodami; wiersz "Profiles and selected
  roots" poprawiony, żeby nie twierdzić już, że `web.xml` jest
  nietknięty). `README.md` (akapit o `web.xml` w "Known limitations"
  przepisany z "not modeled" na opis realnego, ograniczonego zakresu;
  nowy akapit w sekcji 8 o nowej sekcji raportu).
- **Niespełnione/odłożone (zgodnie z własnym zakresem tego przyrostu):**
  XML bean definitions (`<bean id="..." class="...">`) — nietknięte.
  Priorytet/konflikt wielu źródeł konfiguracji — wciąż nietknięty (jest
  teraz DRUGIE źródło, `web.xml`, obok adnotacji, ale nic nie łączy ich
  w jeden scenariusz z priorytetem). Krawędzie grafu z deklaracji
  `web.xml` (serwlet jako root) — celowo nietknięte, osobny,
  świadomy krok. Jawne wejście plików/właściwości scenariusza poza
  `-active-profiles` — nietknięte.
- **Następny krok:** Tak jak poprzedni wpis zauważył — pozostałe punkty
  F1 (XML bean definitions, priorytet konfiguracji) nie są już małymi,
  mechanicznymi krokami w tym samym stylu. Dobry moment na
  `FRAMEWORKS:F2` (DI, `@Bean`, wiązanie qualifier/primary), które może
  już korzystać z `BeanNames` (Część 3) i `web.xml`'s inwentarza
  serwletów (ta faza) jako dodatkowego źródła. Wymaga propozycji i
  zgody użytkownika przed startem.

### FRAMEWORKS:F1 — Adnotacje, zasoby i jawna konfiguracja (wycinek: XML bean definitions + wykrywanie konfliktów nazw beanów) — DONE for this scope; F1 archiwizowalne jako zamknięte (2026-10-23)

- **Zakres zlecenia i zależności:** F1's poprzednie cztery części
  (gotowe wyżej). Kontynuacja na życzenie użytkownika ("konczyc" po
  pytaniu czy dokończyć F1 czy przejść do F2) — ten przyrost domyka
  F1's realny, wykonany zakres (XML bean definitions + priorytet/
  konflikt, w wąskim, szczerym sensie opisanym poniżej).
- **Nowy pakiet `internal/springxml`:** `Parse` dekoduje ograniczony
  podzbiór klasycznego Spring XML wiring — `<bean id="..." class="..."
  scope="..." primary="...">`, z `name` jako fallback dla brakującego
  `id` (pierwszy alias z listy rozdzielanej spacją/przecinkiem).
  **Celowo nie modelowane:** `<property>`/`<constructor-arg>` — nic w
  tym przyroście ich nie konsumuje, więc nie zostały dodane (reguła
  projektu "nie buduj czego nikt nie używa"). Bezpieczny z konstrukcji
  tak jak `internal/webxml` (ten sam test na niezdefiniowaną encję).
- **`internal/input.FindBeansXML`:** skanuje **tylko własne wpisy**
  `app-classes.jar` (nigdy jar zależności — ta sama zasada co
  `Routes`/`BeanNames`) w poszukiwaniu zasobu `.xml`, którego
  zewnętrzny element to `<beans>` (bez względu na namespace, tak jak
  `web-app`'owa detekcja w `internal/webxml`); `META-INF/*` wykluczone.
  Plik, który nie parsuje się jako XML w ogóle, jest cicho odrzucany
  jako "to nie jest plik beanów" (nie błąd) — ale błąd odczytu samego
  wpisu zip trafia do warningów, nie jest przemilczany.
- **Wykrywanie konfliktów (`internal/report`'s `buildConfigConflicts`):**
  grupuje każdą jawnie nazwaną deklarację beana (adnotacja ze swoim
  `value()`, `<bean>`'s own `id`) po nazwie; nazwa zgłoszona przez
  więcej niż jedno źródło z **inną** klasą trafia do "Configuration
  conflicts" jako lista kandydatów — **nigdy nie wybiera zwycięzcy**
  (prawdziwe zachowanie Springa `allow-bean-definition-overriding` jest
  konfigurowalne i nie jest tu modelowane: to detekcja, nie
  rozstrzygnięcie). Ta sama nazwa z tą samą klasą z dwóch źródeł to
  nie konflikt (nieszkodliwa redeklaracja).
- **Zachowanie dostępne użytkownikowi:** raport pokazuje teraz
  zadeklarowane beany XML (tożsamość/klasa/scope/primary) i realne
  konflikty nazw między adnotacjami a XML — dwie nowe sekcje raportu.
- **Dowód na realnym bytecode:** ten sam ograniczenie środowiska (brak
  JDK) jak przy `web.xml`'owym przyroście — zamiast kompilować nową
  aplikację, realny, już skompilowany `app-classes.jar` petclinic
  przepakowany pod `WEB-INF/classes/` plus ręcznie napisany
  `web.xml` + `applicationContext.xml` (2 beany, jeden wskazujący na
  realną, skompilowaną klasę `OwnerController`). `bin/ariadne -offline`
  na tym WAR: **oba beany XML poprawnie wyekstrahowane** (id, klasa,
  scope, primary), **zero fałszywych konfliktów** (bo petclinic's
  własne stereotypy nie mają jawnej nazwy — `Name==""` poprawnie
  wyklucza je z detekcji, zgodnie z zaprojektowanym zachowaniem). Dwa
  przebiegi dają bajt-identyczny `results.json`. Petclinic/WebGoat jako
  zwykłe jary (bez WAR/XML beanów) — zero zmian w zachowaniu, `go run`
  nie loguje nawet linii o XML beanach (bo nic nie znaleziono).
  **Pozytywna ścieżka samego konfliktu** (dwa źródła, różna klasa, ta
  sama nazwa) zweryfikowana tylko na testach jednostkowych
  (`TestBuildConfigConflicts_*`) — żadna z dwóch aplikacji kontrolnych
  nie ma jawnie nazwanego stereotypu, więc nie da się tego zademonstrować
  end-to-end na nich bez modyfikacji ich bytecode (niemożliwej bez
  JDK) — szczerze odnotowany brak pokrycia na realnej aplikacji.
- **Spełnione kryteria (z F1's własnego "Gotowe", częściowo —
  F1's plan miał więcej punktów niż ten całościowy wysiłek zrealizował,
  patrz niżej):** "wzajemnie wykluczające się warunki nie mogą być
  scalane w pozornie jeden aktywny scenariusz" — tak, dla konfliktów
  nazw beanów: `TestBuildConfigConflicts_differentClassSameNameIsConflict`
  (zgłoszone jako kandydaci, nie scalone) i
  `TestBuildConfigConflicts_sameClassSameNameIsNotAConflict`
  (nieszkodliwa redeklaracja poprawnie odróżniona).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy: `internal/springxml`:
  `TestParse_fullShape`, `TestParse_malformedDocumentIsError`,
  `TestParse_undefinedEntityIsError`, `TestParse_emptyBeans`;
  `internal/input`: `TestFindBeansXML_findsSpringBeansFile`,
  `TestFindBeansXML_noBeansFilesIsEmpty`, `TestIsSpringBeansRoot`;
  `internal/report`: `TestBuildConfigConflicts_*` [4]); `go test -race
  ./...` dla całego projektu (zielone); realny WAR zbudowany z
  istniejącego skompilowanego `app-classes.jar` petclinic + ręcznie
  napisany `web.xml`/`applicationContext.xml`, dwa przebiegi
  `bin/ariadne -offline` (opisane wyżej); petclinic/WebGoat jako zwykłe
  jary bez zmian.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Configuration
  (web.xml, XML bean definitions)" rozszerzony o XML bean definitions i
  detekcję konfliktów; wiersz "Profiles and selected roots" poprawiony).
  `README.md` (akapit o `web.xml`/XML beanach w "Known limitations"
  rozszerzony; nowy akapit w sekcji 8 o dwóch nowych sekcjach raportu).
- **Niespełnione/odłożone (zgodnie z własnym zakresem tej fazy — F1's
  plan ma więcej punktów niż ten całościowy wysiłek faktycznie
  zrealizował; nie przez przeoczenie, każdy punkt rozważony i
  świadomie odłożony lub uznany za poza zakresem pilota):**
  `<property>`/`<constructor-arg>` wiring w XML beanach — nietknięte
  (nic go nie konsumuje). **Rozstrzygnięcie** konfliktu (wybór
  zwycięzcy) — z definicji nietknięte, to świadomie detekcja, nie
  rozstrzygnięcie (prawdziwe zachowanie Springa jest konfigurowalne).
  Krawędzie grafu z deklaracji `web.xml`/XML beanów (serwlet/bean jako
  root) — celowo nietknięte, inny krok. Jawne wejście plików/
  właściwości scenariusza poza `-active-profiles` — nietknięte. DI/
  wiring (`@Autowired`/`@Qualifier` na punkcie wstrzyknięcia,
  konstruktor/pole) — to F2's własny zakres, nietknięty tutaj celowo.
- **Następny krok:** F1's realnie wykonany zakres (5 przyrostów:
  wartości adnotacji, rozwiązywanie między klasami, realne wartości
  Spring MVC, `web.xml`, XML bean definitions + konflikty) jest teraz
  w stanie, w którym kolejne punkty planu F1 (rozstrzyganie
  konfliktów, pełny priorytet wielu źródeł, jawne wejście plików) albo
  wymagają decyzji projektowej poza zakresem mechanicznego kroku, albo
  są z definicji poza zakresem (DI to F2). Zalecane: rozważyć
  archiwizację F1 jako zamkniętego dla jego faktycznie zrealizowanego
  zakresu (podobnie jak JVM_GRAPH G0-G5), i przejście do
  `FRAMEWORKS:F2` (DI, `@Bean`, wiązanie qualifier/primary), które
  może już korzystać z `BeanNames` i XML beanów jako dodatkowych
  źródeł. Wymaga propozycji i zgody użytkownika przed startem.

### FRAMEWORKS:F2 — Pierwsze DI i model zewnętrznych wejść (wycinek: rozwiązywanie `@Autowired`) — DONE for this scope; reszta F2 nietknięta (2026-10-23)

- **Zakres zlecenia i zależności:** F1 (gotowe, zarchiwizowane).
  DATA_FLOW D0–D2 (gotowe) dla "prostych fabryk" nie zostało tu użyte —
  ten wycinek nie modeluje konstrukcji `@Bean`, tylko rozwiązanie
  istniejącego katalogu beanów (F1's `BeanNames`) wobec punktów
  wstrzyknięcia. Rejestr B3 (nieodnotowany) nie był wymagany.
  Zgoda użytkownika: "decyduje" (delegacja wyboru) po pytaniu czy
  dokończyć resztę F1 czy przejść do F2 — wybrano F2, pierwszy,
  najwęższy wykonalny wycinek.
- **Nowy plik `internal/profile/di.go`:** `InjectionPoints` — znajduje
  każde pole z `@Autowired` i każdy parametr **jednego** konstruktora,
  którego Spring faktycznie użyłby (jawnie `@Autowired`, albo — reguła
  Springa 4.3 — jedyny zadeklarowany konstruktor, jeśli żaden nie jest
  oznaczony). Klasa z wieloma konstruktorami, z których żaden nie jest
  `@Autowired`, **nie dostaje żadnego** punktu wstrzyknięcia
  konstruktora — brak podstawy do wyboru, nigdy zgadywane.
  `ResolveInjections` — rozwiązuje każdy punkt wobec katalogu beanów
  po typie (dopasowanie dokładne albo jakikolwiek znany podtyp, przez
  `hierarchy.Subtypes`), potem `@Qualifier` (dopasowanie do jawnej
  nazwy kandydata), potem `@Primary` — **ten sam porządek, co
  prawdziwy kontener Springa**. Więcej niż jeden kandydat, który
  przetrwa wszystkie trzy kroki, to `Unknown` z własnym powodem, nigdy
  wybór arbitralny. Nowy eksportowany `classfile.FieldTypeFQCN` —
  konwersja deskryptora typu referencyjnego na FQCN (false dla
  prymitywu/tablicy — iniekcja kolekcji beanów to inna, nietknięta
  funkcja Springa).
- **Zachowanie dostępne użytkownikowi:** nowa sekcja raportu
  "Dependency injection" — każdy punkt wstrzyknięcia z typem,
  qualifierem (jeśli podany) i rozwiązanym beanem albo powodem
  nierozwiązania.
- **Dowód na realnym bytecode, nie tylko testach jednostkowych —
  szczery, zmierzony wynik w obu kierunkach:** `bin/ariadne -offline`
  na petclinic: **8 punktów wstrzyknięcia, 0 rozwiązanych** — ale to
  nie porażka mechanizmu: wszystkie 8 to interfejsy repozytoriów
  Spring Data (`OwnerRepository`, `VisitRepository`, `PetRepository`,
  `VetRepository`) — ich prawdziwe beany to syntetyczne proxy Spring
  Data, obsługiwane przez **inny, już istniejący mechanizm**
  (`synthetic_edges` w profilu YAML), z którym ten wycinek świadomie
  nie integruje się. WebGoat: **74 punkty, 7 naprawdę rozwiązanych**
  przez dokładne dopasowanie typu (np. `WebSecurityConfig`'s własny
  parametr konstruktora typu `UserService` rozwiązuje się do
  konkretnego `org.owasp.webgoat.users.UserService`) — realny,
  deterministyczny, zmierzony sukces, nie fabrykowany. Pozostałe
  67/74 na WebGoat nierozwiązane z jasnym powodem — sprawdzone wprost
  (nie założone), że to głównie typy JDK (`java.lang.String`,
  `java.util.List`), beany dostarczane przez framework
  (`JdbcTemplate`, `DataSource`, `AuthenticationManager`) i interfejsy
  repozytoriów — żadne z nich nie są w katalogu `BeanNames` z
  definicji tego wycinka. **Zero zmian w
  REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN** na obu aplikacjach
  (identyczne z bazową wartością). Dwa przebiegi na każdej aplikacji
  dają bajt-identyczny `results.json`.
- **Spełnione kryteria (z F2's własnego "Gotowe", częściowo — tylko
  fragment "jeden bean/qualifier"):** "jeden bean" —
  `TestResolveInjections_singleCandidateResolves` i realny dowód
  WebGoat (7 rozwiązań przez dokładny typ). "Qualifier" —
  `TestResolveInjections_qualifierDisambiguates`. Dodatkowo (nie
  jawnie wymagane, ale ten sam duch) "primary" —
  `TestResolveInjections_primaryDisambiguates`, i "brak podstawy do
  wyboru konstruktora nie jest zgadywane" —
  `TestInjectionPoints_multipleConstructorsWithoutAutowiredIsSkipped`.
  "Konflikt" i "nieznany profil" — to F1's własne kryteria, już
  spełnione tam, nie powtórzone tutaj.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy w `internal/profile`:
  `TestInjectionPoints_*` [4], `TestResolveInjections_*` [5]); `go
  test -race ./...` dla całego projektu (zielone); `bin/ariadne
  -offline` na petclinic (dwa przebiegi) i WebGoat (dwa przebiegi),
  opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "DI/AOP/
  configuration execution models": Not implemented → Partial, z
  pełnym opisem i dowodami w obie strony — zero na petclinic, 7/74 na
  WebGoat). `README.md` (nowy akapit o DI w "Known limitations" i w
  sekcji 8 o nowej sekcji raportu).
- **Niespełnione/odłożone (zgodnie z własnym, bardzo wąskim zakresem
  tego przyrostu — F2's plan ma dużo więcej punktów niż ten wycinek
  realizuje):** setter injection — nietknięty (rzadki w nowoczesnym
  Springu). Iniekcja kolekcji beanów (`List<Bean>`/`Map<String,Bean>`)
  — nietknięta. Integracja z `BeanNames` + XML beanów (F1's
  `internal/springxml`) jako jeden wspólny katalog — nietknięta
  (osobne pakiety, świadomie nie połączone w tym przyroście).
  Modelowanie `@Bean` factory methods, konstrukcji i alokacji
  (DATA_FLOW D0–D2's "znane fabryki") — nietknięte. Oznaczone korzenie
  HTTP/servlet/filter/lifecycle/scenariusz kontenera — nietknięte.
  Katalog punktów wejścia wskazujących zewnętrzny parser/dispatch —
  nietknięty. Przekazanie tożsamości parametrów do przyszłych źródeł
  taint D4 — nietknięte (D4 nie istnieje). **Żadna nowa krawędź grafu
  wywołań** — rozwiązana iniekcja to fakt o wiązaniu, nie wywołanie,
  ten sam rygor co SERVICE_LOADING S0/S1.
- **Następny krok:** Pozostały zakres F2 jest duży i heterogeniczny —
  oznaczone korzenie HTTP/servlet (prawdopodobnie najwyższa wartość,
  bo bezpośrednio wpływa na reachability), modelowanie `@Bean` factory
  methods, albo integracja DI z XML beanami. Wymaga propozycji i
  zgody użytkownika przed startem, jak każda następna faza.

### FRAMEWORKS:F2 — Pierwsze DI i model zewnętrznych wejść (wycinek: oznaczone korzenie web.xml) — DONE for this scope; reszta F2 nietknięta (2026-10-23)

- **Zakres zlecenia i zależności:** F2's pierwszy wycinek (DI, gotowy
  wyżej). Kontynuacja na "kontynuuj" — wybrano (decyzja własna po
  delegacji "decyduje") "oznaczone korzenie dla HTTP, servlet/filter/
  lifecycle" jako najwyższej wartości pozostały punkt F2 (bezpośredni
  wpływ na reachability pod `-roots profile`, w przeciwieństwie do
  czystego inwentarza).
- **Nowy plik `internal/profile/webxmlroots.go`:** `WebXMLRoots` —
  dla każdej klasy servletu/filtru/listenera zadeklarowanej w
  `web.xml` (z F1's już istniejącego `internal/webxml`/`PreparedApp.WebXMLs`),
  sprawdza, które ze znanych nazw metod cyklu życia (`init`/`destroy`/
  `service`/`doGet`/.../`doFilter`/`contextInitialized`/...) klasa
  faktycznie deklaruje — **nigdy każda publiczna metoda**, zgodnie z
  planu własnym ostrzeżeniem "nie uznawać każdego publicznego helpera
  za endpoint". `WebXMLAppRoots` — ta sama logika, ale zwraca węzły
  grafu wywołań (ten sam wzorzec co istniejący `Profile.overrideRoots`,
  bez rozszerzania na podtypy — `web.xml` nazywa dokładną konkretną
  klasę, kontener nie szuka override'u).
- **`cmd/ariadne`:** zbiera `servletClasses`/`filterClasses`/
  `listenerClasses` (zdeduplikowane) podczas parsowania `web.xml`
  (ten sam krok, co F1's własne `webXMLInput`); pod `-roots profile`
  dołącza `profile.WebXMLAppRoots(...)` do `prof.AppRoots(...)`. Nowa
  sekcja raportu "web.xml-declared roots", pokazywana **niezależnie**
  od trybu `-roots` (fakt inwentarzowy, nie decyzja o wyborze korzeni).
- **Zachowanie dostępne użytkownikowi:** `-roots profile` teraz
  dołącza metody cyklu życia servletów/filtrów/listenerów zadeklarowanych
  w `web.xml`, nawet bez żadnej adnotacji — wcześniej taki "goły"
  serwlet nie dawał żadnego korzenia pod `-roots profile` (tylko
  `-roots all` go pokrywał przez "każda metoda aplikacji").
- **Dowód — szczerze, tylko na poziomie testów jednostkowych, nie
  end-to-end na realnej aplikacji:** ani petclinic, ani WebGoat nie
  mają `web.xml` (obie to zwykłe jary). Jednorazowy diagnostyczny
  program (uruchomiony i usunięty) przeszukał **wszystkie** metody w
  obu app-classes jar pod kątem nazw z listy cyklu życia
  (`init`/`destroy`/`service`/`doGet`/`doPost`/`doFilter`/
  `contextInitialized`/`contextDestroyed`) — **zero przypadkowych
  trafień** na obu aplikacjach, więc nie da się nawet przypadkowo
  zademonstrować pozytywnej ścieżki na tych dwóch kontrolnych
  aplikacjach bez kompilacji nowej klasy (niemożliwej bez JDK w tym
  środowisku — to samo ograniczenie co przy `web.xml`'owym przyroście
  F1). Potwierdzone: `-roots profile` na petclinic (24 punkty) i
  WebGoat (40 punktów) — **bez zmian** względem stanu przed tym
  przyrostem, "0 from web.xml" poprawnie zalogowane na obu. Realny
  WAR (petclinic's `app-classes.jar` + ręcznie napisany `web.xml`
  wskazujący na nieistniejącą klasę `demo.MainServlet`) przechodzi
  przez cały pipeline bez błędu, poprawnie zgłasza "None found" (klasa
  nazwana w `web.xml`, ale nieobecna w analizowanej hierarchii —
  szczerze pominięta, nie fabrykowana). Dwa przebiegi dają
  bajt-identyczny `results.json`.
- **Spełnione kryteria:** "nie uznawać każdego publicznego helpera za
  endpoint" — `TestWebXMLRoots_servletLifecycleMethodsFound` (metoda
  pomocnicza bez nazwy z listy nie staje się korzeniem). "Klasa
  nieznana w hierarchii nie jest fabrykowana" —
  `TestWebXMLRoots_classNotInHierarchyIsSkipped` i realny WAR dowyżej.
  "Metoda abstrakcyjna nie ma ciała do wywołania" —
  `TestWebXMLRoots_abstractMethodSkipped`. "Realne węzły grafu" —
  `TestWebXMLAppRoots_resolvesRealGraphNodes` (przez prawdziwy
  `callgraph.Build`, nie tylko metadane).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone, nowe testy w `internal/profile`:
  `TestWebXMLRoots_*` [4], `TestWebXMLAppRoots_resolvesRealGraphNodes`);
  `go test -race ./...` dla całego projektu (zielone); `bin/ariadne
  -offline -roots profile` na petclinic (dwa przebiegi) i WebGoat;
  realny WAR (jeden przebieg, opisany wyżej); jednorazowy diagnostyczny
  program skanujący oba app-classes jar pod kątem nazw cyklu życia
  (uruchomiony i usunięty).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "DI/AOP/
  configuration execution models" rozszerzony o ten wycinek i szczery
  opis braku end-to-end dowodu na realnej aplikacji). `README.md`
  (nowy akapit "Web.xml roots" w "Known limitations").
- **Niespełnione/odłożone:** `@Bean` factory methods i konstrukcja —
  nietknięte. AOP/pointcut — nietknięte. Katalog punktów wejścia
  wskazujących zewnętrzny parser/dispatch poza samym `web.xml` — 
  nietknięty. Integracja oznaczonych korzeni z DI (np. czy serwlet
  sam ma wstrzyknięte zależności) — nietknięta. Przekazanie
  tożsamości parametrów do przyszłych źródeł taint D4 — nietknięte
  (D4 nie istnieje).
- **Następny krok:** Pozostały zakres F2 (modelowanie `@Bean`,
  katalog punktów wejścia zewnętrznych) wymaga decyzji projektowej.
  Alternatywnie inny obszar. Wymaga propozycji i zgody użytkownika
  przed startem.

### FRAMEWORKS:F2 — Pierwsze DI i model zewnętrznych wejść (wycinek: `@Bean` factory methods) — DONE for this scope; reszta F2 nietknięta (2026-10-23)

- **Zakres zlecenia i zależności:** F2's pierwszy wycinek (`@Autowired`,
  gotowy wyżej) i F1's `BeanNames`. Zgoda użytkownika: wybrano
  `@Bean` factory methods jako następny wycinek z czterech
  przedstawionych opcji (B3, @Bean, katalog zewnętrznych wejść, inne).
- **Zmiana w `internal/profile/springvalues.go`:** nowa `BeanMethods` —
  dla każdej klasy z `@Configuration` szuka metod z `@Bean`, tworzy
  `BeanName{FQCN: <typ zwracany metody>, Name: <@Bean("x")/value()/
  name() albo nazwa metody>, Primary: <@Primary na metodzie>}`.
  Metoda bez `@Bean`, klasa bez `@Configuration` albo zwracany typ
  `void`/prymityw/tablica nie dają wpisu — nigdy zgadywane. Fabryki
  niejednoznaczne/programowalne (zwracany typ z wieloma implementacjami,
  konstrukcja zależna od innych beanów) pozostają nietknięte zgodnie z
  planem F2's "niejednoznaczne... zachowują alternatywy/luki" — raportowany
  jest tylko sam zwracany typ, nic ponad to. `ResolveInjections`/
  `InjectionPoint`/`disambiguate` niezmienione — wynik `BeanMethods`
  jest po prostu scalony z `BeanNames` w `cmd/ariadne/main.go` przed
  wywołaniem `ResolveInjections`.
- **Zachowanie dostępne użytkownikowi:** sekcja "Dependency injection"
  rozwiązuje teraz więcej punktów — bean pochodzący z `@Bean` metody
  liczy się identycznie jak ze stereotypu. Nowa linia logu "N @Bean
  factory methods".
- **Dowód na realnym bytecode, szczery, zmierzony w obie strony:**
  `bin/ariadne -offline -profile spring` na petclinic: 1 `@Bean`
  metoda znaleziona (`CacheConfiguration#...Customizer` →
  `JCacheManagerCustomizer`), ale wstrzyknięcia zostają **0/8**
  (wszystkie 8 to interfejsy Spring Data, nieinjectujące tego typu) —
  poprawnie bez zmiany. WebGoat: **21 `@Bean` metod znalezionych**,
  rozwiązanie skacze z **7/74 do 40/74** — realny, pięciokrotny,
  zmierzony wzrost, nie fabrykowany. Porównanie z binarką przed tym
  przyrostem (zbudowaną z `git stash`) potwierdza: `REACHABLE=42
  FRAMEWORKS=44 NOT_REACHABLE=18 UNKNOWN=56` identyczne przed i po na
  WebGoat (DI to fakt metadanowy, nie krawędź grafu, jak w poprzednich
  wycinkach F2). Dwa przebiegi na obu aplikacjach dają bajt-identyczny
  `results.json`.
- **Spełnione kryteria:** nowe testy jednostkowe w `internal/profile`:
  `TestBeanMethods_explicitNameAndPrimary`,
  `TestBeanMethods_noExplicitNameDefaultsToMethodName`,
  `TestBeanMethods_ignoresNonConfigurationClasses`,
  `TestBeanMethods_ignoresMethodsWithoutBeanAnnotation`,
  `TestBeanMethods_voidReturnContributesNothing`,
  `TestBeanMethods_resolvesRealInjectionPoint` (integracja z
  `ResolveInjections` na syntetycznym `ClassInfo`, wzorem `di_test.go`).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet ./...`,
  `go test ./...` (zielone); `bin/ariadne -offline` na petclinic (dwa
  przebiegi) i WebGoat (dwa przebiegi), opisane wyżej; porównanie z
  binarką z poprzedniego commita (`git stash`/`git stash pop`) na
  WebGoat dla potwierdzenia braku regresji reachability.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "DI/AOP/
  configuration execution models": `@Bean` usunięty z listy gapów,
  dodany zmierzony wynik 1/0 na petclinic i 21/40×74 na WebGoat).
  `README.md` (akapit "DI" w "Known limitations" zaktualizowany o
  `@Bean` factory methods i nowe liczby).
- **Niespełnione/odłożone:** fabryki niejednoznaczne/programowalne —
  nietknięte (zgodnie z planem). Integracja z XML `<bean>` (F1's
  `internal/springxml`) — nietknięta. AOP/pointcut — nietknięte.
  Framework-provided beans (`JdbcTemplate`, `DataSource`, ...) i Spring
  Data repository proxies — nietknięte (WebGoat's pozostałe 34/74).
  Katalog punktów wejścia zewnętrznych, rejestr B3 — nietknięte.
- **Następny krok:** Pozostały zakres F2 (AOP/proxy to już F3's własny
  zakres; w ramach F2 zostają: rejestr B3, katalog zewnętrznych
  punktów wejścia, integracja DI z XML beanami/framework-provided
  beans) wymaga propozycji i zgody użytkownika przed startem.

### FRAMEWORKS:F2 — Pierwsze DI i model zewnętrznych wejść (wycinek: katalog zewnętrznych punktów wejścia) — DONE for this scope; reszta F2 nietknięta (2026-10-23)

- **Zakres zlecenia i zależności:** F1 (`Routes`, `ProfileConditions`),
  F2's drugi wycinek (`WebXMLRoots`), F0's nieużywany dotąd kontrakt
  `ExternalEntry` (`internal/profile/scenario.go`). Zgoda użytkownika:
  wybrano ten wycinek jako rekomendowany z trzech przedstawionych opcji
  (katalog wejść, rejestr B3, przejście do F3).
- **Nowy plik `internal/profile/externalentries.go`:** `ExternalEntries`
  łączy `Routes(h)` (Basis `http-mapping`) i `WebXMLRoots(h, ...)`
  (Basis `servlet`/`filter`/`listener`) w jeden katalog; dla każdego
  wpisu `Active` ustawiane przez `activeCondition`, które sprawdza
  `ProfileConditions` metody, potem klasy, a gdy nic nie gatuje —
  `ConditionTrue` (nigdy zgadywane jako `unknown`, gdy naprawdę nic nie
  warunkuje). `ExternalEntry` (typ) przeniesiony z `scenario.go`'s
  nieużywanego kontraktu F0 do tego pliku jako realna implementacja —
  usunięta duplikacja deklaracji. **Celowo żadnej zmiany w
  `AppRoots`/root-selection** — zgodnie z FRAMEWORKS.MD scenario 4,
  to czysto etykietujący, raportowy przebieg, nie fabrykujący krawędzi
  parser→kontroler.
- **`cmd/ariadne`/`internal/report`:** nowy parametr `Build`, nowa
  sekcja raportu "Entry point catalog" (klasa/metoda/basis/active),
  nowa linia logu "N external entries cataloged".
- **Dowód na realnym bytecode:** `bin/ariadne -offline -profile spring`
  na petclinic: 17 wpisów katalogu (17 routes + 0 web.xml, zgodnie z
  już zweryfikowanymi liczbami), wszystkie `Active: true` (żadna z
  aplikacji nie używa `@Profile`). WebGoat: 19 wpisów (19 routes + 0
  web.xml), też wszystkie `true`. Liczby katalogu zgadzają się 1:1 z
  już zweryfikowanymi `Routes`/`WebXMLRoots` — brak podwójnego liczenia
  czy utraty wpisów. `REACHABLE=42 FRAMEWORK=44 NOT_REACHABLE=18
  UNKNOWN=56` na WebGoat i odpowiadające liczby na petclinic
  niezmienione względem stanu przed tym przyrostem. Dwa przebiegi na
  obu aplikacjach dają bajt-identyczny `results.json`.
- **Spełnione kryteria:** nowe testy jednostkowe w `internal/profile`:
  `TestExternalEntries_httpMappingUnconditionallyActive`,
  `TestExternalEntries_profileGatedRouteUnknownWithoutActiveList`,
  `TestExternalEntries_profileGatedRouteResolvesWithActiveList` (oba
  kierunki: `@Profile` aktywny i nieaktywny), `TestExternalEntries_
  webXMLServletIncluded`, `TestExternalEntries_noEntriesIsEmptyNotNil
  Panic`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet
  ./...`, `go test ./...` (zielone); `bin/ariadne -offline` na
  petclinic (dwa przebiegi) i WebGoat (dwa przebiegi), opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "DI/AOP/
  configuration execution models": czwarty wycinek opisany z
  liczbami; wiersz "Profiles and selected roots": `ExternalEntry`
  odnotowany jako już zbudowany, nie tylko zaprojektowany).
  `README.md` (nowy akapit "Entry point catalog" w "Known
  limitations").
- **Niespełnione/odłożone:** żadna zmiana w `-roots profile`'s własnym
  zestawie korzeni — to osobna decyzja, nietknięta tutaj. `@Scheduled`/
  `@EventListener` jako osobny `Basis` — nietknięte (tylko HTTP i
  web.xml). Deduplikacja `httpMappingVerbs` (springvalues.go) vs
  `spring.yaml`'s `roots.annotations` — nietknięta (osobny refaktor).
  `@Profile`-owa aktywacja sprawdzona tylko na testach jednostkowych —
  żadna z kontrolnych aplikacji nie używa `@Profile` (ten sam, już
  odnotowany gap co F1's `ProfileConditions`). Rejestr B3 — nietknięty.
- **Następny krok:** rejestr B3, integracja DI z XML beanami/
  framework-provided beans, albo przejście do F3 (proxy/AOP) — wymaga
  propozycji i zgody użytkownika przed startem.

### FRAMEWORKS:F2 — Pierwsze DI i model zewnętrznych wejść (wycinek: Spring Data repository proxy beans) — DONE for this scope; reszta F2 nietknięta (2026-10-23)

- **Zakres zlecenia i zależności:** F2's pierwszy wycinek (`@Autowired`,
  gotowy wyżej). Zgoda użytkownika: pierwotnie zaproponowano
  "framework-provided beans" jako następny krok; przed startem
  zmierzono rzeczywisty rozkład pozostałych 34 nierozwiązanych
  wstrzyknięć na WebGoat (12 String, 7 UserTrackerRepository, 6 List,
  3 UserRepository, 1 JdbcTemplate, 1 ResourceLoader, 1 Assignment,
  1 Connection, 1 Integer, 1 Boolean) i 8 na petclinic (wszystkie
  Spring Data) — framework-provided beans dałoby tylko 2/34, Spring
  Data repository proxy 10/34+8/8. Użytkownik zatwierdził zmienioną
  rekomendację po przedstawieniu dokładnych liczb.
- **Zmiana w `internal/profile/springvalues.go`:** nowa
  `RepositoryBeans` — dla każdego interfejsu aplikacji będącego
  (transytywnym) podtypem `org.springframework.data.repository.
  Repository` (ten sam sprawdzian co istniejący `synthetic_edges` w
  `spring.yaml`, ponownie użyty, nie wymyślony od nowa) tworzy
  `BeanName{FQCN: <ten interfejs>}` — realny proxy Spring Data
  implementuje dokładnie ten interfejs, nic więcej do wyliczenia.
  Klasa konkretna (nigdy tak deklarowany repository) i interfejs
  spoza aplikacji (`!IsApp`) pomijane. Scalone z `BeanNames`+
  `BeanMethods` w `cmd/ariadne/main.go` przed `ResolveInjections` —
  `ResolveInjections` sam niezmieniony.
- **Zachowanie dostępne użytkownikowi:** sekcja "Dependency injection"
  rozwiązuje teraz też repozytoria Spring Data. Nowa linia logu
  "N Spring Data repository beans".
- **Dowód na realnym bytecode, zmierzony w obie strony:** `bin/ariadne
  -offline -profile spring` na petclinic: 4 repository beans
  znalezione, wstrzyknięcia skaczą z **0/8 do 8/8**. WebGoat: 2
  repository beans, wstrzyknięcia z **40/74 do 50/74** (dokładnie +10,
  zgodnie z przewidywaniem z inspekcji powodów nierozwiązania).
  `REACHABLE=42 FRAMEWORK=44 NOT_REACHABLE=18 UNKNOWN=56` na WebGoat i
  odpowiadające liczby na petclinic niezmienione. Dwa przebiegi na
  obu aplikacjach dają bajt-identyczny `results.json`.
- **Spełnione kryteria:** nowe testy jednostkowe w `internal/profile`:
  `TestRepositoryBeans_appInterfaceSubtypeOfRepository`,
  `TestRepositoryBeans_concreteClassNotABean`,
  `TestRepositoryBeans_unrelatedInterfaceSkipped`,
  `TestRepositoryBeans_resolvesRealInjectionPoint`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet
  ./...`, `go test ./...` (zielone); `bin/ariadne -offline` na
  petclinic (dwa przebiegi) i WebGoat (dwa przebiegi), opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "DI/AOP/
  configuration execution models": piąty wycinek, z dokładnym
  rozkładem pozostałych 24/74 na WebGoat — tylko 2 to prawdziwe
  framework-provided beans, reszta to `@Value`-owe właściwości).
  `README.md` (akapit "DI" zaktualizowany o repository beans i nowe
  liczby 8/8, 50/74).
- **Niespełnione/odłożone:** framework-provided beans
  (`JdbcTemplate`/`ResourceLoader`, tylko 2 punkty) — nietknięte,
  niska wartość potwierdzona pomiarem. `@Value`-owe właściwości
  (`String`/`List`/`Integer`/`Boolean`, 21/24 pozostałych na WebGoat)
  — osobny mechanizm, nietknięty. Integracja z XML `<bean>`, AOP/
  pointcut, rejestr B3 — nietknięte.
- **Następny krok:** rejestr B3, `@Value`-owe właściwości, albo
  przejście do F3 (proxy/AOP) — wymaga propozycji i zgody użytkownika
  przed startem.

### FRAMEWORKS:F2 — korekta (nie nowy wycinek): fałszywe trafienia niejawnej reguły "jedyny konstruktor" (2026-10-23)

- **Kontekst i jak to znaleziono:** przed podjęciem decyzji "co dalej"
  (kandydat: `@Value`) zmierzono dokładny rozkład 24 pozostałych
  nierozwiązanych wstrzyknięć na WebGoat zamiast zakładać, że to
  wszystko `@Value`. Śledztwo źródła WebGoat wykazało, że tylko 3-4 z
  21 pozycji typu `String`/`List`/`Integer`/`Boolean` to realne
  użycia `@Value` — reszta to fałszywe trafienia niejawnej reguły
  Springa 4.3 "jedyny konstruktor" w `InjectionPoints`
  (`internal/profile/di.go`), która przed tą poprawką stosowała się do
  **każdej** klasy aplikacji z jednym konstruktorem, bez sprawdzenia,
  czy Spring faktycznie zarządza tą klasą. Złapane fałszywie: klasy z
  Lombokiem (`@lombok.Value`, `@AllArgsConstructor` — adnotacja o tej
  samej nazwie co Springowe `@Value`, ale niezwiązana), zwykłe enumy
  (konstruktor generowany przez kompilator), POJO budowane ręcznie
  przez `new`.
- **Poprawka w `internal/profile/di.go`:** `InjectionPoints` przyjmuje
  teraz `stereotypeBeans []BeanName` (katalog F1's `BeanNames`);
  niejawna reguła "jedyny konstruktor" stosuje się tylko, gdy klasa
  jest w tym katalogu **albo** nosi `@Configuration` (sprawdzane
  bezpośrednio, bo `BeanNames` nigdy nie katalogowało `@Configuration`
  jako stereotypu). Jawna ścieżka `@Autowired` na konstruktorze
  pozostaje bez zmian — działa niezależnie od stereotypu/
  `@Configuration`, jak wcześniej. Ważne: `WebSecurityConfig`
  (`@Configuration` + Lombokowy `@AllArgsConstructor`, **bez**
  `@Autowired`) to realny przypadek z już zaraportowanego wyniku F2
  part 1 — stąd `@Configuration` musiało zostać uwzględnione, nie
  tylko stereotyp.
- **Dowód na realnym bytecode:** WebGoat: punkty wstrzyknięcia spadają
  z (zawyżonych) 74 do **43** realnych, rozwiązane z 50 do **37** —
  `WebSecurityConfig`'s własna rezolucja do `UserService` przetrwała
  (potwierdzone w `out/WebGoat/report.md`). Wszystkie **6** pozostałych
  nierozwiązanych są teraz w pełni, uczciwie wyjaśnione: 1 realny
  `@Value` (`DatabaseConfiguration`'s `String`), 4 realne `List<Bean>`
  (`CourseConfiguration`×2, `LessonMenuService`×2 — już udokumentowane
  poza zakresem), 1 realny framework-provided bean (`UserService`'s
  `JdbcTemplate`) — nic niewyjaśnionego nie zostało. petclinic bez
  zmiany (8 punktów, 8/8) — żadna jego klasa nie trafiła na fałszywy
  kształt. `REACHABLE=42 FRAMEWORK=44 NOT_REACHABLE=18 UNKNOWN=56` na
  WebGoat niezmienione. Dwa przebiegi na obu aplikacjach dają
  bajt-identyczny `results.json`.
- **Spełnione kryteria:** nowe testy regresyjne w `internal/profile`:
  `TestInjectionPoints_soleConstructorOnPlainClassIsNotFabricated`
  (klasa bez stereotypu/`@Configuration` → zero punktów),
  `TestInjectionPoints_soleConstructorOnConfigurationClass`
  (`@Configuration` bez `@Autowired` → punkt zachowany). Istniejące
  testy zaktualizowane do nowej sygnatury.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet
  ./...`, `go test ./...` (zielone); `bin/ariadne -offline` na
  petclinic (dwa przebiegi) i WebGoat (dwa przebiegi), opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "DI/AOP/
  configuration execution models": dopisana sekcja "korekta, nie
  nowy wycinek" z poprawionymi liczbami, historyczne zdania o
  częściach 1/3/5 pozostawione nietknięte jako zapis sesji).
  `README.md` (akapit "DI" zaktualizowany o poprawione liczby 43/37/6).
- **Niespełnione/odłożone:** to była korekta błędu, nie nowy zakres —
  `@Value`, framework-provided beans, `List<Bean>` collection
  injection, integracja z XML `<bean>`, AOP/pointcut, rejestr B3
  pozostają nietknięte jak wcześniej.
- **Następny krok:** rejestr B3, `@Value`-owe właściwości (teraz
  realnie tylko 1 punkt na WebGoat, niska wartość), albo przejście do
  F3 (proxy/AOP) — wymaga propozycji i zgody użytkownika przed
  startem.

### FRAMEWORKS:F3 — Proxy i ograniczony model AOP (wycinek: tożsamość proxy/target/interceptor) — DONE for this scope; reszta F3 nietknięta (2026-10-23)

- **Zakres zlecenia i zależności:** F2 (gotowe). Przed startem
  zbadano realne zależności F3 z planu: pełne powiązanie
  `InvocationHandler.invoke` → wywołanie interfejsu wymaga
  REFLECTION:R2 (`Method.invoke`/`Constructor.newInstance`
  rekonstrukcja celu) — **całkowicie UNASSESSED, nieruszone**.
  Migracja mostka Spring Data do katalogu modeli z deskryptorami
  wymaga architektury bliskiej B3. Katalog typów dla GRAPH_PRECISION
  nie ma konsumenta (GRAPH_PRECISION samo UNASSESSED). Zgoda
  użytkownika: wybrano jedyny realnie wykonalny, samodzielny wycinek —
  czysty inwentarz tożsamości, bez rezolucji dispatchu.
- **Nowy pakiet `internal/aop/identity.go`:** `FindProxyHandlers` —
  dla każdej klasy aplikacji będącej (transytywnym) podtypem
  `java.lang.reflect.InvocationHandler` (ten sam wzorzec `h.Subtypes`
  co `RepositoryBeans`), zbiera jej własne pola `final`, instancyjne,
  typu referencyjnego jako "kandydatów na target" — kształt
  zaobserwowany w realnym przykładzie WebGoat (`private final
  Connection targetConnection`). **Celowo zero nowych krawędzi grafu
  wywołań** — wiązanie `invoke`'s argumentu `Method` z realnym
  wywołaniem interfejsu wymaga R2, którego nie ma; zgłoszenie krawędzi
  bez tego sfabrykowałoby dispatch niemożliwy do udowodnienia na tym
  etapie. CGLIB nietknięty i nietykalny w zasadzie — podklasy CGLIB
  nie istnieją jako `.class` w analizowanym bytecode.
- **`cmd/ariadne`/`internal/report`:** nowy parametr `Build`, nowa
  sekcja raportu "Proxy/interceptor identity" (klasa + kandydaci na
  pole celu), nowa linia logu "N InvocationHandler implementations
  found".
- **Dowód na realnym bytecode:** `bin/ariadne -offline -profile
  spring` na WebGoat: **1** implementacja znaleziona —
  `org.owasp.webgoat.lessons.LessonConnectionInvocationHandler` z
  `targetConnection: java.sql.Connection` — potwierdzone bezpośrednim
  odczytem źródła (`private final Connection targetConnection`,
  przypisane w konstruktorze, odczytywane przez `invoke` przed
  `method.invoke(targetConnection, args)`). petclinic: **0** —
  żadna klasa w jego `app-classes.jar` nie implementuje
  `InvocationHandler` — poprawnie "None found", nie fabrykowane.
  `REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN` niezmienione na obu
  aplikacjach. Dwa przebiegi na każdej dają bajt-identyczny
  `results.json`.
- **Spełnione kryteria:** nowe testy jednostkowe w `internal/aop`:
  `TestFindProxyHandlers_finalFieldIsCandidateTarget`,
  `TestFindProxyHandlers_nonFinalFieldSkipped`,
  `TestFindProxyHandlers_staticFieldSkipped`,
  `TestFindProxyHandlers_primitiveFieldSkipped`,
  `TestFindProxyHandlers_nonHandlerClassNotIncluded`,
  `TestFindProxyHandlers_dependencyJarHandlerNotIncluded`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet
  ./...`, `go test ./...` (zielone); `bin/ariadne -offline` na
  petclinic (dwa przebiegi) i WebGoat (dwa przebiegi), opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (nowy wiersz "Proxy/
  interceptor identity"). `README.md` (nowy akapit "Proxy/interceptor
  identity" w "Known limitations").
- **Niespełnione/odłożone (całe pozostałe F3):** powiązanie
  handler↔wywołanie interfejsu (wymaga R2) — nietknięte. Różnice
  JDK Proxy/CGLIB, final/private, self-invocation — nietknięte.
  Pointcut/advice semantyka — nietknięta. Migracja mostka Spring Data
  do katalogu modeli — nietknięta. Katalog typów dla GRAPH_PRECISION —
  nietknięty (brak konsumenta). Heurystyka "kandydat na target" to
  dopasowanie kształtu (final+instancyjne+referencyjne), nie
  śledzenie przepływu danych wewnątrz `invoke` — klasa z więcej niż
  jednym takim polem zgłasza wszystkie, nierozróżnione.
- **Następny krok:** REFLECTION:R2 (odblokowałoby realne krawędzie
  grafu dla F3 i inne obszary — DESERIALIZATION m.in.), rejestr B3,
  albo `@Value`-owe właściwości — wymaga propozycji i zgody
  użytkownika przed startem.

### REFLECTION:R0 + R1 (wycinek A: `Class.forName` z literałem) — DONE for this scope; reszta R1 i R2–R4 nietknięte (2026-10-23)

- **Zakres zlecenia i zależności:** zgoda użytkownika na REFLECTION:R2
  jako "najlepszy" następny krok; przed startem sprawdzono plan i
  stwierdzono, że R2 wymaga najpierw R0 (model wartości) i R1
  (`forName`/`getMethod` lookup) — "R0–R2 to pierwsze wydanie"
  (jedna, powiązana grupa faz), obie całkowicie nieruszone. Zgoda
  użytkownika na skorygowany zakres: R0 + R1's pierwszy, najwęższy
  wycinek (`Class.forName` z literałem), R2 (realne krawędzie
  `invoke`/`newInstance`) na kolejną sesję.
- **Decyzja projektowa R0:** zamiast nowego `Kind` (`KindClass`),
  `Class.forName("literał")` rozwiązuje się do **tego samego**
  `valueflow.KindConst{Type: "java.lang.Class"}`, co istniejący
  literał `Foo.class` (`constValue`'s własny `ConstKindClass` case w
  `step.go`) — to ten sam fakt ("Class o tej nazwie"), osiągnięty
  dwoma różnymi sposobami, a nie dwa różne fakty. Decyzja odnotowana
  jawnie, nie przemilczana: unika nadmiarowej abstrakcji (nowy Kind
  bez realnej potrzeby) i automatycznie korzysta z niej każdy
  istniejący konsument `KindConst`/`java.lang.Class` (np.
  SERVICE_LOADING S1's `ServiceLoader.load(Class)`).
- **Nowy plik `internal/dataflow/reflectapi.go`:** `classForNameValue`
  — rozpoznaje `Class.forName(String)` i `forName(String, boolean,
  ClassLoader)` (katalog przeciążeń po deskryptorze, wzorem
  `lookupFindCatalog`), czyta literał z `args[0]`, zwraca
  `valueflow.Const("java.lang.Class", fqcn)`. Wpięte w `stepCall`
  (`step.go`) przed generycznym fallbackiem `CallResult`, wzorem
  `lookupHandleValue`. Nazwa przekazana przez zmienną (store/load)
  rozwiązuje się identycznie jak literał bezpośredni — za darmo,
  dzięki istniejącemu już propagowaniu wartości lokalnych D2 (R1's
  własne "Gotowe": "literalna nazwa i nazwa przez zmienną dają ten
  sam wynik"). Nazwa runtime-computed (parametr, nieznana) spada do
  generycznego `KindCallResult` — nigdy nie zgadywana (R1's własny
  "poza pierwszym zakresem: dowolne runtime-computed nazwy").
- **Zachowanie dostępne użytkownikowi:** żadnej nowej sekcji raportu
  — to ulepszenie warstwy wartości, konsumowane przez istniejące
  mechanizmy (SERVICE_LOADING S1) automatycznie, bez zmian w ich
  własnym kodzie.
- **Dowód na realnym bytecode:** `bin/ariadne -offline` na petclinic
  i WebGoat — **zero zmiany** w liczbach SERVICE_LOADING S1 (13/29
  petclinic, 31/48 WebGoat, identyczne jak przed zmianą) — potwierdzone
  bezpośrednio (nie założone), że żadna z aplikacji nie używa
  `Class.forName` w swoim kodzie. `REACHABLE/FRAMEWORK/NOT_REACHABLE/
  UNKNOWN` niezmienione na obu. Dwa przebiegi na każdej dają
  bajt-identyczny `results.json`.
- **Spełnione kryteria:** 4 nowe testy jednostkowe w `internal/dataflow`:
  `TestClassForNameValue_literalResolves`,
  `TestClassForNameValue_throughLocalVariableResolvesSame` (R1's
  "Gotowe" bezpośrednio sprawdzone),
  `TestClassForNameValue_runtimeComputedNameIsNotFabricated` (R1's
  non-goal sprawdzony),
  `TestClassForNameValue_threeArgOverloadResolves`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet
  ./...`, `go test ./...`, `go test -race ./...` (wszystkie zielone);
  `bin/ariadne -offline` na petclinic (dwa przebiegi) i WebGoat (dwa
  przebiegi), opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (nowy wiersz
  "Reflection (Class/Method/Constructor)"). `README.md` (akapit
  "Reflection under-approximates" rozszerzony o wyjątek).
- **Niespełnione/odłożone:** `getMethod`/`getDeclaredMethod` —
  nietknięte (wymaga dekodowania literału tablicy `Class[]` — osobna,
  nietrywialna praca, świadomie odłożona na kolejny wycinek R1, nie
  tę sesję). `Method.invoke`/`Constructor.newInstance` (R2, realne
  krawędzie grafu) — nietknięte, zależy od R1 w pełni. R3 (przepływ
  między metodami, `setAccessible`), R4 (kwalifikacja) — nietknięte.
  **Żadna nowa krawędź grafu wywołań** — to fakt o wartości, nie
  wywołanie, ten sam rygor co `KindHandle`/`KindFunctional`.
- **Następny krok:** R1's drugi wycinek (`getMethod`/
  `getDeclaredMethod`, dekodowanie `Class[]`), albo R2 po jego
  ukończeniu — wymaga propozycji i zgody użytkownika przed startem.

### REFLECTION:R1 — drugi wycinek (`getMethod`/`getDeclaredMethod`) — DONE for this scope; R2–R4 nietknięte (2026-10-23)

- **Zakres zlecenia i zależności:** zgoda użytkownika na "najbardziej
  naturalną kontynuację" R1. Przed startem zbadano realny koszt:
  śledzenie literału `Class[]` wymaga **nowej maszynerii**, nie tylko
  recognizera — `aastore` dziś bezwarunkowo odrzuca zapisywaną
  wartość, a `dup` kopiuje `Value` przez wartość (Go struct copy), nie
  przez referencję, więc mutacja przez `aastore` nie byłaby widoczna
  w innych kopiach na stosie. Druga, odkryta w trakcie eskalacja:
  D2 w ogóle nie śledzi literałów `int` (`iconst_*`/`bipush`/`sipush`/
  `ldc` Integer — świadomie poza zakresem, "stałe napisowe/klasowe"
  tylko), a `aastore` potrzebuje indeksu. Obie eskalacje przedstawione
  użytkownikowi z wyceną kosztu przed startem; zgoda na pełny,
  skorygowany zakres.
- **Nowe pole `IsObjectArray`** (`internal/classfile/localops.go`) —
  odróżnia `anewarray`/`aastore` (obiektowe) od `newarray`/prymitywnych
  wariantów store, które współdzielą te same `LocalOpKind`.
- **Nowy `ConstKindInt`** (`internal/classfile/localops.go`) — minimalne
  śledzenie literałów `int` (`iconst_m1`..`iconst_5`, `bipush`,
  `sipush`, `ldc` Integer), celowo nie pełna arytmetyka — tylko do
  odczytu indeksu `aastore`.
- **Nowy `Kind` `KindArray`** (`internal/valueflow/value.go`) —
  jedyny `Kind`, którego payload (`*ArrayValue`) jest **wskaźnikiem
  mutowalnym w miejscu**, nie niemutowalną kopią jak każdy inny `Kind`
  — żeby `dup`'owane kopie tej samej tablicy na stosie widziały
  mutację `aastore` (prawdziwa semantyka referencji JVM, nie kopii).
  `LocalOpNewArray` (`internal/dataflow/step.go`) tworzy `KindArray`
  tylko dla `anewarray` z literalną, nieujemną długością; inaczej
  (prymityw, długość nieprowadna) — bez zmiany, stary opakowany
  `KindAlloc`. `LocalOpArrayStore` mutuje `Elements[idx]` w miejscu
  tylko dla `aastore` z literalnym indeksem w granicach tablicy.
- **Nowy `Kind` `KindReflectiveMethod`** i `internal/dataflow/
  reflectapi.go`'s `reflectiveMethodValue` — rozpoznaje `Class#
  getMethod`/`getDeclaredMethod` na znanym receiverze (`KindConst`/
  `java.lang.Class`), z literalną nazwą i w pełni literalną tablicą
  `Class[]`; zwraca owner FQCN, nazwę, typy parametrów,
  `getMethod`/`getDeclaredMethod` rozróżnione (`Declared` bool).
  Typ prymitywny parametru (np. `int.class`, kompilowany jako
  `getstatic Integer.TYPE`, nie literał `ldc`) poprawnie nie jest
  rozwiązywany — nie fabrykowane.
- **Dowód na realnym bytecode:** `bin/ariadne -offline` na petclinic
  i WebGoat (pełna hierarchia 29k/40k klas) — **brak crasha**, liczby
  SERVICE_LOADING S1 niezmienione (żadna z aplikacji nie używa
  `getMethod`/`getDeclaredMethod` na `java.lang.Class` — jedyne
  trafienia `getMethod()` w kodzie aplikacji to niezwiązane API
  żądania HTTP). `REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN`
  niezmienione. Dwa przebiegi na obu dają bajt-identyczny
  `results.json`.
- **Spełnione kryteria:** nowe testy jednostkowe: `internal/classfile`
  (zmiana istniejącego `TestDecodeLocalOps_ldcStringAndClass` →
  `..._ldcStringClassAndInteger`), `internal/dataflow`:
  `TestReflectiveMethodValue_literalGetMethodResolves`,
  `TestReflectiveMethodValue_twoElementArrayResolves` (potwierdza
  mechanizm współdzielonego wskaźnika przy więcej niż jednym
  `dup`+`aastore`), `TestReflectiveMethodValue_zeroArgArrayResolves`,
  `TestReflectiveMethodValue_nonLiteralNameIsNotFabricated`.
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet
  ./...`, `go test ./...`, `go test -race ./...` (wszystkie zielone);
  `bin/ariadne -offline` na petclinic (dwa przebiegi) i WebGoat (dwa
  przebiegi), opisane wyżej.
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Reflection
  (Class/Method/Constructor)" rozszerzony o drugi wycinek).
  `README.md` (akapit "Reflection under-approximates" rozszerzony).
- **Niespełnione/odłożone:** `Method.invoke`/`Constructor.newInstance`
  (R2, realne krawędzie grafu) — nietknięte, zależy od R0+R1 w pełni
  (teraz oba gotowe dla swojego zakresu). R3 (przepływ między
  metodami, `setAccessible`), R4 (kwalifikacja) — nietknięte. `KindArray`
  obsługuje tylko prostą, liniową konstrukcję literału (bez przechowania
  w zmiennej lokalnej między osobnymi grupami `dup`/`aastore` poza
  tym, co mechanizm wskaźnikowy daje "za darmo") — nietestowane poza
  tym, co projekt czyni poprawnym z definicji.
- **Następny krok:** R2 (`Method.invoke`/`Constructor.newInstance`,
  realne krawędzie grafu — odblokuje F3's pełną rezolucję dispatchu),
  rejestr B3, albo `@Value`-owe właściwości — wymaga propozycji i
  zgody użytkownika przed startem.

### REFLECTION:R2 — pierwsza połowa (`Method#invoke` → realne krawędzie) — DONE for this scope; `Constructor.newInstance`, R3–R4 nietknięte (2026-10-23)

- **Zakres zlecenia i zależności:** kontynuacja uzgodnionej ścieżki
  R0+R1→R2 (oba wycinki R1 gotowe wyżej). Ten przyrost obejmuje
  wyłącznie `Method#invoke`; `Constructor.newInstance` to R2's własna
  druga połowa (alokacja, nie wywołanie), świadomie odłożona, nie
  pominięta przez przeoczenie.
- **Nowy pakiet `internal/reflectinvoke`:** `sites.go` (`FindInvokeSites`
  — każde miejsce wywołania `java.lang.reflect.Method#invoke(Object,
  Object...)`); `resolve.go` (`ResolveInvokes`/`resolveOne` — czyta
  Method-receiver z powrotem do R1's `valueflow.KindReflectiveMethod`,
  z tym samym jednostopniowym pościgiem przez `KindField`/
  `KindCallResult`, jaki `internal/methodhandle`'s własny
  `resolveOne` już stosuje dla receivera `MethodHandle`; receiver,
  który jest parametrem metody, pozostaje nierozwiązany, nie
  zgadywany); `edges.go` (`FindInvokeEdges`/`EdgeTargets` — rozwiązuje
  realny cel dyspozycji po **parametrach**, nie deskryptorze pełnym,
  bo `getMethod` nigdy nie dowodzi typu zwracanego).
- **Nowa maszyneria w `internal/hierarchy`:** `ClassInfo.
  MethodByNameAndParams`/`Hierarchy.ResolveByParams` — odpowiednik
  `MethodByNameAndDesc`/`ResolveExact`, ale dopasowujący tylko po
  liście parametrów (ten sam porządek wyszukiwania JVM), bo `getMethod`
  nigdy nie podaje typu zwracanego do porównania. Nowy eksportowany
  `classfile.ParamDescriptors` (opakowanie istniejącego
  `parseParamDescriptors`) dostarcza mu listę deskryptorów bez pełnego
  deskryptora metody.
- **Nowy `EdgeBasis` `BasisReflection`** (`internal/callgraph/graph.go`)
  — krawędź "może zadziałać", ten sam rygor co `BasisCallback`/
  `BasisMethodHandle`: samo posiadanie obiektu `Method` nigdy nie
  dowodzi, że `invoke` faktycznie wykona się w runtime (złapany
  `ReflectiveOperationException`, martwa gałąź).
- **Rozwiązywanie dyspozycji, nie tylko identyczność:** cel krawędzi to
  realny typ alokacji obiektu docelowego (pierwszy argument `invoke`),
  gdy przepływ wartości go dowodzi (`ObjType`), a `Class`, na którym
  wykonano `getMethod`/`getDeclaredMethod`, tylko jako fallback — ta
  sama zasada rozdziału "żądany typ vs realny cel", którą
  DYNAMIC_LINKAGE L2 już stosuje dla `RequestedOwner`. Krawędź łączy
  się z *każdym* przeciążeniem rozwiązanej nazwy metody na rozwiązanej
  klasie deklarującej (ten sam overload-insensitive kontrakt co
  pozostałe syntetyczne generatory krawędzi w tym kodzie) — sama
  precyzja wyboru klasy deklarującej już jest exact po parametrach,
  nie po samej nazwie.
- **Uczciwe "Gap", nigdy fabrykacja:** receiver nierozwiązany, typ
  parametru niebędący prostym literałem klasy (tablica, prymityw),
  żadna klasa nie deklaruje metody o tej nazwie/liczbie parametrów w
  całej hierarchii dispatch roota, albo **dowiedziona** niezgodność
  liczby argumentów (literalna tablica `Object[]` krótsza/dłuższa niż
  lista parametrów rozwiązanej metody) — każdy z tych czterech
  przypadków raportowany jako `Gap` z konkretnym powodem, zero
  fabrykowanych krawędzi.
- **Dwie nowe sekcje raportu:** "Reflective Method#invoke call sites"
  (każde miejsce wywołania, receiver rozwiązany albo z powodem braku)
  i "Reflective invoke call-graph edges" (każdy rozwiązany Method →
  realny cel albo Gap) — `internal/report/report.go` (`Data.
  ReflectInvokeSites`/`ReflectInvokeEdges`, `Build`'s nowe dwa
  parametry), `report.md.tmpl`.
- **Dowód na realnym bytecode — efekt NIEzerowy, inaczej niż R0/R1:**
  `bin/ariadne -offline -profile spring` na petclinic: 619 miejsc
  `Method#invoke`, 5 rozwiązanych do realnej krawędzi
  (`AprLifecycleListener→SSL#randSet`/`initialize`,
  `SessionRemote→Engine#getInstance`, dwa miejsca `HCANNHelper` w
  warstwie refleksji Hibernate) — `REACHABLE/FRAMEWORK/NOT_REACHABLE/
  UNKNOWN/NOT_APPLICABLE` idzie z **3/59/29/51/6 na 3/62/26/51/6** (3
  przypadki NOT_REACHABLE→FRAMEWORK). WebGoat (`-profile spring`): 628
  miejsc, 4 rozwiązane do 3 krawędzi — z **40/43/21/56/0 na
  42/44/18/56/0**. Dwa przebiegi offline na każdej aplikacji dają
  bajt-identyczny `results.json`. `go build ./...`, `go vet ./...`,
  `go test ./...` wszystkie zielone (23 pakiety, włącznie z nowym
  `internal/reflectinvoke`).
- **Spełnione kryteria:** nowe testy jednostkowe `internal/reflectinvoke`
  (`sites_test.go`, `resolve_test.go` — receiver rozwiązany/nierozwiązany
  przez pole/wynik wywołania/parametr, `edges_test.go` — realna krawędź,
  wszystkie cztery przypadki Gap) i `internal/hierarchy/
  resolve_params_test.go` (`ResolveByParams`/`MethodByNameAndParams`,
  dziedziczenie, brak dopasowania).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Reflection
  (Class/Method/Constructor)" rozszerzony o R2 i zmierzony, niezerowy
  efekt). `README.md` (akapit "Reflection under-approximates"
  rozszerzony o trzeci wycinek).
- **Niespełnione/odłożone (zgodnie z własnym zakresem, nie przez
  przeoczenie):** `Constructor.newInstance` (R2's własna druga
  połowa — alokacja, nietknięta), `getDeclared*`'s własny bypass
  kontroli dostępu/`setAccessible` (R3), kwalifikacja (R4), Method
  receiver docierający jako parametr metody (pościg przez pole/wynik
  wywołania działa, przez parametr — nie). FRAMEWORKS:F3's pełna
  rezolucja dispatchu proxy (`internal/aop`'s `ProxyHandlers`
  podłączone do tych krawędzi) — nie podłączone w tym przyroście, to
  F3's własny następny wycinek.
- **Następny krok:** `Constructor.newInstance` (R2's druga połowa),
  FRAMEWORKS:F3's podłączenie proxy-handlerów do tych krawędzi, R3
  (`setAccessible`/przepływ między metodami), rejestr B3, albo
  `@Value`-owe właściwości — wymaga propozycji i zgody użytkownika
  przed startem.

### REFLECTION:R2 — druga połowa (`Constructor#newInstance` → realne krawędzie) — DONE for this scope; R3–R4 nietknięte (2026-10-23)

- **Zakres zlecenia i zależności:** zgoda użytkownika ("druga polowa
  r2") na dokończenie R2 w całości. Przed startem sprawdzono, że R1
  nigdy nie zbudował odpowiednika `KindReflectiveMethod` dla
  konstruktorów — `getConstructor`/`getDeclaredConstructor` nie były
  wcześniej rozpoznawane wcale, więc ten przyrost musiał najpierw
  dobudować własny, węższy R1-podobny krok (rozpoznanie wartości), nie
  tylko R2's own edge-building. Żadnej nowej maszynerii tablicowej nie
  było jednak potrzeba: `KindArray`/`ConstKindInt` z R1's drugiego
  wycinka są reużyte bez zmian, bo `getConstructor(Class...)` kompiluje
  się do tego samego literalnego `Class[]` idiomu.
- **Nowy `valueflow.KindReflectiveConstructor`/`ReflectiveConstructorValue`**
  (`internal/valueflow/value.go`) — `ReflectiveMethodValue`'s własny
  odpowiednik, bez pola `Name` (konstruktor jest zawsze `"<init>"`, bez
  osobnego argumentu nazwy do rozpoznania). Nowy recognizer
  `reflectiveConstructorValue` (`internal/dataflow/reflectapi.go`),
  wpięty w `stepCall` (`step.go`) tą samą ścieżką co
  `reflectiveMethodValue`.
- **Nowy pakiet-rozszerzenie `internal/reflectinvoke`:**
  `newinstance_sites.go`/`newinstance_resolve.go`/`newinstance_edges.go`
  — ten sam trzyetapowy podział co `Method#invoke`'s własny
  `sites.go`/`resolve.go`/`edges.go`, ale **bez** `ResolveByParams`'s
  własnego przeszukania hierarchii: konstruktor nigdy nie jest
  dziedziczony (JVMS 5.4.3.3, "a class or interface's instance
  initialization method is never inherited"), więc
  `FindNewInstanceEdges` szuka `"<init>"` wyłącznie na samej
  `OwnerFQCN` (`ClassInfo.MethodByNameAndParams`), nigdy na nadklasie,
  i przenosi znalezionej metody **cały, rzeczywisty deskryptor**
  (zawsze `...)V`) bezpośrednio na krawędź grafu — bez
  overload-insensitive "każdy przeciążenie", bo rozwiązane typy
  parametrów już wybrały dokładnie jeden konstruktor. Receiver
  rozwiązywany tym samym jednostopniowym pościgiem przez
  `KindField`/`KindCallResult` co `Method#invoke`'s własny
  `resolveOneNewInstance` (parametr metody wciąż pozostaje
  nierozwiązany — ta sama, nieusunięta granica co w pierwszej
  połowie).
- **Nowy `internal/callgraph/reflectnewinstance_edges.go`:**
  `ApplyReflectNewInstanceEdges` dodaje krawędź do dokładnego
  `"<init>"TargetDesc`, plus powiązaną krawędź do `"<clinit>()V"`, jeśli
  istnieje — ta sama para Kind/Basis (`BasisReflection`, `CallSynthetic`
  dla konstruktora i `CallInit` dla `<clinit>`), jaką
  `ApplyServiceLoaderEdges` już stosuje dla SERVICE_LOADING S2's
  "cele konstruktora ... oraz związane inicjalizacje" — `newInstance`
  wywołuje inicjalizację klasy tak samo jak `new` (JVM_GRAPH G3).
- **Dwie nowe sekcje raportu:** "Reflective Constructor#newInstance
  call sites" i "Reflective newInstance call-graph edges" —
  `internal/report/report.go` (`Data.ReflectNewInstanceSites`/
  `ReflectNewInstanceEdges`, `Build`'s nowe dwa parametry),
  `report.md.tmpl`.
- **Dowód na realnym bytecode — realne krawędzie, ale zero
  dodatkowego efektu na reachability:** `bin/ariadne -offline
  -profile spring` na petclinic: 264 miejsca `Constructor#newInstance`,
  6 rozwiązanych do 7 krawędzi (`GafferUtil#newGafferConfiguratorInstance
  →GafferConfigurator#<init>`, `ContainerProvider#getWebSocketContainer
  →WsWebSocketContainer#<init>`, oraz miejsca w AspectJ/H2/Micrometer) —
  **`REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN/NOT_APPLICABLE`
  niezmienione** (3/62/26/51/6, identyczne jak po pierwszej połowie).
  WebGoat: 264 miejsca, 5 rozwiązanych do 5 krawędzi — niezmienione
  (42/44/18/56/0). Uczciwie odnotowane: nowo-konstruowalne klasy nie
  leżą akurat na ścieżkach podatnych metod w żadnej z dwóch aplikacji
  kontrolnych — to realna, zmierzona krawędź grafu, nie fikcja, po
  prostu bez obecnego wpływu na wynik CVE na tych dwóch aplikacjach.
  Dwa przebiegi offline na każdej dają bajt-identyczny `results.json`.
  `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race
  ./...` wszystkie zielone (23 pakiety).
- **Spełnione kryteria:** 3 nowe testy jednostkowe `internal/dataflow`
  (`TestReflectiveConstructorValue_literalGetConstructorResolves`,
  `..._zeroArgDeclaredResolves`, `..._nonLiteralParamIsNotFabricated`);
  nowe testy `internal/reflectinvoke`
  (`newinstance_sites_test.go`/`newinstance_resolve_test.go`/
  `newinstance_edges_test.go` — pozytywna ścieżka, receiver
  nierozwiązany, dokładny match, niezgodność liczby argumentów,
  parametr-tablica, brak takiego konstruktora, klasa poza hierarchią,
  receiver nierozwiązany nie wnosi niczego, konwersja na krawędź) i
  `internal/callgraph` (`TestApplyReflectNewInstanceEdges` — krawędź
  do `<init>` i `<clinit>`, brak podwójnego liczenia przy ponownym
  zastosowaniu).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "Reflection
  (Class/Method/Constructor)" rozszerzony o R2's drugą połowę).
  `README.md` (akapit "Reflection under-approximates" rozszerzony o
  czwarty wycinek).
- **Niespełnione/odłożone (zgodnie z własnym zakresem, nie przez
  przeoczenie):** `getDeclared*`'s własny bypass kontroli dostępu/
  `setAccessible` (R3), kwalifikacja (R4), Method/Constructor receiver
  docierający jako parametr metody (nietknięte w obu połowach R2).
  FRAMEWORKS:F3's pełna rezolucja dispatchu proxy — wciąż nie
  podłączone, to F3's własny następny wycinek, niezależny od tego
  przyrostu.
- **Następny krok:** R3 (`setAccessible`/przepływ między metodami),
  FRAMEWORKS:F3's podłączenie proxy-handlerów do tych krawędzi,
  rejestr B3, albo `@Value`-owe właściwości — wymaga propozycji i
  zgody użytkownika przed startem.

### FRAMEWORKS:F2 — szósty wycinek (`@Value`-owe punkty wstrzyknięcia) — DONE for this scope (2026-10-23)

- **Zakres zlecenia i zależności:** zgoda użytkownika na rekomendację
  po odrzuceniu dwóch droższych opcji. Przed startem zbadano
  REFLECTION:R3 (parameter-flow tracking) — nowa infrastruktura
  interproceduralna, najprawdopodobniej zerowy efekt na obu
  aplikacjach (proxy-forwarding kończy się na nieobecnym w hierarchii
  interfejsie JDK; biblioteki wewnętrzne używają runtime-computed
  nazw) — i FRAMEWORKS:F3's podłączenie proxy-dispatchu (realny,
  jedyny przypadek na WebGoat celuje w `java.sql.Connection`, którego
  nie ma w analizowanej hierarchii zależności). Obie odrzucone jako
  niekorzystny koszt/wartość. `@Value`-owe właściwości wybrane jako
  tańsza, bezpośrednio w domenie Springa (obie aplikacje kontrolne).
- **Zmienione pliki:** `internal/profile/di.go` (`valueFQCN`, nowe
  pole `InjectionPoint.ValueExpression`, nowy `valueExpression`
  (odpowiednik `qualifierValue`), pole-skan rozszerzony — `@Value`
  samo, bez `@Autowired`, teraz też wyzwala punkt — i
  `constructorInjectionPoints` przechwytuje `@Value` na parametrze tą
  samą ścieżką co `@Qualifier`; `ResolveInjections` odcina
  `ValueExpression`-owy punkt **przed** wyszukiwaniem beanów —
  `Resolved=""` (słusznie — realnie nie ma beana), ale `Reason` mówi
  teraz `a @Value config value ("..."), not a bean lookup`, nigdy
  generyczne "no known bean"); `internal/report/report.go`
  (`InjectionRow.ValueExpression`); `report.md.tmpl` (kolumna
  Resolved renderuje `config value: \`...\`` dla takich wierszy).
- **Zachowanie dostępne użytkownikowi:** raportowa tabela "Dependency
  injection" nie myli już punktu config-value z nierozwiązanym
  wstrzyknięciem beana — każdy `@Value` dostaje własny, uczciwy,
  odrębny wiersz.
- **Dowód na realnym bytecode — ujawnienie wcześniej niewidocznego
  faktu, nie tylko lepsze opisanie istniejącego:** `bin/ariadne
  -offline -profile spring` na WebGoat: punkty wstrzyknięcia skaczą z
  **43 do 45** — pole-samo-`@Value` nigdy wcześniej nie było skanowane
  w ogóle (pętla po polach wymagała `@Autowired`), więc
  `LessonMenuService`'s własne pola `excludeCategories`/
  `excludeLessons` (SpEL `#{'${exclude.categories}'.split(',')}`,
  dzielące property na `List<String>`) były **całkowicie niewidoczne**
  dla `InjectionPoints`, nie tylko błędnie skategoryzowane. Po
  dodaniu: **37 rozwiązanych** (niezmienione), **8** nierozwiązanych w
  pełni wyjaśnionych: **3** realne `@Value` (poprzednia sesja
  zaraportowała tylko 1 — te dwa nowo znalezione pola błędnie trafiły
  do "4 `List<Bean>`", bo nie były jeszcze widoczne jako osobna
  kategoria), **4** realne `List<Bean>` (parametry konstruktora,
  odrębne od tych pól — `CourseConfiguration`×2, pozostałe
  `LessonMenuService`'s konstruktorowe), **1** framework-provided
  bean. petclinic bez zmiany (8 punktów, 8/8, zero `@Value` w kodzie).
  `REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN` niezmienione na obu
  (3/62/26/51/6 petclinic, 42/44/18/56/0 WebGoat) — `@Value` to fakt
  konfiguracyjny, nigdy krawędź grafu, ten sam rygor co reszta tej
  fazy. Dwa przebiegi offline na każdej aplikacji dają bajt-identyczny
  `results.json`. `go build ./...`, `go vet ./...`, `go test ./...`,
  `go test -race ./...` wszystkie zielone.
- **Spełnione kryteria:** 3 nowe testy jednostkowe `internal/profile`:
  `TestInjectionPoints_valueFieldWithoutAutowired`,
  `TestInjectionPoints_valueConstructorParam`,
  `TestResolveInjections_valueExpressionIsNotUnresolved`. Wszystkie
  istniejące testy DI przeszły bez zmian (sygnatura
  `InjectionPoints`/`ResolveInjections` niezmieniona, tylko nowe pole
  struktury).
- **Zmiany w dokumentacji:** `docs/SUPPORT.MD` (wiersz "DI/AOP/
  configuration execution models" rozszerzony o szósty wycinek i
  korektę liczb 43/37/6 → 45/37/8). `README.md` (akapit "DI"
  zaktualizowany o `@Value`'s własne zachowanie i poprawione liczby).
- **Niespełnione/odłożone:** `List<Bean>`/`Map<String,Bean>` collection
  injection, framework-provided beans jako katalog, integracja z XML
  `<bean>`, AOP/pointcut, REFLECTION:R3, FRAMEWORKS:F3's dalszy zakres,
  rejestr B3 — nietknięte, zgodnie z własnym, wąskim zakresem tego
  wycinka.
- **Następny krok:** rejestr B3, REFLECTION:R3 (`setAccessible`/
  przepływ między metodami — zmierzony wcześniej jako
  prawdopodobnie kosztowny i niskiej wartości), albo FRAMEWORKS:F3's
  dalszy zakres (AOP pointcut/advice) — wymaga propozycji i zgody
  użytkownika przed startem.

### ROADMAP:B3 — sama decyzja architektoniczna i kontrakt (`internal/models`) — DONE for this scope; koordynator iteracji NIE (2026-10-23)

- **Zakres zlecenia i zależności:** zgoda użytkownika po przeczytaniu
  ROADMAP.MD na miejscu i ustaleniu, że pełne B3 (rejestr modeli +
  koordynator iteracji z kolejką pracy i punktem stałym) to
  wieloetapowe przedsięwzięcie architektoniczne, nie jednosesyjny
  przyrost jak REFLECTION/FRAMEWORKS. Zgoda ograniczona wyłącznie do
  decyzji o kierunku zależności pakietów i samego kontraktu modelu —
  bez koordynatora.
- **Decyzja architektoniczna (ROADMAP.MD's własny prerequisite
  "ustalić kierunek zależności pakietów przed refaktorem, aby uniknąć
  cyklu profile → callgraph → models → profile"):** `internal/models`
  może zależeć wyłącznie od pakietów-liści bez własnej zależności
  zwrotnej do `internal/profile`/`internal/callgraph` —
  `internal/classfile`, `internal/input` (potwierdzone: `internal/
  input` nie importuje żadnego innego pakietu `internal/` — prawdziwy
  liść), `internal/bytecodeir`. Nigdy nie importuje `internal/profile`
  ani `internal/callgraph` samemu. Przyszły koordynator iteracji
  (`internal/analysis`, nie zbudowany tutaj) jest jedynym pakietem,
  któremu wolno zależeć jednocześnie od `internal/models` i
  `internal/callgraph`; `internal/profile` może zależeć od `internal/
  models` (wybór ID modeli dla aktywnego profilu), nigdy odwrotnie, i
  nigdy od `internal/analysis` — profile zostaje "dostawcą
  konfiguracji i wyboru modeli, nie drugim koordynatorem"
  (dosłowne słowa ROADMAP.MD). Graf zależności zostaje DAG-iem: models
  ← analysis → callgraph, profile → models, nic nie wskazuje z powrotem
  w profile.
- **Nowy pakiet `internal/models` — kontrakt, nie silnik:** `Signature`
  (Owner/Name/Desc dokładnie jak w bytecode, ten sam konwencja co
  `valueflow.CallResultRef`/`FieldRef`) — dopasowanie zawsze po
  dokładnym wywołaniu, nigdy po samej nazwie metody w całej aplikacji
  (B3's własne explicite wymaganie). `VersionRange` (Min/Max,
  honestly nieudawający pełnego semver — tagi pre-release/metadata to
  świadomie nieobsłużony przypadek, patrz niżej). `Model` (ID,
  Artifact — Maven "group:artifact", Versions, Signatures, Source —
  B3's własne "identyfikator, obsługiwane sygnatury, zakres wersji,
  źródło semantyki", wszystkie cztery pola jednocześnie). `Fact`
  (ModelID, Incomplete, Reason — kształt, jaki przyszła ewaluacja
  modelu będzie zwracać; nic w tym pakiecie jeszcze nie produkuje
  `Fact`). `Registry` (`Register`/`Lookup` — `Lookup` dopasowuje po
  dokładnym `Signature`, a gdy `Model.Artifact != ""`, też po
  `input.GAV`'s własnym `PackageName()`+`Version` wewnątrz
  `VersionRange`; `Model` bez `Artifact` dopasowuje niezależnie od
  GAV — API czysto JDK-owe bez koordynatu Maven; `gav == nil`
  (klasa aplikacji) dopasowuje tylko modele bez `Artifact` —
  nigdy zgadywane).
- **Wersja nieudawana za pełną:** `inclusiveVersionInRange` porównuje
  tylko kropkowane segmenty numeryczne (np. "5.3.21") — wersja, której
  nie da się tak sparsować (kwalifikator typu "-RELEASE", segment
  nienumeryczny) **nie jest wykluczana**, tylko włączona (B0's własne
  "każdy limit poszerza do nieznanej wartości z diagnostyką, nigdy po
  prostu nie odrzuca kandydata") — pełny semver pozostaje świadomie
  nieobsłużony, nie udawany.
- **Spełnione kryteria (częściowe — tylko kontrakt):** "rejestr
  modeli z identyfikatorem, obsługiwanymi sygnaturami, zakresem
  wersji i źródłem semantyki" — spełnione. "Obsługę znanego API
  przypinać do konkretnego wywołania/obiektu, nie do wszystkich metod
  o tej samej nazwie" — spełnione przez samą konstrukcję `Signature`.
  "Budżety i stan incomplete z przyczyną" — kształt (`Fact`) istnieje,
  ale nic go jeszcze nie produkuje. **Niespełnione, świadomie, nie
  przez przeoczenie:** "Graf, odkryte alokacje i streszczenia
  wartości wpływają na siebie... kolejka pracy i kontrolowany punkt
  stały" (koordynator), "Przy poprawianiu modelu unieważniać zależne
  wyniki" (invalidation), RTA/points-to update rules
  (GRAPH_PRECISION's własny zakres). **"Gotowe" z B3's własnego planu
  ("model usługi może wprowadzić typ, który uruchamia kolejne
  poprawne rozwinięcie grafu, a limit zatrzymania jest widoczny") NIE
  jest spełnione** — to wymaga koordynatora, nie samego kontraktu.
- **Zero wpływu na działanie binarki:** `internal/models` nie jest
  podłączony do `cmd/ariadne/main.go`, `internal/callgraph` ani
  `internal/profile` w żaden sposób — potwierdzone (`go build`
  niezmieniony dla żadnego innego pakietu, `bin/ariadne` przebudowane
  bez zmiany zachowania, nic nowego do zmierzenia na petclinic/
  WebGoat, bo nic realnie nie działa jeszcze na realnym grafie).
- **Polecenia rzeczywiście wykonane:** `go build ./...`, `go vet
  ./...`, `go test ./...`, `go test -race ./...` (wszystkie zielone,
  24 pakiety, włącznie z nowym `internal/models`).
- **Spełnione kryteria (testy):** nowe testy jednostkowe `internal/
  models`: `TestRegistry_lookupMatchesExactSignatureOnly` (nazwa sama
  nie wystarcza — inny owner/inny deskryptor nie dopasowuje),
  `TestRegistry_artifactAndVersionGateTheMatch` (w zakresie/za stary/
  za nowy/złe artifact/brak GAV), `TestRegistry_
  noArtifactModelMatchesRegardlessOfGAV`,
  `TestInclusiveVersionInRange_unparseableVersionWidensRatherThanExcludes`,
  `TestInclusiveVersionInRange_boundaryValuesAreInclusive`,
  `TestInclusiveVersionInRange_missingTrailingSegmentReadsAsZero`.
- **Zmiany w dokumentacji:** tylko ten plik — `docs/SUPPORT.MD`/
  `README.md` celowo nietknięte, bo ten przyrost nie zmienia żadnej
  obserwowalnej dla użytkownika zdolności (nic nie jest podłączone;
  nie ma nic nowego do opisania w dokumentacji user-facing). To jest
  świadoma decyzja, nie przeoczenie reguły "kod + weryfikacja +
  dokumentacja w tym samym commicie" z catch-up skill — ta reguła
  dotyczy zmian zachowania, a to jest wyłącznie kontrakt.
- **Niespełnione/odłożone (zgodnie z własnym, świadomie wąskim
  zakresem tej sesji, nie przez przeoczenie):** `internal/analysis`
  (koordynator iteracji, kolejka pracy, punkt stały) — nietknięty,
  to jest następny, znacznie większy krok B3. Żaden realny model
  biblioteki (np. Spring `BeanUtils.copyProperties`) nie jest jeszcze
  zarejestrowany — `internal/models` eksportuje tylko typ, zero
  instancji `Model` istnieje poza testami.
- **Następny krok:** `internal/analysis` (koordynator iteracji —
  wymaga własnej propozycji zakresu, to największy pozostały krok
  B3), REFLECTION:R3, albo FRAMEWORKS:F3's dalszy zakres (AOP
  pointcut/advice) — wymaga propozycji i zgody użytkownika przed
  startem.

## Punkt startowy następnej sesji

1. Porównaj ten stan z bieżącym kodem i zmianami w checkoutcie.
2. JVM_GRAPH (G0–G5) jest zarchiwizowane jako DONE — zob.
   [archive/2026-10-22/jvm_graph/COMPLETION.md](../archive/2026-10-22/jvm_graph/COMPLETION.md).
   Przeczytaj ROADMAP:R0, B0–B1 oraz właściwe części CVE:F0/F1, które
   wciąż czekają (nieodnotowane), zanim R1 (DATA_FLOW/SERVICE/ASYNC)
   zostanie potraktowane jako w pełni odblokowane. FRAMEWORKS:F1 jest
   zarchiwizowane jako DONE dla swojego faktycznego zakresu — zob.
   [archive/2026-10-23/frameworks/FRAMEWORKS_F1.md](../archive/2026-10-23/frameworks/FRAMEWORKS_F1.md).
   FRAMEWORKS:F2 ma pięć wycinków gotowych (`@Autowired`, web.xml
   roots, `@Bean` factory methods, katalog zewnętrznych punktów
   wejścia, Spring Data repository beans) plus korektę fałszywych
   trafień niejawnej reguły "jedyny konstruktor" — zob. wpisy wyżej.
   Prawdziwe, skorygowane liczby na WebGoat: 43 punkty, 37
   rozwiązanych, 6 w pełni wyjaśnionych. FRAMEWORKS:F3 ma pierwszy,
   najwęższy wycinek gotowy (tożsamość proxy/target/interceptor,
   `internal/aop`) — bez rezolucji dispatchu, co teraz REFLECTION:R2
   umożliwia, ale F3 jeszcze nie podłączył. REFLECTION:R0+R1+R2 jest
   gotowe **w całości** ("R0–R2 stanowią pierwsze wydanie", zgodnie z
   planem): `Class.forName` z literałem (reużycie `KindConst`);
   `getMethod`/`getDeclaredMethod` z literalną tablicą `Class[]` (nowy
   `KindArray` wskaźnikowy mutowalny przez `aastore`, nowy
   `KindReflectiveMethod`, minimalne śledzenie literałów `int`) —
   zmierzony efekt zerowy na petclinic/WebGoat; `Method#invoke` →
   realne krawędzie (`internal/reflectinvoke`, nowy `BasisReflection`)
   — efekt niezerowy: petclinic 5 nowych krawędzi (3/59/29/51/6→
   3/62/26/51/6), WebGoat 3 nowe krawędzie (40/43/21/56/0→
   42/44/18/56/0); `Constructor#newInstance` → realne krawędzie (nowy
   `KindReflectiveConstructor`, reużywa `KindArray` bez nowej
   maszynerii, dokładny `<init>` nie overload-insensitive, bo
   konstruktor nigdy nie jest dziedziczony) — realne krawędzie istnieją
   (petclinic 7, WebGoat 5) ale zero dodatkowego efektu na
   reachability na obu aplikacjach, uczciwie zmierzone i odnotowane.
   Zbadano i **odrzucono** dwie droższe opcje: REFLECTION:R3
   (parameter-flow tracking — nowa infrastruktura interproceduralna,
   najprawdopodobniej zerowy efekt, bo realne wzorce na obu
   aplikacjach to proxy-forwarding do nieobecnego w hierarchii
   interfejsu JDK albo biblioteki z runtime-computed nazwami) i
   FRAMEWORKS:F3's podłączenie proxy-dispatchu (jedyny realny przypadek
   na WebGoat celuje w `java.sql.Connection`, nieobecny w analizowanej
   hierarchii zależności — też zero). Zamiast nich: FRAMEWORKS:F2's
   szósty wycinek (`@Value`-owe punkty wstrzyknięcia) jest gotowy —
   WebGoat skacze z 43 do **45** punktów (2 pola `@Value` nigdy
   wcześniej nieskanowane), wszystkie 8 nierozwiązanych w pełni
   wyjaśnione (3 `@Value`, 4 `List<Bean>`, 1 framework-provided bean),
   zero wpływu na reachability (fakt konfiguracyjny, nie krawędź).
   ROADMAP:B3 — sprawdzono w ROADMAP.MD, że pełne B3 (rejestr +
   koordynator iteracji z kolejką pracy i punktem stałym) jest
   wieloetapowym przedsięwzięciem architektonicznym, nie jednosesyjnym
   przyrostem. Zgoda użytkownika ograniczona do samej decyzji o
   kierunku zależności pakietów + kontraktu: nowy `internal/models`
   (`Signature`/`VersionRange`/`Model`/`Fact`/`Registry`, dopasowanie
   po dokładnym wywołaniu + GAV, nigdy po samej nazwie) jest gotowy,
   ale **nic jeszcze nie podłączony** do `callgraph`/`profile`/
   `cmd/ariadne` — zero wpływu na działanie binarki, potwierdzone.
   Koordynator iteracji (`internal/analysis`) to następny, osobno
   wymagający zgody, znacznie większy krok. Naturalny następny krok:
   `internal/analysis`, REFLECTION:R3 (mimo zmierzonego niskiego
   prawdopodobieństwa efektu), albo FRAMEWORKS:F3's dalszy zakres (AOP
   pointcut/advice).
3. Zaproponuj najbliższą wykonalną fazę lub minimalny wspólny kontrakt, przedstaw zakres, kolejność i priorytety, a następnie zaczekaj na wyraźną zgodę użytkownika. Do tego czasu wykonuj tylko przegląd do odczytu.
4. Po zatwierdzeniu wykonaj rezultat fazy, jego weryfikację i aktualizację dokumentacji. Jeżeli jest to faza kontraktowa, nie oznaczaj przez to funkcji jako zaimplementowanej.

Wersjonowane dowody i stan repozytorium są podstawą wznowienia. Poprzedni opis czatu nie zastępuje ich weryfikacji.
