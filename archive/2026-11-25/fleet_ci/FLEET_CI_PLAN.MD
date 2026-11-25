# Uruchomienie ariadne w CI dla kilkudziesięciu artefaktów (biblioteki i aplikacje) — plan

Status: **WYCOFANY 2026-11-25** — CI floty będzie robione z poziomu Tekmerion; zob. [WITHDRAWN.md](WITHDRAWN.md). Dawniej: **propozycja, czeka na zatwierdzenie** (zapisana 2026-11-15; 2026-11-19
uzupełniona o tryb biblioteki — flota to głównie biblioteki; 2026-11-23 zaktualizowana o stan
narzędzia i prototyp raportu zbiorczego). Decyzja o starcie: połowa grudnia 2026 (menedżer).
Nie jest fazą żadnego planu z [ROADMAP.MD](../../../RECOMMENDATION/ROADMAP.MD); to narzędzia obok roadmapy, potrzebne
do uruchomienia ariadne na firmowej flocie aplikacji.

## Stan narzędzia (2026-11-23)

Wszystkie obszary planu rozwoju, które nie czekają na decyzje, są zamknięte i zarchiwizowane
(zob. [IMPLEMENTATION_STATUS.MD](../../../RECOMMENDATION/IMPLEMENTATION_STATUS.MD)); otwarte zostały tylko
RUNTIME_NATIVE N2–N5 (brak celu z wieloma kontekstami ładowania); ECOSYSTEMS wycofany 2026-11-25 (Go: oddzielne AriadneGO).
Dla floty istotne od 2026-11-15:

- **Tryb biblioteki kompletny** (LIBRARY_MODE LB0–LB4, zarchiwizowany): `-roots api`,
  `exposed_api`, luka kodu klienta, ogłoszenia Spring/CDI/Micronaut jako korzenie, kto dostarcza
  zależność (`supplied_by`), zestaw 5 bibliotek z wyroczniami JVM.
- **Inne wersje zależności** (`lb3-2`): `-alt-versions newest` — dla biblioteki ocena najwyższej
  wersji w zadeklarowanym zakresie i najnowszej w linii major.minor (przybliżona, tylko v2
  `coverage.api.alternate_versions`); na 5 bibliotekach pokazuje np. „aktualizacja
  jackson-databind w linii usuwa 22 z 32 advisories”.
- **Fakty frameworka w v2** (`frameworks`): beany, `@Profile`, wstrzyknięcia, wejścia, korzenie.
- **Ocena runtime'u JDK** (CVE F8/F9): `-jdk-home`/`-runtime-file` — advisories OpenJDK dla
  docelowego Temurina (kuratorowany katalog 2026), osobno od ustaleń komponentów.
- **Snapshoty** z wersją reguł (`profile.rules`, `library.alt_versions` i in.) — wynik CI da się
  odtworzyć i porównać (`-snapshot-out`/`-snapshot-compare`).
- **Prototyp raportu zbiorczego** `cmd/ariadne-fleet` (2026-11-23, zob. punkt 3).

## Ustalenia z użytkownikiem (2026-11-15)

- **Źródło aplikacji:** pipeline CI — każda aplikacja uruchamia ariadne we
  własnym pipeline, wyniki zbierane centralnie.
- **Sieć:** dozwolona bezpośrednio — każde uruchomienie może wysyłać listę
  współrzędnych zależności (grupa:artefakt:wersja) do `api.osv.dev` i
  pobierać diffy poprawek z `api.github.com` (`GITHUB_TOKEN` dla limitu).
- **Raport zbiorczy:** Markdown.
- **LLM:** wyłączony (domyślnie; `-llm` nie jest używane).
- **Skład floty (2026-11-16):** głównie biblioteki, część aplikacji —
  biblioteki analizowane w trybie biblioteki (`-roots api`, LIBRARY_MODE
  LB0–LB4, gotowy 2026-11-19: README „Library mode in a library's CI”).

## Otwarte decyzje (do rozstrzygnięcia przed startem)

1. **System CI** (GitLab CI, Jenkins, GitHub Actions, inny) — do niego
   powstanie gotowy fragment konfiguracji; skrypt sam w sobie jest
   niezależny od CI (różni się zapis artefaktu i cache).
2. **Zbieranie artefaktów** z pipeline'ów do raportu zbiorczego —
   propozycja na start: ręcznie (pobranie artefaktów + `ariadne-fleet`);
   osobny zaplanowany job później, zależnie od uprawnień w CI.
3. **Repozytorium Mavena w CI** (nowe, 2026-11-19): tryb biblioteki czyta
   pom biblioteki z rodzicami (LB3, `supplied_by`). Kilkadziesiąt
   pipeline'ów pobierających pomy z Maven Central grozi blokadą (Central
   zablokował ariadne 2026-11-18, HTTP 429). Propozycja: `ARIADNE_MAVEN_LOCAL`
   = lokalne repozytorium Mavena runnera (po buildzie ma już wszystkie pomy)
   i — jeśli firma ma proxy (Nexus/Artifactory) — `ARIADNE_MAVEN_CENTRAL`
   na nie. Do potwierdzenia: czy jest proxy. **Dodatkowo (2026-11-23):** `-alt-versions newest`
   pobiera `maven-metadata.xml` każdej zależności biblioteki (listy wersji nie ma w lokalnym
   repozytorium runnera) — pomiar: 4–24 pobrań na bibliotekę. Propozycja: w CI włączać
   `-alt-versions` tylko przy proxy; bez proxy wyłączone (domyślnie `off`).
4. **Runtime aplikacji** (nowe, 2026-11-23): JDK runnera CI zwykle nie jest JDK wdrożenia, więc
   `-jdk-home` runnera nic nie mówi o produkcji. Propozycja: opcjonalny plik
   `ariadne-runtime.yaml` w repozytorium aplikacji (`-runtime-file`, opis docelowego runtime'u);
   bez niego `target_runtime: not_performed`. Do decyzji: czy zespoły mają go utrzymywać.

## Zakres

### 1. Krok CI — `scripts/ci-scan.sh`

- Uruchamiany w pipeline aplikacji po buildzie.
- Wykrywa typ projektu i przygotowuje wejście:
  - Maven: `mvn dependency:copy-dependencies -DincludeScope=runtime` +
    `mvn dependency:list -DincludeScope=runtime` (`-deps`, `-deps-list`);
  - Gradle: `scripts/gradle-copy-deps.init.gradle`;
  - fat jar Spring Boot / WAR / EAR: zależności w środku, bez kroków
    dodatkowych.
- **Biblioteka czy aplikacja** (tryb zapisany w metadanych wyniku):
  - WAR, EAR, fat jar Spring Boot (`BOOT-INF/`) → **aplikacja**;
  - zwykły jar (`packaging` `jar`/`bundle`, Gradle `java-library`) bez
    `Main-Class` w manifeście → **biblioteka**;
  - zwykły jar z `Main-Class` → aplikacja (narzędzie uruchamialne);
  - nadpisanie per projekt: zmienna `ARIADNE_MODE=library|app`.
  - Wielomodułowy projekt Maven/Gradle: każdy moduł-biblioteka osobno
    (osobny wynik per artefakt), moduły `pom` pomijane.
- **Aplikacja:** `ariadne -roots all` (ostrożniejszy tryb do przeglądu
  floty).
- **Biblioteka:** `ariadne -roots api` na zbudowanym jarze (nie `-sources`,
  `-javadoc`, `-tests`), z `-api-exclude` ze zmiennej
  `ARIADNE_API_EXCLUDE` (pakiety publiczne dla JVM, ale wewnętrzne, np.
  `com.firma.lib.internal`) i `ARIADNE_MAVEN_LOCAL` (decyzja 3). Pom
  biblioteki czytany z jara (Maven osadza go w `META-INF/maven/`); dla
  Gradle — wygenerowany pom publikacji przez `-library-pom`, jeśli jest
  (`generatePomFileFor…Publication`), inaczej bez `supplied_by`
  (ostrzeżenie w wyniku). Opcjonalnie `-alt-versions newest` (zmienna
  `ARIADNE_ALT_VERSIONS`, decyzja 3).
- **Aplikacja (opcjonalnie):** `-runtime-file ariadne-runtime.yaml`, jeśli plik jest (decyzja 4).
- Oba tryby: raport Markdown i `-json-format v2`; nazwa artefaktu, commit i
  gałąź ze zmiennych CI — w pliku `ariadne-results/meta.json`
  (`artifact`, `commit`, `branch`, `status`: `ok`/`failed`, `reason`), który czyta
  `ariadne-fleet`. Opcjonalnie `-snapshot-out` (odtwarzalność; petclinic: 23 MB).
- Wyniki zapisuje jako artefakt w stałym katalogu `ariadne-results/`
  (raport, JSON v2, log, metadane: aplikacja, commit, wersja narzędzia,
  data danych OSV).
- **Nie wywraca builda** — domyślnie tylko raportuje; opcjonalny próg
  (np. fail przy `REACHABLE`) włączany później per aplikacja.
- Cache OSV/GitHub (`-cache-dir`) w katalogu cache CI.
- Błąd przygotowania wejścia lub analizy zapisany jako wynik (z powodem),
  nie gubiony.

### 2. Dystrybucja binarki

- Statyczna binarka linux/amd64 (bez JDK), budowana z tagu repozytorium.
- CI pobiera konkretną wersję — wszystkie aplikacje oceniane tą samą
  wersją narzędzia (wersja w metadanych wyniku).

### 3. Raport zbiorczy — `cmd/ariadne-fleet`

**Prototyp (2026-11-23):** `go run ./cmd/ariadne-fleet -in <katalog> -out fleet.md` — czyta
każdy podkatalog z wynikiem v2 (`results.json` albo inny plik JSON z `schema_version` `2`) i
opcjonalnym `meta.json`; tryb biblioteka/aplikacja z `callgraph.roots` (`api` = biblioteka).
Zrobione z listy niżej: osobne tabele bibliotek i aplikacji z liczbami werdyktów, dla bibliotek
`REACHABLE` z `supplied_by: library` vs reszta, `SUPPLIED_DISPATCH_ONLY`, maksymalna liczba
narażonych metod API, pakiety Springa jako framework, advisories usuwane przez aktualizację w
linii (`-alt-versions`); lista `REACHABLE` po liczbie artefaktów i ważności; nieudane analizy z
powodem; wersje narzędzia i najstarsze daty danych OSV; blok interpretacji werdyktów. Testy na
syntetycznych wynikach (`testdata/fleet`). Nie zrobione: powiązanie biblioteka × aplikacje floty.

- Wejście: katalog z zebranymi `ariadne-results/` wielu pipeline'ów.
- Wyjście: jeden plik Markdown, z **osobnymi sekcjami dla bibliotek i
  aplikacji** (werdykty znaczą co innego):
  - tabela aplikacja × liczba `REACHABLE` / `FRAMEWORK` / `NOT_REACHABLE`
    / `UNKNOWN`;
  - tabela biblioteka × te same liczby, plus: ile `REACHABLE` z
    `supplied_by: library` (wersję wybiera biblioteka — naprawa po stronie
    jej zespołu) vs `user`/`undeclared` (wersję wybiera aplikacja), ile z
    `SUPPLIED_DISPATCH_ONLY` (kod zależy od obiektu, który przekazuje
    użytkownik), liczba narażonych metod API per ustalenie (liczba, nie
    lista — `exposed_api` bywa długie: do ~780 metod w fasadzie), oraz czy
    pakiety Springa liczono jako framework (`framework_packages_added`,
    startery);
  - lista `REACHABLE` posortowana po liczbie dotkniętych aplikacji, potem
    po ważności (CVE trafiające wiele aplikacji naraz na górze);
  - aplikacje z nieudaną analizą i powodem;
  - wersje narzędzia i daty danych OSV (widać, gdy aplikacje oceniano na
    różnych danych);
  - ostrzeżenie o interpretacji wprost w raporcie: `REACHABLE` = istnieje
    statyczna ścieżka (priorytet, nie dowód wykorzystania), `FRAMEWORK` =
    ścieżka przez wnętrze frameworka (ręczny przegląd), `NOT_REACHABLE` =
    brak ścieżki w modelu (nie dowód bezpieczeństwa —
    [SUPPORT.MD](../docs/SUPPORT.MD)), `UNKNOWN` = brak mapowania metody;
    dla bibliotek: `REACHABLE` = „użytkownik publicznego API biblioteki może
    dojść do podatnego kodu”, nie „aplikacja to robi”; zmierzone granice
    trybu (LB4): ścieżki przez przekazania asynchroniczne nie są śladami
    wykonania, publiczne klasy wewnętrzne liczą się jako API, ustalenia
    startera Springa to w większości `FRAMEWORK`;
  - później (nie na start): powiązanie biblioteki z aplikacjami floty,
    które jej używają (CVE `REACHABLE` w bibliotece × aplikacje z tą
    biblioteką w `-deps-list`).

### 4. Weryfikacja

- `scripts/ci-scan.sh` przepuszczony lokalnie jak w pipeline przez:
  petclinic (Maven), petclinic-gradle (Gradle), JPetStore (WAR), demo
  Spring Boot (fat jar) — aplikacje; oraz biblioteki: projekt Maven
  zwykłego jara (np. `testdata/library` przerobiony na moduł Maven) i
  biblioteka Gradle `java-library` — wykrycie trybu, `-roots api`, pom z
  jara / `-library-pom`, `ARIADNE_MAVEN_LOCAL` bez sieci do Central.
- Wyniki biblioteki porównane z bezpośrednim uruchomieniem (te same
  werdykty co `scripts/library-measure.sh` dla tej samej biblioteki).
- `ariadne-fleet` na tych wynikach; testy jednostkowe (w tym aplikacja z
  błędem analizy i dwie różne daty danych OSV).
- `go test ./...`, `go vet ./...`.
- Dokumentacja: nowa sekcja README „Uruchomienie w CI dla wielu
  aplikacji”; wpis w [IMPLEMENTATION_STATUS.MD](../../../RECOMMENDATION/IMPLEMENTATION_STATUS.MD).

## Uwagi

- Czas: ok. 2–3 min na dużą aplikację (16 rdzeni, pomiar na petclinic/
  WebGoat/JPetStore); biblioteka kilka–kilkanaście sekund (pięć bibliotek
  LB4: cały pomiar 20 przebiegów ~30 s; z `-alt-versions newest` kilka sekund więcej na
  bibliotekę); pamięci nie mierzono — warto zmierzyć przy weryfikacji, bo runner CI może mieć mniej.
- Wysłanie listy zależności do OSV/GitHub ujawnia inwentarz zależności
  firmy — użytkownik to zaakceptował; zapisane tu dla przejrzystości.
