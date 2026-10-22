# Stan realizacji i przekazanie między sesjami

Ten plik jest rejestrem pracy nad nowymi planami, a nie deklaracją wsparcia produktu. Aktualne możliwości opisuje [SUPPORT.MD](../docs/SUPPORT.MD), zakres prac [ROADMAP.MD](ROADMAP.MD), a sposób prowadzenia sesji [skill](skills/ariadne-implement/SKILL.md).

## Stan początkowy

Przygotowano plany, prompt i skill. W ramach ich tworzenia nie rozpoczęto implementacji nowych faz. Istniejący kod może spełniać część kryteriów; każda kolejna sesja ma to zweryfikować zamiast zakładać, że wszystko trzeba napisać od początku.

Stany faz: `UNASSESSED`, `READY`, `IN_PROGRESS`, `NEEDS_VERIFICATION`, `DONE`, `BLOCKED`. `DONE` oznacza spełnienie wszystkich kryteriów konkretnej fazy oraz wymaganej weryfikacji i dokumentacji. `BLOCKED` wymaga konkretnego powodu i sposobu odblokowania. To etykiety tego pliku, nie stany celów w aplikacji Codex.

## Przegląd obszarów

| Obszar | Plan | Ocena nowych faz | Wznowienie |
|---|---|---|---|
| ROADMAP | [Kontrakty B0–B4 i przyrosty](ROADMAP.MD) | UNASSESSED | Uzgodnić B0–B1 z CVE/JVM; nie traktować całego R0 jako ukończonego po samym projekcie. |
| CVE | [F0–F9](CVE_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Sprawdzić wspólne wymagania F0–F1 i aktualny model wyników. |
| JVM_GRAPH | [G0–G5](JVM_GRAPH_IMPLEMENTATION_PLAN.MD) | G0, G1, G3: DONE; G2: DONE for overloads/inheritance, bridge/default-method real-bytecode verification still pending (needs a JDK, absent here); G4–G5: UNASSESSED | G4 (target Java/multi-release JAR) jest następnym naturalnym krokiem; G5 (integracja/kwalifikacja całego zakresu) zostaje na koniec. |
| DATA_FLOW | [D0–D5](DATA_FLOW_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Wymaga odpowiednich fundamentów JVM; nie zakładać istnienia CFG/value analysis. |
| SERVICE_LOADING | [S0–S4](SERVICE_LOADING_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Rejestry usług i lokalne wartości według zależności planu. |
| REFLECTION | [R0–R4](REFLECTION_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Wspólne wartości DATA_FLOW i reguły rozwiązywania JVM. |
| FRAMEWORKS | [F0–F5](FRAMEWORKS_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Wykorzystać istniejące profile; nie mylić ich z pełnym modelem runtime’u. |
| ASYNC_CALLBACKS | [A0–A4](ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Katalog API i powiązanie callbacka z właściwym obiektem. |
| DYNAMIC_LINKAGE | [L0–L4](DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Uporządkować istniejące lambda/method references przed szerszymi uchwytami. |
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

Przy dodawaniu kolejnego wpisu zapisz:

- Pełny identyfikator fazy i stan.
- Zakres zlecenia i niezbędne wspólne zależności, które faktycznie wykonano.
- Zmienione pliki i zachowanie dostępne użytkownikowi albo rezultat kontraktowy.
- Spełnione i niespełnione kryteria, z odnośnikami do dowodów.
- Polecenia rzeczywiście wykonane, ich wyniki i zakres; osobno kontrole niewykonane z powodem.
- Zmiany w SUPPORT, README, raporcie i innych dotkniętych opisach, albo uzasadnienie braku zmiany zakresu wsparcia.
- Pozostałe ograniczenia, konieczne decyzje oraz następny krok.

## Punkt startowy następnej sesji

1. Porównaj ten stan z bieżącym kodem i zmianami w checkoutcie.
2. Przeczytaj ROADMAP:R0, B0–B1 oraz właściwe części CVE:F0/F1 i JVM_GRAPH:G0.
3. Zaproponuj najbliższą wykonalną fazę lub minimalny wspólny kontrakt, przedstaw zakres, kolejność i priorytety, a następnie zaczekaj na wyraźną zgodę użytkownika. Do tego czasu wykonuj tylko przegląd do odczytu.
4. Po zatwierdzeniu wykonaj rezultat fazy, jego weryfikację i aktualizację dokumentacji. Jeżeli jest to faza kontraktowa, nie oznaczaj przez to funkcji jako zaimplementowanej.

Wersjonowane dowody i stan repozytorium są podstawą wznowienia. Poprzedni opis czatu nie zastępuje ich weryfikacji.
