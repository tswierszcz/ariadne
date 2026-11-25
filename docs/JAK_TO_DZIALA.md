# Jak działa ariadne

*Wersja angielska: [HOW_IT_WORKS.md](HOW_IT_WORKS.md). Jak uruchamiać: [README](../README.md).*

Ten dokument opisuje krok po kroku, co ariadne robi z danymi wejściowymi i jak czytać wyniki. Zakres obsługi wraz z dowodami każdej deklaracji jest w [SUPPORT.MD](SUPPORT.MD). Pełna historia rozwoju, ze szczegółowym opisem każdego modelu, jest w [archive/2026-11-25/docs](../archive/2026-11-25/docs/README-2026-11-25.md).

## 1. Na jakie pytanie odpowiada

ariadne odpowiada na dwa rodzaje pytań o zbudowaną aplikację lub bibliotekę JVM:

1. **Skan CVE** (`ariadne`): które znane podatności w zależnościach są rzeczywiście osiągalne z tego kodu, a nie tylko obecne na classpath.
2. **Badanie** (`ariadne research …`), bez żadnego advisory: czy dowolna metoda jest osiągalna, z których punktów wejścia i jaką ścieżką. Odpowiedź pochodzi z zamrożonej analizy, o którą mogą pytać inne narzędzia (Tekmerion).

Oba zaczynają od tego samego **statycznego pipeline'u**: grafu wywołań zbudowanego wprost z bytecode. Do uruchomienia ariadne nie jest potrzebny JDK, bo narzędzie samo parsuje pliki `.class`.

```
 -app, -deps, -deps-list, -roots, -profile …
                     │
                     ▼
        ┌──────── statyczny pipeline (prepare) ────────┐
        │ 1 wejście i komponenty                       │
        │ 2 bytecode i hierarchia klas                 │
        │ 3 profile i korzenie                         │
        │ 4 graf wywołań i modele                      │
        │ 5 silnik osiągalności                        │
        └──────────────────────────────────────────────┘
              │                          │
              ▼                          ▼
     etap CVE (ariadne)          bundle badawczy (research index)
     6 advisories (OSV)          → research query / check
     7 CVE → podatne metody      → research trace-config / correlate
     8 werdykty
     9 report.md, results.json
```

## 2. Statyczny pipeline

### Krok 1. Wejście i komponenty (`internal/input`, `internal/component`)

`-app` może być katalogiem klas, zwykłym jarem, fat jarem Spring Boot, WAR-em albo EAR-em. Archiwa, które niosą własne biblioteki (`BOOT-INF/lib`, `WEB-INF/lib`, moduły EAR), dostarczają zależności same. `-deps` to katalog jarów zależności. `-deps-list` (wynik `mvn dependency:list`) identyfikuje jary bez `pom.properties`.

Każdy jar staje się **instancją komponentu** z tożsamością Maven (group:artifact:version, GAV) i SHA-256. Komponenty osadzone w innym jarze (kopie shaded i relokowane) są wykrywane z metadanych i potwierdzane przez porównanie klas z opublikowanym jarem. `-embedded-components` i `-sbom` pozwalają takie komponenty zadeklarować; deklaracja jest sprawdzana na klasach, nigdy nie jest przyjmowana na ślepo. Advisories są później dopasowywane do instancji, a nie do nazw jarów.

### Krok 2. Bytecode i hierarchia klas (`internal/classfile`, `internal/hierarchy`)

ariadne parsuje każdy plik `.class` aplikacji i zależności własnym parserem. Dla każdej metody zapisuje instrukcje wywołań (invokevirtual, invokeinterface, invokestatic, invokespecial, invokedynamic), miejsca dostępu do pól i alokacji, adnotacje i numery linii. Hierarchia klas indeksuje nadklasy, interfejsy, podtypy i zadeklarowane metody.

Klasy tworzą jedną płaską ścieżkę klas: najpierw aplikacja, potem jary zależności w podanej kolejności (wygrywa pierwsza kopia klasy). Przegrana kopia zostaje zapisana jako **przesłonięta** (shadowed); jej kod nigdy nie trafia do ścieżki. Jary multi-release są rozstrzygane według `-target-java`.

### Krok 3. Profile i korzenie (`internal/profile`, `internal/libapi`)

**Profil** zawiera wiedzę o frameworku: które pakiety są kodem frameworka, które metody framework wywołuje (kontrolery, listenery, metody cyklu życia beanów) i jakie syntetyczne krawędzie framework dodaje. `-profile auto` (domyślnie) wykrywa obecne frameworki (spring, servlet, jakarta, micronaut, quarkus, stripes) i łączy ich profile; `generic` nie ma wiedzy o frameworkach.

**Korzenie** to miejsca, od których może zacząć się wykonanie (`-roots`):

| `-roots` | Korzenie |
|---|---|
| `all` (domyślnie) | każda metoda aplikacji oraz metody cyklu życia servletów, filtrów i listenerów zarejestrowanych w `web.xml` |
| `profile` | tylko punkty wejścia znane aktywnym profilom (handlery, listenery, beany) oraz te same metody z `web.xml` |
| `api` | **tryb biblioteki**: publiczne API jara z `-app`. Obejmuje też metody, które klasy API dziedziczą z zależności, oraz klasy, które biblioteka ogłasza frameworkowi (Spring, CDI, Micronaut); `-api-exclude` wyłącza pakiety |

W trybie biblioteki deklaracje zależności (jej `pom.xml` albo `-library-pom`) mówią, kto dostarcza wersję każdej zależności (`supplied_by`). Z `-alt-versions` ocenia się również inne wersje, które może dostać użytkownik biblioteki (sekcja 3).

### Krok 4. Graf wywołań i modele (`internal/callgraph` i pakiety modeli)

Każda metoda jest węzłem. Wywołania bezpośrednie stają się krawędziami. Wywołanie wirtualne lub interfejsowe daje krawędzie do wszystkich implementacji w hierarchii (**CHA**, domyślnie). `-callgraph-mode rta` zawęża je do klas, które osiągalny kod faktycznie tworzy; `pta` zawęża dalej tam, gdzie znane są alokacje odbiorcy.

Wywołania na typach JDK i `javax` nie są rozwijane przez CHA, bo JDK połączyłby wtedy wszystko ze wszystkim. Zamiast tego ograniczone **modele** dodają krawędzie, którymi JVM naprawdę przechodzi przez takie wywołania. Każdy model ma wersję, a większość jest sprawdzona wyrocznią na prawdziwym JVM:

- inicjalizacja klas (`<clinit>`), lambdy i referencje metod (wywołania SAM), MethodHandle,
- refleksja ze znanymi nazwami (literały, wartości wpływające do wywołania, konfiguracja),
- ServiceLoader i JPMS `provides`, dispatch sterowników i obiektów JDBC, dispatch kontenera servletów,
- proxy frameworków (repozytoria Spring Data, mappery MyBatis, proxy transakcyjne), advice Spring AOP, zdarzenia Springa, JPA,
- callbacki i wykonanie asynchroniczne (wątki, executory, CompletableFuture, listenery, reactive),
- punkty wejścia deserializacji i kandydaci łańcuchów gadgetów.

Modele zależne od siebie (jeden tworzy obiekty, na które drugi kieruje wywołania) liczą się razem, aż nic się nie zmienia (**punkt stały**). Model dodaje krawędź tylko wtedy, gdy kod tworzący potrzebny obiekt jest osiągalny.

To, czego żaden model nie rozstrzygnie, jest zapisywane jako **luka modelu** z miejscem i mechanizmem: nierozwiązana refleksja, metoda `native`, nieznany callback i tak dalej. Luka nigdy nie dodaje krawędzi. Wskazuje miejsce, w którym graf może być niepełny.

### Krok 5. Silnik osiągalności (`internal/reach`)

Dwa przeszukiwania wszerz biegną raz, ze wszystkich korzeni:

- **Przebieg 1 (`app_only`)** odwiedza metody frameworka, ale nie idzie przez nie dalej. Frameworki kierują wywołania przez refleksję i proxy, których graf statyczny nie umie wiernie śledzić.
- **Przebieg 2 (`full`)** nie ma tego ograniczenia.

Metoda znaleziona w przebiegu 1 jest osiągalna bez przechodzenia przez wnętrze frameworka. Metoda znaleziona tylko w przebiegu 2 jest osiągalna wyłącznie przez kod frameworka. Ścieżki są odtwarzane od najkrótszych, z preferencją kroków niezależnych od samego CHA. Każdy krok ścieżki niesie dowód: podstawę krawędzi, jej charakter (`jvm_rule`, `library_model`, `input_declaration`, `heuristic`) i miejsce wywołania.

Tu kończy się statyczny pipeline (`prepare` w `cmd/ariadne`). `ariadne` przechodzi dalej do etapu CVE; `ariadne research index` zamiast tego zapisuje stan do bundle'a.

## 3. Etap CVE

### Krok 6. Advisories (`internal/cves`)

Zidentyfikowane komponenty są wysyłane do OSV (`api.osv.dev`) w paczkach. Odpowiedzi trafiają do cache na dysku (`-cache-dir`, domyślnie `.cache`) razem z czasem pobrania:

- `-offline` korzysta wyłącznie z cache.
- `-data-ttl`, `-refresh`, `-allow-stale` i `-max-data-age` sterują świeżością danych.
- `-snapshot-out` zapisuje wszystko, co przebieg przeczytał, do samosprawdzającego się snapshotu. `-snapshot-in` odtwarza go bez sieci, a `-snapshot-compare` wyjaśnia różnice między przebiegami.

### Krok 7. Od advisory do podatnych metod (`internal/vulnmap`)

Advisory wskazuje pakiet i wersje, a nie kod. ariadne ustala podatne metody z tych źródeł, w kolejności:

1. wbudowana tabela ręczna (`internal/vulnmap/vuln_methods.yaml`), kuratorowana, z wariantami dla wersji,
2. metody zmienione przez commity z poprawką (GitHub; `GITHUB_TOKEN` podnosi limit zapytań),
3. dla advisory bez commitu z poprawką: metody, których bytecode zmienił się między ostatnim wydaniem przed poprawką a wydaniem z poprawką, w klasach wskazanych przez advisory, czytane z Maven Central (`-release-diff`, domyślnie włączone),
4. opcjonalnie, `-llm`: model językowy dla advisories, których pozostałe źródła nie zmapowały.

### Krok 8. Werdykty

Każde advisory dostaje werdykt dla każdej instancji komponentu, której dotyczy. W wynikach v1 jest jeden wiersz na advisory.

| Werdykt | Znaczenie |
|---|---|
| `REACHABLE` | Podatna metoda jest osiągana w przebiegu 1. Raport podaje najkrótszą ścieżkę i jej pierwsze wywołanie z kodu aplikacji do zależności. |
| `FRAMEWORK` | Podatna metoda jest osiągana tylko w przebiegu 2, przez kod frameworka. Wymaga przejrzenia. |
| `NOT_REACHABLE` | Zmapowane metody są w grafie, ale żaden korzeń do nich nie dochodzi. **Czytać jako „brak ścieżki w modelu”, nie „bezpieczne”.** Powody mówią, dlaczego werdykt może być słabszy: `MAPPING_PARTIAL`, `MODEL_INCOMPLETE` (osiągalne luki modelu mogą prowadzić do komponentu), `MAPPED_CODE_ABSENT`. |
| `UNKNOWN` | Nie ma mapowania albo żadna zmapowana metoda nie istnieje w analizowanej wersji (`METHOD_NOT_FOUND`), albo klasa instancji przegrała z inną kopią (`CLASSPATH_CONFLICT`). |
| `NOT_APPLICABLE` | Komponent nie zawiera kodu JVM (WebJar) albo advisory w ogóle nie dotyczy kodu JVM. |

Dodatkowe oceny są dołączane bez zmiany werdyktów:

- **advisories docelowego JDK** (`-jdk-home` / `-runtime-file`): tylko stosowalność,
- **inne wersje zależności** w trybie biblioteki (`-alt-versions`):
  - `approx`: advisories tej wersji oceniane na analizowanym kodzie,
  - `rerun`: cały statyczny pipeline liczony ponownie z jarem tej wersji,
- **sprawdzenie exploitu i frameworka** przez model językowy (`-exploit-check`, `-framework-check`) jako druga opinia,
- **dowody z działania** z nagranego przebiegu (`-runtime-trace`, sekcja 5).

### Krok 9. Wyniki (`internal/report`)

- `report.md`: raport do czytania; każde ustalenie ze ścieżką, powodami, lukami modelu i pokryciem.
- `results.json`: deterministyczne wyniki dla maszyn. `v1` (domyślnie) ma jeden wiersz na advisory. `v2` (`-json-format v2`) ma komponenty, jedno ustalenie na advisory × instancję, pokrycie, frameworki i sekcję runtime. Schemat jest w [RESULTS_CONTRACT.MD](RESULTS_CONTRACT.MD).
- Opcjonalnie: graf wywołań (`-graph-out`), porównanie z wcześniejszym grafem (`-graph-compare`), filtr JFR dla podatnych metod przebiegu (`-runtime-filter-out`) i snapshot.

Wszystkie wyniki zapisują się razem albo żaden.

## 4. Interfejs badawczy

`ariadne research index` uruchamia ten sam statyczny pipeline bez advisories i bez sieci. Zapisuje **bundle**: katalog z manifestem, tożsamością (wejścia z hashami, konfiguracja, ograniczenia pokrycia) i jednym zbiorem JSONL na każdy typ rekordu:

- komponenty,
- klasy (wybrane i przesłonięte),
- węzły grafu (ze specjalizacjami odbiorcy i lokalizacjami źródeł),
- krawędzie (z miejscami wywołań i warunkami),
- korzenie,
- luki modelu,
- wszystkie miejsca wywołań z bytecode.

`-scope reachable` zostawia tylko to, do czego dochodzą korzenie albo własne metody wejścia. `-sources` dodaje SHA-256 pliku źródłowego każdej metody.

`ariadne research query` odpowiada na jedno żądanie wyłącznie z bundle'a. Operacje:

- `entries`, `methods`, `calls`: jakie jest API i co wywołuje,
- `callers`, `callees`: jeden krok grafu,
- `reachability`: do 256 metod docelowych w żądaniu. Każdy cel jest rozstrzygany (`matched`, `missing`, `shadowed`, `ambiguous`), szukany w ramach budżetu pracy i zwracany z punktami wejścia, ścieżką-świadkiem dla każdego wejścia (każda krawędź jako dowód) i lukami modelu, które go dotyczą.

Kontraktem (identyfikatory, schematy, reguły, błędy) zarządza ariadne; jest opisany w [research-interface](research-interface/README.md). Dwa dodatkowe polecenia służą konsumentowi:

- `ariadne research check` wydaje werdykt producenta o dowolnej odpowiedzi. Pakiet akceptacyjny z prawdziwymi i celowo zepsutymi odpowiedziami opisuje [ACCEPTANCE.md](research-interface/ACCEPTANCE.md).
- `research trace-config` i `research correlate` (eksperymentalne) łączą nagrany przebieg z węzłami bundle'a; zob. [RUNTIME.md](research-interface/RUNTIME.md).

## 5. Dowody z działania

Graf statyczny może przybliżać z nadmiarem (CHA) i może pomijać to, czego nie obejmuje żaden model. Nagrany przebieg daje niezależny dowód:

1. `-runtime-filter-out` (skan CVE) albo `research trace-config` (badanie) zapisuje ustawienia JFR śledzące dokładne wejścia do wybranych metod (`jdk.MethodTrace`, JDK 25+).
2. Aplikacja działa pod Java Flight Recorder z tymi ustawieniami. `scripts/JfrToTrace.java` zamienia nagranie na format śladu [`ariadne-runtime-trace-1`](RUNTIME_TRACE.MD): załadowane klasy, wykonane metody i zmierzone relacje wywołujący → wywoływany.
3. `-runtime-trace` (skan CVE) albo `research correlate` przypisuje ślad do analizowanych klas **po hashu treści**. Klasa o innych bajtach nie dostaje niczego. Zmierzone wywołanie, dla którego graf nie ma krawędzi, jest klasyfikowane według przyczyny (refleksja, proxy, async, pominięty odbiorca, …).

Dowody z działania nigdy nie zmieniają krawędzi ani werdyktu. **Brak zdarzenia niczego nie dowodzi**, bo inny scenariusz może ten kod uruchomić. Wejście do metody dowodzi, że kod się wykonał, a nie że dotarły do niego dane atakującego.

## 6. Jak czytać wyniki

- `REACHABLE` i `FRAMEWORK` mają ścieżkę, którą można sprawdzić krok po kroku. Kroki oznaczone `heuristic` zależą wyłącznie od CHA i mogą być przybliżeniem z nadmiarem.
- `NOT_REACHABLE` obowiązuje w ramach modelu. Trzeba sprawdzić jego powody i luki modelu oraz pamiętać, że wejścia muszą odpowiadać wdrożonej ścieżce klas.
- W trybie biblioteki ustalenie znaczy, że **użytkownik API biblioteki** może dojść do podatnego kodu. `exposed_api` wymienia metody API, które do niego prowadzą, a `supplied_by` mówi, kto wybiera wersję zależności.
- Wszystko, czego ariadne wie, że nie widzi (luki, ograniczenia, nierozwiązany dispatch, granice natywne), jest wymienione w sekcjach pokrycia raportu, a nie ukryte.
