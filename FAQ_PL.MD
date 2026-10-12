# FAQ

Notatki wyjaśniające decyzje projektowe ariadne, spisane w trakcie pracy nad
narzędziem. Uzupełnienie do [PLAN.md](PLAN.md) (historia faz) i
[README.md](README.md) (jak działa pipeline) — tutaj trafia *dlaczego*,
nie *co*.

Wersja angielska tego pliku: [FAQ.md](FAQ.md).

## Dlaczego lepiej analizować bytecode (`.class`), a nie kod źródłowy Javy?

Kilka powodów, od najważniejszego:

### 1. Zależności nie mają dostępnego kodu źródłowego — tylko bytecode

To jest argument rozstrzygający dla tego konkretnego narzędzia. Cel to
prześledzenie wywołania **z aplikacji w głąb biblioteki** (np. do
`XStream#fromXML`). Kod źródłowy swojej aplikacji się ma, ale **prawie
nigdy nie ma się kodu źródłowego zależności** w normalnym procesie
budowania — Maven ściąga tylko skompilowane jary
(`target/dependency/*.jar`), nie `-sources.jar`. Żeby analizować źródło
całego drzewa zależności (w petclinic to 87 jarów, w WebGoat 126),
trzeba by osobno dociągać `-sources` dla każdego, dopasowywać dokładnie
wersje i ufać, że się ściągnęły poprawnie. Bytecode zawsze się ma — to
właśnie to, co realnie się uruchamia.

### 2. Kompilator już rozwiązał za Ciebie całą niejednoznaczność

Kod źródłowy Javy jest pełen rzeczy, które trzeba *wywnioskować*: które
przeciążenie metody zostanie wywołane, na co wymazują się generyki, czy
autoboxing wstawi niejawne wywołanie, czy kompilator doda syntetyczny
akcesor do prywatnego pola z zagnieżdżonej klasy. Żeby to policzyć na
poziomie źródła, trzeba by napisać praktycznie własny front-end
kompilatora Javy.

W bytecode to wszystko jest już rozstrzygnięte i zapisane jawnie.
Widzieliśmy to dosłownie w trakcie prac: kiedy sprawdzaliśmy `javap` na
identycznym źródle skompilowanym dwoma kompilatorami (JDK11 vs JDK17),
bytecode różnił się (`invokespecial` vs `invokevirtual` dla wywołania
własnej prywatnej metody, `StringBuilder` vs `invokedynamic` dla
konkatenacji stringów) — kompilator podjął decyzję, narzędzie tylko ją
czyta, zamiast zgadywać.

### 3. Format jest mały, stabilny i dokładnie opisany

Specyfikacja JVM (format `.class`, zestaw instrukcji) jest wąska i od
dekad prawie niezmienna — cały parser (`internal/classfile`) powstał w
kilkuset liniach Go, bez żadnej zewnętrznej biblioteki. Gramatyka języka
Java jest ogromna i stale rośnie (rekordy, `sealed`, pattern matching,
text blocks...) — parsowanie samego źródła odpornie na to wszystko to
dużo większy i bardziej kruchy projekt.

### 4. Bytecode jest wspólnym mianownikiem dla całego JVM

Kotlin, Scala, Groovy — wszystko kompiluje się do tego samego formatu
`.class`. Analiza na bytecode automatycznie obsługuje mieszany projekt
(np. część modułów w Kotlinie), bez pisania osobnego parsera na każdy
język.

### Gdzie świadomie zrobiono inaczej

W fazie 4 (`internal/vulnmap`, mapowanie CVE → metoda) analizowany jest
**kod źródłowy**, nie bytecode — bo tam źródłem danych są diffy z
GitHuba (fix-commity), które z natury są tekstem Java, nie skompilowanym
artefaktem. Zasada ogólna: **używaj takiej reprezentacji, w jakiej dane
naturalnie istnieją**, a nie dogmatycznie trzymaj się jednej.

### Cena tej decyzji

- Bytecode jest mniej czytelny do debugowania — stąd w testach
  jednostkowych wynik własnego parsera był stale porównywany z `javap`
  jako wzorcem.
- Tracisz nazwy zmiennych, komentarze (nieistotne dla grafu wywołań,
  więc nie szkodzi).
- Trzeba samemu napisać dekoder instrukcji (`invokevirtual`,
  `invokeinterface` itd.) zamiast skorzystać z gotowego parsera AST Javy
  — to jest dokładnie to, co robi `internal/classfile`.

## Które projekty ariadne wspiera i jak je przygotować przed odpaleniem

**Na poziomie mechanizmu:** każda aplikacja skompilowana do bytecode JVM
— czyli technicznie Java, Kotlin, Scala, Groovy, bo wszystkie kompilują
się do tego samego formatu `.class` (patrz wyżej, punkt 4). W praktyce
jednak **realnie przetestowane i zweryfikowane są tylko dwa projekty**,
oba w Mavenie:

- **stary spring-petclinic** (Spring Boot 2.4.5, commit sprzed czerwca
  2021) — przypadek realistyczny, dużo CVE, mało faktycznie osiągalnych.
- **WebGoat v8.2.2** — przypadek kontrolny, celowo podatny (stary
  XStream), ma potwierdzać że narzędzie nie zawsze mówi „bezpieczne".

Wszystko inne (Kotlin, Gradle, czyste JAX-RS/Jakarta EE, Micronaut,
Quarkus) **powinno działać mechanicznie**, ale nie zostało na tym
uruchomione ani przetestowane.

### Co realnie NIE jest wspierane

- **Gradle** — nie ma gotowego sposobu wyciągnięcia zależności (dla
  Mavena jest `mvn dependency:copy-dependencies`; dla Gradle nic takiego
  nie napisano — to zaplanowana Faza 7 w [PLAN.md](PLAN.md)).
- **WAR/EAR** — tylko zwykły jar, katalog klas, albo fat jar Spring
  Boota. WAR (`WEB-INF/`) też odłożone na Fazę 7.
- **System profili** — lista pakietów „frameworkowych" (decydująca o
  statusie `FRAMEWORK`) jest na sztywno dobrana pod stos Spring/Tomcat/
  Jackson/Hibernate/Logback/Log4j. Na aplikacji z Jersey/JAX-RS albo
  Quarkusem rdzeń analizy (`REACHABLE`/`NOT_REACHABLE`/`UNKNOWN`) będzie
  działał poprawnie, ale klasyfikacja `FRAMEWORK` będzie niedokładna —
  ścieżki idące przez wewnętrzności tamtego frameworka mogą wyjść jako
  `REACHABLE`, mimo że w praktyce są sterowane refleksją tak samo jak w
  Springu.

### Jak przygotować projekt

Ariadne analizuje **skompilowany bytecode**, więc projekt musi być
najpierw zbudowany standardowym narzędziem (Maven):

```bash
# 1. Zbuduj aplikację
mvn package -DskipTests

# 2. Skopiuj zależności runtime do osobnego katalogu
mvn dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory=target/dependency

# 3. (opcjonalnie, ale zalecane) wygeneruj listę zależności jako fallback identyfikacji
mvn dependency:list -DincludeScope=runtime -DoutputFile=target/deps.txt
```

Krok 3 jest ważny praktycznie — na petclinic same `pom.properties` z
wnętrza jarów identyfikowały tylko 42/87 zależności; z `deps.txt` jako
fallbackiem — 87/87. Bez tej listy część jarów trafi do raportu jako
„unidentified" i nie zostanie w ogóle sprawdzona pod kątem CVE.

**Alternatywa — fat jar Spring Boota:** jeśli build produkuje jeden fat
jar (`target/myapp-1.0.0.jar`), wystarczy podać go samego, bez `-deps`/
`-deps-list` — ariadne sam wyciąga `BOOT-INF/classes` i
`BOOT-INF/lib/*.jar`. Minus: identyfikacja zależności jest wtedy niższa
(tylko z `pom.properties`, bo nie ma skąd wziąć `deps.txt`), więc ten
wariant znajdzie mniej CVE niż wariant z `-deps-list`.

**Czego NIE trzeba mieć na maszynie, gdzie odpala się `ariadne`:** JDK.
Jest potrzebny tylko do zbudowania analizowanej aplikacji — sam
`ariadne` czyta `.class` własnym parserem.

## Jak działa pipeline — krok po kroku

### Krok 1: `internal/input` — normalizacja wejścia

Zadanie: zamienić cokolwiek podano w `-app`/`-deps`/`-deps-list` na dwie
jednolite rzeczy, których reszta programu już nie musi rozróżniać: jedną
ścieżkę do jara appki, i listę `[]input.DepJar` (każdy element to
`{Path, GAV, Source, Warnings}`).

**`PrepareAppInput(appPath)`** rozpoznaje, z czym ma do czynienia:

- **katalog** (`target/classes`) → `packDirAsJar`: chodzi po drzewie
  (`filepath.Walk`), każdy plik dopisuje do nowego archiwum ZIP z
  zachowaniem ścieżki względnej, zapisuje do pliku tymczasowego;
- **zwykły jar** → używany wprost, zero kopiowania;
- **fat jar Spring Boota** (wykrywany po wpisie zaczynającym się od
  `BOOT-INF/`) → `prepareFromBootJar`: przepisuje `BOOT-INF/classes/**`
  do nowego jara **obcinając prefiks** (bo bez tego parser `.class`
  szukałby klasy pod złą ścieżką), a każdy `BOOT-INF/lib/*.jar`
  wypakowuje bajt-w-bajt do osobnego pliku — tak jakby ktoś ręcznie
  zrobił `mvn dependency:copy-dependencies`.

**`ResolveDepJars(depsDir, depsList)`** dla każdego jara zależności, w
tej kolejności:

1. Zajrzyj do środka jara — każdy jar z Maven Central ma w sobie
   `META-INF/maven/<groupId>/<artifactId>/pom.properties`. Działa
   niezależnie od narzędzia budowania. Komplikacja znaleziona na
   petclinic: niektóre jary (np. `ehcache-3.9.2.jar`) mają **kilka**
   takich plików (zbundlowane własne podmoduły) — wtedy wygrywa ten,
   którego `artifactId-version` pasuje do nazwy samego pliku jara, a
   resztę dopisuje się jako ostrzeżenie.
2. Jeśli jar nie ma `pom.properties`, dopasuj nazwę pliku
   (`artifactId-version[-classifier].jar`) do listy z `-deps-list`
   (`ParseDepsList` musi po drodze przeciąć kody ANSI i dopiski
   ` -- module x [auto]`, którymi nowsze wersje `mvn dependency:list`
   zaśmiecają wynik).
3. Jeśli nic nie zadziałało, jar trafia na listę „unidentified" w
   raporcie — nie jest po cichu pomijany.

### Krok 2: `internal/classfile` — parsowanie bytecode

Wejście: surowe bajty jednego pliku `.class`. Wyjście: struktura klasy
(nadklasa, interfejsy, metody, adnotacje) plus — dla każdej metody —
lista wywołań.

**`Parse()`** czyta plik sekwencyjnie według specyfikacji JVM: magic
number `0xCAFEBABE`, wersja bytecode (zbierana do raportu jako
„najwyższa widziana wersja"), **constant pool** (tablica wszystkiego, do
czego klasa się odwołuje — nazwy klas, sygnatury metod, literały,
uchwyty `MethodHandle`; wszystkie typy wpisów aż do Javy 21; wpis
`Long`/`Double` zajmuje dwa sloty, nie jeden — dziwactwo specyfikacji),
flagi/nadklasa/interfejsy, metody (nazwa, sygnatura,
`RuntimeVisibleAnnotations` parsowane rekurencyjnie, surowe bajty
`Code`), i atrybuty klasy, zwłaszcza `BootstrapMethods` (potrzebne do
lambd — patrz niżej).

**`ExtractCalls()`** chodzi po bytecode metody instrukcja po instrukcji
— każda instrukcja to bajt kodu operacji plus argument o długości
zależnej od konkretnej instrukcji (tabela dla ~200 kodów operacji, ze
specjalnymi przypadkami `tableswitch`/`lookupswitch` i prefiksu `wide`).
Interesują nas cztery rodzaje wywołań: `invokevirtual`, `invokespecial`,
`invokestatic`, `invokeinterface` — wszystkie rozwiązywane wprost z
constant poola.

**`invokedynamic`** (lambdy, referencje do metod, lub na nowszym
bytecode konkatenacja stringów) nie da się odczytać wprost — informacja
o prawdziwym celu jest schowana w `BootstrapMethods`. Jeśli bootstrap to
`LambdaMetafactory`/`altMetafactory`, program czyta z niego drugi
argument (uchwyt do prawdziwej implementacji) i tworzy normalną
krawędź grafu. Jeśli bootstrap to coś innego (np. `StringConcatFactory`),
program **świadomie nie zgaduje** — liczy to jako „nierozwiązane" i tak
pokazuje w raporcie.

Każdy element było sprawdzane ręcznie przeciwko `javap -c -p -v` na tym
samym pliku skompilowanym na Javę 8 i Javę 21. Na realnych danych: 29154
klas z petclinic i 40811 z WebGoat sparsowane bez jednego błędu.

### Krok 3: `internal/hierarchy` — mapa dziedziczenia

Wejście: jar appki + `[]input.DepJar` z kroku 1. Wyjście: struktura
`Hierarchy` znająca każdą przeczytaną klasę, z appki i zależności.

Dla każdego jara, dla każdego wpisu `.class` (pomijając
`module-info.class` i `META-INF/versions/**` — multi-release jary):
parsowanie przez `classfile.Parse`, i **w tym samym miejscu**
wywołanie `ExtractCalls` dla każdej metody — bo `classfile.ClassFile`
świadomie nie udostępnia na zewnątrz constant poola i `BootstrapMethods`,
więc to jedyny moment, w którym są jeszcze dostępne. Dzieje się to
równolegle (8 gorutyn), wyniki trafiają do mapy `FQCN → *ClassInfo`.

**Konflikty na classpath:** jeśli ta sama klasa istnieje w dwóch
różnych jarach (realny przypadek na petclinic: `jakarta.activation` vs
`jakarta.activation-api`), program nie wywala błędu — zostaje pierwsza
przeczytana, druga trafia do ostrzeżeń.

**Odwrócenie relacji dziedziczenia:** program buduje mapę
`direct: supertyp → lista podtypów`, idąc po wszystkich klasach w
**sortowanej** kolejności nazw (nie w losowej kolejności iteracji po
mapie Go — to poprawka po realnym błędzie: bez sortowania dwa
uruchomienia tego samego programu na tych samych danych dawały różne,
choć równie poprawne, ścieżki w `results.json`).

**`Subtypes(fqcn)`** zwraca wszystkich transytywnych podtypów (BFS po
mapie `direct`, wynik zapamiętywany w `subtypesCache`, bo ten sam typ
bywa pytany tysiące razy w jednym przebiegu). Typ nigdy nie przeczytany
(typowo coś z JDK, np. `java.lang.Runnable`) zwraca po prostu puste —
to jest ten właśnie mechanizm, który w kroku 4 chroni przed eksplozją
grafu.

**`ResolveMethod(fqcn, name)`** jest zaimplementowana i przetestowana
(odtwarza porządek wyszukiwania metody w JVM: klasa, nadklasy,
interfejsy), ale **żaden inny pakiet jej jeszcze nie wywołuje** —
gotowy klocek czekający na dalsze fazy, nie element dzisiejszego
działającego pipeline'u.

### Krok 4: `internal/callgraph` — budowa grafu wywołań i CHA

Wejście: `Hierarchy` z kroku 3 — każda klasa już ma swoje metody wraz z
wyciągniętymi z nich wywołaniami (`Calls`). Wyjście: `Graph` — graf
skierowany, gdzie węzły to metody, a krawędzie to „kto kogo wywołuje",
rozszerzony o wywołania polimorficzne, których bytecode sam z siebie nie
pokazuje w pełni.

**Internowanie.** Każda metoda ma swój klucz tekstowy
`"fqcn#nazwa(sygnatura)"`. Trzymanie milionów krawędzi jako par stringów
byłoby wolne i ciężkie pamięciowo, więc każdy klucz dostaje raz jeden
mały numer (`uint32`) i od tego momentu cały graf operuje na liczbach.
Budowany jest też drugi indeks: `"fqcn#nazwa"` (bez sygnatury) → lista
wszystkich metod o tej nazwie niezależnie od przeciążenia — potrzebny w
kroku 7 do dopasowania podatnej metody z `vuln_methods.yaml`, gdzie nikt
nie podaje dokładnej sygnatury.

**Dwa przebiegi budowy.** Przebieg 1: zanotuj każdą zadeklarowaną
metodę, nawet bez wywołań — krok 7 potrzebuje punktu startowego dla
każdej metody appki jako potencjalnego korzenia BFS. Przebieg 2: dodaj
krawędzie — dla `invokespecial`/`invokestatic`/`invokedynamic` wprost do
rozwiązanego celu; dla `invokevirtual`/`invokeinterface` do celu z
bytecode **i** do rozszerzenia CHA.

**CHA (class-hierarchy analysis).** Bytecode zapisuje wywołanie
interfejsu tak, jak kompilator je statycznie widział — w czasie
działania może to być którykolwiek podtyp. Statyczny graf nie wie
który, więc zachowuje się konserwatywnie: dodaje krawędź do **każdego**
podtypu (z `hierarchy.Subtypes()`), który sam deklaruje tę metodę.

**Ochrona przed eksplozją.** CHA nigdy nie rozszerza wywołań na typach
`java.*`/`javax.*`/`jdk.*`/`sun.*` (`DefaultSkip`) — bez tej ochrony
sam bootstrap frameworka wciągałby niemal cały classpath. Zmierzone na
petclinic: z ochroną graf ma 1,2 mln krawędzi i buduje się w 1,9 s; bez
niej — 11,4 mln krawędzi (9,3×) i 2 min 43 s (85× dłużej). Bez tego
praktycznie każde CVE w każdej zależności wyszłoby jako „osiągalne".

Oba przebiegi chodzą po klasach w **sortowanej** kolejności nazw (nie w
losowej kolejności iteracji po mapie Go) — druga połowa poprawki
determinizmu z kroku 3: bez tego dwa uruchomienia na tych samych danych
dawały różne, choć równie poprawne, ścieżki w `results.json`.

Liczby: petclinic — 273 180 węzłów, 1 229 600 krawędzi; WebGoat —
339 786 węzłów, 2 591 404 krawędzi.

### Krok 5: `internal/cves` — zapytania do OSV

Wejście: unikalne pary `(groupId:artifactId, wersja)` z kroku 1. Wyjście:
lista CVE, każdy zdeduplikowany, z podsumowaniem, severity, listą
dotyczących go zależności i referencjami (zwłaszcza linkami do commitów
naprawiających).

**Zapytanie wsadowe.** `POST /v1/querybatch` na `api.osv.dev` przyjmuje
wiele zapytań naraz (limit 1000/żądanie, większe listy są dzielone na
kawałki). Odpowiedź dla każdej pary pakiet+wersja to tylko **lista
identyfikatorów** podatności, bez szczegółów — celowo lekka. Jest też
obsłużona paginacja (`/v1/query` + `page_token`), choć w praktyce nigdy
nie była potrzebna.

**Szczegóły.** Dla zbioru unikalnych identyfikatorów `GET /v1/vulns/{id}`
— równolegle, 8 na raz.

**Deduplikacja aliasów.** OSV często ma dwa osobne rekordy dla tej samej
podatności (jeden pod GHSA, jeden pod CVE, wzajemnie wskazujące się w
polu `aliases`). Union-find łączy je w jedną grupę, z identyfikatora
`CVE-*` jako głównym, gdy taki istnieje.

**Severity.** Pole `severity[].score` dla `CVSS_V3` to cały wektor CVSS
jako tekst, nie liczba — pierwsza wersja raportu pokazywała to jako
tabelę z ~30 praktycznie unikalnych wierszy. Poprawka: program używa
`database_specific.severity` — czytelnej etykiety (`LOW`/`HIGH`/
`CRITICAL`...), którą OSV wypełnia dla podatności z GitHub Advisory
Database — z wektorem CVSS jako zapasowym źródłem.

Każde zapytanie przechodzi przez `internal/httpcache`; przy `-offline`
brak wpisu w cache to jasny błąd, nie ciche pominięcie.

Liczby: petclinic — 148 CVE dla 87 zależności; WebGoat — 160 CVE dla 126
zależności (w tym 34 dla samego `xstream:1.4.5`).

### Krok 6: `internal/vulnmap` — CVE → konkretne metody

Wejście: lista CVE z kroku 5. Wyjście: mapa CVE → `{Methods, Source}` —
dla CVE, które się zmapowało, konkretna lista podatnych metod. CVE bez
wpisu tutaj dostanie w kroku 7 status `UNKNOWN`.

**Źródło 1 — tabela ręczna** (`internal/vulnmap/vuln_methods.yaml`,
wbudowana przez `//go:embed`, nadpisywalna przez `-vuln-methods`).
Dopasowanie po głównym ID CVE **i** po każdym aliasie. Zawsze wygrywa,
gdy ma wpis.

**Źródło 2 — diff z fix-commitu na GitHubie.** Dla referencji w formacie
`github.com/<owner>/<repo>/commit/<sha>`: ściągnij `<link>.diff`,
zostaw tylko `.java` pod `src/main/` (bez testów), dla każdego takiego
pliku ściągnij jego **pełną treść** z commitu naprawiającego
(`raw.githubusercontent.com`), i skanerem (`internal/javadiff`) znajdź,
która metoda otacza każdą zmienioną linię. Skaner najpierw zamazuje
komentarze/stringi/text-blocki (zachowując numery linii), żeby stray
`{` nigdy nie rozjechał liczenia nawiasów, potem śledzi stos ramek
klasa/metoda/„coś innego" (bloki `if`/`for`/lambda/anonimowa klasa są
przezroczyste), i zmienia nazwę metody na `<init>`, gdy ta zgadza się z
nazwą otaczającej klasy (konstruktor).

Źródło 3 to LLM czytający advisory + diff poprawki, włączany opcją
`-llm` (projekt i dokładny prompt: [docs/phase4-llm.md](docs/phase4-llm.md),
po angielsku). Uruchamia się tylko dla CVE, których źródła 1 i 2 nie
zmapowały, a każda metoda, którą model wskaże, jest sprawdzana, czy
istnieje w analizowanym bytecode. Gdy model widział tylko tekst advisory
(bez diffu), wynik `NOT_REACHABLE` jest obniżany do `UNKNOWN`, bo model
mógł wskazać istniejącą metodę, która nie jest tą podatną.

Liczby: petclinic — 56/148 zmapowanych (1 ręcznie, 55 przez diff);
WebGoat — 104/160 (37 ręcznie — cała rodzina XStream, 67 przez diff).

### Krok 7: `internal/reach` — przeszukiwanie grafu i statusy

Wejście: graf z kroku 4, hierarchia z kroku 3, CVE z kroku 5, mapowanie
z kroku 6. Wyjście: dla każdego CVE jeden z czterech statusów, a dla
`REACHABLE`/`FRAMEWORK` — konkretna ścieżka.

**Korzenie:** wszystkie metody kodu aplikacji (`IsApp == true`), bez
wyjątku — łącznie z prywatnymi.

**Dwa przeszukania BFS, policzone raz** (nie per CVE — stąd analiza
148–160 CVE w &lt;100 ms nawet na grafie z 2,6 mln krawędzi):
- **Przeszukanie 1 (REACHABLE):** węzeł z pakietu „frameworkowego"
  (Spring Web/WebMVC, Spring Boot, Tomcat, Jackson-databind, Hibernate,
  Logback, Log4j) jest zaznaczany jako odwiedzony, ale program **nie
  idzie dalej** jego krawędziami — framework jest ślepym zaułkiem, do
  którego wolno dojść jednym skokiem, ale nie przez który wolno iść
  dalej.
- **Przeszukanie 2 (FRAMEWORK):** te same korzenie, bez ograniczenia.

Ten jeden trik („zaznacz jako odwiedzony, ale nie rozwijaj") realizuje
dokładnie wyjątek z planu: bezpośrednie wywołanie appki w metodę
frameworka wciąż liczy się jako REACHABLE, ale ścieżka wymagająca kilku
skoków *przez* framework — już nie.

**Dopasowanie metody z YAML do węzłów grafu** (`matchVulnerableMethod`):
szukaj wprost po nazwie klasy+metody (ignorując przeciążenia), **i
dodatkowo** sprawdź każdy transytywny podtyp, który nie nadpisuje tej
metody — bo wywołanie przez referencję podklasy zapisuje w bytecode
nazwę podklasy, nawet gdy metoda jest tylko dziedziczona.

**Status dla CVE:** najpierw, jeśli każda podatna zależność to WebJar
(ani jednego pliku `.class`, zasoby w `META-INF/resources/webjars/`) →
`NOT_APPLICABLE`: JavaScript/CSS po stronie przeglądarki, którego graf
wywołań nie przeanalizuje, co nie znaczy „bezpieczne”. (Sam pusty jar się
nie liczy: agregatory Mavena, jak `spring-boot-starter-web`, też są puste,
a Spring4Shell jest przypisany właśnie do niego.) Potem: brak wpisu z kroku 6 → `UNKNOWN`. Inaczej: dopasowane
węzły odwiedzone w przeszukaniu 1 → `REACHABLE` (najkrótsza taka
ścieżka). Odwiedzone tylko w przeszukaniu 2 → `FRAMEWORK`. Nigdzie →
`NOT_REACHABLE`.

**Ścieżka i „entry call":** BFS zapamiętuje dla każdego węzła, z którego
został pierwszy raz osiągnięty (wskaźnik rodzica) — odtworzenie ścieżki
to spacer od celu do korzenia i odwrócenie. „Entry call" to pierwsza
krawędź na tej ścieżce, gdzie wywołujący jest kodem appki, a wywoływany
już nie — najbardziej użyteczna pojedyncza linia dla człowieka
przeglądającego wynik.

Liczby: petclinic — 13 REACHABLE / 18 FRAMEWORK / 25 NOT_REACHABLE / 92
UNKNOWN (z 148). WebGoat (control case) — 51/32/21/56 (z 160), z całą
rodziną XStream jako REACHABLE przez dwuskokową ścieżkę
`VulnerableComponentsLesson#completed → XStream#fromXML`.

### Krok 8: `internal/report` — złożenie wszystkiego w wynik

Wejście: wszystko, co powstało wcześniej. Wyjście: `report.md` (dla
człowieka) i `results.json` (dla maszyny). Ten krok niczego nie liczy —
tylko układa gotowe dane w czytelną formę.

**Przygotowanie danych.** Wyniki są sortowane po ID CVE, a wewnątrz
każdego wyniku lista metod też — to jest właśnie mechanizm, który
gwarantuje, że dwa uruchomienia na tych samych danych dają identyczny
plik wynikowy. Liczone są: liczniki statusów, liczniki severity
(grupowane po czytelnej etykiecie z kroku 5), wiersze tabeli (jeden na
każdą kombinację CVE+zależność), i osobne listy `REACHABLE`/`FRAMEWORK`
dla sekcji z pełnymi ścieżkami.

**Renderowanie.** Szablon Markdown jest wbudowany (`//go:embed`,
`templates/report.md.tmpl`), wypełniany przez `text/template` — cała
logika była wcześniej, szablon tylko formatuje. Wersja HTML (`-html`)
była w planie jako opcja, nie została zrobiona.

Raport zawiera: nagłówek, podsumowanie (statusy + severity), pełną
tabelę CVE, sekcję pełnych ścieżek dla każdego `REACHABLE` (z
podświetlonym „entry call"), sekcję `FRAMEWORK` z opisem „wymaga
ręcznego przeglądu", listę niezidentyfikowanych jarów, i stałą sekcję
Limitations (refleksja, dynamiczne proxy, gadżety serializacji,
przeszacowania CHA — ta sama treść co w README).

`results.json` to ta sama lista wyników jako JSON z wcięciami, bez
żadnej logiki renderującej — łatwe do porównania między przebiegami
albo wczytania w innym narzędziu.

## Czy graf wywołań można zapisać/przechować?

**Nie — obecnie nigdzie się nie zapisuje.** Żyje tylko w pamięci RAM
jednego przebiegu `ariadne` i jest wyrzucany, gdy proces się kończy.
`internal/callgraph` nie ma żadnej funkcji zapisu/odczytu (zero tagów
JSON, zero `Save`/`Load`).

Co właściwie jest zapisywane: tylko wynik końcowy analizy
(`results.json` — dla każdego CVE status, metody, ścieżka) i cache
odpowiedzi HTTP z OSV/GitHuba (`.cache/`). Żadne z tych dwóch to nie
sam graf.

**Dlaczego nikt go nie zapisuje:** graf jest deterministycznie
odtwarzalny z tych samych wejść, a budowa trwa ok. 2 sekundy nawet dla
270 tys. węzłów i 1,2 mln krawędzi (zmierzone na petclinic) — jak wynik
pośredni kompilatora: policz, użyj, zapomnij.

**Czy dałoby się zapisać — technicznie tak,** bo `Graph` to w środku
proste dane (lista stringów, mapa string→int, tablica list krawędzi).
Dwie opcje: `encoding/gob` (kompaktowy, binarny, czytelny tylko dla Go)
albo JSON (czytelny dla człowieka/innych narzędzi, ale przy 1,2 mln
krawędzi plik wyjdzie dziesiątki–setki MB). Miałoby to sens, gdyby
trzeba było: (1) puszczać wiele analiz na tym samym buildzie bez
przebudowy grafu za każdym razem, (2) zrobić wizualizację grafu w
osobnym narzędziu (eksport do `.dot`), albo (3) debugować sam graf bez
przechodzenia całego pipeline'u. Żaden z tych przypadków jeszcze nie
wystąpił, więc nie zostało to zaimplementowane.
