# Stan realizacji i przekazanie między sesjami

Ten plik jest rejestrem pracy nad nowymi planami, a nie deklaracją wsparcia produktu. Aktualne możliwości opisuje [SUPPORT.MD](../docs/SUPPORT.MD), zakres prac [ROADMAP.MD](ROADMAP.MD), a sposób prowadzenia sesji [skill](skills/ariadne-implement/SKILL.md).

## Interfejs badawczy RI0 — dokumentacja kontraktu (2026-11-24)

- Stan IN_PROGRESS dla RI0. Przygotowano specyfikację `1.0.0-draft.1`, JSON Schema, hash-weryfikowalny bundle danych kontrolnych, przykłady żądań/odpowiedzi i walidator dokumentacji w [docs/research-interface](../docs/research-interface/README.md).
- Kontrakt obejmuje katalogi API/metod/wywołań, sąsiedztwo grafu, paczkową osiągalność, tożsamości, dowody ścieżek, warunki krawędzi, luki, limity, błędy i wersjonowanie. Tekmerion otrzymuje kopię pakietu z pinem hashy i instrukcją konsumenta.
- Weryfikacja dotyczy schematów, integralności danych, odwołań i semantyki przypadków kontrolnych. Implementacja CLI RI1–RI6 i akceptacja wykonującego klienta TM0 pozostają planowane.
- Następny przyrost obejmuje akceptację kontraktu w adapterze TM0 i wydzielenie pipeline'u RI1, zgodnie z [planem](RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD).

## Stan początkowy

Przygotowano plany, prompt i skill. W ramach ich tworzenia nie rozpoczęto implementacji nowych faz. Istniejący kod może spełniać część kryteriów; każda kolejna sesja ma to zweryfikować zamiast zakładać, że wszystko trzeba napisać od początku.

Stany faz: `UNASSESSED`, `READY`, `IN_PROGRESS`, `NEEDS_VERIFICATION`, `DONE`, `BLOCKED`. `DONE` oznacza spełnienie wszystkich kryteriów konkretnej fazy oraz wymaganej weryfikacji i dokumentacji. `BLOCKED` wymaga konkretnego powodu i sposobu odblokowania. To etykiety tego pliku, nie stany celów w aplikacji Codex.

## Przegląd obszarów

| Obszar | Plan | Ocena nowych faz | Wznowienie |
|---|---|---|---|
| ROADMAP | [Kontrakty B0–B4 i przyrosty](ROADMAP.MD) | **B0–B4 DONE (2026-11-23)**: B0 DONE for this scope (`b0-1`…`b0-5`, decyzja o ścieżkach przykładowych 2026-11-20); B1 DONE for this scope, B2 DONE (oceny 2026-11-23); B3 DONE for this scope (`b3-1`…`b3-4`; iteracja przebiegów jednorazowych i unieważnianie per model świadomie niebudowane); B4 DONE (każdy obszar ma wyrocznię JVM albo „nie dotyczy”). Historia przyrostów — wpisy ROADMAP:B* w rejestrze | Najnowszy wpis ROADMAP:B3 w rejestrze. |
| CVE | [F0–F10, zarchiwizowane](../archive/2026-11-23/cve/COMPLETION.md) | F0: DONE (kontrakt, 2026-11-05); F1: DONE (F1a–F1c, 2026-11-05); F2, F3, F4: DONE (2026-11-05); F5: DONE (F5a–F5c, 2026-11-06); F6: DONE (F6a, F6b 2026-11-06); F7: DONE dla jawnej relokacji (F7a, 2026-11-06); F8: F8a DONE (2026-11-13, identyfikacja runtime'u); F8b DONE dla mechanizmu (2026-11-13: dostawca upstream OpenJDK przez kuratorowany katalog — warunki openjdk.org wykluczają automatyczny dostęp; pierwsza kuracja 2026-11-20: 8 advisories z 2026, wersja katalogu 1), F8c: zamknięte bez budowy (2026-11-23, decyzja użytkownika: werdykt metodowy i sygnał `api_exposure` niebudowane); F8 DONE, [zarchiwizowane](../archive/2026-11-23/cve/CVE_F8.md); F9: DONE dla F0–F7 (F9a, F9b 2026-11-06); F10: DONE (2026-11-10, kontynuacje 2026-11-13 i 2026-11-16; PR #48, #49, [zarchiwizowane](../archive/2026-11-13/cve/CVE_F10.md); redukcja `UNKNOWN` PR #84–#87, [zarchiwizowane](../archive/2026-11-16/cve/CVE_F10_UNKNOWN_REDUCTION.md)) | Kontrakt wyników `docs/RESULTS_CONTRACT.MD` (osie, stany z v2 `not_assessed`, kody powodów `internal/findings`, reguła B0 dla luk, zgodność JSON, agregacja v1), audyt 39 mapowań ręcznych (38 punktów wejścia, 1 ujście, 0 pełnych), zestaw kontrolny 33 przypadków pokrywający 5 rozszerzeń i całą macierz §5, odtwarzalny baseline z manifestem odpowiedzi. Szacunek: 7/26 i 6/18 dzisiejszych NOT_REACHABLE to "metoda nie znaleziona". F1b: kody powodów w wynikach, ustalenia per advisory × instancja z agregacją, reguła luk w formie report-only (zmierzona: obejmuje wszystkie NOT_REACHABLE obu aplikacji, więc nie zmienia werdyktu). F1c: `-json-format v2`, v1 jako eksport ustaleń sprawdzany w każdym przebiegu. F2: metadane cache, TTL danych OSV, `-refresh`/`-allow-stale`/`-max-data-age`, wycofane rekordy i aliasy, odświeżenie bez płatnych zapytań LLM. F3: snapshoty danych (`-snapshot-out`/`-snapshot-in`/`-snapshot-compare`), klucze wyprowadzeń. F4: komparator wersji Maven (wyrocznia Maven 3.9.12), wersjonowane mapowania, selektory z deskryptorem. F5a: obie strony diffu (rodzic z GitHub REST API), mapowanie = metody zmienione + usunięte sprzed poprawki, dodane helpery tylko w dowodach, `MAPPING_NEW_SIDE_ONLY`/`SOURCE_UNAVAILABLE`, błąd źródła per CVE. F5b: klasy członkowskie `Outer$Inner`, weryfikacja kandydatów w bytecode (deskryptor tylko przy jednoznacznym przeciążeniu), dopasowanie commita do wydania po obecności kandydatów, `possible_rename`. F5c: linia wydań z historii git (tag badanej wersji + GitHub compare: `descends_from_release`/`in_release`/`other_line`, tag poza gałęzią), wykluczenie commitów z innej linii, prompt LLM v2 z obiema stronami poprawki. F6a: komponenty osadzone z `pom.properties` (moduł projektu vs biblioteka), potwierdzenie klasami względem JAR-a z Maven Central (bajty albo odcisk niezależny od serializacji), osobna instancja z własnymi klasami i zapytaniem OSV, niepotwierdzeni kandydaci jako luka pokrycia. F7a: jawna relokacja z `pom.xml` (maven-shade) i z `-embedded-components` (YAML, SHA-256), porównanie przez odwrotną relokację, przepisywanie selektorów w silniku, kolizje/konflikty/minimalizacja bez potwierdzenia. F10 (2026-11-10) i ręczne mapowania z diffów wydań (2026-11-13, 46 → 68 wpisów; `UNKNOWN` 49/45/8 → 27/23/5); redukcja `UNKNOWN` (2026-11-16, 68 → 107 wpisów, kody `NOT_CODE_MAPPABLE` i `MAPPED_CODE_ABSENT`; `UNKNOWN` 27/26/5 → 1/0/0). **Obszar CVE DONE (2026-11-23), zarchiwizowany** — F9 dla F8 zrobione; zob. [COMPLETION.md](../archive/2026-11-23/cve/COMPLETION.md). |
| JVM_GRAPH | [G0–G5, zarchiwizowane](../archive/2026-10-22/jvm_graph/COMPLETION.md) | G0, G1, G3, G4, G5: DONE; G2: DONE (overloads/inheritance; bridge/default-method real-bytecode verification closed 2026-11-07 by GRAPH_PRECISION P0 — `testdata/callgraph`, JVM oracle; korekta 2026-11-08, PR #29: CHA celuje też w ciała dziedziczone spoza typu wywołania, builder `3`) | Cały JVM_GRAPH's R0-owy zakres (G0–G5) zamknięty; ograniczenie G2 domknięte i korekta CHA (addendum w COMPLETION.md). Następny krok to inny obszar R0/R1 z ROADMAP.MD, nie JVM_GRAPH. |
| DATA_FLOW | [D0–D5, zarchiwizowane](../archive/2026-11-07/data_flow/COMPLETION.md) | D0–D5: DONE (dla zadeklarowanych zakresów); obszar zamknięty 2026-11-07 (D5 — PR #13) | IR/CFG/wartości lokalne (D0–D2), pięć wycinków międzyproceduralnych wartości i streszczenia zależności `internal/flowsum` (D3), pilotaż taint `-taint` opt-in (D4), kwalifikacja: wyrocznia JVM, kształty, pomiar (D5, `docs/TAINT_QUALIFICATION.MD`). Ograniczenia: zob. COMPLETION.md i SUPPORT.MD (wiersz Partial). |
| SERVICE_LOADING | [S0–S4, zarchiwizowane](../archive/2026-11-05/service_loading/COMPLETION.md) | S0–S4: DONE (dla zadeklarowanych zakresów); obszar zamknięty 2026-11-05; po zamknięciu: dispatch SPI na pomijanych odbiorcach (`service-provider:dispatch`, 2026-11-08) | Load sites, lazy materialization, dispatch, JPMS (`-jpms-scenario`), rozpoznane loadery, D3 parametry/pola, `doPrivileged`; S4: wyrocznia prawdziwego JVM (62 pary, 0 nadmiarowych krawędzi, 1 zadeklarowana luka), pomiar historyczny (`results.json` bajtowo stały na 11 buildach). Ograniczenia: zob. COMPLETION.md i SUPPORT.MD (wiersz Partial). |
| REFLECTION | [R0–R4, zarchiwizowane](../archive/2026-11-23/reflection/COMPLETION.md) | R0+R1+R2: DONE dla całego zadeklarowanego zakresu ("R0–R2 stanowią pierwsze wydanie"); R3: zmierzone dwukrotnie (2026-11-05, 2026-11-09) z zerowym zyskiem; wycinek 3 (`r3-3`, 2026-11-15) DONE for this scope — spis pochodzenia znalazł 2 miejsca (Thymeleaf, `Method` w parametrze z pola statycznego), rozwiązywane przez wywołujących; reszta R3 (konfiguracja, `setAccessible`, `Constructor` przez parametr) zamknięta bez budowy 2026-11-23 (zerowy zysk w dwóch pomiarach; R3 DONE for this scope); R4: oceniony 2026-11-21 — kryteria w większości spełnione przez B0/B3/B4; `r4-1` DONE (2026-11-21) — `Class.forName` → `<clinit>`, wyniki v1 bez zmian w 12 konfiguracjach; `r4-2` DONE (2026-11-22) — przypadki fixtures (samo wyszukanie, lookupy odrzucane przez JVM jako luka, brak `setAccessible` jako jawna nadaproksymacja); **R4 DONE**; **obszar REFLECTION DONE (2026-11-23), zarchiwizowany** | `Class.forName("literał")` rozwiązuje się do `KindConst`/`java.lang.Class` (reużycie, nie nowy typ). `getMethod`/`getDeclaredMethod` z literalną nazwą i literalną tablicą `Class[]` rozwiązuje się do nowego `KindReflectiveMethod` — wymagało nowej maszynerii: `KindArray` (wskaźnikowy, mutowalny przez `aastore`, prawdziwa semantyka referencji JVM) i minimalnego śledzenia literałów `int` (wcześniej całkowicie poza zakresem D2). Zmierzony efekt R0+R1 na petclinic/WebGoat: **zero** (żadna z aplikacji nie używa tych API), uczciwie odnotowane; brak crasha na pełnej hierarchii (29k/40k klas), bajt-identyczny `results.json`. R2's pierwsza połowa: `internal/reflectinvoke` łączy każde miejsce `Method#invoke`, którego receiver rozwiązuje się do powyższego faktu, realną krawędzią grafu (`BasisReflection`) do realnego celu dyspozycji (typ alokacji, nie tylko `Class` z lookupu), z uczciwym `Gap` przy niezgodności liczby argumentów/nierozpoznanym typie parametru/braku deklaracji — **efekt niezerowy**, inaczej niż R0/R1: petclinic 3/59/29/51/6→3/62/26/51/6 (5 krawędzi), WebGoat 40/43/21/56/0→42/44/18/56/0 (3 krawędzie). R2's druga połowa (nowa): analogiczna maszyneria dla `getConstructor`/`getDeclaredConstructor` (`KindReflectiveConstructor`, reużywa R1's `KindArray`, bez nowego modelu tablicy) i `Constructor#newInstance` → realna krawędź do **dokładnie jednego** `<init>` (konstruktor nigdy nie jest dziedziczony — JVMS 5.4.3.3, więc żadnego overload-insensitive doboru jak przy `Method#invoke`), plus powiązana krawędź `<clinit>` tą samą parą Kind/Basis, jaką `ApplyServiceLoaderEdges` już stosuje dla SERVICE_LOADING S2. Zmierzone: realne krawędzie istnieją (petclinic 264 miejsc/6 rozwiązań/7 krawędzi, WebGoat 264/5/5), ale **zero dodatkowego efektu na reachability** poza tym, co dała już pierwsza połowa — liczby REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN identyczne przed/po na obu aplikacjach, uczciwie odnotowane (nowo-konstruowalne klasy nie leżą akurat na ścieżkach podatnych metod w tych dwóch aplikacjach). Dwa przebiegi offline na każdej dają bajt-identyczny `results.json`. |
| FRAMEWORKS | [F0–F5, zarchiwizowane](../archive/2026-11-23/frameworks/COMPLETION.md) | **Obszar DONE (2026-11-23), zarchiwizowany**; F0, F1: DONE; F2: DONE for this scope (zamknięte 2026-11-23) (F1 dla swojego faktycznego zakresu); F2: sześć wycinków (DI, oznaczone korzenie web.xml, `@Bean` factory methods, katalog zewnętrznych punktów wejścia, Spring Data repository beans, `@Value`-owe właściwości) + korekta false positives zrobione; F2 (siódmy wycinek, 2026-11-08): dispatch kontenera servletów — model B3 `servlet-container:dispatch` do zarejestrowanych servletów/filtrów; F2 (`s2`, 2026-11-09): wywołania na obiektach żądania/odpowiedzi/łańcucha do implementacji tworzonych przez osiągalny kod (WebGoat: 15 Undertow `FRAMEWORK` → `REACHABLE`, 4 wiarygodne, 11 artefaktów CHA); F2 (`l2`/`l2-2`, 2026-11-10): appendery i komponenty Joran logback spoza konfiguracji (też domyślnej Spring Boot) + builder `6` — petclinic/WebGoat: CVE-2024-12798 `FRAMEWORK` → `NOT_REACHABLE`, CVE-2021-42550 zostaje (selektor JNDI z właściwości JVM); F2 (`l1`, 2026-11-10): appendery log4j 1.2 spoza pakowanej konfiguracji nie są celami CHA/RTA (JPetStore: 3 fałszywe `REACHABLE` usunięte, aplikacja kontrolna zgodna z JVM); F2 (`j1`, 2026-11-09): wywołania na obiektach JDBC (`java.sql`/`javax.sql.DataSource`) do implementacji tworzonych przez osiągalny kod (ślad petclinic: wszystkie 22 relacje JDBC w grafie, bez zmian werdyktów); F3: pierwszy wycinek (tożsamość proxy/target/interceptor, `internal/aop`) zrobiony; wycinek 3 (2026-11-10, `framework:proxies` `x1`): katalog proxy MyBatis/Spring Data/Spring ORM — ślady: petclinic `proxy` 13 → 0, JPetStore 8 → 2, JPetStore CVE-2022-41853 `REACHABLE`; wycinek 2 (dispatch przez proxy JDK) zmierzony 2026-11-09 i NOT STARTED — 0 z 12 relacji `proxy` ma ograniczony zbiór interfejsów; advice Spring AOP (`spring:aop` `a1`, `a2`, 2026-11-16) DONE for this scope — wywołanie metody beana objętej pointcutem → metoda advice `@AspectJ` lub `<aop:config>`, wiązanie zmiennych jak w Springu, `args`/`this`/`target`/`@target`/`bean`; F3 i F4: DONE for this scope (2026-11-23, decyzja użytkownika; [F3](../archive/2026-11-23/frameworks/FRAMEWORKS_F3.md), [F4](../archive/2026-11-23/frameworks/FRAMEWORKS_F4.md)); F4: wycinek 1 (`spring:events` `ev1`, 2026-11-15) DONE for this scope — `publishEvent` → listenery `@EventListener`/`@TransactionalEventListener` beanów, `@Schedules` jako korzeń; F5: DONE for this scope (2026-11-23; `f5-1` wersja reguł korzeni w snapshotcie, `f5-2` wyrocznia Spring MVC — 14/26 ciał naprawione, `f5-3` dwa scenariusze profili i wstrzyknięcia — 2 fałszywe negatywy i 2 błędne fakty naprawione; 8 z 11 przypadków fixtures z wyrocznią JVM, odłożone: konsumenci wiadomości, priorytety konfiguracji; [zarchiwizowane](../archive/2026-11-23/frameworks/FRAMEWORKS_F5.md)) | Dispatch kontenera (2026-11-08): `Servlet#service`/`Filter#doFilter`/szablon `HttpServlet` → servlety/filtry z web.xml, `@WebServlet`/`@WebFilter`, beanów i katalogu autokonfiguracji Boot; petclinic: `DispatcherServlet` osiągalny, fałszywie negatywne CVE-2024-38809 przy RTA usunięte; WebGoat bez rejestracji (launcher poza wejściem). F0 oddziela wykrycie frameworka od aktywnego profilu. F1 dostarcza realne wartości adnotacji (Spring MVC routes/bean names/`@Profile`), ograniczony `web.xml`/XML bean definitions odczyt, i wykrywanie konfliktów nazw beanów. F2's pierwszy wycinek: `@Autowired` injection points rozwiązane po typie/qualifier/primary wobec F1's `BeanNames` — zmierzone pozytywnie na WebGoat (7/74), poprawnie 0/8 na petclinic (Spring Data repository, inny mechanizm). F2's drugi wycinek: `web.xml`-owe servlety/filtry/listenery dostają realne korzenie pod `-roots profile` — zweryfikowane tylko na testach jednostkowych (petclinic/WebGoat nie mają `web.xml`). F2's trzeci wycinek: `@Bean` factory methods na `@Configuration` klasach dodane do tego samego katalogu beanów — WebGoat skacze z 7/74 do 40/74 rozwiązanych (21 `@Bean` metod znalezionych), petclinic bez zmiany (0/8, niezależne typy). F2's czwarty wycinek: `ExternalEntries` scala `Routes`+`WebXMLRoots` w jeden katalog z `Basis`/`Active` (F0's `ExternalEntry` kontrakt, teraz realny) — czysto raportowy, zero nowych korzeni/krawędzi; petclinic 17, WebGoat 19 wpisów. F2's piąty wycinek: `RepositoryBeans` (interfejsy aplikacji — podtypy `Repository`, ten sam sprawdzian co `synthetic_edges`) — petclinic skacze z 0/8 do **8/8**. **Korekta (nie wycinek):** liczby WebGoat (74/7/40/50) były zawyżone fałszywymi trafieniami niejawnej reguły "jedyny konstruktor" Springa 4.3 (Lombok, enumy, ręczne `new`) — po poprawce w `InjectionPoints` prawdziwe liczby to **43 punkty, 37 rozwiązanych, 6 w pełni wyjaśnionych** (1 `@Value`, 4 `List<Bean>`, 1 framework-provided bean). F2's szósty wycinek: `@Value`-owe punkty wstrzyknięcia (pole lub parametr konstruktora z `@Value(...)`) rozpoznane jako **własna, odrębna kategoria** — nigdy nie przechodzą przez wyszukiwanie beanów (Spring rozwiązuje je przez property sources, nie katalog beanów), więc `Reason` mówi teraz "a @Value config value (...), not a bean lookup", nie mylące "no known bean". Ujawniło też, że pole-samo-`@Value` (bez `@Autowired`) **nigdy wcześniej nie było skanowane** — WebGoat skacze z 43 do **45** punktów (2 nowo znalezione pola `LessonMenuService#excludeCategories`/`excludeLessons`), z czego **8** nierozwiązanych teraz w pełni wyjaśnionych: **3** realne `@Value` (było 1 — pozostałe 2 to właśnie te nowo znalezione pola, które poprzednia sesja błędnie zaliczyła do "4 `List<Bean>`"), **4** realne `List<Bean>` (parametry konstruktora, odrębne od pól), **1** framework-provided bean. petclinic bez zmian (8/8, brak `@Value` w kodzie). Zero wpływu na reachability na obu aplikacjach (`@Value` to fakt konfiguracyjny, nigdy krawędź grafu). Rejestr B3 — nietknięty. |
| ASYNC_CALLBACKS | [A0–A4, zarchiwizowane](../archive/2026-11-20/async_callbacks/COMPLETION.md) | A0–A2 (MVP): DONE; A3: wycinek 1 (2026-11-15) DONE for this scope — katalog kontynuacji `CompletionStage`, warunek `exceptional` w dowodzie; wycinek 2 (`a3-2`, 2026-11-19) — zadania przekazywane przez parametry rozwiązywane u wywołujących; wycinek 3 (`a3-3`, 2026-11-19) — przekazania przez kolejkę w polu i załącznik klucza NIO; wycinek 4 (`a3-4`, 2026-11-20) — zadania zwracane z metody (bramka `async-callbacks:returned-tasks`); wycinek 5 (`a3-5`, 2026-11-20) — listenery: lista w polu i `PropertyChangeSupport` (bramka `async-callbacks:listeners`); wycinek 6 (`a3-6`, 2026-11-20) — reactive: JDK `Flow`/`SubmissionPublisher` (bramka `async-callbacks:reactive`); wycinek 7 (`a3-7`, 2026-11-20) — zakończenie ręcznie kończonych `CompletableFuture` (bramka `async-callbacks:completion`); A3 DONE for this scope (2026-11-20); A4: `a4-1` (2026-11-20) — podklasy `Thread` i odrzucone wykonanie; A4: `a4-2` (2026-11-20) — kwalifikacja: pełny pipeline na 9 fixtures vs wyrocznie JVM; **A0–A4 DONE for this scope**; obszar zamknięty i zarchiwizowany 2026-11-20 (PR #109–#116) | A2 dodaje realną krawędź "może wykonać" do `run`/`call` albo (po DYNAMIC_LINKAGE L0/L1) prosto do implementacji lambdy — 67/44 z 69/46 z 135/134 rozwiązanych na petclinic/WebGoat, z 2 szczerymi lukami dla klas JDK poza analizowanym uniwersum. A0–A2 domykają zadeklarowane MVP tego obszaru; A3 (2026-11-15…20): kontynuacje, przekazania (parametry, kolejki, klucze NIO, wartości zwracane), listenery, JDK `Flow`, zakończenie ręcznie kończonych future — każdy wycinek z wyrocznią JVM; werdykty na bibliotekach i aplikacjach referencyjnych bez zmian. |
| DYNAMIC_LINKAGE | [L0–L4, zarchiwizowane](../archive/2026-11-23/dynamic_linkage/COMPLETION.md) | L0, L1 (creation site; SAM-call materialization 2026-11-09; PR #30, [L1 zarchiwizowane](../archive/2026-11-09/dynamic_linkage/DYNAMIC_LINKAGE_L1.md)), L2 (same-method + static-field case): DONE; L2: DONE for this scope (2026-11-23, dalsza rozbudowa zamknięta po pomiarze 2026-11-21, [zarchiwizowane](../archive/2026-11-23/dynamic_linkage/DYNAMIC_LINKAGE_L2.md)); L3–L4: DONE for this scope (2026-11-23; L3 niebudowane po pomiarze); **obszar DONE, zarchiwizowany** | L1: lambda/method reference uruchamiana przez interfejs funkcyjny w innej metodzie niż tworząca dostaje krawędź `sam` (`internal/samcall`, builder `4`); na śladzie petclinic luki `sam-call` 50 → 22 (12 z nich to nie lambdy). L2 wciąż daje **0 z 10/46** realnych krawędzi na petclinic/WebGoat (pola instancyjne / wynik innego wywołania). Pomiar 2026-11-21: 0/10 i 2/46 miejsc `invoke` rozwiązanych, 2 z 611 istotnych miejsc luk petclinic — dalsze L2 nie podjęte (decyzja). |
| DESERIALIZATION | [Z0–Z4, zarchiwizowane](../archive/2026-11-07/deserialization/COMPLETION.md) | Z0–Z4: DONE (dla zadeklarowanych zakresów); obszar zamknięty 2026-11-07 (Z3 — PR #11, `e818150`) | Modele wejść (Java serialization, XStream), krawędzie „może wykonać” do callbacków, polityki literalne (Z2), katalog 5 łańcuchów `z3-1` per wejście i per kopia klas, dopływ danych żądania per wejście × łańcuch (`-taint`), v2 i powiązania przy ustaleniach, wyrocznia JVM i zestaw CVE. Ograniczenia: zob. COMPLETION.md i SUPPORT.MD (wiersz Partial). |
| GRAPH_PRECISION | [P0–P4, zarchiwizowane](../archive/2026-11-23/graph_precision/COMPLETION.md) | P0: DONE (2026-11-07; PR #15, [zarchiwizowane](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md)); P1: DONE (2026-11-08; PR #17, [zarchiwizowane](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md)); P2: DONE (2026-11-08; PR #19, [zarchiwizowane](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md)); P4: DONE (2026-11-08; PR #21, [zarchiwizowane](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P4.md)); P3: DONE for this scope (zamknięte 2026-11-23; wycinek `p3-1` — eksperymentalny tryb `-callgraph-mode pta`, 1–3,4 % krawędzi CHA ponad RTA, 0 zmian werdyktów, 2,7–3,7× czasu RTA; **obszar DONE, zarchiwizowany**); trzecia aplikacja pomiarowa JPetStore (WAR, 2026-11-09): żaden model nie zmienia werdyktu, 3 fałszywe negatywy domyślnego `-roots all` (korzenie web.xml, naprawione); 2026-11-10: ręczne mapowania HSQLDB/log4j 1.2 (7 sinków) — 3 artefakty CHA przez appendery log4j, RTA poprawne, `j1` bez wpływu; WebGoat: CVE-2022-41853 `REACHABLE` (lekcja SQLi wykonuje SQL z żądania) | Ponowny benchmark po modelu dispatchu servletów (2026-11-08): petclinic — CVE-2024-38809 i CVE-2024-22259 FRAMEWORK przez ścieżkę żądania, zostają 4 poprawne zmiany RTA; WebGoat bez zmian (launcher Boot poza wejściem); domyślnie nadal CHA. P4: benchmark CHA vs RTA (`scripts/graph-benchmark.sh`, `testdata/graphbench`) — RTA <1 s, `-roots all` bez zmian werdyktów, przy `-roots profile` 5 z 8 zmian poprawnych, 1 nierozstrzygnięta, 2 fałszywie negatywne przez niemodelowany dispatch kontener → servlet (`javax.*` skip); domyślny tryb pozostaje CHA. P2: opt-in `-callgraph-mode rta` — punkt stały na katalogu P1 zawęża krawędzie CHA do klas tworzonych przez osiągalny kod, fallback nieznanych alokacji z jawnymi lukami, wyrocznia JVM `testdata/rta`; petclinic/WebGoat `-roots profile`: −17% krawędzi CHA, 6/2 FRAMEWORK → NOT_REACHABLE (wszystkie z `MODEL_INCOMPLETE`), żaden REACHABLE nie stracony; domyślny tryb bez zmian. Następny: P3 (ograniczone points-to) albo P4 (benchmark). Szczegóły we wpisie GRAPH_PRECISION:P2. |
| RUNTIME_NATIVE | [N0–N5](RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD) | N0, N1: DONE (2026-11-08; PR #24, [zarchiwizowane](../archive/2026-11-08/runtime_native/RUNTIME_NATIVE_N0_N1.md)); N2: oceniony 2026-11-23 — NOT STARTED, odłożony (brak celu z wieloma kontekstami ładowania; kontrakt śladu już niesie loadery); N3–N4: UNASSESSED; N5: wycinek `n5-1` DONE (2026-11-23) — granice natywne jako luki `native` i inwentarz ładowań bibliotek | Przyczyny luk grafu względem śladu (`g1`, PR #28, addendum w archiwum N0–N1): na petclinic wskazały defekt CHA (naprawiony, PR #29) i wywołania SAM (L1, PR #30). `-runtime-filter-out`: ustawienia JFR śledzące wszystkie podatne metody ustaleń — petclinic: 505 metod / 91 advisories, 26 wykonanych (3/3 REACHABLE, 23/74 FRAMEWORK, CVE-2025-35036 NOT_REACHABLE), 65 śledzonych bez wejścia. Kontrakt śladu `ariadne-runtime-trace-1` i importer `-runtime-trace` (dopasowanie po hashach klas, definicje per loader, wywołania vs graf statyczny, podatne metody zaobserwowane per ustalenie); rejestrator z JFR bez agenta. petclinic: CVE-2024-38809 śledzone ~1 400 razy, CVE-2025-35036 wykonane mimo NOT_REACHABLE (luka `javax.validation` SPI). Graf i werdykty bez zmian. Szczegóły we wpisie RUNTIME_NATIVE:N0–N1. |
| ECOSYSTEMS | [E0–E4](ECOSYSTEMS_IMPLEMENTATION_PLAN.MD) | UNASSESSED | Osobny pilot; nie jest warunkiem ulepszania JVM. |
| LIBRARY_MODE | [LB0–LB4, zarchiwizowane](../archive/2026-11-23/library_mode/COMPLETION.md) | LB0, LB1: DONE (2026-11-16; kontrakt §8, `testdata/library`, `-roots api`, `exposed_api`); LB2: wycinek 1 (`lb2-1`/`lb2-2`, 2026-11-17) DONE for this scope — luka `client_code`, `exposed_api[].client_dispatch`; pomiar na 3 prawdziwych bibliotekach (2026-11-17, `scripts/library-measure.sh`); metody odziedziczone z zależności jako API (`lb2-3`), specjalizowane do odbiorcy (`lb2-4`); ogłoszenia Spring (`spring.factories`, `*.imports`) jako korzenie (`lb2-5`); archiwa beanów CDI jako korzenie (`lb2-6`, 2026-11-18); obiekty dostarczane przez wywołującego jako luka (`lb2-7`, 2026-11-19); biblioteki ogłaszające klasy Springowi przez framework (`lb2-8`, 2026-11-19); definicje beanów Micronauta jako korzenie (`lb2-9`, 2026-11-20); LB3: wycinek 1 (`lb3-1`, 2026-11-18) DONE for this scope — deklaracje zależności z pomu biblioteki, `supplied_by` (wyrocznia Maven); LB4: wycinek 1 (`lb4-1`, 2026-11-19) DONE for this scope — zestaw kwalifikacyjny 5 bibliotek (fasada, fabryka, SPI, starter, callbacki), wyrocznia JVM jetty-client, użycie w CI; `lb4-2` (2026-11-19): wyrocznie JVM pozostałych trzech bibliotek, 0 fałszywych negatywów; LB2 i LB4: DONE for this scope (2026-11-23, decyzja użytkownika; [LB2](../archive/2026-11-23/library_mode/LIBRARY_MODE_LB2.md), [LB4](../archive/2026-11-23/library_mode/LIBRARY_MODE_LB4.md)); LB3: DONE for this scope (`lb3-1`, `lb3-2` 2026-11-23: opcja A + wybór X pod flagami `-alt-versions`/`-alt-versions-mode`; na przyszłość: tryb `rerun` (opcja B), inne strategie wyboru); **obszar DONE, zarchiwizowany** | Analiza jara biblioteki: korzenie z publicznego API zamiast wszystkich metod, lista narażonych metod API per advisory, kod klienta jako jawna luka, wersje zależności. Dziś przybliżenie: `-app` biblioteki z `-roots all` (także metody prywatne). |

Po pierwszej sesji wpisz poniżej konkretne identyfikatory faz i ich stan. Ogólne `UNASSESSED` w tabeli obszarów zastąp krótkim podsumowaniem rzeczywistego postępu; nie oznaczaj całego obszaru DONE po jednej fazie.

## Rejestr faz

### Raport: klasy w kilku kopiach (sekcja report.md) — DONE (2026-11-23)

- **Zakres:** dane o przesłoniętych klasach były tylko w v2 (`components[].shadowed`, `shadowed_differing`); report.md dostaje sekcję „Classes in several copies”: liczba klas zdefiniowanych przez więcej niż jeden komponent, ile z inną treścią, kolejność classpath (`ClasspathOrder`) i tabela par komponentów (kopia wzięta, kopia przesłonięta, klasy, z inną treścią, przykład), pary z różną treścią najpierw, ≤ 30 wierszy (`shadowReportRows`, `cmd/ariadne/gaps.go`).
- **Test:** `TestShadowedClassesReport` — commons-lang3 3.11 trzy razy (jar, identyczna kopia, jar z `Validate` o zmienionej stałej): dwa wiersze, najpierw para z różną treścią. `go test ./...`, `go vet ./...`.
- **Zmiana tylko w report.md:** v1 i v2 bez zmian. Pomiar na aplikacjach referencyjnych nie został dokończony w tej sesji (limit użycia).

### RUNTIME_NATIVE:N5 — wycinek `n5-1`: granice natywne (metody `native`, ładowanie bibliotek) — DONE for this scope (2026-11-23)

- **Zakres (zatwierdzony 2026-11-23: „N5 minimalne”):** jawna inwentaryzacja granic Java ↔ natywne, bez modelu kodu natywnego (plan N5: „pierwszy tor JNI może zakończyć się na jawnej inwentaryzacji granic”).
- **Zrobione:** `classfile.Method.IsNative`; `internal/nativebound` (`n5-1`): `Methods` — metody `native` z klas wejścia; `Loads` — wywołania `System`/`Runtime` `loadLibrary`/`load(String)` z nazwą biblioteki, gdy jest literałem (fakty `dataflow`); `Summarize` z osiągalnością z silnika (`MethodReachable`). Każda metoda `native` to miejsce luki rodzaju `native` (`nativeGapSites`), nieograniczone — reguła luk B0 (§4) rozpatruje osiągalne. Słownik luk: nowy mechanizm `native` (site/`native`, wersja słownika `n5-1`). Wynik: v2 `coverage.native`, report.md „Native code boundaries” (≤ 50 metod, ładowania), log; snapshot `algorithms` `native.boundaries`.
- **Testy:** `TestNativeBoundaries` (fixture `testdata/native`, `scripts/native-fixture.sh`: metoda natywna wołana z `main` — osiągalna i luka; nieużywana — nie; `loadLibrary` literału w `<clinit>` i `Runtime#loadLibrary` parametru), `TestEveryProducerCodeMapped`, `TestVocabularyDocumented`; `go test ./...`, `go vet ./...`.
- **Pomiar (`main` vs gałąź, offline, `-profile spring`, `-roots profile`/`all`, CHA):** petclinic 361 metod natywnych (tomcat-embed-core: APR/OpenSSL), **72 osiągalne**, 7 ładowań (5 osiągalnych); WebGoat 206 (jffi 204, wildfly-common 3), **108 osiągalnych**, 3 ładowania (0); JPetStore 3, 0 osiągalnych — wyniki v1 bajtowo identyczne. **0 zmian werdyktów i kodów powodów** w 6 konfiguracjach; `model_gaps` dostaje `native` w 26 (petclinic) i 21 (WebGoat) ustaleniach, które już były `MODEL_INCOMPLETE`.
- **Dokumentacja:** README, `docs/RESULTS_CONTRACT.MD` (§4, słownik luk), SUPPORT.MD (wiersz „Runtime tracing, dynamic loader identity and JNI models”: `Not implemented` → `Partial` — był nieaktualny od N1), RUNTIME_NATIVE_IMPLEMENTATION_PLAN (N5).
- **Ograniczenia:** brak streszczeń Java ↔ natywne (co kod natywny woła), `RegisterNatives`, FFM API (`java.lang.foreign`), metody natywne JDK (poza wejściem); osiągalne metody natywne w Tomcacie/jffi to w większości wiązania opcjonalne (APR), które CHA osiąga szeroko.
- **Następny krok:** pełne N5 (streszczenia) i N2–N4 — odłożone do realnego celu.

### Plan CI floty — aktualizacja i prototyp raportu zbiorczego (`cmd/ariadne-fleet`) (2026-11-23)

- **Kontekst:** plan [FLEET_CI_PLAN.MD](FLEET_CI_PLAN.MD) (propozycja od 2026-11-15, dotąd niezacommitowana) czeka na decyzję menedżera w połowie grudnia; na prośbę użytkownika zaktualizowany i zacommitowany, z prototypem raportu zbiorczego.
- **Aktualizacja planu:** sekcja „Stan narzędzia (2026-11-23)” (zamknięte obszary; tryb biblioteki, `-alt-versions`, v2 `frameworks`, ocena runtime'u, snapshoty); decyzja 3 uzupełniona o ruch do Central przy `-alt-versions` (4–24 pobrań `maven-metadata.xml` na bibliotekę — włączać tylko przy proxy); nowa decyzja 4 — opis runtime'u aplikacji (`-runtime-file` z pliku w repozytorium, bo JDK runnera to nie JDK wdrożenia); krok CI zapisuje `ariadne-results/meta.json` (`artifact`, `commit`, `branch`, `status`, `reason`); usunięta nieaktualna uwaga o `FRAMEWORKS:F3`.
- **Prototyp `cmd/ariadne-fleet`** (`-in <katalog> [-out fleet.md]`): podkatalog na artefakt z wynikiem v2 (`results.json` albo plik JSON z `schema_version` `2`) i opcjonalnym `meta.json`; tryb biblioteka/aplikacja z `callgraph.roots`. Raport: osobne tabele bibliotek (liczby werdyktów, `REACHABLE` z `supplied_by: library` vs reszta, `SUPPLIED_DISPATCH_ONLY`, maks. narażonych metod API, pakiety Springa jako framework, advisories usuwane przez najnowszą wersję w linii) i aplikacji; lista `REACHABLE` po liczbie artefaktów i ważności; nieudane analizy z powodem; wersje narzędzia i daty danych OSV z ostrzeżeniem przy różnicach; blok interpretacji werdyktów.
- **Testy:** `TestFleetReport` na syntetycznych wynikach `testdata/fleet` (biblioteka, aplikacja, nieudana analiza, inna wersja narzędzia, wersja nieoceniona, zerowa data OSV). Smoke test na prawdziwych wynikach v2 (5 bibliotek LB4 z `-alt-versions newest`, petclinic/WebGoat/JPetStore): wyłapał błąd (pole `callgraph.roots`, nie `callgraph.config.roots` — biblioteki wychodziły jako aplikacje) i zerową datę OSV, oba poprawione; liczby zgodne z pomiarami (configuration2: 778 metod API; jjwt-jackson: 22 advisories usuwane w linii). `go test ./...`, `go vet ./...`.
- **Nie zrobione (czeka na decyzje planu):** `scripts/ci-scan.sh`, dystrybucja binarki, powiązanie biblioteka × aplikacje floty, sekcja README „Uruchomienie w CI dla wielu aplikacji”.

### LIBRARY_MODE:LB3 — zamknięcie (2026-11-23); obszar LIBRARY_MODE DONE i zarchiwizowany

- **Decyzja użytkownika (2026-11-23):** zamknąć LIBRARY_MODE po `lb3-2`.
- **LB3 → DONE for this scope:** kryterium gotowości („wynik mówi, której wersji zależności dotyczy każde ustalenie i kto ją dostarcza”) spełnione przez `lb3-1`; opcjonalna ocena innych wersji — `lb3-2`. Odłożone na przyszłość: tryb `rerun` (flaga zarezerwowana, odrzucana z „not built yet”) i inne strategie wyboru.
- **Obszar LIBRARY_MODE (LB0–LB4) DONE:** plan i opis przeniesione do [archive/2026-11-23/library_mode/](../archive/2026-11-23/library_mode/COMPLETION.md) z `COMPLETION.md` (obok archiwów LB2, LB4); link w `docs/RESULTS_CONTRACT.MD` i indeks RECOMMENDATION/README.MD zaktualizowane. Powiązanie z CI floty — `FLEET_CI_PLAN.MD`.
- **Dokumentacja:** SUPPORT.MD bez zmiany statusu („Library mode” `Partial`).

### LIBRARY_MODE:LB3 — przyrost `lb3-2`: inne wersje zależności, które użytkownik może dostać — DONE for this scope (2026-11-23)

- **Decyzja użytkownika (2026-11-23):** ocena wersji spoza wejścia — opcja A (OSV i wariant mapowania tej wersji, osiągalność na kodzie analizowanej) pod dedykowaną flagą, z miejscem na B (drugi przebieg grafu z jarem tej wersji); wybór wersji — X (najwyższa w zadeklarowanym zakresie + najnowsza w linii major.minor), też pod flagą, z miejscem na inne strategie.
- **Zrobione:** `-alt-versions off|newest` (domyślnie `off`), `-alt-versions-mode approx` (`rerun` zarezerwowany — odrzucany z komunikatem „not built yet”); wymaga `-roots api`. `internal/component/altversions.go` (`AltVersionCandidates`, `chooseNewest`: `range_max`, `line_latest`; lista wersji z `maven-metadata.xml` przez cache; `Inventory.WithInstanceVersion`), `reach.Engine.WithComponents` (płytka kopia: te same przebiegi BFS, inwentarz z inną wersją instancji — wybór wariantu mapowania i `MAPPING_VERSION_UNCOVERED` jak dotąd), `cmd/ariadne/altversions.go` (OSV tylko dla wybranych GAV, mapowania tylko dla nowych advisories, ocena per wersja). Wynik: v2 `coverage.api.alternate_versions` (strategia, tryb, `rules_version` `lb3-2`, zestawy z wersjami, `chosen_by`, `basis`, `not_assessed`, ustalenia) — nigdy w `findings`, v1 ani werdyktach. Brak listy wersji → notatka; brak OSV/mapowań → `not_assessed`, przebieg idzie dalej. Snapshot: `config` `alt-versions`, `alt-versions-mode`; `algorithms` `library.alt_versions`.
- **Testy:** `TestChooseNewest`, `TestMajorMinor` (`internal/component`), `TestWithComponentsAltVersion` (`internal/reach`: wariant innej wersji na tym samym kodzie, `MAPPING_VERSION_UNCOVERED`, oryginał bez zmian), `TestLibraryAltVersions` i `TestLibraryAltVersionsFlags` na nowym fixture `testdata/library-altversions` (pom przez `-library-pom` do `testdata/library-deps/lib.jar`; commons-lang3 `[3.11,3.18.0)` → 3.17.0, CVE-2025-48924 `REACHABLE`; commons-text bez listy wersji → notatka; bez odpowiedzi OSV → `not_assessed`; v1 bajtowo identyczne; ustalenia bazowe bez zmian; flagi odrzucane bez `-roots api`, dla nieznanej strategii i `rerun`). Cache fixture: `maven-metadata.xml` commons-lang3 (z cache 2026-11-10) i jedna odpowiedź OSV (pobrana 2026-11-23 samym klientem OSV, bez Maven Central). `go test ./...`, `go vet ./...`.
- **Pomiar (5 bibliotek LB4, `-roots api`, CHA; jeden przebieg online po listy wersji — 0 × HTTP 429, potem offline):** żadna biblioteka nie deklaruje zakresu — każdy wybór to `line_latest`. configuration2: 3 zależności, nic nowszego w linii; jackson-yaml: 3 wersje (jackson-databind 2.12.3 → 2.12.7.2: 11 z 15 advisories zostaje), jjwt-jackson: 2 (jackson-databind 2.9.10.4 → 2.9.10.8: 22 z 32 znika, 10 zostaje, `REACHABLE`), mybatis-starter: 11 (np. spring-expression 5.2.4 → 5.2.25: 5 z 8 zostaje), jetty-client: 3 (jetty-http 9.4.36 → 9.4.58: 2 z 5 zostaje). Żadna nowsza wersja nie przyniosła advisory, którego nie miała analizowana.
- **Dokumentacja:** README (opis i tabela flag), `docs/RESULTS_CONTRACT.MD` (`coverage.api.alternate_versions`), SUPPORT.MD (wiersz trybu biblioteki), LIBRARY_MODE_IMPLEMENTATION_PLAN (LB3).
- **Ograniczenia:** `approx` — kod innej wersji może się różnić (dlatego `basis`); zależności bez instancji w wejściu (nieobecne, niezidentyfikowane) nie są oceniane; nadpisania wersji w aplikacji użytkownika niewidoczne.
- **Następny krok (przyszłość, decyzja użytkownika):** tryb `rerun` (opcja B) i inne strategie wyboru (granice zakresu, wszystkie wersje).

### ROADMAP:B0, B3, B4 — ocena i zamknięcie (2026-11-23) — kontrakty B0–B4 DONE

- **Tryb:** praca autonomiczna na prośbę użytkownika (2026-11-23); zamknięcie oparte na kryteriach gotowości i istniejących pomiarach, do cofnięcia przy przeglądzie PR.
- **B0 → DONE for this scope.** Kryterium („ta sama luka nie jest różnie interpretowana przez dwa analizatory, eksport legacy nie ukrywa niepewności”): wspólny słownik luk (`b0-3`) dla wszystkich analizatorów; reguła luk oznacza `NOT_REACHABLE` jako `MODEL_INCOMPLETE` także w v1 (petclinic 29/29 — oznaczenie jest uczciwe, bo istotne nieograniczone miejsca to głównie refleksja 512 i deserializacja 99, długi ogon w dziesiątkach bibliotek, `b0-5`); dowód kroku ścieżki i charakter ustalenia (`b0-1`), ścieżka bez kroków tylko-CHA (`b0-2`); heurystyka ścieżek przykładowych odrzucona pomiarem (2026-11-20).
- **B3 → DONE for this scope.** Kryterium („model usługi może wprowadzić typ, który uruchamia kolejne poprawne rozwinięcie grafu; limit zatrzymania widoczny”): wspólny punkt stały modeli i bramek twórców z raportowanym budżetem, niezależny od kolejności (`b3-2`), wybór uczestników z profilu w raporcie/v2/snapshotcie (`b3-3`), wersje reguł każdego rodzaju krawędzi (`b3-4`, test wymusza). Świadomie niebudowane: (a) jednorazowe przebiegi (bazowe callbacki, MethodHandle, refleksja) w punkcie stałym — ich wejście to fakty przepływu wartości i hierarchia, nie krawędzie grafu, więc iteracja nie doda krawędzi; ich wersje są w `algorithms`; (b) unieważnianie zależnych wyników per model zamiast per przebieg — porównanie snapshotów i tak nazywa zmianę wersji; zysk tylko w zwięzłości raportu.
- **B4 → DONE.** Tabela kwalifikacji: każdy obszar ma przypadki i wyrocznię JVM albo jawne „nie dotyczy” (mapowania CVE, runtime JDK); wiersz „Runtime JDK” zaktualizowany o katalog i testy F9.
- **Weryfikacja:** przegląd rejestru (wpisy ROADMAP:B0/B3/B4), `internal/analysis`, `internal/models`, `cmd/ariadne/snapshot_versions_test.go`; bez zmian w kodzie.
- **Dokumentacja:** ROADMAP.MD (stany B0, B3, B4 — dokument zostaje jako zasady dla planów), wiersz przeglądu.

### GRAPH_PRECISION:P3 — zamknięcie (2026-11-23); obszar GRAPH_PRECISION DONE i zarchiwizowany

- **Tryb:** praca autonomiczna na prośbę użytkownika (2026-11-23); zamknięcie oparte na istniejącym pomiarze, do cofnięcia przy przeglądzie PR.
- **P3 → DONE for this scope:** kryterium gotowości („dwa receivery, alias pól, tablica i fabryka dają uzasadnione różnice względem RTA; metryki wskazują koszt i jakość, zanim zwiększy się czułość”) spełnione przez `p3-1` — wyrocznia `testdata/pta` (9 miejsc, każde z 10 wywołań JVM zachowane) i pomiar `testdata/graphbench/RESULTS.MD` (0 zmian werdyktów przy 2,7–3,7× czasu RTA). Reguła stopu: większa czułość niebudowana.
- **Obszar GRAPH_PRECISION (P0–P4) DONE:** pliki obszaru przeniesione do [archive/2026-11-23/graph_precision/](../archive/2026-11-23/graph_precision/COMPLETION.md) z `COMPLETION.md`; PLAN.md i indeks zaktualizowane. Domyślny tryb CHA, RTA/PTA opt-in.
- **Dokumentacja:** SUPPORT.MD bez zmiany statusu („RTA and points-to” `Partial`).

### DYNAMIC_LINKAGE:L3 i L4 — ocena i zamknięcie (2026-11-23); obszar DYNAMIC_LINKAGE DONE i zarchiwizowany

- **Tryb:** praca autonomiczna na prośbę użytkownika („bez mojego udziału”, 2026-11-23); zamknięcie oparte na pomiarze, do cofnięcia przy przeglądzie PR.
- **L3 (adaptacje i bootstrapy) — pomiar:** kombinatory `bindTo`/`asType` — MethodHandle to 2 z 611 istotnych luk petclinic (`b0-5`); `StringConcatFactory` — tymczasowy test w `internal/classfile` (niezacommitowany) przeszedł po klasach: miejsca konkatenacji `invokedynamic` z argumentem typu spoza JDK (jedyne, dla których model dodałby krawędź `toString`): **petclinic z zależnościami 0** (3 miejsca konkatenacji, 29 188 klas, 6 381 `invokedynamic`), **WebGoat z zależnościami 1** (119 miejsc, 4 z obiektem, 1 typ spoza JDK; 42 317 klas). Biblioteki pod Javę 8 konkatenują przez `StringBuilder.append(Object)` — odbiorca `java.*`, nie rozwijany, więc to ogólna luka wywołań z JDK. **Wniosek: model niebudowany.**
- **L4 (weryfikacja):** dla zbudowanego zakresu L0–L2 spełnione — wyrocznie `testdata/samcall`, `testdata/async` (MethodHandle, `b4-2`), dowody w `path_edges`, wersje w snapshotcie (`b3-4`), luki według B0.
- **Obszar DYNAMIC_LINKAGE (L0–L4) DONE:** `DYNAMIC_LINKAGE.MD` i plan przeniesione do [archive/2026-11-23/dynamic_linkage/](../archive/2026-11-23/dynamic_linkage/COMPLETION.md) z `COMPLETION.md`; indeks zaktualizowany.
- **Weryfikacja:** pomiar opisany wyżej; `go test ./...`; bez zmian w kodzie produktu.

### FRAMEWORKS:F0 i F2 — zamknięcie (2026-11-23); obszar FRAMEWORKS DONE i zarchiwizowany

- **Decyzja użytkownika (2026-11-23):** zamknąć F0 i F2.
- **F0 → DONE:** już DONE w przeglądzie (oddzielenie wykrycia frameworka od aktywnego profilu, kontrakt scenariusza `internal/profile/scenario.go`); brakował tylko znacznik w planie.
- **F2 → DONE for this scope:** siedem wycinków i późniejsze (`s2`, `l1`/`l2`, `j1`); kryterium gotowości sprawdzone wyroczniami: bean/qualifier (`TestDIInjectionOracle`, `f5-3`), konflikt nazw (F1), nieznany profil → `unknown`, kontroler z helperem (`TestMVCOracle`, `f5-2`: `unmapped()` nieosiągalny); tożsamość parametrów dla taint (D4). Odłożone: programowalne fabryki, filtr `@Profile` dla kandydatów wstrzyknięć.
- **Obszar FRAMEWORKS (F0–F5) DONE:** `FRAMEWORKS.MD` i plan przeniesione do [archive/2026-11-23/frameworks/](../archive/2026-11-23/frameworks/COMPLETION.md) z `COMPLETION.md`; linki (PLAN.md, archiwa faz) i indeks RECOMMENDATION/README.MD zaktualizowane.
- **Dokumentacja:** SUPPORT.MD bez zmiany statusów.

### FRAMEWORKS:F3, F4 i LIBRARY_MODE:LB2, LB4 — zamknięcie (2026-11-23) — DONE for this scope

- **Decyzja użytkownika (2026-11-23):** zamknąć wszystkie cztery fazy zgodnie z rekomendacjami ocen (PR #137, #139).
- **Przeniesione:** teksty faz do `archive/2026-11-23/frameworks/FRAMEWORKS_F3.md`, `FRAMEWORKS_F4.md` i `archive/2026-11-23/library_mode/LIBRARY_MODE_LB2.md`, `LIBRARY_MODE_LB4.md` z zapisami ukończenia (zrobione, odłożone, weryfikacja); w planach wskaźniki.
- **Stan obszarów:** FRAMEWORKS — F1, F3, F4, F5 rozstrzygnięte; F0 i F2 bez formalnego zamknięcia (oba zrobione w wycinkach: F0 model scenariusza, F2 siedem wycinków DONE for this scope) — obszar zostaje w indeksie. LIBRARY_MODE — LB0, LB1, LB2, LB4 rozstrzygnięte; reszta LB3 wstrzymana decyzją projektową użytkownika — obszar zostaje w indeksie.
- **Dokumentacja:** plany, ten wpis, wiersze przeglądu; SUPPORT.MD bez zmian (opisuje zachowanie, wiersze `Partial`).
- **Następny krok:** formalne zamknięcie FRAMEWORKS F0/F2 (wtedy archiwizacja obszaru) — do decyzji.

### LIBRARY_MODE:LB2 i LB4 — ocena reszty (2026-11-23) — kryteria gotowości spełnione; rekomendacja: zamknąć z jawną listą odłożonych

- **LB2 (kod klienta i wejścia frameworka), punkty planu → stan:**
  - *miejsca, gdzie biblioteka woła kod klienta → luka „kod klienta”* — `lb2-1`/`lb2-2`, `lb2-7` (`supplied_dispatch`); przypadek (d) w `testdata/library`.
  - *ścieżki warunkowe po powrocie z kodu klienta* — **nie zrobione; semantyka „do zdefiniowania”** (plan): wymaga reguły na poziomie CFG (podatne wywołanie zdominowane przez wywołanie kodu klienta, którego wynik je warunkuje); brak zmierzonego przypadku w zestawie LB4.
  - *wejścia z autokonfiguracji ogłaszanej przez bibliotekę → korzenie z pochodzeniem* — `lb2-5` (Spring Boot), `lb2-6` (CDI/Jandex), `lb2-8` (klasy ogłaszane przez framework), `lb2-9` (Micronaut 3.x).
  - *weryfikacja: przypadek (d), wyrocznia z przykładowym klientem* — `scripts/library-oracle.sh` uruchamia klienta `fx.Client` z własnym `Listener` dla `Pipeline#run`; `observed.txt` zapisuje tylko ciała biblioteki, więc wykonanie listenera klienta nie jest zapisane (drobne uzupełnienie, opcjonalne).
  - zmierzone i świadomie niebudowane: wartości dostarczone przekazane do metod zależności (2026-11-21: 0–107 miejsc, żaden werdykt by się nie zmienił), `@ConditionalOn…` (0 z 26 ustaleń), typ pola z konstruktorów dla odziedziczonych (jackson: bez zysku); niesprawdzone: Micronaut 4.x.
  - **Kryterium gotowości** („callback klienta jest jawną luką, autokonfiguracja biblioteki — korzeniem”): spełnione.
- **LB4 (kwalifikacja i CI), punkty → stan:** zestaw 5 bibliotek z ręcznymi oczekiwaniami i wyroczniami JVM (`lb4-1`, `lb4-2`; 0 `NOT_REACHABLE` wykonanych), README „Library mode in a library's CI”, SUPPORT.MD `Partial` ze zmierzonymi ograniczeniami — **zrobione**; powiązanie z planem CI floty — zależy od decyzji o CI (połowa grudnia), należy do `FLEET_CI_PLAN.MD`, nie do LB4. **Kryterium gotowości** (SUPPORT.MD opisuje tryb jako `Partial` ze zmierzonymi ograniczeniami): spełnione.
- **Rekomendacja:** zamknąć LB2 i LB4 jako „DONE for this scope” (odłożone: ścieżki warunkowe po kodzie klienta, wartości dostarczone do zależności, `@ConditionalOn…`, Micronaut 4.x; powiązanie z flotą → plan CI). LIBRARY_MODE zostałby otwarty tylko przez wstrzymaną resztę LB3. **Decyzja użytkownika.**
- **Weryfikacja:** przegląd planu, rejestru, `testdata/library` (EXPECTED.MD, `library-oracle.sh`) i testów `library_mode_test.go`; bez zmian w kodzie.

### REFLECTION:R3 i DYNAMIC_LINKAGE:L2 — zamknięcie (2026-11-23) — DONE for this scope; REFLECTION DONE i zarchiwizowany

- **Decyzja użytkownika (2026-11-23, propozycja „zamknięcie obszarów uśpionych”):** obie fazy mają zmierzony zerowy zysk — zamknąć bez dalszej budowy.
- **REFLECTION:R3:** zrobiony wycinek 3 (`r3-3`); konfiguracja, `setAccessible`/`trySetAccessible`, `Constructor` przez parametr niebudowane — pomiary 2026-11-05 i 2026-11-09 bez zysku, spis `r3-3`: 2 miejsca w trzech aplikacjach; odmowa dostępu rozróżniona od celu przez `r4-2`. **Obszar REFLECTION (R0–R4) DONE:** `REFLECTION.MD` i plan przeniesione do [archive/2026-11-23/reflection/](../archive/2026-11-23/reflection/COMPLETION.md) z `COMPLETION.md`; indeks i linki zaktualizowane.
- **DYNAMIC_LINKAGE:L2:** pomiar 2026-11-21 (0/10 i 2/46 miejsc, 2 z 611 istotnych luk) — tekst fazy w [archive/2026-11-23/dynamic_linkage/DYNAMIC_LINKAGE_L2.md](../archive/2026-11-23/dynamic_linkage/DYNAMIC_LINKAGE_L2.md); obszar zostaje otwarty (L3–L4 nieocenione).
- **Dokumentacja:** RECOMMENDATION/README.MD, PLAN.md (link), plany; SUPPORT.MD bez zmian (wiersze „Reflection” i MethodHandle pozostają `Partial`, opisują zachowanie).

### FRAMEWORKS:F3 i F4 — ocena reszty (2026-11-23) — prawie całość zrobiona wycinkami; rekomendacja: zamknąć obie fazy z jawną listą odłożonych

- **F3 (proxy i ograniczony model AOP), punkty planu → stan:**
  1. *tożsamość proxy/target/interceptor; JDK Proxy: handler + `Method`* — tożsamość: wycinek 1 (`internal/aop`); katalog proxy MyBatis/Spring Data/Spring ORM/`@Transactional` (`framework:proxies` `x1`–`x3`, wyrocznie `mybatis`, `springdata`, `jpa`, `tx`); **ogólny JDK Proxy (wycinek 2) — NOT STARTED po pomiarze 2026-11-09: 0 z 12 relacji `proxy` ma ograniczony zbiór interfejsów** (reguła stopu).
  2. *JDK/CGLIB, final/private, self-invocation* — `spring:aop` `a1`/`a2`, wyrocznia `testdata/aop` (CGLIB z self-invocation i metodą final, proxy JDK).
  3. *ograniczona semantyka pointcutów i advice, nieznane oznaczone, `around` warunkuje* — `a1`/`a2` (podzbiór AspectJ jak w Springu, wiązanie zmiennych, XML `<aop:config>`).
  4. *Spring Data bridge → katalog z deskryptorami, stara reguła jako fallback* — `x1`/`x2`; `synthetic_edges` `Repository → SimpleJpaRepository` zostaje w `spring.yaml` jako jawny fallback.
  5. *katalog typów tworzonych przez DI/proxy dla GRAPH_PRECISION, typ generowany oznaczony* — katalog P1 (`container`, `model`) i luka `container-proxy` (`internal/alloc`).
  - **Kryterium gotowości** (proxy/target/self-invocation, dwa handlery, warunkowe advice, typ generowany bez mieszania ścieżek): spełnione poza „dwoma handlerami” ogólnego JDK Proxy (wycinek 2).
- **F4 (rozszerzanie wejść i frameworków), punkty → stan:**
  1. *schedulery i event/message listeners* — `spring:events` `ev1` (zdarzenia, `@Scheduled`/`@Schedules` jako korzeń, wyrocznia `testdata/events`); listenery wiadomości tylko jako korzenie z adnotacji — model odłożony w F5 (broker).
  2. *Servlet/Jakarta/Micronaut/Quarkus z rzeczywistych artefaktów* — profile z aplikacjami kontrolnymi (`scripts/prepare-*-demo.sh`), dispatch kontenera servletów (`servlet-container:dispatch`), definicje beanów Micronauta w trybie biblioteki (`lb2-9`); Spring MVC z wyrocznią (`f5-2`).
  3. *programowalne rejestracje, dynamiczne profile, SpEL* — poza zakresem (plan: „dopiero w jawnym zakresie”); `@Profile` jako wyrażenia od `f5-3`.
  4. *osobny tryb klasyfikacji `reach` dla modelowanych wejść* — nie zrobiony; werdykt `FRAMEWORK` i reguła luk B0 (decyzja 2026-11-20) pokrywają dziś granicę frameworka; brak zmierzonego przypadku, który by go wymagał.
  - **Kryterium gotowości** (lista wejść wersjonowana; nowe wejścia nie zmieniają niejawnie znaczenia wyników): spełnione przez `profile.rules`/`profile.builtin` (`f5-1`) i strażnika `TestRulesVersion_tracksRuleSource`.
- **Rekomendacja:** zamknąć F3 i F4 jako „DONE for this scope” z odłożonymi: ogólny JDK Proxy (zmierzony 0/12), model listenerów wiadomości, SpEL/programowalne rejestracje, osobny tryb klasyfikacji. Wtedy FRAMEWORKS F0–F5 byłby rozstrzygnięty i obszar można zarchiwizować (F2 to siedem wycinków DONE for this scope). **Decyzja użytkownika.**
- **Weryfikacja:** przegląd wpisów rejestru i kodu (`internal/aop`, `proxycatalog`, `springaop`, `springevents`, `alloc`), bez zmian w kodzie.

### CVE:F9 dla przyrostu F8 (ocena runtime'u JDK) — DONE (2026-11-23); obszar CVE DONE i zarchiwizowany

- **Zadania F9 względem przyrostu F8:**
  - *1 (CLI, pomoc, README, JSON, raport, ograniczenia):* `-jdk-home`, `-runtime-file`, `-openjdk-catalog` w pomocy i tabeli flag README, sekcja raportu „Target runtime”/„Target runtime advisories”, v2 `target_runtime`, ograniczenia w SUPPORT.MD (F8a/F8b, zamknięcie F8c) — sprawdzone, bez zmian.
  - *2 (raportowanie osobno):* `target_runtime` oddzielnie od ustaleń komponentów; `applicability` z katalogu, `reachability=not_assessed` — sprawdzone testem.
  - *3 (migracja ręcznych mapowań do v2):* nie dotyczy (F8 nie ma mapowań).
  - *4 (eksport v1):* wyniki v1 bajtowo identyczne bez runtime'u, z Temurin 17 i z Debianem 25 — `TestTargetRuntimeE2E`; petclinic (`-roots profile`) bez runtime'u i z `-jdk-home` Temurin 17: v1 identyczne.
  - *5 (stabilność przy tym samym snapshotcie):* dwa przebiegi → ten sam `analysis_id`; odtworzenie snapshotu → ta sama ocena runtime'u i ustalenia (inny `analysis_id` z projektu: `replayed`). **Znalezione i naprawione:** (a) notatka o wieku katalogu liczona od zegara — odtworzenie > 100 dni po nagraniu dawałoby inną treść; teraz przy `-snapshot-in` liczona od `created_at` snapshotu (`assessTargetRuntime(…, now)`); (b) własny katalog `-openjdk-catalog` zapisywany tylko jako wersja — szkic z tą samą wersją i inną treścią był nieodróżnialny; teraz hashowany jako wejście `openjdk-catalog` (jak `-sbom`, `-profile-file`). Tożsamość runtime'u już była w `config["target-runtime"]`.
  - *6:* `go test ./...`, `go vet ./...`.
  - *7 (aplikacja kontrolna):* petclinic z Temurin 17.0.9: 88 advisories — 70 `affected`, 18 `not_affected`, wszystkie `not_assessed`; notatka „runtime starszy niż najstarsze advisory katalogu dla linii 17”; v1 bez zmian.
  - *8 (błędy danych):* oba wejścia naraz, katalog bez pliku `release`, uszkodzony `-runtime-file`, uszkodzony i brakujący `-openjdk-catalog` → kod błędu i **żaden** `r.md`/`r.json` niezapisany — `TestTargetRuntimeErrors`.
  - *9 (artefakty spięte identyfikatorem):* jak dla F0–F7 (`analysis_id` w raporcie, v2 i manifeście).
- **Testy (nowe):** `TestTargetRuntimeE2E`, `TestTargetRuntimeSnapshot` (mutacja: notatka od zegara — wyłapana), `TestTargetRuntimeErrors` (`cmd/ariadne/targetruntime_e2e_test.go`).
- **Kryteria wydania F9:** scenariusze kontrolne wyjaśnione; niepewność jawna (`not_assessed`, `not_performed` z powodem, notatki o pokryciu); brak zapytań sieciowych offline (ocena z katalogu osadzonego); dowody i wejścia każdego rozstrzygnięcia w snapshotcie (runtime w konfiguracji, katalog jako wejście); dokumentacja rozróżnia działające/nieobsługiwane (SUPPORT.MD).
- **Obszar CVE (F0–F10) DONE:** `CVE.MD` i `CVE_IMPLEMENTATION_PLAN.MD` przeniesione do [archive/2026-11-23/cve/](../archive/2026-11-23/cve/COMPLETION.md) z `COMPLETION.md`; linki zaktualizowane; wiersz w SUPPORT.MD „Target JDK vulnerability assessment” zostaje `Partial` (tylko Temurin, katalog od 2026, bez werdyktu metodowego).
- **Dokumentacja:** README (opis snapshotu: katalog jako wejście, datowanie notatki), SUPPORT.MD, RECOMMENDATION/README.MD (indeks), ten wpis.

### CVE:F8c — zamknięcie (2026-11-23) — bez budowy; CVE:F8 DONE

- **Decyzja użytkownika (2026-11-23):** po ocenie (wpis niżej, PR #131) F8c zamknięte bez budowy — ani werdykt metodowy dla CVE JDK, ani raportowy sygnał `api_exposure` (`f8c-1`) nie powstają. Ustalenia runtime'u bez zmian: `applicability` z katalogu, `reachability=not_assessed` z powodem.
- **CVE:F8 DONE:** F8a, F8b i F8c rozstrzygnięte; tekst fazy przeniesiony do [archive/2026-11-23/cve/CVE_F8.md](../archive/2026-11-23/cve/CVE_F8.md) z zapisem ukończenia i kryteriami F8; w planie wskaźnik.
- **Obszar CVE zostaje otwarty:** F9 (integracja, migracja i kwalifikacja wydania) dla przyrostu F8 nie została powtórzona, jak wymaga jej własna treść.
- **Dokumentacja:** SUPPORT.MD („Target JDK vulnerability assessment”) — dopisane, że werdykt metodowy dla CVE JDK świadomie nie jest budowany; README/FAQ bez zmian (nie obiecywały go).
- **Następny krok:** F9 dla F8 albo inny obszar.

### RUNTIME_NATIVE:N2 — ocena (2026-11-23) — nie zaczęty; odłożony do pierwszego celu z wieloma kontekstami ładowania

- **Kryteria N2 względem `main`:**
  - *jawne plugin JAR-y i deklaracja kontekstu loadera:* brak — wejścia to `-app`, `-deps`, `-deps-list`, WAR/EAR, `-jpms-scenario`; hierarchia jest jedną mapą FQCN (`internal/hierarchy`), kolizje rozstrzygane deterministycznie (app, potem zależności w kolejności) i zapisywane (`Hierarchy.Duplicates`), a CVE F1/F6 tworzy z nich osobne instancje komponentów;
  - *relacje i delegacja loaderów, widoczność:* brak modelu po stronie statycznej; po stronie śladu kontrakt `ariadne-runtime-trace-1` (N0/N1) już niesie loadery (`loader`: id, nazwa, typ) i definicje klas per loader z SHA-256 (`internal/runtimeevents`) — importer odróżnia dwie definicje tej samej nazwy;
  - *klasy przechwycone w runtime (hash przed/po transformacji), klasy generowane bez tożsamości Maven, zakaz tworzenia ścieżek między kontekstami:* brak (dziś import N1 jest dowodem, nie częścią grafu, więc ścieżek między kontekstami też nie tworzy).
  - **Kryterium gotowości** („dwie definicje tej samej klasy pozostają odrębne, znany plugin ma warunkowe/zaobserwowane użycie”): spełnione tylko po stronie śladu.
- **Zapotrzebowanie:** żadna aplikacja referencyjna (petclinic, WebGoat — płaski classpath Boot; JPetStore — jeden WAR) nie ma kilku kontekstów ładowania; flota to głównie biblioteki (tryb biblioteki, `-roots api`), gdzie kontekst ładowania nie istnieje. N2 jest dużym przyrostem (hierarchia z kontekstem w tożsamości klasy dotyka grafu, CVE F1, snapshotów).
- **Rekomendacja:** odłożyć N2 (i zależne N3–N5) do pierwszego rzeczywistego celu z wtyczkami lub wieloma modułami ładowanymi osobno (np. EAR z izolowanymi modułami, Jenkins, OSGi); wtedy zacząć od fixture z dwiema definicjami tej samej klasy i wyrocznią śladu.
- **Weryfikacja:** przegląd kodu (`internal/hierarchy`, `internal/component`, `internal/runtimeevents`) i logów przebiegów v2 na 3 aplikacjach; bez zmian w kodzie. Dokumentacja: RUNTIME_NATIVE_IMPLEMENTATION_PLAN (N2), ten wpis.
- **Następny krok:** brak w RUNTIME_NATIVE do czasu takiego celu.

### FRAMEWORKS — fakty frameworka w v2 JSON (`frameworks`) — DONE (2026-11-23)

- **Zakres:** luka odnotowana w ocenie F5 — fakty profilu były tylko w logu i `report.md`. Nowa addytywna sekcja v2 `frameworks` (`internal/resultsv2/frameworks.go`, `resultsv2.BuildFrameworks`): `profile`, `rules_version`, `active_profiles`; `beans[]` (nazwa w kontenerze — jawna albo domyślna, źródło `stereotype`/`bean_method`/`repository`, `primary`, `qualifier`), `conditions[]` (`@Profile` i stan), `injections[]` (punkt, typ, qualifier, `@Value`, `resolved`, kandydaci, powód), `entries[]` (wejścia zewnętrzne), `roots[]` (korzenie cyklu życia beanów, XML i ciał Spring MVC z pochodzeniem), `summary`. Listy sortowane; `schema_version` bez zmian (jak `deserialization`, `callgraph`, `target_runtime`). Nigdy część `reachability` ani wiersza v1.
- **Test:** `TestV2Frameworks` — przebieg `-json-format v2` na `testdata/di`: 4 wstrzyknięcia `Shop` zgodne z tym, co wstrzyknął JVM (wyrocznia `f5-3`), stany `@Profile` przy `-active-profiles default`, bean z nazwą domyślną, liczniki, dwa przebiegi bajtowo identyczne w sekcji.
- **Aplikacje referencyjne (`main` vs gałąź, `-roots profile`, v2):** dokument poza `frameworks` (i `analysis_id`/`tool`) identyczny na wszystkich trzech. Sekcja: petclinic 9,6 KB (12 beanów, 8/8 wstrzyknięć, 17 mapowań HTTP, 9 korzeni), WebGoat 116 KB (172 beany, 118 punktów: 93 rozwiązane, 19 `@Value`, 6 nierozstrzygniętych; 159 mapowań; 199 korzeni), JPetStore 5 KB (3 beany, 8 punktów nierozstrzygniętych — beany z XML nie są w katalogu wstrzyknięć; 6 wejść `web.xml`, 8 korzeni).
- **Weryfikacja:** `go test ./...`, `go vet ./...`. Dokumentacja: `docs/RESULTS_CONTRACT.MD` (wiersz `frameworks`), README (`-json-format`), SUPPORT.MD („Profiles and selected roots”).
- **Ograniczenia:** beany z XML Springa nie biorą udziału w rozwiązywaniu wstrzyknięć (JPetStore: 8 nierozstrzygniętych); kandydaci niefiltrowani przez `@Profile`; sekcja nie zależy od `-roots` (te same fakty w obu scenariuszach).
- **Następny krok:** raport zbiorczy floty (`ariadne-fleet`, plan CI — grudzień) może czytać `frameworks.summary`.

### ROADMAP:B1 i B2 — ocena (2026-11-23) — oba spełnione przez istniejący kod

- **B2 (wspólna reprezentacja kodu i wartości) → DONE**, wykonany przez DATA_FLOW D0–D5 (zamknięty 2026-11-07):
  - tabela wyjątków zachowana w parserze (`classfile.Method.ExceptionTable`) i używana przez CFG (`internal/classfile/cfg.go`, bloki handlerów) — bez drugiego formatu;
  - domena wartości (`internal/valueflow`): stałe tekstowe/klasowe (`KindConst`), typy (`KindTyped`), alokacje (`KindAlloc`), `null`, skończone zbiory (`KindSet`), nieznana (`KindUnknown`) oraz rodzaje funkcyjne/uchwyty/pola/wyniki wywołań/tablice/refleksja;
  - limity poszerzają do `KindUnknown` z powodem (`Merge` z budżetem, `MaxCallTraceDepth`, `Summary.Incomplete` w `internal/flowsum`) — kandydaci nie są odrzucani;
  - streszczenia międzyproceduralne przyrostowo (`internal/flowsum`, D3);
  - **kryterium gotowości** („dwa konsumenty używają tych samych faktów o argumencie bez ponownego dekodowania stosu”): z `internal/dataflow` korzysta 20 pakietów — m.in. `reflectinvoke`, `callback`, `methodhandle`, `serviceload`, `samcall`, `springevents`, `springaop`, `deserialize`, `libapi`, `pta`; bajtkod dekoduje wyłącznie `internal/classfile` (sprawdzone wyszukaniem opkodów w `internal/*`).
- **B1 (identyfikatory i dowody krawędzi) → DONE for this scope:**
  - miejsce wywołania: krawędź niesie `pc`, rodzaj i podstawę (`callgraph.EdgeID{From, To, Kind, Basis}`), linie tylko do prezentacji;
  - relacje oddzielone od wywołań: rejestracje callbacków (`callback.RegistrationSite`), katalog tworzonych typów (`alloc`, P1), inicjalizacja jako osobny rodzaj (`CallInit`), przepływ danych w `flowsum`/`taint`;
  - **kryterium gotowości** („dwa powody utworzenia tej samej krawędzi zachowane i odtwarzalne z raportowanej ścieżki”): ta sama para metod z dwiema podstawami to dwie krawędzie (`EdgeID` zawiera `Basis`), a `path_edges` v2 (`b0-1`, `reach.PathEdge.Evidence`) wypisuje dowody **wszystkich** krawędzi między węzłami kroku;
  - poza zakresem: tożsamość metody nie zawiera kontekstu ładowania — dziś deterministyczny wybór klasy (`hierarchy.Duplicates`, instancje CVE F1/F6); pełne izolowane classloadery należą do RUNTIME N1–N2, jak mówi sam B1.
- **Weryfikacja:** przegląd kodu (`internal/classfile`, `valueflow`, `dataflow`, `flowsum`, `callgraph`, `reach/pathevidence.go`) i wyszukanie dekodowania opkodów poza `classfile`; bez zmian w kodzie. Dokumentacja: ROADMAP.MD (B1, B2), ten wpis; SUPPORT.MD bez zmian.
- **Następny krok:** ROADMAP zostaje otwarty przez B0 (reguła luk) i B3 (przebiegi poza punktem stałym, unieważnianie per model).

### CVE:F8c — ocena (2026-11-23) — osiągalność kodu JDK na poziomie metod niewykonalna ze źródeł, które mamy; możliwy raportowy sygnał ekspozycji API

- **Materiał:** kuratorowany katalog OpenJDK (`internal/runtimevuln/openjdk_advisories.yaml`, wersja 1): 8 advisories z 2026, **88 CVE**; każdy wpis podaje tylko komponent JDK (`security-libs/java.security` 15, `core-libs/java.net` 13, `client-libs/2d` 8, `security-libs/javax.net.ssl` 7, `hotspot/compiler` 7, `security-libs/org.ietf.jgss` 6, …), bez metody, identyfikatora błędu JDK ani commita.
- **Kryterium 1 (mapowanie na kod tylko dla wiarygodnych przypadków):** brak źródła. Advisories OpenJDK nie podają błędów JDK; błędy bezpieczeństwa w JBS nie są publiczne; Red Hat CSAF (`access.redhat.com/hydra/rest/securitydata/csaf/…`) wiąże CVE z własnym Bugzillą, tytuły to „JDK: …” bez numeru błędu JDK (sprawdzone wyszukiwaniem, 2026-11-23). Jedyna droga do metod to diff kolejnych wydań aktualizacji (`github.com/openjdk/jdk*u`, tagi CPU) jak `release-diff` z CVE F10 — dałby **zbiór zmienionych metod per komponent**, nie per CVE (kilka poprawek na komponent w jednym CPU, tytuły commitów celowo ogólne), i wymagałby wczytania klas docelowego runtime'u do hierarchii (`-jdk-home`), czego kryterium 2 zabrania robić globalnie. **Wniosek: werdykt metodowy dla CVE JDK — nie budować** (wiarygodnego mapowania nie ma dla żadnego z 88).
- **Kryterium 3 (JIT/VM bez werdyktu):** 8 CVE `hotspot/*` i 4 narzędziowe (`tools/jlink`, `tools/jshell`, `core-svc`) — już dziś `reachability=not_assessed` (F8b), bez zmian.
- **Pomiar alternatywy — ekspozycja API komponentu** (graf `-roots profile` CHA każdej aplikacji, osiągalność od metod aplikacji, wywołania do publicznych pakietów komponentu: `javax.imageio.*`, `java.awt.*`, `org.ietf.jgss.*`, `javax.xml.crypto.*`, `org.omg.*`, `javax.sound.*`, …; skrypt roboczy, nie w repo):

  | Aplikacja | API komponentu nieużywane | tylko przez biblioteki | przez kod aplikacji | komponent bez mapowania API | hotspot + narzędzia |
  |---|---|---|---|---|---|
  | petclinic | 3 | 68 | 0 | 5 | 12 |
  | WebGoat | 9 | 26 | 36 | 5 | 12 |
  | JPetStore | 30 | 41 | 0 | 5 | 12 |

  CHA przez biblioteki osiąga prawie każde API (`javax.imageio` w petclinic: 31 celów przez Tomcat/Spring), więc sygnał jest użyteczny tylko w jedną stronę: „API komponentu nie jest wywoływane przez osiągalny kod” (JPetStore: 30 z 88, m.in. całe `client-libs`, `javax.net.ssl`, `org.ietf.jgss`). Nawet to nie jest `NOT_REACHABLE`: kod JDK woła własne wnętrza (SPI `ImageIO`, `ServiceLoader`, JNDI), a graf ich nie widzi.
- **Propozycja (`f8c-1`, do decyzji):** raportowe pole przy ustaleniu runtime'u — `api_exposure: none | libraries | application` z listą pakietów i powodem, bez zmiany `reachability` (dalej `not_assessed`); mapowanie komponent → pakiety jako wersjonowany katalog. Koszt mały; wartość: priorytetyzacja (JPetStore: 30 z 88 CVE z ekspozycją „none”), nie werdykt. Alternatywa: zamknąć F8c jako „DONE for this scope: werdykt metodowy świadomie niebudowany”.
- **Weryfikacja:** pomiar na grafach petclinic/WebGoat/JPetStore (`-roots profile`, CHA, offline, binarka z `main` + #130); bez zmian w kodzie. Dokumentacja: CVE_IMPLEMENTATION_PLAN (F8c), ten wpis; SUPPORT.MD bez zmian (stan produktu ten sam).
- **Następny krok:** decyzja użytkownika: `f8c-1` albo zamknięcie F8c.

### FRAMEWORKS:F5 — zamknięcie (2026-11-23) — DONE for this scope

- **Zakres wykonany:** ocena (2026-11-22) i przyrosty `f5-1`, `f5-2`, `f5-3` (PR #126–#129), każdy z wyrocznią albo testem wymuszającym i pomiarem na aplikacjach referencyjnych (0 zmian werdyktów).
- **Kryteria:** fixtures — 8 z 11 przypadków z wyrocznią JVM, 1 częściowo (wartości domyślne adnotacji: testy jednostkowe), 2 odłożone z powodem (konsument wiadomości: broker; priorytety konfiguracji: brak modelu źródeł właściwości); kontrolowane uruchomienia — tak, jedna wersja Springa 5.3.6; raportowanie — log i `report.md`, bez v2 JSON; snapshot — scenariusz i reguły (`f5-1`); B4 i dokumentacja — tak.
- **Przeniesione:** tekst fazy F5 z `FRAMEWORKS_IMPLEMENTATION_PLAN.MD` do [archive/2026-11-23/frameworks/FRAMEWORKS_F5.md](../archive/2026-11-23/frameworks/FRAMEWORKS_F5.md) z zapisem ukończenia; w planie wskaźnik. Obszar FRAMEWORKS zostaje otwarty (F3 reszta, F4 reszta).
- **Dokumentacja:** SUPPORT.MD bez zmiany statusu (wiersz „Profiles and selected roots” już opisuje `f5-2`/`f5-3`, zostaje `Implemented` dla profili; ograniczenia wymienione); README/FAQ bez zmian (opisują zachowanie, nie stan fazy).
- **Pozostałe ograniczenia:** jak w zapisie archiwum.
- **Następny krok:** CVE:F8c (ocena) albo fakty frameworka w v2 JSON (przy raporcie zbiorczym floty, grudzień).

### FRAMEWORKS:F5 — przyrost `f5-3`: qualifier/`@Primary` i dwa scenariusze profili w wyroczni DI — DONE (2026-11-22)

- **Wyrocznia (`testdata/di`, `scripts/di-oracle.sh`):** cztery beany `Store` (`@Primary`, jawna nazwa `@Component("slow")`, `@Qualifier("archive")` na klasie, nazwa domyślna) i `Shop` z czterema polami `@Autowired` (bez qualifiera, `@Qualifier("slow")`, `"archive"`, `"coldStore"`) — JVM wypisuje, co kontener wstrzyknął (`# inject …`); beany z `@Profile("eu")`, `"!eu"`, `"eu & cloud"`, `"eu | us"`, `{"us", "cloud"}`. `Main` przyjmuje aktywne profile z argumentów; skrypt uruchamia JVM dwa razy: bez profilu (`observed.txt`) i z `eu,cloud` (`observed-eu-cloud.txt`).
- **Wynik przed poprawką:** scenariusz domyślny zgodny. Przy `-active-profiles eu,cloud` graf nie osiągał konstruktorów `CloudEuTax` (`eu & cloud`) i `EitherTax` (`eu | us`) — `profileCondition` znał tylko nazwy i `!nazwa`, wyrażenie brał za nazwę profilu. Rozwiązywanie wstrzyknięć podawało `FastStore` dla `@Qualifier("archive")` (qualifier na klasie beana) i `@Qualifier("coldStore")` (nazwa domyślna) — JVM wstrzyknął `ArchiveStore` i `ColdStore`: qualifier porównywany tylko z jawną nazwą, a przy braku dopasowania kod przechodził do `@Primary` (błędny fakt zamiast „nierozstrzygnięte”).
- **Poprawka:** `evalProfileExpression` — wyrażenia Springa 5.1+ (`!`, `&`, `|`, nawiasy; mieszanie `&`/`|` bez nawiasów i błędna składnia → warunek nieznany, jak odrzucenie przez Springa); `BeanName.Qualifier` (`@Qualifier` klasy lub metody `@Bean`) i `BeanName.EffectiveName()` (nazwa domyślna jak `AnnotationBeanNameGenerator`/`Introspector.decapitalize`); `disambiguate`: qualifier pasuje do nazwy (jawnej lub domyślnej) albo `@Qualifier` beana, zero dopasowań → nierozstrzygnięte z powodem (Spring nie startuje), kilka → `@Primary` wśród nich. `profile.RulesVersion` `f5-2` → `f5-3` (wymuszone przez strażnika).
- **Testy:** `TestDIOracle` (scenariusz domyślny, nowe ujemne: 4 beany profili), `TestDIProfilesOracle` (eu,cloud: wszystkie ciała osiągalne, `UsTax` nie), `TestDIInjectionOracle` (4 z 4 wstrzyknięć zgodne z JVM), `TestEvalProfileExpression`, `TestBeanName_EffectiveName`; `go test ./...`, `go vet ./...`.
- **Aplikacje referencyjne (`main` vs `f5-3`, offline, `-profile spring`, `-roots profile`/`all`, CHA):** wyniki v1 bajtowo identyczne; rozwiązane wstrzyknięcia bez zmian (petclinic 8, WebGoat 93, JPetStore 0), `report.md` różni się tylko nagłówkiem (i losowymi ścieżkami tymczasowymi WAR-a) — żadna z aplikacji nie używa wyrażeń `@Profile` ani qualifierów, których dotyczy poprawka.
- **Ograniczenia:** kandydaci wstrzyknięć nie są filtrowani przez `@Profile` (bean nieaktywnego profilu dalej kandydatem); dopasowanie po nazwie pola/parametru (fallback Springa przy kilku kandydatach), `@Priority`, własne adnotacje qualifierów (meta-`@Qualifier`) — niemodelowane; rozwiązywanie wstrzyknięć pozostaje raportowe (bez krawędzi).
- **Stan F5:** z 11 przypadków pokryte 8 (złożony marker, parametry, XML, proxy, self-invocation, kontener przed kontrolerem, qualifier, dwa profile), częściowo 1 (wartości domyślne adnotacji — testy jednostkowe), odłożone 2 (konsumenci wiadomości, priorytety konfiguracji). Fakty frameworka wciąż poza v2 JSON.
- **Następny krok:** decyzja, czy zamknąć F5 z odłożonymi dwoma przypadkami, czy dodać fakty frameworka do v2 JSON; alternatywnie CVE:F8c.

### FRAMEWORKS:F5 — przyrost `f5-2`: wyrocznia JVM Spring MVC (kontener przed kontrolerem) i korzenie wywołań MVC — DONE (2026-11-22)

- **Wyrocznia (`testdata/mvc`, `scripts/mvc-oracle.sh`):** prawdziwy wbudowany Tomcat 9.0.45 z `DispatcherServlet` nad kontekstem `@EnableWebMvc` (Spring 5.3.6, Jackson 2.11.4; jary bez `META-INF/maven`, SHA-256 oryginałów w `deps/ORIGINALS.sha256`, jary rdzenia Springa z `testdata/tx/deps`), prawdziwe żądania HTTP: mapowanie zwracające `String`, zmienna ścieżki przez zarejestrowany `Converter` z wynikiem pisanym przez Jacksona, `@RequestBody` czytane przez Jacksona, wyjątek obsłużony przez `@RestControllerAdvice`/`@ExceptionHandler`, interceptor, `@ModelAttribute`, `@InitBinder`. Ujemne: publiczna metoda bez mapowania; kontroler spoza skanowania komponentów (JVM: 404). JVM wykonał 26 ciał aplikacji.
- **Wynik wyroczni przed poprawką:** graf przy `-roots profile` nie osiągał **14 z 26**: `@ExceptionHandler`, `@ModelAttribute`, `@InitBinder` (wołane jak handler przez `InvocableHandlerMethod` → `Method#invoke`), `WebConfig#addInterceptors`/`addFormatters` i tworzone w nich obiekty, `preHandle`/`postHandle`/`afterCompletion` interceptora, gettery zwracanego DTO, konstruktor i setter typu `@RequestBody`.
- **Poprawka:** profil `spring` — `roots.annotations` += `ExceptionHandler`, `ModelAttribute`, `InitBinder`; `roots.overrides` = wszystkie metody `WebMvcConfigurer` (18) i `HandlerInterceptor#preHandle`/`postHandle`/`afterCompletion`, `AsyncHandlerInterceptor#afterConcurrentHandlingStarted` (interceptor, którego nikt nie rejestruje, też jest korzeniem — jawna nadaproksymacja); nowy rodzaj korzenia `roots.web_bodies` (`profile.WebBodyRoots`, `internal/profile/webbodies.go`): gettery typu zwracanego handlera `@ResponseBody`/`@RestController`/`@RestControllerAdvice` oraz konstruktor bezargumentowy (gdy brak — wszystkie) i settery typu parametru `@RequestBody`, rekurencyjnie po typach aplikacji, z podtypami aplikacji dla typów abstrakcyjnych. Log `spring web body roots: …`. `profile.RulesVersion` `f5-1` → `f5-2` (wymuszone przez strażnika `f5-1`).
- **Wynik:** 26 z 26 ciał osiągalnych; `unmapped()` nie; `never()` (mapowanie nieżądane) i `GhostController` (poza skanowaniem) osiągalne celowo — profil nie czyta skanowania. Test sprawdza też dokładny zbiór korzeni ciał (2 gettery, konstruktor, setter).
- **Aplikacje referencyjne (`main` vs `f5-2`, offline, `-profile spring`, `-roots profile`/`all` × CHA/RTA):** **0 zmian werdyktów w 12 konfiguracjach**; `-roots all` bajtowo identyczne; przy `-roots profile` zmienione tylko ścieżki przykładowe: petclinic 3 (CVE-2022-22968, CVE-2024-38820, CVE-2025-22233 zaczynają się od `OwnerController#setAllowedFields` z `@InitBinder` zamiast `PetController#initOwnerBinder`), WebGoat 2 (CVE-2026-22737, CVE-2026-47884 od `MvcConfiguration#addViewControllers`, 12→7 i 13→8 kroków). Korzenie ciał: petclinic 1, WebGoat 35 (27 getterów, 2 konstruktory, 6 setterów), JPetStore 0.
- **Weryfikacja:** `TestMVCOracle` (przed poprawką 14 braków), `TestRulesVersion_tracksRuleSource`, `go test ./...`, `go vet ./...`. Dokumentacja: README, FAQ.md, PLAN.md (faza 7), SUPPORT.MD („Profiles and selected roots”), tabela kwalifikacji.
- **Ograniczenia:** typy elementów generycznych (`List<Item>`), dostęp do pól, `@JsonProperty` na prywatnych, buildery, mix-iny, `@JsonCreator` z argumentami (wszystkie konstruktory tylko gdy brak bezargumentowego); skanowanie komponentów nieczytane; własne `HandlerMethodArgumentResolver`/`ReturnValueHandler` tylko przez CHA; WebFlux poza zakresem.
- **Następny krok:** `f5-3` — qualifier/`@Primary` i dwa przebiegi `-active-profiles` w `testdata/di`.

### FRAMEWORKS:F5 — przyrost `f5-1`: wersja reguł korzeni i profili w snapshotcie — DONE (2026-11-22)

- **Zmiana:** `internal/profile/rules.go` — `profile.RulesVersion` (`f5-1`) dla reguł zamieniających profil w korzenie i fakty frameworka (rodzaje korzeni, `roots.beans` z wyborem konstruktora i cyklu życia, korzenie `web.xml` i XML Springa przez `internal/springxml`, katalog wejść, punkty wstrzyknięcia, fakty `@Value`) oraz `profile.BuiltinDigest()` — skrót SHA-256 (16 znaków) osadzonych YAML-i profili, aktualny bez ręcznego podbijania. Oba w `algorithms` snapshotu (`profile.rules`, `profile.builtin`); własny `-profile-file` jak dotąd hashowany jako wejście.
- **Strażnik:** `TestRulesVersion_tracksRuleSource` — skrót kodu `internal/profile` i `internal/springxml` (bez testów i komentarzy, przez `go/printer`) porównany z wartością zapisaną dla `RulesVersion`; komunikat mówi, co podbić. Sprawdzone: zmiana samego komentarza przechodzi, zmiana kodu w `springxml` pada. `TestBuiltinDigest_followsProfileContent` — 7 profili, skrót zmienia się ze zmianą YAML-a.
- **End-to-end:** `testdata/di` — `-snapshot-out` zawiera oba klucze, `-snapshot-in` odtwarza bajtowo identyczny `results.json`. petclinic (`-roots profile`): snapshot z `main` porównany przez `-snapshot-compare` — każde advisory podaje `algorithm profile.builtin: "" → "ae0a683eab383afb"; algorithm profile.rules: "" → "f5-1"`; wyniki v1 identyczne.
- **Weryfikacja:** `go test ./...`, `go vet ./...`. Dokumentacja: README (snapshoty), SUPPORT.MD („Data snapshots and replay”), FRAMEWORKS_IMPLEMENTATION_PLAN (F5).
- **Ograniczenia:** jak przy `b3-4` porównanie jest per przebieg — zmiana reguł korzeni unieważnia wszystkie advisories, nie tylko te, których ścieżki zaczynają się od zmienionego korzenia; skrót obejmuje wszystkie wbudowane profile, nie tylko aktywny. Fakty frameworka wciąż nie trafiają do v2 JSON.
- **Następny krok:** `f5-2` — wyrocznia JVM `DispatcherServlet` → kontrolery.

### FRAMEWORKS:F5 — ocena (2026-11-22) — 5 z 11 przypadków pokrytych wyroczniami B4, 3 częściowo, 3 brak; modele korzeni bez wersji w snapshotcie

- **Punkt wyjścia:** tabela kwalifikacji podawała dla „DI/adnotacje/konfiguracja” wyrocznię **brak** — nieaktualne od `b4-7`/`b4-8` (2026-11-15); wiersz poprawiony. Oceniono kryteria F5 względem `main` + `r4-2` bez zmian w kodzie.
- **Przypadki fixtures z F5 → dowód:**

  | Przypadek | Stan | Dowód |
  |---|---|---|
  | adnotacje z wartościami i defaults | częściowo | `@Bean(initMethod, destroyMethod)` w `testdata/di` (JVM); wywnioskowany domyślny `destroyMethod` i wartości domyślne elementów (`internal/annotation`) — tylko testy jednostkowe |
  | złożony marker | pokryty | `@Service` (meta-`@Component`) w `testdata/di`/`tx`/`events`; własne adnotacje złożone w `testdata/aop` |
  | parametry | pokryty | wstrzykiwanie przez konstruktor/setter/pole w `testdata/di`; wiązanie argumentów advice w `testdata/aop` |
  | XML | pokryty | `testdata/xmlbeans` (`b4-8`) |
  | priorytety konfiguracji | brak | `@Value`/źródła właściwości są faktem raportowym (F2 wycinek 6), nie dają krawędzi; aktywne profile tylko z `-active-profiles`, nie z `application.properties` |
  | qualifier | brak w wyroczni | rozwiązywanie punktów wstrzyknięcia (qualifier/`@Primary`/nazwa) jest raportowe — tylko testy jednostkowe |
  | dwa profile | częściowo | `testdata/di`: jeden profil nigdy nieaktywny (`Dormant`); brak porównania dwóch scenariuszy `-active-profiles` z JVM |
  | proxy | pokryty | `testdata/tx`, `aop`, `jpa`, `springdata`, `mybatis` |
  | self-invocation | pokryty | `testdata/aop` (`OrderService`, CGLIB) |
  | kontener przed kontrolerem | częściowo | `testdata/servletdispatch`: kontener → servlet/filtr; `DispatcherServlet` → metoda `@RequestMapping` bez wyroczni (handlery są korzeniami z adnotacji, `-roots profile`) |
  | konsument wiadomości | brak | `@JmsListener`/`@KafkaListener`/`@RabbitListener` jako korzenie w `spring.yaml`, bez fixture |

- **Kontrolowane uruchomienia beanów/endpointów:** tak dla DI, XML, AOP, zdarzeń, tx i proxy (jedna wersja: Spring 5.3.6, bez przenoszenia na inne wersje); nie dla endpointów MVC i listenerów wiadomości.
- **Raportowanie:** log i `report.md` podają punkty wstrzyknięcia (znalezione/rozwiązane), katalog wejść zewnętrznych (`ExternalEntries`) i korzenie cyklu życia per rodzaj; beany z `@Conditional`/`@Lazy`/`@Profile` bez listy liczone jako „mogą powstać”. **Brak w v2 JSON** — żadnego z tych faktów nie ma w wynikach maszynowych.
- **Snapshot:** scenariusz zapisany (`profile`, `roots`, `active-profiles` w `config`; własny plik profilu jako wejście `profile-file`). **Luka:** reguły wbudowanych profili (`internal/profile/profiles/*.yaml`) i modele korzeni/DI w Go (`roots.beans`, `web.xml`, XML Springa, `ExternalEntries`, wybór konstruktora) nie mają klucza w `algorithms` — zmiana reguły korzeni jest widoczna tylko przez tożsamość binarki, a `-snapshot-compare` nie przypisze jej przyczyny (`b3-4` wymusza wersje tylko dla rodzajów krawędzi, korzenie to nie krawędzie).
- **B4 i dokumentacja:** wiersz tabeli kwalifikacji poprawiony; SUPPORT.MD bez zmian (opisuje stan modelu, nie kwalifikację).
- **Proponowane przyrosty (w kolejności):**
  1. `f5-1` — wersja reguł korzeni/profili w snapshotcie (`profile.rules`: stała wersji + skrót osadzonych YAML-i), test na wzór `b3-4`. Mały, zamyka obowiązkowe kryterium.
  2. `f5-2` — wyrocznia JVM Spring MVC: prawdziwy `DispatcherServlet` (jary z petclinic) obsługujący żądania do kontrolerów z `@RequestMapping`/`@GetMapping`, metoda bez mapowania jako przypadek ujemny; sprawdza korzenie handlerów i ścieżkę kontener → `DispatcherServlet` → handler.
  3. `f5-3` — `testdata/di` rozszerzone o qualifier/`@Primary` i dwa przebiegi z różnymi `-active-profiles`.
  - Konsumenci wiadomości i priorytety konfiguracji — odłożone: wymagają brokera albo modelu źródeł właściwości, którego nie ma.
- **Następny krok:** `f5-1`.

### REFLECTION:R4 — lookupy odrzucane przez JVM i brakujące przypadki fixtures (`r4-2`) — DONE (2026-11-22); R4 DONE

- **Przypadki wyroczni JVM (`testdata/reflection`):** `lookupOnly` (samo `getMethod` bez `invoke`), `getMethodNotPublic` (`getMethod` metody pakietowej), `declaredInherited` (`getDeclaredMethod` metody odziedziczonej z nadklasy), `constructorNotPublic` (`getConstructor` konstruktora niepublicznego) — w JVM żadne ciało się nie wykonuje (`NoSuchMethodException`); `privateWithoutAccess` (`getDeclaredMethod` prywatnej bez `setAccessible`, JVM rzuca `IllegalAccessException`).
- **Defekt znaleziony przez wyrocznię:** trzy z nich dawały fałszywą krawędź `BasisReflection` — rozwiązanie celu ignorowało rozróżnienie `getMethod`/`getDeclaredMethod` (`Declared`) zapisane już w R1. **Reguła (`reflectinvoke.Version` `r4-1` → `r4-2`):** lookup sprawdzany na klasie, na której go wykonano, przed dyspozycją — `getDeclaredMethod` tylko metody deklarowane w tej klasie, `getMethod` tylko publiczne (także odziedziczone; `hierarchy.ResolvePublicByParams`), `getConstructor` tylko publiczne konstruktory; dowiedzione odrzucenie daje lukę z powodem zamiast krawędzi. Gdy metody nie widać w analizowanej hierarchii (klasa spoza wejścia), zachowanie bez zmian.
- **Jawna nadaproksymacja:** `setAccessible` nie jest śledzone, więc prywatna metoda wywołana bez niego zachowuje krawędź (nadmiarowa krawędź, nigdy pominięta) — test wyroczni dopuszcza tylko ją.
- **Testy:** `TestReflectionOracle` — bez poprawki 3 fałszywe krawędzie i 3 nieistniejące luki wyłapane; z poprawką 4 luki `Method#invoke` (w tym oba odrzucone lookupy) i 1 luka konstruktora.
- **Aplikacje referencyjne (`main` vs `r4-2`, offline, `-profile spring`, `-roots profile`/`all` × CHA/RTA):** krawędzie `reflection` identyczne (petclinic 46, WebGoat 47, JPetStore 8), **wyniki v1 bajtowo identyczne we wszystkich 12 konfiguracjach** — wzorce nie występują w tych aplikacjach; poprawka dotyczy precyzji, nie zmienia dziś werdyktów.
- **Weryfikacja:** `go test ./...`, `go vet`. Dokumentacja: README (siódmy wycinek refleksji), SUPPORT.MD (wiersz „Reflection”), REFLECTION_IMPLEMENTATION_PLAN (R4), tabela kwalifikacji.
- **Ograniczenia:** `setAccessible`, sprawdzanie dostępu zależne od wywołującego (np. metoda chroniona z innego pakietu), runtime-computed nazwy — niemodelowane. Reszta R3 (konfiguracja, `setAccessible`, `Constructor` przez parametr) NOT STARTED.
- **Następny krok:** REFLECTION R0–R2, R4 DONE; obszar otwarty tylko przez resztę R3 (bez zmierzonego zysku). Następny obszar wg propozycji: FRAMEWORKS F5 (wyrocznia DI) albo CVE F8c.

### REFLECTION:R4 — inicjalizacja przez `Class.forName` (`r4-1`) — DONE for this scope (2026-11-21)

- **Reguła (`internal/reflectinvoke/forname.go`, `reflectinvoke.Version` `r2-1` → `r4-1`):** wywołanie `Class.forName` z nazwą będącą dowiedzionym literałem klasy z wejścia inicjalizuje ją — `forName(String)` zawsze, `forName(String, boolean, ClassLoader)` gdy `boolean` nie jest stałą `false` (niestały: krawędź „może wykonać”, licznik). Krawędź `BasisReflection`/`CallInit` (jak `<clinit>` przy `newInstance`) do `<clinit>` klasy, a gdy jej brak — najbliższej nadklasy, która go ma (JVMS 5.5: nadklasa najpierw; dalszy łańcuch daje istniejąca krawędź G3 `<clinit>` → `<clinit>` nadklasy); interfejsy bez nadinterfejsów. `callgraph.ApplyReflectClassInitEdges`; log `Class.forName call sites found …`. Nazwa nieliteralna zostaje luką `reflection` jak dotąd.
- **Wyrocznia JVM (`testdata/reflection`, `scripts/reflection-oracle.sh`):** nowe przypadki `forNameInit` (`Inited#<clinit>`), `forNameSuper` (`InitedSub` bez inicjalizatora → `InitedBase#<clinit>`), `forNameMaybe` (`initialize` z parametru, w JVM `true`), `forNameNoInit` (`initialize=false` — JVM nie inicjalizuje `NotInited`); graf: dokładnie trzy nowe krawędzie do wykonanych ciał, `NotInited#<clinit>` nieosiągalny. `TestReflectionOracle` (mutacje: bez krawędzi; ignorowanie `initialize=false` — obie wyłapane), `TestFindForNameInits` (statystyki: 6 miejsc, 1 `initialize=false`, 2 bez inicjalizatora, 3 krawędzie, 1 niestały).
- **Aplikacje referencyjne (`main` vs `r4-1`, offline, `-roots profile`/`all` × CHA/RTA):** petclinic 186 literałów `forName` (5 `initialize=false`, 149 poza wejściem — JDK, 8 bez inicjalizatora), 24 inicjalizujące, 22 krawędzie dodane; WebGoat 232/6/178/19 → 29, 27 krawędzi; JPetStore 54/1/43/2 → 8, 8 krawędzi; żadne z `initialize` niestałym. **Wyniki v1 bajtowo identyczne we wszystkich 12 konfiguracjach** — żaden werdykt ani ścieżka przykładowa się nie zmienił.
- **Weryfikacja:** `go test ./...`, `go vet ./...`. Dokumentacja: README (szósty wycinek refleksji), SUPPORT.MD (wiersze „Implicit class initialization” i „Reflection”), REFLECTION_IMPLEMENTATION_PLAN (R4).
- **Ograniczenia:** `ClassLoader#loadClass` (bez inicjalizacji — poprawnie bez krawędzi), `MethodHandles.Lookup#findClass`/`ensureInitialized` i nazwy nieliteralne — niemodelowane; kolejność inicjalizacji nie jest modelowana (jak w G3).
- **Następny krok:** brakujące przypadki fixtures R4 (samo wyszukanie, niedostępna metoda) albo decyzja o selektywności reguły luk (wpis oceny R4 niżej).

### REFLECTION:R4 — ocena (2026-11-21) — kryteria w większości spełnione przez B0/B3/B4; jedna luka modelu, selektywność reguły luk poza zasięgiem R4

- **Punkty R4 względem stanu `main` (`c777c67`):**
  - *Fixtures:* `testdata/reflection` (wyrocznia JVM `b4-1`: literały `forName`/`getMethod`, wyszukanie na nadtypie, statyczna z `null`, prywatna `getDeclaredMethod`, `Method` w polu statycznym, `Constructor#newInstance`, niezgodna liczba argumentów, nazwa z wejścia jako luka) i `testdata/reflectparam` (`r3-1`/`r3-3`). Brakuje: `forName` z `initialize` true/false, samo wyszukanie bez wywołania jako przypadek ujemny wyroczni, niedostępna metoda.
  - *Integracja z `internal/reach` wg B0:* zrobiona przez `b0-1`…`b0-5` — miejsca refleksji to luki `reflection` (wspólny słownik `b0-3`), reguła luk oznacza `NOT_REACHABLE` jako `MODEL_INCOMPLETE`, v2 `gaps`/`gap_probe`/`unbounded_sites`, spis `coverage.reflection_census` (`r3-3`).
  - *Wersje modeli w snapshotach:* `reflection.invoke`, `reflection.parameters` w derywacji; `TestAlgorithmVersions_everyEdgeBasis` (`b3-4`) wymusza wersję dla `BasisReflection`.
  - *Kontrole B4 i pokrycie wg API:* `b4-1` (wyrocznia JVM), SUPPORT.MD wiersz „Reflection”.
- **Luka modelu znaleziona przy ocenie:** `Class.forName(literał)` (domyślnie i z `initialize=true`) nie daje krawędzi do `<clinit>` celu — SUPPORT.MD „Implicit class initialization” wymienia to jako niemodelowane. Pomiar (test roboczy): petclinic 460 miejsc `forName`, 186 z literałem (5 z `initialize=false`), 32 z celem w wejściu, **15 z `<clinit>`**; WebGoat 629/232/6/48/**21** (JAXB `Utils` → `ReflectionNavigator`, Tomcat JNI, `HdrHistogram` → `DatatypeConverter`, XStream `MXParser`, Ant, AspectJ…). Wpływ na werdykty niezmierzony.
- **Selektywność reguły luk (pytanie stojące za R4):** petclinic (`-roots` domyślne, CHA) — 613 istotnych nieograniczonych miejsc (refleksja 512, deserializacja 99, `MethodHandle` 2), reguła oznacza **29/29** `NOT_REACHABLE`. Spis 512 miejsc refleksji: nieznane 218 (198 — wynik wywołania wirtualnego), pole 206 (92 o wartości typowanej), wynik wywołania statycznego 45, parametr 35, inne 8; `convertible` 0. Nawet rozwiązanie wszystkich pól typowanych zostawia ponad 400 miejsc, więc żadne oznaczenie się nie zmieni (jak w `b0-5`). Selektywność wymaga zmiany samej reguły (trafność miejsca względem metody podatnej zamiast całej instancji) — decyzja projektowa, nie kryterium R4.
- **Stan R4:** oceniony; do zrobienia w zakresie R4 zostają wycinek `forName` → `<clinit>` (z przypadkami true/false w wyroczni) i brakujące przypadki fixtures. Kod bez zmian.

### DYNAMIC_LINKAGE:L2 — miejsca `MethodHandle` na aplikacjach referencyjnych (pomiar) — DONE dla pomiaru (2026-11-21); rozbudowa NIE podjęta (decyzja użytkownika)

- **Pytanie:** ile z wywołań `MethodHandle#invoke`/`invokeExact` (dawniej „0 z 10/46”) rozwiązuje się dziś, po `ResolveField`/`ResolveCallResult` (DATA_FLOW D3), i czy dalsze L2 coś zmieni.
- **Metoda (test roboczy, niecommitowany):** `methodhandle.FindInvokeSites` + `ResolveInvokes` na pełnym wejściu petclinic i WebGoat.
- **Wynik:** petclinic 0 z 10 (terracotta logging 4 — wynik wywołania wirtualnego; classgraph 3 — zbiór wartości; micrometer/Hazelcast 2 — parametr, `findVirtual` z nazwą nieznaną statycznie; hibernate-validator 1); WebGoat 2 z 46 (objenesis `DefineClassHelper$Java11`), reszta głównie JRuby (~35: pola `HandleMethod#target*`/`CompiledIRMethod#specific` z nieznanym zapisem, kombinatory `guardWithTest`/`foldArguments`/`insertArguments` — zakres L3, wartości sprzed handlera wyjątku), invokebinder, te same co petclinic.
- **Wpływ:** w `b0-5` `MethodHandle` to 2 z 611 istotnych nieograniczonych miejsc luk petclinic (refleksja 510, deserializacja 99) — pełne L2/L3 nie zmieniłoby żadnego oznaczenia reguły luk ani werdyktu. Decyzja: nie rozbudowywać; następny krok — ocena REFLECTION R4 (wpis wyżej, gdy gotowy).

### LIBRARY_MODE:LB2 — wartości dostarczone przekazane do metod zależności (pomiar) — DONE dla pomiaru (2026-11-21); model NIE budowany (decyzja użytkownika)

- **Pytanie:** czy rozszerzyć `lb2-7` o wartość dostarczoną przez wywołującego, przekazaną jako argument do metody zależności i wywoływaną tam (oraz zapisaną w polu obiektu zależności — przypadek (cc) `testdata/library-supplied`), i co by to zmieniło.
- **Metoda (test roboczy, niecommitowany):** dla każdego korzenia `-roots api` parametry typu zależności (spoza biblioteki i JDK, interfejs lub klasa nie-`final`) śledzone przez argumenty wywołań do metod zależności (`ResolveExact`, do 4 poziomów, `dataflow.AnalyzeWithParameterRefs`); liczone wywołania wirtualne/interfejsowe na tej wartości w zależności i zapisy jej do pól instancyjnych.
- **Wynik (miejsca wywołań / zapisy do pól):** commons-configuration2 0/0, jjwt-jackson 0/0, starter mybatis 2/1 (`SqlSessionTemplate#<init>` → `SqlSessionFactory#getConfiguration`), jetty-client 1/6, jackson-dataformat-yaml 107 (61 korzeni, głównie `JsonGenerator#_copyCurrentContents`)/37, fixture `library-supplied` 0/1 ((cc), `StringSubstitutor#setVariableResolver`).
- **Wpływ:** jedyni kandydaci to dwa niewykonane `REACHABLE` jackson-yaml (`lb4-2`): CVE-2026-68494, CVE-2026-89425 — po 53 wejścia, **wszystkie korzenie odziedziczone** z jackson-core/databind: 13 z parametrem `JsonParser` (wartość dostarczona, ale wprost na korzeniu odziedziczonym — `lb2-7` analizuje tylko klasy biblioteki; osobna luka), 3 `JsonParser#readValueAs*` (`this` przekazany jako argument — pozostałość `lb2-4`), 37 tworzy parser sam przez fabrykę z pola `_jsonFactory` (nic dostarczonego). Model nie dałby więc `SUPPLIED_DISPATCH_ONLY` żadnemu ustaleniu — tylko znaczniki części wejść. `-callgraph-mode pta` (i RTA) na jackson-yaml: oba nadal `REACHABLE`, 53 wejścia; werdykty 17/9 jak CHA.
- **Decyzja (2026-11-21):** nie budować (jak warunki `@ConditionalOn…`) — bez biblioteki, na której zmienia ustalenie. Kod bez zmian.

### LIBRARY_MODE:LB4 — krawędź przekazania asynchronicznego w jetty-client (pomiar) — DONE, ograniczenie zamknięte (2026-11-21)

- **Pytanie:** czy ograniczenie (1) z `lb4-1` — brak krawędzi od selektora NIO do `ReadCallback#succeeded`, werdykt tylko przez obcą drogę CHA — nadal obowiązuje po wycinkach ASYNC_CALLBACKS `a3-2`–`a4-1`, `b3-4` i `lb2-9`.
- **Przebieg:** `main` (`ddcc181`), jetty-client 9.4.36, `-roots api -profile generic -offline`, CHA i RTA, `-graph-out`; wyrocznia krawędzi `scripts/MethodEdgeTrace.java` na `fx.JettyGet` (`testdata/library-lb4/jetty-client`, wzorzec `org.eclipse.jetty.*`): 1853 pary, `status=200 body=hello`; `scripts/real-edge-paths.py`.
- **Wynik:** graf ma pełną prawdziwą ścieżkę `HttpClient#GET` → `HttpDestination#send` → resolver → `SelectorManager#connect` → `Connect#update` (`queued`, `a3-3`) → `processSelected` (`selected`, `a3-3`) → `processConnect` → `SelectorManager#execute` → `EatWhatYouKill#produce`/`#execute` → `ChannelEndPoint$1#run` (`a3-4`) → `FillInterest#fillable` → `AbstractConnection$ReadCallback#succeeded` → `HttpConnectionOverHTTP#onFillable` → `HttpReceiverOverHTTP#parse` → `HttpParser#convertContentLength` — każda krawędź wykonana przez JVM albo modelowana międzywątkowa z wykonanymi końcami. Po prawdziwych krawędziach: `convertContentLength` z 36, `HttpURI#parse` z 43 ze 103 wykonanych wejść API — jak po `a3-4`, w CHA i RTA. Krawędź `FillInterest#fillable` → `ReadCallback#succeeded` (callback w `AtomicReference`) jest w grafie przez CHA po interfejsie `Callback` Jetty (20 wywołujących `succeeded` w grafie) — trafna, ale nadaproksymowana; `AtomicReference` nadal niemodelowany.
- **Co zostaje (bez zmian, decyzje wcześniejsze):** zgłaszana ścieżka przykładowa to najkrótsza trasa BFS, nie wykonana (preferencja ścieżek odrzucona 2026-11-20, ROADMAP B0); werdykty bez zmian.
- **Dokumentacja:** LIBRARY_MODE_IMPLEMENTATION_PLAN (LB4 — punkt zamknięty), wpis `lb4-1` niżej (adnotacja), `testdata/library-lb4/EXPECTED.MD` (ograniczenie 1). Kod bez zmian.

### LIBRARY_MODE:LB2 — definicje beanów Micronauta jako ogłoszenia (`lb2-9`) — DONE for this scope (2026-11-20)

- **Zakres (zatwierdzony przez użytkownika 2026-11-20, „Tak, buduj”, po pomiarze `lb2-9m` niżej):** model ogłoszeń Micronauta, oczekiwanie zapisane w `testdata/library-micronaut/EXPECTED.MD` przed implementacją.
- **Reguła (`libapi.MicronautAnnouncements`, `micronautTargets`, `AnnouncedRoots`; `libapi.Version` `lb2-9`):** lista definicji z jara biblioteki — `META-INF/micronaut/io.micronaut.inject.BeanDefinitionReference/<klasa>` (3.x) albo `META-INF/services/io.micronaut.inject.BeanDefinitionReference` (1.x/2.x); klasa definicji = referencja bez `$Reference`. Korzenie: każda metoda biblioteki wołana **bezpośrednią** krawędzią z metod definicji i klas w niej zagnieżdżonych (`$Exec#dispatch`), jeśli jej klasa nie jest syntetyczna, nie jest samą definicją ani jej nadklasą; plus publiczne metody instancyjne takiej klasy (jak dla beanu CDI). Pochodzenie: wpis listy. `-api-exclude` działa jak dla innych ogłoszeń.
- **Weryfikacja:** `TestLibraryMicronaut` (CLI, CHA i RTA: 4 ogłoszone klasy, oba advisories `REACHABLE`, `exposed_api` dokładnie 4 metody, każda wykonana na JVM, `announced_by` z wpisu listy) — zastąpił `TestLibraryMicronaut_measuredGap`; `TestMicronautAnnouncements` (3.x, plik usług 2.x, jar bez listy), `TestMicronautRoots` (10 korzeni, bez `Expand`/`Unmanaged`/klas generowanych; `-api-exclude`). Regresja: 5 bibliotek kwalifikacyjnych × `-roots all|api` × CHA/RTA, offline — stara binarka (`b3-4`) vs nowa: 20/20 identyczne (ustalenia, statusy, `exposed_api`, liczba ogłoszeń). Prawdziwe jary Micronauta 3.9.7 (zależności demo, snakeyaml 1.30): micronaut-inject / -context / -http — 11 / 24 / 23 ogłoszone klasy, +0 / +11 / +9 metod wejścia, werdykty i `exposed_api` bez zmian (beany Micronauta są publiczne). Poprawka w trakcie: ręcznie pisane definicje w micronaut-inject wołały własną nadklasę — nadklasy definicji wyłączone (12 → 11 klas).
- **Ograniczenia:** prywatne metody beanów (Micronaut woła je refleksją) nie są znajdowane; układ Micronautu 4.x niesprawdzony; ścieżki przez kod samego Micronauta nie są oznaczane jako framework (inaczej niż Spring w `lb2-8`); dla jarów samego Micronautu definicje ręczne dokładają też klasy pomocnicze.
- **Dokumentacja:** README (tryb biblioteki), SUPPORT.MD, plan LB2, EXPECTED.MD fixture, ten rejestr.
- **Następny krok:** sprawdzenie układu 4.x (wymaga pobrania jarów 4.x); ewentualnie Micronaut jako framework dla ścieżek (`lb2-8` dla Micronauta).

### ROADMAP:B3 — wersje reguł każdego rodzaju krawędzi w derywacji (`b3-4`) — DONE for this scope (2026-11-20)

- **Zakres (punkt „invalidation zależnych wyników po zmianie modelu” z listy zatwierdzonej przez użytkownika 2026-11-20):** stan zastany — uczestnicy punktu stałego trafiają do snapshotu z wersjami (konfiguracja `models`, `id@wersja`), więc ich zmiana unieważnia wyniki; ale jednorazowe przebiegi poza punktem stałym nie miały wersji: bazowe reguły callbacków (A0–A2, a3-1…a3-3), MethodHandle (L2), refleksja R0–R2 (`Method#invoke`, `newInstance`), a także tabela interfejsów JDK (`jdkapi`, krawędzie obiektów i SPI) i reguły luk (`gaps`). Ich zmiana unieważniała wyniki tylko przez wersję analizatora — `-snapshot-compare` nie umiał podać przyczyny.
- **Zrealizowane:** `callback.Version` (`a3-3`), `methodhandle.Version` (`l2-1`), `reflectinvoke.Version` (`r2-1`); w `algorithmVersions`: `callback.resolver`, `methodhandle.edges`, `reflection.invoke`, `jdk.api`, `gaps.rules`. `TestAlgorithmVersions_everyEdgeBasis` czyta (go/parser) każdą stałą `EdgeBasis` z `internal/callgraph` i wymaga dla niej wersji (klucz algorytmu albo `config:models`) — nowy rodzaj krawędzi bez wersji nie przejdzie; mutacja (pusta wersja `methodhandle.edges`) wyłapana. `TestParticipantCatalog_versioned`: każdy uczestnik ma wersję w etykiecie.
- **Weryfikacja:** `go test ./...`, `go vet ./...`; petclinic (offline): stara binarka (`openjdk-catalog-curation`) vs nowa — v1 bajtowo identyczne; `-snapshot-compare` nowej względem snapshotu starej: 48 advisories, werdykty „unchanged”, przyczyny wymieniają 5 nowych kluczy (`callback.resolver "" → "a3-3"` …).
- **Ograniczenia:** unieważnienie nadal per przebieg, nie per model: zmiana dowolnej wersji unieważnia wszystkie advisories (bezpiecznie, ale grubo). Drobniejsze (tylko advisories, których ścieżki używają zmienionych krawędzi) wymaga ostrożnej semantyki — bramki twórców ucinają też krawędzie CHA, a wyniki negatywne zależą od wszystkich modeli — nie podjęte. Raporty-tylko (`reach.GapProbeVersion`, `UnboundedVersion`, `reflectinvoke.ProvenanceVersion`) poza derywacją, bo nie zmieniają werdyktów. Jednorazowe przebiegi nadal poza katalogiem uczestników (nie da się ich wyłączyć z profilu).
- **Dokumentacja:** README (sekcja snapshotów), SUPPORT.MD (wiersz „Data snapshots and replay”), ten rejestr.
- **Następny krok (wymaga zgody):** jednorazowe przebiegi do katalogu uczestników; albo invalidation per model, jeśli porównania snapshotów floty okażą się zbyt hałaśliwe.

### LIBRARY_MODE:LB2 — warunki `@ConditionalOn…` (pomiar) — DONE dla pomiaru (2026-11-20); model nie budowany

- **Pytanie:** czy znakowanie wejść autokonfiguracji ich warunkami (`@ConditionalOnClass` itd.) zmieniłoby ocenę ustaleń prawdziwego startera — np. ustalenia osiągalne tylko przez konfigurację warunkową na opcjonalnej zależności.
- **Pomiar (mybatis-spring-boot-autoconfigure 2.1.2, `work/libs/mybatis-starter`, `-roots api`, CHA, offline, cache `work/libs/cache`):** adnotacje warunków odczytane z bajtkodu (skrypt roboczy, poza repo): `MybatisAutoConfiguration` — `@ConditionalOnClass(SqlSessionFactory, SqlSessionFactoryBean)` (obie z zależności zadeklarowanych jako nieopcjonalne — przy użyciu startera zawsze spełnione), `@ConditionalOnSingleCandidate(DataSource)`, `@Bean` z `@ConditionalOnMissingBean`; klasa wewnętrzna `MapperScannerRegistrarNotFoundConfiguration` — `@ConditionalOnMissingBean`; `MybatisLanguageDriverAutoConfiguration` i 5 konfiguracji wewnętrznych — `@ConditionalOnClass` na opcjonalnych sterownikach FreeMarker/Velocity/Thymeleaf. Wynik: 25 `FRAMEWORK` + 1 `REACHABLE`; **każde** z 26 ma wejścia w publicznym API bez warunków (`VFS#list`, `SpringBootVFS`, `MybatisProperties`) oraz w `MybatisAutoConfiguration`; konfiguracje sterowników językowych nie są wejściem żadnego ustalenia (0 wejść). Znakowanie warunkami zmieniłoby 0 z 26.
- **Decyzja (w ramach zatwierdzonego „warunki `@ConditionalOn…`”):** modelu nie budujemy; wracamy przy bibliotece, na której ustalenie zależy wyłącznie od konfiguracji warunkowej na opcjonalnej/nieobecnej klasie. Kod i dokumentacja wsparcia bez zmian (plan LB2 zaktualizowany).

### LIBRARY_MODE:LB2 — pomiar Micronauta (`lb2-9m`) — DONE dla pomiaru (2026-11-20); model NIE zaczęty

- **Pytanie (plan LB2):** czy wygenerowane przez Micronaut definicje beanów są już publicznym API, tak że `-roots api` dochodzi do metod beanów bez osobnego modelu (jak `lb2-6` dla CDI).
- **Fixture:** `testdata/library-micronaut` — te same przypadki (t)–(x) co `testdata/library-cdi`, w Micronaucie 3.9.7: pakietowe `@Factory`/`@Singleton` producer, `@EventListener`, `@PostConstruct`, `@Inject`, klasa bez adnotacji beana; jedyny wywołujący `StringSubstitutor#replace` (commons-text 1.9) to pakietowe `Expand#expand`. Budowa z procesorem adnotacji (`scripts/JavaBuild.java` przyjmuje teraz `proc` → `-proc:full`; JDK 25 bez tego nie uruchamia procesorów). Wyrocznia JVM `scripts/library-micronaut-oracle.sh` (JDI, `observed.txt`): wykonane `Producers#banner`, `AuditListener#onStartup`, `CachedStore#warm`, `Wiring#init`, `Expand#expand`, `replace`; `Unmanaged` nigdy. Micronaut pobrany raz przez `mvn dependency:copy-dependencies` do `work/micronaut-oracle/lib` (niecommitowany).
- **Wynik:** **nie** — `$X$Definition` (wołające `build`/`injectBean`/`initialize`, `$Exec#dispatch` dla `@EventListener`) są pakietowe i syntetyczne; publiczne `$X$Definition$Reference` też syntetyczne i tylko tworzą definicję; ogłoszenie w `META-INF/micronaut/io.micronaut.inject.BeanDefinitionReference/<klasa>` (3.x; 2.x: `META-INF/services/…`). `-roots api` (CHA, RTA; profil `generic` i wykryty `micronaut`; z jarami Micronauta w `-deps` i bez): CVE-2022-42889 i CVE-2025-48924 `NOT_REACHABLE`, `exposed_api` puste, `coverage.api` 4 klasy / 5 metod / 0 ogłoszonych — **fałszywy negatyw**. `-roots all`: `REACHABLE` (każda metoda korzeniem, także `Unmanaged`).
- **Utrwalone:** `TestLibraryMicronaut_measuredGap` (sprawdza `observed.txt` i obecny wynik; do odwrócenia, gdy model powstanie); EXPECTED.MD z oczekiwaniem dla modelu. Dokumentacja: README (tryb biblioteki: Micronaut nieobsługiwany, sprawdzać `-roots all`), SUPPORT.MD (wiersz Library mode), plan LB2.
- **Następny krok (wymaga zgody użytkownika):** model ogłoszeń Micronauta — każda definicja z listy `BeanDefinitionReference` → korzenie `build`/`injectBean`/`initialize`/`Exec#dispatch` z `announced_by`; oczekiwanie w EXPECTED.MD (4 metody w `exposed_api`, `Unmanaged` nigdy). Do sprawdzenia przy tym: układ Micronautu 4.x.

### ROADMAP:B0 — preferencja ścieżek przykładowych (pomiar) — NIE REALIZOWANE, decyzja 2026-11-20

- **Pytanie:** czy inna reguła wyboru ścieżki przykładowej (zamiast najkrótszej) da więcej ścieżek, które naprawdę się wykonują.
- **Pomiar (2026-11-20, prototyp w Pythonie, poza repo):** jetty-client, zapisany graf, krawędzie wzięte przez JVM (`scripts/MethodEdgeTrace.java`, logika jak `scripts/path-divergence.py`). Zgłaszane dziś ścieżki: 18 z 408 prawdziwych krawędź po krawędzi. Najkrótsza: 14–27; „najmniej kroków CHA, potem długość”: 18–28 — rozrzut zależy wyłącznie od kolejności rozstrzygania remisów; ważenie rozgałęzieniem CHA nie lepsze; górna granica (prawdziwa ścieżka istnieje w grafie): 81. RTA i `-callgraph-mode pta` dają te same liczby. Jedyny stały efekt „najmniej CHA”: −26 % kroków CHA, +2 średniej długości.
- **Dlaczego heurystyka nie pomoże:** brakujące ścieżki zależą od danych (resolver Sync vs Async ustawiany publicznym setterem, callback `FillInterest`, ścieżki błędów `abort`) — w trybie biblioteki legalnie możliwe, bo te obiekty może dostarczyć klient. Odzyskać je mogłaby tylko wiedza o typach odbiorców, nie kolejność przeszukiwania.
- **Decyzja (użytkownik, 2026-11-20, rekomendacja „zrezygnować”):** reguły nie zmieniamy; ścieżka przykładowa pozostaje najkrótszą ścieżką z dowodem kroków (B0 `b0-2`: bez kroków tylko-CHA, gdy istnieje). Nic nie zmienione w kodzie ani dokumentacji wsparcia.
- **Gdyby wracać:** tylko z modelem typów odbiorców (np. listenery przez parametr — jetty `ResponseNotifier`), nie z heurystyką kolejności.


### ASYNC_CALLBACKS:A4 `a4-2` — kwalifikacja obszaru — DONE for this scope (2026-11-20)

- **Wymaganie (plan A4):** fixtures dla direct executor, thread override, task object, lambda, constructor, dwa zadania, rejection, future success/failure, listener i wybrany pipeline reactive; porównanie celów z kontrolowanymi uruchomieniami bez wnioskowania o niemożliwości niezaobserwowanych scenariuszy; liczba dodanych krawędzi i błędnych skojarzeń; B4 i snapshoty; udokumentowane ograniczenia współbieżności.
- **Kwalifikacja (`cmd/ariadne/async_qualify_test.go`, `testdata/async-qualify/RESULTS.MD`):** `TestAsyncQualification` uruchamia pełny pipeline (`-profile generic -roots profile`, CHA) na dziewięciu fixtures async (`async`, `continuations`, `async-forward`, `async-handoff`, `async-returned`, `async-listeners`, `async-reactive`, `async-completion`, `async-threads` — razem cała lista z planu) i porównuje z ich wyroczniami JVM: 166 wykonanych ciał, 1 nieosiągnięte (`MethodHandle` z nazwą wybieraną w czasie wykonania — luka DYNAMIC_LINKAGE, nie tego obszaru); 60 krawędzi `callback` z osiągniętych metod, 6 do ciał niezapisanych przez wyrocznie: 2 wykonane bez znacznika (wyrocznia `continuations` zapisuje tylko ciała z `Ran.here()`), 4 udokumentowane „może wykonać” (`then*`/`exceptionally` na etapach `supplyAsync`/`completedFuture`, zadanie zwracane bramkowane metodą osiąganą przez CHA mimo braku instancji, `onError` subskrybenta, którego `onNext` nie rzucił) — **żadne błędne skojarzenie**. `TestAsyncTwoTasks` (scenariusz 3): w żadnym fixture miejsce rejestracji z bezpośrednim `new`/lambdą nie rozwiązuje się do więcej niż jednego celu. Liczby oczekiwane w teście (regresja); `ARIADNE_WRITE_ASYNC_RESULTS=1` odtwarza tabelę.
- **Pomiar krawędzi na wejściach rzeczywistych (sumarycznie, `a3-2`…`a4-1`):** każdy wycinek mierzony osobno na 5 bibliotekach (20 przebiegów) i 3 aplikacjach (12 konfiguracji, `-graph-compare`) — wpisy powyżej; werdykty i ścieżki przykładowe aplikacji nie zmieniły się w żadnym wycinku `a3-3`…`a4-1`; jetty-client: `convertContentLength` po prawdziwych krawędziach z 3 do 36 ze 103 wykonanych wejść API (`a3-4`).
- **Snapshoty (B4):** modele bramkowane `async-callbacks:{returned-tasks,listeners,reactive,completion,rejection}` z wersjami są w konfiguracji snapshotu (`models`) i w v2 `callgraph.participants` — porównanie snapshotów nazywa zmianę modelu; reguły zwykłych krawędzi (A2, `a3-2`, `a3-3`, `start()` z `a4-1`) widać tylko jako różnice krawędzi.
- **Ograniczenia współbieżności (README, „What an asynchronous edge means”):** krawędź asynchroniczna to relacja „może wykonać”, nie stos jednego wątku ani porządek czasowy; tożsamość wątku, przeplecenia, polityka schedulera i anulowania, żądanie w reactive streams i widoczność pamięci — niemodelowane; nieznana możliwość anulowania nie dowodzi, że zadanie się nie wykona.
- **Następny krok:** obszar ASYNC_CALLBACKS zamknięty i zarchiwizowany 2026-11-20 (`archive/2026-11-20/async_callbacks/COMPLETION.md`) po scaleniu łańcucha PR #109–#116; kandydaci poza zakresem: Reactor/RxJava (po dostępie do artefaktów), listy listenerów przez parametr (jetty `ResponseNotifier`), preferowanie krawędzi modelowanych w ścieżkach przykładowych.

### ASYNC_CALLBACKS:A4 `a4-1` — podklasy `Thread` i odrzucone wykonanie — DONE for this scope (2026-11-20)

- **Wymaganie (plan A4, lista fixtures):** „thread override” i „rejection”. Wzorzec „podklasa `Thread` nadpisuje `run()`, bez `Runnable`” był udokumentowaną luką A0 (katalog zna tylko konstruktory z `Runnable`; `Thread#start` to metoda JDK, której CHA nie rozwija).
- **Reguła (`internal/callback/threads.go`):** `start()` na podklasie `Thread` z wejścia (odbiorca rozwiązany przez `receiverObjs` do `new`, inaczej zadeklarowana podklasa i jej podklasy) → zwykła krawędź (warunek `started`) do `run()`, które ta klasa wykonuje, gdy to `run()` z wejścia (nie `Thread#run` — to A2). `RejectedExecutionHandler` przekazany `ThreadPoolExecutor` (konstruktor albo `setRejectedExecutionHandler`) → `execute`/`submit` na tym samym executorze (miejsce `new`) dostają krawędź do `rejectedExecution` (warunek `rejected`, bramka `async-callbacks:rejection@a4-1` na metodzie rejestrującej). Saturacja puli nierozstrzygana; zadanie odrzucone zachowuje krawędź A2.
- **Weryfikacja:** fixture `testdata/async-threads` (wyrocznia JVM `scripts/async-threads-oracle.sh`): uruchomiona podklasa; podklasa w zmiennej nadklasy — tylko jej własne `run()`; nieuruchomiona — nic; jednowątkowa pula bez kolejki, której drugie zadanie jest odrzucane — handler wykonany i osiągnięty; handler rejestrowany tylko w niewołanym kodzie — nic. Osiągalność zgodna z JVM. `TestThreadModels` (mutacja: bez rozwiązywania odbiorcy — wyłapana).
- **Pomiar (`a3-7` vs `a4-1`):** biblioteki — werdykty i `exposed_api` identyczne w 20 przebiegach (starter mybatis: 7 krawędzi `start()`); aplikacje (12 konfiguracji) — petclinic/WebGoat/JPetStore +28/+26/+14 krawędzi (podklasy `Thread` w bibliotekach uruchamiane w osiągalnym kodzie: 63/70/24 wywołań `start()`, 28/34/14 celów `run()`; jedna krawędź odrzucenia w petclinic), WebGoat `-roots profile` RTA +112 (nowo osiągnięte ciała `run()` czynią 6 klas więcej żywymi — RTA usuwa 86 krawędzi CHA mniej), 0 usuniętych; **żaden werdykt ani ścieżka przykładowa nie zmienione**.
- **Ograniczenia:** `start()` na odbiorcy typu `Thread` bez rozwiązanego obiektu — pominięte (licznik); saturacja, polityki JDK (`AbortPolicy`, `CallerRunsPolicy`…) i `ScheduledThreadPoolExecutor` — niemodelowane osobno.

### ASYNC_CALLBACKS:A3 wycinek 7 — zakończenie ręcznie kończonych `CompletableFuture` (`a3-7`) — DONE for this scope (2026-11-20)

- **Wymaganie (ASYNC_CALLBACKS.MD, scenariusz 4; plan A3):** operacja future zależna od sukcesu/porażki ma odpowiedni warunek i analiza nie udaje, że każda kontynuacja wykonuje się bezwarunkowo. Wycinek 1 zostawiał każdej kontynuacji krawędź „może wykonać” niezależnie od losu etapu.
- **Reguła (`internal/callback/completion.go`, bramka `async-callbacks:completion@a3-7`):** dla krawędzi kontynuacji A2, których etap rozwiązuje się (`receiverObjs`) wyłącznie do `new CompletableFuture<>()`: `then*`/`*Either`/`thenCompose` wymagają zakończenia normalnego (`complete`, `completeAsync`, `obtrudeValue`, `completeOnTimeout`), `exceptionally*` — wyjątkowego (`completeExceptionally`, `obtrudeException`, `cancel`, `orTimeout`), `handle`/`whenComplete` — dowolnego. Wywołania kończące parowane po miejscu `new` (lokalna, wywołujący dla parametru, zapisy pola private/final); kończące z nierozwiązanym odbiorcą liczą się dla każdego ręcznego future (nigdy powód do usunięcia). Krawędź z miejsca rejestracji (jak dotąd) bramkowana metodami wywołań kończących; brak zakończenia właściwego rodzaju — krawędź usunięta (licznik). Pozostałe kontynuacje (etapy `supplyAsync`/`runAsync`, inne kontynuacje, biblioteki, nierozwiązane) — bez zmian. Lambdy nadal osiągalne krawędzią utworzenia (DYNAMIC_LINKAGE); decyzja dotyczy krawędzi `callback`, a osiągalności — dla instancji klas.
- **Weryfikacja:** fixture `testdata/async-completion` (wyrocznia JVM `scripts/async-completion-oracle.sh`): zakończony normalnie — `thenApply` tak, `exceptionally` nie; nigdy niezakończony — nie; zakończony błędem — `exceptionally` i `whenComplete` tak, `thenApply` nie; kończony przez metodę, której jest przekazany — tak (bramka na niej); future w polu final kończony tylko przez niewołaną metodę — nie (odcięty bramką); `runAsync` — bez zmian. Osiągalność zgodna z JVM w każdym przypadku. `TestCompletion` (mutacja: bez rozróżniania rodzaju zakończenia — wyłapana). `testdata/continuations`: kontynuacja nigdy niezakończonego future straciła krawędź (na JVM się nie wykonała) — `TestContinuationsOracle` zaktualizowany.
- **Pomiar (`a3-6` vs `a3-7`):** żadne wejście nie rejestruje kontynuacji na ręcznie kończonym future (petclinic i WebGoat po 8 krawędzi kontynuacji, wszystkie na innych etapach) — grafy, werdykty i ścieżki identyczne we wszystkich 32 przebiegach.
- **Ograniczenia:** zakończenie etapów pochodnych (`supplyAsync`, kontynuacje) nierozstrzygane; drugi etap `thenCombine`/`*Either` pomijany (bramka tylko na odbiorcy); future przekazany przez kolekcję, wynik metody lub pole nie-final — nierozwiązany (krawędź bez zmian); podklasy `CompletableFuture` — bez zmian.

### ASYNC_CALLBACKS:A3 wycinek 6 — reactive: JDK `Flow`/`SubmissionPublisher` (`a3-6`) — DONE for this scope (2026-11-20)

- **Wymaganie (plan A3):** jedna biblioteka reactive i wersja; warunki subskrypcji i wykonania wybranego katalogu operatorów; nieaktywny pipeline i znana subskrypcja rozróżnione. Wybór: JDK `java.util.concurrent.Flow` z `SubmissionPublisher` (Java 9+, sprawdzane na JDK 25) — Reactora/RxJavy nie ma w lokalnym repozytorium Mavena, a Maven Central blokuje pobieranie (2026-11-18).
- **Reguła (`internal/callback/reactive.go`, bramka `async-callbacks:reactive@a3-6`):** `subscribe(s)` → `s.onSubscribe` (warunek `subscribed`); `submit`/`offer` na tym samym wydawcy → `s.onNext` (`published`) i `s.onError` (`exceptional` — wyjątek w `onNext` kończy subskrypcję przez `onError`, potwierdzone na JVM); `close()` → `onComplete` (`completed`); `closeExceptionally` → `onError`; `consume(c)` → `c.accept` przy `submit`/`offer`. Wydawca identyfikowany obiektami odbiorcy (`receiverObjs`: miejsce `new`, `this` w podklasie `SubmissionPublisher` — procesor republikujący przez `submit`, wywołujący dla parametru, zapisy pola private/final); subskrypcja i emisja parują, gdy dzielą obiekt; nierozwiązany wydawca — luka (licznik), nigdy „każdy wydawca”. Bramka: metoda subskrybująca. `receiverObjs` (wspólne z `a3-4`) zna teraz `this`.
- **Weryfikacja:** fixture `testdata/async-reactive` (wyrocznia JVM `scripts/async-reactive-oracle.sh`, prawdziwi wydawcy na wspólnej puli): subskrybent, łańcuch przez procesor (`Upper#onNext` → `submit` → `Sink#onNext`), lambda `consume`, `onError` subskrybenta rzucającego w `onNext` — wykonane i osiągnięte; wydawca zasubskrybowany i zamknięty bez emisji — brak `onNext`; `closeExceptionally` — tylko `onError`; subskrybent niezasubskrybowany i zasubskrybowany w niewołanym kodzie (wydawca jako parametr bez wywołujących) — nic. Osiągalność po bramce zgodna z JVM dla każdego subskrybenta. `TestReactive` (mutacja: parowanie wydawców po samym typie — wyłapana).
- **Pomiar (`a3-5` vs `a3-6`):** żadne z wejść (5 bibliotek, petclinic, WebGoat, JPetStore) nie używa `Flow` — 0 subskrypcji; grafy, werdykty i ścieżki identyczne we wszystkich 32 przebiegach. Model sprawdzony tylko na fixture.
- **Ograniczenia:** inne biblioteki reactive (Reactor/RxJava — osobny przyrost po dostępie do artefaktów), implementacje `Flow.Publisher` w wejściu (ich `subscribe` to zwykły kod wołający subskrybenta — parametr typu JDK), żądanie (`request`) i anulowanie, wątek wykonawcy.

### ASYNC_CALLBACKS:A3 wycinek 5 — listenery (`a3-5`) — DONE for this scope (2026-11-20)

- **Wymaganie (plan A3):** rejestracja listenera powiązana z modelem źródła zdarzenia; samo `addListener` nie tworzy zdarzenia. `a3-3` celowo pomijało obejście listy bez czyszczenia („rozgłoszenie, zakres A3”).
- **Reguła (`internal/callback/listeners.go`, bramka `async-callbacks:listeners@a3-5`):** dwa modele źródła: (1) lista w polu — magazyn `a3-3` do kolekcji JDK w polu + *rozgłoszenie*: metoda obchodząca elementy tego pola (iteracja, `List#get(i)`, `peek`) bez zdejmowania i czyszczenia, wołająca metodę na elemencie; (2) `PropertyChangeSupport` w polu — `addPropertyChangeListener([nazwa,] l)` + `firePropertyChange`/`fireIndexedPropertyChange` na tym samym polu. Krawędź „może wykonać” (warunek `notified`) z wywołania powiadamiającego do metody listenera, bramkowana metodą, w której listener wchodzi (rejestrująca dla `new`/lambdy, wywołująca dla parametru — forwarder `a3-2`). Skan wspólny z `a3-3` (`ScanHandoffs`). DATA_FLOW: `List#get(int)` śledzone jak pobranie z kolekcji (w `a3-3` jak odczyt bez zdjęcia — wymaga `clear`).
- **Weryfikacja:** fixture `testdata/async-listeners` (wyrocznia JVM `scripts/async-listeners-oracle.sh`): listenery `Consumer` (typ JDK, CHA milczy) rozgłaszane w drugim wątku, pętla `get(i)` po `CopyOnWriteArrayList` własnego interfejsu (krawędź CHA już jest), prawdziwy `PropertyChangeSupport`; osiągalność po bramce zgodna z JVM dla każdego listenera — zarejestrowane i powiadomione: tak; zarejestrowany tylko w niewołanym kodzie (bramka), na źródle, które nie powiadamia, utworzony bez rejestracji, na beanie bez `fire`: nie. `TestListeners` (mutacja: bramka na metodzie rozgłaszającej przepuszcza niewołaną rejestrację — wyłapana).
- **Biblioteki:** werdykty i `exposed_api` identyczne we wszystkich 20 przebiegach; krawędzie listenerów: jjwt-jackson (databind) 13, pozostałe 0 — jetty-client obchodzi listenery odpowiedzi przez parametr (`ResponseNotifier#notify*(List, …)`), nie pole.
- **Aplikacje referencyjne (`a3-4` vs `a3-5`, 12 konfiguracji):** petclinic/WebGoat/JPetStore +14/+12/+13 krawędzi `callback` (listenery 10/5/2, drenaże indeksowane `a3-3` 1/2/11 — np. MyBatis `BatchExecutor`), 0 usuniętych; kandydatów listenerów 650/143/156, z czego CHA ma już 452/99/112, bramka odcina 187/38/42; **żaden werdykt ani ścieżka przykładowa nie zmienione**. Rozgłoszenie to każde obejście elementów pola, nie tylko listenerów (Spring `ExtendedBeanInfo` → deskryptory właściwości — realne „może wykonać” na typie JDK).
- **Ograniczenia:** tablice listenerów (copy-on-write `T[]`), lista obchodzona przez parametr lub wynik gettera, `forEach` z lambdą, `java.util.Observable`, usunięcie listenera, nazwa właściwości, na którą czeka listener — niemodelowane.

### ASYNC_CALLBACKS:A3 wycinek 4 — zadania zwracane z metody (`a3-4`) — DONE for this scope (2026-11-20)

- **Problem (po `a3-3`, jetty-client):** `EatWhatYouKill` uruchamia (`runTask`: `task.run()`, albo `execute(task)` na puli) to, co zwraca `ExecutionStrategy.Producer#produce`; `SelectorProducer#produce` zwraca wynik `processSelected`, a ten — `Selectable#onSelected`; `ChannelEndPoint#onSelected` zwraca jedno ze swoich pól final (`_runFillable`… — klasy anonimowe tworzone w konstruktorze). Wywołanie `Runnable#run` nie jest rozwijane przez CHA (typ JDK), a wynik wywołania instancyjnego był dla DATA_FLOW nieznany — brakowało ostatniej krawędzi na prawdziwej ścieżce do `convertContentLength`.
- **Reguła (`internal/callback/returned.go`, bramka `async-callbacks:returned-tasks@a3-4`):** miejsca: `Runnable#run`/`Callable#call` i skatalogowane rejestracje A2. Zadanie rozwiązywane: parametr — u wywołujących; wynik wywołania — wartości zwracane metody (cel dokładny dla static/special; dla wirtualnego/interfejsowego zwracającego `Runnable`/`Callable` — DATA_FLOW śledzi teraz te wyniki jako `CallTrace` — ciała, które wykonają klasy odbiorcy, gdy są znane: `new`, parametr, pole private/final z zapisami jego klasy i parametrami konstruktora u wywołujących konstruktor; inaczej wszystkie implementacje, do 16); pole final/static — zapisy (`ResolveFieldOneHop`); `new`/lambda — zadanie. Tylko zadania, które przeszły przez `return`; zwracany parametr nie jest śledzony. Krawędź „może wykonać” (warunek `produced`) z miejsca do `run`/`call` klasy albo ciała lambdy, bramkowana (creator gate, punkt stały B3) najgłębszą metodą zwracającą zadanie. Limity: 8 ogniw łańcucha, 256 analiz metod na miejsce. Log `returned tasks (ASYNC_CALLBACKS a3-4): …`.
- **Precyzja:** wersja bez zawężania odbiorcy dawała krawędzie z `ProduceConsume`/`ExecuteProduceConsume`/`ProduceExecuteConsume#produce` (osiągalne przez CHA na `ExecutionStrategy#produce`, w trybie biblioteki „tworzone” przez klienta) do zadań `ChannelEndPoint` — 23 kandydatów; ich pole `_producer` ustawia tylko konstruktor, którego nic nie woła → 11.
- **Weryfikacja:** fixture `testdata/async-returned` (wyrocznia JVM `scripts/async-returned-oracle.sh`: pętla strategii w drugim wątku, prawdziwy executor): `Strategy#runTask` → `Conn$1#run` i lambda `write` (wykonane), → `IdleTask#run` bramkowane na `Idle#onSelected` (nic nie tworzy `Idle`; z korzeni bez metody zwracającej nic nie dochodzi); `Main#main` (`submit`) → lambda `Callable` z `Factory#job` (wykonana); `CloseTask` zwracany do kodu, który go nie uruchamia — brak; `OtherStrategy` (jak `ProduceConsume`, nigdy nie tworzona) — brak (mutacja: bez zawężania krawędzie są — wyłapana). `TestReturnedTasks`.
- **jetty-client (`-roots api`):** 11 krawędzi: `EatWhatYouKill#runTask`/`#execute`, `Invocable#invokeNonBlocking` → `ChannelEndPoint$1/$2/$3#run`, `SslConnection$DecryptedEndPoint#needsFillInterest` → `SslConnection$1#run`, `ReservedThread#run` → `ReservedThreadExecutor$1#run`. Wyrocznia krawędzi odtworzona (`MethodEdgeTrace`, 1853 pary): do `convertContentLength` po prawdziwych krawędziach dochodzi **36 ze 103** wykonanych wejść API (3 przed; `real-edge-paths.py` przyjmuje teraz też krawędź `sam` — lambda przekazana executorowi — o wykonanych końcach; bazowe 3 bez zmian), `HttpClient#GET` → `HttpDestination#send` → resolver → `Connect#update` → `processSelected` → `EatWhatYouKill#produce` → `ChannelEndPoint$1#run` → `FillInterest#fillable` → `onFillable` → parser; do `HttpURI#parse` 43 bez zmian. Werdykty i `exposed_api` bez zmian (CHA, RTA). Ścieżki raportowane idą teraz tym mechanizmem (przez `runTask`, gdy ten przebieg użył `execute`): wykonane w całości po metodach 66 → 40 z 408, prawdziwe krawędź po krawędzi 18 → 18 — utracone 26 pasowały tylko po metodach, przez skoki CHA, których Jetty nie robi.
- **Biblioteki (`-offline`):** werdykty i `exposed_api` identyczne we wszystkich 20 przebiegach; krawędzie: jetty 11, starter 6, pozostałe 0.
- **Aplikacje referencyjne (`a3-3` vs `a3-4`, 12 konfiguracji, `-graph-compare`):** petclinic/WebGoat/JPetStore 40/55/4 z 541/607/123 miejsc rozwiązanych, +63/+80/+4 krawędzi `callback` (6/7/0 kandydatów odciętych bramką), 0 usuniętych; **żaden werdykt ani ścieżka przykładowa nie zmienione**.
- **Ograniczenia:** zwracany parametr (`return task;` w metodzie pomocniczej) nie jest śledzony; pola nie-final i nie-private — nieznane (licznik „fields unresolved”: 26/33/6 na aplikacjach); `AtomicReference` (`FillInterest`) i drenaż przez metodę pomocniczą (`QueuedThreadPool#idleJobPoll`) — jak po `a3-3`; rozwiązywanie niekontekstowe (miejsce dostaje to, co zwraca każda implementacja, którą może wywołać); BFS raportuje najkrótszą ścieżkę, nie wykonaną.
- **Następny krok:** reszta A3 (listenery, reactive) albo plan CI floty (od 2026-11-23); dla ścieżek raportowanych — preferowanie krawędzi modelowanych przed szerokim CHA (osobny temat).

### ASYNC_CALLBACKS:A3 wycinek 3 — przekazania przez kolejkę i klucz selektora (`a3-3`) — DONE for this scope (2026-11-19)

- **Problem (po `a3-2`, jetty-client):** zgłoszenie połączenia z `HttpClient#GET` idzie przez własną kolejkę Jetty — `ManagedSelector#submit(update)` → `_updates.offer(update)`, a wątek selektora (`SelectorProducer#processUpdates`) zamienia dwie kolejki, iteruje, woła `update.update(selector)` i czyści — oraz przez załącznik klucza NIO (`Connect#update`: `channel.register(selector, OP_CONNECT, this)`, `processSelected`: `key.attachment()` → `processConnect`). Żadne z nich nie jest wywołaniem ani `Executor`em.
- **Reguła (`internal/callback/handoff.go`):** *kolejka* — magazyn: `add`/`offer`/`push`/`put`… na kolekcji JDK z pola (także przez akcesor `access$NNN`); drenaż: element zdjęty z kolekcji z pola (`poll`/`take`/`pop`…) albo odczytany (iteracja, `peek`) z polem czyszczonym w tej samej metodzie (`clear`), z wywołaniem metody na elemencie — iteracja bez czyszczenia to rozgłoszenie do listenerów (zakres A3), nie drenaż. Para: to samo pole albo dwa pola kolekcji tej samej klasy przy znanym typie elementu zgodnym z wołaną metodą (podwójne buforowanie). Krawędź od miejsca, gdzie obiekt wchodzi (magazyn dla `new`/lambdy, a dla parametru — wywołujący z `a3-2`, `Origin`), do metody, którą wywołanie drenażu uruchamia na tym obiekcie; warunek `queued`. *Klucz NIO* — `register(Selector, int, Object)`/`attach` → metody tego samego zagnieżdżenia klas czytające `attachment()` i sprawdzające gotowość (`isReadable`…/`readyOps`, albo same biorące `selectedKeys()`); warunek `selected`. Do 8 drenaży na magazyn. Dataflow: wyniki `Collection#iterator` i pobrań z kolejek są śledzone (`CallTrace`, `dataflow.IsCollectionRetrieval`); `Iterator#next`, którego iterator zgubił handler wyjątku w pętli, przypisany do jedynego iterowanego pola metody. Log `hand-offs (ASYNC_CALLBACKS a3-3): …`.
- **Precyzja:** pierwsza wersja (krawędź magazyn → metoda drenażu) dawała `ManagedSelector#dump` (wkłada `DumpKeys`) dostęp przez CHA do `Connect#update` i dalej do kodu podatnego — 18 fałszywie narażonych metod API (`dump`, `sweep`); krawędź od miejsca wejścia do metody elementu usuwa je (`dump` → tylko `DumpKeys#update`).
- **Weryfikacja:** fixture `testdata/async-handoff` (wyrocznia JVM `scripts/async-handoff-oracle.sh`: drenaże w drugim wątku, prawdziwy `Pipe`/`Selector`): `Loop` (`poll`), `Swap` (podwójne buforowanie, iteracja z `catch` w pętli + `clear`), lambda, `Api#other` → `C#work` tylko od `Api#other` (nigdy nie wywołane — `C#work` się nie wykonał), `Listeners` (lista bez czyszczenia, nigdy nie odpalona) — brak krawędzi, `Unused` — brak, `Sel#register` → `Sel#loop`, nie `Sel#dumpKeys` (obchodzi wszystkie klucze; wykonał się). `TestHandoffs` (mutacja: iteracja bez czyszczenia jako drenaż — wyłapana).
- **jetty-client (`-roots api`):** 15 krawędzi (9 `queued`, 6 `selected`): `SelectorManager#connect` → `Connect#update`, `#accept` → `Accept#update`, `ManagedSelector#dump` → `DumpKeys#update`, `doStart` → `Start#update`… oraz `Connect#update`/`Accept#update`/`processConnect`… → `processSelected`. Nowa wyrocznia krawędzi `scripts/MethodEdgeTrace.java` (każda para wywołujący → wywoływany, którą wykonała JVM; 1853 pary) i `scripts/real-edge-paths.py` (ścieżka w grafie wyłącznie po krawędziach wykonanych lub modelowanych międzywątkowych z wykonanymi końcami): do `HttpURI#parse` dochodzi po prawdziwych krawędziach 43 ze 103 wykonanych wejść API (36 przed); `HttpClient#GET` dochodzi teraz po prawdziwych krawędziach do pętli selektora, `processConnect`, `CreateEndPoint#run`, `onOpen`, `ChannelEndPoint#onSelected`. Do `convertContentLength` bez zmian (3 ze 103): brakuje zadania *zwracanego* przez `onSelected` (`_runFillable`), które uruchamia `EatWhatYouKill`. Ścieżki raportowane bez zmian (66 z 408 wykonanych w całości — BFS wybiera krótszą trasę CHA); `exposed_api` 200 → 203 (`SelectorManager#accept`/`acceptor`, strona serwera); werdykty bez zmian (CHA i RTA).
- **Biblioteki (`scripts/library-measure.sh -offline`):** werdykty identyczne we wszystkich 20 przebiegach; liczby `exposed_api` identyczne poza jetty-client (200 → 203, CHA; 199 → 202, RTA). Krawędzie: jetty 15, starter 5 (`DeferredLog#log` → gettery `DeferredLog$Line` — odtwarzane i czyszczone; `BaseExecutor#deferLoad` → `DeferredLoad#load`; `ConfigFileApplicationListener$Loader`), pozostałe 0. `peek`/`element`/`getFirst`/`getLast` wymagają czyszczenia jak iteracja: bez tego stos zakresów SpEL (`ExpressionState#enterScope` → `VariableScope#setVariable`) i kolejka zdarzeń `YAMLGenerator` dawały krawędzie do bieżącego elementu, nie drenażu.
- **Aplikacje referencyjne (`a3-2` vs `a3-3`, 12 konfiguracji):** petclinic/WebGoat/JPetStore: 41/47/1 krawędzi `callback` dodanych (48/52/4 magazyny z drenażem, 2/9/0 załączników kluczy), 0 usuniętych (`-graph-compare`); **żaden werdykt ani ścieżka przykładowa nie zmienione** we wszystkich 12 konfiguracjach (`-roots profile`/`all` × CHA/RTA).
- **Ograniczenia:** zadanie przekazane jako wartość zwracana (`Producer#produce` → `task.run()`) i przez `AtomicReference` (`FillInterest`) nie są śledzone; drenaż przez metodę pomocniczą zwracającą element (`QueuedThreadPool#idleJobPoll`) nie jest rozpoznany; obiekt w kolejce nieznany (nie `new`/lambda/parametr) — brak krawędzi (licznik); klucz NIO parowany po zagnieżdżeniu klas, nie po selektorze; kolekcje inne niż JDK nie są modelowane.
- **Następny krok:** zadania zwracane z metody (`onSelected` → `_runFillable`, `ExecutionStrategy.Producer`) — ostatnia brakująca krawędź na prawdziwej ścieżce jetty do `convertContentLength`.

### ASYNC_CALLBACKS:A3 wycinek 2 — zadania przekazywane przez parametry (`a3-2`) — DONE for this scope (2026-11-19)

- **Problem (LB4, jetty-client):** ścieżki przykładowe bibliotek z callbackami nie były śladami wykonania. Nowa wyrocznia stosów (`scripts/MethodStackTrace.java`: stos wątku przy pierwszym wejściu do metody, z bci każdej ramki) pokazała, że na prawdziwych ścieżkach do `HttpURI#parse` i `HttpParser#convertContentLength` brakuje w grafie dokładnie jednej krawędzi: `QueuedThreadPool#runJob(Runnable)` → `job.run()` (odbiorca z JDK, zadanie z kolejki puli). Zadania docierają tam przez `Executor#execute(task)` w `SelectorManager#execute(Runnable task)` — parametr, więc A1 zostawiał miejsce nierozwiązane.
- **Reguła (`internal/callback/forward.go`):** gdy argument callbacku miejsca rejestracji jest parametrem metody rejestrującej, jest rozwiązywany u jej wywołujących (`ResolveExact` celów wywołań w całej hierarchii): `new` albo lambda/referencja → cel (osobny `ResolvedCallback` z łańcuchem `Forwarded`), parametr wywołującego → dalej (3 poziomy, 32 miejsca wywołań na miejsce rejestracji). Krawędź „może wykonać” z miejsca rejestracji (jak A2), bez modelu kolejki. Log `callbacks: … (N through their method's callers — ASYNC_CALLBACKS a3-2 — with T targets)`.
- **Weryfikacja:** fixture `testdata/async-forward` (wyrocznia JVM `scripts/async-forward-oracle.sh`, prawdziwy `ThreadPoolExecutor`): `Task` (1 poziom), lambda, `Deep` (3 poziomy) — wykonane i z krawędziami, `Unused` — ani jedno, ani drugie (`TestForwardedTasks`). Istniejące wyrocznie: luki „parametr” w `testdata/async` (`FromParam`) i `testdata/continuations` (lambda `fromParam`) — obie wykonane na JVM — są teraz krawędziami (`TestAsyncOracle`, `TestContinuationsOracle` zaktualizowane).
- **jetty-client:** krawędź `SelectorManager#execute` → `ManagedSelector$CreateEndPoint#run` (prawdziwe przekazanie przed `HttpURI#parse`); z 402 zgłoszonych ścieżek 66 wykonuje się w całości (40 przed; `scripts/path-divergence.py` na śladzie `MethodBreakTrace`), dla CVE-2023-40167 wejścia z w pełni wykonaną ścieżką 20 → 46 z 200; werdykty bez zmian. Ścieżka z `HttpClient#GET` nadal nie jest prawdziwa: zgłoszenie połączenia idzie przez kolejkę selektora (`ManagedSelector#submit`), własną abstrakcję Jetty, nie `Executor`.
- **Biblioteki (`scripts/library-measure.sh -offline`):** werdykty i liczby `exposed_api` identyczne we wszystkich 20 przebiegach.
- **Aplikacje referencyjne (`main` vs `a3-2`, 12 konfiguracji):** petclinic/WebGoat/JPetStore — 1/5/8 miejsc rozwiązanych przez wywołujących (1/7/23 cele); liczniki luk callbacków maleją (petclinic 28 → 25); **jedna zmiana statusu**: WebGoat `-roots profile` RTA CVE-2023-5685 (XNIO) `NOT_REACHABLE` + `MODEL_INCOMPLETE` → `REACHABLE`, tak jak już w CHA i `-roots all` — nowy osiągalny kod sprawia, że RTA zachowuje krawędź CHA `AbstractIoFuture#setException` → `CancelRequestedState#notifyFailed` (`-graph-compare`); ścieżka przykładowa przez kanał klienta AJP Undertow nie jest sprawdzona na JVM. 3 ścieżki przykładowe WebGoat zmienione.
- **Ograniczenia:** suma po wywołujących bez kontekstu (metoda przekazująca dostaje krawędzie do wszystkich zadań, które ktokolwiek jej podaje); pola, wartości zwracane i kolekcje nie są śledzone; wywołania przez nadtyp do nadpisującej metody przekazującej nie są liczone jako wywołujący.

### LIBRARY_MODE:LB4 — wyrocznie JVM dla pozostałych bibliotek (`lb4-2`) — DONE (2026-11-19)

- **Cel:** dokończyć kwalifikację LB4 — werdykty commons-configuration2, jackson-dataformat-yaml i jjwt-jackson były sprawdzone tylko ręcznie.
- **Wyrocznia (`scripts/library-lb4-jvm-oracle.sh`, `testdata/library-lb4/<lib>/`):** klient używający wyłącznie publicznego API (typowe użycie + scenariusze celowane w zmapowane metody), śledzony `MethodBreakTrace` na klasach biblioteki, klasach zmapowanych metod wszystkich ustaleń (`mapped.txt`) i — dla jackson-yaml — `ObjectMapper` (API dziedziczone); `observed.txt` stabilne w dwóch przebiegach. `TestLibraryLB4Oracles` (z mutacją: zmiana werdyktu wykonanego ustalenia na `NOT_REACHABLE` wyłapana).
- **Wynik:** commons-configuration2 2/2 `REACHABLE` wykonane; jackson-yaml 15/17; jjwt-jackson 14/15; **żadne `NOT_REACHABLE` (0/9, 0/21) nie wykonało się** — brak fałszywych negatywów. Dla każdego wykonanego ustalenia metoda API użyta przez klienta jest w `exposed_api`. Niepokazane: CVE-2026-68494 (parser asynchroniczny, w obu bibliotekach Jacksona; w jjwt RTA już `NOT_REACHABLE`) i CVE-2026-89425 w jackson-yaml (parsery JSON) — w modelu tylko przez parser przekazany przez wywołującego do metody dziedziczonej / CHA po `JsonParser#nextToken`; nadaproksymacja dla zwykłego użycia.
- **Poprawka przy okazji:** `testdata/library-lb4/jetty-client/client.jar` (z `lb4-1`) zawierał cały katalog fixture (jary Jetty, cache; 1,7 MB) — `JavaBuild` pakuje katalog źródeł; źródła klientów przeniesione do `<lib>/client/fx`, jar 2,3 KB; `scripts/library-lb4-oracle.sh` pomija ukryte klasy lambd (nazwy zmienne między uruchomieniami), `observed.txt` stabilne.
- **Kod analizy bez zmian.** Każda biblioteka zestawu ma teraz wyrocznię JVM.

### LIBRARY_MODE:LB2 — biblioteki ogłaszające klasy Springowi przez framework (`lb2-8`) — DONE for this scope (2026-11-19)

- **Decyzja użytkownika (2026-11-19):** po pomiarze rozbieżności ścieżek („implement it”) — dla bibliotek ogłaszających klasy Springowi pakiety Springa liczą się jako framework automatycznie.
- **Reguła (`springFrameworkPatterns` w `cmd/ariadne`, `libapi.SpringAnnouncement`, `libapi.Version` = `lb2-8`):** z `-roots api`, gdy biblioteka ma ogłoszenia Springa (`spring.factories`, `META-INF/spring/*.imports`; nie CDI), wzorce `framework_packages` wbudowanego profilu `spring`, których wybrany profil nie ma, są dodawane do dopasowania frameworka — przy dowolnym `-profile`. v2 `coverage.api.framework_packages_added`, log `library mode (lb2-8): …`.
- **Testy:** `TestSpringFrameworkPatterns` (spring.factories, `*.imports`, CDI, brak ogłoszeń, profil spring już wybrany), `TestLibraryAutoconfig` (wzorce dodane, werdykty fixture bez zmian — ścieżka tylko przez commons-text).
- **Pomiar (`scripts/library-measure.sh -offline`, `lb2-7` vs `lb2-8`):** starter mybatis `-roots api` CHA 25 `REACHABLE` → `FRAMEWORK`, 1 `REACHABLE` (CVE-2026-113730, OGNL przez `XMLConfigBuilder`, tylko z konfiguracją XML); RTA 23 → `FRAMEWORK`, 1 `REACHABLE`. Wobec wyroczni: wszystkie 12 wykonanych na JVM są `FRAMEWORK` (nadal zgłaszane). Pozostałe 4 biblioteki i wszystkie przebiegi `-roots all`: bez zmian.
- **Weryfikacja:** `go test ./...`, `go vet ./...`; aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, `main` vs `lb2-8`, offline): v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Ograniczenia:** `FRAMEWORK` nie odróżnia ustaleń, które kontener naprawdę wykonuje, od tych zależnych od konfiguracji aplikacji (pomiar rozbieżności: żadna poprawka wysyłki tego nie rozdziela); `spring-context` nie jest w pakietach frameworka profilu `spring` (bez zmian w profilu).

### LIBRARY_MODE: rozbieżność ścieżek startera z wykonaniem (po `lb2-7`) — DONE, pomiar (2026-11-19)

- **Pytanie (decyzja użytkownika: „measure it”):** czy nadaproksymacja inicjalizacji logowania trzyma niewykonane ustalenia startera jako `REACHABLE` i co by dał jej model.
- **Metoda:** klasy wszystkich 364 ścieżek 26 ustaleń (`scripts/path-divergence.py classes`), śledzone `MethodBreakTrace` w obu scenariuszach aplikacji startera; dla każdej ścieżki pierwsza niewykonana metoda (`scripts/path-divergence.py diverge`); dla miejsc CHA cele CHA z grafu (narzędzie robocze) też śledzone.
- **Wynik:** 9 z 364 ścieżek ma wszystkie metody wykonane (zbiór, nie kolejność: CVE-2022-22965/22970 przez `BeanWrapperImpl`, CVE-2021-42550 przez inicjalizację logowania); 18 krawędzi rozbieżności — przeskok logowania `StaticLoggerBinder#getLoggerFactory` → `ContextJNDISelector` (93 ścieżki), gałęzie zależne od wartości (70), korzenie niewołane w scenariuszu (79), dalsze miejsca CHA w kontenerze. **W 11 z 13 rozbieżnych miejsc CHA wywołanie nastąpiło, ale JVM wybrała inną implementację** (`DefaultContextSelector`, `LiteralExpression`, `DependencyObjectProvider`, `ClassPathResource#exists`…) — wybraną przez konfigurację lub zbiór beanów.
- **Eksperymenty (binarka robocza, niecommitowana):** E1 — sam model selektora logback (usunięta krawędź do `ContextJNDISelector`): zero zmian werdyktów na starterze (CHA/RTA) i petclinic (`-roots all`, CHA; 3 CVE konfiguracji logback zostają `FRAMEWORK` innymi drogami). E2 — górna granica: w każdym rozbieżnym miejscu CHA, gdzie JVM wybrała inny cel, zostają tylko cele wykonane, powtarzane po ponownym śledzeniu nowych ścieżek (4 rundy, 254 krawędzie w 18 miejscach, zbieżne): **wszystkie 26 nadal `REACHABLE`**, wejść 338 → 175. *(Pierwszy przebieg E2 pokazał „wszystkie `NOT_REACHABLE`” — błąd eksperymentu: separator reguł `;` występuje w deskryptorach, pusta reguła usunęła wszystkie krawędzie; poprawione.)* Potem ustalenia trzymają **gałęzie zależne od konfiguracji/danych** wołane wprost (`DataSourceInitializerInvoker#initialize` tylko z `schema.sql`, `ContextInitializer#configureByResource` tylko z `logback.xml`, lambda tłumacza wyjątków tylko przy błędzie, `scanClasses` tylko z `type-aliases-package`, ścieżki błędów) — tego analiza niewrażliwa na ścieżkę nie rozróżni.
- **`-profile spring`** na starterze: 25/26 `FRAMEWORK` (CHA i RTA), 1 `REACHABLE` — CVE-2026-113730 (OGNL przez `XMLConfigBuilder`, tylko z konfiguracją XML i `${…}`). Pomiar bibliotek używa `-profile generic`.
- **Wniosek:** ani model selektora logback, ani nawet idealna wysyłka w zmierzonych miejscach nie zmienia werdyktu: `REACHABLE` startera trzyma kod kontenera, którego wykonanie zależy od konfiguracji aplikacji (wybór implementacji i gałęzie zależne od zasobów/właściwości). Rozróżnia to dopiero status: z `-profile spring` 25/26 to `FRAMEWORK`. Kod analizy bez zmian. Szczegóły: `testdata/library-starter/EXPECTED.MD`.
- **Następny krok (decyzja użytkownika):** dla bibliotek ogłaszających klasy Springowi (`lb2-5`) traktować pakiety Springa jako framework (jak `-profile spring`) — starter: 26 `REACHABLE` → 1 `REACHABLE` + 25 `FRAMEWORK`; albo zostawić i opisać w użyciu CI.

### LIBRARY_MODE:LB2 — obiekty dostarczane przez wywołującego (`lb2-7`) — DONE for this scope (2026-11-19)

- **Decyzja użytkownika (2026-11-19):** luka jak `client_code` — znakowanie, werdykty bez zmian (opcja „drop edges” odrzucona).
- **Reguła (`libapi.SuppliedSites`, `libapi.Version` = `lb2-7`, `reach.EngineVersion` 4, `gaps.Version` `lb2-7`):** wywołanie wirtualne/interfejsowe na typie zależności (spoza biblioteki i JDK, interfejs lub klasa nie-`final`), którego odbiorca jest *dostarczony*: parametr metody, gdy ta jest korzeniem — tylko dla ścieżek startujących w tym korzeniu (biblioteka może wołać ten sam korzeń z własnym obiektem) — albo pole instancyjne biblioteki typu zależności, którego każdy zapis zapisuje parametr korzenia nigdy nie wołanego przez bibliotekę (na każdej ścieżce). Luka `client_code` (kod miejsca `supplied`, pusty zbiór kandydatów, nigdy `MODEL_INCOMPLETE`); krawędzie bez zmian. `exposed_api[].supplied_dispatch`, gdy każda ścieżka wejścia przechodzi przez takie wywołanie; powód `SUPPLIED_DISPATCH_ONLY` (oś modelu), gdy wszystkie wejścia ustalenia są oznaczone (v1: wiersz, gdy wszystkie pozytywne ustalenia). v2 `coverage.api.supplied` (`sites`, `field_sites`, `edges`, `fields`, `unanalyzed_methods`); log `caller-supplied objects (LIBRARY_MODE lb2-7): …`; raport „caller-supplied”.
- **Fixture `testdata/library-supplied`** (oczekiwania przed implementacją, pomiar „przed” na `lb2-6`): (y) parametr korzenia, (z) pole z settera → oznaczone; (aa) własny lookup, (bb) pole też z domyślnej wartości biblioteki, (cc) wartość przekazana do konstruktora zależności (granica), (dd) korzeń (y) wołany przez bibliotekę z własnym obiektem → nieoznaczone; z `-api-exclude sp.api` CVE-2025-48924 `SUPPLIED_DISPATCH_ONLY` (CHA; pod RTA `NOT_REACHABLE` — istniejące przybliżenie RTA: nic nie tworzy `ConstantStringLookup`). `TestLibrarySupplied` (z mutacją: znakowanie niezależne od kontekstu wyłapane), `TestSuppliedSites`.
- **Poprawki w trakcie (zapisane w EXPECTED.MD):** pierwsza wersja (a) traktowała parametr korzenia jako dostarczony na każdej ścieżce — jackson-yaml oznaczał 21 wejść przez `YAMLParser#<init>` wołany przez bibliotekę; (b) liczyła pola ustawiane z konstruktorów, które biblioteka woła sama (jetty `AbstractConnectionPool` z `Callback` = własny `HttpDestination`); (c) znakowała przekazanie dostarczonej wartości do metody zależności — na starterze cała krawędź `sqlSessionTemplate` → `new SqlSessionTemplate(…)`, choć dalsza ścieżka obiektu nie dotyka. (c) usunięte, (a)/(b) zawężone; (dd) dodany do fixture.
- **Biblioteki (`scripts/library-measure.sh -offline`, `lb4-1` vs `lb2-7`):** werdykty i listy `exposed_api` identyczne we wszystkich 10 przebiegach. Oznaczenia: starter 26/338 (CHA; `MybatisAutoConfiguration#<init>` w każdym ustaleniu), jetty-client 6/400, pozostałe 0; żadne ustalenie `SUPPLIED_DISPATCH_ONLY`. Powód na starterze: w każdym ustaleniu 7–11 z 13 wejść dochodzi do tego samego kodu przez **inicjalizację logowania** (slf4j → `ContextJNDISelector` → konfigurator logback → `SpringPropertyAction` → `Environment` → SpEL/`BeanFactory`) — wywołania interfejsowe w samych zależnościach, których implementację wybiera konfiguracja. To, nie obiekty z kontenera, trzyma 14 niewykonanych ustaleń jako `REACHABLE`.
- **Weryfikacja:** `go test ./...`, `go vet ./...`; aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, `main` vs `lb2-7`, offline, ten sam cache): v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Następny krok (decyzja użytkownika):** nadaproksymacja inicjalizacji logowania (CHA po `ContextSelector`/akcjach joran w zależnościach) — osobny problem, dotyczy też aplikacji.

### LIBRARY_MODE: wyrocznia startera mybatis (po `lb4-1`) — DONE, pomiar (2026-11-19)

- **Pytanie:** które z 26 `REACHABLE` startera mybatis-spring-boot-autoconfigure 2.1.2 (`-roots api`, CHA) naprawdę się wykonują — pomiar LB4 pokazał 22 ścieżki `heuristic` przez kontener Springa.
- **Wyrocznia (`testdata/library-starter`, `scripts/library-starter-oracle.sh`):** najmniejsza aplikacja startera (Boot 2.2.5, HSQLDB 2.5.1 z `~/.m2`, jeden `@Mapper`), scenariusze `plain` i `yml` (`application.yml` z jedną właściwością `mybatis.*`); nowy `scripts/MethodBreakTrace.java` — jednorazowe breakpointy JDI na metodach zmapowanych klas (9 s; `MethodTrace` z wejściem do każdej metody nie skończył w 8 min). Bez sieci.
- **Wynik:** wykonane w obu scenariuszach 5 (CVE-2022-22965/22970 `CachedIntrospectionResults`, CVE-2026-41848 `AntPathMatcher` z `SpringBootVFS#list`, CVE-2021-42550 i CVE-2026-41851 tylko przez metody pomocnicze szerokich mapowań diff); tylko z `application.yml` 7 (snakeyaml, `undeclared` — na JVM z ładowania konfiguracji przez sam Boot, zgłoszone ścieżki przez mybatis się nie wykonały); nigdy 14 (7 SpEL, 4 konfiguracja logback, serwer WWW, `MimeType`, OGNL — wymagają, by aplikacja coś zrobiła).
- **Wzorzec:** 10 z 14 niewykonanych i wszystkie 7 snakeyaml opuszczają kod mybatis wywołaniem na obiekcie **przekazanym bibliotece przez kontener** (parametr `ObjectProvider` → `getIfAvailable`, `BeanFactory` w `AutoConfigurationPackages#get`), który CHA rozsyła do wszystkich implementacji Springa. To otwarty punkt LB2 „wywołania na typach zależności”, zmierzony na prawdziwej bibliotece. `heuristic` ich nie oddziela (12 z 14 niewykonanych, ale też 9 z 12 wykonanych).
- **Kod analizy bez zmian.** Szczegóły: `testdata/library-starter/EXPECTED.MD`.
- **Następny krok (decyzja użytkownika):** jak traktować wywołanie na odbiorcy typu zależności, którego biblioteka nie tworzy (parametr, wartość z kontenera) — np. jak `client_code`: luka i oznaczenie zamiast pełnego CHA po implementacjach zależności.

### LIBRARY_MODE:LB4 — zestaw kwalifikacyjny i użycie w CI (`lb4-1`) — DONE for this scope (2026-11-19)

- **Zestaw (`testdata/library-lb4/EXPECTED.MD`):** pięć prawdziwych bibliotek, po jednej na styl API — commons-configuration2 2.7 (fasada/builder), jackson-dataformat-yaml 2.12.3 (fabryka, podklasa mappera), jjwt-jackson 0.11.2 (SPI), mybatis-spring-boot-autoconfigure 2.1.2 (starter), **nowa: jetty-client 9.4.36** (callbacki, selektor NIO). `scripts/library-measure.sh` (piąta biblioteka, lista nazw z tablicy) × `-roots all`/`api` × CHA/RTA, offline na `lb2-6`; jetty: jeden przebieg online za zgodą użytkownika — 1 plik z Central, bez 429, reszta z `~/.m2`.
- **Wynik:** `-roots api` nie traci żadnego `REACHABLE` z `-roots all` (jedna różnica RTA w jjwt — parser asynchroniczny, RTA trafnie); zyskuje 16 na jackson-yaml (metody odziedziczone, `lb2-3`). Werdykty pozostałych czterech jak w pomiarach LB2/LB3.
- **Mapowania ręczne (2):** jetty-client miał 4 z 5 advisory `UNKNOWN` (`MAPPING_MISSING`, OSV bez commitów poprawki). Dodane z porównania bytecode jetty-http 9.4.36 z kopią w jetty-runner 9.4.57 (`work/jardiff`, `work/calldump`): CVE-2023-40167 → `HttpParser#convertContentLength` (9.4.36: `Long.parseLong`, przyjmuje `+16`; 9.4.57: pętla po cyfrach), CVE-2022-2047 → prywatne `HttpURI#parse(State,String,int,int)` (9.4.57 sprawdza znaki authority). Tylko linia 9.4 (`[9.4.0,9.4.52)`/`[9.4.0,9.4.47)`), 109 wpisów, `TestServerManualMappings`.
- **Wyrocznia JVM (`scripts/library-lb4-oracle.sh`, fixture `testdata/library-lb4/jetty-client`: jary + 25 wpisów cache):** klient woła tylko `HttpClient#start`/`GET`/`stop` wobec lokalnego gniazda z `Content-Length: +5` → `status=200 body=hello` (podatność realna dla użytkownika biblioteki); JDI: obie zmapowane metody wykonane. Analiza (CHA i RTA): obie `REACHABLE`, `HttpClient#GET` (oba przeciążenia) w `exposed_api`; 85 z 200 (RTA 199) narażonych metod jetty-client wykonało się w tym jednym scenariuszu. CVE-2024-6763 `NOT_REACHABLE` — zgodnie z advisory (użycie `HttpURI` przez samo Jetty nie jest podatne), choć powód `MAPPED_CODE_ABSENT` jest przypadkowy (mapowanie z Jetty 12); CVE-2025-11143, CVE-2026-2332 `UNKNOWN` (bez commitu, 9.4 EOL). `TestLibraryJetty` (offline, pusty `ARIADNE_MAVEN_LOCAL`).
- **Sprawdzone ręcznie:** starter mybatis — 22 z 26 ścieżek `REACHABLE` `heuristic` przez kontener Springa (`ObjectProvider#getIfAvailable` → CHA po `BeanFactory` → `YamlMapFactoryBean`/SpEL), 12 z 26 zależności `undeclared` (logback przez slf4j `<clinit>` — realne, ale aplikacji); jjwt — 14 z 15 przez `JacksonDeserializer#readValue` (bez JVM).
- **Ograniczenia znalezione (nienaprawione):** (1) przekazanie asynchroniczne (selektor NIO → `ReadCallback#succeeded`) nie ma krawędzi — werdykt jetty wychodzi dobrze, bo CHA znajduje inną drogę (`HttpRequest#abort` → `notifyComplete` → `AuthenticationListener#onComplete` → `send`), której środek się nie wykonał; ścieżki przykładowe bibliotek z callbackami nie są śladami wykonania *(zamknięte w grafie przez `a3-2`–`a3-4`, potwierdzone pomiarem 2026-11-21 — wpis wyżej; ścieżki przykładowe nadal najkrótsze, nie wykonane)*; (2) publiczne ≠ zamierzone API (`HttpReceiverOverHTTP#receive`); (3) `heuristic` nie filtruje szumu (realne na databind); (4) `exposed_api` do ≈780 metod.
- **Dokumentacja:** README „Library mode in a library's CI” (uruchomienie, `-api-exclude`, `ARIADNE_MAVEN_LOCAL`, interpretacja `exposed_api`/`supplied_by`, granice); SUPPORT.MD — wiersz trybu biblioteki z kwalifikacją i zmierzonymi ograniczeniami (stan `Partial`).
- **Weryfikacja:** `go test ./...`, `go vet ./...`. Kod analizy bez zmian (`libapi.Version` `lb2-6`); zmiana wyników tylko przez 2 nowe mapowania ręczne jetty-http 9.4.
- **Następny krok:** plan CI floty (tydzień od 2026-11-23) z trybem biblioteki; kandydaci: krawędź przekazania selektora/executora (ASYNC_CALLBACKS) dla bibliotek z callbackami, warunki `@ConditionalOn…` dla starterów.

### LIBRARY_MODE:LB2 — archiwa beanów CDI jako korzenie (`lb2-6`) — DONE for this scope (2026-11-18)

- **Problem:** biblioteka będąca archiwum beanów CDI (Jakarta EE/Weld, Quarkus/ArC) ma kod, który kontener wywołuje niezależnie od dostępu — pakietowe metody `@Produces`, obserwatory `@Observes`, `@Inject`, `@PostConstruct` — a reguła publicznego API go nie widziała. Fixture: przed zmianą obie podatności `NOT_REACHABLE` pod `-roots api`, a Weld wykonuje `replace`.
- **Zakres (decyzja użytkownika, 3a):** CDI (`META-INF/beans.xml`, indeks Jandex) zamiast osobno Quarkusa — Quarkus nie generuje kodu w jarze biblioteki, odkrywa ją jak każdy kontener CDI; prawdziwa aplikacja Quarkus jako wyrocznia wymagałaby setek artefaktów z Central (po 429). Micronaut osobno, później (najpierw pomiar, czy wygenerowane `$Definition` są już publicznym API).
- **Reguła (`libapi.CDIAnnouncements`, gałąź CDI w `AnnouncedRoots`, `libapi.Version` = `lb2-6`):** tryb odkrywania z `bean-discovery-mode`; bez niego pusty lub sprzed 4.0 → `all`, 4.0+ → `annotated`; `none` → nic; sam Jandex → `annotated`. `annotated`: klasy z adnotacją definiującą beana (zakresy, `@Dependent`, `@Singleton`, `@Interceptor`, `@Decorator`, `javax`/`jakarta`, adnotacje biblioteki z `@Stereotype`/`@NormalScope`/`@Scope`); `all`: konkretne klasy z konstruktorem bezargumentowym lub `@Inject`. Wejścia beana: konstruktory, `<clinit>`, metody `@Inject`/`@Produces`/`@PostConstruct`/`@PreDestroy`/`@AroundInvoke`/`@AroundConstruct` i z parametrem `@Observes`/`@ObservesAsync`/`@Disposes` (dowolny dostęp), publiczne metody instancyjne. Pochodzenie: `<plik> bean-discovery-mode=<tryb>[ @<adnotacja>]: <klasa>`. Alternatywy, specjalizacja, weta i rozszerzenia nieoceniane.
- **Fixture `testdata/library-cdi`** (oczekiwania zapisane przed implementacją, pomiar „przed” na `lb3-1`): (t) pakietowy `@Produces`, (u) pakietowy obserwator, (v) pakietowy `@PostConstruct`, (w) pakietowe `@Inject`, (x) klasa bez adnotacji definiującej (nie bean). Wyrocznia `scripts/library-cdi-oracle.sh` (Weld SE 5.1.2.Final, CDI 4.0, JDI; Weld pobrany raz Mavenem do `work/cdi-oracle`): wykonane (t)–(w) i `replace`, (x) nigdy. Wynik (CHA i RTA): obie podatności `REACHABLE`, `exposed_api` = dokładnie (t)–(w) z pochodzeniem, wszystkie wykonane na JVM. `TestLibraryCDI`, `TestCDIAnnouncements` (tryby `annotated`/`all`/`none`, sam Jandex, jar bez archiwum), `TestDiscoveryMode`, `TestCDIRoots` (wejścia, pochodzenie, `-api-exclude`).
- **Prawdziwe jary (zależności Quarkus 3.9.5 z `work/quarkus-demo`, `-roots api`, CHA/RTA, `main` vs `lb2-6`, offline po jednym rozgrzaniu cache — 25 + 8 plików z Central, bez 429):** smallrye-context-propagation 2.1.0 (`beans.xml`, `annotated`): 1 bean — publiczny interceptor `SmallRyeCurrentThreadContextInterceptor`, 0 nowych korzeni, werdykty i `exposed_api` bez zmian; mutiny 2.5.8 (sam Jandex): 0 beanów. Cztery biblioteki pomiarowe (`scripts/library-measure.sh`) nie są archiwami beanów — wynik bez zmian z konstrukcji.
- **Weryfikacja:** `go test ./...`, `go vet ./...`; aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, `main` vs `lb2-6` offline): v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Ograniczenia:** tryb `all` bierze każdą konkretną klasę z odpowiednim konstruktorem (może zawyżać); producenci w polach nie są modelowani (nie wykonują kodu); `@Singleton` traktowany jak definiujący (Quarkus), choć w CDI nim nie jest; rozszerzenia CDI (`Extension`) tylko jako dostawcy `META-INF/services`.
- **Następny krok:** pomiar Micronauta (3b) albo warunki `@ConditionalOn…` (LB2).

### LIBRARY_MODE:LB2 — odbiorca w polu (kontynuacja `lb2-4`) — NOT IMPLEMENTED, decyzja (2026-11-18)

- **Pytanie:** czy specjalizacja odziedziczonych korzeni powinna objąć pole ustawiane w konstruktorze klasy API (`YAMLMapper` → `ObjectMapper#_jsonFactory`), żeby `ObjectMapper#createParser` na `YAMLMapper` nie sięgał parsera JSON (wpis `lb2-4` nazywał tę ścieżkę pozostałym przybliżeniem).
- **Sprawdzenie na JVM (`scripts/YAMLMapperFactoryProbe.java`, jackson 2.12.3):** `new YAMLMapper().createParser(…)` → `YAMLParser`; `new YAMLMapper((YAMLFactory) null).createParser(…)` → `ReaderBasedJsonParser`. Publiczny konstruktor przyjmuje `null`, a `ObjectMapper(JsonFactory)` zapisuje wtedy `new MappingJsonFactory(this)`. Ścieżka jest więc osiągalna z API — nie jest nadmiarem modelu.
- **Decyzja (z użytkownikiem):** reguły nie implementujemy. Poprawna (typy wszystkich zapisów pola w łańcuchu konstruktorów, z gałęzią `null`) nie usuwa na jackson nic; pomijająca gałąź `null` ukryłaby realną ścieżkę. Zysk byłby tylko przy polach przypisywanych bezwarunkowo, bez zmierzonego przykładu. Dokumentacja (README, plan LIBRARY_MODE, wpis `lb2-4`, SUPPORT.MD) poprawiona.
- **Następny krok:** ogłoszenia innych frameworków (Quarkus, Micronaut) jako korzenie (LB2).

### LIBRARY_MODE: pomiar z `-deps-list` (po `lb3-1`) — DONE (2026-11-18)

- **Problem (z `lb3-1`):** w pomiarze mybatis-spring-boot-autoconfigure 8 jarów Spring Boot/Spring (budowane Gradle, bez `pom.properties`) było w wejściu bez tożsamości — ich advisory nigdy nie były oceniane.
- **Zmiana:** `scripts/library-measure.sh` zapisuje `mvn -o dependency:list` każdej biblioteki (`work/libs/<name>/deps.txt`, Maven offline na lokalnym repozytorium) i przekazuje go jako `-deps-list`. Kodu analizy bez zmian.
- **Wynik (`-offline`; OSV dla nowych jarów pobrane jednym przebiegiem online tylko dla mybatis):** commons-configuration2, jackson-dataformat-yaml, jjwt-jackson — ustalenia, werdykty i `exposed_api` identyczne z `lb3-1` (wszystkie 12 przebiegów). mybatis: 25/25 jarów zidentyfikowanych, 19 → 39 ustaleń; 20 nowych na spring-boot, -autoconfigure, -beans, -context, -core, -expression, wszystkie `supplied_by: library`; `REACHABLE` z API 13 (CHA) / 11 (RTA), m.in. CVE-2022-22965 (spring-beans), CVE-2022-27772 (spring-boot), CVE-2024-38820 (spring-context) i 8 advisory SpEL (spring-expression); `lb3-1` raportuje 0 `unidentified`. Uwaga: CVE-2020-26945 (mybatis) i CVE-2024-12798 (logback-core) przeszły z `UNKNOWN` (`SOURCE_UNAVAILABLE`/`MAPPING_MISSING`) na `NOT_REACHABLE` z `MODEL_INCOMPLETE` — przez źródła poprawek dociągnięte do cache w przebiegu online, nie przez `-deps-list` (ten sam cache bez listy daje to samo).
- **Następny krok:** LB3 część opcjonalna (advisory dla wersji z deklarowanego zakresu) albo LB4.

### LIBRARY_MODE:LB3 — deklaracje zależności biblioteki (`lb3-1`) — DONE for this scope (2026-11-18)

- **Problem:** wyniki dotyczą wersji z `-deps`, ale użytkownik biblioteki może dostać inne, a część zależności (`provided`, `optional`) istnieje tylko, gdy doda je sam. Wynik nie mówił, kto dostarcza instancję ustalenia.
- **Reguła (`component.LibraryDeclarations`, `libapi.Version` = `lb3-1`):** przy `-roots api` czytany jest pom biblioteki (wbudowany `META-INF/maven/<g>/<a>/pom.xml`; przy kilku ten, którego artifactId nazywa plik jara; albo `-library-pom`) z rodzicami, właściwościami i `dependencyManagement` (wersje i zakresy, także importowane BOM-y — tylko w LB3, F6b bez zmian). Drzewo konsumenta jak w Mavenie: deklaracje `compile`/`runtime` bez `optional`, przechodnio przez pomy zależności (wersja z wejścia, gdy spełnia deklarację, inaczej deklarowana dokładna), najbliższy wygrywa, `exclusions`, łączenie zakresów, ≤ 6 poziomów → `supplied_by: library`; `provided`/`system`/`optional` i to, co przychodzi tylko przez nie → `user`; jar wejścia spoza drzewa → `undeclared`; `test` pomijany. Dopasowanie po `group:artifact` do instancji (bez osadzonych), `version_match`: `declared`, `differs`, `in_range`, `out_of_range`, `undetermined`, `unresolved`, `unidentified` (jar `<a>-<v>.jar` w wejściu bez tożsamości), `absent`. v2: `coverage.api.dependencies`, `components[].declaration`, `findings[].supplied_by`; raport: sekcja „Library dependencies”; ostrzeżenie dla każdej zależności dostarczanej przez bibliotekę, która jest `absent`/`unidentified` (jej advisory nie są oceniane). Werdykty i kody powodów bez zmian.
- **Dostęp do Maven Central (`cmd/ariadne/central.go`):** pierwszy pomiar online (2026-11-18) dostał od repo1.maven.org HTTP 429 („tool identified as abusive”; ok. 50 udanych i kilkanaście ponowionych nieudanych zapytań — błędy nie trafiały do cache). Teraz: `ARIADNE_MAVEN_LOCAL` (lokalne repozytorium Mavena, tylko pliki pobrane z repozytorium zdalnego wg `_remote.repositories`) przed siecią, własny User-Agent `ariadne/<wersja>`, nieudany URL nie jest ponawiany w przebiegu, po 429 nic więcej nie jest wysyłane (ostrzeżenie). Dotyczy też F6a/F6b. `scripts/library-measure.sh` ustawia `ARIADNE_MAVEN_LOCAL`; pomiar powtórzony z `-offline`. `TestCentralClient`.
- **Fixture `testdata/library-deps`** (oczekiwania zapisane przed implementacją; po pierwszym przebiegu jawnie dopisany `spring-jcl` — zależność compile opcjonalnego `spring-core`): (p) commons-text `provided` z zakresem `[1.6,1.10.0)` przez właściwość (wejście 1.9: `in_range`, user), (q) commons-lang3 3.11 compile (library, `declared`), (r) commons-io 2.6 `runtime` (library, `absent`, ostrzeżenie), (s) spring-core `optional` (user, `absent`), junit `test` (pominięty). Wyrocznia `scripts/library-deps-oracle.sh`: Maven 3.9.14 `dependency:list` konsumenta z repozytorium plikowego = dokładnie (q) i (r); zbiór `supplied_by: library` jest mu równy (CHA i RTA). `TestLibraryDeps`, `TestLibraryDeclarations` (najbliższy wygrywa, wykluczenia, przechodnie `provided`, `optional` poniżej biblioteki, wersja i zakres z rodzica, zakres wersji, nierozwiązana właściwość, niezidentyfikowany jar, niezadeklarowane), `TestLibraryDeclarationsBOM`, `TestLibraryPOMChoice`.
- **Biblioteki (`scripts/library-measure.sh -offline`, `lb2-5` vs `lb3-1`, ten sam cache):** werdykty i zbiory `exposed_api` identyczne we wszystkich 16 przebiegach. commons-configuration2: 3 library (commons-text 1.8, commons-lang3 3.9, commons-logging 1.2 — `declared`), 17 user (opcjonalne). jackson-dataformat-yaml: 4 library, `declared` (jackson-core/-annotations przez databind). jjwt-jackson: 4 library, `declared`. mybatis-spring-boot-autoconfigure: wersje przez BOM Spring Boot; 8 jarów Spring Boot/Spring przychodzi przez `spring-boot-autoconfigure` i jest w wejściu **bez tożsamości** (brak `pom.properties`) — ich advisory nigdy nie były oceniane w tym pomiarze (wiadomość nowa); mybatis i mybatis-spring `optional` (user), slf4j-api `differs` (1.7.30 deklarowane, 1.7.25 w wejściu), 9 jarów `undeclared` (wnosi je starter, nie autokonfiguracja); osadzony ognl bez `supplied_by` (idzie z hostem).
- **Weryfikacja:** `go test ./...`, `go vet ./...`; aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, `lb2-5` vs `lb3-1` offline): v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Ograniczenia:** advisory oceniane tylko dla wersji z wejścia (nie dla innych wersji z zakresu; `MAPPING_VERSION_UNCOVERED` nie zaimplementowane); profile pomu, metadane modułów Gradle nieczytane; zakresy `import` tylko w pomie biblioteki i jej rodzicach (nie w pomach zależności — tam wersje są zwykle jawne); dopasowanie `unidentified` tylko po nazwie pliku.
- **Następny krok:** LB3 część opcjonalna (advisory dla wersji z deklarowanego zakresu, najnowsze w linii) albo `-deps-list` w `scripts/library-measure.sh`, żeby jary Springa w mybatis zostały ocenione.

### LIBRARY_MODE:LB2 — autokonfiguracja jako korzenie (`lb2-5`) — DONE for this scope (2026-11-17)

- **Problem:** kod, który biblioteka udostępnia Springowi (autokonfiguracje, słuchacze z `spring.factories`), bywa pakietowy — metody `@Bean` bez modyfikatora, zagnieżdżone `static class …Configuration`, klasy z `@Import` — więc reguła publicznego API go nie widziała, choć Spring go wywołuje. Fixture: przed zmianą obie podatności `NOT_REACHABLE`, a JVM wykonuje `replace`.
- **Reguła (`libapi.Announcements`, `libapi.AnnouncedRoots`, `libapi.Version` = `lb2-5`):** odczyt jara biblioteki — każda klasa każdego klucza `META-INF/spring.factories` (składnia `Properties`, kontynuacje `\`), każda klasa z `META-INF/spring/*.imports` (klucz = nazwa pliku). Wpis `EnableAutoConfiguration`/`AutoConfiguration` albo klasa z adnotacją konfiguracji: konstruktory, `<clinit>`, metody `@Bean` (dowolny dostęp, także statyczne), metody `@Autowired`/`@Value`/`@Inject`/`@Resource`/`@PostConstruct`/`@PreDestroy`/`@EventListener`, publiczne metody callbacków kontenera (`InitializingBean`, `DisposableBean`, `ImportSelector`, `ImportBeanDefinitionRegistrar`, post-procesory), zagnieżdżone klasy konfiguracji i cele `@Import` — przechodnio. Inny klucz `spring.factories` (słuchacz, inicjalizator, …): konstruktory, `<clinit>`, publiczne metody instancyjne. Warunki nieoceniane („może działać”); `-api-exclude` usuwa klasy ogłoszone jak dostawców. Pochodzenie: v2 `exposed_api[].announced_by` (`<plik> <klucz>: <klasa>` + ` > member …`/` > @Import …`), raport „announced: …”; `coverage.api.announced_classes`, `coverage.api.announcements[]`; log `library API (lb2-5): …, K more from C framework-announced classes in A announcements`.
- **Fixture `testdata/library-autoconfig`** (oczekiwania zapisane przed implementacją): starter z commons-text (zależności i cache z `testdata/library`), Spring jako `provided` (tylko kompilacja i wyrocznia). Przypadki (j) `@Bean` pakietowy, (k) zagnieżdżona konfiguracja, (l) `@PostConstruct` klasy z `@Import`, (m) słuchacz z `spring.factories`, (n) konfiguracja nieogłoszona, (o) wpis tylko w `AutoConfiguration.imports`. Wyrocznia `scripts/library-autoconfig-oracle.sh` (JDI, Spring Boot 2.4.5 z `@EnableAutoConfiguration`): wykonane (j)–(m) i `replace`, (n) nie, (o) nie (Boot 2.4 nie czyta plików `.imports` — oczekiwana różnica). Wynik (CHA i RTA): obie podatności `REACHABLE`, `exposed_api` = dokładnie (j), (k), (l), (m), (o) z pochodzeniem; (n) nigdy. `TestLibraryAutoconfig`, `TestAnnouncedRoots`, `TestParseFactories`.
- **Prawdziwy starter (`scripts/library-measure.sh`, nowy wpis):** mybatis-spring-boot-autoconfigure 2.1.2 z zależnościami startera — 2 ogłoszenia, 9 klas (2 autokonfiguracje, 6 zagnieżdżonych konfiguracji, rejestrator z `@Import`), 5 nowych korzeni (48 → 53); werdykty i `exposed_api` identyczne z `lb2-4` (19 ustaleń, CHA i RTA) — jego autokonfiguracje są publiczne, więc zmiana dokłada tylko `announced_by` (7 z 13 metod w `exposed_api` CHA). Trzy dotychczasowe biblioteki: brak ogłoszeń, wyniki bez zmian.
- **Weryfikacja:** `go test ./...`, `go vet ./...`; aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, `lb2-4` vs `lb2-5` na tym samym cache offline): v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Ograniczenia:** warunki `@ConditionalOn…`/`@Profile` nieoceniane (może zawyżać); produkty metod `@Bean` (ich `init`/`destroy`, wywołania kontenera) nie są korzeniami, chyba że są publicznym API; `@ImportAutoConfiguration`, `@ComponentScan` i meta-adnotacje własne biblioteki nie są śledzone; klasy zagnieżdżone rozpoznawane po nazwie `Outer$Inner` z adnotacją konfiguracji; katalogi innych frameworków (Quarkus `META-INF/quarkus-extension.properties`, Micronaut) nieczytane.
- **Następny krok:** ścieżki warunkowe po powrocie z kodu klienta albo LB3 (wersje i zakresy zależności z poma).

### LIBRARY_MODE: specjalizacja odziedziczonych korzeni do odbiorcy (`lb2-4`) — DONE (2026-11-17)

- **Problem (z `lb2-3`):** korzeń odziedziczony był wspólnym węzłem metody zależności, więc wywołanie na `this` w nim rozchodziło się jak dla dowolnego obiektu — `JsonFactory#createJsonParser` na `YAMLFactory` sięgał parsera JSON, choć `YAMLFactory` nadpisuje `createParser`.
- **Reguła (`libapi.Specialize`, `libapi.Version` = `lb2-4`):** dla każdej pary (metoda odziedziczona, klasa API) kopia węzła (`callgraph.Graph.AddClone`; ten sam klucz, `ID`/`IDsByName` zwracają oryginał). W kopii wywołanie `invokevirtual`/`invokeinterface` na niezmienionym `this` (`dataflow.AnalyzeWithThis`: lokalna 0 oznaczona `valueflow.This`; zapis do niej, `checkcast` albo złączenie zdejmuje znacznik) idzie tylko do ciał, które wybiera klasa API lub jej podtypy (`ResolveExact` + reguła wysyłki grafu z klasą API jako odbiorcą); gdy nic się nie rozwiązuje, krawędź zostaje. Wywołanie na `this` (także `invokespecial`: prywatne, `super`) metody zadeklarowanej w właściwym nadtypie klasy API idzie do kopii tej metody dla tej samej klasy — przechodnio. Inne krawędzie (inne odbiorniki, modele, `<clinit>`, SAM) kopiowane bez zmian; metoda, której analiza nie przeszła, zachowuje wszystkie krawędzie (`unanalyzed_methods`). Po RTA/PTA, przed reach.
- **Reach:** `bfsResult.path` dla oryginału nieodwiedzonego bezpośrednio bierze najkrótszą ścieżkę przez jego kopie (pierwszą przy remisie), więc statusy, `MethodReachable` i luki działają bez zmian; `exposed_api` liczy wsteczny BFS także od kopii celu i scala kopie tej samej metody w jeden wpis (`inherited_by` = klasy, których kopia sięga kodu; ścieżka: `app_only` > bez `client_dispatch` > krótsza). Sonda (`probe`) traktuje cel-kopię jak oryginał. v2 `coverage.api.receiver_specialization` (`clones`, `this_sites`, `edges_dropped`, `unanalyzed_methods`); log `receiver specialization (LIBRARY_MODE lb2-4): …`.
- **Fixture:** przypadek (i) `StrictExpander` nadpisuje chroniony `substitute(TextStringBuilder,int,int)` (woła go na `this` każde odziedziczone `replace`); oczekiwanie zapisane przed zmianą w `EXPECTED.MD`. Wyrocznia: `scripts/library-oracle.sh` uruchamia też `fx.Client strict` → `observed-strict.txt` (JVM: `StringSubstitutor#replace(String)` i `StrictExpander#substitute` tak, `StringSubstitutor#substitute` i `resolveVariable` nie). Wynik (CHA i RTA): `StrictExpander` w `inherited_by` przeciążeń `replace` dla CVE-2022-42889 (to ich kod zmapowany), dla CVE-2025-48924 w żadnym — tylko w `resolveVariable`; binarka `lb2-3` na tym samym fixture wymieniała go dla obu. `TestSpecialize` (kopie, krawędzie, wyrocznia), `TestLibraryMode`, `TestRoots`.
- **Biblioteki (`-roots api`, offline, ten sam cache, `lb2-3` vs `lb2-4`):** jackson-dataformat-yaml — 682 kopie, 1892 miejsca na `this` (RTA 1642), 1007 krawędzi usuniętych (RTA 777), 0 nieprzeanalizowanych; werdykty bez zmian (17/26 `REACHABLE`, CHA i RTA); `exposed_api` CVE-2025-49128 101 → 87 (m.in. `JsonFactory#createJsonParser`, `JsonParser#nextFieldName`/`nextTextValue`… na `YAMLParser`, `JsonGenerator#writeBinary`… na `YAMLGenerator`), CVE-2026-68494 i CVE-2026-89425 61 → 53. Co zostaje: odbiorca w polu/parametrze (`ObjectMapper#createParser` → `_jsonFactory.createParser` → parser JSON — *korekta 2026-11-18: to nie przybliżenie, ścieżka jest realna, zob. wpis „odbiorca w polu”*; `readValueAs` przekazuje `this` do `ObjectMapper`) oraz sam kod zmapowany (`JsonParseException#getMessage` → `JsonLocation#_appendSourceDesc` w wyjątkach YAML — przy specjalizacji ścieżka idzie przez nadpisane `getMessage`, o krok dłuższa, więc przykładowa ścieżka findingu to teraz `ObjectMapper#createParser`). commons-configuration2, jjwt-jackson: 0 metod odziedziczonych, wyniki bez zmian. Czas: jackson-yaml +0,7 s (1,7 → 2,5 s).
- **Weryfikacja:** `go test ./...`, `go vet ./...`; aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, `lb2-3` vs `lb2-4` na tym samym cache offline): v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Ograniczenia:** specjalizacja tylko dla `this` w ciele metody; `this` zapisany do pola lub przekazany jako argument traci kontekst; konstruktory i metody statyczne nie są kopiowane; liczba węzłów grafu w `coverage`/v1 obejmuje kopie (tylko `-roots api`).
- **Następny krok:** autokonfiguracja jako korzenie (LB2) albo odbiorca w polu ustawianym w konstruktorze klasy API (`YAMLMapper` → `YAMLFactory`) — *odłożone 2026-11-18, zob. wpis „odbiorca w polu”*.

### LIBRARY_MODE: metody odziedziczone z zależności jako API (`lb2-3`) — DONE (2026-11-17)

- **Decyzja użytkownika (2026-11-17):** metody, które klasa API dziedziczy z nadtypu zależności, są API.
- **Reguła:** dla dostępnej klasy API każda metoda instancyjna z ciałem, `public` (lub `protected`, gdy klasa nie jest `final`), zadeklarowana w nadklasie lub nadinterfejsie z zależności, do której wywołanie na klasie API się rozwiązuje (nie nadpisana po drodze). Korzeniem jest węzeł metody zależności. v2: `exposed_api[].inherited_by` (klasy API) i `coverage.api.inherited_methods`; raport: „inherited by …”; log `… N methods (M inherited from dependency supertypes)`.
- **Fixture:** przypadki (g) `Expander extends StringSubstitutor` i (h) `SafeExpander` z nadpisanym `replace(String)`; wyrocznia JVM odtworzona (`scripts/library-oracle.sh`, klient kompilowany z commons-text). `TestLibraryMode`: `replace(String)` odziedziczone tylko przez `Expander`, pozostałe przeciążenia przez obie klasy; zbiory własnych metod bez zmian.
- **Biblioteki (`scripts/library-measure.sh`):** jackson-dataformat-yaml 640 korzeni odziedziczonych, `REACHABLE` 1 → 17 z 26 (CHA i RTA). Ręcznie: 13 z 16 nowych to deserializacja databind przez odziedziczone `readValue`/`readValueAs` (realne dla użytkownika `YAMLMapper`/`YAMLParser`); CVE-2025-49128 to przybliżenie — `JsonFactory#createJsonParser` → `createParser` → parser JSON, choć `YAMLFactory` nadpisuje `createParser`/`_createParser` (węzeł odziedziczony nie jest specjalizowany do odbiorcy); CVE-2026-68494 i CVE-2026-89425 przychodzą z parametru `JsonParser` w `JsonGenerator#_copyCurrentContents` (dowolny parser klienta — to samo przybliżenie parametrów co w LB1). commons-configuration2 i jjwt-jackson: 0 metod odziedziczonych, wyniki bez zmian.
- **Weryfikacja:** `go test ./...`, `go vet ./...`; aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, porównanie z LB1 na tym samym cache offline): v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Ograniczenia:** brak specjalizacji do odbiorcy (wywołania na `this` w metodzie odziedziczonej rozchodzą się jak dla każdego obiektu zależności); metody statyczne nadtypów nie są korzeniami; konstruktory nadklas tylko przez konstruktory klasy API.
- **Następny krok:** specjalizacja odziedziczonych korzeni do odbiorcy (wywołania na `this` rozwiązywane z klasy API) albo autokonfiguracja (LB2).

### CVE (kontynuacja F10) — mapowania tylko z inicjalizatorów statycznych — DONE (2026-11-17)

- **Problem:** poprawki „gadżetowych” CVE jackson-databind (typowanie polimorficzne) tylko dopisują nazwy klas do czarnej listy w statycznym inicjalizatorze, więc źródło diff mapowało je na `SubTypeValidator#<clinit>` (wcześniej `BeanDeserializerFactory#<clinit>`). Ten inicjalizator uruchamia każda deserializacja — każdy użytkownik Jacksona dostawał je jako `REACHABLE`. Pomiar 2026-11-17: jjwt-jackson 0.11.2 (databind 2.9.10.4), 21 z 36 `REACHABLE` to ta rodzina.
- **Decyzja:** (1) mapowanie diff/release-diff złożone wyłącznie z inicjalizatorów rodziny Jacksona jest zastępowane API włączającym default typing (`ObjectMapper#enableDefaultTyping*`/`activateDefaultTyping*`/`setDefaultTyping`, odpowiedniki `MapperBuilder`; sprawdzone w bytecode 2.0.0–2.13.5), rola `entrypoint`, inicjalizatory jako dowód (v2 `mapping.initializer_redirect`, notatka wiersza); `@JsonTypeInfo(use = CLASS)` nie jest pokryte (mapowanie częściowe). (2) Każde inne mapowanie diff/release-diff złożone tylko z `<clinit>` zachowuje werdykt i dostaje kwalifikator `MAPPING_INITIALIZER_ONLY` (v2 `mapping.initializer_only`). `ResolverVersion` 8. Kod: `internal/vulnmap/initializers.go`.
- **Weryfikacja:** `TestFixDiff_jacksonBlocklistRedirect`, `TestRedirectInitializers`, `TestInitializerOnly`, `TestInitializerReasons` (z mutacjami); `scripts/library-measure.sh -offline` (gałąź pomiaru LB2 + ta zmiana): jjwt-jackson `REACHABLE` 36 → 15 (`-roots all` CHA/RTA, `-roots api` CHA), 35 → 14 (`-roots api` RTA), suma `exposed_api` 78 → 32 (CHA); configuration2 i jackson-yaml (databind 2.12.3) bez zmian. CVE-2020-25649 (`DOMDeserializer#<clinit>`) zostaje `REACHABLE` z `MAPPING_INITIALIZER_ONLY`. Aplikacje referencyjne (petclinic/WebGoat/JPetStore, `-roots all`/`profile` × CHA/RTA, v1, offline): żaden status się nie zmienił; `MAPPING_INITIALIZER_ONLY` dochodzi do 2/2/3 wierszy (`UriComponentsBuilder#<clinit>` CVE-2024-22243/22262, w JPetStore też `MimeTypeUtils#<clinit>` CVE-2018-1272).
- **Ograniczenia:** polimorfizm przez adnotację `@JsonTypeInfo(use = CLASS/MINIMAL_CLASS)` w kodzie aplikacji nie jest wykrywany (możliwy fałszywy `NOT_REACHABLE`, oznaczony `MAPPING_PARTIAL`); `disableDefaultTyping`/`deactivateDefaultTyping` wołają `setDefaultTyping(null)`, więc ich wywołanie da ścieżkę; instancyjne inicjalizatory (`C#<init>`) nie dostają kwalifikatora (selektor nie odróżnia ich od jawnego konstruktora).

### LIBRARY_MODE: pomiar na prawdziwych bibliotekach + `lb2-2` — DONE (2026-11-17)

- **Biblioteki** (`scripts/library-measure.sh`, Maven Central → `work/libs/`, `-roots all`/`api` × CHA/RTA):
  - commons-configuration2 2.7 (fasada/builder; commons-text 1.8, commons-lang3 3.9);
  - jackson-dataformat-yaml 2.12.3 (fabryka, `YAMLMapper extends ObjectMapper`; snakeyaml 1.27, databind 2.12.3);
  - jjwt-jackson 0.11.2 (SPI przez `META-INF/services`; databind 2.9.10.4).
- **Wyniki (`api`, CHA):**

  | Biblioteka | Metody API | Werdykty (`all` = `api`) | `exposed_api` na ustalenie | Miejsca `client_code` (osiągalne) |
  |---|---|---|---|---|
  | commons-configuration2 | 2156 | 2 `REACHABLE` | 772, 778 | 2346 (2259) |
  | jackson-dataformat-yaml | 248 | 1 `REACHABLE`, 25 `NOT_REACHABLE` | 15 | 126 (101) |
  | jjwt-jackson | 10 | 36 `REACHABLE` (RTA: 35 + 1 `NOT_REACHABLE`) | 1–6 | 2 (2) |

  Werdykty `-roots api` = `-roots all` we wszystkich 12 przebiegach (poza jednym RTA w jjwt: CVE-2026-68494, parser asynchroniczny, którego jjwt nie używa — RTA trafniej). Tryb nie zmienia więc werdyktów na tych bibliotekach; jego wartością jest lista `exposed_api`.
- **Sprawdzone ręcznie:**
  - commons-configuration2: `ConfigurationInterpolator#interpolate` → `StringSubstitutor#replace` — prawdziwe (interpolacja przy każdym getterze), stąd ~770 metod API.
  - jackson-yaml: wszystkie 6 CVE snakeyaml `NOT_REACHABLE` — poprawnie, Jackson używa parsera zdarzeń, nie `Composer`/`Yaml#load`.
  - jjwt: 20 z 36 `REACHABLE` kończy się na `SubTypeValidator#<clinit>` — mapowania z diffów czarnej listy gadżetów databind; ta inicjalizacja wykonuje się przy każdym użyciu Jacksona, więc to szum mapowania (oś CVE), nie trybu biblioteki.
- **Błąd znaleziony i naprawiony (`lb2-2`):** w `lb2-1` 98 wpisów `client_dispatch` w commons-configuration2 (np. `AbstractConfiguration#addErrorLogListener` → `PatternSubtreeConfigurationWrapper#addEventListener`) wskazywało publiczne klasy, które klient tworzy przez `new` — to zwykły kod biblioteki, nie zastępstwo kodu klienta. Teraz klasa, którą klient może utworzyć lub rozszerzyć bez nadpisania (`libapi.Constructible`), nie jest zastępstwem: 98 → 0; przypadek (d) fixture bez zmian (`TestConstructible`, `TestLibraryMode`).
- **Granica znaleziona (niezmieniona):** metody publiczne odziedziczone przez klasę API z nadklasy zależności (`YAMLMapper#readValue` z `ObjectMapper`) nie są korzeniami — ani w `api`, ani w `all`. Eksperyment (tymczasowa łata, wycofana): jackson-yaml +640 korzeni, `REACHABLE` 1 → 17 z 26 (CVE databind/core przez odziedziczone `readValue`); pozostałe dwie biblioteki bez zmian. To decyzja o definicji API (kontrakt §8), nie poprawka.
- **Szum `client_code`:** liczba miejsc rośnie z rozmiarem API (2346 w commons-configuration2 — prawie każde wywołanie `this.` na nadpisywalnej metodzie). Raportowa, nie zmienia werdyktów; do floty przyda się raczej liczba miejsc na ścieżkach `exposed_api` niż suma.

### LIBRARY_MODE:LB2 wycinek 1 — kod klienta jako luka (`lb2-1`) — DONE for this scope (2026-11-17)

- **Zakres:** miejsca kodu klienta i oznaczenie ścieżek przez nie; bez nowych krawędzi i bez zmian werdyktów.
- **Miejsca:** `libapi.ClientSites` — wywołanie wirtualne/interfejsowe w metodzie biblioteki, którego rozwiązana metoda jest zadeklarowana w dostępnym, nie wyłączonym typie biblioteki, który klient może rozszerzyć (interfejs albo klasa nie `final` z konstruktorem `public`/`protected`), ma dostęp `public`/`protected`, nie jest `static` ani `final`. Reguły dostępności wspólne z korzeniami (`selector`).
- **Luka:** kod miejsca `client`, mechanizm `client_code` (słownik `internal/gaps`, tabela w kontrakcie). Zbiór kandydatów pusty — klasy klienta nie są w wejściu, a kod klienta wchodzi do biblioteki tylko przez API, które już jest korzeniem — więc nigdy `MODEL_INCOMPLETE`; liczona w `coverage.gaps` (`count`, `reachable`).
- **`client_dispatch`:** krawędź, której każde miejsce jest miejscem klienta, cel nie jest ciałem wołanej metody, a klasy celu (ani podtypu dziedziczącego ciało) wejście nie tworzy (katalog P1, bez założonego `entry`), to wysyłka zastępująca kod klienta. `reach.SetExposure` liczy wsteczny BFS także bez takich krawędzi: metoda API sięgająca kodu tylko przez nie ma `client_dispatch: true`, inaczej jej ścieżka przykładowa ich unika. Raport: „client dispatch” w sekcji Exposed API; logi `library API (lb2-1): …, N client-code call sites`, `client-code dispatch (LIBRARY_MODE LB2): …`.
- **Weryfikacja:**
  - `testdata/library`: 3 miejsca (`Pipeline#run` → `Listener#onEvent`, `Codecs#decodeAll` → `Codec#decode`, `Templates#render` → `Templates#prepare`), wszystkie osiągalne; pod CHA (także z `-api-exclude lib.internal`) `client_dispatch` ma tylko `Pipeline#run` — jedyna narażona metoda, której JVM nie potwierdza (`observed.txt`); `Codecs#decodeAll` nie, bo `ExpandCodec` tworzy `ServiceLoader` biblioteki; pod RTA nic. Zbiory `exposed_api` bez zmian. Zapisane w `EXPECTED.MD` (sekcja LB2, pomiar po implementacji).
  - `TestLibraryMode` (rozszerzony), `TestClientSites`, `TestEveryProducerCodeMapped`, `TestVocabularyDocumented`; `go test ./...`, `go vet ./...`.
  - Aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, porównanie z LB1 na tym samym cache offline): v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Ograniczenia:**
  - wywołania na typach zależności i JDK, które klient może zaimplementować, nie są miejscami klienta;
  - podatny kod osiągalny dopiero po powrocie z kodu klienta nie jest oznaczany jako warunkowy;
  - „tworzy wejście” liczone z katalogu P1 całego wejścia, nie z osiągalności z API;
  - klucz snapshotu `library.api` (zapisywany w każdym trybie) to teraz `lb2-1`: replay snapshotu z `lb1-1` odmówi, jak przy każdym podbiciu wersji algorytmu;
  - brak pomiaru na prawdziwych bibliotekach (LB4).
- **Następny krok:** autokonfiguracja Spring Boot biblioteki jako korzenie (LB2 wycinek 2) albo pomiar na bibliotekach z floty.

### LIBRARY_MODE:LB1 — korzenie z publicznego API i narażone metody — DONE (2026-11-16)

- **Zakres:** `-roots api` i `-api-exclude`. Korzenie wybiera `internal/libapi` (`lb1-1`) według kontraktu §8; przy `-jpms-scenario module-path` liczy się `exports`/`provides` z `module-info`.
- **Wynik:** silnik (`reach.SetExposure`) wsteczny BFS od podatnych metod, z regułą przebiegu „tylko kod aplikacji” (bez przechodzenia przez framework), z pamięcią podręczną na zbiór celów. Wyjścia:
  - v2 `findings[].exposed_api[]` (metoda, przebieg, przykładowa ścieżka z dowodem i charakterem);
  - v2 `coverage.api`;
  - sekcja raportu „Exposed API”;
  - logi `library API (lb1-1)` i `roots: api`;
  - snapshot `library.api`.
- **RTA:** w trybie `api` klasy dostępne liczą się jako tworzone (`EntryRoots`), bez „dowolnego podtypu” parametrów.
- **Biblioteka kontrolna** rozszerzona o klasę zagnieżdżoną (publiczną i prywatną) oraz `protected` w klasie zwykłej i w `final`; wyrocznia odtworzona.
- **Weryfikacja:**
  - `TestLibraryMode`: wynik dokładnie jak w `EXPECTED.MD` — CHA 5, RTA 4, `-api-exclude lib.internal` 3. Pod RTA każda narażona metoda to metoda wykonana przez JVM. Test działa offline na odpowiedziach OSV i GitHuba skopiowanych do `testdata/library/cache` (14 wpisów z manifestu `-snapshot-out`).
  - `TestRoots` (`internal/libapi`): reguły dostępności.
  - `go test ./...`, `go vet ./...`.
  - Porównanie v1 z `main` w 12 konfiguracjach (`-roots all`/`profile`) — wynik w opisie PR.
- **Dokumentacja:** README (sekcja i flagi), SUPPORT.MD (nowy wiersz), RESULTS_CONTRACT §8.
- **Ograniczenia:**
  - kod klienta niemodelowany (LB2), autokonfiguracja i wersje zależności (LB2, LB3);
  - Kotlin `internal` liczy się jako API;
  - dostępność klas zagnieżdżonych z flag klasy i klas otaczających, nie z `InnerClasses`;
  - brak pomiaru na prawdziwych bibliotekach (przeniesiony do LB4).
- **Następny krok:** LB2 albo pomiar na bibliotekach z floty.

### LIBRARY_MODE:LB0 — kontrakt trybu biblioteki — DONE (2026-11-16)

- **Zakres:** kontrakt bez implementacji. W `docs/RESULTS_CONTRACT.MD` §8:
  - wybór trybu przez `-roots api`;
  - definicja publicznego API (dostępność klas zagnieżdżonych, `protected` tylko w klasach nie-`final`, `<clinit>`, dostawcy `META-INF/services`, bez metod syntetycznych i lambd);
  - zawężenia: JPMS `exports` przy `-jpms-scenario module-path` oraz `-api-exclude`;
  - znaczenie statusów;
  - nowe pola v2 `findings[].exposed_api[]` i `coverage.api`;
  - zarezerwowany mechanizm luki `client_code` (LB2).
- **Biblioteka kontrolna** `testdata/library`: `lib.jar` z przypadkami (a)–(f) na commons-text 1.9 (CVE-2022-42889) i commons-lang3 3.11; klient `client.jar`; `EXPECTED.MD` z oczekiwaniami LB1 dla CHA, RTA i `-api-exclude`.
- **Wyrocznia:** `scripts/library-oracle.sh` (JDI, `scripts/MethodTrace.java`). JVM wykonuje `replace` z przypadków (a), (e) i (f). Nigdy nie wchodzi w `Formatter#unsafe`, `Hidden#run` ani `ExpandingListener#onEvent`, bo w (d) działa listener klienta (`testdata/library/observed.txt`).
- **Pomiar „przed”:** `-roots all` (CHA i RTA) daje ścieżkę z `Engine#expand` (klasa pakietowa). Do `replace` prowadzi 9 metod biblioteki, w tym 3, których żaden klient nie wywoła.
- **Weryfikacja:** wyrocznia uruchomiona. Brak zmian kodu, więc `go test ./...` bez wpływu.
- **Następny krok:** LB1 — korzenie `api`, `exposed_api`, test na `testdata/library` według `EXPECTED.MD`.

### FRAMEWORKS — sterowniki JDBC według URL (`jdbc:drivers` `u1`) — DONE (2026-11-16)

- **Problem:** `DriverManager#getConnection(url)` oferuje URL każdemu
  zarejestrowanemu sterownikowi, a `service-provider:dispatch` łączy
  `Driver#connect` ze wszystkimi sterownikami z `META-INF/services`. Przez
  to WebGoat (HSQLDB) „osiągał” cały kod połączenia pgjdbc. Z 8 nowych
  `REACHABLE` z części 2–3 redukcji `UNKNOWN` 2 brały się stąd.
- **Model (`internal/jdbcurl`, domyślnie włączony — decyzja 2026-11-16):**
  sterownik z katalogu (pgjdbc 42.x, Connector/J 8.0.23+, H2 1.4/2.x),
  którego prefiks nie pasuje do żadnego URL-a aplikacji, ma `connect`
  ucięte po sprawdzeniu URL-a (`callgraph.Graph.CutSitesAfter`: późniejsze
  miejsca wywołań usunięte, a krawędzie modeli z nich odrzucane we
  wszystkich ścieżkach dodawania). Dzieje się to przed modelami i bramkami
  twórców. Źródła URL-i: stałe `jdbc:` klas aplikacji, pliki konfiguracji
  (wszystkie profile, Spring XML, `web.xml`), H2/HSQLDB/Derby przy
  wsparciu baz wbudowanych Springa. **Założenie:** spakowana konfiguracja
  to wdrażana. Model wyłącza się dla przebiegu, gdy ma nieustawiony
  placeholder URL-a, JNDI albo URL bez literalnego `jdbc:<sub>:`.
  Nadpisania zmienną środowiskową nie widać.
- **Wyrocznia JVM:** `testdata/jdbcdrivers` (`scripts/jdbcdrivers-oracle.sh`;
  nowe `scripts/MethodTrace.java` śledzi wejścia do metod przez JDI) —
  `jdbc:fx:` z pgjdbc 42.2.18 i Connector/J 8.0.23 na classpath: JVM
  wchodzi w `connect` i sprawdzenie URL-a obu, dalej nic. Graf osiąga oba
  sprawdzenia i nic za nimi; bez modelu z `main` osiągalnych jest 6993
  metody, z modelem 82, żadna nowa. Nieosiągnięte, a wykonane przez JVM:
  ładowanie sterowników przez `DriverManager` (`<clinit>`, konstruktory,
  rejestracja, wątek czyszczący MySQL) — istniejąca luka SPI, ta sama z
  modelem i bez.
- **Wynik (v1 vs `main`, 12 konfiguracji):** petclinic i JPetStore bez
  zmian (petclinic nie ma pgjdbc, MySQL zostaje przez profil `mysql`;
  JPetStore ma tylko HSQLDB). WebGoat we wszystkich 4: CVE-2022-21724,
  CVE-2022-26520, CVE-2026-42198 `REACHABLE` → `NOT_REACHABLE`, poza tym
  tylko liczby `model_gaps` (2 miejsca refleksji w pgjdbc mniej).
- **Weryfikacja:** `TestJDBCDriversOracle`, `TestCutSitesAfter`,
  `TestCollect`, `TestDecide`; `go test ./...`, `go vet ./...`.
  Dokumentacja: README, SUPPORT.MD.
- **Poza zakresem:** MySQL w petclinic (CVE-2023-22102, ścieżka przez H2 do
  `ResultSet#getObject`) zostaje, bo profil `mysql` naprawdę go używa; jsoup
  w WebGoat (`NodeVisitor` → `Cleaner`) i HTTP/2 Undertow to osobne
  nadmiarowe przybliżenia.

### FRAMEWORKS — profil `spring`: listenery wiadomości jako korzenie — DONE (2026-11-16)

- **Zakres (zatwierdzony 2026-11-16):** metody `@JmsListener`,
  `@KafkaListener`/`@KafkaHandler`, `@RabbitListener`/`@RabbitHandler`
  (z kontenerami `@Repeatable`) jako korzenie `-roots profile` profilu
  `spring`; bez osobnego modelu. Ustalenie przed implementacją: plan floty
  CI używa `-roots all`, gdzie każda metoda aplikacji (więc i listener)
  już jest korzeniem — zmiana poprawia `-roots profile`, nie wyniki floty;
  użytkownik zatwierdził wersję małą, następnie redukcję `UNKNOWN`.
- **Zrealizowane:** `internal/profile/profiles/spring.yaml`
  `roots.annotations` += 8 adnotacji (spring-jms, spring-kafka,
  spring-rabbit). Klasa z `@KafkaListener`/`@RabbitListener` — korzenie to
  jej publiczne metody (istniejąca reguła adnotacji klasy; także publiczny
  konstruktor).
- **Weryfikacja:** `TestMessageListenerRoots` (`testdata/listeners`:
  adnotacje-zaślepki o prawdziwych nazwach, metody i klasy, powtórzone;
  bez zmiany profilu test nie przechodzi). Aplikacja kontrolna
  `scripts/prepare-listeners-demo.sh` (Boot 2.4.3, ActiveMQ w pamięci,
  `@JmsListener` → commons-text 1.9 `StringSubstitutor#replace`):
  `-roots profile` CVE-2022-42889 `NOT_REACHABLE` (binarka z `main`) →
  `REACHABLE` (korzeń: `OrderListener#onOrder`), JVM uruchamia listener
  przez kontener; wariant bez adnotacji — `NOT_REACHABLE`, JVM listenera
  nie uruchamia; `-roots all` — `REACHABLE` w obu.
  Aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, porównanie z `main`): żadna nie ma listenera wiadomości — v1 bajtowo identyczne we wszystkich 12 konfiguracjach. `go test ./...`, `go vet ./...`.
  Dokumentacja: README, SUPPORT.MD.
- **Ograniczenia:** dopasowanie po nazwie adnotacji, bez sprawdzania
  konfiguracji kontenera (`@EnableJms`/`@EnableKafka`, autokonfiguracja);
  bez Spring Cloud Stream (funkcje), `@SqsListener` i programowej
  rejestracji `MessageListener`.
- **Następny krok (zatwierdzony kierunek):** redukcja `UNKNOWN` (brak
  mapowania CVE → metoda) — zrobiona 2026-11-16 (PR #84–#87, wpisy CVE:F10
  niżej, [archiwum](../archive/2026-11-16/cve/CVE_F10_UNKNOWN_REDUCTION.md)).

### FRAMEWORKS:F3 — advice Spring AOP, domknięcie (`spring:aop` `a2`) — DONE for this scope (2026-11-16)

- **Zakres (zatwierdzony 2026-11-16, „kończ AOP”):** luki a1 — wiązanie
  zmiennych advice, `args`/`this`/`target`/`@target`/`bean`, wzorce
  adnotacji metody w `execution`, parametryzowane `@Pointcut`, filtr typu
  `returning`, aspekty z XML `<aop:config>`.
- **Zrealizowane:** `classfile.Method.ParameterNames` (atrybut
  `MethodParameters`, inaczej `LocalVariableTable` po slotach; nil, gdy nie
  każdy parametr ma nazwę). `springaop.methodScope` — kolejność jak w
  Springu: `argNames`/`arg-names`, nazwy z pliku klasy, wnioskowanie
  `AspectJAdviceParameterNameDiscoverer` (wiodący `JoinPoint`, `throwing` →
  jedyny `Throwable`, `returning` → jedyny pozostały, zmienna adnotacji →
  jedyny parametr typu adnotacji, jedna zmienna `this`/`target`/`args` →
  jedyny pozostały; niejednoznaczne zostaje niezwiązane). Nowe węzły:
  `args` (typy lub zmienne; argument pasuje, gdy jego typ deklarowany może
  zawierać instancję typu — test Springa w czasie wykonania; prymitywy
  opakowane), `this`/`target` (klasa beana), `@target`, `bean` (nazwy:
  jawne lub domyślne stereotypów, nazwy metod `@Bean`, id XML —
  `springBeanNamesByType` w `cmd/ariadne`), `@Adnotacja` na początku
  `execution`, odwołania do `@Pointcut` z parametrami (wiązane parametrami
  metody pointcutu). `@AfterReturning`: typ zmiennej `returning` filtruje
  po typie zwracanym (`Object` — każda metoda, także `void`).
  `springxml.AOPConfig` (sekcje `<aop:config>` z `<beans>` i
  zagnieżdżonych `<beans>`: pointcuty, aspekty, advice; `<aop:advisor>`
  liczony); advice XML: metoda o nazwie z najmniejszą liczbą parametrów
  (`BeanUtils.resolveSignature`), `and`/`or`/`not` w wyrażeniu, bramka:
  osiągalny `ConfigBeanDefinitionParser#parse` (obiekty advice tworzy
  kontener refleksyjnie). `scripts/JavaBuild.java` kopiuje zasoby
  (pliki inne niż `.java`) do jara — pozostałe fixture'y mają tylko
  źródła, ich jary nie zmieniają się. Wejście modelu: `springaop.Input`
  (typy beanów, nazwy, sekcje XML); linia logu liczy advice z
  `<aop:config>` i nierozwiązane.
- **Wyrocznia:** `testdata/aop` rozszerzony (bez informacji debug — zmienne
  wiąże wnioskowanie, jak Spring na JVM): 16 advice w `TracingAspect`
  (każdy desygnator powyżej, `bean()` bez beana, advice bez dopasowania),
  `XmlAudit` przez `fx/aop/aop.xml` (`and not`, `pointcut-ref`, advice bez
  dopasowania), `Repo#name` zwracające `String`. JVM uruchomił 15 ciał
  advice: a2 osiąga **15/15** (a1: 6, bez modelu: 0); 4 nieuruchamiane
  (`ghost`, `noBean`, `Unregistered#never`, `XmlAudit#ghost`) bez krawędzi;
  self-invocation bez krawędzi. Dopasowania „może” ponad wykonanie (jawne
  w `TestFindOnFixture`): `args(long)` na `save(Object)`, `onString` na
  `find` (zwraca `Object`, rzuca), `onError` na `name`.
- **Aplikacja kontrolna** `scripts/prepare-aop-demo.sh` (trzy warianty;
  Boot kompiluje z `-parameters`, nazwy z `MethodParameters` sprawdzone na
  klasach z Mavena): `-roots profile` CVE-2022-42889 — `advised` (`@Aspect`
  z `args(text)`): a1 `NOT_REACHABLE` (advice nierozstrzygnięte) → a2
  `REACHABLE`, JVM uruchamia advice; `unadvised`: `NOT_REACHABLE`, JVM nie
  uruchamia; `xml` (`<aop:config>`, `and args(text)`): a1 `NOT_REACHABLE` →
  a2 `REACHABLE`, JVM uruchamia; `-roots all` — `REACHABLE` we wszystkich.
  Aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, porównanie z `main` (a1) na tym samym cache offline): żadna nie ma klasy `@Aspect` ani `<aop:config>` (0 aspektów w logu) — v1 bajtowo identyczne we wszystkich 12 konfiguracjach.
- **Weryfikacja:** `TestAOPOracle`, `TestFindOnFixture`,
  `TestMethodScope`, `TestArgsAndReturn`, `TestParseXMLWords`,
  `TestParseExecution`, `TestParseOperatorsAndUnknown`,
  `springxml.TestParse_aopConfig`, `classfile.TestParameterNames`;
  `go test ./...`, `go vet ./...`. Dokumentacja: README, SUPPORT.MD.
- **Ograniczenia (pozostałe):** `@args`, `cflow`, wzorce adnotacji
  parametrów, zmienne, których Spring też by nie związał — bez krawędzi;
  `proceed()` w `@Around` nieśledzony; kolejność advice, typ `throwing`,
  `<aop:advisor>`, CTW/LTW, `this()` na proxy JDK, przechwycenie metody
  `final` przez interfejs proxy JDK — niemodelowane.
- **Następny krok:** obszar F3 dla AOP domknięty w zadeklarowanym zakresie;
  w tygodniu od 2026-11-23 plan floty CI (`FLEET_CI_PLAN.MD`).

### FRAMEWORKS:F3 — advice Spring AOP (`spring:aop` `a1`) — DONE for this scope (2026-11-16)

- **Zakres (zatwierdzony 2026-11-16):** krawędź „może wykonać” z
  wywołania metody beana objętej pointcutem do metody advice aspektu
  `@AspectJ` (`@Before`/`@After`/`@AfterReturning`/`@AfterThrowing`/
  `@Around`), wąski język pointcutów (`execution`, `within`,
  `@annotation`, `@within`, odwołania do `@Pointcut`, `&&`/`||`/`!`),
  reszta jako oznaczona luka; bez krawędzi dla self-invocation i metod
  `final`/`private`; `@Around`: `proceed()` opisany, krawędź do targetu bez
  zmian; model w koordynatorze B3 (wersja, `models.disable`). Poza
  zakresem: `<aop:config>`, CTW/LTW, `@Order`. Plan floty CI
  (`FLEET_CI_PLAN.MD`) przesunięty przez użytkownika na tydzień od
  2026-11-23.
- **Wyrocznia:** `testdata/aop` (`scripts/aop-oracle.sh`, `TestAOPOracle`;
  jary Spring 5.3.6 z `testdata/tx/deps`, aspectjweaver 1.9.6 tylko do
  uruchomienia, suma SHA-256 w skrypcie): `OrderService` przez CGLIB
  (self-invocation `place` → `internal`, metoda `final`, metoda
  `@Audited`), `JdbcRepo` przez proxy JDK (`save` wraca, `find` rzuca),
  `ReportService` z `@Monitored`; `TracingAspect`: `@Before` na nazwanym
  pointcucie, `@AfterReturning` (`Repo+`), `@AfterThrowing` (`within`),
  `@Around` z powiązaną adnotacją, `@Before` na `@within`, advice bez
  dopasowania, advice z `args(id)`; `Unregistered` — `@Aspect` bez
  `@Component`. Zapis wyroczni obejmuje też, dla których punktów złączenia
  uruchomiło się `before` (JVM: nie dla `internal` ani `fixed`). Przed
  modelem graf nie osiągał **żadnego** z 6 ciał advice uruchomionych przez
  JVM.
- **Zrealizowane:** `internal/springaop` — parser i trójwartościowe
  dopasowanie pointcutów (`pointcut.go`), katalog aspektów i advice,
  punkty złączenia (metody klas beanów: własne i dziedziczone, bez
  `Object`, statycznych, prywatnych, `final`, abstrakcyjnych,
  syntetycznych; aspekty nie są doradzane), miejsca wywołań (wywołanie
  wirtualne/interfejsowe o tej nazwie i deskryptorze, klasa beana
  przypisywalna do właściciela), self-invocation z faktu odbiorcy D2
  (`this` = wartość typowana klasy bez pochodzenia), budżet 50 000
  krawędzi, bramka: osiągalny twórca podtypu `AbstractAspectJAdvice`.
  `callgraph.SAMEdge` dostał `Condition` (przenoszony przez bramkę do
  `CallSite`): `returned` dla `@AfterReturning`, `exceptional` dla
  `@AfterThrowing`. Podłączenie w `cmd/ariadne` jako bramka punktu stałego
  B3 (uczestnik `spring:aop`, klucz snapshotu `spring.aop`, linia logu).
- **Wynik:** wyrocznia — wszystkie 5 ciał advice z rozstrzygniętymi
  pointcutami osiągnięte (krawędzie z `Main#main`, z warunkami); brak
  krawędzi do advice bez dopasowania, do aspektu niebędącego beanem i z
  self-invocation; `withArgs` (`args(id)`) nieosiągnięty — zadeklarowana
  luka (2 pary nierozstrzygnięte). Kontrola negatywna: ten sam przebieg z
  `models.disable: [spring:aop]` — 0 krawędzi do advice. Aplikacja
  kontrolna `scripts/prepare-aop-demo.sh` (Boot 2.4.3,
  `spring-boot-starter-aop`, `@Before` → commons-text 1.9
  `StringSubstitutor#replace`): `-roots profile` CVE-2022-42889
  `NOT_REACHABLE` (model wyłączony) → `REACHABLE` (ścieżka `main` →
  `AuditAspect#audit` → `replace`), JVM uruchamia advice; wariant z
  pointcutem na niewywoływanym `refund`: `NOT_REACHABLE`, JVM advice nie
  uruchamia; `-roots all` — `REACHABLE` w obu (każda metoda aplikacji jest
  korzeniem). Aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, porównanie z `main` na tym samym cache offline): żadna nie ma klasy `@Aspect` (0 aspektów w logu) — v1 bajtowo identyczne we wszystkich 12 konfiguracjach; efekt mierzalny tylko na wyroczni i aplikacji kontrolnej.
- **Weryfikacja:** `TestAOPOracle`, `TestFindOnFixture`,
  `TestNoCreatorWithoutSpringAOP`, `TestParseExecution`,
  `TestParseOperatorsAndUnknown`, `TestTypeAndParamPatterns`,
  `TestDescParts`, `TestTriLogic`; `go test ./...`, `go vet ./...`.
  Dokumentacja: README, SUPPORT.MD.
- **Ograniczenia:** krawędź „może wykonać”; `args`/`this`/`target`/`bean`,
  wzorce adnotacji w `execution` i nazwane pointcuty z parametrami —
  nierozstrzygnięte (bez krawędzi); `proceed()` w `@Around` nieśledzony;
  kolejność advice, ograniczenia typu `returning`/`throwing`,
  `<aop:config>`, CTW/LTW, proxy JDK dla metody `final` wołanej przez
  interfejs — niemodelowane; rejestracja to relacja typów, nie rozwiązana
  definicja beana.
- **Następny krok (wymaga zgody):** F3 — `args(...)` z typami parametrów
  advice (najczęstsza przyczyna nierozstrzygnięcia) albo `<aop:config>`;
  w tygodniu od 2026-11-23 plan floty CI.

### ASYNC_CALLBACKS:A3 — kontynuacje `CompletionStage`, wycinek 1 — DONE for this scope (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** katalog operacji
  `CompletionStage`/`CompletableFuture` (kontynuacje normalne,
  `exceptionally`, `handle`/`whenComplete`, `*Async`, źródła
  `supplyAsync`/`runAsync`) z krawędzią „może wykonać” przez istniejący
  mechanizm A1/A2; także w kodzie bibliotek (decyzja 1); gałąź wyjątku jako
  adnotacja w dowodzie, bez nowego rodzaju krawędzi (decyzja 2); budżet.
  Poza zakresem: listenery, reactive, taint D4, `ListenableFuture`.
- **Wyrocznia:** `testdata/continuations` (`scripts/continuations-oracle.sh`,
  `TestContinuationsOracle`, tylko JDK): dwie kontynuacje po
  `supplyAsync`, `thenCompose`, `thenCombine`, nieudany etap
  (`thenApply` pominięte, `exceptionally`/`handle`/`whenComplete`
  wykonane), udany etap (`exceptionally` niewykonane), `runAsync`
  z klasą `Runnable` i `thenRunAsync` na executorze, wywołanie przez
  `CompletionStage`, future nigdy nieukończony, callback przez parametr.
  **Pomiar przed modelem:** lambdy i referencje metod były już osiągane
  (krawędź z `invokedynamic` tworzącego lambdę, DYNAMIC_LINKAGE L0) —
  także przekazana przez parametr; nieosiągalny był tylko `Task#run`
  (obiekt klasy przekazany do `runAsync`). Realny zysk osiągalności
  dotyczy więc obiektów klas przekazanych do kontynuacji; dla lambd model
  dodaje krawędź rejestracji i warunek (wyjaśnienie ścieżki).
- **Zrealizowane:** `internal/callback` — `completionCatalog` (83 wpisy:
  każda metoda `CompletionStage` z callbackiem na obu właścicielach,
  formy `*Async`, indeks 1 dla `thenCombine`/`thenAcceptBoth`/
  `runAfterBoth`/`*Either`, `exceptionallyAsync`/`exceptionallyCompose*`
  tylko na `CompletableFuture`, `supplyAsync`/`runAsync`), pole
  `RegistrationModel.Condition` (`exceptional` dla `exceptionally*`);
  warunek przechodzi przez `CallbackEdge` → `callgraph.CallSite.Condition`
  → `reach.EdgeEvidence.Condition` (`"condition"` w v2, pominięte gdy
  puste; v1 bez zmian) i strzałkę raportu (`→[library_model: callback,
  if exceptional]`).
- **Budżet:** nie wprowadzony — krawędzie są per miejsce rejestracji,
  bez rozwijania łańcucha, więc liczba krawędzi jest liniowa względem
  liczby miejsc; nie ma czego obcinać. Zapisane jako decyzja, nie luka.
- **Wynik:** wyrocznia — 12 z 12 ciał wykonanych przez JVM osiągalne,
  16 krawędzi `callback` dokładnie zgodnych z oczekiwanymi; jedyna luka
  rejestracji to `attach` (callback w parametrze); trzy kontynuacje
  niewykonane przez JVM (normalna gałąź na nieudanym etapie,
  `exceptionally` na udanym, future nigdy nieukończony) zachowują krawędź
  „może” — model nie rozstrzyga ukończenia. Aplikacja kontrolna
  `scripts/prepare-continuations-demo.sh` (commons-text 1.9,
  `thenApply(new Render())`, `-profile generic -roots profile`):
  CVE-2022-42889 `NOT_REACHABLE` (przed) → `REACHABLE` (`main` →
  `Render#apply(Object)` → `apply(String)` → `replace`), JVM uruchamia
  `Render`; wariant bez rejestracji — `NOT_REACHABLE`, zgodnie z JVM.
  Aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/
  `profile` × CHA/RTA, porównanie z `28f1ebb` na tym samym cache offline):
  miejsca rejestracji 435/453/98 → 444/462/100, krawędzie `callback`
  288/287/74 → 297/296/76 (wszystkie nowe w bibliotekach); v1 bajtowo
  identyczne we wszystkich 12 konfiguracjach.
- **Weryfikacja:** `TestContinuationsOracle`, `TestAsyncOracle` (bez
  zmian), `TestCompletionCatalog`, `TestPathEvidenceCondition`;
  `go test ./...`, `go vet ./...`. Dokumentacja: README, SUPPORT.MD.
- **Ograniczenia:** brak analizy ukończenia etapu (kryterium „nieaktywny
  pipeline” z A3 nierozstrzygnięte — krawędź zostaje „może”); listenery,
  reactive, `ListenableFuture`, przekazanie argumentów do taint D4 —
  nietknięte; callback przez parametr/pole — raportowana luka jak w A1;
  przy dwóch rejestracjach tej samej pary metod z różnym warunkiem dowód
  niesie warunek pierwszej.
- **Następny krok (wymaga zgody):** A3 — analiza ukończenia (future bez
  `complete*`), listenery albo wybrana biblioteka reactive; albo inny
  obszar.

### FRAMEWORKS:F4 — zdarzenia aplikacyjne Springa (`spring:events` `ev1`), wycinek 1 — DONE for this scope (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** krawędź z `publishEvent` do
  metod listenerów (`@EventListener`, meta-adnotacje, w tym
  `@TransactionalEventListener`) zarejestrowanych beanów, z bramką twórców
  i dopasowaniem po typie zdarzenia (z dziedziczeniem); `@Scheduled` pod
  `-roots profile`; wyrocznia JVM, aplikacja kontrolna, pomiar na
  aplikacjach referencyjnych.
- **Korekta zakresu po przeglądzie kodu:** `@EventListener` i `@Scheduled`
  już były korzeniami adnotacji profilu `spring` (bez sprawdzania
  rejestracji). Reguła korzeni pozostała bez zmian (żadnego nowego
  wyniku negatywnego); model dodaje ścieżki publikujący → listener i
  domyka to, czego korzenie nie obejmowały: `@TransactionalEventListener`
  (meta-adnotacja, nie korzeń) i `@Schedules` (dwa `@Scheduled` na jednej
  metodzie — w pliku klasy jest tylko kontener `@Repeatable`).
- **Wyrocznia:** `testdata/events` (`scripts/events-oracle.sh`,
  `TestEventsOracle`; jary Spring 5.3.6 z `testdata/tx/deps`, bez kopii):
  listenery na dokładnym typie, nadtypie, `classes=` bez parametru
  (`ContextRefreshedEvent` publikowany przez kontener), na klasie z
  `@Bean`, `@TransactionalEventListener` po commicie, `ApplicationListener`;
  przypadki ujemne: zdarzenie nigdy nie publikowane, klasa niebędąca beanem.
  Przed modelem (bez korzeni adnotacji) graf nie osiągał **5 z 7** ciał
  listenerów (CHA dociera tylko do `ApplicationListener`), a `Ticker#twice`
  (`@Schedules`) nie był korzeniem.
- **Zrealizowane:** `internal/springevents` — katalog listenerów (klasy
  aplikacji; adnotacja `@EventListener` lub meta-adnotacja do głębokości 3;
  typy z `classes`/`value` albo z jedynego parametru), miejsca publikacji
  (`publishEvent(Object|ApplicationEvent)` na podtypach
  `ApplicationEventPublisher` w całym wejściu), typ publikowany z wartości
  lokalnej D2 (`new` dokładny, parametr przez `CallerIndex`, inaczej typ
  statyczny, inaczej typ parametru `publishEvent` — miejsce
  „unbounded”), pominięcie przekazywania własnego parametru przez
  implementację wydawcy, re-publikacja wyniku listenera
  (`ApplicationListenerMethodAdapter`) = typy zwracane listenerów;
  dopasowanie jak w Springu (nadtypy, opakowanie `PayloadApplicationEvent`);
  rejestracja = relacja typu z typem beana (stereotypy, produkty `@Bean`,
  beany XML); bramka: osiągalny twórca `ApplicationListenerMethodAdapter`.
  Podłączenie w `cmd/ariadne` jako bramka punktu stałego B3 (uczestnik
  `spring:events`, klucz snapshotu `spring.events`, linia logu). Profil
  `spring`: `@Schedules` w `roots.annotations`.
- **Wynik:** wyrocznia — wszystkie 7 ciał i oba `@Scheduled` osiągnięte;
  brak krawędzi do listenera nieopublikowanego typu i do klasy niebędącej
  beanem. Jedyna dodatkowa krawędź na fixture: `EventPublicationInterceptor`
  (publikuje `ApplicationEvent` tworzony refleksyjnie → ograniczenie
  statyczne) → listener `ContextRefreshedEvent` — dopuszczalne „może”.
  Aplikacja kontrolna `scripts/prepare-events-demo.sh` (Boot 2.4.3,
  `@TransactionalEventListener` → commons-text 1.9
  `StringSubstitutor#replace`): `-roots profile` CVE-2022-42889
  `NOT_REACHABLE` (przed) → `REACHABLE` (ścieżka `main` →
  `OrderService#place` → listener → `replace`), JVM uruchamia listener;
  wariant publikujący inny typ: `NOT_REACHABLE`, JVM listenera nie
  uruchamia; `-roots all` — `REACHABLE` w obu (każda metoda aplikacji jest
  korzeniem). Aplikacje referencyjne (petclinic, WebGoat, JPetStore; `-roots all`/`profile` × CHA/RTA, porównanie z `main` na tym samym cache offline): żadna nie ma listenera adnotowanego — v1 bajtowo identyczne we wszystkich 12 konfiguracjach; efekt mierzalny tylko na wyroczni i aplikacji kontrolnej.
- **Weryfikacja:** `TestEventsOracle`, `TestFindOnFixture`,
  `TestFindWithoutAdapterCreator`, `TestMatches`, `TestParamTypes`;
  `go test ./...`, `go vet ./...`. Dokumentacja: README, SUPPORT.MD.
- **Ograniczenia:** krawędź „może wykonać”: bez SpEL `condition`, fazy
  transakcji (`@TransactionalEventListener` działa tylko dla zdarzeń
  publikowanych w transakcji), `@Async`, argumentów generycznych zdarzenia;
  bez listenerów klas bibliotecznych, wczesnych zdarzeń Boot (tylko przez
  multicaster) i `addApplicationListener`; rejestracja to relacja typów, nie
  rozwiązana definicja beana (`@Conditional`/`@Profile` nie stosowane).
- **Następny krok (wymaga zgody):** F4 — kolejne wejścia (listenery
  wiadomości: `@JmsListener`/`@KafkaListener`/`@RabbitListener`) albo
  ASYNC_CALLBACKS:A3 (kontynuacje `CompletionStage`).

### FRAMEWORKS — profil `spring`: `spring-tx` jako pakiet frameworka (po `x3`) — DONE (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** dopisać
  `org.springframework:spring-tx` do `framework_packages` profilu `spring`
  (tylko spring-tx, bez spring-context) i zmierzyć wobec `x3`.
- **Powód:** `x3` (prawdziwa krawędź do `TransactionInterceptor#invoke`)
  wpuścił kod aplikacji we wnętrze spring-tx, gdzie CHA sięga daleko poza
  wykonanie: JPetStore `-roots all` CVE-2026-41850 `FRAMEWORK` →
  `REACHABLE` przez `TransactionSynchronization#beforeCommit` → adapter
  `@TransactionalEventListener` (nierejestrowany), przy RTA także
  CVE-2022-41853.
- **Wynik:** petclinic, WebGoat (obie konfiguracje korzeni) i JPetStore
  `-roots profile` — v1 bajtowo identyczne ze stanem sprzed `x3`;
  JPetStore `-roots all` — statusy jak przed `x3` (CVE-2026-41850
  `FRAMEWORK`), zmienione tylko ścieżki `FRAMEWORK` (przez prawdziwy
  interceptor) i liczniki `model_gaps`; RTA — CVE-2022-41853 i
  CVE-2026-41850 znów `FRAMEWORK`. Krawędzie `x3` zostają w grafie.
- **Weryfikacja:** `TestSpringProfileFrameworksSpringTx`; `go test ./...`,
  `go vet ./...`. Dokumentacja: README, SUPPORT.MD.
- **Następny krok (wymaga zgody):** inny obszar.

### FRAMEWORKS:F3 — proxy transakcyjne Springa (`framework:proxies` `x3`, `spring-tx:transactional`) — DONE for this scope (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** wyrocznia JVM, potem model
  wywołań metod transakcyjnych → `TransactionInterceptor#invoke` z bramką
  twórców; pomiar na aplikacjach referencyjnych, każda zmiana werdyktu
  wyjaśniona.
- **Wyrocznia:** `testdata/tx` — prawdziwe jary Spring 5.3.6 (`jcl`,
  `core`, `beans`, `aop`, `context`, `expression`, `tx`),
  `@EnableTransactionManagement`, własny `PlatformTransactionManager`;
  `@Transactional` na klasie (CGLIB), na jednej metodzie, za interfejsem
  (JDK), wyjątek (`rollback`), serwis bez adnotacji.
  `scripts/tx-oracle.sh`, `TestTxOracle`. Przed modelem graf nie osiągał
  **9** ciał (`TransactionInterceptor#invoke`, `$1#proceedWithInvocation`,
  `TransactionAspectSupport` ×4, `getTransaction`/`commit`/`rollback`).
- **Zrealizowane:** wpis `spring-tx:transactional` w
  `internal/proxycatalog` (wersja `x2` → `x3`): metoda transakcyjna, gdy
  `@Transactional` (Spring, `javax`/`jakarta.transaction`) jest na niej,
  na jej klasie albo na nadtypie, który ją deklaruje (lub na nim samym);
  tylko publiczne, niestatyczne, niefinalne; miejsce wywołania, gdy metoda
  na właścicielu lub dowolnym konkretnym podtypie jest transakcyjna; bez
  wywołań z tej samej klasy. Bramka: `new TransactionInterceptor` albo
  klasa nazywająca ją literałem.
- **Wynik:** wyrocznia — wszystkie ciała osiągalne. Kandydaci: petclinic
  20 (repozytoria Spring Data z `@Transactional`), WebGoat 10, JPetStore 4.
  v1: WebGoat bajtowo identyczne; petclinic — 4 ścieżki `REACHABLE`
  (CVE-2021-46877, CVE-2022-45868, CVE-2023-6378, CVE-2026-41850) zaczynają
  się teraz w kontrolerze i idą przez menedżera transakcji, werdykty bez
  zmian; JPetStore `-roots profile` bez zmian, `-roots all` —
  **CVE-2026-41850 `FRAMEWORK` → `REACHABLE`**: pierwszy krok
  (`AccountActionBean#newAccount` → `TransactionInterceptor#invoke`)
  prawdziwy, dalej CHA z `TransactionSynchronization#beforeCommit` do
  adaptera `@TransactionalEventListener`, którego JPetStore nie rejestruje
  — nadmiarowe przybliżenie CHA w spring-tx (zostaje też przy RTA, bo
  adapter tworzy osiągalna fabryka); ścieżki 10 ustaleń `FRAMEWORK` i
  liczniki `model_gaps` (reflection 147 → 172) się zmieniły. RTA
  (`-roots all`): dodatkowo CVE-2022-41853 `FRAMEWORK` → `REACHABLE` przez
  `doBegin` → `JDBCConnection#setAutoCommit` (prawdziwe) i wnętrze HSQLDB
  (CHA).
- **Weryfikacja:** `TestTransactionalEdges` (klasa, metoda JTA,
  interfejs, implementacja; `final`/`static`, wywołanie wewnątrz klasy,
  bean bez adnotacji, brak twórcy), `TestTxOracle`; `go test ./...`,
  `go vet ./...`. Dokumentacja: README, SUPPORT.MD.
- **Następny krok (wymaga zgody):** `org.springframework:spring-tx` (i
  ewentualnie `spring-context`) w `framework_packages` profilu `spring`,
  tak jak spring-beans/core/aop — ścieżka przez wnętrze spring-tx byłaby
  wtedy `FRAMEWORK` do ręcznego przeglądu zamiast `REACHABLE`.

### FRAMEWORKS:F2 — skanowanie z XML Springa (`<mybatis:scan>`, `<context:component-scan>`): pomiar i naprawa ograniczenia — DONE for this scope (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** zmierzyć na JPetStore, co dodaje
  `WEB-INF/*.xml` (mappery z `<mybatis:scan>`, pakiety
  `<context:component-scan>`, beany elementów przestrzeni nazw);
  przekazać `WEB-INF/*.xml` do katalogu proxy niezależnie od wyniku;
  zawężenie `roots.beans` przez `component-scan` i katalog beanów
  przestrzeni nazw — tylko przy zmierzonym efekcie.
- **Pomiar (JPetStore 6.0.2):** 7 interfejsów mapperów, każdy z XML
  mappera — wszystkie 7 już wykrywane po `namespace`; `<mybatis:scan
  base-package="org.mybatis.jpetstore.mapper">` nie dodaje żadnego.
  Klasy ze stereotypem: 3 (`AccountService`, `CatalogService`,
  `OrderService`), wszystkie w skanowanym `org.mybatis.jpetstore.service`
  — zawężenie przez `component-scan` nie zmieniłoby niczego (nie
  zbudowane). `<jdbc:embedded-database>`: `EmbeddedDatabaseFactoryBean#
  afterPropertiesSet`, `EmbeddedDatabaseFactory#initDatabase`,
  `ResourceDatabasePopulator#populate`, konfigurator HSQLDB — już
  osiągalne (CHA). `<tx:annotation-driven>`: `TransactionInterceptor#invoke`
  **nieosiągalny** (proxy transakcyjne wokół serwisów `@Transactional`) —
  model proxy AOP poza tym zakresem, do osobnej propozycji.
- **Zrealizowane:** `proxycatalog.Find(h, jars, extraXML)` czyta też
  `PreparedApp.WebInfBeansXMLs` (skan mapperów, przestrzenie nazw
  mapperów) — usunięte udokumentowane ograniczenie README.
- **Wynik:** petclinic, WebGoat, JPetStore (`-roots profile` i `all`): v1
  **bajtowo identyczne**; linia katalogu proxy JPetStore bez zmian.
- **Weryfikacja:** `TestFindMappersInWebInfXML` (skan tylko w `WEB-INF`,
  pakiet poza skanem, brak pliku), `TestFindMappers`; `go test ./...`,
  `go vet ./...`. Dokumentacja: README, SUPPORT.MD.
- **Następny krok (wymaga zgody):** model proxy transakcyjnego Springa
  (`TransactionInterceptor` wokół beanów `@Transactional`, wyzwalany przez
  `<tx:annotation-driven>`/`@EnableTransactionManagement`) albo inny obszar.

### GRAPH_PRECISION:P3 — wycinek 1 (`p3-1`): zawężanie odbiorników według alokacji — DONE for this scope (2026-11-15); tryb eksperymentalny

- **Zakres (zatwierdzony 2026-11-15):** ograniczone points-to we wspólnej
  domenie wartości D2/D3, opt-in `-callgraph-mode pta` nad RTA, wyrocznia
  JVM z przypadkami z kryteriów „Gotowe”, pomiar z regułą stopu i decyzją
  o trybie.
- **Zrealizowane:** `internal/pta` (wersja `p3-1`, `callgraph.pta` w
  snapshotach, v2 `callgraph.pta`, log „receiver narrowing”); strategia
  `callgraph.PTA` (RTA z własną nazwą trybu), `Graph.PruneCHASites`.
  Dowód typu odbiornika w miejscu CHA osiągalnej metody: `new` w tej samej
  metodzie (klucz alokacji = miejsce `new`), złączenie takich wartości
  (≤ 8 klas), wynik statycznej fabryki, wartość pola `final` (zapisują je
  tylko `<clinit>`/konstruktory), parametr metody statycznej/prywatnej/
  konstruktora, jeśli wszyscy wywołujący są w wejściu (nie korzeń, brak
  krawędzi refleksji/modelu do niej); głębokość ≤ 3; bez kontekstu
  wywołań. Brak dowodu (element tablicy, wynik wywołania wirtualnego,
  pole niefinalne, pole zapisane z parametru konstruktora, klasa, dla
  której wybór metody JVM jest nieznany — tablica lub spoza wejścia) →
  miejsce zachowuje cele RTA. Dowód nie zależy od krawędzi grafu, więc
  jeden przebieg jest punktem stałym.
- **Wyrocznia:** `testdata/pta` (`scripts/pta-oracle.sh`, `TestPTAOracle`):
  dla każdego z 9 miejsc wywołania `Shape#area` RTA daje wszystkie trzy
  cele; PTA zawęża dwa odbiorniki, pole `final`, fabrykę, parametr i
  fabrykę z dwiema alokacjami, zostawia trzy cele dla tablicy, pola
  niefinalnego i pola z parametru konstruktora; każde z 10 wywołań JVM
  pozostaje krawędzią.
- **Pomiar** (`testdata/graphbench/RESULTS.MD`): ponad RTA petclinic
  7 062 / 7 991 krawędzi CHA (3,4 %), WebGoat 9 555 / 12 302 (2,6–2,9 %),
  JPetStore 292 / 299 (1,0 %) przy `-roots profile` / `all`; **0 zmian
  werdyktów**; 2 ścieżki WebGoat (`-roots all`) przeszły na inną prawdziwą
  trasę — usunięte krawędzie `ClearSiteDataHeaderWriter` →
  `MvcRequestMatcher`/`ELRequestMatcher` są artefaktami CHA (pole
  `private final` = `new SecureRequestMatcher()`); koszt 2,7–3,7× czasu
  RTA, +30–45 % pamięci. Dowody: najczęściej pola `final` (~60 %).
- **Decyzja:** tryb zostaje eksperymentalny (opt-in), domyślny CHA bez
  zmian; większa czułość nieuzasadniona.
- **Weryfikacja:** `TestPTAOracle`, `TestPruneCHASites`,
  `TestStrategyForPTA`; `go test ./...`, `go vet ./...`. Dokumentacja:
  README, SUPPORT.MD, plan P3, RESULTS.MD.
- **Następny krok (wymaga zgody):** `<mybatis:scan>`/
  `<context:component-scan>` z XML albo inny obszar.

### REFLECTION:R3 — wycinek 3 (`r3-3`): spis pochodzenia odbiorców i `Method` przez parametry — DONE for this scope (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** spis pochodzenia odbiorców
  osiągalnych nierozwiązanych `Method#invoke`/`Constructor#newInstance` na
  petclinic, WebGoat i JPetStore z regułą stopu (0 miejsc do zamiany →
  tylko spis); przy zysku — implementacja wyłącznie zmierzonych kształtów
  z wyrocznią JVM.
- **Spis (przed zmianą, `-roots all`):** petclinic 510 miejsc — wartość
  nieśledzona (głównie wynik wywołania wirtualnego, element tablicy) 216,
  pole 206 (najczęściej trzymające wartość typowaną albo nieśledzoną),
  wynik wywołania statycznego 45, parametr 37, zbiór 6; WebGoat 537
  (227/207/51/46/6); JPetStore 147 (62/53/12/17/2 + 1 wynik wywołania
  instancyjnego). Wśród parametrów, po jednym kroku przez pole/helper u
  wywołujących: **2** miejsca do zamiany w petclinic i 2 w WebGoat, 0 w
  JPetStore — Thymeleaf `BinaryOperationExpression#
  doComposeBinaryOperationExpression` (parametry `Method` 4 i 5): 7
  wywołujących, każdy przekazuje pole statyczne `getDeclaredMethod(
  "isLeftAllowed"/"isRightAllowed", …)` jednej z 6 podklas; zbiór
  kompletny. Poprzednie pomiary (0 z 619 / 0 z 37) nie sprawdzały kroku
  parametr → pole.
- **Zrealizowane:** `reflectinvoke.ResolveThroughParameters` (wersja
  `r3-3`, `reflection.parameters` w snapshotach): odbiorca-parametr →
  `dataflow.CallerIndex.ResolveParameter` → każda wartość wprost albo o
  jeden krok (pole, wynik statycznego helpera) → `KindReflectiveMethod`
  → osobny `ResolvedInvoke` (`ThroughParameter`) z celem wywołania i
  tablicą argumentów odczytanymi w miejscu; komplet → wpis nierozwiązany
  zastąpiony (luka znika), część → luka zostaje
  (`PartlyThroughParameter`). Krawędzie przez istniejące
  `FindInvokeEdges` (basis `reflection`). `Constructor#newInstance` —
  0 zmierzonych miejsc, bez zmian. Spis zostaje jako funkcja raportowa
  (`reflectinvoke.Provenance`/`SummarizeProvenance`; log „reflection
  census”, sekcja raportu, v2 `coverage.reflection_census`).
- **Wynik:** petclinic i WebGoat +12 krawędzi (Thymeleaf
  `isLeftAllowed`/`isRightAllowed`), luki `reflection` 510 → 508 i
  536/537 → 534/535; v1 różni się **tylko** `model_gaps`, żaden werdykt
  się nie zmienił (obie konfiguracje korzeni). JPetStore: 1 miejsce
  częściowe, 3 tożsamości, 0 krawędzi; v1 bajtowo identyczne. Spis po
  zmianie: 0 miejsc do zamiany na trzech aplikacjach.
- **Wyrocznia:** `testdata/reflectparam` (`scripts/reflectparam-oracle.sh`,
  `TestReflectParamOracle`): pola statyczne przekazywane do helpera (kształt
  Thymeleaf), wyszukanie w miejscu wywołania, cel instancyjny, helper z
  jednym wywołującym o nazwie z czasu działania (luka częściowa), `Method`
  nieprzekazany (ujemny). `main` sprzed zmiany: 4 z 10 ciał; po: 9 z 10
  (dziesiąte — nazwa z czasu działania, udokumentowana luka).
- **Weryfikacja:** `TestResolveThroughParameters` (zbiory celów, luka
  częściowa, cel statyczny, spis), `TestReflectParamOracle`; `go test
  ./...`, `go vet ./...`. Dokumentacja: README, SUPPORT.MD.
- **Ograniczenia:** jeden krok przez pole/helper u wywołującego;
  wywołujący tylko statycznie dokładni (`CallerIndex`); wartości
  nieśledzone przez D2 (wyniki wywołań wirtualnych, elementy tablic)
  pozostają lukami — to większość spisu i zakres P3/R4, nie R3.
- **Następny krok (wymaga zgody):** GRAPH_PRECISION P3 (ograniczone
  points-to — największa kategoria spisu to wartości nieśledzone) albo
  `<mybatis:scan>`/`<context:component-scan>` z XML.

### ROADMAP:B0 — przyrost `b0-5`: gdzie są nieograniczone miejsca luk — DONE for this scope (2026-11-15); katalog dyspozytorów odrzucony pomiarem

- **Zakres (zatwierdzony 2026-11-15):** zmierzyć rozkład osiągalnych
  nieograniczonych miejsc reguły luk i — jeśli skupiają się w kilku
  dyspozytorach refleksyjnych frameworków — ograniczyć je katalogiem
  (zbiory `Owners` z modeli `roots.beans`, `x2`, `jpa:objects`, tras MVC,
  Stripes). Klauzula ryzyka planu: przy długim ogonie wynikiem jest
  udokumentowany pomiar i tylko wpisy, które się bronią.
- **Pomiar (domyślne przebiegi, obie konfiguracje korzeni):** petclinic
  611 miejsc (refleksja 510, deserializacja 99, `MethodHandle` 2) w **477**
  różnych metodach, 40 komponentów (Tomcat 98, Byte Buddy 88, Hibernate 79,
  spring-core 36, H2 33, Jackson 25, …); WebGoat 681 w 549 (55
  komponentów; Byte Buddy, Hibernate, JRuby, XStream, …); JPetStore 161 w
  129 (14; MyBatis, log4j 1.2, Spring, HSQLDB, Stripes). 15 najczęstszych
  metod: 82 / 83 / 44 miejsc. Planowane dyspozytory (tworzenie i cykl
  życia beanów Springa, `RepositoryMethodInvoker`, handlery Spring ORM,
  `InvocableHandlerMethod`, `ApplicationListenerMethodAdapter`, Stripes,
  `ReflectionUtils`/`MethodInvoker`): **25 / 25 / 19** miejsc (4 %, 4 %,
  12 %).
- **Decyzja:** katalog nie zbudowany — ograniczenie wszystkich tych miejsc
  zostawiłoby każdemu `NOT_REACHABLE` setki istotnych nieograniczonych
  miejsc, więc nie zmieniłoby żadnego oznaczenia reguły (kod bez
  użytkownika). Selektywność reguły wymaga ogólnego ograniczenia
  nierozwiązanej refleksji (REFLECTION R3/R4, GRAPH_PRECISION P3), nie
  katalogu. Zapis w kontrakcie §4.
- **Zrealizowane (report-only):** `reach.UnboundedSummary` (wersja
  `b0-5`): miejsca, różne metody, rodzaje, grupy po komponencie klasy
  wołającej (GAV), najczęstsze metody; log „unbounded gap sites”, sekcja
  raportu „Unbounded gap sites”, v2 `coverage.unbounded_sites`.
- **Weryfikacja:** `TestUnboundedSummary` (tylko osiągalne i
  nieograniczone, grupowanie po komponencie i metodzie, `TopN`); petclinic,
  WebGoat, JPetStore (`-roots profile` i `all`): v1 **bajtowo
  identyczne** z `main`; `go test ./...`, `go vet ./...`. Dokumentacja:
  kontrakt §4, README, SUPPORT.MD.
- **Następny krok (wymaga zgody):** REFLECTION R3 (dotąd 0 z 37
  zmierzonych relacji miało kształt R3 — ponowny pomiar na nieograniczonych
  miejscach z tej sekcji) albo GRAPH_PRECISION P3; ewentualnie
  `<mybatis:scan>`/`<context:component-scan>` z XML.

### ROADMAP:B4 — przyrost `b4-8`: wyrocznia JVM dla beanów z XML Springa, korzenie XML w `roots.beans` — DONE (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** cykl życia beanów z definicji XML
  (konstruktor, settery właściwości, `init-method`/`destroy-method`),
  także dla klas bibliotecznych; czytanie XML z `WEB-INF/` WAR-a.
- **Wyrocznia:** `testdata/xmlbeans` — prawdziwe jary Spring 5.3.6 (`jcl`,
  `core`, `beans`, `aop`, `context`, `expression`) z zależności petclinic
  bez `META-INF/maven`, własny jar biblioteczny `deps/fx-lib.jar`
  (`LibService` jak `SqlSessionFactoryBean` w JPetStore),
  `ClassPathXmlApplicationContext` nad `fx/app/beans.xml`:
  `<constructor-arg>`, `c:`, `<property>`, `p:` (wartość i `-ref`), bean
  wewnętrzny, `init-method`/`destroy-method`, `default-init-method`,
  definicja `parent` z abstrakcyjnym rodzicem, ujemne: klasa nazwana tylko
  w definicji abstrakcyjnej, klasa bez definicji, niewybrany konstruktor;
  nierozstrzygnięte: `lazy-init`. `scripts/xmlbeans-oracle.sh`,
  `TestXMLBeansOracle`.
- **Wynik wyroczni przed poprawką:** JVM wykonał 18 ciał; graf osiągał
  **1** (`main`) — konstruktory, settery i metody init/destroy Spring woła
  refleksyjnie; dla klasy bibliotecznej nie pomogłoby nawet `-roots all`.
- **Poprawka:** `springxml` czyta właściwości (`<property>`, `p:`),
  argumenty konstruktora (`<constructor-arg>`, `c:`), `init-method`/
  `destroy-method` i domyślne z `<beans>`, `parent`/`abstract`/
  `lazy-init`, `factory-method`/`factory-bean`, beany wewnętrzne (na
  dowolnej głębokości) i zagnieżdżone `<beans>`. `input` zbiera
  `WEB-INF/*.xml` z korzeniem `<beans>` z WAR-a i modułów EAR
  (`PreparedApp.WebInfBeansXMLs`). `profile.XMLBeanLifecycleRoots`
  (część `roots.beans`, oba scenariusze): dla nieabstrakcyjnej definicji
  po scaleniu łańcucha `parent`, której klasa jest w wejściu (aplikacja
  lub biblioteka) — konstruktory o liczbie parametrów równej liczbie
  argumentów, jednoparametrowe settery właściwości (też w nadklasach;
  ścieżki zagnieżdżone pominięte), `init-method` albo domyślna,
  `destroy-method` (`(inferred)`: `close`/`shutdown`) albo domyślna;
  definicja z `factory-method` — metody fabryki (statyczne albo na klasie
  `factory-bean`), bez produktu.
- **Wynik:** wyrocznia 18 z 18; ujemne nieosiągalne, `lazy-init` celowo
  osiągalny. Aplikacje referencyjne: **żaden werdykt v1 się nie
  zmienił**; petclinic i WebGoat bez XML — bez zmian korzeni; JPetStore
  czyta `WEB-INF/applicationContext.xml` po raz pierwszy (2 beany):
  konstruktory i settery `SqlSessionFactoryBean`/
  `DataSourceTransactionManager` (wcześniej nieosiągalne z aplikacji i
  korzeni `web.xml`) są korzeniami — 8 nowych przy `-roots profile`, 5 przy
  `-roots all`.
- **Weryfikacja:** `TestParse_lifecycle`, `TestParse_fullShape`
  (`internal/springxml`), `TestPrepareAppInput_capturesWebInfBeansXML`
  (WAR i EAR), `TestXMLBeanLifecycleRoots` (rodzic, fabryka statyczna i
  instancyjna, `(inferred)`, domyślna metoda bez ciała, ścieżka
  zagnieżdżona, klasa abstrakcyjna, brak klasy), `TestXMLBeansOracle`;
  `go test ./...`, `go vet ./...`. Dokumentacja: README, FAQ.md, PLAN.md,
  SUPPORT.MD.
- **Ograniczenia:** elementy przestrzeni nazw (`<jdbc:…>`, `<tx:…>`,
  `<mybatis:scan>`, `<context:component-scan>`), pliki spoza ścieżki klas i
  `WEB-INF/`, wybór przeciążenia konstruktora po typach (liczy się każdy o
  właściwej liczbie parametrów); katalog proxy MyBatis nadal nie czyta
  `<mybatis:scan>` z `WEB-INF` (mappery JPetStore wykrywa po XML mapperów).
- **Następny krok (wymaga zgody):** B0 — zawężenie nieograniczonych miejsc
  refleksji, albo `<mybatis:scan>`/`<context:component-scan>` z XML.

### ROADMAP:B4 — przyrost `b4-7`: wyrocznia JVM dla DI i cyklu życia beanów Springa, korzenie `roots.beans` — DONE (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** wstrzykiwanie zależności i cykl
  życia beanów Springa na prawdziwym JVM; ograniczoną lukę naprawić w tym
  samym przyroście.
- **Wyrocznia:** `testdata/di` — prawdziwe jary Spring 5.3.6 (`jcl`,
  `core`, `beans`, `aop`, `context`, `expression`) i API JSR-250
  (`jakarta.annotation-api` 1.3.5) z zależności petclinic, bez
  `META-INF/maven`, SHA-256 oryginałów w `deps/ORIGINALS.sha256`;
  `AnnotationConfigApplicationContext` ze skanowaniem `fx.beans`:
  wstrzykiwanie przez konstruktor, pole i setter, `@PostConstruct`/
  `@PreDestroy`, `@Bean(initMethod, destroyMethod)`, `BeanNameAware`,
  `ApplicationContextAware`, `InitializingBean`, `DisposableBean`,
  `SmartInitializingSingleton`, `FactoryBean`, `BeanPostProcessor`,
  `BeanFactoryPostProcessor`, `ApplicationListener`, `@EventListener`,
  `Condition`; ujemne: nieaktywny `@Profile`, niespełniony `@Conditional`,
  nieżądany `@Lazy`, klasa bez stereotypu. `scripts/di-oracle.sh`,
  `TestDIOracle`.
- **Wynik wyroczni:** JVM wykonał 21 ciał aplikacji, żadnego z ujemnych.
  Przy `-roots profile` graf nie osiągał **7**: obu konstruktorów beanów,
  settera `@Autowired`, `@PostConstruct`, `@PreDestroy`, `initMethod` i
  `destroyMethod` z `@Bean` — wszystkie wołane przez `Constructor#newInstance`/
  `Method#invoke` bez literału. Interfejsy callbacków CHA już osiągała.
- **Poprawka:** rodzaj korzenia profilu `roots.beans` (`spring.yaml`;
  `profile.BeanLifecycleRoots`, w obu scenariuszach korzeni jak korzenie
  `web.xml`): dla konkretnych klas aplikacji ze stereotypem,
  `@Configuration` albo `@ControllerAdvice`/`@RestControllerAdvice` —
  konstruktor wybierany przez Springa (jedyny, inaczej `@Autowired`/
  `@Inject`, inaczej bezargumentowy), niestatyczne metody `@Autowired`/
  `@Inject`/`@Resource`/`@Value`, `@PostConstruct`/`@PreDestroy` (javax i
  jakarta, także w nadklasach aplikacji); dla metod `@Bean` — `initMethod`/
  `destroyMethod` na klasie produktu (zadeklarowany konkretny typ zwracany
  i podtypy tworzone `new` w metodzie fabrykującej), przy domyślnym
  `destroyMethod` wywnioskowane `close()`, inaczej `shutdown()`. Bean z
  `@Profile` nieaktywnym według `-active-profiles` pominięty; `@Profile`
  bez listy, `@Conditional`, `@Lazy` — liczone („może powstać”, reguła
  nie zgaduje fałszu). Log `spring bean lifecycle roots: …`.
- **Wynik:** wyrocznia (z `-active-profiles default`): 21 z 21 ciał
  osiągalnych; `Dormant` (nieaktywny profil) i `Plain` nie; `Unmet` i
  `Sleepy` osiągalne celowo (przypadki nierozstrzygnięte, test to
  sprawdza). Aplikacje referencyjne: **żaden werdykt v1 się nie zmienił**;
  petclinic 8 metod (7 nowych korzeni przy `-roots profile`), JPetStore 3,
  WebGoat 165 (154 konstruktory, 6 `@PostConstruct`, 1 `@PreDestroy`,
  1 metoda wstrzykiwania, 3 `@Bean` init/destroy; przy `-roots all` 3 nowe
  korzenie biblioteczne). WebGoat `-roots profile`: dwie ścieżki
  `REACHABLE` zaczynają się teraz od korzeni cyklu życia (CVE-2022-22976
  od `WebSecurityConfig#configureGlobal` z `@Autowired`, CVE-2022-41853 od
  `HSQLDBDatabaseConfig#shutdown` z `@PreDestroy`) — obie prawdziwe.
- **Weryfikacja:** `TestBeanLifecycleRoots` (wybór konstruktora: jedyny,
  `@Autowired`, bezargumentowy, brak; metody nadklasy; metoda statyczna
  wykluczona; `@Bean` jawne i wywnioskowane, `destroyMethod=""`; klasa bez
  stereotypu, abstrakcyjna, biblioteczna; `@Profile` nieznany/nieaktywny/
  aktywny), `TestSpringProfileHasBeanRoots`, `TestDIOracle`; `go test
  ./...`, `go vet ./...`. Dokumentacja: README, FAQ.md, PLAN.md (faza 7),
  SUPPORT.MD.
- **Ograniczenia:** beany z XML Springa (`init-method`, argumenty
  konstruktora), metody `@Bean` poza `@Configuration` (tryb lite), własne
  stereotypy (meta-adnotacja `@Component`), warunki `@Conditional`/Boot nie
  są oceniane.
- **Następny krok (wymaga zgody):** B0 — zawężenie nieograniczonych miejsc
  refleksji, albo beany z XML Springa (JPetStore).

### ROADMAP:B4 — przyrost `b4-6`: wyrocznia JVM dla proxy JPA Spring ORM i model `jpa:objects` `e1` — DONE (2026-11-15)

- **Zakres (zatwierdzony 2026-11-15):** wpisy `spring-orm:*` katalogu
  `framework:proxies` (współdzielony i rozszerzony `EntityManager`,
  odroczone zapytanie, zarządzana fabryka) na prawdziwym JVM; ograniczoną
  lukę naprawić w tym samym przyroście.
- **Wyrocznia:** `testdata/jpa` — prawdziwe jary Spring 5.3.6 (`jcl`,
  `core`, `beans`, `tx`, `orm`) i API JPA 2.2 (`jakarta.persistence-api`
  2.2.3, pakiety `javax.persistence`) z zależności petclinic, bez
  `META-INF/maven`, SHA-256 oryginałów w `deps/ORIGINALS.sha256`; zamiast
  Hibernate'a minimalny dostawca w fixture (`EntityManagerFactory`,
  `EntityManager`, `Query`, `EntityTransaction`; ciała zapisują cały stos),
  podklasa `AbstractEntityManagerFactoryBean`, `OrphanEntityManager`
  (nikt nie tworzy); `scripts/jpa-oracle.sh`, `TestJPAOracle`.
- **Wynik wyroczni:** JVM wykonał wszystkie cztery handlery i za nimi
  `FxEntityManager#find`/`createQuery`, `FxQuery#getResultList`,
  `FxEntityManagerFactory#createEntityManager`. Graf osiągał handlery (wpisy
  katalogu poprawne), ale **żadnego z czterech ciał dostawcy**: handler
  woła cel przez `Method#invoke`, a wywołania na `javax.persistence` nie są
  rozwijane przez CHA (reguła pomijania).
- **Poprawka:** model `jpa:objects` (`internal/jpadispatch`, wersja `e1`)
  — ta sama bramka twórców co `jdbc:objects` (`callgraph.ObjectEdges` +
  `NewCreatorGate`, krok punktu stałego B3, uczestnik w katalogu
  `models.disable`, `jpa.objects` w snapshotach) dla `javax.persistence`
  `EntityManager`, `EntityManagerFactory`, `Query`, `TypedQuery`,
  `StoredProcedureQuery`, `EntityTransaction`. `jakarta.persistence`
  pominięte celowo: nie podlega regule pomijania, CHA już je rozwija
  (test pilnuje obu faktów). Wyrocznia po poprawce: 12 z 12 ciał wejścia
  osiągalnych, krawędzie `model` prosto z miejsc wywołań w `Main`, brak
  krawędzi do `OrphanEntityManager`; bez modelu test zgłasza 9 błędów.
- **Aplikacje referencyjne** (CHA, `-roots profile` i `all`; RTA
  `-roots profile`): v1 **bajtowo identyczne** na petclinic, WebGoat,
  JPetStore. petclinic +139, WebGoat +138 krawędzi (zapytania Hibernate
  5.4, `SessionImpl#find`/`remove`/`flush`/`contains`, …); część dostępu
  do produktów Hibernate w osiągalnych miejscach dawał już
  `service-provider:dispatch` (rejestracja `PersistenceProvider`). Log
  przypisuje 381 krawędzi nowej bramce: 139 nowych i 242 wcześniej
  przypisane modelom B3 (3 620 → 3 403) i `jdbc:objects` (13 583 →
  13 558) — te same krawędzie, inna kolejność atrybucji. JPetStore 0
  kandydatów (brak JPA).
- **Weryfikacja:** `TestObjectEdgesGatedByCreator` (łańcuch fabryka →
  `EntityManager` → zapytanie jako punkt stały, wywołanie przez `Query` i
  `TypedQuery`, brak krawędzi bez bootstrapu dostawcy, sierota),
  `TestTypesAreSkipped`, `TestJPAOracle`; `go test ./...`, `go vet ./...`.
  Dokumentacja: README, FAQ.md, SUPPORT.MD.
- **Ograniczenia:** bramka nie rozróżnia kontekstu (każdy utworzony
  `EntityManager` jest celem każdego wywołania); klasa dostawcy tworzona
  tylko refleksyjnie (`persistence.xml`) nie ma twórcy; fragmenty
  repozytoriów Spring Data — bez zmian.
- **Następny krok (wymaga zgody):** wyrocznia DI (ostatni obszar B4 bez
  wyroczni) albo B0 — zawężenie nieograniczonych miejsc refleksji.

### FRAMEWORKS — profil `stripes` (`roots.handlers`) po śladzie JPetStore — DONE (2026-11-14)

- **Zakres (zatwierdzony 2026-11-14):** zamknąć fałszywy negatyw werdyktu
  znaleziony przez ślad: Stripes woła handlery zdarzeń action beanów przez
  `Method#invoke`, więc przy `-roots profile` nic pod action beanem nie
  było osiągalne.
- **Zrealizowane:** nowy rodzaj korzenia profilu `roots.handlers`
  (`implements` + `returns`): publiczne, niestatyczne, nieabstrakcyjne
  metody bez parametrów klas aplikacji implementujących typ, zwracające
  `returns` albo jego podtyp (konwencja Stripes — handler nie ma stałej
  nazwy ani wymaganej adnotacji). Profil `profiles/stripes.yaml`
  (wykrywany po `ActionBean`): handlery, `@Before`/`@After`/
  `@ValidationMethod`, `ActionBean#setContext`, pakiet frameworka
  `net.sourceforge.stripes:*`.
- **Wynik:** JPetStore wykrywa `spring+stripes`, 28 punktów wejścia (było
  6). Na tym samym śladzie: CVE-2020-26945 `NOT_REACHABLE` → **`REACHABLE`**
  przy `-roots profile` (ścieżka z `AccountActionBean#newAccount`), zgodnie
  z 884 wykonaniami `SerializedCache#deserialize`; CVE-2022-41853
  `FRAMEWORK` → `REACHABLE` (jak przy `-roots all`); pozostałe zmiany —
  tylko liczniki `model_gaps`. JPetStore `-roots all`, petclinic, WebGoat:
  v1 bajtowo identyczne. Relacje śladu bez zmian (korzenie nie dodają
  krawędzi).
- **Weryfikacja:** `TestHandlerRoots` (podtyp `Resolution`, metoda w klasie
  abstrakcyjnej, wykluczone: parametry, prywatna, statyczna, abstrakcyjna,
  inny typ zwracany, klasa biblioteki), `TestStripesProfile`; `go test
  ./...`, `go vet ./...`. Dokumentacja: README, PLAN.md (faza 7), FAQ.md,
  SUPPORT.MD, TRACE_GAPS.MD.
- **Ograniczenia:** wiązanie parametrów żądania z setterami action beana
  (też refleksja Stripes) nie jest modelowane; `@HandlesEvent` na metodzie
  z parametrami — poza konwencją.
- **Następny krok (wymaga zgody):** dokończenie B4 (DI, JPA) albo B0 —
  zawężenie nieograniczonych miejsc refleksji.

### ROADMAP:B4 — ślady JVM po wyroczniach (2026-11-14): pomiar, bez zmian kodu

- **Zakres (zatwierdzony 2026-11-14):** nagrać od nowa petclinic i
  JPetStore (`scripts/runtime-trace-*.sh`, bez zmian) i przeanalizować oba
  ślady analizatorem sprzed dzisiejszych poprawek (`8287fba`) i po nich
  (`c43976e`). Szczegóły: `testdata/graphbench/TRACE_GAPS.MD`, sekcja „Re-recorded
  after the B4 oracles”.
- **Wynik 1 — relacje:** petclinic 121 z 2 736 poza grafem (4,4 %),
  JPetStore 38 z 764 (5,0 %) — **identycznie przed i po** poprawkach.
  `x2` dodaje krawędzie z miejsca wywołania repozytorium (osiągalność),
  a zmierzona relacja startuje w `RepositoryMethodInvoker` (refleksja);
  `p2` — żadna z aplikacji nie łączy się przez `DriverManager`; `j2`/`p3`
  — relacje statementów HSQLDB nie trafiły na próbkowane stosy. Miara
  relacji nie nadaje się do oceny modeli skracających krok refleksyjny.
- **Wynik 2 — fałszywy negatyw werdyktu:** JPetStore CVE-2020-26945
  (MyBatis, `SerializedCache#deserialize` wykonane 884 razy według
  dokładnych wejść) ma przy `-roots profile` werdykt `NOT_REACHABLE`
  (przed i po), przy `-roots all` — `REACHABLE`. Przyczyna zmierzona
  łańcuchem relacji śladu: Stripes woła handler zdarzenia action beana
  (`AccountActionBean#signon`) przez `Method#invoke`, więc z korzeni
  `web.xml` nic pod action beanem nie jest osiągalne. Pozostałe
  obserwowane metody podatne obu aplikacji mają werdykty pozytywne.
- **Następny krok:** model Stripes — zrobiony tego samego dnia (profil
  `stripes`, wpis powyżej).

### ROADMAP:B4 — poprawka po `b4-5`: bazowe klasy repozytoriów Spring Data (`framework:proxies` `x2`) — DONE (2026-11-14)

- **Zakres (zatwierdzony 2026-11-14):** zamknąć lukę znalezioną przez
  wyrocznię Spring Data: ostatni krok proxy (`RepositoryMethodInvoker` →
  `Method#invoke` → metoda bazowej klasy) był refleksją bez literału.
- **Zrealizowane:** wpis `spring-data:base-repository` — wywołanie na
  interfejsie repozytorium aplikacji (z katalogu) → ciało, które wykonuje
  dla tej metody każda bazowa klasa (konkretny podtyp `Repository` w
  wejściu). Dopasowanie jak w Spring Data: dokładny deskryptor, inaczej
  nazwa i przypisywalne parametry (zatarty `Object` przyjmuje referencję),
  bez typu zwracanego; pominięte metody z adnotacją `*.Query` i metody
  `default`. Bramka twórców: metoda robiąca `new` bazowej klasy albo każda
  metoda klasy, która nazywa ją literałem (`getRepositoryBaseClass`
  zwracające `SimpleJpaRepository.class` — Spring tworzy ją refleksyjnie).
  Krawędzie z miejsca wywołania repozytorium (jak krawędź do
  `JdkDynamicAopProxy#invoke`), basis `model`. Wersja `x1` → `x2`.
- **Wynik:** `testdata/springdata` 16 z 16 (było 15, i `save`/`findById`
  tylko przypadkiem); `OrphanRepository` (nikt nie tworzy ani nie nazywa) bez
  krawędzi. Pierwsza wersja (tylko dokładny deskryptor) dała petclinic 0
  krawędzi — repozytoria petclinic deklarują metody z konkretnymi typami
  (`void save(Owner)`) — stąd dopasowanie po przypisywalności. Aplikacje
  referencyjne: werdykty v1 bez zmian; petclinic +9 krawędzi (kontrolery →
  `SimpleJpaRepository#save/findById/findAll`), ścieżka CVE-2026-41850
  (nadal `REACHABLE`) wchodzi teraz z `OwnerController#processCreationForm`
  zamiast z syntetycznego mostu profilu; WebGoat +7, v1 bajtowo
  identyczne; JPetStore 0 (MyBatis), bajtowo identyczne.
- **Weryfikacja:** `TestBaseRepositoryEdges` (`new`, literał, sierota,
  metoda zapytania, `@Query`, `default`, dopasowanie `save(Pet)` →
  `save(Object)`), `TestSpringDataOracle`, `TestMyBatisOracle`; `go test
  ./...`, `go vet ./...`.
- **Ograniczenia:** fragmenty repozytoriów (`…RepositoryImpl`, własne
  interfejsy fragmentów) — ten sam krok refleksyjny, nie modelowany.
- **Następny krok (wymaga zgody):** wyrocznia JPA `EntityManager` albo DI.

### ROADMAP:B4 — przyrost `b4-5`: wyrocznia JVM dla proxy Spring Data — DONE for this scope (2026-11-14); znaleziona luka

- **Zakres (zatwierdzony 2026-11-14):** repozytoria Spring Data
  (`framework:proxies`, wpis `spring-data:repository`) na prawdziwym JVM.
  Bez JPA: dostawca JPA to kilkanaście MB; zamiast niego minimalny moduł
  magazynu w fixture (podklasa `RepositoryFactorySupport` i bazowa klasa
  repozytorium) — Spring buduje proxy tą samą drogą co dla JPA.
- **Zrealizowane:** `testdata/springdata` — prawdziwe jary Spring 5.3.6
  (`jcl`, `core`, `beans`, `aop`, `context`, `expression`, `tx`) i Spring
  Data Commons 2.4.8 z zależności petclinic (5,5 MB; bez `META-INF/maven`,
  SHA-256 oryginałów w `deps/ORIGINALS.sha256`); bazowa klasa zapisuje
  cały stos; `scripts/springdata-oracle.sh`, `TestSpringDataOracle`.
- **Wynik:** 16 ciał z wejścia na stosach; krawędź `framework:proxies` z
  wywołania repozytorium do `JdkDynamicAopProxy#invoke` jest, cały łańcuch
  interceptorów (`ReflectiveMethodInvocation`,
  `QueryExecutorMethodInterceptor`, `RepositoryComposition`,
  `RepositoryMethodInvoker`, …) osiągalny. **Luka:** ostatni krok —
  `RepositoryMethodInvoker` → `Method#invoke` → metoda bazowej klasy — to
  refleksja bez literału (0 z 96 miejsc `Method#invoke` rozwiązanych):
  `count` nieosiągalny, `save`/`findById` tylko przypadkiem (krawędzie CHA z
  `CrudRepositoryInvoker`, którego JVM tu nie wykonuje). Most `spring` →
  `SimpleJpaRepository` obejmuje ten krok tylko dla JPA i nie jest tu
  sprawdzany.
- **Weryfikacja:** test utrwala lukę (padnie, gdy `count` stanie się
  osiągalny albo pojawi się krawędź z `core.support` do bazowej klasy);
  `go test ./...`, `go vet ./...`.
- **Następny krok (wymaga zgody):** model kroku „repozytorium → bazowa
  klasa”: wywołanie metody interfejsu repozytorium (z katalogu) → ta sama
  metoda w klasach bazowych repozytoriów, np. z bramką twórców (klasa
  bazowa tworzona przez osiągalny `getTargetRepository`), z pomiarem na
  petclinic (JPA); albo wyrocznia DI.

### ROADMAP:B4 — poprawka po `b4-4`: nadinterfejsy JDK typów JDBC (`jdkapi` `k1`, `jdbc:objects` `j2`, `service-provider:dispatch` `p3`) — DONE (2026-11-14)

- **Zakres (zatwierdzony 2026-11-14):** zamknąć lukę znalezioną przez
  wyrocznię MyBatis: hierarchia nie czyta JDK, więc nie wie, że
  `PreparedStatement` rozszerza `Statement`, a klasa implementująca tylko
  podinterfejs nie jest podtypem nadinterfejsu.
- **Zrealizowane:** `internal/jdkapi` — tabela bezpośrednich nadinterfejsów
  `java.sql`/`javax.sql` odczytana refleksją z JDK 25 (stała od JDBC 4.1),
  `Subinterfaces`, `Subtypes` (podtypy z wejścia dla interfejsu i każdego
  jego podinterfejsu JDK). Używają jej `callgraph.ObjectEdges`
  (`jdbc:objects`; model servletów dzieli funkcję, ale tabela nie ma typów
  servletów) i produkty `service-provider:dispatch`.
- **Wynik:** `testdata/mybatis` 51 z 51 (było 50); `testdata/jdbc` bez
  zmian. **Korekta oceny z `b4-4`:** ten kształt ma prawdziwy sterownik —
  HSQLDB `JDBCPreparedStatement`/`JDBCCallableStatement` implementują tylko
  podinterfejsy, H2 `JdbcClob` — `NClob`. Aplikacje referencyjne: v1 —
  werdykty bez zmian; petclinic i WebGoat bajtowo identyczne, JPetStore:
  zmieniona tylko ścieżka CVE-2022-41853 (nadal `REACHABLE`) — teraz
  przez `BaseExecutor#closeStatement` → `JDBCPreparedStatement#close`, tak
  jak MyBatis zamyka statement. Graf: +295 (petclinic: H2, Hibernate
  `MaterializedNClob`), +526 (WebGoat: 391 do statementów HSQLDB,
  PostgreSQL), +102 (JPetStore: HSQLDB).
- **Weryfikacja:** `TestSubinterfaces`, `TestSubtypes`, `TestMyBatisOracle`
  (teraz bez wyjątków), `TestJDBCOracle`, `TestSPIDispatchOracle`;
  `go test ./...`, `go vet ./...`.
- **Następny krok (wymaga zgody):** wyrocznia proxy Spring Data/JPA albo DI.

### ROADMAP:B4 — przyrost `b4-4`: wyrocznia JVM dla proxy MyBatis — DONE for this scope (2026-11-14); znaleziona luka

- **Zakres (zatwierdzony 2026-11-14):** proxy frameworków
  (`framework:proxies`) na prawdziwym JVM, zaczynając od MyBatis (jar był
  w lokalnym repo Mavena). Spring Data/JPA — osobno.
- **Zrealizowane:** `testdata/mybatis` — prawdziwy `mybatis-3.4.1.jar`
  (Apache 2.0; skopiowany bez `META-INF/maven`, żeby test offline nie
  pytał OSV — klasy bajtowo identyczne, SHA-256 oryginału w
  `deps/mybatis-3.4.1.sha256`), sterownik JDBC, którego ciała zapisują
  **cały stos** (`StackWalker`), więc wyrocznia obejmuje też ciała MyBatis
  wykonane przez JVM; `app.jar` woła metodę mappera `@Mapper`/`@Select`
  przez `UnpooledDataSource` (czyli `DriverManager`) i metodę proxy JDK
  interfejsu niebędącego mapperem z własnym handlerem;
  `scripts/mybatis-oracle.sh`, `TestMyBatisOracle`.
- **Wynik:** 51 ciał z wejścia na stosach (39 MyBatis); osiągalnych 50 —
  mapper przez `framework:proxies` do `MapperProxy#invoke`, dalej
  `MapperMethod`, executor, handlery, `JdbcTransaction`, sterownik przez
  fasadę `DriverManager` (`p2`). Interfejs bez `@Mapper` nie jest łączony z
  `MapperProxy`; jego handler (lambda) osiągalny przez krawędź utworzenia
  lambdy. **Luka:** `FxStatement#close` — MyBatis zamyka przez
  `java.sql.Statement#close`, a statement sterownika implementuje tylko
  `PreparedStatement`; dziedziczenie interfejsów JDK nie jest w hierarchii,
  więc `jdbc:objects` (i produkty `service-provider:dispatch`) nie widzą
  implementacji `Statement`. Prawdziwe sterowniki (H2, Hikari) implementują
  `Statement` zwykle po linii klas, więc wpływ jest pewnie mały (ocena
  błędna — patrz poprawka powyżej: HSQLDB ma ten kształt); ten sam
  mechanizm dotyczy `Wrapper#unwrap`, `CallableStatement`.
- **Weryfikacja:** test mutacyjny; `go test ./...`, `go vet ./...`.
- **Następny krok (wymaga zgody):** tabela nadinterfejsów JDK dla typów
  JDBC (`PreparedStatement` → `Statement` → `Wrapper` itd.) w
  `jdbc:objects`/`spidispatch`, z pomiarem wpływu; potem Spring Data/JPA i DI.

### ROADMAP:B4 — poprawka po `b4-3`: fasada `DriverManager#getConnection` (`service-provider:dispatch` `p2`) — DONE (2026-11-14)

- **Zakres (zatwierdzony 2026-11-14):** zamknąć fałszywy negatyw znaleziony
  przez wyrocznię JDBC.
- **Zrealizowane:** `spidispatch.Facades` — statyczna metoda JDK, która
  sama woła metodę zarejestrowanego serwisu; jedyny wpis:
  `java.sql.DriverManager#getConnection` (każde przeciążenie, tylko
  `invokestatic`) → `Driver#connect(String, Properties)` zarejestrowanych
  sterowników (krawędź `model`, rodzaj `interface`). Typ zwracany fasady
  wchodzi do produktów serwisu (`Connection` → `Statement` → `ResultSet`).
  Fasada jest sygnaturą tylko przy zarejestrowanym sterowniku. Wersja
  `p1` → `p2`.
- **Wynik:** `testdata/jdbc` — 13 z 13 wykonanych ciał osiągalnych (było
  5); nowe nadmiarowe cele tylko w jarze sterownika (nigdy nie tworzony
  `UnusedStatement` i dziedziczone ciało zaślepki), zgodnie z zasadą
  rozwijania produktów bez bramki twórców (jak `UnusedCodec`). Aplikacje
  referencyjne: v1 **bajtowo identyczne**; graf +6 (petclinic: Hibernate
  `DriverManagerConnectionCreator`, H2 `DbUpgrade`), +4 (WebGoat: Hibernate,
  Spring `DriverManagerDataSource`), 0 (JPetStore: wywołujący MyBatis
  nieosiągalni z korzeni) — tam sterowniki były już osiągalne przez
  bezpośrednie `driver.connect` (Hikari).
- **Weryfikacja:** `TestDriverManagerFacade` (fasada, produkt, wywołanie
  niestatyczne, brak rejestracji), `TestJDBCOracle` (zaktualizowany: każde
  wykonane ciało osiągalne, nadmiarowe cele wymienione dokładnie),
  `TestSPIDispatchOracle`; `go test ./...`, `go vet ./...`.
- **Następny krok (wymaga zgody):** wyrocznia proxy frameworków (MyBatis,
  Spring Data) albo DI.

### ROADMAP:B4 — przyrost `b4-3`: wyrocznia JVM dla obiektów JDBC — DONE for this scope (2026-11-14); znaleziony fałszywy negatyw

- **Zakres (zatwierdzony 2026-11-14):** obiekty JDBC (`jdbc:objects`) i
  dispatch sterownika (`service-provider:dispatch`) na prawdziwym JVM.
  Proxy frameworków (MyBatis, Spring Data, Spring ORM) wymagają
  prawdziwych bibliotek — osobny przyrost.
- **Zrealizowane:** `testdata/jdbc` — `stub.jar` (abstrakcyjne zaślepki
  `Connection`/`PreparedStatement`/`ResultSet` generowane refleksją przez
  `scripts/JdbcStubs.java`, nie commitowane jako źródła), `driver.jar`
  (sterownik w `META-INF/services/java.sql.Driver`, rejestrujący się w
  `<clinit>` jak każdy sterownik JDBC 4), `pool.jar` (kształt Hikari:
  `DataSource`, proxy połączenia, proxy statementu dziedziczący
  `executeQuery` z klasy abstrakcyjnej), `other.jar` (statementy nigdzie
  nie tworzone), `app.jar` (zapytanie przez `DriverManager` i przez pulę);
  `scripts/jdbc-oracle.sh`, `TestJDBCOracle`.
- **Wynik:** pula — 5 z 5 wykonanych ciał osiągalnych, każda krawędź
  `model` do ciała wykonanego przez JVM; statementy bez osiągalnego twórcy
  nieosiągalne. **Sterownik — 0 z 8:** JVM dochodzi do `FxDriver#connect`
  z `java.sql.DriverManager#getConnection` (kod JDK poza wejściem), więc w
  wejściu nie ma wywołania na `java.sql.Driver`; `service-provider:dispatch`
  nie ma czego rozwinąć, połączenie sterownika nie ma osiągalnego twórcy, a
  bramka `jdbc:objects` odcina połączenie, statement i result set. Żadna
  luka nie jest zgłaszana — fałszywy negatyw bez ujawnienia. Ten kształt
  występuje w zależnościach aplikacji referencyjnych: Spring
  `DriverManagerDataSource`, MyBatis `UnpooledDataSource`/`PooledDataSource`,
  Hibernate `DriverManagerConnectionCreator`, Hikari `DriverDataSource`,
  logback `DriverManagerConnectionSource`, log4j `JDBCAppender`.
- **Weryfikacja:** test utrwala obecne zachowanie (padnie, gdy luka
  zostanie zamknięta); `go test ./...`, `go vet ./...`.
- **Następny krok (wymaga zgody):** model `DriverManager#getConnection`
  (wszystkie przeciążenia) → `Driver#connect` zarejestrowanych sterowników
  (rozszerzenie `service-provider:dispatch`), z pomiarem wpływu na v1
  trzech aplikacji; potem wyrocznia proxy frameworków i DI.

### ROADMAP:B4 — przyrost `b4-2`: wyrocznia JVM dla callbacków i MethodHandle — DONE for this scope (2026-11-14)

- **Zakres (zatwierdzony 2026-11-14):** drugi brak z macierzy B4 —
  callbacki (ASYNC A0–A2) i MethodHandle (DYNAMIC_LINKAGE L2) nie miały
  wyroczni na prawdziwym JVM.
- **Zrealizowane:** `testdata/async` (źródła, `app.jar`, `observed.txt`;
  `fx.trace.Ran` zapisuje ciała z każdego wątku),
  `scripts/async-oracle.sh` (przez `scripts/JavaBuild.java`),
  `TestAsyncOracle` (bez JDK). Przypadki: dodatnie — `new Thread(new
  Task())`, lambdy do `ExecutorService#execute`/`submit`, `Callable` do
  `submit`, `AccessController#doPrivileged`, `findVirtual`/`findStatic` z
  literałami; ujemne — `Runnable` zaalokowany, nigdzie nie przekazany;
  `findVirtual` nieistniejącej metody (JVM rzuca wyjątek); nierozstrzygnięte
  — callback dochodzący do `new Thread` przez parametr, nazwa uchwytu z
  wejścia.
- **Wynik:** model zgodny z JVM bez poprawek: dokładnie 5 krawędzi
  `callback` i 2 `methodhandle`, każda do ciała wykonanego przez JVM
  (krawędzie `Callable`/`PrivilegedAction` trafiają w metodę pomostową,
  która woła wykonane ciało); 8 z 10 wykonanych ciał osiągalnych z
  `main`; brakują tylko dwa przypadki nierozstrzygnięte, zgłoszone jako
  luki; brakująca metoda uchwytu — luka z powodem.
- **Weryfikacja:** test mutacyjny (usunięta oczekiwana krawędź → test
  czerwony); `go test ./...`, `go vet ./...`.
- **Ograniczenia:** pojedyncze przypadki na kształt; nie sprawdzono
  `findSpecial`/`findConstructor`, `Thread` z `ThreadGroup`, przeciążeń
  `doPrivileged` z kontekstem. Ślad na prawdziwej aplikacji nie był
  częścią przyrostu.
- **Następny krok (wymaga zgody):** wyrocznia dla obiektów JDBC i proxy
  frameworków, potem DI.

### ROADMAP:B4 — pierwszy przyrost (`b4-1`): wyrocznia JVM dla refleksji R0–R2 — DONE for this scope (2026-11-14)

- **Zakres (zatwierdzony 2026-11-14, „Kontynuuj B4”):** pierwszy brak z
  macierzy B4 — refleksja nie miała wyroczni na prawdziwym JVM.
- **Zrealizowane:** `testdata/reflection` (źródła, `app.jar`,
  `observed.txt`), `scripts/reflection-oracle.sh` (kompilacja przez
  `scripts/JavaBuild.java`, bez `javac` na PATH), `TestReflectionOracle`
  (bez JDK). Przypadki: dodatnie — `forName`+`getMethod` z literałami,
  wyszukanie na nadtypie z alokacją podtypu (JVM wybiera `Sub#greet`),
  statyczna z `null`, prywatna przez `getDeclaredMethod`+`setAccessible`,
  `Method` w polu statycznym, `Constructor#newInstance`; ujemne —
  dowiedziona niezgodność liczby argumentów (JVM rzuca wyjątek, ciało nie
  rusza), ciało nadtypu niewybrane przez JVM; nierozstrzygnięty — nazwa
  metody z wejścia.
- **Wynik:** model zgodny z JVM bez poprawek: dokładnie 6 krawędzi
  `reflection`, każda do ciała, które JVM wykonał; 7 z 8 wykonanych ciał
  osiągalnych z `main`; jedyne nieosiągnięte to przypadek
  nierozstrzygnięty, zgłoszony jako luka (`unresolved`), a niezgodność
  liczby argumentów — jako luka z powodem.
- **Weryfikacja:** test mutacyjny (usunięta oczekiwana krawędź → test
  czerwony); `go test ./...`, `go vet ./...`.
- **Ograniczenia:** pojedyncze przypadki na kształt; poza fixture
  zostają `Class.forName` przez zmienną, tablice parametrów z prymitywami,
  `Method` z wyniku wywołania (tu: luka). Ślad na prawdziwej aplikacji nie
  był częścią tego przyrostu.
- **Następny krok (wymaga zgody):** wyrocznia dla callbacków/MethodHandle,
  potem obiekty JDBC/proxy i DI.

### ROADMAP:B0 — czwarty przyrost (`b0-4`): pomiar wariantów reguły luk — DONE for this scope (2026-11-14)

- **Zakres (zatwierdzony 2026-11-14):** pomiar rozszerzenia reguły luk o luki
  SAM i pomijanych odbiorców, **report-only** — bez zmiany reguły i v1;
  decyzja o regule po pomiarze.
- **Punkt wyjścia (zmierzony na `main`):** reguła oznacza wszystkie
  NOT_REACHABLE (18/18, 18/18, 15/15 petclinic/WebGoat/JPetStore) dzięki
  611/681/161 osiągalnym miejscom bez ograniczonego zbioru kandydatów —
  samo dołożenie nowych rodzajów luk nie mogło więc zmienić żadnej flagi;
  pytanie brzmiało, czy nowe rodzaje są *selektywne* w zakresie instancji.
- **Zrealizowane:** `reach.SetProbe`/`ApplyProbe` (wersja `b0-4`): per
  ustalenie NOT_REACHABLE liczniki `unbounded`, `bounded` (+ rozbicie po
  rodzaju), `lambda` (osiągalne nierozwiązane wywołania SAM, których
  interfejs ma lambdę tworzoną w klasie instancji — z `SamInterface`
  wywołań invokedynamic), `skipped_receiver` (osiągalne bezpośrednie
  wywołania wirtualne/interfejsowe na typach `java.`/`javax.`/`jdk.`/`sun.`,
  których pomijane przez CHA rozwinięcie wybrałoby ciało w instancji —
  to samo `CHA.DispatchTargets` z regułą pomijania zwalniającą typ
  odbiorcy) i ich część na metodach `java.lang.Object`. Pięć wariantów
  (`current`, `scoped`, `scoped+lambda`, `…+skipped-object`, `…+skipped`);
  v2 `findings[].gap_probe`, `coverage.gap_probe`; raport „Gap-rule
  probe”; linia logu. `SetGaps` korzysta ze wspólnego `reachableGaps`.
- **Wynik:** current 18/18/15, scoped 17/15/14, scoped+lambda 17/16/14,
  +skipped (z i bez `Object`) 18/18/15. Luki lambda dodają tylko jedno
  oznaczenie względem `scoped` (WebGoat, spring-security-web: 200 luk);
  pomijani odbiorcy trafiają w **każdą** instancję (petclinic 1 917–16 122
  wywołań na instancję, z czego ~950 na metodach `Object`; JPetStore
  372–3 759) — jako składnik reguły nic nie rozróżniają. Nieoznaczone przez
  `scoped`: jakarta.el (petclinic, WebGoat), commons-lang 2.6 i
  spring-security-web (WebGoat), ognl (JPetStore). **Decyzja: reguła bez
  zmian** (zapis w kontrakcie §4).
- **Weryfikacja:** `TestGapProbe`, `TestGapProbeVariants`,
  `TestGapProbeSkippedFromGraph` (`javac` niedostępny — pozytywny przypadek
  pomijanego odbiorcy białoskrzynkowo, ścieżka grafu własną regułą
  pomijania na fixture); `go test ./...`, `go vet`. petclinic, WebGoat,
  JPetStore: v1 **bajtowo identyczne** z `main`; v2 identyczne poza nowym
  `gap_probe`; koszt pomiaru 0,1–1,1 s.
- **Ograniczenia (świadome):** pomiar używa rozwinięcia CHA także w trybie
  RTA; luki lambda liczone tylko przez lambdy (klasy implementujące
  interfejs to wymiar CHA/pomijanych odbiorców); wiersze bez instancji
  (`advisories_without_instance`) nie są mierzone.
- **Następny krok (wymaga zgody):** selektywność reguły wymaga zawężenia
  *nieograniczonych* miejsc (refleksja: 510+ osiągalnych w petclinic) albo
  trafności względem metody podatnej zamiast całej instancji; ewentualnie
  brakujące wyrocznie B4 albo inny obszar.

### ROADMAP:B0 — trzeci przyrost (`b0-3`): wspólny słownik luk — DONE for this scope (2026-11-13)

- **Zakres (zatwierdzony 2026-11-13):** słownik i mapowanie, **bez zmiany
  reguły luk** (decyzja użytkownika: wariant bez rozszerzania
  `MODEL_INCOMPLETE`).
- **Inwentaryzacja przed zmianą:** sześć osobnych słowników — miejsca reguły
  luk (`reflection`, `methodhandle`, `callback`, `serviceloader`,
  `deserialization`, `allocation`), nieznane alokacje katalogu typów
  (`reflection-unresolved`, `dynamic-proxy`, `container-proxy`,
  `unsafe-allocation`, `constructor-handle`, `service-unresolved`,
  `external-input`), luki SAM (`parameter`, `field`, `call-result`, …),
  tekstowe powody koordynatora B3, przyczyny śladu (`reflection`,
  `method-handle`, `proxy`, `async`, `sam-call`, `skipped-receiver`, …),
  luki taint/servletów/proxy. Ten sam mechanizm pod 2–3 nazwami (refleksja,
  uchwyty, ServiceLoader, wątki).
- **Zrealizowane:** `internal/gaps` (wersja `b0-3`): 16 mechanizmów,
  mapowanie każdego kodu pięciu źródeł (`site`, `allocation`, `lambda`,
  `runtime`, `coordinator` — szczegół = id modelu) na jeden mechanizm;
  kody źródeł bez zmian. v2 `coverage.gaps` (mechanizm, źródło, kod,
  liczba, dla `site` liczba osiągalnych), `findings[].gaps` (`model_gaps`
  w nowym słowniku); raport: tabela „Model gaps by mechanism”; tabela w
  kontrakcie wyników generowana z mapowania.
- **Weryfikacja:** `TestEveryProducerCodeMapped` (każda stała luk alloc,
  samcall, runtimeevents i każdy rodzaj miejsca ma mechanizm),
  `TestVocabularyDocumented`, `TestFromCountsAndSummary`; `go test ./...`,
  `go vet ./...`. petclinic, WebGoat, JPetStore: v1 **bajtowo identyczne**
  z `main`; v2/raport z tabelą (petclinic m.in. `reflection` 816 miejsc,
  510 osiągalnych + 258 alokacji; `lambda` 1 138 luk SAM).
- **Ograniczenia (świadome):** reguła luk nadal bierze tylko miejsca `site`
  — nierozwiązane wywołania SAM i pomijane odbiorcy `java.*`/`javax.*` są
  raportowane, ale nie osłabiają `NOT_REACHABLE`; pomijani odbiorcy nie
  mają jeszcze licznika statycznego (tylko przyczyna w śladzie); luki
  taint, rejestracji servletów i proxy bez twórcy nie są jeszcze w
  słowniku.
- **Następny krok (wymaga zgody):** rozszerzenie reguły luk (SAM, pomijani
  odbiorcy) z pomiarem wpływu na v1 albo brakujące wyrocznie B4.

### ROADMAP:B0 — drugi przyrost (`b0-2`): zgłaszana ścieżka bez kroków tylko-CHA — DONE for this scope (2026-11-13)

- **Zakres (zatwierdzony 2026-11-13):** wybór zgłaszanej ścieżki — następny
  krok z `b0-1`.
- **Zrealizowane:** dwa dodatkowe przebiegi BFS (app-only i pełny) z
  pominięciem krawędzi o podstawie `cha` (`bfsKeep`, `noCHA`; cel osiągalny
  inną regułą od tego samego wywołującego zostaje osiągalny). Werdykt
  rozstrzygają jak dotąd zwykłe przebiegi; jeśli cel jest osiągalny w
  przebiegu bez CHA, zgłaszana jest ta ścieżka, także dłuższa. Odtąd
  `heuristic` w charakterze ustalenia znaczy: metoda osiągalna wyłącznie
  przez jakieś rozwinięcie CHA. `reach.EngineVersion` 2 → 3.
- **Weryfikacja:** `TestPreferPathWithoutCHAOnlySteps` (wszystkie metody
  fixtures wyroczni `cg.jar` i `rta.jar` od `main`: werdykt jak z
  dotychczasowych przebiegów; ścieżka bez kroku tylko-CHA zawsze, gdy
  istnieje; oba przypadki występują); `go test ./...`, `go vet ./...`.
  Względem binarki z `main` (v2): petclinic, WebGoat, JPetStore — **0 zmian
  werdyktów**; zmienione ścieżki 7/8/0 (wszystkie dłuższe); ustalenia
  pozytywne zależne od CHA 106 → **99 z 109**, 89 → **81 z 137**, 16 → **16
  z 16**; koszt +0,1–0,4 s, +~20–50 MB.
- **Wniosek:** dla ok. 90 % pozytywnych ustaleń petclinic i całego
  JPetStore podatna metoda jest osiągalna tylko przez rozwinięcie CHA —
  RTA (`-callgraph-mode rta`) i modele dispatchu decydują o tym, które z
  nich są realne. Charakter jest dowodem, nie werdyktem.
- **Dokumentacja:** `docs/RESULTS_CONTRACT.MD`, README, `docs/SUPPORT.MD`,
  ten rejestr.
- **Następny krok (wymaga zgody):** wspólny słownik luk (kryterium B0 „ta
  sama luka tak samo interpretowana”) albo brakujące wyrocznie B4.

### ROADMAP:B0 — ocena i pierwszy przyrost (`b0-1`): dowód kroku ścieżki, charakter ustalenia, `MAPPING_MISSING` — IN_PROGRESS (2026-11-13)

- **Zakres (zatwierdzony 2026-11-13):** ocena B0 i B4 punkt po punkcie oraz
  pierwszy przyrost B0.
- **Ocena B0 (przed `b0-1`):**
  1. *Jeden wersjonowany schemat wyniku, instancji i dowodu* — spełnione
     przez v2 (CVE F0/F1); dowody innych torów (deserializacja, ślady,
     taint, `target_runtime`) to osobne sekcje addytywne, nie wspólny typ.
  2. *Rozdzielone osie* (tożsamość, mapowanie, osiągalność, dane,
     obserwacja) — spełnione: kontrakt wyników, taint i ślady osobno.
  3. *Kod powodu, miejsce, wersja modelu, warunki* — częściowo: kody i
     wersje algorytmów są, ale ścieżka była listą nazw metod; podstawę
     (direct/CHA/model…) miała tylko krawędź wejścia z aplikacji.
  4. *Charakter ustalenia* (reguła JVM / model / heurystyka / deklaracja /
     obserwacja) — **brak** → `b0-1`.
  5. *Wpływ luk na wynik* — częściowo: luki globalne i per ustalenie są
     rozdzielone, ale `MODEL_INCOMPLETE` jest tylko informacyjne i pokrywa
     wszystkie `NOT_REACHABLE`, więc niczego nie różnicuje (otwarte).
  6. *Znaczenie `NOT_REACHABLE`/`NOT_APPLICABLE`* — spełnione (kontrakt).
  *Kryterium „eksport legacy nie ukrywa niepewności”* — prawie: v1 niesie
  kody powodów i luki, ale 20 z 27 wierszy `UNKNOWN` petclinic (advisory
  bez mapowania) nie miało żadnego kodu → `b0-1`. *„Ta sama luka nie jest
  różnie interpretowana przez dwa analizatory”* — nie zweryfikowane:
  słowniki luk są osobne (globalne miejsca modeli, luki RTA nieznanych
  alokacji, luki SAM, nierozwiązane fakty B3); ujednolicenie to dalszy
  przyrost.
- **Zrealizowane w `b0-1`:** `reach.PathEdge`/`EdgeEvidence`: każdy krok
  zgłoszonej ścieżki odbudowany z grafu ze **wszystkimi** krawędziami między
  dwiema metodami (`kind`, `basis`, offsety wywołań — kryterium B1 „dwa
  powody tej samej krawędzi zachowane”, `callgraph.EdgesBetween`), charakter
  kroku (najsilniejszy dowód: krok poparty też wywołaniem bytecode to
  `jvm_rule`) i zbiór charakterów ustalenia: `jvm_rule`, `library_model`,
  `input_declaration`, `heuristic` (tylko CHA), `unclassified`. v2
  `findings[].path_edges`, `findings[].character`; raport: strzałki kroków
  innych niż `jvm_rule` opisane (`→[heuristic: cha]`) i linia „Path
  evidence”; v1 bez tych pól. Nowy kod `MAPPING_MISSING` (advisory, którego
  żadne źródło nie zmapowało) — także w `reasons` v1.
- **Weryfikacja:** `TestPathEvidence` (fixture overload: liczba kroków,
  dowody, podstawy, brak pól w v1), `TestBasisCharacterAndLines`,
  `TestReasonCodesDocumented`; `go test ./...`, `go vet ./...`. Względem
  binarki z `main`: petclinic, WebGoat, JPetStore — **0 zmian werdyktów**,
  jedyna różnica v1 to `MAPPING_MISSING` na 20/19/2 wierszach; żaden
  `UNKNOWN` bez powodu. Pomiar charakteru (zgłoszona najkrótsza ścieżka,
  nie wszystkie ścieżki): ustalenia pozytywne z co najmniej jednym krokiem
  tylko-CHA — petclinic 106 z 109, WebGoat 89 z 137, JPetStore 16 z 16;
  tylko `jvm_rule`: 3, 39, 0.
- **Dokumentacja:** `docs/RESULTS_CONTRACT.MD` (kod, pola v2), README,
  `docs/SUPPORT.MD`, ten rejestr.
- **Ograniczenia:** charakter opisuje zgłoszoną ścieżkę, nie wszystkie
  możliwe; nie zmienia werdyktu; obserwacja (ślady) nie jest jeszcze
  charakterem kroku; ujednolicenie słowników luk i różnicujące użycie luk
  w negatywnych werdyktach — otwarte.
- **Następny krok (wymaga zgody):** preferowanie ścieżki bez kroków
  tylko-CHA przy wyborze zgłaszanej ścieżki (ta sama osiągalność, lepszy
  dowód) albo wspólny słownik luk.

### ROADMAP:B4 — ocena (2026-11-13)

Stan punktów B4: *fixtures w repo, testy Go bez JDK i sieci* — spełnione;
*porównywanie grafów, przyczyn i werdyktów na tej samej konfiguracji* —
spełnione (`-graph-compare`, `-snapshot-compare`, `scripts/graph-benchmark.sh`);
*pomiar pominiętych celów, błędnych krawędzi, nierozstrzygniętych miejsc,
czasu i pamięci* — spełnione dla grafu (benchmark, `TRACE_GAPS`, wyrocznie),
nie dla każdego modelu; *`go test`/`go vet` przed wydaniem* — spełnione;
*dokumentacja z dokładnymi wersjami* — `docs/SUPPORT.MD`.
Macierz obszarów (na podstawie testów i fixtures w repo; „wyrocznia” =
zachowanie zmierzone na prawdziwym JVM i zapisane w fixture):

| Obszar / model | Przypadki +/−/nierozstrz. | Wyrocznia JVM | Prawdziwa aplikacja |
|---|---|---|---|
| Graf wywołań (JVM_GRAPH) | tak (`testdata/callgraph`) | tak (`callgraph-oracle.sh`) | benchmark 3 aplikacji |
| RTA, typy tworzone (GRAPH_PRECISION) | tak (`testdata/rta`, `alloc`) | tak (`rta-oracle.sh`, `alloc-oracle.sh`) | benchmark |
| ServiceLoader (SERVICE_LOADING) | tak, z jawnymi lukami (`testdata/oracle`) | tak | petclinic/WebGoat |
| Dispatch servletów, SPI | tak (`testdata/servletdispatch`, `spidispatch`) | tak (`observed.txt`) | petclinic/WebGoat |
| Lambdy/SAM (DYNAMIC_LINKAGE L1) | tak (`testdata/samcall`) | tak | ślad petclinic |
| Deserializacja, taint | tak (`testdata/deserialize`, `taint`) | tak | WebGoat, aplikacje kontrolne |
| log4j 1.2 / logback | testy jednostkowe | aplikacje kontrolne na JVM (`prepare-*-demo.sh`) | JPetStore, petclinic |
| Obiekty JDBC | testy jednostkowe + `testdata/jdbc` (13 dodatnich — pula i sterownik przez `DriverManager`; 2 ujemne poza jarami sterownika/puli) | tak (`jdbc-oracle.sh`, 2026-11-14) | tak |
| Proxy frameworków | testy jednostkowe + `testdata/mybatis` (MyBatis: 51 ciał ze stosów, wszystkie osiągalne; 1 ujemny) + `testdata/springdata` (Spring Data: 16 ciał, wszystkie osiągalne; 1 ujemny); JPA `EntityManager` bez fixture | tak dla MyBatis i Spring Data (`mybatis-oracle.sh`, `springdata-oracle.sh`, 2026-11-14) | tak |
| Refleksja (R0–R4) | testy jednostkowe + `testdata/reflection` (9 dodatnich, 7 ujemnych, 1 nierozstrzygnięty, 1 jawna nadaproksymacja) | tak (`reflection-oracle.sh`, 2026-11-14) | petclinic/WebGoat (efekt zmierzony) |
| Callbacki (ASYNC A0–A2), MethodHandle (L0/L2) | testy jednostkowe + `testdata/async` (7 dodatnich, 2 ujemne, 2 nierozstrzygnięte) | tak (`async-oracle.sh`, 2026-11-14) | liczby na petclinic/WebGoat |
| DI/adnotacje/konfiguracja (FRAMEWORKS F1–F2) | testy jednostkowe + `testdata/di` (DI i cykl życia: 21 ciał; ujemne: nieaktywny `@Profile`, klasa bez stereotypu; nierozstrzygnięte: `@Conditional`, `@Lazy`) + `testdata/xmlbeans` | tak (`di-oracle.sh` b4-7, `xmlbeans-oracle.sh` b4-8, 2026-11-15); rozwiązywanie wstrzyknięć (qualifier, `@Primary`) — wyrocznia od `f5-3` (`TestDIInjectionOracle`), dwa scenariusze profili (`TestDIProfilesOracle`); `@Value` tylko raportowe — bez wyroczni; konsumenci wiadomości — bez wyroczni (ocena F5 2026-11-22); Spring MVC: `mvc-oracle.sh` (`f5-2`, 2026-11-22) | petclinic/WebGoat |
| Mapowania CVE | zestaw kontrolny 33 przypadków (`testdata/cve-control`), baseline | nie dotyczy | 3 aplikacje |
| Runtime JDK (CVE F8) | testy jednostkowe na stronach syntetycznych + kuratorowany katalog (88 CVE, 2026) + testy end-to-end CLI (`targetruntime_e2e_test.go`, F9 2026-11-23) | nie dotyczy | petclinic z Temurin 17 (F9) |

**Braki B4 (kolejne kroki):** wyrocznie JVM w fixture dla
proxy JPA `EntityManager` oraz DI; przypadki
„nierozstrzygnięte” jawnie tylko tam, gdzie fixture ma zapisane luki
(ServiceLoader, deserializacja, RTA). Refleksja, callbacki/MethodHandle,
obiekty JDBC, proxy MyBatis i Spring Data: zrobione 2026-11-14 (`b4-1`…`b4-5`).

### CVE:F8b — ocena wydania runtime'u: dostawca upstream OpenJDK (katalog kuratorowany) — DONE dla mechanizmu (2026-11-13)

- **Zakres (zatwierdzony 2026-11-13):** dostawca upstream OpenJDK, Temurin
  jako pierwsza dystrybucja. **Zmiana drogi danych (decyzja użytkownika):**
  warunki korzystania z openjdk.org zabraniają automatycznego dostępu bez
  pisemnej zgody Oracle (wyjątki: indeks wyszukiwarki, odczyt „w normalnym
  toku prac rozwojowych” nad kodem); API Adoptium nie podaje CVE, NVD ma
  tylko CPE Oracle JDK w pojedynczych wersjach. Wybrano katalog
  kuratorowany: ariadne nigdy nie łączy się z openjdk.org.
- **Zrealizowane:** `internal/runtimevuln`: interfejs `Provider`, dostawca
  `openjdk-vulnerability-group` nad katalogiem osadzonym w binarce
  (`openjdk_advisories.yaml`, schemat 1, walidacja dat, URL, CVE, wersji i
  linii; dziś pusty), dopasowanie per CVE: linia nieoznaczona w macierzy →
  `not_affected`, wersja późniejsza niż ostatnia podatna → `not_affected`,
  równa → `affected`, wcześniejsza → `affected` tylko przy stwierdzeniu
  „and earlier” (`earlier_affected`), inaczej `undetermined`; advisory bez
  danej linii → notatka (brak danych, nie „bezpieczny”); katalog starszy
  niż 100 dni → notatka; reachability zawsze `not_assessed`. Tylko Temurin
  (inne dystrybucje i nierozpoznana → `not_performed` z powodem).
  `cmd/openjdk-catalog`: konwerter stron zapisanych ręcznie
  (`YYYY-MM-DD.html`, tylko pliki lokalne, wynik przechodzi walidację
  loadera). `-openjdk-catalog FILE` do sprawdzenia szkicu. Raport: linia
  nagłówka z licznikami i sekcja „Target runtime advisories” (affected i
  undetermined); v2 `target_runtime.advisories`; snapshot
  `runtime.identity`, `runtime.openjdk_catalog` (i konfiguracja
  `openjdk-catalog` przy nadpisaniu). Procedura kuracji i reguły w
  `docs/RUNTIME_SOURCES.MD`.
- **Weryfikacja:** `TestAssess` (syntetyczny katalog: affected, fixed,
  undetermined, „and earlier”, JDK 8 `8uNNN`, linia bez pokrycia),
  `TestAssess_notPerformed` (Corretto, nierozpoznana dystrybucja, brak
  wersji, pusty katalog, katalog przeterminowany), `TestParseCatalog_rejects`,
  `TestEmbeddedCatalog`, `TestParseAdvisoryHTML` (syntetyczna strona:
  nagłówek dwuwierszowy, macierz OpenJFX pominięta, „and earlier”);
  `go test ./...`, `go vet ./...`. petclinic: Temurin 17.0.16 z
  `-runtime-file` — z katalogiem osadzonym „not assessed (no advisories
  yet)”, z syntetycznym katalogiem z konwertera 1 affected / 1 not
  affected i sekcja w raporcie; Debian openjdk-25 — „not covered”. Wyniki v1
  bez zmian.
- **Ograniczenia:** **katalog jest pusty** — ocena działa dopiero po
  kuracji (zapisanie stron ręcznie, konwersja, przegląd); konwerter
  sprawdzony tylko na stronie syntetycznej (format rzeczywistych stron
  potwierdzi pierwsza kuracja); tylko Temurin; brak osiągalności kodu JDK
  (F8c).
- **Następny krok (wymaga działania użytkownika):** zapisać strony advisories
  (2019–2026) jako `work/openjdk-advisories/YYYY-MM-DD.html`; potem
  konwersja, przegląd i commit katalogu.
- **Kuracja 1 (2026-11-20):** użytkownik zapisał ręcznie 8 stron
  (2026-01-20, 04-21, 07-21, 08-18, 10-20, 10-21, 10-24, 11-17; w tym
  advisories pozakwartalne). Prawdziwe strony ujawniły dwa błędy
  konwertera: CVSS i wektor w jednej komórce (CVSS wychodził 0) oraz wersje
  awaryjne z piątym elementem (`17.0.21.0.2`) — cięte na `17.0.21.0` i
  fałszywą linię `2`. Poprawki: `runtimeinfo.Release.Extra` (piąty i dalsze
  elementy JEP 322), porównanie po nich (brakujący = 0), `versionRe` bez
  limitu elementów, CVSS z pierwszego tokenu komórki. Nowa notatka: runtime
  starszy niż ostatnia podatna wersja najstarszego advisory dla linii →
  „advisories sprzed katalogu nie są uwzględnione”. Katalog wersja 1: 8
  advisories, 88 CVE, linie 8/11/17/21/25/26/27; sprawdzony niezależnym
  parserem HTML (0 różnic). Testy: `TestParseAdvisoryHTML_emergency`
  (nowa strona syntetyczna w kształcie prawdziwej),
  `TestCompare_extraElements`, `TestParseVersion` (5 elementów),
  `TestAssess` (notatka); nowe testy padają na starym kodzie.
  E2E (`bin/ariadne`, petclinic, `-runtime-file` Temurin, v2): 17.0.21 →
  10 affected (10-21 równa, 10-24 wcześniejsza wersja awaryjna, 11-17: 8),
  17.0.21.0.2 → 9 (10-21 naprawione), 17.0.21.0.3 → 8 (10-24 naprawione),
  17.0.22 → 0/88, 21.0.13.0.3 → 9 (11-17), 8u512 → 5 (11-17),
  11.0.32.1 → 46 (10-20: 39 = liczba CVE oznaczonych dla linii 11),
  17.0.10 → 70 + notatka o advisories sprzed katalogu; sekcja „Target
  runtime advisories” w raporcie; `go test ./...`, `go vet` czyste.
- **Ograniczenia po kuracji 1:** brak advisories sprzed 2026 (34 strony
  2019–2025 nie zapisane) — dla starszych runtime'ów notatka zamiast
  danych; tylko Temurin; F8c (osiągalność kodu JDK) bez zmian.

### CVE:F8a — identyfikacja docelowego runtime'u i wybór źródła — DONE for this scope (2026-11-13)

- **Zakres (zatwierdzony 2026-11-13):** F8a zadania 1–6; pierwszy dostawca
  F8b wybrany przez użytkownika: **upstream OpenJDK** (Vulnerability Group).
- **Zrealizowane:** `internal/runtimeinfo` (`rules f8a-1`): `-jdk-home`
  czyta wyłącznie plik `release` instalacji (nic nie uruchamia, JDK
  skanera nigdy nie jest domyślny), `-runtime-file` — deklarowany YAML
  (nieznane pola = błąd); flagi wzajemnie wykluczające się. Surowe
  wartości zachowane; vendor/dystrybucja tylko ze stałej tabeli
  `IMPLEMENTOR` (`Oracle Corporation`, `Red Hat, Inc.` — vendor bez
  dystrybucji, bo IMPLEMENTOR wspólny dla kilku produktów); wersje JEP 322
  i `1.8.0_NNN-bBB`; dla pakietów Debian/Ubuntu nazwa i wersja pakietu
  z wersji runtime'u, wydanie dystrybucji nierozstrzygnięte. Braki
  (vendor, build, OS, arch, dystrybucja) jawnie w `unresolved`. Nagłówek
  raportu „Target runtime”, v2 `target_runtime` (zawsze, `assessment:
  not_performed` z powodem: brak opisu albo brak dostawcy), konfiguracja
  snapshotu `target-runtime` (pola znormalizowane, bez ścieżki).
  Studium źródeł: `docs/RUNTIME_SOURCES.MD` (OpenJDK Vulnerability Group:
  34 strony 2019–2026, macierz CVE × komponent × linie; Debian/Ubuntu
  przez OSV; Oracle CPU).
- **Weryfikacja:** `TestParseVersion`, `TestFromJDKHome` (prawdziwy plik
  `release` pakietu Debian openjdk-25 z tej maszyny; Temurin, Oracle JDK 8 i
  plik bez IMPLEMENTOR skonstruowane według formatu — oznaczone w
  `testdata/README.MD`), `TestFromFile`; `go test ./...`, `go vet ./...`.
  petclinic: bez flag v1 bajtowo identyczne z `main`; z `-jdk-home`
  tego JDK opis „Debian OpenJDK package, 25.0.5+8-1-Debian, Linux x86_64,
  package openjdk-25 25.0.5+8-1”, nierozstrzygnięte wydanie Debiana;
  obie flagi naraz i brak pliku `release` — czytelne błędy przed analizą.
- **Dokumentacja:** README (flagi), `docs/SUPPORT.MD` (wiersz Target JDK →
  Partial), `docs/RUNTIME_SOURCES.MD`, plan CVE, ten rejestr.
- **Ograniczenia:** brak oceny podatności runtime'u (F8b); nieznany lub
  wspólny IMPLEMENTOR nie daje dystrybucji; warunki automatycznego
  pobierania stron OpenJDK do potwierdzenia przed F8b.
- **Następny krok (wymaga zgody):** F8b — adapter upstream OpenJDK
  (wersjonowany katalog z advisories, cache/snapshot, Temurin jako
  pierwsza obsługiwana dystrybucja).

### GRAPH_PRECISION — pełne wejście WebGoata (`webgoat-full-input`) — DONE for this scope (2026-11-13)

- **Zakres (zatwierdzony 2026-11-13):** wejście WebGoata = zawartość
  wdrożenia `webgoat-server` (kontener, wszystkie lekcje, launcher), zależności
  z rozwiązania `webgoat-server`.
- **Zmiana (`scripts/prepare-webgoat.sh`):** instalacja pom-ów rodziców (`-N`)
  i kontenera; lekcje kompilowane przez bezpośrednie wywołanie celów
  `resources:resources compiler:compile` (lekcja cross-site-scripting nie
  uruchamia wtedy asciidoctora, który pod JDK 17 pada); każda lekcja
  instalowana jako jar (`install:install-file`), potem `webgoat-server`
  kompilowany Mavenem (koniec obejścia `javac` + Lombok z PR #51) i jego
  zależności (`dependency:copy-dependencies`/`list`, bez modułów WebGoata i
  opcjonalnego `spring-boot-devtools`). Do `app-classes.jar` trafiają klasy
  kontenera (z zasobami), klasy lekcji (bez zasobów — nieczytane przez
  ariadne, i18n ma kolizje nazw) i launchera.
- **Wynik (WebGoat):** 72 → 246 klas aplikacji, 126 → 131 jarów zależności
  (+ PostgreSQL JDBC, jsoup, jjwt, zxcvbn, commons-lang 2.6, checker-qual;
  − `jakarta.activation-api`, mediowany przez Mavena); 7 rejestracji servletów
  (+ `@WebServlet` `ImageServlet` lekcji challenge); +39 173 krawędzi (CHA
  `-roots all`); 164 → 172 advisories, **żaden z dotychczasowych 164
  werdyktów się nie zmienił** (offline, ten sam cache). Nowe: PostgreSQL
  CVE-2022-21724, CVE-2022-41946, CVE-2024-1597 `REACHABLE` przez sterownik
  rejestrowany w `META-INF/services` (domyślna konfiguracja WebGoata używa
  HSQLDB — użycie zależy od URL JDBC), CVE-2022-31197 `NOT_REACHABLE`, 2
  `UNKNOWN`; jsoup CVE-2026-75140 `REACHABLE` przez `Jsoup.parse` lekcji XSS,
  CVE-2022-36033 `UNKNOWN`. CHA 62/65/14/23/0 → 66/65/15/26/0 w obu
  scenariuszach korzeni; RTA bez nowych zmian werdyktów.
- **Weryfikacja:** przebiegi WebGoata stare/nowe wejście × `-roots all/profile`
  offline na tym samym cache (po jednym przebiegu online uzupełniającym dane
  OSV nowych zależności); `scripts/graph-benchmark.sh --repeat 2`.
  **Skrypt nie był uruchomiony w całości** (brak JDK 17 `javac`/`jar`): kroki
  Mavena (instalacje, `build-classpath`, `copy-dependencies`, `list` dla
  `webgoat-server`) wykonane tu; kompilacja lekcji i launchera przez API
  kompilatora JDK 25 z `-proc:full` i Lombokiem 1.18.42 (`-source/-target 15`),
  bo `maven-compiler-plugin` 3.8.0 nie przekaże `-proc:full`, a JDK 23+ bez
  niego nie uruchamia Lomboka. `vulnerable-components` w zmierzonym jarze to
  dotychczasowe klasy (JDK 17).
- **Ograniczenia:** WebWolf (osobna aplikacja) poza wejściem; zasoby lekcji
  (migracje SQL, szablony) nie są analizowane.
- **Następny krok (wymaga zgody):** inny obszar (B3 — jednorazowe przebiegi,
  CVE:F8).

### ROADMAP:B3 — trzeci przyrost: wybór uczestników z profilu (`b3-3`) — DONE for this scope (2026-11-13)

- **Zakres (zatwierdzony 2026-11-13):** kryterium B3 „`internal/profile`
  pozostaje dostawcą konfiguracji i wyboru modeli” oraz zapis wyboru w
  wynikach i snapshocie (invalidation według reguł CVE F3).
- **Zrealizowane:** katalog uczestników punktu stałego w CLI
  (`cmd/ariadne/participants.go`: 5 modeli — ServiceLoader ×3,
  `servlet-container:dispatch`, `service-provider:dispatch` — i 4 bramki
  — `dynamic-linkage:sam-calls`, `servlet-container:objects`,
  `jdbc:objects`, `framework:proxies`, każdy z wersją reguł; wersja
  modeli ServiceLoader nazwana `s4`); `analysis.Select` (nieznane ID =
  błąd z listą znanych, kolejność katalogu), `analysis.Labels`;
  `models.Registry.Remove/IDs`; sekcja profilu `models: {disable: [...]}`
  (tylko dane; scalanie profili = suma list); wyłączony model usunięty z
  rejestru, wyłączona bramka nie jest krokiem ani nie dokłada krawędzi w
  podsumowaniu; `callgraph.Config.Participants`/`DisabledParticipants`
  (raport, v2 `callgraph`, `-graph-out`/`-graph-compare`), konfiguracja
  snapshotu `models`/`models-disabled`; `CoordinatorVersion` `b3-3`.
  Nazwy kroków w logu: `dynamic-linkage:sam-calls`,
  `servlet-container:objects` (były `sam-calls`,
  `servlet-container:dispatch objects`).
- **Weryfikacja:** `TestSelect`, `TestRegistry_removeAndIDs`,
  `TestModelsSection`, `TestParticipantCatalog`; `go test ./...`,
  `go vet`. Integracyjnie względem binarki z `main` (`b3-2`): petclinic,
  WebGoat, JPetStore CHA `-roots all` i WebGoat RTA `-roots profile` —
  v1 bajtowo identyczne, `-graph-compare` 0 dodanych/0 usuniętych
  krawędzi (różni się tylko ustawienie `participants`). petclinic z
  wyłączonymi `servlet-container:dispatch` i `servlet-container:objects`:
  −5 918 krawędzi, 45 ustaleń zmienionych, 7/96/17/27/6 → 7/83/30/27/6.
  Literówka w ID: przebieg kończy się przed analizą z listą znanych ID.
  `scripts/graph-benchmark.sh` nie był powtarzany: domyślny graf jest
  identyczny (`-graph-compare`), więc `testdata/graphbench` się nie zmienia.
- **Dokumentacja:** README (akapit o wyborze uczestników), `docs/SUPPORT.MD`
  (wiersz „Library model iteration foundation”), ten rejestr.
- **Ograniczenia:** wybór tylko przez wyłączanie (brak modeli opt-in);
  jednorazowe przebiegi (krawędzie syntetyczne, callbacki, uchwyty,
  refleksja, deserializacja) poza katalogiem; invalidation zależnych
  wyników po zmianie samego modelu nadal tylko przez wersje algorytmów.
- **Następny krok (wymaga zgody):** jednorazowe przebiegi do katalogu
  uczestników albo inny obszar (pełne wejście WebGoata, CVE:F8).

### GRAPH_PRECISION / FRAMEWORKS F2 — launcher Boot WebGoata w wejściu (`webgoat-boot-launcher`) — DONE for this scope (2026-11-13)

- **Zakres (zatwierdzony 2026-11-13):** domknięcie wielokrotnie odkładanego
  kroku „WebGoat bez launchera Boot w wejściu” — zmiana przygotowania
  wejścia, nie modelu.
- **Przyczyna:** `scripts/prepare-webgoat.sh` budował tylko
  `webgoat-container` i lekcję vulnerable-components; `StartWebGoat`
  (`@SpringBootApplication`, `@ServletComponentScan`) jest w
  `webgoat-server`, więc żadna klasa wejścia nie włączała
  auto-konfiguracji Boot, model `servlet-container:dispatch` nie
  rejestrował nic (0 servletów/filtrów) i cała ścieżka żądania MVC była
  nieosiągalna.
- **Zmiana:** skrypt kompiluje dwie klasy `webgoat-server` (`StartWebGoat`,
  `HSQLDBDatabaseConfig`) przez `javac` z classpathem kontenera i jego
  zależności oraz Lombokiem w wersji z pom-u rodzica (`LOMBOK_JAR` do
  nadpisania) i dokłada je do `app-classes.jar`. Mavenem się nie da:
  `webgoat-server` zależy od wszystkich lekcji, w tym cross-site-scripting,
  które nie buduje się pod JDK 17. Skrypt ostrzega, gdy launchera brak w
  jarze.
- **Wynik (WebGoat):** 6 rejestracji (`DispatcherServlet`,
  `OrderedCharacterEncodingFilter`, `OrderedFormContentFilter`,
  `OrderedRequestContextFilter`, `DelegatingFilterProxy`,
  `FilterChainProxy`); +1 709 krawędzi (CHA, `-roots all`); CHA
  62/53/26/23/0 → 62/65/14/23/0 w obu scenariuszach korzeni — 12
  `NOT_REACHABLE` → `FRAMEWORK` (zasoby statyczne CVE-2026-22741/-22745/
  -41841/-41842/-41843/-41844, CVE-2024-38819, CVE-2025-22233,
  CVE-2022-22968, Thymeleaf CVE-2021-43466/-40478/-41901) przez
  dispatch Undertow → `DispatcherServlet`; to dokładnie podejrzewane
  fałszywie negatywne z wpisu CVE:F10 (kontynuacja). RTA `-roots
  profile`: **CVE-2024-38809 przestaje być fałszywie negatywne**
  (otwarte od GRAPH_PRECISION P4); CVE-2026-40477 zostaje `REACHABLE` w
  obu trybach (z osiągalnym renderowaniem widoków
  `ThymeleafEvaluationContext` jest żywy dla RTA — nadal nadmiarowa
  ścieżka z doradcy Spring Security). petclinic i JPetStore bez zmian.
- **Weryfikacja:** WebGoat ze starym i nowym `app-classes.jar` w obu
  scenariuszach korzeni, ręcznie sprawdzona ścieżka CVE-2026-22745
  (`ServletHandler#handleRequest` → `HttpServlet#service` →
  `DispatcherServlet#doService`); `scripts/graph-benchmark.sh --repeat 2`.
  **Skrypt nie był uruchomiony w całości** — na tej maszynie nie ma `javac`
  JDK 17; te same kroki (kompilacja dwóch plików z Lombokiem, dołożenie
  do jara) wykonane ręcznie przez API kompilatora JDK 25 z Lombokiem
  1.18.42, `-source/-target 15` jak reszta WebGoata.
- **Ograniczenia:** (rozwiązane wpisem `webgoat-full-input` powyżej) wejście nadal zawiera tylko jedną lekcję (pozostałe ~30
  lekcji wdrożenia `webgoat-server` są poza nim).
- **Następny krok (wymaga zgody):** pełne wejście WebGoata (wszystkie
  lekcje) albo ROADMAP B3.

### CVE:F10 — „kod nieobecny”: `MAPPED_CODE_ABSENT` — DONE (2026-11-16)

- **Decyzja (zatwierdzona 2026-11-16):** gdy żadna klasa wskazana przez
  wybrany wariant mapowania (także po relokacji instancji) nie istnieje na
  analizowanym classpath, ustalenie jest `NOT_REACHABLE` z nowym kodem
  `MAPPED_CODE_ABSENT`, a notatka wymienia klasy. Klasa obecna bez metody
  zostaje `METHOD_NOT_FOUND` (`UNKNOWN`). Reguła luk B0 nie dotyczy takich
  ustaleń (klasy nie ma, więc żadne nierozstrzygnięte miejsce jej nie
  uruchomi). Wiersz v1 niesie kod, gdy niosą go wszystkie ustalenia
  (agregacja §5). Założenie jak przy każdym `NOT_REACHABLE`: wejście to
  wdrażany classpath.
- **Wynik:** `UNKNOWN` (v2, `-roots all` CHA) 3/1/3 → 1/0/0: JPetStore
  CVE-2018-11040/-1257/-1271 (brak `spring-webmvc`/`spring-messaging`),
  petclinic CVE-2026-34483/-34487 (`tomcat-embed-core` 9.0.45 bez
  `JsonAccessLogValve` i `tomcat-tribes`), WebGoat CVE-2021-22119 (brak
  `spring-security-oauth2-client`). Zostaje petclinic CVE-2022-21363.
- **Weryfikacja:** `TestAnalyze_unmatchedMethodIsMethodNotFound` (klasa
  obecna / część obecna / żadna), `TestFindingsForUnmatchedVerdicts`
  (instancja i wiersz), `TestReasonCodesDocumented`; `go test ./...`,
  `go vet ./...`; porównanie v1 z `main` w 12 konfiguracjach (petclinic/
  WebGoat/JPetStore × `-roots all`/`profile` × CHA/RTA): zmieniają się
  wyłącznie te 6 advisory. Dokumentacja: RESULTS_CONTRACT (§2, §3), README.

### CVE:F10 (redukcja `UNKNOWN`, część 3: bazy, sterowniki, reszta + `NOT_CODE_MAPPABLE`) — DONE for this scope (2026-11-16)

- **Mapowania (9, metoda jak w częściach 1–2):** H2 CVE-2021-42392 i
  CVE-2022-23221 (Console: `WebApp#login`/`#test`, `entrypoint`),
  MySQL CVE-2021-2471 (`MysqlSQLXML#getSource`), CVE-2023-22102
  (`ResultSetImpl#getObject` — 8.2.0 usuwa `autoDeserialize`; warianty
  5.1 / 8.0 / `mysql-connector-j`), pgjdbc CVE-2022-26520
  (`Driver#setupLoggerFromProperties`) i CVE-2026-42198 (krok SCRAM z
  liczbą iteracji: `jre8`/`jre7` `processServerFirstMessage`, od 42.7.4
  `core.v3` `handleAuthenticationSASLContinue`), jsoup CVE-2022-36033
  (`Cleaner#createSafeElement`), hibernate-validator CVE-2023-1932
  (`SafeHtmlValidator#isValid`), Spring Boot CVE-2026-22733 (wykluczenie
  ścieżek CloudFoundry z bezpieczeństwa: konfigurator 2.0–2.3,
  `WebSecurityCustomizer` 2.4–3.4, łańcuch filtrów 3.5; 4.x poza modułem —
  `MAPPING_VERSION_UNCOVERED`). Grupa C okazała się mniejsza: CVE-2026-22733
  dotyczy kodu (autokonfiguracji), więc dostała mapowanie.
- **`NOT_CODE_MAPPABLE` (nowy kod, kontrakt):** wpis `not_code: "<powód>"`
  w tabeli (bez metod, wariantów, roli i pokrycia) → status
  `NOT_APPLICABLE` (jak WebJar: graf nie może odpowiedzieć; nie werdykt
  „bezpieczne”), powód w notatce; v1 `reasons`, v2 `reason_codes`, log
  „no JVM code=N”. `ResolverVersion` 6 → 7. Jedyny wpis: CVE-2025-49124
  (instalator Tomcata dla Windows). Dokumentacja: RESULTS_CONTRACT §2–3,
  README.
- **Bez mapowania:** CVE-2022-21363 (MySQL 8.0.28) — advisory Oracle nie
  nazywa komponentu, diff 8.0.27→8.0.28 (62 metody: TLS, MFA, nawigacja
  wyników) nie ma zmiany do przypisania; zostaje `MAPPING_MISSING`.
- **Wynik:** `UNKNOWN` (v2, `-roots all` CHA offline): petclinic 10 → 3,
  WebGoat 6 → 1, JPetStore 3 → 3 (łącznie od 27/26/5). Zmiany (12): 5 →
  `NOT_REACHABLE` (H2 Console ×2, `SQLXML`, CloudFoundry ×2 — petclinic i
  WebGoat nie działają na Cloud Foundry, ścieżka nieosiągalna), 2 →
  `FRAMEWORK` (`@SafeHtml` przez walidator), 1 → `NOT_APPLICABLE`
  (CVE-2025-49124), 4 → `REACHABLE` — wszystkie nadmiarowe lub zależne od
  konfiguracji, także przy RTA: WebGoat CVE-2022-26520 i CVE-2026-42198
  (przez `java.sql.Driver#connect` do sterownika PostgreSQL; WebGoat łączy
  się z HSQLDB), CVE-2022-36033 (`Element#text` → dowolny `NodeVisitor` →
  `Cleaner`; lekcja tylko parsuje, nie czyści), petclinic CVE-2023-22102
  (ścieżka przez H2 do `ResultSet#getObject` MySQL; MySQL jest tylko w
  profilu `mysql`). Mapowania poprawne, nieprecyzja grafu na interfejsach
  JDBC/jsoup.
- **Pozostaje `UNKNOWN`:** kod nieobecny w jarze (`METHOD_NOT_FOUND`):
  JPetStore CVE-2018-11040/-1257/-1271, petclinic CVE-2026-34483/-34487,
  WebGoat CVE-2021-22119 (`spring-security-core` bez klas OAuth2);
  petclinic CVE-2022-21363. Decyzja, czy „kod nieobecny” ma dawać inny
  status, pozostaje otwarta (zmiana kontraktu).
- **Weryfikacja:** `TestManualNotCode`, `TestFindingsForUnmatchedVerdicts`
  (`NOT_APPLICABLE` + `NOT_CODE_MAPPABLE`), `TestServerManualMappings`
  (warianty części 3), `TestLoadManualMeta_embeddedTable` (107 wpisów),
  `TestReasonCodesDocumented`; `go test ./...`, `go vet ./...`;
  porównanie v2 przed/po (zmieniło się tylko 12 omówionych), v1/Markdown
  petclinic dla CVE-2025-49124; RTA dla 4 nowych `REACHABLE`.

### CVE:F10 (redukcja `UNKNOWN`, część 2: Tomcat, Jackson, logback, Undertow/XNIO) — DONE for this scope (2026-11-16)

- **Metoda:** jak w części 1 (diff znormalizowanego bytecode'u wydań w
  `work/jardiff`, który dostał `-list` do wypisania metod klasy; każda
  metoda sprawdzona z nazwy w wydaniach wariantu). Gdy commit poprawki
  nazywa klasę, której badana linia nie ma (parser async Jacksona od 2.14,
  procesory modelu logbacka od 1.3), wariant starszej linii wskazuje metodę
  z tym samym kodem.
- **Wynik:** 14 wpisów (83 → 97; role 53 entrypoint / 8 sink / 36
  vulnerable). Tomcat: CVE-2024-56337 (`DefaultServlet#doPut`, zapis przy
  `readonly=false`; poprawka 9.0.99 w `DirResourceSet`), CVE-2026-34483
  (`JsonAccessLogValve$JsonWrappedElement#addElement`), CVE-2026-65905
  (`DigestAuthenticator$NonceInfo#nonceCountValid`). Jackson:
  CVE-2026-68494 (liczby w parserze async: `NonBlockingJsonParser` < 2.14,
  `NonBlockingUtf8JsonParserBase` od 2.14 i w 3.x), CVE-2026-50193
  (`InternalNodeMapper#nodeToString`/`nodeToPrettyString`), CVE-2026-54514
  (`FromStringDeserializer$Std#_deserialize` — wspólna gałąź typów JDK,
  nadmiarowe przybliżenie przypadku `InetSocketAddress`; 2.3:
  `InetSocketAddressDeserializer`), CVE-2026-54515
  (`BeanDeserializerBase#createContextual`, 2.x i 3.x). logback:
  CVE-2025-11226 (`IfAction#begin` < 1.3, `IfModelHandler#handle`),
  CVE-2026-1225 (`NestedComplexPropertyIA#begin` < 1.3,
  `ImplicitModelHandler#doComplex`), CVE-2026-10532
  (`HardenedObjectInputStream#<init>`). Undertow: CVE-2021-3597
  (`AbstractFramedStreamSinkChannel#flushComplete`), CVE-2021-3629
  (`Http2Channel#updateSettings`), CVE-2026-3260 (`MultiPartParserDefinition$
  MultiPartUploadHandler#parseBlocking`, brak poprawki). XNIO:
  CVE-2020-14340 (`NioXnio#getSelector`).
  `UNKNOWN` (v2, ustalenia, `-roots all` CHA offline): petclinic 19 → 10,
  WebGoat 17 → 6, JPetStore 3 → 3. Zmiany (20): 12 → `FRAMEWORK` (logback przy
  konfiguracji, Jackson przez konwertery Spring MVC, DIGEST przez potok
  wentyli Tomcata), 4 → `NOT_REACHABLE` (petclinic nie zapisuje przez
  `DefaultServlet`, nie wywołuje `JsonNode#toString`; żadna aplikacja nie
  uruchamia odbiorników gniazd logbacka), 4 → `REACHABLE` w WebGoat:
  CVE-2026-3260 i CVE-2020-14340 (wiarygodne: `VerifyAccount` woła
  `getParameterNames`, co parsuje treść żądania — multipart na dysk,
  blokujący odczyt przez selektor wątku), CVE-2021-3597 i CVE-2021-3629
  (ścieżki przez kanały HTTP/2; WebGoat nie włącza HTTP/2, więc to
  nadmiarowe przybliżenie zależne od konfiguracji — także przy RTA).
- **Pozostaje (`METHOD_NOT_FOUND`, kod nieobecny w jarze):** petclinic
  CVE-2026-34483 (`tomcat-embed-core` 9.0.45 nie ma `JsonAccessLogValve` —
  jest w nim od linii 9.0.6x) i CVE-2026-34487 (członkostwo w chmurze z
  `tomcat-tribes`, którego jar embed nie zawiera) — ten sam przypadek co
  JPetStore CVE-2018-11040/-1257/-1271 (decyzja o kontrakcie otwarta).
- **Weryfikacja:** `TestServerManualMappings` (wybór wariantu wg wersji,
  wydanie naprawione nic nie wybiera), `TestLoadManualMeta_embeddedTable`
  (97 wpisów), `go test ./...`, `go vet ./...`; porównanie ustaleń v2
  przed/po na trzech aplikacjach (zmieniły się tylko omówione wyżej, plus
  powód CVE-2026-34483 `MAPPING_MISSING` → `METHOD_NOT_FOUND`); WebGoat
  ponownie przy RTA (te same 4 `REACHABLE`). Dokumentacja: README.
- **Następny krok:** część 3 — zrobiona, sekcja wyżej.

### CVE:F10 (redukcja `UNKNOWN`, część 1: Spring) — DONE for this scope (2026-11-16)

- **Zakres (zatwierdzony 2026-11-16):** przegląd 41 unikalnych `UNKNOWN`
  na petclinic/WebGoat/JPetStore (27/26/5) — grupy: A brak commita
  poprawki i nazwy klasy (~27), B mapowanie z commita na metody nieobecne w
  badanej linii (9), C advisory niedotyczące kodu (~5); ręczne mapowania z
  diffów wydań w 2–3 PR-ach (Spring; Tomcat/Jackson/logback/Undertow; H2,
  sterowniki JDBC, reszta + kod powodu dla grupy C). Ta część: Spring.
- **Metoda:** jak 2026-11-13 (`work/jardiff`: znormalizowany bytecode
  ostatniego wydania podatnego vs naprawionego z Maven Central, przypisanie
  zmian po treści advisory, sprawdzenie każdej metody z nazwy w wydaniach
  wariantu). Gdy poprawka zmienia współdzielone narzędzie wykonywane i tak
  przy starcie (`StringUtils#cleanPath`, `UrlResource#createRelative`,
  cache ścieżek właściwości Spring Data), wpis nazywa punkt wejścia z
  treści advisory (`entrypoint`).
- **Wynik:** 15 wpisów (68 → 83; role 49 entrypoint / 8 sink / 26
  vulnerable): CVE-2016-1000027 (eksportery HTTP invoker/RMI, poprawki brak
  — commity zmieniają tylko Javadoc), CVE-2018-11040/-1257/-1271/-15756
  (linia 4.3; dla 15756 osobny wariant 4.2 — `HttpRange#toResourceRegions`
  powstało w 4.3), CVE-2024-38828 (`ByteArrayHttpMessageConverter#
  readInternal`, 5.3.39 vs 6.0.0 — bufor z `Content-Length`),
  CVE-2025-41242 (`PathResourceResolver#getResource`; poprawka w
  `spring-core` `createRelative`), CVE-2025-22235 (`EndpointRequest#to`),
  CVE-2026-40972 (`HttpHeaderAccessManager#isAllowed`), CVE-2026-41711/
  -41716/-41721 (resolvery argumentów web Spring Data; jedno wydanie
  3.5.11→3.5.12 naprawia wszystkie trzy), CVE-2021-22119 (repozytorium
  żądań autoryzacji OAuth2, warianty dla `spring-security-oauth2-client` i
  `-core`), CVE-2026-22732 (`OnCommittedResponseWrapper#addHeader`),
  CVE-2026-47838 (`SubjectDnX509PrincipalExtractor#extractPrincipal`).
  `UNKNOWN` (v2, ustalenia, `-roots all` CHA offline): petclinic 27 → 19,
  WebGoat 26 → 17, JPetStore 5 → 3. Zmiany: 11 → `FRAMEWORK` (ścieżki przez
  wnętrze Springa), 6 → `NOT_REACHABLE` (aplikacje nie wołają
  `EndpointRequest.to()`, nie używają zdalnych DevTools, HTTP invokera
  (JPetStore), zakresów zasobów (JPetStore)), 2 → `REACHABLE` w WebGoat:
  CVE-2026-22732 (ścieżka wiarygodna: nagłówki bezpieczeństwa zapisywane
  przez `OnCommittedResponseWrapper#addHeader`) i CVE-2026-47838
  (nadmiarowa ścieżka CHA `ClearSiteDataHeaderWriter` → dowolny
  `RequestMatcher` → filtr X.509; WebGoat nie używa X.509; także przy RTA,
  bo konfigurator Spring Security tworzy filtr) — mapowanie poprawne,
  nieprecyzja grafu.
- **Pozostaje w tej grupie:** JPetStore CVE-2018-11040/-1257/-1271 —
  `METHOD_NOT_FOUND`, bo aplikacja nie ma `spring-webmvc` ani
  `spring-messaging` (kod podatny nieobecny; obecny kontrakt: `UNKNOWN`);
  CVE-2026-22733 (CloudFoundry, zależne od wdrożenia — grupa C).
- **Weryfikacja:** `TestLoadManualMeta_embeddedTable` (83 wpisy),
  `go test ./...`, `go vet ./...`; porównanie ustaleń v2 przed/po na trzech
  aplikacjach (każda zmiana wyjaśniona wyżej). Dokumentacja: README.
- **Następny krok:** część 2 (Tomcat, Jackson, logback, Undertow/XNIO) — zrobiona, sekcja wyżej.

### CVE:F10 (kontynuacja) — ręczne mapowania z diffów wydań — DONE for this scope (2026-11-13)

- **Archiwum:** [archive/2026-11-13/cve/CVE_F10.md](../archive/2026-11-13/cve/CVE_F10.md) (zmergowane: PR #48 `4b88699`, PR #49 `8d9ff45`).

- **Zakres (zatwierdzony 2026-11-13):** największe grupy advisories
  wciąż niezmapowanych po F10 — bez commita poprawki i bez nazwy klasy w
  tekście, więc ani źródło diffu, ani `release-diff` nie miały z czego
  korzystać: Spring webmvc/expression, Thymeleaf, SnakeYAML; dodatkowo
  dwa advisories Springa o wstrzykiwaniu do logów (spring-core).
- **Metoda:** dla każdej poprawki ręcznie przeczytany diff wydań
  (znormalizowany bytecode `classfile.MethodFingerprint`, ostatnie wydanie
  przed poprawką vs wydanie z poprawką, Maven Central), zmienione metody
  przypisane do advisory po jego treści, potem każda metoda odnaleziona z
  nazwy w bytecode całego zakresu wydań wariantu (np. webmvc 5.3.0, 5.3.4,
  5.3.6, 5.3.20, 5.3.39, 6.0.0, 6.0.23, 6.1.0, 6.1.21, 6.2.0 i ostatnie 6.2
  przed poprawką; SpEL 4.3.0–6.2.18 na 13 wydaniach; Thymeleaf
  3.0.11–3.1.4 na 8; SnakeYAML 1.17–1.31 na 8). Poprawki Springa z 2026 r.
  są na Central tylko dla 6.2.x (5.3.x komercyjne) — warianty 5.3/6.0/6.1
  nazywają te same metody tych samych klas. Wersja spoza sprawdzonego
  zakresu → `MAPPING_VERSION_UNCOVERED`.
- **Wynik:** 22 wpisy schematu v2 (46 → 68): SnakeYAML DoS ×5 i SpEL DoS
  jako `entrypoint` (`Composer#getSingleNode/getNode`,
  `SpelExpression#getValue`), reszta `vulnerable` (webmvc ×8, SpEL ×2,
  `LogFormatUtils#formatValue` ×2, Thymeleaf ×4). Szczegóły i decyzje
  w `testdata/cve-control/MANUAL_MAPPING_AUDIT.MD`.
- **Wpływ (`-roots all`, CHA):** `UNKNOWN` petclinic 49 → 27, WebGoat
  45 → 23, JPetStore 8 → 5; zmieniają się wyłącznie dotychczasowe
  `UNKNOWN`. petclinic 7/96/17/27/6, WebGoat 62/53/26/23/0, JPetStore
  2/13/15/5/0. Nowe `REACHABLE`: CVE-2026-41850 (obie aplikacje, ścieżki
  do ewaluacji SpEL przez Spring Data / Spring Security — wejście, nie
  dowód, że wyrażenie jest od atakującego), CVE-2026-40477 na WebGoat
  (przez `PreInvocationAuthorizationAdviceVoter`, długa ścieżka CHA).
- **Ręcznie sprawdzone wątpliwe werdykty** (WebGoat — rozwiązane wpisem `webgoat-boot-launcher` powyżej): na WebGoat zasoby statyczne
  (CVE-2026-22741/-22745/-41841/-41842/-41843), CVE-2021-43466 i
  CVE-2026-40478/-41901 są `NOT_REACHABLE` z `MODEL_INCOMPLETE`, a na
  petclinic `FRAMEWORK` — to znana luka WebGoata (launcher poza wejściem,
  0 zarejestrowanych servletów, więc `DispatcherServlet` nie jest
  osiągalny), prawdopodobnie fałszywie negatywne. CVE-2026-41853 na
  WebGoat `FRAMEWORK` przez `UndertowHttpHandlerAdapter` (reaktywny
  adapter, artefakt CHA w aplikacji servletowej). CVE-2026-22735 (SSE)
  `NOT_REACHABLE` na obu — żadna nie używa `SseEmitter`.
- **Weryfikacja:** `TestReleaseDiffManualMappings` (każde sprawdzone
  wydanie pokryte dokładnie jednym wariantem z oczekiwanymi metodami,
  wydanie z poprawką i wersje spoza zakresu bez wariantu),
  `TestLoadManualMeta_embeddedTable` (68 wpisów: 44 entrypoint, 8 sink,
  16 vulnerable); `go test ./...`, `go vet ./...`; trzy aplikacje offline
  przed/po; `scripts/graph-benchmark.sh --repeat 2` — grafy bez zmian
  (liczby krawędzi identyczne we wszystkich 12 konfiguracjach), dwa nowe
  wiersze `verdict-changes.tsv` (WebGoat `-roots profile`, RTA usuwa
  artefakty CHA: CVE-2026-41853 przez reaktywny adapter Undertow,
  CVE-2026-40477 przez `EvaluationContext` Spring Security → Thymeleaf).
- **Ograniczenia:** gdy jedno wydanie poprawia dwa advisories
  (CVE-2026-22741/-22745, CVE-2026-40477/-40478), podział jego metod
  między nie wynika z treści advisories (decyzja ręczna);
  CVE-2026-41842/-41843 dzielą jedną listę. Bez mapowania zostają
  CVE-2025-41242 (webmvc 6.2.9 → 6.2.10 nie zmienia żadnej metody webmvc)
  i CVE-2024-38828 (poprawka tylko w komercyjnym 5.3.42).
- **Następny krok (wymaga zgody):** pozostałe `UNKNOWN` to w większości
  advisories bez kodu w JVM (instalator Tomcata, konfiguracja) albo
  produkty bez publicznego kodu poprawki (MySQL Connector/J, Oracle);
  albo inny obszar.

### CVE:F10 — pokrycie mapowań bez commita poprawki — DONE for this scope (2026-11-10)

- **Archiwum:** [archive/2026-11-13/cve/CVE_F10.md](../archive/2026-11-13/cve/CVE_F10.md) (zmergowane: PR #48 `4b88699`, PR #49 `8d9ff45`).

- **Pomiar wyjściowy:** 69 niezmapowanych advisories (petclinic, WebGoat,
  JPetStore): 52 bez commita poprawki w OSV, 10 tylko z PR, 5 z diffem bez
  metod, 2 inne. Szukanie numeru CVE w commitach: 1 trafienie na 8.
- **A — pull requesty:** odnośnik `…/pull/N` → commit merge (API GitHub);
  merge dwurodzicowy czytany względem pierwszego rodzica (`compare`),
  squash jak zwykły commit; PR niezmergowany i PR-y w
  `github/advisory-database` pomijane z adnotacją; brak metadanych/diffu
  PR (także offline) jest błędem odnośnika, nie przebiegu (snapshot nadal
  odrzuca).
- **B — inicjalizatory pól:** `javadiff` przypisuje zmienione inicjalizatory
  i bloki inicjalizacyjne do `<clinit>`/`<init>(*)` (wszystkie
  konstruktory, `BytecodeAllConstructors`); w mapowaniu tylko gdy poprawka
  nie zmieniła żadnej metody (inaczej ETag CVE-2024-38809 mapowałby się na
  `<clinit>` osiągalny przy każdym użyciu klasy). Wynik: CVE-2018-1272
  (`MimeTypeUtils#<clinit>`), CVE-2018-11040, CVE-2026-83557;
  CVE-2016-1000027 i Tomcat CVE-2025-49124 słusznie bez metod (javadoc /
  instalator).
- **C — `release-diff`:** metody zmienione między wydaniem przed poprawką a
  wydaniem patch z poprawką (Maven Central, normalizowany bytecode
  `classfile.MethodFingerprint`), w klasach nazwanych w advisory;
  ta sama linia major.minor, dopasowanie nazwy wersji (`.Final`,
  `.RELEASE`), konstruktory/`<clinit>` tylko bez innych zmian, ≤ 40 metod.
  **Walidacja** na 130 mapowaniach z commitów: precyzja metod 95 % (52/55),
  pokrycie w nazwanych klasach 85 %
  (`testdata/cve-control/RELEASE_DIFF_VALIDATION.MD`) → domyślnie włączone
  (`-release-diff=false`). Ruch do Central szeregowany co 750 ms z własnym
  User-Agentem po odpowiedzi HTTP 429 przy pierwszym podejściu; offline
  zapas z `~/.m2/repository` (nigdy w odtwarzaniu snapshotu).
- **Wynik:** `UNKNOWN` petclinic 57 → 49, WebGoat 60 → 45, JPetStore 12 → 8
  (`-roots all`, CHA). Zmieniają się tylko dotychczasowe `UNKNOWN` oraz
  CVE-2024-6162 na WebGoat (`NOT_REACHABLE` → `REACHABLE`: backport z PR
  #1612 do linii WebGoata). Ręcznie sprawdzone mapowania `release-diff`
  (UriComponentsBuilder, RegexRequestMatcher, LoadBalancerDrainingValve,
  ApplicationTemp, HardenedObjectInputStream, CookieRequestCache) oraz
  JPetStore CVE-2020-26945 (MyBatis, `<cache />` we wszystkich mapperach →
  `SerializedCache#deserialize` — ścieżka zgodna z konfiguracją;
  wykorzystanie wymaga wpływu na zawartość cache).
- **Weryfikacja:** testy `TestFixDiff_pullRequest*`,
  `TestFixDiff_fieldInitializers`,
  `TestFixDiff_constructorNextToFieldInitializer`, `TestInitializers`,
  `TestReleaseDiff_*`, `TestReleaseBefore`; `go test ./...`,
  `go vet ./...`, `scripts/graph-benchmark.sh --repeat 3` (offline).
  `ResolverVersion` 6, flaga `release-diff` w konfiguracji snapshotu, v2
  `release_diffs`.
- **Przy okazji:** naprawiony regex `scripts/graph-benchmark.sh` po zmianie
  linii logu w B3 (`b3-2`, PR #47) — benchmark nie działał od tamtej
  zmiany.
- **Ograniczenia:** advisory bez nazwy klasy lub z poprawką w nowym wydaniu
  minor zostaje `UNKNOWN`; `release-diff` to heurystyka (unreviewed); PR z
  rebase wielu commitów czytany przez ostatni commit.
- **Następny krok (wymaga zgody):** ręczne mapowania dla największych
  pozostałych grup (Spring webmvc/expression bez nazw klas, Thymeleaf,
  SnakeYAML) albo inny obszar.

### ROADMAP:B3 — drugi przyrost: wspólny punkt stały modeli i bramek twórców (`b3-2`) — DONE for this scope (2026-11-10)

- **Zakres (zatwierdzony):** bramki twórców (SAM, obiekty servletów,
  JDBC, proxy frameworków) jako kroki koordynatora `internal/analysis`
  obok modeli z rejestru; jeden punkt stały z raportowanym budżetem;
  niezależność od kolejności. Poprawka nieaktualnego wiersza ROADMAP w
  tabeli (koordynator był w CLI od 2026-10-25).
- **Zrealizowane:** `analysis.Options.Steps` (`Step{Name, Apply}`),
  iteracja = wszystkie kroki, potem przebieg modeli; koniec, gdy nic nie
  dodano; `Result.StepEdges`, `LastAdders` (uczestnicy wciąż dodający
  przy wyczerpaniu budżetu); `CoordinatorVersion` `b3-2` w snapshotach
  (`analysis.coordinator`). `cmd/ariadne`: dawna sekwencja bramki →
  modele → bramki zastąpiona krokami; linia logu `models and gates
  (ROADMAP B3 fixed point): …`.
- **Weryfikacja:** `TestRun_stepsAndModelsShareOneFixedPoint` (łańcuch
  model → bramka → model; ta sama fixtura pokazuje, że dawna sekwencja
  gubi ostatnią krawędź), `TestRun_stepBudgetIsVisible`,
  `TestRun_orderDoesNotChangeTheGraph`. Integracyjnie 24 przebiegi
  (petclinic, WebGoat, JPetStore, aplikacje kontrolne log4j/logback ×
  `-roots profile/all` × CHA/RTA) z `-graph-compare` względem binarki
  sprzed zmiany: 0 krawędzi dodanych/usuniętych, `results.json` bajtowo
  identyczne; punkt stały w 1–5 iteracjach (budżet 14–62), żaden przebieg
  nie dobił do limitu („incomplete” pochodzi z istniejących faktów
  ServiceLoader). `go test ./...`, `go vet ./...`.
- **Ograniczenia / pozostałe kryteria B3:** jednorazowe przebiegi
  (krawędzie syntetyczne, callbacki, uchwyty, refleksja, deserializacja)
  zostają przed punktem stałym — dodają krawędzie we wszystkich miejscach
  niezależnie od osiągalności (bez korzeni), więc ich kolejność nie
  zmienia wyniku; brak wyboru modeli z profilu; brak invalidation
  zależnych wyników poza wersjami algorytmów w snapshotach.
- **Dokumentacja:** README, `docs/SUPPORT.MD`, ten rejestr.
- **Następny krok (wymaga zgody):** wybór modeli z profilu albo inny
  obszar (RUNTIME N2, FRAMEWORKS F4).

### FRAMEWORKS:F3 — wycinek 3: katalogowy model proxy (`internal/proxycatalog`, `framework:proxies` `x1`) — DONE for this scope (2026-11-10)

- **Zakres (zatwierdzony):** A — mappery MyBatis (`MapperProxy`) i
  `SqlSessionTemplate` (`SqlSessionInterceptor`); B — repozytoria Spring
  Data (`JdkDynamicAopProxy`) i proxy JPA Spring ORM (`EntityManager`,
  `Query`, `EntityManagerFactory`). Zbiory interfejsów z katalogu, nie
  wnioskowane (wycinek 2: nieograniczone na wszystkich zmierzonych
  miejscach).
- **Zrealizowane:** `internal/proxycatalog` — katalog (bytecode mybatis
  3.4.1–3.5.16, mybatis-spring 1.3.0–3.0.3, spring-aop/orm
  4.3.4–6.1.14), wykrywanie mapperów (namespace XML mappera, `@Mapper`,
  `@MapperScan`, `<mybatis:scan>`, `MapperScannerConfigurer`),
  repozytoria (aplikacyjne interfejsy pod `Repository`); krawędzie
  kandydujące do istniejącej bramki twórcy (handler musi być tworzony
  przez osiągalny kod). `cmd/ariadne`: czwarta bramka we wspólnej pętli,
  linia logu, snapshot `framework.proxies`.
- **Weryfikacja:** `TestFindMappers`, `TestFixedInterfacesAndCallerClass`;
  ślady runtime: petclinic (nowe nagranie) relacje `proxy` poza grafem
  13 → 0 (wszystkie 148 → 127), JPetStore 8 → 2; `-graph-compare` z
  `main`: JPetStore `-roots all` CHA **CVE-2022-41853 `FRAMEWORK` →
  `REACHABLE`** (zapytania aplikacji przez mappery), petclinic/WebGoat
  bez zmian werdyktów (63/23 ścieżki krótsze przez CHA w łańcuchu Spring
  AOP — nadaproksymacja, sprawdzone ręcznie). `go test ./...`,
  `go vet ./...`, `scripts/graph-benchmark.sh --repeat 3`.
- **Dokumentacja:** README, FAQ, `docs/SUPPORT.MD` (nowy wiersz),
  `TRACE_GAPS.MD`, `RESULTS.MD`, `results.tsv`, ten rejestr.
- **Ograniczenia:** krawędź kończy się na handlerze (`Method#invoke` to
  refleksja); mappery rejestrowane w kodzie albo skanowane przez XML
  spoza wejścia (`WEB-INF/*.xml` w WAR) niewykryte; interceptor
  `SqlSessionTemplate` odcięty, gdy szablon tworzy tylko refleksja
  Springa; CHA za `JdkDynamicAopProxy#invoke` rozwija cały łańcuch AOP.
- **Łańcuch interceptorów Spring AOP — zmierzony, NOT STARTED
  (2026-11-10):** górna granica — usunięcie wszystkich krawędzi CHA z
  `ReflectiveMethodInvocation#proceed`/`JdkDynamicAopProxy#invoke`/
  `CglibAopProxy` (petclinic 274, WebGoat 281, JPetStore 244) nie odbiera
  ostatniej ścieżki żadnemu dodatniemu ustaleniu na trzech aplikacjach;
  reguła stopu, nic nie zbudowano (`TRACE_GAPS.MD`).
- **Następny krok (wymaga zgody):** `ROADMAP:B3` (koordynator w CLI).

### FRAMEWORKS:F2 — komponenty Joran logback spoza konfiguracji (`l2-2`) i builder `6` — DONE for this scope (2026-11-10)

- **Zakres (zatwierdzony):** rozszerzenie `l2` na komponenty, które tworzy
  tylko Joran (implementacje `LifeCycle`), żeby usunąć artefakty CHA po
  `LifeCycle#start`/`#stop` na petclinic/WebGoat.
- **Zrealizowane:** `internal/logbackconfig` (`l2-2`): konfiguracja
  zbiera każdy atrybut `class` i elementy bez `class`; komponent
  Joran-only = konkretna klasa `ch.qos.logback.*` z `LifeCycle`, nie
  appender, nie konwerter wzorca, bez żadnego `new` w wejściu; włączony
  przez atrybut `class`, klasę domyślną Joran (`layout`, `encoder`,
  `evaluator` → `JaninoEventEvaluator`, `discriminator` — odczytane z
  bytecode, `@DefaultClass` tylko w `SiftingAppender`), referencję z kodu
  aplikacji albo typ parametru jednoargumentowego settera. **Poprawka
  buildera (`6`):** builder `5` (i `l1`/`u1` przed nim) usuwał też ciało
  dziedziczone przez włączoną klasę po wyłączonej klasie deklarującej
  (log4j `ConsoleAppender` tracił `WriterAppender#append`) — fałszywe
  przycięcie, naprawione; test `TestExcludeAbstractBodies`.
- **Weryfikacja:** `TestDecideComponents`, `TestReadConfig`,
  `TestExcludeAbstractBodies`; aplikacja kontrolna — czwarty wariant
  (`EvaluatorFilter` z `<evaluator>` bez `class`): JVM `Assuming default
  type [JaninoEventEvaluator]`, model zgodny; CVE-2024-12798
  `NOT_REACHABLE` w wariantach bez Janino, `FRAMEWORK` w `smtp`/`filter`.
  petclinic/WebGoat vs `main`: −582/−1046 krawędzi CHA, **CVE-2024-12798
  `FRAMEWORK` → `NOT_REACHABLE`** w 8/8 przebiegach (poprawne: Boot nie
  tworzy ewaluatora); JPetStore +9 krawędzi przywróconych przez builder 6,
  werdykty bez zmian. `go test ./...`, `go vet ./...`,
  `scripts/graph-benchmark.sh --repeat 3`.
- **Zostaje:** CVE-2021-42550 `FRAMEWORK` przez `ContextJNDISelector`
  (tworzony tylko przy `-Dlogback.ContextSelector=JNDI`) — prawdziwy kod
  zależny od ustawienia JVM, poza zakresem.
- **Dokumentacja:** README, FAQ, `docs/SUPPORT.MD`,
  `testdata/graphbench/RESULTS.MD`, `results.tsv`, ten rejestr.
- **Następny krok (wymaga zgody):** model proxy MyBatis/Spring Data albo
  aplikacja do oceny `j1`.

### FRAMEWORKS:F2 — appendery logback spoza konfiguracji (`internal/logbackconfig`, `l2`) — DONE for this scope (2026-11-10)

- **Zakres (zatwierdzony):** odpowiednik `l1` dla logback, z domyślną
  konfiguracją Spring Boot (gdy nie ma pakowanego pliku logback);
  aplikacja kontrolna na JVM; pomiar na petclinic/WebGoat.
- **Zrealizowane:** `internal/logbackconfig` — katalog 20 appenderów
  logback-core/classic (bytecode 1.1.11, 1.2.3, 1.2.13, 1.3.14, 1.4.14,
  1.5.12), odczyt `logback*.xml` (z `include resource`, `import`,
  `springProfile`), domyślna konfiguracja Boot (`ConsoleAppender`;
  `RollingFileAppender` przy `logging.file.*`/`logging.path`), reguły
  „nic nie wyłączaj” (brak pliku i Boot, `logging.config`, Groovy,
  usługa `Configurator`, placeholder, include plikiem/URL-em, Joran w
  kodzie aplikacji). Wspólny `internal/appref` (referencje z kodu, jary)
  — `log4jconfig` przepięty. **Builder `5`** (`internal/callgraph`): przy
  `Exclude` ciało zadeklarowane w klasie abstrakcyjnej/interfejsie nie
  jest celem CHA, gdy każda konkretna podklasa je wybierająca jest
  wyłączona (wcześniej `SMTPAppenderBase#append` zostawał celem mimo
  wyłączonego `SMTPAppender`). Snapshot `logback.appenders`.
- **Weryfikacja:** `TestReadConfig`, `TestDecide`,
  `TestExcludeAbstractBodies`; `scripts/prepare-logback-demo.sh` (Boot
  2.4.5: domyślna, `logging.file.name`, `logback-spring.xml` z
  `SMTPAppender` + Janino) — appendery modelu = appendery root loggera na
  JVM we wszystkich trzech. `-graph-compare` z `main`: petclinic −288
  CHA, WebGoat −755 CHA, JPetStore −7 (builder 5 przy `l1`); jedyna
  zmiana ustalenia: CVE-2023-6378 (instancja logback-core) `FRAMEWORK` →
  `NOT_REACHABLE`; **żaden werdykt advisory się nie zmienił**.
  `go test ./...`, `go vet ./...`, `scripts/graph-benchmark.sh --repeat 3`.
- **Wynik ręcznej weryfikacji:** 3 ścieżki przez appendery na
  petclinic/WebGoat były artefaktami i zniknęły. CVE-2021-42550 i
  CVE-2024-12798 zostają `FRAMEWORK` przez inny artefakt: CHA po
  `LifeCycle#start`/`#stop`/`#isStarted` (np.
  `StatusListenerConfigHelper#initAndAddListener` → każdy `LifeCycle` →
  `JaninoEventEvaluatorBase#start`).
- **Dokumentacja:** README, FAQ, `docs/SUPPORT.MD` (nowy wiersz),
  `testdata/graphbench/RESULTS.MD`, `results.tsv`, ten rejestr.
- **Ograniczenia:** tylko appendery; logback-access poza katalogiem;
  tylko pakowana konfiguracja i właściwości (zmienne środowiskowe,
  `-Dlogback.configurationFile`, zewnętrzny `logging.config` nieczytane).
- **Następny krok (wymaga zgody):** rozszerzenie na komponenty Joran
  (`LifeCycle`: ewaluatory, źródła połączeń, listenery statusu) tworzone
  z konfiguracji — to zamknęłoby oba pozostałe artefakty; albo model
  proxy MyBatis / aplikacja do oceny `j1`.

### FRAMEWORKS:F2 — appendery log4j 1.2 spoza konfiguracji (`internal/log4jconfig`, `l1`) — DONE for this scope (2026-11-10)

- **Zakres (zatwierdzony):** wzorzec `u1` dla log4j 1.2 — appendery
  log4j, których pakowana konfiguracja ani kod aplikacji nie wymieniają,
  przestają być celami CHA/RTA; tylko appendery (bez layoutów, filtrów,
  logback/log4j2); aplikacja kontrolna na prawdziwym JVM.
- **Zrealizowane:** `internal/log4jconfig` — katalog 18 appenderów
  (bytecode 1.2.8–1.2.17; `RewriteAppender` od 1.2.17), odczyt
  `log4j.properties`/`log4j.xml` z korzeni jarów aplikacji i zależności
  oraz `log4jConfigLocation` z `web.xml` (tylko `classpath:` z korzenia),
  referencje z kodu aplikacji. Model nie wyłącza niczego bez pakowanej
  konfiguracji, przy placeholderze `${…}`, nieparsowalnym XML, innej
  lokalizacji w `web.xml` lub gdy kod aplikacji sam woła konfigurator.
  Znacznik obecności: `helpers.AppenderAttachableImpl` (most
  log4j-over-slf4j go nie ma). `cmd/ariadne`: reguły `Exclude` `u1` i
  `l1` łączone, linia logu, reguła w skip rule; snapshot
  `log4j.appenders`.
- **Weryfikacja:** `TestReadConfig`, `TestDecide` (krawędzie CHA przez
  dziedziczone `AppenderSkeleton#doAppend` → `this.append`).
  `scripts/prepare-log4j1-demo.sh` — te same klasy, konfiguracja
  console-only vs `JDBCAppender` → HSQLDB; JVM (`-Dlog4j.debug`, liczba
  wierszy) i ariadne zgodne: CVE-2022-23305 `NOT_REACHABLE`/`REACHABLE`,
  CVE-2021-4104 `NOT_REACHABLE` w obu (przed modelem 4× `REACHABLE`, 3
  fałszywe). JPetStore `-graph-compare`: CHA −170 krawędzi, RTA −106;
  CHA 3/9/11/12/0 → 0/10/13/12/0 (CVE-2021-4104, CVE-2022-23305 →
  `NOT_REACHABLE`, CVE-2022-41853 → `FRAMEWORK`), RTA bez zmian
  werdyktów; CHA i RTA różnią się już tylko CVE-2024-38820.
  `go test ./...`, `go vet ./...`, `scripts/graph-benchmark.sh --repeat 3`.
- **Dokumentacja:** README, FAQ, `docs/SUPPORT.MD` (nowy wiersz),
  `testdata/graphbench/RESULTS.MD`, `results.tsv`, ten rejestr.
- **Ograniczenia:** tylko pakowana konfiguracja (zewnętrzny plik może
  włączyć appender pominięty przez model); każdy znaleziony plik się
  liczy, także taki, którego log4j nie wybierze; linie kontynuacji w
  `.properties` nieczytane.
- **Następny krok (wymaga zgody):** ten sam model dla logback/log4j2
  (jeśli pomiar pokaże artefakty), katalogowy model proxy MyBatis
  (krok 4 z wpisu JPetStore) albo aplikacja do oceny `j1`.

### GRAPH_PRECISION:P4 — krok 3: ręczne mapowania HSQLDB i log4j 1.2 — DONE (2026-11-10)

- **Zakres (zatwierdzony):** ręczne mapowania (sink) dla 7 niezmapowanych
  advisories JPetStore: HSQLDB CVE-2022-41853 i log4j 1.2 CVE-2019-17571,
  CVE-2021-4104, CVE-2022-23302, CVE-2022-23305, CVE-2022-23307,
  CVE-2023-26464; sprawdzenie, czy `j1` zmienia przez nie werdykt.
- **Zrealizowane:** 7 wpisów schematu v2 w
  `internal/vulnmap/vuln_methods.yaml` (39 → 46; `role: sink`,
  `partial`, `reviewed`). Bez commitów poprawek (HSQLDB w SVN, log4j 1.2
  EOL), więc metody wskazane w bytecode: HSQLDB `Routine#getMethods`
  (`Class.forName` po nazwie z SQL) i `#invokeJavaMethod`
  (`Method.invoke`), wariant `[2.0.0,2.7.1)` sprawdzony na 2.0.0, 2.2.9,
  2.3.4, 2.5.0, 2.6.1, 2.7.0, 2.7.1 (`Routine` identyczny w 2.7.0/2.7.1,
  poprawka zmieniła tylko bramkę `supportsJavaMethod`); log4j
  `SocketNode#run`, `JMSAppender#activateOptions`/`#lookup`,
  `JMSSink#<init>`/`#onMessage`,
  `JDBCAppender#getLogStatement`/`#flushBuffer`/`#execute`, Chainsaw
  `LoggingReceiver$Slurper#run`, wariant `[1.2.8,1.2.17]` sprawdzony na 7
  wydaniach. Test `internal/vulnmap/jpetstore_mappings_test.go`.
- **Wynik (JPetStore, oba scenariusze korzeni identycznie):** CHA
  0/9/7/19/0 → 3/9/11/12/0, RTA 0/8/8/19/0 → 0/9/14/12/0. Trzy CHA
  `REACHABLE` (HSQLDB, JMSAppender, JDBCAppender) to artefakty dispatchu
  appenderów log4j (konfiguracja ma tylko `ConsoleAppender`); RTA je
  usuwa — log4j `NOT_REACHABLE` poprawnie, HSQLDB `FRAMEWORK` przez
  `EmbeddedDatabaseFactory#shutdownDatabase` → `Statement#execute`
  (ścieżka istnieje dla pewnego tekstu SQL, graf go nie modeluje).
  Nowy ślad runtime: wszystkie zmapowane metody śledzone na wejściu,
  żadna niewykonana. **`j1` nie zmienia werdyktu:** krawędź
  `Statement` daje już `service-provider:dispatch` (`p1`, produkty
  `java.sql.Driver` z META-INF/services HSQLDB) — binarka sprzed `j1`
  (`95d7c85^`, `-vuln-methods`) daje te same werdykty i ścieżki pod
  `-roots profile`.
- **Weryfikacja:** `go test ./...`, `go vet ./...` — PASS;
  `scripts/graph-benchmark.sh --repeat 3` (petclinic bez zmian);
  `scripts/runtime-trace-jpetstore.sh`.
- **WebGoat (HSQLDB 2.5.1):** CVE-2022-41853 `UNKNOWN` → `REACHABLE` we
  wszystkich 4 konfiguracjach (CHA 53/37/13/61/0 → 54/37/13/60/0, RTA
  52/35/… → 53/35/…). **Prawdziwe trafienie z warunkiem wykorzystania:**
  `SqlInjectionLesson4#completed(@RequestParam String query)` wykonuje
  parametr żądania przez `Statement#executeUpdate` — SQL od atakującego
  może utworzyć procedurę Javy. Raportowana (najkrótsza) ścieżka idzie
  przez `LessonConnectionInvocationHandler#invoke`; binarka sprzed `j1`
  daje ten sam werdykt (krawędź `Statement` z `p1`).
- **Dokumentacja:** `testdata/cve-control/MANUAL_MAPPING_AUDIT.MD`,
  `testdata/graphbench/RESULTS.MD` (JPetStore), `results.tsv`,
  `docs/SUPPORT.MD`, ten rejestr.
- **Ograniczenia:** mapowania wskazują sink, nie wejście (SQL od
  atakującego, konfiguracja appendera); wersje spoza sprawdzonego zakresu
  → `MAPPING_VERSION_UNCOVERED`; JPetStore nie jest aplikacją, na której
  `j1` cokolwiek rozstrzyga (sterownik zarejestrowany jako usługa, bez
  puli).
- **Następny krok (wymaga zgody):** rozpoznawanie servletów bez API
  servletów w wejściu (krok 2 z wpisu JPetStore) albo katalogowy model
  proxy MyBatis/Spring ORM (krok 4); dla oceny `j1` na werdyktach
  potrzebna aplikacja z pulą połączeń i sterownikiem bez rejestracji
  usługi.

### FRAMEWORKS:F2 — korzenie `web.xml` także pod `-roots all` — DONE (2026-11-09)

- **Zakres (zatwierdzony):** krok 1 z wpisu JPetStore — metody cyklu życia
  servletów/filtrów/listenerów z `web.xml` korzeniami także domyślnego
  `-roots all` (dotąd tylko `-roots profile`).
- **Zrealizowane:** `cmd/ariadne`: `-roots all` = `reach.AppRoots` +
  `profile.WebXMLAppRoots` bez duplikatów, w ustalonej kolejności
  (`appendNewRoots`); linia logu `roots: all (N methods: every application
  method, M more from web.xml)`; sekcja raportu „web.xml-declared roots”
  opisuje oba scenariusze.
- **Weryfikacja:** `TestAppendNewRoots`. JPetStore `-roots all`
  (`-graph-compare` z wersją sprzed poprawki): 312 korzeni (6 z web.xml),
  +418 krawędzi, 8 advisories `NOT_REACHABLE` → `FRAMEWORK` — te same
  werdykty co `-roots profile`; 3 z nich (CVE-2022-22965/22970,
  CVE-2026-41848) z podatną metodą wykonaną na śladzie.
  `scripts/graph-benchmark.sh --repeat 3`: petclinic i WebGoat (bez
  `web.xml`) bez zmian; JPetStore `all` RTA–CHA: CVE-2024-38820 jak w
  `profile` (sprawdzone wcześniej, poprawne). `go test ./...`,
  `go vet ./...` — PASS.
- **Dokumentacja:** README (flaga `-roots`, akapit o korzeniach web.xml),
  FAQ, `docs/SUPPORT.MD` (wiersz Profiles and selected roots),
  `RESULTS.MD`, `TRACE_GAPS.MD`, szablon raportu, ten rejestr.
- **Następny krok (wymaga zgody):** rozpoznawanie servletów bez API
  servletów w wejściu albo ręczne mapowania HSQLDB CVE-2022-41853 i log4j
  1.2 (kroki 2–3 z wpisu JPetStore).

### GRAPH_PRECISION:P4 — trzecia aplikacja pomiarowa: JPetStore 6.0.2 (WAR) — DONE for this scope (2026-11-09)

- **Zakres (zatwierdzony):** aplikacja z innym zestawem mechanizmów i CVE
  niż petclinic/WebGoat, żeby ocenić modele `s1`/`p1`/`s2`/`j1` na
  werdyktach: skrypt przygotowania, przebiegi CHA/RTA × korzenie,
  przypisanie werdyktów do modeli binarkami sprzed kolejnych modeli,
  ręczna weryfikacja zmian, ślad runtime, benchmark, dokumentacja.
- **Zrealizowane:** `scripts/prepare-jpetstore.sh` (tag `jpetstore-6.0.2`;
  na JDK 25 bez `ct.sym`/`javac`/`jar` na PATH: `-source/-target 8`,
  `maven-war-plugin` 3.4.0 wołany wprost, `app-classes.jar` pakowany
  Pythonem), `scripts/runtime-trace-jpetstore.sh` (WAR w Jetty 9.4
  `jetty-runner`, `--add-opens java.base/java.lang` dla CGLIB Springa 4.3,
  scenariusz katalog/szukanie/koszyk/logowanie kontem demo z danych
  aplikacji), JPetStore w `scripts/graph-benchmark.sh` (wejście: sam WAR +
  `-deps-list`; bez `-deps-list` tylko 7/17 jarów rozpoznanych i 8 z 35
  advisories).
- **Wynik:** 35 advisories, 18 zmapowanych. **Żaden model** (`s1`, `p1`,
  CHA dziedziczone ciała, L1, `s2`, `u1`, `j1`) **nie zmienia werdyktu**
  (`-graph-compare` kolejnych binarek `95ae6ec`…HEAD, oba scenariusze;
  przyrost krawędzi: `p1` +98/+7, builder `3` +3 190, `s2` +255/+0, `j1`
  +811/+739). Rozbieżność scenariuszy korzeni: 8 advisories `FRAMEWORK`
  pod `-roots profile` i `NOT_REACHABLE` pod domyślnym `-roots all` —
  ścieżki z `ContextLoaderListener` (biblioteka zarejestrowana w
  `web.xml`, korzeń tylko w `profile`); ślad potwierdza wykonanie
  podatnych metod CVE-2022-22965/22970 (`CachedIntrospectionResults`) i
  CVE-2026-41848 (`AntPathMatcher`) → **3 fałszywe negatywy trybu
  domyślnego**. RTA vs CHA: 2 nowe zmiany, sprawdzone ręcznie
  (CVE-2024-38820 poprawna; CVE-2026-41851 usuwa artefakt CHA przez
  appendery log4j, RTA traci przy tym refleksyjnie tworzony
  `Log4JLogger`). Ślad: 722 relacje, 37 poza grafem (`skipped-receiver`
  17 — m.in. `RequestDispatcher` do JSP 9, `DataSource` →
  `SimpleDriverDataSource` 1 (ograniczenie `j1` zmierzone), `FilterChain`
  1; `reflection` 11 — handlery Stripes 4; `proxy` 7 — mappery MyBatis 5,
  `SqlSessionTemplate` 2 z ograniczonym zbiorem interfejsów;
  `class-init` 2). `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `RESULTS.MD` (sekcja „JPetStore”, tabela), 
  `TRACE_GAPS.MD` (sekcja „A second application”), README (skrypty,
  korzenie web.xml a `-roots all`), `docs/SUPPORT.MD` (wiersze Profiles
  and selected roots, Servlet container dispatch, JDBC objects), ten
  rejestr.
- **Ograniczenia pomiaru:** jedna aplikacja, jeden scenariusz; 17 z 35
  advisories niezmapowanych (m.in. HSQLDB CVE-2022-41853, log4j 1.2 —
  bez commitów poprawek), więc modele JDBC nie mają tu podatnej metody do
  osiągnięcia; kontener spoza wejścia.
- **Następny krok (wymaga zgody), w kolejności wartości:**
  1. korzenie `web.xml` także pod `-roots all` (lub ostrzeżenie dla WAR-a)
     — usuwa 3 zmierzone fałszywe negatywy trybu domyślnego;
  2. rozpoznawanie servletów bez API servletów w wejściu (znane nadtypy
     `HttpServlet`/`GenericServlet`);
  3. ręczne mapowania dla HSQLDB CVE-2022-41853 i log4j 1.2
     (JDBCAppender, SocketServer, JMS) — dopiero wtedy JPetStore sprawdzi
     `j1` na werdyktach;
  4. model katalogowy proxy (MyBatis `SqlSessionTemplate`/mappery,
     Spring ORM).

### FRAMEWORKS:F3 — wycinek 2 (dispatch przez proxy JDK): pomiar na śladzie, reguła stopu zadziałała — NOT STARTED (2026-11-09)

- **Zlecenie (zatwierdzone z regułą stopu):** wywołanie na interfejsie `I`
  → krawędź do `H#invoke`, gdy osiągalne miejsce `Proxy.newProxyInstance`
  ma handler `H` i `I` w zbiorze interfejsów (literał lub prosty
  przepływ); nieznany zbiór = jawna luka. Przed budową: czy zbiory
  interfejsów dla 12 relacji `proxy` ze śladu są ograniczone; jeśli dla
  większości nie — stop.
- **Metoda:** ślad z `j1` (CHA, `-roots profile`), pełna lista relacji;
  dla każdej miejsce tworzenia proxy i jego argumenty odczytane
  `dataflow.Analyze`/`ArgumentAt` (jednorazowa sonda, niezatwierdzona).
- **Wynik:** handler znany we wszystkich 6 miejscach tworzenia (5×
  `new`, `JdkDynamicAopProxy` jako `this`), zbiór interfejsów ograniczony
  w **żadnym**: pole `proxiedInterfaces` (konfiguracja AOP, 5 relacji
  repozytoriów Spring Data), tablica z parametru + `EntityManagerProxy`
  (`SharedEntityManagerCreator`, 2), wynik `ClassUtils.toClassArray`/
  `getAllInterfacesForClass` (`ExtendedEntityManagerCreator` 2,
  `ManagedEntityManagerFactoryInvocationHandler` 2,
  `DeferredQueryInvocationHandler` 1). W całym wejściu: petclinic 59
  miejsc, 24 z literalną tablicą (21 też z `new` handlerem); WebGoat
  52/21 (16) — proxy adnotacji, Spring JDBC, sesja Hibernate, Connector/J;
  żadne na zmierzonej relacji. **0 z 12** — reguła stopu, nic nie
  zbudowano. Szczegóły: `testdata/graphbench/TRACE_GAPS.MD`, „What the
  `proxy` relations are”.
- **Dokumentacja:** `TRACE_GAPS.MD`, ten rejestr; `docs/SUPPORT.MD`,
  README i FAQ bez zmian (wsparcie się nie zmieniło).
- **Możliwa alternatywa (wymaga decyzji):** model katalogowy zamiast
  wnioskowania — proxy `SharedEntityManagerCreator`/
  `ExtendedEntityManagerCreator` implementują `EntityManager` (`Query`/
  `EntityManagerFactory` dla pozostałych), repozytoria Spring Data
  (`RepositoryBeans` z F2) są opakowane w `JdkDynamicAopProxy`. Krawędź
  kończy się jednak w `invoke` handlera (dalej `Method#invoke` na
  targecie — relacje `reflection`), a ścieżka repozytoriów już dochodzi do
  `SimpleJpaRepository` przez krawędzie syntetyczne profilu; oczekiwany
  wpływ na osiągalność mały.
- **Następny krok (wymaga zgody):** trzecia aplikacja pomiarowa (ocena
  modeli `j1`/`s2`/SPI na werdyktach), model wywołań `@Bean` albo model
  katalogowy proxy powyżej.

### FRAMEWORKS:F2 — obiekty JDBC (`internal/jdbcdispatch`, `j1`) — DONE for this scope (2026-11-09)

- **Zakres:** wywołania wirtualne/interfejsowe na `java.sql` `Connection`,
  `Statement`, `PreparedStatement`, `CallableStatement`, `ResultSet`,
  metadanych, LOB-ach, `Savepoint`, `Wrapper` i `javax.sql.DataSource`
  (pomijani odbiorcy CHA). Model `service-provider:dispatch` sięga tylko
  klas w jarach sterowników, nie proxy puli (Hikari), przez które
  Hibernate i Spring JDBC faktycznie wołają.
- **Zrealizowane:** `callgraph.ObjectEdges(h, types)` — logika
  `servletdispatch.ObjectEdges` wyciągnięta do wspólnej funkcji (servlety
  bez zmian semantyki, `s2`); `jdbcdispatch.ObjectEdges` z listą typów
  JDBC; bramka twórcy (`callgraph.NewCreatorGate`, podstawa `model`) w
  punkcie stałym razem z bramką SAM i obiektów servletowych, przed modelami
  bibliotek i po nich. `java.sql.Driver` zostaje przy modelu SPI. Linia
  logu `JDBC objects (jdbc:objects j1): …`, `jdbc.objects` w snapshotach.
- **Weryfikacja:** `TestObjectEdgesGatedByCreator` (proxy puli z
  odziedziczonym ciałem; sterownik dopiero gdy jego `connect` osiągalny —
  punkt stały przez krawędź połączenia; klasa tworzona tylko przez
  nieosiągalną metodę i klasa bez twórcy — bez krawędzi),
  `TestTypesAreSkipped`. Nowe nagranie petclinic (poprzednie nie
  przetrwało restartu; ten sam skrypt i scenariusz, JDK 25.0.5, 2 946
  relacji): poza grafem 164 → 141, `skipped-receiver` 85 → 62 — wszystkie
  22 relacje `java.sql`/`javax.sql` (Hibernate/Spring JDBC → proxy Hikari
  → H2, `HikariDataSource#getConnection`, `DriverDataSource`) plus
  `DriverDataSource#getConnection` → `org.h2.Driver#connect`; żadna nowa
  luka. `-graph-compare` z `main` (`fac65dc`), oba scenariusze korzeni:
  petclinic +11 342 krawędzi, WebGoat +4 882/+4 883, 0 usuniętych, **0
  zmian werdyktów i ścieżek**; kontrolne CVE-2021-23463 (H2 `JdbcSQLXML`)
  zostaje `NOT_REACHABLE`. `scripts/graph-benchmark.sh --repeat 3`:
  `verdict-changes.tsv` bez zmian; czas WebGoat jeden po drugim bez/z
  modelem 19,9–20,4 s / 20,4 s (model 150–180 ms). `go test ./...`,
  `go vet ./...` — PASS.
- **Dokumentacja:** README (akapit „JDBC objects”), FAQ, `docs/SUPPORT.MD`
  (nowy wiersz JDBC objects), `TRACE_GAPS.MD`, `RESULTS.MD`, ten rejestr.
- **Ograniczenia:** implementacja tworzona tylko refleksyjnie nie ma
  twórcy (krawędź `HikariDataSource` na petclinic istnieje tylko dzięki
  `new HikariDataSource` w `HikariConnectionProvider`/`HikariJNDIFactory`
  Hikari, nie dzięki tworzeniu przez Boot); bramka niewrażliwa na kontekst
  (proxy tworzone przez osiągalną fabrykę jest celem każdego wywołania na
  swoim interfejsie); pule/XA (`javax.sql.PooledConnection`,
  `XAConnection`) i `RowSet` poza listą.
- **Następny krok (wymaga zgody):** model wywołań `@Bean` (18 relacji
  `reflection` z pomiaru R3; skraca ścieżki, werdyktów nie zmieni) albo
  kolejni pomijani odbiorcy ze śladu (`java.util.Map` 11, `ClassLoader` 6,
  `Runnable` 6, `java.io.Reader` 5 — rozproszone typy JDK, ryzyko
  rozrostu), albo proxy JPA (`proxy`, 12; FRAMEWORKS F3/F4).

### FRAMEWORKS:F2 — funkcje Undertow wyłączone w aplikacji Spring Boot (`internal/bootfeatures`, `u1`) — DONE for this scope (2026-11-09)

- **Zakres:** 11 fałszywych `REACHABLE` na WebGoat po `s2` (ścieżki z kodu
  aplikacji przez nadaproksymację CHA wewnątrz Undertow). Pomiar: każda
  ścieżka idzie przez rozwinięcie CHA (`AbstractFramedStreamSourceChannel#close`
  → każdy XNIO `ChannelListener`, `endExchange` → `Http2ServerConnection`,
  stream → każdy `StreamSinkConduit`); klasy docelowe są tworzone przez kod
  osiągalny w CHA, więc bramka twórcy i RTA ich nie usuwają — decyduje
  konfiguracja. Decyzja użytkownika: model funkcji Boot+Undertow.
- **Zrealizowane:** `internal/bootfeatures` — katalog funkcji (HTTP/2, TLS,
  AJP, protokół PROXY, timeout zapisu, mechanizmy uwierzytelniania Undertow,
  klient HTTP) z klasami i tym, co je włącza: właściwość w spakowanych
  `application*.properties`/`.yml` (każdy profil; parser właściwości i
  podzbioru YAML), odwołanie w klasie aplikacji (wywołanie, `new`,
  `getstatic` API funkcji), `login-config` w web.xml; przełącznik będący
  placeholderem środowiska (`${VAR:default}`) zostawia funkcję włączoną.
  Działa, gdy w wejściu są undertow-core i fabryka Undertow Spring Boot (albo
  `@SpringBootApplication`). `callgraph.CHA`/`RTA` dostały `Exclude` /
  `ExcludeRule` (klasy wyłączonych funkcji nie są celami; reguła w opisie
  skip rule); linia logu z decyzją dla każdej funkcji; `boot.features` w
  snapshotach.
- **Weryfikacja:** `TestReadConfig`, `TestDecideAndExclude`.
  `-graph-compare` z `main` (oba scenariusze korzeni): petclinic bez zmian
  (Tomcat); WebGoat −3 215 krawędzi, **4 advisories `REACHABLE` →
  `NOT_REACHABLE`** (HTTP/2 HPACK CVE-2021-3859, AJP CVE-2024-6162, PROXY
  CVE-2024-7885, uwierzytelnianie formularzem Undertow CVE-2023-1973;
  wszystkie z `MODEL_INCOMPLETE`), sprawdzone ręcznie (`RESULTS.MD`,
  „Undertow features (u1)”); z 11 artefaktów: 4 usunięte, 5 wiarygodnych
  lub zależnych od konfiguracji (TLS przez placeholder, szerokie mapowania
  trafiające w parser HTTP/1.1), 2 artefakty zostają (pong WebSocket,
  CVE-2023-5685 — to ostatnie RTA usuwa: nowa, poprawna zmiana RTA–CHA).
  `scripts/graph-benchmark.sh --repeat 3`; czas na WebGoat jeden po drugim
  w granicach rozrzutu. `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** README, FAQ, `docs/SUPPORT.MD` (nowy wiersz Embedded
  Undertow features), `RESULTS.MD`, ten rejestr.
- **Ograniczenia:** tylko Undertow i tylko spakowana konfiguracja (zmienne
  środowiskowe, linia poleceń i zewnętrzne pliki mogą włączyć funkcję);
  podzbiór YAML; wywołania bezpośrednie do klas funkcji zostają; inne
  zależne od konfiguracji części kontenerów (konektory Tomcat, Jetty) nadal
  rozwijane przez CHA.
- **Następny krok (wymaga zgody):** JDBC jako pomijany odbiorca (bramka
  twórcy dla `java.sql`, około 15 relacji na śladzie) albo model wywołań
  `@Bean`.

### FRAMEWORKS:F2 — dispatch kontenera servletów `s2`: obiekty żądania/odpowiedzi/łańcucha — DONE for this scope (2026-11-09)

- **Zakres:** krok 4 z `testdata/graphbench/TRACE_GAPS.MD` — wywołania na
  `javax.servlet` `ServletRequest`/`ServletResponse`, ich typach `Http*` i
  `*Wrapper` oraz `FilterChain` (pomijani odbiorcy CHA): 35 z 102 relacji
  `skipped-receiver` na śladzie petclinic.
- **Zrealizowane:** `servletdispatch.ObjectEdges` — dla każdego wywołania
  wirtualnego/interfejsowego na tych typach kandydat do ciała, które
  wykonuje każda konkretna implementacja z wejścia (`ResolveExact`), z
  metodami tworzącymi ją (`new`) jako twórcami; `callgraph.NewCreatorGate`
  (bramka SAM uogólniona o podstawę krawędzi; tu `model`) dodaje krawędź,
  gdy twórca jest osiągalny z korzeni — w CLI w punkcie stałym z bramką SAM,
  przed modelami bibliotek i po nich. `jakarta.servlet` pominięte (CHA już
  rozwija). Katalog Boot: `ResourceUrlEncodingFilter`
  (ThymeleafAutoConfiguration, gdy `org.webjars.WebJarAssetLocator` włącza
  łańcuch zasobów; właściwości nie są czytane). Wersja modelu `s1` → `s2`
  (snapshoty).
- **Weryfikacja:** `TestObjectEdgesGatedByCreator` (krawędź do ciała
  klasy tworzonej przez osiągalny kod; wrapper tworzony tylko przez
  nieosiągalną metodę i klasa bez twórcy — bez krawędzi),
  `TestBootCatalogThymeleafResourceFilter`. Ślad petclinic (builder `4`,
  `g2`): poza grafem 175 → 139, `skipped-receiver` 102 → 69 (33 z 35
  relacji servletowych; zostają `WsFilter` i jego
  `ServletContainerInitializer`), `no-call-at-site` 3 → 0.
  `-graph-compare` z `s1` (oba scenariusze korzeni): petclinic +5 870
  krawędzi, 0 zmian werdyktów, 32 ścieżki; WebGoat +3 447, **15 advisories
  Undertow `FRAMEWORK` → `REACHABLE`** — każde sprawdzone ręcznie
  (`RESULTS.MD`, „Servlet request/response objects”): 4 wiarygodne
  (parsowanie cookie przy `getSession`, `getQueryString` przy szerokim
  mapowaniu CVE-2024-3884/4027, `ParseState#reset`), 11 to artefakty CHA
  wewnątrz Undertow (listenery kanałów HTTP/2/AJP/WebSocket/klienta, TLS,
  uwierzytelnianie Undertow); zmiany RTA–CHA te same (4 + 2).
  `scripts/graph-benchmark.sh --repeat 3`; czas: `s1` i `s2` na WebGoat
  8,4–8,6 s jeden po drugim (model ok. 0,1 s). `go test ./...`,
  `go vet ./...` — PASS.
- **Decyzja użytkownika (2026-11-09):** wdrożyć model i opisać wszystkie 15
  zmian zamiast zawężać go do akcesorów żądania.
- **Dokumentacja:** README (sekcja „Request, response and filter-chain
  objects”), FAQ, `docs/SUPPORT.MD` (wiersz Servlet container dispatch),
  `TRACE_GAPS.MD`, `RESULTS.MD`, ten rejestr.
- **Ograniczenia:** bramka twórcy nie odróżnia filtrów osiągalnych tylko
  przez CHA (`OncePerRequestFilter#doFilter` → każde `doFilterInternal`),
  więc ich wrappery też są celami; ścieżka z kodu aplikacji do kontenera
  przechodzi dalej przez nadaproksymację CHA jego wnętrza; `WsFilter`,
  `ServletContainerInitializer`, `ServletContext#addFilter` nadal nie są
  modelowane.
- **Następny krok (wymaga zgody):** zawężenie wewnętrznego dispatchu
  Undertow (listenery kanałów, conduity) albo model wywołań `@Bean` (18
  relacji `reflection` z pomiaru R3).

### REFLECTION:R3 — wycinek 2 (refleksja między metodami): pomiar na śladzie, reguła stopu zadziałała — NOT STARTED (2026-11-09)

- **Zlecenie:** zaimplementować zmierzoną część R3 — `Method`/`Constructor`
  rozwiązany przez R1/R2 w jednej metodzie i wywołany w innej (argument,
  zwrot, pole, lambda) — z maszynerią L1 (`samcall`, `CallerIndex`,
  bramka osiągalności). Przed budową: sprawdzenie, czym są 37 relacji
  `reflection` ze śladu petclinic.
- **Metoda:** ten sam ślad co L1 (builder 4, CHA, `-roots profile`);
  pełna lista relacji przez lokalne zniesienie limitu 30 wierszy w
  szablonie raportu (niezatwierdzone); dla każdego miejsca kod bajtowy
  wywołującego — skąd pochodzi obiekt refleksji.
- **Wynik:** 18 — metody `@Bean` przez `SimpleInstantiationStrategy`
  (metoda z `getDeclaredMethods()` filtrowana nazwą z definicji beana);
  6 — `InvocationHandler` proxy JDK przekazujący swój parametr `Method`
  (EntityManager); 3 — `InvocableHandlerMethod#doInvoke` (metody
  `@RequestMapping`); 3 — klasy z definicji beanów / `spring.factories`;
  2 — domyślne klasy Hibernate (wywołanie interfejsu, `Map`); 2 — SpEL;
  2 — wyszukanie w tej samej metodzie z klasą z fasady
  (`ClassLoaderService#classForName`) lub pola (devtools `main`).
  **0 z 37** ma kształt R3, więc zmierzony zysk = 0 relacji; reguła
  stopu jak przy sondzie z 2026-11-05 (0 miejsc
  `Method#invoke` na petclinic i WebGoat). Filtry `getDeclaredMethods()`
  biorą nazwę z adnotacji/definicji beanów — zbiór nieograniczony przez
  przepływ danych; ograniczone są zbiory adnotacyjne (`@Bean`,
  mapowania żądań), które profil spring już ma jako korzenie (krawędź
  skróciłaby ścieżki, nie zmieniłaby werdyktu). Nazwy z konfiguracji (3)
  i `setAccessible` — nie zrobione.
- **Zmiany w kodzie:** brak. Tabela w `testdata/graphbench/TRACE_GAPS.MD`
  (sekcja "What the `reflection` relations are").
- **Następny krok (wymaga zgody):** jeśli w ogóle — model adnotacyjny
  (wywołanie metod `@Bean`, przekazanie `Method` w handlerach proxy JDK —
  styka się z FRAMEWORKS F3); sam R3 dopiero z aplikacją kontrolną, która
  przekazuje stałe `Method` między metodami.

### DYNAMIC_LINKAGE:L1 — materializacja wywołań SAM — DONE for this scope (2026-11-09)

- **Archiwum:** PR #30 (`a19eab5`, merge `2aa71d9`); tekst fazy przeniesiony do
  [archive/2026-11-09/dynamic_linkage/DYNAMIC_LINKAGE_L1.md](../archive/2026-11-09/dynamic_linkage/DYNAMIC_LINKAGE_L1.md).

- **Zakres zlecenia:** otwarta połowa L1 — lambda/method reference tworzona
  w jednej metodzie i uruchamiana przez metodę interfejsu funkcyjnego w
  innej (przekazana jako argument, zwrócona, zapisana w polu) ma dostać
  krawędź od miejsca wywołania SAM do implementacji, z wartością śledzoną
  przez DATA_FLOW, bez zastępczego „każda lambda tego interfejsu” bez
  ujawnionej luki; osobna podstawa krawędzi; zmierzone na śladzie
  petclinic (`TRACE_GAPS.MD`: `sam-call` 58 z 203). Gałąź od
  `cha-inherited` (PR #29, na #28).
- **Zrealizowane:**
  1. `internal/classfile`: `Call` dla rozwiązanego `invokedynamic` niesie
     `SamInterface`/`SamMethod` (z `NameAndType`). `newReader` nie
     alokuje już bufora 32 KiB na każdy atrybut w pamięci (parsowanie było
     ograniczone alokacjami — patrz „Czas”).
  2. `internal/dataflow`: `CallerIndex.SetOptions(CallerOptions{MaxDepth,
     MaxCallers, FollowCaptures, VirtualCallers})` — parametr
     implementacji lambdy rozwiązywany przez wartości przechwycone w
     miejscach `invokedynamic` (z przesunięciem o odbiorcę dla uchwytów
     instancyjnych; parametr spoza przechwyconych to argument wywołania
     SAM — nieznany z tym powodem); wywołujący metody wirtualnej jako suma
     miejsc nazywających ją i deklaracje, które nadpisuje w wejściu
     (nadtyp spoza wejścia blokuje, chyba że nie może zadeklarować
     metody: interfejs znacznikowy albo typ JDK i deskryptor z typem spoza
     JDK); `Facts`, `Prefetch` (równoległa analiza, wynik niezależny od
     kolejności), `MethodFacts.ReturnValues`/`FieldWrites` (każdy zapis
     pola z wartością, dowolny odbiorca); pamięć podręczna 512
     sparsowanych klas.
  3. Nowy `internal/samcall` (`Version = "l1-1"`): `Find` — wywołania
     metody SAM na interfejsie implementowanym przez jakąś lambdę wejścia
     (lub jego nadtypie) oraz wywołania metody spoza wejścia z argumentem
     typu takiego interfejsu („escape”); `Resolver` — odbiorca/argument
     przez fakty metody, bezpośrednich wywołujących (`CallerOptions.Direct`,
     do 64 na metodę), pola (wszystkie zapisy w klasie i klasach
     zagnieżdżonych; pole ani prywatne, ani finalne → luka) i zwroty
     metod static/private; każde ogniwo (parametr, pole, zwrot) liczone
     raz i zapamiętane, cykle cięte (wynik przecięty cyklem ważny tylko
     do końca bieżącego miejsca); parametr/pole o typie, który nie może
     trzymać lambdy (nie jest interfejsem funkcyjnym wejścia, jego
     nadtypem ani `Object`), nie jest śledzone; budżet 20 000 ogniw na
     miejsce; luki z kategorią (`parameter`, `sam-argument`, `field`,
     `call-result`, `instance-call`, `unknown-value`, `depth`,
     `bootstrap`, `budget`); `Edges`, `Summarize`.
  4. `internal/callgraph`: `BasisSAM = "sam"`, `SAMGate` — kandydaci
     dodawani dopiero, gdy któraś metoda tworząca wartość jest osiągalna
     z korzeni (punkt stały; wywoływane przed i po modelach bibliotek);
     referencja metody instancyjnej rozwijana jak wywołanie wirtualne
     (zapamiętane per właściciel/nazwa/deskryptor); para już w grafie
     (krawędź `invokedynamic` twórcy) nie jest dublowana.
     `BuilderVersion` 3 → 4.
  5. CLI: dwie linie logu (miejsca i luki; krawędzie dodane / pokryte /
     pominięte przez bramkę), nagłówek raportu (`SAM calls
     (DYNAMIC_LINKAGE L1)`, liczba krawędzi SAM w wierszu Call graph),
     wersja `samcall.rules` w snapshotach, kolumny `sam_edges` /
     `sam_edges_gated` w `scripts/graph-benchmark.sh`.
- **Weryfikacja:** fixture z prawdziwym javac i JVM `testdata/samcall`
  (`scripts/samcall-oracle.sh`: każde ciało lambdy/cel referencji
  zapisuje najbliższą ramkę fixture pod sobą, StackWalker bez ukrytych
  ramek) — `TestSAMFixture`: wszystkie 8 zmierzonych relacji
  (parametr, zwrot, pole ustawiane setterem, pole finalne z referencją
  związaną do nadpisanej metody — celem jest nadpisanie, referencja
  konstruktora, `List#forEach`, lambda przechwycona przez lambdę, helper
  wołany przez interfejs) to krawędzie `sam` pod zmierzonym bci;
  nieużywana lambda i lambda z nigdy niewywoływanej metody nie mają
  żadnej; `TestSAMGapsDisclosed`, `TestMayDeclareOutside`.
  `TestRuntimeTrace`: luka `sam-call` (`Gaps#drive` → `lambda$sam$0`)
  zniknęła, zostaje `reflection` 2. petclinic, nowe nagranie tym samym
  skryptem (stare nie zostało zachowane), CHA `-roots profile`: poza
  grafem 204 → 175 z 2 702, `sam-call` 50 → 22 (12 to nie lambdy —
  kolekcje JDK wołające `equals`/`hashCode`, spliteratory, `Comparator`;
  7 realnych z utraconą wartością — element tablicy, zmienna przez
  handler wyjątku, wynik wywołania instancyjnego, argument SAM, nadtyp
  JDK; 3 pominięte przez bramkę osiągalności). Pierwsza wersja bez bramki
  zmieniła CVE-2021-23463 z `NOT_REACHABLE` na `FRAMEWORK` przez
  kontekstowo-niewrażliwą sumę (`JdbcTemplate#execute(ConnectionCallback)`
  „uruchamiał” lambdę `AbstractJdbcInsert`, której twórcy nic osiągalnego
  nie woła) — bramka to usunęła. `scripts/graph-benchmark.sh --repeat 3`
  (2026-11-09, `testdata/graphbench/RESULTS.MD`): żadnej zmiany werdyktu w
  8 konfiguracjach, zmiany RTA–CHA te same (4 + 2); `-graph-compare` z
  grafem buildera 3: petclinic +1 626 `sam` i +171 `model`, 2 krótsze
  ścieżki (`SpringApplicationRunListeners#doWithListeners` →
  `lambda$contextLoaded$4` przez `List#forEach`); WebGoat +1 557–1 558
  `sam` i +77 `model`, 1 ścieżka (`Lazy#getNullable` →
  `JpaQueryMethod#lambda$new$0`); usuniętych krawędzi 0. Czas analizy
  niższy niż z builderem 3: petclinic 14,6 → 10,4–11,1 s, WebGoat
  23,2–24,0 → 16,2–16,3 s, szczytowa pamięć bez zmian (1,34–1,36 /
  1,71–1,81 GB).
  `go test ./...`, `go vet ./...`, `go test -race ./internal/samcall
  ./internal/dataflow` — PASS.
- **Czas:** pierwsza wersja: krok SAM 9,2 s na petclinic (ponowne
  czytanie i parsowanie klasy dla każdej metody); po pamięci podręcznej
  klas, równoległym `Prefetch` i usunięciu bufora 32 KiB na atrybut w
  parserze — ok. 2,8 s; cała analiza jest krótsza niż przed zmianą
  (poprawka parsera przyspiesza też budowę hierarchii) — liczby w
  `testdata/graphbench/RESULTS.MD`. Pamięć:
  pierwszy pomiar dał +300 MB szczytowego RSS (1,37 → 1,66 GB);
  przyczyną była kopia całej `hierarchy.Method` w każdym miejscu
  wywołania `dataflow.CallerIndex` (ok. 600 tys.) — teraz wskaźnik, a
  fakty rozwiązywane partiami po 1 024 metody z limitem 4 096 w
  pamięci. Poprawka indeksu
  zmniejsza też pamięć SERVICE_LOADING, który używa tego samego indeksu.
  WebGoat: pierwszy pomiar benchmarku dał krok SAM 12,5 s (rekurencyjne
  rozwijanie łańcuchów parametrów w `ResolveParameter`); po przejściu na
  bezpośrednich wywołujących z zapamiętaniem ogniw w `samcall` kolejna
  wersja zawisła na interpreterze JRuby (`Iterable#iterator`, łańcuchy
  parametrów `IRubyObject` > 60 poziomów, silnie rekurencyjne) —
  przycinanie po typie zadeklarowanym i budżet to usunęły. Końcowo krok
  SAM: petclinic 1,9 s, WebGoat 2,4 s, te same krawędzie co przed
  optymalizacjami.
- **Dokumentacja:** README (sekcja 4 „SAM calls”, Known limitations),
  FAQ (invokedynamic), `docs/SUPPORT.MD` (wiersz Lambdas and method
  references), `testdata/graphbench/TRACE_GAPS.MD` i `RESULTS.MD`, ten
  rejestr.
- **Ograniczenia:** wartość nie jest śledzona przez wynik wywołania
  instancyjnego, element tablicy, kolekcję ani zmienną żywą przez handler
  wyjątku; analiza niewrażliwa na kontekst (suma po wywołujących, ze
  złagodzeniem przez bramkę twórcy); wartość przekazana do metody JDK
  traktowana jako przez nią uruchamiana, także gdy JDK ją tylko
  przechowuje (`Map#put`); krawędzie `sam` nie są zawężane przez RTA.
  Klasyfikator luk śladu liczy jako `sam-call` także wywołania nie-lambd
  przez ramki JDK.
- **Następny krok (wymaga zgody):** refleksja między metodami (REFLECTION
  R3, 37 relacji na nowym śladzie) albo zawężenie klasyfikatora
  `sam-call` i śledzenie wyniku wywołań instancyjnych.

### FRAMEWORKS:F2 — siódmy wycinek: dispatch kontenera servletów (model B3 `servlet-container:dispatch`) — DONE for this scope (2026-11-08)

- **Zakres zlecenia (2026-11-08, sesja nadrzędna po GRAPH_PRECISION:P4):**
  zamodelować dispatch kontenera do servletów i filtrów, którego graf nie
  rozwija (`DefaultSkip` pomija odbiorców `javax.`/`jakarta.`):
  ograniczony wyjątek dla `javax.servlet.Servlet#service`,
  `javax.servlet.Filter#doFilter` i odpowiedników `jakarta.servlet`,
  wyłącznie do servletów/filtrów rejestrowanych przez aplikację, z
  zapisaną podstawą krawędzi; fixture z wyrocznią JVM; ponowny
  `scripts/graph-benchmark.sh --repeat 3` z wyjaśnieniem CVE-2024-38809.
  Gałąź `servlet-dispatch` od `graph-precision-p4` (`a64edf7`, PR P4 jeszcze
  nie scalony).
- **Wybór miejsca:** model rejestru B3 (nie wyjątek w `internal/callgraph`):
  cele zależą od faktów frameworka (web.xml, adnotacje, beany, autokonfiguracja
  Boot), których `callgraph` nie zna, a koordynator `internal/analysis`
  ocenia tylko osiągalne miejsca wywołań i daje krawędzie z podstawą
  `model` i PC miejsca. Reguła skip pozostaje bez zmian i opisana jako taka.
- **Zrealizowane:**
  1. `internal/servletdispatch` (`ModelID = "servlet-container:dispatch"`,
     `Version = "s1"`): sygnatury `Servlet`/`GenericServlet`/`HttpServlet#
     service(ServletRequest, ServletResponse)`, `Filter#doFilter` oraz
     szablon `HttpServlet` (`service(HttpServletRequest, …)`, `doGet`/
     `doPost`/`doPut`/`doDelete`/`doHead`/`doOptions`/`doTrace`/`doPatch`,
     `getLastModified`) dla `javax` i `jakarta`. Cel: metoda wybrana przez
     JVM (`ResolveExact`) na każdej zarejestrowanej klasie będącej podtypem
     odbiorcy wywołania; tylko `invokevirtual`/`invokeinterface` (nie
     `super.`), szablon tylko gdy wołającym jest sam `HttpServlet`.
     Rejestracje (`Find`): web.xml, klasy aplikacji z `@WebServlet`/
     `@WebFilter`, a w aplikacji Spring Boot (`@SpringBootApplication`/
     `@EnableAutoConfiguration` na klasie aplikacji) beany będące
     servletem/filtrem (stereotyp, konkretny typ zwracany `@Bean`) oraz
     katalog autokonfiguracji sprawdzony w źródłach Boot: `DispatcherServlet`,
     `OrderedCharacterEncodingFilter`, `OrderedFormContentFilter`,
     `OrderedRequestContextFilter`, `DelegatingFilterProxy` + `FilterChainProxy`
     — każdy tylko przy obecnej autokonfiguracji i jej klasach
     `@ConditionalOnClass` oraz bez `exclude`/`excludeName`. `@Bean`
     zwracający `ServletRegistrationBean`/`FilterRegistrationBean`/goły
     `Servlet`/`Filter` → niekompletny fakt modelu z przyczyną.
  2. `models.Router` — wiele ewaluatorów w jednym przebiegu koordynatora
     (ServiceLoader domyślnie, servlety po ID); budżet iteracji +3
     (`MaxChain`).
  3. CLI: wiersz logu `servlet dispatch (s1): N registered servlets/filters
     (klasa [rodzaj, podstawa]; …)`; zarejestrowane klasy jako typy
     `container` w katalogu P1 (`p1-2`), więc żywe dla RTA; snapshoty
     zapisują `servlet.dispatch`; `callgraph.BuilderVersion` `2`.
  4. Fixture `testdata/servletdispatch` + `scripts/servletdispatch-oracle.sh`:
     WAR z prawdziwym API `javax.servlet` (klasy `jakarta.servlet-api-4.0.4`
     bez metadanych Maven, żeby test offline nie pytał OSV), własny kontener
     tworzący klasy z web.xml po nazwie i wołający `Filter#doFilter`/
     `Servlet#service` jak `ApplicationFilterChain`; `HttpServlet` przez
     szablon, `GenericServlet`, filtr; niezarejestrowane `UnusedServlet`/
     `UnusedFilter`. Wyrocznia JVM (OpenJDK 25): 5 ciał. W
     `TestGraphBenchmarkFixtures`: 5/5 osiągnięte z `main` w CHA i RTA, przy
     `-roots profile` i `all`, niezarejestrowane nigdy (asercja); analizator
     `a64edf7`: 0/5 w obu trybach.
  5. `scripts/graph-benchmark.sh`: kolumny `servlet_registrations`,
     `model_edges`, wersja `servlet_dispatch`.
- **Wyniki (benchmark `--repeat 3`, porównanie per ustalenie z `a64edf7`):**
  petclinic — 4 rejestracje (autokonfiguracja Boot; bez Spring Security),
  12 krawędzi modelu (`ApplicationFilterChain#internalDoFilter` →
  `HttpServlet#service`, → `OncePerRequestFilter#doFilter`; `HttpServlet` →
  `FrameworkServlet#service`/`doGet`/…; plus dwa miejsca Spring MVC
  `SimpleServletHandlerAdapter`/`ServletWrappingController` — ograniczona
  nadaproksymacja), `DispatcherServlet#doDispatch` i
  `ResourceHttpRequestHandler#handleRequest` osiągalne z `main` w obu
  trybach. Werdykty: NOT_REACHABLE → FRAMEWORK we wszystkich scenariuszach
  CVE-2024-38819 (`isInvalidPath`, zasoby statyczne — poprawne),
  CVE-2026-41844 (`getViewName` przez `applyDefaultViewName` — wiarygodne),
  CVE-2023-20883 (`WelcomePageHandlerMapping` — poprawne); dodatkowo przy
  RTA `-roots profile` CVE-2024-22259 (CORS przez `CorsInterceptor` —
  wiarygodne, było nierozstrzygnięte) i **CVE-2024-38809** (było fałszywie
  negatywne; rzeczywista ścieżka `ResourceHttpRequestHandler#handleRequest`
  → `checkNotModified(J)` → `checkNotModified(String, J)` →
  `validateIfNoneMatch` jest w grafie, raport pokazuje najkrótszą, reaktywną).
  Zmiany RTA vs CHA na petclinic: 6 → 4 (Tomcat, poprawne). Żaden REACHABLE
  nie zmieniony. Czas/pamięć w szumie (petclinic CHA 15,2–15,8 s,
  1381–1411 MB). WebGoat — 0 rejestracji, graf i werdykty jak w P4: wejście
  `webgoat-container` nie zawiera `StartWebGoat` (`@SpringBootApplication`
  w `webgoat-server`), więc fałszywie negatywne CVE-2024-38809 przy RTA
  pozostaje (wada wejścia, nie modelu).
- **Weryfikacja:** `scripts/servletdispatch-oracle.sh` (OpenJDK 25),
  `TestGraphBenchmarkFixtures`, testy `internal/servletdispatch` (web.xml,
  `@WebServlet`, beany tylko w Boot, katalog Boot z `@ConditionalOnClass`
  i `exclude`, `super.`, obcy wołający, nieznana rejestracja → niekompletny
  fakt), `TestRouterSendsEachModelToItsEvaluator`, `scripts/graph-benchmark.sh
  --repeat 3` (zielony), `bin/ariadne` na petclinic/WebGoat × profile/all ×
  cha/rta przeciw `a64edf7`, `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** README (wyjątek od skip, P4, skrypt, układ), FAQ,
  `docs/SUPPORT.MD` (nowy wiersz, wiersz RTA), `docs/RESULTS_CONTRACT.MD`
  (katalog `p1-2`), `testdata/graphbench/RESULTS.MD`, plany FRAMEWORKS i
  GRAPH_PRECISION, ten rejestr.
- **Ograniczenia:** tylko rejestracje widoczne w wejściu; niemodelowane
  `init`/`destroy`, listenery, `ServletContainerInitializer` i rejestracje
  bibliotek (filtr WebSocket Tomcata, filtr metryk Actuatora), wewnętrzne
  filtry Spring Security za `FilterChainProxy`, `ServletContext#addServlet`;
  brakująca rejestracja nie jest luką ustalenia (`MODEL_INCOMPLETE` jej nie
  nazywa); każde osiągalne `Servlet#service` dostaje cały zarejestrowany
  zbiór (także adaptery Spring MVC); katalog Boot nie ocenia warunków
  właściwości ani `@ConditionalOnMissingBean` (przyjęte wartości domyślne).
- **Następny krok (wymaga zgody):** scalić P4, potem tę gałąź; ewentualnie
  dołączyć `webgoat-server` do przygotowania WebGoat (żeby benchmark
  odzwierciedlał rzeczywisty launcher), lifecycle `init`, wewnętrzny łańcuch
  Spring Security, luka ustalenia przy niezarejestrowanym dispatchu.

### JVM_GRAPH:G2 (korekta) — CHA dla implementacji dziedziczonych spoza typu wywołania — DONE (2026-11-08)

- **Zakres zatwierdzony (2026-11-08):** naprawa defektu CHA wskazanego
  przez klasyfikację luk względem śladu (`inherited-impl`, 17 relacji na
  petclinic). Gałąź `cha-inherited` od `trace-gaps` (PR #28).
- **Zrealizowane:** `callgraph.CHA.DispatchTargets` — poza podtypami
  deklarującymi metodę, dla każdego konkretnego podtypu bez własnej
  deklaracji cel = ciało wybrane przez `ResolveExact` (także w nadklasie
  spoza hierarchii typu wywołania i domyślna metoda interfejsu;
  abstrakcyjne pominięte, reguła skip jak dla podtypów). RTA korzysta z
  tego zbioru kandydatów bez zmian. `BuilderVersion` 2 → 3.
- **Weryfikacja:** `TestCHAInheritedBody` (fixture `testdata/runtime`:
  `LoudGreeter extends BaseGreeter implements Greeter` — krawędź CHA do
  `BaseGreeter#greet`, którą JVM wykonał w nagranym śladzie);
  `TestKnownTargets` (wyrocznia JVM P0) bez nowych nadmiarowych celów;
  `TestRuntimeTrace` — w fixture zostają `reflection` 2 i `sam-call` 1.
  petclinic, ten sam ślad: CHA 221 → 203 relacji poza grafem
  (`inherited-impl` 17 → 0), RTA `-roots profile` 326 → 234
  (`rta-pruned` 105 → 31). `-graph-compare` z grafem buildera 2 (oba tryby
  korzeni): petclinic +11 713 krawędzi, 0 zmian werdyktów, 43–44 ścieżki
  krótsze (np. `OwnerController#processFindForm` →
  `AbstractErrors#rejectValue`); WebGoat +16 439, 0 zmian werdyktów, 1–3
  ścieżki. Benchmark `scripts/graph-benchmark.sh --repeat 3` powtórzony
  (2026-11-09, `testdata/graphbench/RESULTS.MD`): bez zmian werdyktów;
  pierwszy przebieg wykazał 2× dłuższą analizę WebGoat (47 s zamiast
  28 s — `ResolveExact` dla każdego miejsca wywołania), naprawione
  zapamiętaniem celów dispatchu per właściciel/nazwa/deskryptor w
  `callgraph.BuildWith` (WebGoat 23–25 s, petclinic 14,5–15 s).
  `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz Direct calls, CHA and
  inheritance), README (sekcja 4), FAQ (CHA),
  `testdata/graphbench/TRACE_GAPS.MD` i `RESULTS.MD`, ten rejestr.
- **Następny krok (wymaga zgody):** materializacja wywołań SAM
  (DYNAMIC_LINKAGE L1, 58 relacji) albo refleksja między metodami (R3).

### RUNTIME_NATIVE:N1 — ostrzejszy klasyfikator `sam-call` (przyczyny `g2`) — DONE (2026-11-09)

- **Zakres:** poprawka pomiaru, nie grafu. `g1` liczył jako `sam-call` każde
  wywołanie wirtualne, którego zmierzony cel nazywa się inaczej niż wołana
  metoda — przez ramki JDK także kolekcje wołające `equals`/`hashCode`,
  spliteratory, `Comparator` (12 z 22 na śladzie petclinic po L1).
- **Zrealizowane:** `gapClassifier` (`cmd/ariadne/runtime.go`): `sam-call`
  tylko, gdy cel implementuje lambdę/referencję metody tworzoną w wejściu
  (`invokedynamic` z `SamInterface`/`SamMethod`; także nadpisanie celu
  referencji), której interfejs to typ wywołania (lub jego podtyp), a metoda
  to wołana metoda. `GapCauseVersion` `g1` → `g2` (wersja w snapshotach).
- **Weryfikacja:** `TestGapClassifier` (nowy przypadek: cel, którego żadna
  lambda nie implementuje, w miejscu SAM → nie `sam-call`); ten sam ślad
  petclinic, builder `4`: `sam-call` 22 → 12, `skipped-receiver` 92 → 102,
  łącznie 175 bez zmian (pozostałe 12 sprawdzone pojedynczo:
  `TRACE_GAPS.MD`, „A sharper `sam-call`”). `go test ./...`, `go vet` — PASS.
- **Dokumentacja:** `docs/RUNTIME_TRACE.MD`, `docs/SUPPORT.MD`,
  `testdata/graphbench/TRACE_GAPS.MD`, ten rejestr.
- **Ograniczenia:** gdy cel jest też celem referencji metody tworzonej w
  wejściu, klasyfikator nie odróżni wywołania przez lambdę od zwykłego
  obiektu tej klasy (2 z 12).

### RUNTIME_NATIVE:N1 — przyczyny luk grafu względem śladu — DONE (2026-11-08)

- **Zakres zatwierdzony (2026-11-08):** pogrupowanie relacji zmierzonych w
  przebiegu, których graf statyczny nie ma, według przyczyny ustalonej z
  bytecode wywołującego; pomiar CHA i RTA na petclinic; ranking dalszych
  modeli według pomiaru. Gałąź `trace-gaps` od `main` (po PR #26, #27).
- **Zrealizowane:**
  1. `internal/runtimeevents`: `Inputs.Classify`, przyczyna i szczegół na
     każdej relacji `not_in_graph`, `Result.GapCauses` (liczby per
     przyczyna, 10 najczęstszych szczegółów), `GapCauseVersion = "g1"`.
  2. `cmd/ariadne/runtime.go` `gapClassifier`: instrukcja wywołania pod
     zmierzonym `bci` w metodzie wywołującej → `class-init`,
     `invokedynamic`, `reflection`, `method-handle`, `proxy` (cel
     `InvocationHandler#invoke`), `async` (Thread/Executor przez ramki
     runtime), `sam-call` (ciało lambdy albo cel referencji o innej nazwie
     niż metoda interfejsu), `skipped-receiver`, `inherited-impl` (ciało
     wybrane przez JVM dla podtypu pochodzi z nadklasy spoza typu
     wywołania), `rta-pruned` / `dispatch-missing`, `no-call-at-site`.
  3. Raport: tabela przyczyn z typami odbiorców, kolumna przyczyny w
     tabeli relacji; v2 `runtime.gap_cause_version`,
     `runtime.calls_not_in_graph_by_cause`; snapshot: wersja
     `runtime.gap_causes`.
  4. Fixture `testdata/runtime` rozszerzony (`Gaps.java`: dziedziczone
     ciało `BaseGreeter#greet` za `Greeter`, lambda tworzona w `sam()` i
     wywoływana w `drive()`), ślad nagrany ponownie.
  5. `testdata/graphbench/TRACE_GAPS.MD`: pomiar na petclinic i ranking.
- **Weryfikacja:** `TestGapClassifier` (7 przypadków na prawdziwym bytecode
  `rtm.jar`: dispatch-missing/rta-pruned zależnie od trybu,
  inherited-impl, sam-call, reflection, class-init, brak indeksu),
  `TestRuntimeTrace` (fixture: reflection 2, inherited-impl 1, sam-call 1;
  tabela w raporcie). Pierwsza wersja zostawiała 41 relacji jako
  „pośrednie”; przejrzane przykłady okazały się dziedziczonymi
  implementacjami i wywołaniami SAM — dodane jako osobne przyczyny, „pośrednie”
  spadło do 0. Błąd szablonu (`$` wewnątrz `with`) wykryty przy pierwszym
  renderowaniu, poprawiony. petclinic (ślad z 505 metodami śledzonymi):
  CHA 221 z 2 867 relacji poza grafem (skipped-receiver 87, sam-call 58,
  reflection 33, inherited-impl 17, proxy 12, class-init 8, async 3,
  no-call-at-site 3), identycznie dla `-roots all` i `profile`; RTA
  `-roots all` 222 (1 rta-pruned), RTA `-roots profile` 326 (105
  rta-pruned, 35 z nich handlery AttoParser/Thymeleaf). `go test ./...`,
  `go vet ./...` — PASS.
- **Wniosek:** największe mierzalne luki to wywołania SAM (DYNAMIC_LINKAGE
  L1), refleksja między metodami (REFLECTION R3) i dziedziczone
  implementacje — te ostatnie to defekt CHA (`DispatchTargets` bierze tylko
  podtypy deklarujące metodę), tani do naprawy; RTA jako domyślny tryb
  dodatkowo nieuzasadniony (105 pominiętych relacji przy `-roots profile`);
  points-to (P3) nieuzasadnione.
- **Dokumentacja:** `docs/RUNTIME_TRACE.MD`, `docs/RESULTS_CONTRACT.MD`,
  `docs/SUPPORT.MD`, README, `testdata/graphbench/TRACE_GAPS.MD`, ten
  rejestr.
- **Ograniczenia:** jeden przebieg jednej aplikacji; przyczyna z
  instrukcji pod zmierzonym indeksem (ramki inline mogą go przesunąć —
  `no-call-at-site`); mierzy pominięte cele, nie nadaproksymację.
- **Następny krok:** naprawa CHA dla dziedziczonych implementacji —
  zrobione (wpis „JVM_GRAPH:G2 (korekta)”). Pozostaje (wymaga zgody):
  materializacja wywołań SAM (L1).

### SERVICE_LOADING (po zamknięciu obszaru): dispatch SPI na pomijanych odbiorcach — DONE (2026-11-08)

- **Zakres zatwierdzony (2026-11-08):** model B3 dla wywołań na
  interfejsach SPI JDK/javax, których CHA nie rozwija, ograniczony do
  zarejestrowanych providerów (`META-INF/services`) i ich „produktów”;
  wskazany przez ślad runtime (CVE-2025-35036 na petclinic). Obszar
  SERVICE_LOADING jest zamknięty (S0–S4), to jego dodatkowy wycinek. Gałąź
  `spi-dispatch` od `main` (po PR #25).
- **Zrealizowane:**
  1. `internal/spidispatch` (`ModelID = "service-provider:dispatch"`,
     `Version = "p1"`): `Find` — usługi pomijane przez `DefaultSkip` z
     zarejestrowanymi providerami obecnymi w wejściu (klasa konkretna,
     podtyp usługi), rodzina API (pakiet usługi bez `.spi`), produkty
     (domknięcie po typach zwracanych metod usługi — z deklaracji w
     wejściu i z wywołań w bytecode — w obrębie rodziny i pomijanych
     pakietów), sygnatury z faktycznych wywołań w wejściu; `Evaluator` —
     usługa → zarejestrowani providerzy, produkt → implementacje z jarów
     providerów; cel = metoda wybrana przez JVM (`ResolveExact`, bez
     abstrakcyjnych).
  2. CLI: rejestracja w koordynatorze obok modelu servletów, log
     `service-provider dispatch (…)`, budżet iteracji +2, wersja
     `spi.dispatch` w snapshotach.
  3. `testdata/spidispatch` + `scripts/spidispatch-oracle.sh` (cztery
     jary: API `javax.ariadnefx` z usługą, produktem i bootstrapem przez
     `ServiceLoader`; provider z zarejestrowaną usługą i nietworzonym
     `UnusedCodec`; inny jar z `RogueCodec` i niezarejestrowanym
     `RogueProvider`; aplikacja).
- **Weryfikacja:** `TestSPIDispatchOracle` (pełne CLI: oba ciała
  uruchomione przez JVM osiągalne z `main`; krawędzie modelu do
  `Base64Provider#create`, `Base64Codec#encode` i — zgodnie z projektem —
  `UnusedCodec#encode`; żadnej do `fx.other.*`); mutacja (bez produktów)
  → test zawodzi. Pierwsza wersja z całą rodziną API (bez domknięcia po
  typach zwracanych) dawała fałszywe FRAMEWORK dla CVE-2021-22118
  (`ServletContainerInitializer` wciągał `javax.servlet` do servletów
  spring-web) — poprawione przed commitem. Real data (`-graph-compare` z
  binarką `main`, ten sam cache): petclinic 8 usług / 1 096 sygnatur, 5 085
  krawędzi dodanych; zmiany: CVE-2025-35036 NOT_REACHABLE → FRAMEWORK
  (ślad: `HibernateValidator#buildValidatorFactory` 3 wejścia, metody
  advisory 26 wejść), CVE-2021-46877, CVE-2022-45868, CVE-2023-6378
  FRAMEWORK → REACHABLE przez `Eh107CacheManager#createCache` (ślad: 1
  wejście — pierwszy krok realny; dalej istniejąca nadaproksymacja
  magazynu dyskowego Ehcache i deserializacji — 0 wejść); WebGoat 11
  usług / 1 105 sygnatur, 1 769 krawędzi, jedyna zmiana CVE-2025-35036.
  Relacje ze śladu petclinic poza grafem: 227 → 221. Benchmark
  `scripts/graph-benchmark.sh --repeat 3` powtórzony (`testdata/graphbench`):
  zmiany RTA vs CHA te same (4 + 2); koszt modelu 0,6–2 s na WebGoat, w
  granicach szumu. `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz JDK/javax service-provider
  dispatch), README (sekcja 4, skrypt wyroczni),
  `testdata/graphbench/RESULTS.MD`, ten rejestr.
- **Ograniczenia:** ograniczenie przez rejestrację i jar, nie przez
  alokację (nietworzona implementacja w jarze providera jest celem);
  implementacje spoza jarów providerów, providerzy tylko z `provides`
  modułu i to, którego providera dostaje dane miejsce wywołania — nie
  modelowane; usługi spoza pomijanych pakietów zostają CHA.
- **Następny krok (wymaga zgody):** benchmark grafu względem śladu
  (pominięte cele per przyczyna) albo launcher Spring Boot dla WebGoat.

### RUNTIME_NATIVE:N1 — filtr dokładnego śledzenia z mapowań CVE — DONE (2026-11-08)

- **Zakres zatwierdzony (2026-11-08):** generowanie z mapowań CVE ustawień
  JFR śledzących na wejściu podatne metody ustaleń, odczyt aktywnego filtra
  z nagrania, oznaczenie ustaleń „śledzone, bez wejścia w tym przebiegu”.
  Gałąź `runtime-filter` od `main` (po PR #24); w tym samym commicie
  archiwizacja N0–N1.
- **Zrealizowane:**
  1. `-runtime-filter-out FILE.jfc` i `-runtime-filter-status` (domyślnie
     wszystkie werdykty poza NOT_APPLICABLE; nieznany werdykt → błąd):
     `runtimeFilter` — filtry `Class::method` (konstruktory `<init>`,
     escapowane w XML) dla zmapowanych metod, których klasa jest w
     wejściu; plik `.jfc` z `jdk.MethodTrace` (`enabled`, `stackTrace`,
     `threshold 0 ms`, `filter`), podawany obok `profile`
     (`settings=profile,settings=FILE`); pusty wybór → plik wyłączony i
     ostrzeżenie.
  2. `scripts/JfrToTrace.java`: filtr `jdk.MethodTrace` aktywny w
     nagraniu (`jdk.ActiveSetting`) trafia do `scope.method_filters`;
     `--method-filter` tylko dopisuje.
  3. `runtimeTraced`/`tracedNoEntry`: ustalenia, których zmapowana metoda
     była objęta filtrem — v2 `findings[].runtime_traced`; raport: lista
     „Traced on entry, not entered in this run” z zastrzeżeniem, że to nie
     dowód nieosiągalności.
  4. `scripts/runtime-trace-petclinic.sh`: najpierw analiza z
     `-runtime-filter-out`, potem przebieg z wygenerowanym plikiem.
- **Weryfikacja:** sprawdzono najpierw na JDK 25, że JFR przyjmuje drugi
  plik ustawień obok `profile`, że filtr obsługuje `<init>` i że aktywny
  filtr jest zapisany w nagraniu (`jdk.ActiveSetting`). `TestRuntimeFilter`
  (wybór per werdykt, klasa spoza wejścia pominięta, escapowanie,
  pokrycie i „bez wejścia”), `TestRuntimeFilterFlag` (pusty wybór →
  wyłączony plik i ostrzeżenie w raporcie; nieznany werdykt odrzucony).
  petclinic: filtr 505 metod z 91 advisories (79 zmapowanych metod spoza
  wejścia), przebieg z ok. 193 000 dokładnych wejść, 39 936 uciętych
  stosów (głębokie stosy wejść); 70 śledzonych metod wykonanych — 26
  advisories z dokładnym wejściem: wszystkie 3 statycznie REACHABLE, 23 z
  74 FRAMEWORK, CVE-2025-35036 (NOT_REACHABLE, 26 wejść); 65 advisories
  śledzonych bez wejścia w tym scenariuszu; werdykty bez zmian.
  `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `docs/RUNTIME_TRACE.MD`, `docs/RESULTS_CONTRACT.MD`
  (`findings[].runtime_traced`), `docs/SUPPORT.MD`, README (sekcja, flagi,
  skrypt), plan RUNTIME_NATIVE (archiwizacja N0–N1), ten rejestr.
- **Ograniczenia:** śledzenie wielu gorących metod spowalnia nagrywany
  przebieg (`-runtime-filter-status` zawęża); filtr JFR nie rozróżnia
  overloadów (śledzi każdy o tej nazwie); metody zmapowane na klasy spoza
  wejścia pominięte; „bez wejścia” dotyczy tylko tego scenariusza.
- **Następny krok:** model wywołań SPI `javax.*` przez zarejestrowanych
  providerów — zrobione (wpis „SERVICE_LOADING (po zamknięciu obszaru):
  dispatch SPI”). Pozostaje (wymaga zgody): N2.

### RUNTIME_NATIVE:N0–N1 — kontrakt śladu i importer offline — DONE (2026-11-08)

- **Zakres zatwierdzony (2026-11-08):** N0 (wersjonowany format śladu i
  zasady wiarygodności) i N1 (importer offline, korelacja, raport/v2/
  snapshot) bez agenta Java; jako pierwszy rejestrator — konwerter
  nagrania Java Flight Recorder. Gałąź `runtime-n0-n1` od
  `servlet-dispatch` (PR #23 jeszcze niezmergowany).
- **Zrealizowane:**
  1. `docs/RUNTIME_TRACE.MD`: format `ariadne-runtime-trace-1` (JSON
     Lines; nagłówek z metadanymi przebiegu, zakresem, próbkowaniem i
     kompletnością; artefakty, loadery, definicje klas per loader z
     hashem i `source_basis`, metody; dowody `loaded`/`executed`/`call`/
     `handoff` z id), limity, reguły czytania, statusy korelacji.
  2. `scripts/JfrToTrace.java`: konwersja nagrania JFR (API
     `jdk.jfr.consumer`, bez narzędzi spoza JRE): klasy tylko z podanych
     artefaktów (hash pliku klasy z classpath — JFR nie podaje źródła
     bajtów), loader z `jdk.ClassLoad`/ramek, `executed` z próbek
     (`sampled`, raz na próbkę) i z `jdk.MethodTrace` (`traced`),
     relacje z sąsiednich ramek (`stack-adjacent`) lub przez ramki spoza
     zakresu (`stack-through-runtime`), z bci i linią wywołującego;
     ucięte stosy liczone; przyczyny niekompletności w nagłówku.
  3. `internal/runtimeevents` (`ImportVersion = "n1-1"`): `Parse`
     (ścisła walidacja, limity 1 GiB / 1 MiB na linię / 5 mln rekordów /
     4 KiB na napis), `Correlate` (statusy klas po SHA-256 pliku klasy lub
     artefaktu, dwie definicje osobno, metody → węzły grafu, wywołania →
     `in_graph` z podstawami krawędzi / `not_in_graph` / `unusable`),
     `ObservedVulnerable` (mapowanie bez deskryptora = każdy overload),
     `NotInGraph`.
  4. CLI: `-runtime-trace` (wczytanie i walidacja przed analizą),
     `cmd/ariadne/runtime.go` (hash klasy z wpisu jara, z którego klasa
     została przeczytana; obserwacje per ustalenie i instancja), log,
     raport „Runtime evidence”, v2 `runtime` i
     `findings[].runtime_observed`, snapshot: wejście `runtime-trace`,
     konfiguracja `runtime-trace`, wersja `runtime.import` (tylko ze
     śladem).
  5. `testdata/runtime` + `scripts/runtime-trace-fixture.sh`: `rtm.jar`,
     `rtm-v2.jar` (inna kompilacja `Square`), `run.trace` nagrany na JDK 25
     (wywołanie przez interfejs, `Lib.parse` śledzone dokładnie 1000 razy,
     klasa tylko załadowana, `Plugin` w loaderze aplikacji i w izolowanym
     `URLClassLoader`, wywołanie refleksyjne o nazwie budowanej w czasie
     działania); `scripts/runtime-trace-petclinic.sh`.
- **Weryfikacja:** `TestParseFixture`, `TestParseRejects` (20 naruszeń
  kontraktu odrzucanych), `TestCorrelateStatuses` (matched / not_in_input
  / different_version / name_only, dwie definicje), `TestObservedVulnerable`,
  `TestRuntimeTrace` (pełne CLI: 7 klas `matched`, `rtm.Plugin` z dwiema
  definicjami, 1 metoda `traced`, 2 wywołania refleksyjne `not_in_graph`
  — po jednym na loader —, graf i ustalenia identyczne jak bez śladu; na
  `rtm-v2.jar` `Square` `different_version`, jego próbki nieużyte),
  `TestRuntimeTraceSnapshot` (wejście i wersja w manifeście, zły plik
  odrzucony przed analizą); mutacja (dopasowanie bez porównania hashy) —
  oba testy zawodzą. Prawdziwy przebieg: petclinic na JDK 25 pod JFR
  (strony + 200 × zasoby statyczne z `If-None-Match`), dwa niezależne
  przebiegi skryptu: 10 527 klas `matched`, ok. 2 400–2 650 metod
  wykonanych, 2 414–2 746 relacji `in_graph`, 217–251 `not_in_graph`, 244–
  319 uciętych stosów; CVE-2024-38809 `ServletWebRequest#
  validateIfNoneMatch` `traced` 1 409 / 1 419 razy (statycznie FRAMEWORK);
  CVE-2025-35036 `ValidatorFactoryImpl#<init>` wykonane (`sampled`), a
  statycznie NOT_REACHABLE z `MODEL_INCOMPLETE` — wywołujący w śladzie:
  `HibernateValidator#buildValidatorFactory`, osiągany przez
  `javax.validation.spi.ValidationProvider#buildValidatorFactory`
  (pomijany odbiorca `javax.*`); 12–13 advisories z zaobserwowaną podatną
  metodą (liczba zmienna — próbkowanie); werdykty bez zmian
  (3/74/13/57/6). `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `docs/RUNTIME_TRACE.MD` (nowy), `docs/SUPPORT.MD`
  (wiersz Runtime evidence), `docs/RESULTS_CONTRACT.MD` (`runtime`,
  `findings[].runtime_observed`), README (sekcja, flaga, skrypty,
  ograniczenia), FAQ, plan RUNTIME_NATIVE, ten rejestr.
- **Ograniczenia:** dowód jednego przebiegu; JFR nie podaje źródła bajtów
  klasy (hash z classpath, `source_basis` `classpath-lookup`); id loadera
  nie są stabilne między przebiegami; bez agenta (N3), bez analizy
  kontekstu loaderów poza rozdzieleniem definicji (N2), bez korelacji
  handoff z JFR (N4), bez native/JNI (N5); filtr `jdk.MethodTrace` trzeba
  podać ręcznie.
- **Wydanie i archiwum:** zmergowane jako PR #24 (2026-11-08); tekst faz
  przeniesiony do [archive/2026-11-08/runtime_native/RUNTIME_NATIVE_N0_N1.md](../archive/2026-11-08/runtime_native/RUNTIME_NATIVE_N0_N1.md).
- **Następny krok:** lista `jdk.MethodTrace#filter` z mapowań CVE —
  zrobione (wpis „RUNTIME_NATIVE:N1 — filtr dokładnego śledzenia”).
  Model wywołań SPI `javax.validation` — zrobione (wpis „SERVICE_LOADING
  (po zamknięciu obszaru): dispatch SPI”). Pozostaje (wymaga zgody): N2
  (kontekst loaderów).

### GRAPH_PRECISION:P4 — benchmark i kwalifikacja — DONE (2026-11-08)

- **Zakres zatwierdzony (2026-11-08):** pomiar CHA vs RTA na fixtures z
  wyrocznią JVM i na petclinic/WebGoat w obu scenariuszach korzeni: błędne
  i pominięte cele, nieznane miejsca, zmiany werdyktów CVE, czas, pamięć;
  wersje reguł w snapshotach; ręczny opis każdej zmiany werdyktu; decyzja
  o domyślnej strategii; dokumentacja migracji. P3 (points-to) odłożone do
  decyzji po wynikach. Gałąź `graph-precision-p4` od `main` (`2082ff7`).
- **Zrealizowane:**
  1. `scripts/graph-benchmark.sh`: na konfigurację `--repeat` mierzonych
     przebiegów bez pliku grafu (czas ścienny i szczytowe RSS procesu przez
     `os.wait4`) i jeden niemierzony z `-graph-out` (CHA) /
     `-graph-compare` (RTA); każda liczba poza czasem/pamięcią musi być
     identyczna we wszystkich przebiegach. Zbiera z v2 i logu: rozmiar
     grafu, krawędzie CHA (usunięte / tylko przez fallback), klasy żywe,
     aktywne luki alokacji, katalog, miejsca luk modelu (wszystkie /
     osiągalne), werdykty v1, porównanie z CHA. Wynik:
     `testdata/graphbench/{results.tsv,verdict-changes.tsv,versions.txt}`.
     Pierwsza wersja liczyła odczyt ~350 MB pliku grafu do kosztu RTA
     (+1 GB RSS) — stąd rozdzielenie pomiaru i porównania.
  2. `cmd/ariadne/graphbench_test.go` (`TestGraphBenchmarkFixtures`): dla
     `testdata/callgraph` i `testdata/rta` — ciała śledzone (wołające
     helper śladu) uruchomione przez JVM, a nieosiągnięte (pominięte) i
     osiągnięte, a nieuruchomione (nadmiarowe); asercje: CHA nic nie gubi,
     RTA gubi tylko udokumentowany `Unnamed#area`, nadmiar RTA ⊆ nadmiar CHA.
  3. `cmd/ariadne/snapshot.go`: `algorithmVersions(mode)` — w trybie `rta`
     także `alloc.catalog` i `callgraph.rta` (CHA bez zmian, żeby nie
     unieważniać snapshotów CHA); `TestSnapshotAlgorithmsRTA`.
  4. `testdata/graphbench/RESULTS.MD`: wyniki, ręczna kontrola każdej
     zmiany werdyktu (skan bytecode wszystkich jarów po nazwach z `/` i
     `.`, BFS po wyeksportowanych grafach), decyzja, migracja.
- **Wyniki:** fixtures — `callgraph` 13/13 obserwowanych w obu trybach, 0
  nadmiarowych; `rta` przy `-roots profile` CHA 0 pominiętych / 2 nadmiarowe
  (`Ghost`, `Hidden`), RTA 1 pominięte (`Unnamed`, luka aktywna) / 0
  nadmiarowych; przy `-roots all` oba 0 / 2. Aplikacje (3 przebiegi):
  petclinic CHA 13,5–17,9 s / 1335–1394 MB, RTA 13,9–14,7 s / 1315–1435 MB;
  WebGoat CHA 23,1–23,4 s / 1772–1829 MB, RTA 23,8–24,3 s / 1776–1819 MB.
  `-roots profile`: usunięte 52 561 z 310 418 (petclinic) i 95 378 z
  556 769 (WebGoat) krawędzi CHA, z zachowanych 50 305 / 178 188 tylko przez
  fallback; osiągalne miejsca luk 596 → 532 / 676 → 590; `-roots all`:
  6 113 / 14 823 usuniętych, werdykty bez zmian. Zmiany werdyktów
  (FRAMEWORK → NOT_REACHABLE, wszystkie z `MODEL_INCOMPLETE`):
  CVE-2022-45143, CVE-2025-55752, CVE-2026-53404, CVE-2026-55957
  (petclinic; zawory/realm Tomcata — żadna klasa wejścia ich nie tworzy poza
  `MBeanFactory#createJNDIRealm`, operacją JMX) i CVE-2021-22118 (WebGoat;
  reaktywny multipart w aplikacji servlet) — poprawne; CVE-2024-22259
  (petclinic; ścieżka CHA przez nieużywany `CorsWebFilter`, ścieżka
  servletowa niewidoczna w obu trybach) — nierozstrzygnięta;
  CVE-2024-38809 (oba) — fałszywie negatywne: `ResourceHttpRequestHandler#
  handleRequest` → `ServletWebRequest#checkNotModified(J)` →
  `validateIfNoneMatch` wykonuje się przy każdym żądaniu zasobu
  statycznego, ale `ApplicationFilterChain#internalDoFilter` →
  `javax.servlet.Servlet#service` nie jest rozwijane (skip `javax.`), więc
  `DispatcherServlet` jest nieosiągalny w obu trybach, a RTA usunęło
  artefakt (reaktywny `WriteBarrier` → servletowy
  `DeferredResultSubscriber#onNext`), który trzymał CHA na FRAMEWORK.
  Żaden REACHABLE nie zmieniony.
- **Decyzja:** domyślna strategia pozostaje CHA; RTA opt-in jako druga
  opinia z `-roots profile` i `-graph-compare`. Ponowna ocena po
  zamodelowaniu dispatchu kontener → servlet/filter. Points-to (P3)
  nieuzasadnione: o większości zachowanych krawędzi decyduje fallback
  (refleksja), nie precyzja dispatchu.
- **Weryfikacja:** `scripts/graph-benchmark.sh --repeat 3` (zielony, 32
  przebiegi), `TestGraphBenchmarkFixtures`, `TestSnapshotAlgorithmsRTA`,
  `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `testdata/graphbench/RESULTS.MD`, `docs/SUPPORT.MD`
  (wiersz RTA and points-to: benchmark, decyzja, ograniczenie skip),
  README (sekcja 4, skrypt), plan GRAPH_PRECISION, ten rejestr.
- **Ograniczenia:** dwie aplikacje Spring, jeden profil; czas/pamięć z
  jednej maszyny (3 przebiegi, zakres min–max); bez zewnętrznego silnika
  referencyjnego; „nadmiarowe” na fixtures to ciała nieuruchomione w tym
  przebiegu, nie dowiedzione błędne.
- **Wydanie i archiwum:** zmergowane jako PR #21 (`95ae6ec`, 2026-11-08);
  model dispatchu servletów (PR #22) zmergowany do gałęzi `graph-precision-p4`
  po merge'u #21 i wprowadzony do `main` razem z archiwizacją; tekst fazy
  przeniesiony do [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P4.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P4.md).
- **Następny krok:** zrobione 2026-11-08 — model dispatchu kontener →
  servlet/filter (`servlet-container:dispatch`, wycinek FRAMEWORKS F2) i
  ponowny benchmark (petclinic: fałszywy negatyw usunięty). Pozostaje
  (wymaga zgody): WebGoat bez launchera Boot w wejściu; P3 dopiero gdy
  pomiar to uzasadni.

### GRAPH_PRECISION:P2 — RTA i punkt stały — DONE (2026-11-08)

- **Zakres zatwierdzony (2026-11-08):** opt-in `-callgraph-mode rta`
  z polityką `cha-fallback-with-gap` ustaloną w P0, na katalogu P1;
  fixtures z trzema implementacjami, pośrednią alokacją, cyklem i nieznaną
  fabryką; porównanie z CHA przez `-graph-compare`, `-roots profile` i
  `-roots all` jako osobne scenariusze. Domyślny tryb (`cha`) bez zmian.
  Gałąź `graph-precision-p2` od `archive-graph-precision-p1` (PR #18).
- **Zrealizowane:**
  1. `internal/callgraph/rta.go` (`RTAVersion = "p2-1"`): strategia
     `RTA` (przy budowie zbiór kandydatów CHA, w metadanych tryb `rta` i
     polityka `cha-fallback-with-gap`); `ApplyRTA(g, h, RTAInputs)` —
     punkt stały: kolejka osiągalnych metod (krawędzie inne niż CHA
     zawsze przechodzone), klasy żywe z `Creation` (miejsce-metoda albo
     od startu), cele CHA bez żywej klasy parkowane w indeksie po klasie
     deklarującej i budzone przy ożywieniu klasy-potomka (przodkowie
     klasy), decyzja: istnieje żywa klasa `C` z `ResolveExact(C, nazwa,
     deskryptor)` = klasa celu. Na końcu każdy odrębny cel CHA oceniany
     raz wobec końcowego zbioru żywych klas (`created` > `fallback` >
     brak); krawędzie bez żywej klasy usuwane, zachowane tylko przez
     fallback liczone. `RTAResult`: metody osiągalne, klasy żywe (w tym
     tylko przez fallback), krawędzie CHA przed/usunięte/przez fallback,
     aktywne luki wg rodzaju.
  2. `internal/classfile/names.go` (`NamedClasses`: literały klas
     `ldc`/`ldc_w` i stałe tekstowe w postaci nazwy binarnej klasy),
     `hierarchy.ClassInfo.NamedClasses`.
  3. `cmd/ariadne/rta.go`: tłumaczenie katalogu (pochodzenia → `Creation`;
     `<container>`/miejsce spoza grafu → od startu; luki →
     `UnknownCreation` z fallbackiem: `reflection-unresolved`/`unsafe`/
     `constructor-handle` → `nameableClasses` (nazwane w kodzie, z
     adnotacją klasy poza `kotlin.Metadata`, typy pól/parametrów/wyników;
     tylko klasy konkretne), `service-unresolved` → zarejestrowani
     providerzy, `external-input` → konkretne podtypy typów parametrów
     korzeni, proxy → brak); `rtaGapSites` (`Unsafe`/uchwyty konstruktora
     jako luki modelu `allocation`, nieograniczone). `main.go`: katalog
     liczony przed silnikiem osiągalności (w trybie CHA bez wpływu), RTA
     po koordynatorze modeli, log, raport „RTA”, v2 `callgraph.rta`;
     opis flagi.
  4. `testdata/rta` + `scripts/rta-oracle.sh` (javac + JVM; zasób
     `META-INF/services`) → `rta.jar`, `observed.txt` (9 ciał metod).
- **Weryfikacja:** `TestRTAOracle` (pełne CLI, bez JDK): każde ciało
  uruchomione przez JVM osiągalne z `main` w grafie RTA poza `Unnamed#area`
  (klasa z fabryki refleksyjnej o nazwie budowanej w czasie działania —
  oczekiwana strata pokryta aktywną luką `reflection-unresolved`);
  usunięte dokładnie 3 krawędzie CHA (`Ghost` — nigdy nie tworzony,
  `Hidden` — tworzony w nieosiąganej metodzie, `Unnamed`), `Dynamic`
  (z adnotacją) zachowany przez fallback; cykl `StepA`↔`StepB` z `StepC`
  tworzonym dopiero w `StepB#go` zachowany; provider ServiceLoadera
  zachowany; `-roots all`: 0 usuniętych, aktywna luka `external-input`.
  Mutacje: pominięcie alokacji odkrytych w trakcie (5 ciał traconych) i
  pominięcie natychmiastowego sprawdzenia dispatchu (`StepC` tracony) —
  test zawodzi. `TestRTAOrderIndependent` (5 permutacji kreacji, korzeni
  i kolejności krawędzi → identyczny wynik i graf), `TestRTAFallbackAndGap`,
  `TestNameableClasses`, `TestStrategyForRTA`. Real data (offline,
  `-profile spring`, `-graph-compare` z grafem CHA tej samej gałęzi):
  petclinic `-roots profile`: 83 421 metod osiągalnych, 15 818 klas żywych
  (7 562 tylko przez fallback), usunięte 52 561 z 310 418 krawędzi CHA
  (50 305 zachowanych tylko przez fallback); werdykty FRAMEWORK 71 → 65,
  NOT_REACHABLE 16 → 22: CVE-2022-45143 (`JsonErrorReportValve`),
  CVE-2025-55752 i CVE-2026-53404 (`RewriteValve`), CVE-2026-55957
  (`JNDIRealm`) — komponenty Tomcata tworzone tylko z `server.xml`,
  którego Spring Boot petclinic nie używa; CVE-2024-22259 (`CorsWebFilter`)
  i CVE-2024-38809 (`ChannelSendOperator`) — stos reaktywny nieużywany
  przez aplikację servlet MVC; wszystkie 6 z `MODEL_INCOMPLETE` (luki
  `allocation`, `reflection` i in.). `-roots all`: usunięte 6 113,
  werdykty bez zmian. WebGoat `-roots profile`: usunięte 95 378 z 556 769
  (178 188 przez fallback), FRAMEWORK 51 → 49 (CVE-2021-22118 i
  CVE-2024-38809 — stos reaktywny), `-roots all`: 14 823, werdykty bez
  zmian. Żaden REACHABLE nie zmieniony w żadnym scenariuszu. Czas
  przejścia RTA 0,6–1,3 s (całość: petclinic 18 → 21–22 s, WebGoat
  27–30 → 31–33 s). Dwa przebiegi RTA: `-graph-out` bajtowo identyczny,
  v2 identyczne poza `comparison`. `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz RTA and points-to →
  Partial), README (sekcja 4, `-callgraph-mode`, skrypt wyroczni), FAQ
  (dlaczego RTA jest opt-in), `docs/RESULTS_CONTRACT.MD` (`callgraph.rta`,
  luka `allocation`), szablon raportu, PLAN.md (Faza 7 i backlog),
  plan GRAPH_PRECISION, ten rejestr, skill catch-up.
- **Ograniczenia:** fallback jest heurystyką z uzasadnieniem, nie dowodem:
  klasa tworzona refleksyjnie i nazwana tylko w zasobie/konfiguracji
  (np. zawór Tomcata w `server.xml`) albo tylko w sygnaturze generycznej
  nie jest żywa — luka oznacza negatyw `MODEL_INCOMPLETE`, nie przywraca
  krawędzi; fallback jest globalny (klasy, nie miejsca wywołań); krawędzie
  bezpośrednie do implementacji w nadklasie nie są zawężane; usuwane są
  też krawędzie do abstrakcyjnych deklaracji (nie prowadzą dalej);
  z `-roots all` RTA usuwa niewiele (każda metoda jest korzeniem).
  Domyślny tryb pozostaje CHA — zmiana domyślnej strategii to osobna
  decyzja po P4.
- **Wydanie i archiwum:** zmergowane jako PR #19 (`2082ff7`, 2026-11-08;
  zawiera też archiwizację P1 z zamkniętego PR #18); tekst fazy przeniesiony
  do [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md).
- **Następny krok (wymaga zgody):** GRAPH_PRECISION:P3 — ograniczona
  analiza points-to (DATA_FLOW D3), albo P4 — benchmark i kwalifikacja
  (pomiar błędnych/pominiętych celów, decyzja o domyślnej strategii).

### GRAPH_PRECISION:P1 — katalog typów tworzonych jawnie i pośrednio — DONE (2026-11-08)

- **Zakres zatwierdzony (2026-11-08):** katalog klas tworzonych (`new` i
  modele) z pochodzeniem, jawne luki dla niemodelowanych mechanizmów,
  reguła wejść zewnętrznych przy `-roots all`; raport, v2, lista w
  `-graph-out` (rozszerzenie formatu zatwierdzone); wyrocznia JVM; tylko
  raport — graf i werdykty bez zmian. Gałąź `graph-precision-p1` od
  `archive-graph-precision-p0` (PR #16).
- **Zrealizowane:**
  1. `internal/alloc` (`CatalogVersion = "p1-1"`): `Build(Inputs)` —
     `new` z `InitRefs` (tylko klasy konkretne); krawędzie modeli do
     `<init>` w końcowym grafie (`serviceloader` → `service`, `reflection`,
     `model`); beany kontenera (stereotypy, konkretne typy zwracane przez
     `@Bean`); kandydaci deserializacji per format z wejściami tego formatu;
     klasy metod-korzeni instancyjnych przy `-roots profile` (`entry`);
     miejsca lambd; luki: `reflection-unresolved`, `dynamic-proxy`,
     `unsafe-allocation`, `constructor-handle`, `service-unresolved`,
     `container-proxy` (z ograniczeniem do interfejsu), `external-input`
     (`-roots all`, z liczbą parametrów referencyjnych). `Summarize`:
     liczby wg pochodzenia i luk, `cha_edges_to_classes_not_created`
     (informacyjnie).
  2. `cmd/ariadne/alloc.go` (zbieranie faktów modeli), log, raport (sekcja
     „Created types”), v2 `callgraph.created_types`, `-graph-out` →
     `ariadne-callgraph-2` z rekordami `type` (klasa, pochodzenia z liczbą
     miejsc); `-graph-compare` czyta też v1 i podaje dodane/usunięte typy.
  3. `testdata/alloc` + `scripts/alloc-oracle.sh` (javac + JVM; zasób
     `META-INF/services` dopisywany do JAR-a) → `alloc.jar`, `observed.txt`.
- **Weryfikacja:** `TestCreatedTypesOracle` (pełne CLI na fixture, bez
  JDK): 7 klas utworzonych przez JVM — `Square`/`BigSquare`/`Circle`
  (`new`), `Hexagon` (`reflection`), `PluginImpl` (`model`: provider
  ServiceLoadera materializowany przez model B3), `Saved`
  (`deserialization:java`), proxy (luka `dynamic-proxy`); jedyny nadmiar
  `Never` (`new` w metodzie, której nic nie woła — katalog nie zna
  osiągalności, to P2); klasa abstrakcyjna nieobecna; mutacja (bez
  pochodzenia `reflection`) → test zawodzi. `TestOriginsAndGaps`,
  `TestTypesInExport`. Real data (offline, binarka `main` vs gałąź):
  petclinic 17 763 klas (container 7, deserialization 497, model 7, new
  17 710, reflection 6; 53 tylko przez model), 6 358 miejsc lambd, luki 329
  (reflection-unresolved 258, dynamic-proxy 59, service-unresolved 6,
  container-proxy 4, unsafe 1, external-input 1); WebGoat 26 594 (146 tylko
  przez model), 6 249 lambd, luki 328; informacyjnie krawędzie CHA do klas
  spoza katalogu: petclinic 57 970 / 310 418, WebGoat 92 779 / 556 769.
  `-graph-compare` z grafem `main`: 0 krawędzi dodanych/usuniętych, 0
  zmienionych ustaleń; v2 różni się tylko `created_types`; v2 i
  `-graph-out` bajtowo identyczne w dwóch przebiegach; czas WebGoat
  23,5–27,4 s (`main`) vs 23,7–28,7 s (w granicach szumu). `go test ./...`,
  `go vet ./...` — PASS.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz RTA and points-to), README
  (sekcja 4, `-graph-out`, skrypt wyroczni), `docs/RESULTS_CONTRACT.MD`,
  szablon raportu, plan GRAPH_PRECISION, ten rejestr.
- **Ograniczenia:** katalog bez osiągalności (to P2); deserializacja bez
  odrzuceń polityki (nadmiar, bezpieczny); typy zwracane przez `@Bean`
  tylko konkretne (interfejs: klasa z `new` w ciele metody fabrycznej);
  tablice nieskatalogowane; `clone()` nie tworzy nowego typu; w eksporcie
  liczba miejsc per pochodzenie, nie same miejsca.
- **Wydanie i archiwum:** zmergowane jako PR #17 (`46076b5`, 2026-11-08);
  archiwizacja weszła do `main` z PR #19 (PR #18 zamknięty); tekst fazy przeniesiony do
  [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md).
- **Następny krok (wymaga zgody):** GRAPH_PRECISION:P2 — RTA i punkt stały
  (opt-in `-callgraph-mode rta`, polityka `cha-fallback-with-gap`, fixtures
  z trzema implementacjami, pośrednią alokacją, cyklem i nieznaną fabryką,
  porównanie z CHA przez `-graph-compare`).

### GRAPH_PRECISION:P0 — uczciwy baseline i tryby porównania — DONE (2026-11-07)

- **Zakres zatwierdzony (2026-11-07):** rozdzielenie strategii dispatch od
  budowy grafu z zachowaniem CHA przez adapter; ustawienia grafu w
  metadanych; stabilne identyfikatory krawędzi i porównanie grafów;
  przypadki ze znanymi poprawnymi celami (z domknięciem weryfikacji
  bridge/default z JVM_GRAPH:G2 na prawdziwym bytecode); polityka nieznanej
  alokacji. Gałąź `graph-precision-p0` od `archive-data-flow` (PR #14).
- **Zrealizowane:**
  1. `internal/callgraph/strategy.go`: `Strategy` (`Mode`,
     `DispatchTargets`, `UnknownAllocationPolicy`, `SkipRule`), `CHA`
     (dotychczasowa logika: podtypy zadeklarowanego odbiorcy z dokładnym
     deskryptorem, skip JDK), `StrategyFor`, `Config` + `Differences`;
     polityki `all-subtypes` (CHA) i zarezerwowana `cha-fallback-with-gap`
     (RTA: zachować zbiór CHA w miejscu z odbiorcą z niemodelowanego
     mechanizmu i zapisać lukę ograniczającą wynik negatywny).
     `build.go`: `BuildWith(h, strategy)`, `Build` = adapter CHA.
     `Graph.Config()`/`SetAnalysis(roots, targetJava)`.
  2. `internal/callgraph/compare.go`: `Snapshot` (krawędzie końcowego grafu
     z tożsamością metoda→metoda/rodzaj/podstawa i PC, ścieżki ustaleń),
     `Write`/`Read` (`ariadne-callgraph-1`, posortowany tekst), `Compare`
     (różnice ustawień, dodane/usunięte wg podstawy, zmiana dowodów PC,
     ustalenia ze zmienionym werdyktem/ścieżką z krawędziami
     wyjaśniającymi, ustalenia tylko po jednej stronie).
  3. CLI: `-callgraph-mode` (tylko `cha`, nieznany tryb → błąd), `-graph-out`
     (zapis atomowy razem z raportem), `-graph-compare` (plik czytany przed
     analizą); log porównania; snapshot config `callgraph-mode`. Raport:
     ustawienia grafu w nagłówku, sekcja „Call graph compared with …”. v2:
     `callgraph` (ustawienia, rozmiar, `comparison`).
  4. `testdata/callgraph` (`cg.Sites`, kształty, `cg.Oracle`, `Trace`) +
     `scripts/callgraph-oracle.sh` (javac + JVM) → `cg.jar`, `observed.tsv`.
- **Weryfikacja:** `TestKnownTargets` — wszystkie 13 ciał uruchomionych
  przez JVM obecne w grafie, 3 nadmiarowe z uzasadnieniem (bridge generyczny,
  bridge kowariantny, pomocnik wywołania super), brak krawędzi CHA do metody
  niezadeklarowanej; mutacja (dopasowanie po nazwie) → test zawodzi;
  błędów deskryptorów do korekty nie znaleziono. `TestExportRoundTrip`,
  `TestCompareExplains`. Real data (offline, binarka `main` vs gałąź):
  werdykty petclinic 3/71/16/57/6, WebGoat 42/51/10/61/0 bez zmian; v2 różni
  się tylko nową sekcją `callgraph` (oraz `analysis_id`/wersją), raport
  tylko linią nagłówka grafu; dwa przebiegi bajtowo identyczne; `-graph-out`
  bajtowo identyczny w dwóch przebiegach (petclinic 991 314 krawędzi,
  WebGoat 1 409 559, ≈ 370 MB); porównanie grafu z samym sobą puste;
  porównanie petclinic z `-roots profile` ujawnia różnicę `roots` i 9 ustaleń
  ze zmienioną ścieżką bez zmiany krawędzi. `go test ./...`, `go vet ./...`
  — PASS.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersze Direct calls/CHA i RTA),
  README (flagi, sekcja 4, skrypt wyroczni), `docs/RESULTS_CONTRACT.MD`,
  szablon raportu, plan i wymagania GRAPH_PRECISION, addendum w
  `archive/2026-10-22/jvm_graph/COMPLETION.md`, ten rejestr.
- **Ograniczenia:** jedyny tryb to CHA (RTA — P2); plik `-graph-out` jest
  duży dla dużych aplikacji; `EdgeCount` liczy też powtórzone tożsamości
  krawędzi dodane osobno przez modele (eksport je scala: petclinic 991 328 vs
  991 314); ścieżki ustaleń w eksporcie to wersja v1 (jedna na advisory).
- **Wydanie i archiwum:** zmergowane jako PR #15 (`d9614c1`, 2026-11-07);
  tekst fazy przeniesiony do
  [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md).
- **Następny krok (wymaga zgody):** GRAPH_PRECISION:P1 — katalog typów
  tworzonych jawnie i pośrednio (alokacje `new`, typy z modeli SERVICE/
  FRAMEWORK/REFLECTION/DYNAMIC/DESERIALIZATION, luki dla niemodelowanych
  mechanizmów).

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

### SERVICE_LOADING:S2 — Materializacja i krawędzie dispatch — DONE for declared scope, zarchiwizowane (2026-11-05; wpisy poniżej z 2026-10-26)

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
- **Niespełnione/odłożone:** powiązanie iteratora/stream/uchwytu
  `Provider` z konkretnym `load()` i rejestracją oraz krawędzie do
  materializacji/dispatch pozostają otwarte. Kontynuacja poniżej dodaje
  dokładne dopasowanie tych sygnatur i jawne luki, lecz nie zgaduje celu.
  "Typy przekazać do
  późniejszego RTA" — nie dotknięte (RTA samo nie istnieje jeszcze,
  GRAPH_PRECISION). Java 9+'s modułowy `provider()` (statyczna
  fabryka) — poza zakresem, to S3 (JPMS). Weryfikacja loadera
  (czy jawnie podany `ClassLoader`/`ModuleLayer` faktycznie widzi dany
  provider) — nie sprawdzana; S2 łączy niezależnie od `LoaderKind`,
  co jest świadomą, odnotowaną niedokładnością (nadmiarowa, nie
  brakująca, ścieżka), nie przemilczanym założeniem.
- **Następny krok:** kontynuować `SERVICE_LOADING:S2` przez
  międzywywołaniowe śledzenie tożsamości load → iterator/stream → provider
  i dispatch metod. `SERVICE_LOADING:S3` (JPMS) pozostaje odrębne.

#### SERVICE_LOADING:S2 — D3 field provenance (2026-10-26)

- **Zakres:** kontynuacja poprzedniego zatwierdzonego przyrostu S2; wykorzystać
  D3 dla pól, bez points-to i bez tworzenia celów dla nierozstrzygniętych pól.
- **Zrealizowane:** `ResolveFieldOneHop` zwraca bezpośrednie streszczenie D3,
  a `serviceLoadOrigins` śledzi jego wynik z własnym limitem głębokości i
  ochroną przed cyklami wspólną dla ścieżki pól/helperów. Działa dla pól
  statycznych oraz `final` pól instancji zapisywanych przez `this` w
  konstruktorach — dokładnie w kształcie udostępnianym przez D3. Pozostałe
  pola i parametry pozostają nierozstrzygnięte.
- **Weryfikacja w tym przyroście:** JDK zawiera moduł `jdk.compiler`, mimo że
  nie udostępnia komend `javac`/`javap`; artefakty fixture odświeżono z JDK
  przez `java --module jdk.compiler/com.sun.tools.javac.Main -source 8 -target 8`.
  `TestCompiledServiceLoaderFixture` PASS: przepływ przez `static final` pole
  dochodzi do konstruktora i descriptor-exact dispatch providera; `mutable`
  pole instancji oraz helper bez wartości pozostają jawnymi gapami bez
  provider edges. `GOCACHE=/tmp/ariadne-gocache go test ./internal/serviceload
  ./internal/dataflow`, `go test ./...`, `go vet ./...`, `git diff --check`
  oraz `go build -o /tmp/ariadne-s2-fields ./cmd/ariadne` — PASS. Kontrolne
  przebiegi bez nowych wpisów OSV offline: petclinic 29,123 klas, 29 load
  sites i 57 consumption/dispatch (39 linked); WebGoat 40,569 klas, 48 load
  sites i 115 consumption/dispatch (90 linked). Późniejszy etap analizy CVE
  przerwał brakujący wpis OSV; pełne wyniki reachability tych przebiegów nie
  są raportowane.
- **Pozostało w S2:** inne pola, parametry, wirtualne helper returns oraz
  stream callback/pipeline flows pozostają nieobsługiwane lub jawnie
  niepełne. Widoczność przez explicit/system/installed loaderów pozostaje
  gapem. Bazowy CHA nadal dodaje szersze cele; RTA nie istnieje. JPMS S3
  pozostaje odrębne.

#### SERVICE_LOADING:S2 — skompilowany fixture pól (2026-10-26)

- **Zrealizowane:** rozszerzono `Entry` w fixture o `static final LOADER`
  inicjalizowany przez `ServiceLoader.load`, konsumenta tego pola oraz
  konsumenta pola instancji mutowalnego. Zaktualizowano checked-in classfiles
  do classfile major 52. Test integracyjny sprawdza pełny graf przez pole
  statyczne (konstruktor i descriptor-exact dispatch providera) oraz
  potwierdza, że ścieżka pola mutowalnego daje `Incomplete` bez modelowych
  krawędzi. Poprzednie plany „uzupełnić fixture” są tym samym zamknięte.
- **Weryfikacja:** `TestCompiledServiceLoaderFixture -count=1` PASS;
  pełne `go test ./...`, `go vet ./...`, build i `git diff --check` PASS.
  Świeży petclinic `-offline -profile spring`: 29,123 klas, 29 load sites,
  57 consumption/dispatch (39 linked), 4 model edges; analiza osiągalności
  zakończyła się (`2/15/9/15/6`). Świeży WebGoat: 40,569 klas, 48 load sites,
  115 consumption/dispatch (90 linked); analiza doszła do OSV, lecz zakończyła
  się błędem braku wpisu `b1e63ee08008e9916ecd8fb4d3495aacb7c0edffdb844a556869cb679b822a55`
  w offline cache, więc reachability WebGoat nie jest potwierdzona w tym
  przebiegu.
- **Pozostało:** tylko dalsze ograniczenia S2 wymienione wyżej; osobne
  `SERVICE_LOADING:S3` obejmuje JPMS.

#### SERVICE_LOADING:S2 — final instance field i cykle (2026-10-26)

- **Zrealizowane:** fixture ma teraz `final` pole instancji inicjalizowane
  przez `ServiceLoader.load()` w konstruktorze. Sprawdza dodatni przepływ
  `constructor load → final field → iterator → next → provider.decode`,
  w tym krawędzie do konstruktora i dokładnej metody providera. Dodano też
  dwa statyczne pola z cyklicznymi referencjami. Cykliczne i mutowalne pole
  kończą analizę jawnym `Incomplete`, bez provider edges.
- **Weryfikacja:** fixture skompilowany JDK-owym `javac` przez
  `java --module jdk.compiler/com.sun.tools.javac.Main -source 8 -target 8`;
  classfile major 52. `go test ./internal/serviceload -run
  TestCompiledServiceLoaderFixture -count=1`, `go test ./...`, `go vet ./...`,
  build i `git diff --check` — PASS. Petclinic świeżo: 29,123 klas, 29 load
  sites, 57 consumption/dispatch (39 linked), 4 model edges i zakończona
  reachability `2/15/9/15/6`. WebGoat: 40,569 klas, 48 load sites,
  115 consumption/dispatch (90 linked); przebieg znów zatrzymał się na tym
  samym brakującym wpisie OSV offline, więc reachability WebGoat pozostaje
  niezweryfikowana w tym przebiegu.
- **Pozostało:** nowe przypadki pól bezpiecznych i cyklicznych są zweryfikowane
  end-to-end. Inne ograniczenia S2 pozostają wymienione wyżej; JPMS jest S3.

#### SERVICE_LOADING:S2 — podstawienie zwracanego parametru (2026-10-26)

- **Zakres:** dokładny `invokestatic`/`invokespecial` helper, którego ciało
  zwraca formalny parametr; rozwiązać ten marker tylko z argumentu na jego
  konkretnym call site. Nie rozciągać tego na wirtualną dispatch lub dowolne
  parametry zewnętrznych callerów.
- **Zrealizowane:** D2 może opcjonalnie zachować indeks formalnego parametru
  przy analizie zwrotów. `ResolveCallResultOneHopWithParameterRefs` oddaje
  parametr wraz z indeksem, a `ResolveCallArgumentAtCall` odtwarza argument z
  metody caller'a wskazanej w `CallResultRef.Site`. S2 podstawia tę wartość
  i kontynuuje ograniczony traversal; zestawy branchy są podstawiane
  kandydat po kandydacie. Zwykła analiza D2 pozostaje bez zmian, chyba że
  konsument jawnie zażąda markerów.
- **Weryfikacja:** skompilowany Java 8 fixture sprawdza `load → passLoader(x)
  → return x → iterator → next → provider.decode` z krawędziami do
  konstruktora i metody; caller z nieznanym parametrem nie dostaje interfejsu
  usługi ani modelowych celów. `GOCACHE=/tmp/ariadne-gocache go test
  ./internal/serviceload ./internal/dataflow ./internal/valueflow`, pełne
  `go test ./...`, `go vet ./...`, build i `git diff --check` — PASS.
  Petclinic: 29,123 klas, 29 load sites, 57 consumption/dispatch (39 linked),
  4 model edges, reachability `2/15/9/15/6`, bez zmiany względem poprzedniego
  pomiaru. WebGoat: 40,569 klas, 48 load sites, 115 consumption/dispatch
  (90 linked), po czym OSV offline przerwało analizę z brakującym kluczem
  `b1e63ee08008e9916ecd8fb4d3495aacb7c0edffdb844a556869cb679b822a55`;
  reachability WebGoat nie jest potwierdzona.
- **Pozostało:** marker formalnego parametru jest substytuowany wyłącznie
  przez konkretny direct-call site; nieznane argumenty pozostają gapem.
  Wirtualne helper returns, parametry zewnętrznych API, stream callback flows
  oraz pozostałe ograniczenia S2 nie zmieniają statusu. Następny krok pozostaje
  `SERVICE_LOADING:S2`.

#### SERVICE_LOADING:S2 — liniowy terminal Stream API (2026-10-26)

- **Zakres:** weryfikacja `ServiceLoader.stream().findFirst().get().get()`;
  bez callbacków i transformacji streama.
- **Zrealizowane:** analizator rozróżnia stream bez operacji terminalnej,
  wsparty łańcuch z powiązanym `Provider.get()` i nierozpoznany terminal.
  Stream bez terminala pozostaje lazy; prosty łańcuch daje krawędź do providera
  dokładnie przy `Provider.get()` oraz późniejszy descriptor-exact dispatch.
  Terminal jest wspierany tylko wtedy, gdy provenance wskazuje dokładną
  instrukcję `stream()` i ten sam load site; współdzielenie jednego loadera
  przez osobne streamy nie łączy ich przypadkowo.
  Terminal ze ścieżką callback `map(provider -> provider.get())` pozostaje
  `Incomplete`; sam `Provider.get()` w callbacku nie jest łączony z load site.
- **Weryfikacja:** skompilowany fixture sprawdza dodatnią ścieżkę i providerowe
  krawędzie, lazy stream bez materializacji oraz lukę callbacku bez zgadywanych
  krawędzi; test jednostkowy odrzuca też powiązanie terminala z innym PC
  `stream()`. `go test ./internal/serviceload ./internal/dataflow`, pełne
  `go test ./...`, `go vet ./...`, `go build -o /tmp/ariadne-s2-stream
  ./cmd/ariadne` oraz `git diff --check` — PASS. Petclinic: 29,123 klas,
  29 load sites, 57 consumption/dispatch (39 linked), 4 model edges,
  reachability `2/15/9/15/6`, bez zmiany. WebGoat: 40,569 klas, 48 load sites,
  115 consumption/dispatch (90 linked); późniejszy etap OSV przerwał brak
  wpisu `b1e63ee08008e9916ecd8fb4d3495aacb7c0edffdb844a556869cb679b822a55`,
  więc reachability WebGoat nie jest potwierdzona.
- **Ograniczenia:** wsparty jest tylko liniowy przepływ `findFirst` → Optional
  `get` → `Provider.get`; `map`/`filter`/inne callbacki i terminale bez
  rozpoznanego `Provider.get` pozostają lukami. S2 nadal trwa, JPMS pozostaje
  osobnym S3.

#### SERVICE_LOADING:S2 — kontynuacja: lazy materialization z lokalnym provenance (2026-10-26)

- **Zakres zatwierdzony:** w ramach S2 powiązać wybrane lokalne użycia
  iteratora/stream i `ServiceLoader.Provider.get()` z konkretnym
  `ServiceLoader.load()` bez zgadywania provider identity; materializować
  provider dopiero w odpowiedniej operacji pobrania.
- **Zrealizowane:** `dataflow` zachowuje ograniczony łańcuch odbiorców dla
  zwracających referencję wywołań JDK `ServiceLoader`, `Iterator`,
  `Optional`, `Spliterator` i `java.util.stream.*`. `FindConsumptionSites`
  rozwiązuje wybrane punkty do dokładnego load site i interfejsu usługi.
  Kontynuacja S2 wykorzystuje `dataflow.ResolveCallResultOneHop` dla
  dokładnych wyników `invokestatic`/`invokespecial`: statyczny helper
  zwracający `ServiceLoader.load(...)` może przenieść pochodzenie do metody
  konsumenta, z zachowaniem PC factory call-site. Pola nie były obsługiwane
  w tym pomiarze; późniejszy przyrost poniżej dodaje ograniczoną obsługę
  streszczeń D3. Wirtualne wyniki, parametry i helpery o nierozstrzygalnym
  wyniku nie tworzą provider edges; ich zużycie pozostaje jawnie niepełne.
  `Iterator.next()` i `Provider.get()` dodają krawędzie z własnego PC do
  publicznych konstruktorów providerów i istniejących `<clinit>`;
  `ServiceLoader.load()` oraz `Iterator.hasNext()` nie tworzą krawędzi
  konstruktora. Gdy późniejszy receiver metody usługi nadal ma ten sam
  lokalny ślad pochodzenia, evaluator dodaje descriptor-exact krawędź do
  implementacji z rejestracji, po sprawdzeniu, że provider jest znanym
  podtypem usługi. Existing CHA dispatch pozostaje szerszym
  overapproximation. Nierozpoznane receiver flows i stream pipelines
  pozostają jawnie `Incomplete`; provider targets nie są zgadywane.
- **Zmienione pliki:** `internal/valueflow/value.go`,
  `internal/dataflow/step.go`, `callresults.go` i testy,
  `internal/serviceload/consumptions.go`,
  `inventory.go`, `model.go`, testy provenance i grafu, `cmd/ariadne/main.go`,
  szablon raportu, `internal/report/report.go` i testy renderowania, `README.md`,
  `docs/SUPPORT.MD` oraz ten rejestr. Potwierdzenie ścieżki na plikach
  classfile dodaje `internal/serviceload/fixture_integration_test.go` oraz
  `testdata/serviceloader/{src,classes}`: źródła Java i gotowe artefakty
  Java 8 (classfile major 52), w tym `META-INF/services`.
- **Weryfikacja:** `GOCACHE=/tmp/ariadne-gocache go test ./...` — PASS;
  `GOCACHE=/tmp/ariadne-gocache go vet ./...` — PASS; `git diff --check`
  — PASS; `GOCACHE=/tmp/ariadne-gocache go build -o
  /tmp/ariadne-s2-final ./cmd/ariadne` — PASS. Testy obejmują dokładne
  powiązanie call-trace z load-site, niezgodny PC jako jawny gap,
  `hasNext()` bez materializacji, `next()`/`Provider.get()` z konstruktorami
  i `<clinit>`, descriptor-exact dispatch, filtrację po ownerze usługi,
  brak warningu dla identycznej sygnatury bez provenance oraz render tabeli.
  Test `TestS2_linkedNextReachesProviderConstructorAndCHAServiceMethod`
  sprawdza pełny modelowy graf `load → iterator → next → constructor →
  provider.decode()V`, z dodatkowym overloadem jako negatywnym przypadkiem.
  Nowy `TestCompiledServiceLoaderFixture` przechodzi przez prawdziwe
  classfile’y: `hierarchy.Build` → `FindLoadSites` → `Inventory` →
  `FindConsumptionSites` → `analysis.Run`; potwierdza przepływ
  `load → createLoader() → iterator → next → decode`, krawędzie do
  konstruktora oraz `decode()Ljava/lang/String;`, brak materializacji dla
  `loadWithoutConsuming()` i jawny gap bez celu dla brakującego providera.
  `runWithUnknownLoader()` zwraca jawny `Incomplete` i nie dodaje krawędzi
  modelowej. `ResolveCallResultOneHop` zachowuje bezpośredni wynik
  helpera bez rekurencyjnego rozwiązywania jego zwróconego `ServiceLoader`;
  łańcuch jest ograniczany przez limit call-trace.
  Artefakty wygenerowano z dołączonych źródeł przez Eclipse Java Compiler
  3.17.0; test Go czyta gotowe `.class` i nie wymaga kompilatora Java.
  Finalne przebiegi `-offline -profile spring` po tej zmianie: petclinic
  29,123 klas / 29 load sites (13 resolved) / 55 consumption/dispatch sites
  (39 linked) /
  4 model edges / 3-62-26-51-6 statusy; WebGoat 40,569 klas / 48 load sites
  (31 resolved) / 111 consumption/dispatch sites (87 linked) / 75 model
  edges / 42-44-18-56-0 statusy. Liczba nowych krawędzi modelu pozostała
  bez zmian na obu aplikacjach; dodatkowe wykryte miejsca nie są połączone
  z load-site. Oba wyjściowe
  `results.json` są bajtowo identyczne z `out/petclinic/results.json` i
  `out/WebGoat/results.json`. Nowa tabela "ServiceLoader consumption and
  dispatch sites" jest obecna w obu raportach.
- **Pozostało w S2:** helpery są śledzone tylko przez dokładne
  `invokestatic`/`invokespecial` wyniki do limitu call-trace; poza nowym
  podzbiorem streszczeń D3 inne pola, parametry i virtual helper returns
  pozostają nierozstrzygnięte. Stream
  callback/pipeline flows są niepełne;
  weryfikacja widoczności przez explicit/system/installed loaderów pozostaje
  gapem. Dispatch dodany przez model jest dokładny dla wykrytej rejestracji,
  ale bazowy CHA nadal dodaje szersze cele. Test pełnej ścieżki używa
  kontrolowanego classpath fixture, a świeże aplikacje kontrolne nie mają
  odpowiadającej, namierzonej ścieżki podatności. Typy nie są przekazywane
  do osobnego RTA, bo RTA nie istnieje w obecnym kodzie. JPMS S3 pozostaje
  osobne.

#### SERVICE_LOADING:S2 — weryfikacja przed commitem przyrostów pól/helperów/stream (2026-11-05)

- **Zakres:** sama weryfikacja niezacommitowanych przyrostów 2026-10-26
  (pola D3, final instance field i cykle, podstawienie zwracanego
  parametru, liniowy terminal Stream API); bez zmian w kodzie.
- **Weryfikacja:** `go vet ./...`, `go test ./...`, `git diff --check`,
  `make build` — PASS. `bin/ariadne -offline -profile spring` na petclinic
  dwukrotnie: bajtowo identyczny JSON, 29 load sites (13 resolved),
  57 consumption/dispatch (39 linked), 4 model edges, statusy
  `3/62/26/51/6`. WebGoat tym razem przeszedł do końca (wcześniejszy
  brak wpisu OSV w offline cache nie wystąpił): 48 load sites
  (31 resolved), 115 consumption/dispatch (90 linked), 75 model edges,
  statusy `42/44/18/56/0`. Na obu aplikacjach `results.json` jest bajtowo
  identyczny z buildem z poprzedniego commita (`2e3ab2f`): nowe
  przyrosty rozpoznają więcej miejsc konsumpcji (WebGoat 107→115,
  87→90 linked), ale nie dodają krawędzi modelu ani nie zmieniają
  reachability — fixture pozostaje jedynym dowodem dodatniej ścieżki.

### SERVICE_LOADING:S3 — JPMS i warstwy (wycinek 1: deskryptory modułów i inwentarz `provides`, scenariusz classpath) — faza DONE for declared scope, zarchiwizowana (2026-11-05)

- **Zakres zatwierdzony:** pierwszy wycinek S3 — odczyt `module-info`
  i inwentarz JPMS jako fakty raportowe pod jawnym scenariuszem
  `classpath`; bez scenariusza module path, bez reguły `provider()` jako
  krawędzi, bez warstw. Razem z zamknięciem S2 (archiwum
  `archive/2026-11-05/service_loading/SERVICE_LOADING_S2.md`).
- **Zrealizowane:**
  - `internal/classfile/module.go`: `AccModule`, `ClassFile.IsModule()`,
    `ClassFile.Module *ModuleInfo` — pełne dekodowanie atrybutu `Module`
    (JVMS 4.7.25): nazwa, flagi (`open`), wersja, `requires` (flagi,
    wersja), `exports`/`opens` (z kwalifikowanymi celami), `uses`,
    `provides … with …` w kolejności deklaracji; obcięty atrybut to błąd.
  - `internal/serviceload/modules.go`: `ModuleScenario`
    (`ScenarioClasspath` — jedyny modelowany), `InventoryModules`
    → `ModuleInventory{Descriptors, Provides, Mismatches}`. Deskryptor
    trzymany poza indeksem klas (hierarchia dalej pomija
    `module-info.class`). Explicit: z `module-info.class`, a w jarze
    `Multi-Release: true` także z `META-INF/versions/N/` — z
    `-target-java` najwyższe N ≤ target (reguła G4), bez targetu korzeń,
    a gdy go brak, najniższe N z notatką; bez manifestu MR wpisy
    wersjonowane są ignorowane. Automatic: `Automatic-Module-Name`
    (z łączeniem linii kontynuacji manifestu) albo nazwa z pliku wg
    algorytmu `ModuleFinder` (nielegalna nazwa → "" z powodem).
    `ModuleProviderDecl`: provider znaleziony?, `public static
    provider()`?, także w `META-INF/services` tego jara?
    `DeclarationMismatch`: provider tylko w module-info albo tylko w
    `META-INF/services` jawnego modułu, z opisem efektu zależnego od
    scenariusza. Uszkodzony deskryptor → ostrzeżenie, jar pominięty
    (nie zgadywany jako automatic).
  - CLI: linia `JPMS: scenario classpath …` z licznikami; raport: sekcja
    "JPMS module descriptors" (tabela jawnych modułów, liczba jarów bez
    deskryptora, `provides`, niezgodności) i dopisek w Limitations.
    `results.json` bez zmian formatu.
- **Kryteria S3 a ten wycinek:** "classpath, named module, automatic module
  i niedostępny provider są rozróżnione" — rozróżnione jako *fakty*
  (rodzaj modułu, provider `provides`-only niewidoczny na classpath),
  ale nie jako scenariusz wykonania module path. "Nieznana warstwa nie
  daje pewnego negatywnego wyniku" — nietknięte (warstwy niemodelowane;
  `load(ModuleLayer, Class)` nadal `LoaderModuleLayer` z luką widoczności).
  Klucz modelu z wariantem modułu i snapshot CVE F3 — nietknięte.
- **Weryfikacja:** fixture `testdata/jpms` skompilowany JDK 25 przez
  `java --module jdk.compiler/com.sun.tools.javac.Main` (`-source/-target`,
  bo JDK nie ma `ct.sym` dla `--release`); README fixture opisuje
  regenerację. Testy: `TestParse_moduleInfoFromCompiledFixture`,
  `TestParse_ordinaryClassHasNoModule`,
  `TestParseModuleAttribute_truncatedIsError`,
  `TestInventoryModules_compiledFixture` (explicit/automatic z manifestu
  i z nazwy pliku, MR bez targetu, MR bez manifestu, `provider()`,
  obie strony niezgodności), `TestInventoryModules_multiReleaseTarget`
  (target 8 → automatic z notatką, 11 → explicit z versions/9),
  `TestInventoryModules_classpathSemanticsUnchanged` (`FactoryCodec`
  z samego `provides` nie staje się kandydatem konstruktora; `LegacyCodec`
  z samego `META-INF/services` tak), `TestInventoryModules_malformedDescriptorIsWarning`,
  `TestAutomaticModuleName`, `TestRenderJPMSModuleInventory`.
  `go vet ./...`, `go test ./...`, `git diff --check`, `make build` — PASS.
  `bin/ariadne -offline -profile spring`, po dwa przebiegi: petclinic —
  19 jawnych deskryptorów (4 tylko w `META-INF/versions/`: HikariCP,
  byte-buddy, classgraph, log4j-api), 69 bez deskryptora, 6 `provides`,
  7 niezgodności (m.in. jackson-datatype-jsr310: `provides` z
  `JSR310Module`, a `META-INF/services` z `JavaTimeModule`); WebGoat —
  20 / 107 / 5 / 7. Na obu `results.json` bajtowo identyczny z
  baseline'em sprzed zmiany i między przebiegami; raport petclinic
  stabilny między przebiegami. Petclinic z `-target-java 8`: 15 jawnych
  (4 wersjonowane wypadają), 5 niezgodności; z `-target-java 17`: 19 / 7;
  statusy bez zmian.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz ServiceLoader/JPMS), README
  (ServiceLoader w ograniczeniach), szablon raportu (sekcja + Limitations),
  FAQ.md (krok hierarchii: gdzie są czytane deskryptory), ten rejestr.
- **Ograniczenia:** jary automatyczne tylko liczone w raporcie (notatka
  o wersjonowanym deskryptorze powyżej targetu nie jest pokazywana dla
  nich); słowa kluczowe Javy w nazwie automatycznej nie są sprawdzane;
  `uses` tylko zapisane — brak kontroli `uses` aplikacji wobec jej load
  sites; deklaracje modułu aplikacji zależą od tego, czy przygotowany
  `app-classes.jar` zawiera `module-info.class`.
- **Następny krok (S3 wycinek 2, wymaga zgody):** jawny scenariusz module
  path (np. flaga wybierająca, które jary są na module path), w nim:
  providery z `provides` jawnych modułów zamiast ich `META-INF/services`,
  reguła `provider()` → krawędź do tej metody zamiast konstruktora,
  automatic modules z `META-INF/services`; czytelność (`requires`) i
  eksport pakietu usługi jako warunki widoczności z jawną luką. Zmierzone
  dziś: realne aplikacje petclinic/WebGoat mają 5–6 `provides` i 7
  niezgodności, więc efekt na reachability będzie wymagał aplikacji
  kontrolnej uruchamianej na module path (dziś żadnej nie ma).

### SERVICE_LOADING:S3 — wycinek 2: scenariusz module path (`-jpms-scenario`) — DONE (2026-11-05)

- **Zakres zatwierdzony:** flaga `-jpms-scenario classpath|module-path`
  (wszystkie jary na module path, bez wariantu mieszanego), zbiór
  providerów zależny od scenariusza, reguła `provider()`, kontrola `uses`,
  ostrzeżenie o brakującym `requires`, aplikacja kontrolna uruchamiana w
  obu trybach.
- **Zrealizowane:**
  - `serviceload.ScenarioModulePath`, `ParseScenario`, `ApplyScenario`
    → `ScenarioView{Registrations}`: w module-path rejestracje
    `META-INF/services` jawnych modułów są pomijane, a każda para
    `provides` staje się `ServiceRegistration` z nowym
    `Mechanism = "module-info provides"` i `Module`; automatic modules
    zachowują `META-INF/services` (nazwa modułu dopisana). Ostrzeżenia:
    `requires` (bez `static`, poza `java.*`/`jdk.*`) na moduł spoza
    wejścia, jar automatyczny bez legalnej nazwy.
  - `ServiceRegistration.Mechanism`/`Module` (S0 ustawia
    `META-INF/services`), `ProviderMaterialization.FactoryDesc`:
    `materialize` dla rejestracji `provides` z `public static provider()`
    daje cel fabryki zamiast konstruktora (także dla klasy z prywatnym
    konstruktorem); automatic/`META-INF/services` — nigdy.
  - `ModelEvaluator.SetScenario`: konsumpcja dodaje krawędź `CallStatic`
    do `provider()` albo jak dotąd `<init>` + `<clinit>`; `usesAllowed`
    sprawdza moduł wywołującego dla powiązanych load sites (load site
    w jawnym module bez `uses S` → brak celów, `Incomplete` z powodem
    `ServiceConfigurationError`); dispatch po obiekcie z `provider()` →
    jawna luka (typ zwracany nieśledzony, CHA zostaje). Komunikat "no
    provider" zależy od scenariusza.
  - CLI: `-jpms-scenario` (walidacja), materializacje z
    `scenarioView.Registrations`, scenariusz w logu `JPMS:` i nagłówku
    raportu. Raport: tekst sekcji JPMS zależny od scenariusza, w tabeli
    kandydatów źródło `provides` i cel `provider()`. Format
    `results.json` bez zmian (scenariusz nie trafia do JSON, bo plik jest
    tablicą wyników — odstępstwo od planu, gdzie proponowano pole
    warunkowe; scenariusz jest w raporcie i logu).
- **Weryfikacja:** fixture `testdata/jpms` rozszerzony o moduł
  `fixture.jpms.nouses` (load bez `uses`). Nowe testy:
  `TestScenario_providerSetsDiffer` (classpath: ModuleCodec, LegacyCodec,
  AutoCodec konstruktory, bez FactoryCodec; module-path: ModuleCodec,
  `FactoryCodec#provider()` zamiast prywatnego konstruktora, AutoCodec,
  bez LegacyCodec; dokładny dispatch do `ModuleCodec.encode`, brak do
  FactoryCodec/LegacyCodec, luka `provider()`),
  `TestScenario_missingUsesBlocksProviders`,
  `TestApplyScenario_missingRequiresWarns`, `TestParseScenario`.
  Mutacje kontrolne: wyłączenie kontroli `uses` i pominięcia
  `META-INF/services` jawnych modułów — odpowiednie testy padają.
  `go vet ./...`, `go test ./...`, `git diff --check`, `make build` — PASS.
  Aplikacja kontrolna `scripts/prepare-jpms-demo.sh` (JDK 25, Maven
  online): realny JVM — `-cp` 0 providerów, `--module-path` 1 provider.
  `bin/ariadne -profile generic -roots profile` (online, potem 2× offline
  bajtowo identycznie): `classpath` — CVE-2022-42889 i CVE-2025-48924
  NOT_REACHABLE, 0 krawędzi modelu; `module-path` — obie REACHABLE,
  ścieżka `demo.Main#main → demo.InterpolatingCodec#provider() →
  StringSubstitutor#replace` (1 krawędź modelu). Regresja `-offline
  -profile spring`: domyślny scenariusz petclinic/WebGoat — `results.json`
  bajtowo identyczny z baseline'em; `module-path` — statusy bez zmian
  (`3/62/26/51/6`, `42/44/18/56/0`), krawędzie modelu 4→2 i 75→73
  (providery z `META-INF/services` jawnych modułów odpadają), 39/153
  rejestracji w mocy, po 5 ostrzeżeń o `requires` HikariCP 3.4.5
  (sprawdzone: nie-`static` `requires slf4j.api`/`hibernate.core`/…
  — realny problem tej wersji na module path).
- **Dokumentacja:** README (tabela flag, lista skryptów, układ projektu,
  ograniczenia ServiceLoader/JPMS), FAQ.md (aplikacja kontrolna),
  `docs/SUPPORT.MD`, szablon raportu (nagłówek, sekcja JPMS,
  Limitations), README fixture, ten rejestr.
- **Kryteria S3 a stan:** classpath, named module, automatic module i
  niedostępny provider są teraz rozróżnione także w wykonaniu modelu.
  Nadal otwarte: scenariusz mieszany (część jarów na classpath), pełne
  rozwiązywanie modułów/czytelność i kwalifikowane eksporty, warstwy
  (`ModuleLayer`, "nieznana warstwa nie daje pewnego negatywnego
  wyniku" — dalej tylko luka `LoaderModuleLayer`), wariant modułu w
  kluczu modelu i snapshot CVE F3. S3 pozostaje IN_PROGRESS.
- **Następny krok (wymaga zgody):** albo S3 wycinek 3 (scenariusz
  mieszany przez listę jarów na module path + `load(ModuleLayer, Class)`
  jako jawna luka z nazwaną warstwą), albo przejście do S4 (regresja i
  rozszerzenia stream API) / innego obszaru — efekt wycinka 3 na
  petclinic/WebGoat byłby zerowy (nie są uruchamiane modułowo).

### SERVICE_LOADING:S3 — wycinek 3: scenariusz mieszany i `ModuleLayer.boot()` — DONE; faza zamknięta i zarchiwizowana (2026-11-05)

- **Zakres zatwierdzony:** `-jpms-scenario mixed` + `-jpms-module-path`,
  `load(ModuleLayer.boot(), S)`, nieznana warstwa jako luka bez
  negatywnego wyniku, trzecie i czwarte uruchomienie demo. Zmierzone przed
  startem i przyjęte świadomie: petclinic/WebGoat mają 0 wywołań
  `load(ModuleLayer, …)` i nie startują modułowo, więc efekt na nich zero.
- **Zrealizowane:**
  - `ScenarioMixed`, `ScenarioConfig{Scenario, ModulePathJars, AppJar}`,
    `ParseModulePathJars` (wzorce `filepath.Match` na nazwie pliku, `app`
    = aplikacja). `ApplyScenario(cfg, …)` liczy zbiór jarów na module
    path (`module-path` = wszystkie, `classpath` = żaden), ustawia nowe
    `ModuleDescriptor.OnModulePath`; jary z classpath zachowują
    rejestracje `META-INF/services` bez modułu, ich `module-info` i
    `provides` są ignorowane. Ostrzeżenia: wzorzec bez dopasowania,
    `requires` spełniane tylko przez jar z classpath ("the unnamed
    module, not readable by named modules"), osobno od modułu zupełnie
    nieobecnego. `usesCheck` działa tylko dla wywołujących w jawnych
    modułach na module path. Założenie jawne: moduły z module path są
    rozwiązane jak przy `--add-modules ALL-MODULE-PATH`.
  - `ServiceLoadSite.Layer`/`LayerReason` (`LayerBoot`, `LayerUnknown`):
    argument `ModuleLayer` rozpoznany jako `KindCallResult`
    `java.lang.ModuleLayer#boot()` w tej samej metodzie. Ewaluator
    (`visibleProviders`, `bootLayerOnly`): gdy wszystkie powiązane load
    sites są boot-layer, widoczni są tylko providery z `Module != ""`;
    w `classpath` — żaden z wejścia, z powodem i `Incomplete` (providerów
    JDK nie analizujemy); luka "module-layer loader" znika tylko w tym
    przypadku. Inna warstwa — bez zmian: wszyscy kandydaci + luka.
  - CLI: `-jpms-module-path`, walidacja (mixed wymaga listy, lista tylko
    z mixed, zły wzorzec). Raport: tekst sekcji JPMS dla `mixed` z listą
    jarów na module path i założeniem `ALL-MODULE-PATH`, opis
    `ModuleLayer.boot()`; Limitations zaktualizowane.
- **Weryfikacja:** fixture: nowa klasa `fixture.jpms.app.Layers`
  (`boot()`, `given(ModuleLayer)`), skompilowana jak reszta modułu.
  Testy: `TestScenario_bootLayer` (classpath: zero providerów z wejścia
  + powód; module-path: ModuleCodec, `FactoryCodec#provider()`,
  automatyczny AutoCodec; mixed {app, api, impl}: ModuleCodec tak,
  AutoCodec z classpath nie), `TestScenario_unknownLayerKeepsGap`,
  `TestScenario_mixed` (provides z module path + `META-INF/services`
  z classpath, LegacyCodec z jawnego modułu pominięty, `uses` tylko dla
  module path), `TestApplyScenario_mixedWarnings`,
  `TestJPMSScenarioFlags`. Mutacje: wyłączenie filtra warstwy boot i
  ostrzeżenia o `requires` na classpath — testy padają. `go vet ./...`,
  `go test ./...`, `git diff --check`, `make build` — PASS.
  Demo (`prepare-jpms-demo.sh`, realny JVM 25): mixed poprawny
  (`-p classes:commons-text`, `-cp commons-lang3`) — 1 provider; mixed
  nieuruchamialny (commons-text na classpath) — `FindException: Module
  org.apache.commons.text not found, required by demo.jpms`. ariadne
  offline: `mixed -jpms-module-path 'app,commons-text-*.jar'` — 2×
  REACHABLE, bez ostrzeżeń; `mixed -jpms-module-path app` — ostrzeżenie
  "requires org.apache.commons.text, which is on the class path".
  `classpath`/`module-path` na demo bajtowo jak w wycinku 2. Regresja:
  petclinic/WebGoat domyślnie bajtowo identyczne z baseline'em,
  petclinic `module-path` bajtowo jak w wycinku 2; WebGoat `mixed
  -jpms-module-path 'jackson-*.jar'` — statusy bez zmian, 75 krawędzi.
- **Dokumentacja:** README (flagi `-jpms-scenario mixed`,
  `-jpms-module-path`, ograniczenia), FAQ.md (demo), `docs/SUPPORT.MD`,
  szablon raportu, README fixture, plan fazy, ten rejestr.
- **Kryteria S3 a stan:** classpath / named / automatic / niedostępny
  provider rozróżnione w modelu; nieznana warstwa nie daje negatywnego
  wyniku (test). Otwarte: pełne rozwiązywanie modułów (w tym fakt, że
  main z classpath bez `--add-modules` nie rozwiązuje modułów z module
  path), kwalifikowane eksporty, warstwy inne niż boot, dispatch po
  `provider()`, klucz modelu z wariantem modułu i snapshot CVE F3.
- **Zamknięcie (2026-11-05):** na decyzję użytkownika S3 jest DONE dla
  zadeklarowanego zakresu, a pozostałe punkty to jawne ograniczenia —
  rekord: `archive/2026-11-05/service_loading/SERVICE_LOADING_S3.md`.
- **Następny krok (wymaga zgody), kandydatury: rozpoznawanie jawnych `ClassLoader` (19/29 load sites
  petclinic, 32/48 WebGoat), SERVICE_LOADING:S4, dalsze modele B3.

### SERVICE_LOADING — korekta po S2 (nie nowa faza): rozpoznawanie jawnego `ClassLoader` w `load(Class, ClassLoader)` — DONE for this scope (2026-11-05)

- **Zakres zatwierdzony:** rozpoznać typowe postacie argumentu loadera
  i traktować je jak modelowany classpath, usuwając lukę widoczności;
  nierozpoznane zostają luką. Wymóg S1 "odróżnić classpath, system loader
  i jawnie podany loader" — doprecyzowanie, nie zmiana kontraktu.
- **Zrealizowane:** `dataflow`: `isClassLoaderTraceCall` — dokładne
  `Thread#getContextClassLoader()` i `Class#getClassLoader()` dają
  `KindCallTrace` (jak wcześniej wywołania JDK ServiceLoader/stream), nie
  `Unknown`. `serviceload`: `ServiceLoadSite.LoaderIdentity`/`LoaderDetail`
  (`context`, `system` dla `getSystemClassLoader()` i `null`, `defining`
  dla `X.class.getClassLoader()` lub nieznanej klasy); ewaluator
  (`explicitLoaderRecognized`) pomija lukę "explicit loader", gdy każdy
  powiązany load site z jawnym loaderem ma rozpoznaną tożsamość.
  Raport: tożsamość w kolumnie Loader i opis założenia jednego loadera
  aplikacji. Krawędzie modelu bez zmian (już wcześniej były dodawane jako
  możliwość) — zmienia się wyłącznie luka.
- **Weryfikacja:** fixture `fixture.jpms.app.Loaders` (context, system,
  null, defining, parametr). `TestLoadSites_explicitLoaderIdentity`;
  mutacje (wyłączenie pominięcia luki; wyłączenie śledzenia wywołań
  loadera) — test pada. `go vet ./...`, `go test ./...`, `make build` —
  PASS. Offline `-profile spring`, po dwa przebiegi: petclinic 1/19
  jawnych loaderów rozpoznanych, WebGoat 5/32 (wszystkie
  `X.class.getClassLoader()`: asciidoctor, xnio); pozostałe to parametry
  metod (Jackson `secureGetServiceLoader`, Undertow, Hibernate, log4j,
  validation) i pola klas anonimowych (`ObjectMapper$2`, `Xnio$2`).
  `results.json` bajtowo identyczny z baseline'em i stabilny; statusy bez
  zmian. Ostrzeżenie zbiorcze o luce zostaje (pozostałe miejsca).
- **Dokumentacja:** `docs/SUPPORT.MD`, README (ograniczenia), szablon
  raportu, ten rejestr. FAQ bez zmian (nie opisuje loaderów).
- **Ograniczenia / następny krok:** efekt mały i zmierzony uczciwie.
  Większość realnych luk to loader jako parametr — wymagałoby to
  propagacji argumentów od wywołujących (DATA_FLOW D3 dla parametrów,
  dziś tylko zwrot parametru) albo pól klas anonimowych zapisywanych w
  konstruktorze z parametru. Założenie jednego loadera aplikacji jest
  jawne w raporcie.

### DATA_FLOW:D3 — Przepływ między metodami i streszczenia (wycinek: parametry od wywołujących) — DONE for this shape (2026-11-05)

- **Zakres zatwierdzony:** wartości parametru ze wszystkich wywołań w
  analizowanym wejściu, bez kontekstu, głębokość 3, maks. 16 wywołań,
  tylko metody o statycznie dokładnych wywołaniach; pierwszy odbiorca —
  argumenty `Class` i `ClassLoader` przy `ServiceLoader.load`.
- **Zrealizowane:** `internal/dataflow/params.go`: `CallerIndex`
  (budowany raz, na żądanie; cele kluczowane klasą deklarującą przez
  `ResolveExact`; cele `invokedynamic` jako "ukryci wywołujący"),
  `ResolveParameter` → `ParameterValues{Values, Unknown, Callers}`.
  Metody: statyczne, prywatne, konstruktory albo wirtualne bez nadpisań
  w obie strony (nadtyp spoza wejścia, inny niż `Object`, liczy się jako
  możliwe nadpisanie). Nieznane z powodem: lambda/referencja metody,
  brak wywołań, przekroczone limity, przerwana analiza wywołującego.
  Rekurencja na tym samym parametrze zbiega (nic nie dodaje do sumy).
  `serviceload`: `FindLoadSites` używa `AnalyzeWithParameterRefs`;
  `ServiceLoadSite.Interfaces`/`ParamSource`/`ClassParamSet`/
  `ClassParamIndex`, `Candidates()`; częściowy zbiór zachowuje
  `UnresolvedReason`. Loader z parametru: rozpoznany tylko, gdy wszystkie
  wartości od wywołujących są rozpoznanymi postaciami (nowa tożsamość
  `mixed` przy różnych). **Precyzja (kryterium D3 "dwa call sites nie
  tworzą nieuzasadnionej ścieżki"):** konsumpcja dochodząca do load site
  przez jedno dokładne wywołanie helpera jest zawężana do argumentu tego
  wywołania (`narrowToCallSite`), więc suma bez kontekstu zostaje tylko
  w inwentarzu load site, nie w krawędziach konkretnego wywołującego.
- **Weryfikacja:** fixture `fixture.jpms.app.Params` (zbiór z dwóch
  wywołań, kształt Jacksona `findModules()`→`findAll(null)`→`secure`,
  rekurencja, referencja metody, brak wywołań, metoda nadpisana).
  `TestLoadSites_parametersFromCallers`,
  `TestParameters_noCrossCallerPath`, `TestResolveParameter_limits`.
  Mutacje: wyłączenie zawężenia do wywołania (krawędź Codec z
  `viaRunnable` się pojawia) i ochrony przed nadpisaniem — testy padają.
  `go vet ./...`, `go test ./...`, `make build` — PASS. Offline
  `-profile spring`, po dwa przebiegi (stabilne): petclinic load sites z
  usługą 13→20, powiązane konsumpcje 39→45, krawędzie modelu 4→7, 4
  loadery z parametru rozpoznane; WebGoat 31→38, 90→99, 75→79, 5
  loaderów. Rozwiązane realne miejsca: Jackson
  `ObjectMapper`/`MapperBuilder#secureGetServiceLoader` →
  `databind.Module` (loader `null` z `findModules()`), logback
  `EnvUtil#loadFromServiceLoader` → `Configurator`, JAXB
  `ServiceLoaderUtil#firstByServiceLoader` → `JAXBContextFactory`,
  ehcache `ClassLoading#servicesOfType`, xnio `SaslUtils#getFactories`.
  Statusy CVE bez zmian, `results.json` bajtowo identyczny z baseline'em
  (nowe krawędzie nie leżą na ścieżkach podatnych metod). Czas WebGoat
  13,6 s → ok. 15,0 s.
- **Dokumentacja:** `docs/SUPPORT.MD` (DATA_FLOW i ServiceLoader), README,
  szablon raportu, ten rejestr. FAQ bez zmian.
- **Ograniczenia:** suma bez kontekstu w inwentarzu; zakłada pełny
  zbiór wywołujących w wejściu (refleksja/framework niewidoczne poza
  "brak wywołań"); wywołania wirtualne tylko przy jednym celu; pola
  klas anonimowych zapisywane z parametru konstruktora (np.
  `ObjectMapper$2`) nadal nierozwiązane; brak wersjonowania streszczeń.
- **Następny krok (wymaga zgody):** kolejni odbiorcy `ResolveParameter`
  — REFLECTION:R3 (nazwy klas/metod jako parametry), MethodHandle,
  callbacki; albo pola zapisywane z parametru konstruktora (połączenie
  z istniejącym streszczeniem pól `final`).

### REFLECTION:R3 — wycinek 1 (parametry i fabryki): pomiar, reguła stopu zadziałała — NOT STARTED, zmierzone (2026-11-05)

- **Zakres zatwierdzony:** krok 0 = pomiar; jeśli na petclinic i WebGoat
  zero realnych miejsc daje się zamienić na krawędź, przerwać i zapisać
  pomiar zamiast budować odbiorców (decyzja użytkownika przy
  zatwierdzeniu planu).
- **Metoda:** tymczasowa sonda (tymczasowy test Go, usunięty po użyciu)
  na prawdziwych jarach: dla każdego `Method#invoke` analiza wywołującego
  z `AnalyzeWithParameterRefs`; receiver-parametr rozwiązywany przez
  `dataflow.CallerIndex.ResolveParameter` (D3); receiver będący wynikiem
  helpera, którego `ResolveCallResult` nie rozwiązuje, sprawdzany pod
  kątem stałych argumentów w miejscu wywołania
  (`ResolveCallArgumentAtCall`).
- **Wynik:**
  - petclinic: 619 miejsc `invoke`; 72 z receiverem-parametrem — **0**
    z jakimkolwiek rozpoznanym `Method` od wywołujących; 48
    nierozwiązanych wyników helperów, 4 ze stałymi argumentami.
  - WebGoat: 628; 75 parametrów — **0**; 52 helpery, 6 ze stałymi.
  - Kandydaci "fabryka": Tomcat `IntrospectionUtils#findMethod(
    obj.getClass(), "setExecutor", …)` i XStream
    `SerializationMethodInvoker#getMethod(type, "readResolve", …)` —
    klasa znana dopiero w czasie działania (nieznana / parametr);
    Hibernate `ReflectHelper#getMethod(Stream.class, "takeWhile"|
    "dropWhile", …)` — jedyne z pełnymi stałymi, ale cel
    `java.util.stream.Stream` jest poza analizowanym wejściem (JDK), więc
    krawędź i tak nie powstałaby.
  - Łącznie **0** miejsc możliwych do zamiany na krawędź na obu
    aplikacjach → reguła stopu. Zgodne z wcześniejszą oceną (sesja
    2026-10-23), teraz potwierdzone pomiarem, a nie przewidywaniem.
- **Zmiany w kodzie:** brak (sonda usunięta). Dokumentacja wsparcia bez
  zmian — nic nowego nie jest wspierane.
- **Następny krok:** R3 wycinek 1 ma sens dopiero z inną aplikacją
  kontrolną, która przekazuje `Method`/nazwy przez helpery z klasami z
  wejścia, albo w połączeniu z wartościami klas znanymi w czasie
  działania (np. `obj.getClass()` przy znanym typie alokacji — RTA,
  GRAPH_PRECISION). Inne kandydatury: pola zapisywane z parametru
  konstruktora (D3), SERVICE_LOADING:S4, dalsze modele B3, FRAMEWORKS:F3.

### DATA_FLOW:D3 — wycinek: pola `final` zapisywane z parametru konstruktora — DONE for this shape (2026-11-05)

- **Zakres zatwierdzony:** rozszerzyć streszczenie pól `final`
  zapisywanych przez `this` w konstruktorze o wartość będącą parametrem
  konstruktora, rozwiązywaną przez wywołujących (`ResolveParameter`,
  limity jak w poprzednim wycinku); odbiorca: argumenty `Class`/
  `ClassLoader` przy `ServiceLoader.load`; reguła stopu przy zerze
  rozwiązanych miejsc (nie zadziałała).
- **Zrealizowane:** `dataflow.CallerIndex.ResolveFinalField` (pole
  `final`, niestatyczne; każdy zapis `this.f = …` w `<init>` analizowany
  z `AnalyzeWithParameterRefs`; parametr → wywołujący tego konstruktora;
  suma po wszystkich `new`, bez rozróżniania instancji — bezpieczne
  dzięki `final` i zapisowi przez `this`). `collect` przyjmuje jawną
  metodę-właściciela parametru. `serviceload`: `resolveFieldArg` —
  najpierw istniejące `ResolveField` (pole statyczne / `final` ze stałą,
  co przy okazji rozwiązuje też stałe pola statyczne `Class`), potem
  `ResolveFinalField`; to samo dla loadera. Raport: źródło "final field
  set from a constructor parameter…".
- **Weryfikacja:** fixture `fixture.jpms.app.Captures` (anonimowy
  `PrivilegedAction` w kształcie Jacksona, `Box` z dwoma `new`,
  `Mutable` z polem nie-`final`).
  `TestLoadSites_finalFieldsFromConstructorParameters`; mutacje
  (pominięcie warunku `final`; wyłączenie odbiorcy) — test pada.
  `go vet ./...`, `go test ./...`, `make build` — PASS. Offline
  `-profile spring`, po dwa stabilne przebiegi: petclinic load sites z
  usługą 20→25, powiązane konsumpcje 45→46; WebGoat 38→43, 99→100.
  Rozwiązane: Jackson `ObjectMapper$2#run`/`MapperBuilder$1#run`
  (`databind.Module`; loader `null` dla `MapperBuilder$1`, dla
  `ObjectMapper$2` łańcuch przekracza 3 poziomy), Hibernate Validator
  `GetInstancesFromServiceLoader` (`ValueExtractor`,
  `ConstraintValidator`), xnio `Xnio$2` (`XnioProvider`, loader
  `Undertow.class.getClassLoader()`). Krawędzie modelu bez zmian
  (7 / 79): te loadery wracają przez `AccessController.doPrivileged`,
  którego wynik nie jest modelowany. `results.json` bajtowo
  identyczny, czas bez zmian.
- **Dokumentacja:** `docs/SUPPORT.MD`, README, szablon raportu, ten
  rejestr.
- **Ograniczenia:** suma po instancjach; tylko odbiorca ServiceLoader
  (inni użytkownicy `ResolveField` bez zmian); pola nie-`final` i
  zapisy poza konstruktorem nieznane.
- **Następny krok (wymaga zgody):** model
  `AccessController.doPrivileged(PrivilegedAction|
  PrivilegedExceptionAction)` → wynik `run()` (ASYNC_CALLBACKS lub
  model B3) — jedyny brakujący element, żeby te rozwiązane miejsca
  dały krawędzie; albo SERVICE_LOADING:S4.

### DATA_FLOW:D3 — streszczenia zależności (`internal/flowsum`) — DONE (2026-11-07)

- **Zakres zatwierdzony (2026-11-07):** zależności parametr/`this`/pole →
  wynik/pole/argument wywołania, klucz kontekstu przez podstawienie per
  miejsce wywołania, jawna polityka konserwatywna dla nieznanych metod z
  pochodzeniem poszerzenia, rekurencja z limitem, wersjonowanie w
  snapshotcie, mały katalog JDK pod D4, pierwszy odbiorca w deserializacji
  (raport + v2). Pola instancyjne per obiekt (points-to) poza zakresem.
  Gałąź `data-flow-d3-summaries` od `main` (ca0a761).
- **Zrealizowane:**
  1. `internal/flowsum/dep.go`: domena — posortowane zbiory internowanych
     źródeł (`param:i`, `this`, `field:Owner.name`, `call:Owner#name desc`,
     wewnętrzne `new@pc`), `substitute` (param → argument miejsca, this →
     odbiorca), limit 32 źródeł `call:` → `call:*` (monotoniczny top).
     `SummaryVersion = "d3-1"`.
  2. `method.go`: interpreter na CFG i `LocalOp` z D1 — stos/zmienne,
     `dup*` wg kategorii, tablice i obiekty (alokacja wchłania argumenty
     konstruktora/mutatora/`putfield`; `getfield` zależy od pola i obiektu),
     krawędzie wyjątków (zmienne z wejścia ∪ wyjścia bloku), bez zależności
     sterowania; streszczenie: `Return`, `Fields`, `Calls[pc]`, `IntoThis`,
     `IntoParams`, `Widened`, `Incomplete`.
  3. `engine.go`: odkrywanie od korzeni (budżet 20 000 metod), cel
     wywołania: `invokestatic`/`invokespecial` dokładnie, virtual/interface
     przy jednej implementacji CHA w wejściu; SCC (Tarjan, iteracyjnie,
     najpierw wywoływane), w SCC kolejka z najmniejszą metodą najpierw, limit
     30 ewaluacji na członka → `Incomplete`; niekompletność wywoływanego
     przechodzi na wywołujących. `catalog.go`: StringBuilder/StringBuffer,
     String, Base64, opakowania strumieni/readerów i ich `read*`, Optional,
     Objects, `parse*`/`valueOf`, URLDecoder, `Object#<init>`,
     `setObjectInputFilter`/`createFilter`; invokedynamic → argumenty.
  4. Odbiorca: `deserialize.AttachArgumentFlow` (Java: odbiorca
     `readObject`; XStream: argument 0) → `Site.DataFrom*`; raport (kolumna
     „Data read depends on”), v2 `entries[].data_depends_on`/
     `data_flow_incomplete`/`data_flow_widened`; snapshot `algorithms`
     `flowsum.summaries`; log z liczbami i czasem.
  5. Fixture `testdata/flowsum` (`scripts/build-flowsum-fixture.sh`, `ext.Lib`
     kompilowane, ale poza JAR-em = metoda zewnętrzna); aplikacja kontrolna
     `scripts/prepare-dataflow-demo.sh` (sześć kształtów, uruchamiana na
     JVM).
- **Weryfikacja:** `flowsum_test.go` — tożsamość, transformacja z katalogu
  (bez nowego źródła), stała (pusto), konkatenacja (indy), konstruktor +
  getter, CHA 1 cel / 2 cele (poszerzenie z powodem), opakowanie strumienia,
  mutacja przez metodę użytkownika, rekurencja (zbieżna), handler wyjątku,
  pole statyczne, metoda zewnętrzna, **dwa miejsca wywołania helpera bez
  mieszania**, budżet (poszerzenie „summary budget”), limit rekurencji
  (`Incomplete`). `argflow_test.go`: wejścia fixture'u wyroczni zależą od
  parametru `InputStream`. Demo: `fromBytes`/`fromBase64` → parametr 0,
  `fromField` → this + pole, `fromConstant` → nic, `fromSource` → oba
  parametry + wynik `Source#bytes` (2 implementacje), `twoSites` → tylko
  parametr 1. Prawdziwe dane (offline, main ca0a761 vs gałąź): werdykty
  petclinic 3/71/16/57/6, WebGoat 42/51/10/61/0 bez zmian; v2 identyczny poza
  nowymi polami; dwa przebiegi bajtowo identyczne; petclinic 1252 metod, 1317
  ewaluacji, 1,8 s, 119/130 wejść zależnych od parametru; WebGoat 2420 metod,
  3484 ewaluacje, 6,7 s, 134/143 (`VulnerableComponentsLesson#completed` →
  parametr 0 `String`). Wydajność: pierwsza wersja 47 s na WebGoat
  (olbrzymia metoda JRuby `Sprintf#rubySprintfToBuffer`, 2605 instrukcji, w
  SCC 153 metod) → slice zamiast mapy zmiennych, brak alokacji przy
  niezmienionym scalaniu, internowanie źródeł, SCC i kolejność „najmniejsza
  najpierw” → 6,7 s. `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz Data flow), README, FAQ,
  `docs/RESULTS_CONTRACT.MD`, plan DATA_FLOW, ten rejestr.
- **Ograniczenia:** to nie taint (skąd biorą się parametry — D4); brak
  zależności sterowania; pola per pole, nie per obiekt; katalog JDK mały,
  reszta konserwatywna (nadmiar, nigdy brak); odbiorcą są tylko wejścia
  deserializacji; streszczenia nie zasilają jeszcze `internal/dataflow`
  (wartości) ani innych modeli.
- **Następny krok (wymaga zgody):** DATA_FLOW:D4 — pierwszy taint (źródła:
  parametry handlerów Spring MVC/servletów, `getParameter`/`getInputStream`;
  propagacja przez streszczenia D3 w górę po wywołujących; ujście: wejścia
  deserializacji) — domyka resztę DESERIALIZATION:Z3.

### DATA_FLOW:D4 — pierwsza analiza taint (pilotaż: żądanie → wejścia deserializacji) — DONE (2026-11-07)

- **Zakres zatwierdzony (2026-11-07):** `internal/taint` z wersjonowanym
  katalogiem źródeł i zabezpieczeń; jeden pilotaż: dane żądania HTTP →
  dane czytane przez wejścia deserializacji (Java: odbiorca `readObject`;
  XStream: argument 0 `fromXML`); propagacja wstecz po streszczeniach D3
  per miejsce wywołania; ścieżka danych; stany `flow_found`/
  `no_flow_in_model`/`undetermined`/`not_assessed`; zabezpieczenia
  specyficzne dla modelu (polityka Z2 skuteczna, reszta niezwiązana);
  osobna ocena w v2 i raporcie; **flaga `-taint`, domyślnie wyłączona**
  (decyzja użytkownika: opt-in, zgodnie z D5). Gałąź `data-flow-d4-taint`
  od `main` (d683dd7 + merge PR #8).
- **Zrealizowane:**
  1. `internal/taint/catalog.go` (`CatalogVersion = "taint-1"`): źródła —
     parametry handlerów Spring MVC z `@RequestParam`/`@RequestBody`/
     `@RequestHeader`/`@PathVariable`/`@CookieValue`/`@RequestPart`/
     `@MatrixVariable` (`src:spring-mvc-param`), nieadnotowane
     `String`/`String[]`/`byte[]`/`InputStream`/`Reader`/`MultipartFile`
     metody z adnotacją mapowania (`src:spring-mvc-implicit`), JAX-RS
     `@QueryParam`/`@FormParam`/`@PathParam`/`@HeaderParam`/`@CookieParam`/
     `@MatrixParam`/`@BeanParam` (`src:jaxrs-param`), wyniki akcesorów
     `ServletRequest`/`HttpServletRequest`/wrapperów (`javax`/`jakarta`,
     `src:servlet-request`); katalog „zabezpieczeń” niezwiązanych
     (zamiana napisów, escaping/encoding Spring/commons/OWASP/Jsoup) i
     sprawdzeń sterujących (`isEmpty`/`isBlank`/`matches`/`startsWith`/
     `contains`/`equals`, `StringUtils`, `Pattern.matches`).
  2. `internal/taint/taint.go`: indeks wywołań klas aplikacji (po nazwie +
     deskryptorze; dokładnie przez `ResolveExact`, virtual/interface gdy
     klasa metody jest podtypem nazwanego właściciela), `Closure` (metody
     z ujściami + wywołujący do głębokości 8 → korzenie streszczeń),
     `Analyze`: wyszukiwanie wstecz — `param:i`/`this` → źródło albo każde
     miejsce wywołania (argument/odbiorca tego miejsca w terminach
     wywołującego), `call:` → źródło servletowe albo luka, `field:`/
     `call:*` → luka, brak wywołujących → wejście spoza katalogu, metoda z
     invokedynamic → luka; limity: głębokość 8, 2000 węzłów, 5 ścieżek
     (licznik dokładny), 20 luk (licznik dokładny). Zabezpieczenia na
     ścieżce: wywołania z katalogu, których odbiorca/argument zależy od
     wejścia danej metody (scalane per metoda, lista PC).
  3. `internal/flowsum`: nowa opcja `Options.Opaque` — wskazane wywołania
     (akcesory żądania) nigdy nie są rozwijane w ciało jedynej
     implementacji, więc ich wynik zostaje źródłem `call:`; bez opcji
     zachowanie D3 niezmienione (`SummaryVersion` bez zmian).
  4. `internal/deserialize/taint.go`: ujścia (`sink:deser-java-stream`,
     `sink:xstream-fromxml`), polityka Z2 → stan zabezpieczenia (`evaluated`
     odrzucająca wszystkie kandydatki lub `default_allowlist` → `effective`;
     częściowo → `partial`; `unknown`/`version_unknown` → `unknown`),
     `TaintRoots`, `AttachTaint` (wejście biblioteczne bez wywołujących z
     aplikacji → `not_assessed`), `Site.Taint`.
  5. CLI: `-taint`; korzenie streszczeń = domknięcie wywołujących (jeden
     przebieg `Summarize`), log ze stanami; snapshot `algorithms`
     `taint.catalog`, konfiguracja `taint`. v2: `deserialization.
     taint_catalog`, `entries[].taint`, `summary.entries_by_taint` (tylko
     z `-taint`). Raport: podsekcja „Request data reaching the entries
     (taint)” (bez `-taint`: jawne „Not assessed”; wejścia `not_assessed`
     tylko liczone).
  6. Fixture `testdata/taint` (`scripts/build-taint-fixture.sh`) i
     aplikacja kontrolna `scripts/prepare-taint-demo.sh` (stuby adnotacji
     Spring, `HtmlUtils`, `ServletRequest` o prawdziwych nazwach — bez
     zależności, offline).
- **Weryfikacja:** `internal/taint/taint_test.go` (pełny potok jak w CLI na
  fixture): dane żądania przez dwa poziomy wywołań (`flow_found`, guard
  `none`, ścieżka z PC), stała (`no_flow_in_model`), helper wywołany z
  danymi żądania i z literałem — czytany tylko literał (`no_flow_in_model`,
  brak mieszania miejsc), filtr `createFilter("java.lang.String;!*")`
  (`effective`), `isEmpty` + `HtmlUtils.htmlEscape` (`unrelated`, żadne
  zabezpieczenie nieskuteczne), interfejs z dwiema implementacjami
  (`undetermined`), `ServletRequest#getParameter` przy jednej
  implementacji w wejściu (`src:servlet-request`), parametr niejawny,
  argument `main` (`no_flow_in_model` + `outside_inputs`), pole
  (`undetermined`), wywołujący przez lambdę (luka), limit głębokości
  (`undetermined`) i ta sama ścieżka z limitem 12 (`flow_found`, 11
  kroków), wspólny helper z wywołującym z żądania i z literałem (jedna
  ścieżka, bez luk). `resultsv2`: taint tylko z `-taint`. `report`:
  renderowanie z i bez `-taint`, wejścia biblioteczne tylko liczone.
  Demo na OpenJDK 25: aplikacja działa (filtr odrzuca `Payload` w czasie
  wykonania), ariadne: 9/9 wejść w oczekiwanym stanie (`flow_found` 5,
  `no_flow_in_model` 3, `undetermined` 1). Prawdziwe dane (offline, binarka
  `main` vs gałąź): werdykty WebGoat 42/51/10/61/0 i petclinic 3/71/16/57/6
  bez zmian; bez `-taint` v2 identyczny z `main` poza `analysis_id`/wersją;
  z `-taint` identyczny z przebiegiem bez niej poza polami taint; dwa
  przebiegi z `-taint` bajtowo identyczne. WebGoat: 143 wejścia —
  `VulnerableComponentsLesson#completed` i wewnętrzne `XStream#fromXML`
  `flow_found` z `@RequestParam payload` (guard `unrelated`: 5×
  `String#replace`, `StringUtils#isEmpty`; polityka `unrestricted`), 141
  bibliotecznych `not_assessed`; petclinic: 130 `not_assessed` (brak wejść
  w aplikacji). Lekcja insecure-deserialization WebGoat nie jest w tym
  wejściu (brak w `app-classes.jar`), więc przypadek Java pokrywa demo.
  Czas: WebGoat 22,6 s z i bez `-taint`, petclinic 12,5 s / 12,4 s
  (streszczenia 6,5 s / 1,7 s, ta sama liczba metod — domknięcie
  wywołujących mieści się w już streszczanych). `go test ./...`,
  `go vet ./...` — PASS.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersze Data flow and taint,
  Deserialization), README (sekcja D4, flaga, demo, ograniczenia),
  `FAQ.md` (nowe pytanie o `-taint`), `docs/RESULTS_CONTRACT.MD`, szablon
  raportu, plany DATA_FLOW/DESERIALIZATION, `DATA_FLOW.MD`, ten rejestr.
  `FAQ_PL.MD` celowo nietknięty (stała decyzja użytkownika).
- **Ograniczenia:** jedna rodzina ujść (wejścia deserializacji); mały
  katalog źródeł; wywołujący szukani tylko w klasach aplikacji; przepływy
  przez pola (statyczne i instancyjne), wywołujący przez lambdy/referencje
  metod i niemodelowane wywołania → `undetermined`; brak zależności
  sterowania; ścieżka na poziomie metod (bez kroków transformacji wewnątrz
  metody); źródła z Spring rozpoznawane po adnotacjach (bez `@ModelAttribute`
  i konwerterów); `no_flow_in_model` względne wobec katalogu; wyniki
  łańcuchów (Z3a) nadal mają `flow: not_analyzed`.
- **Wydanie i archiwum:** zmergowane jako PR #9 (`003e9d3`, 2026-11-07);
  tekst fazy przeniesiony do
  [archive/2026-11-07/data_flow/DATA_FLOW_D4.md](../archive/2026-11-07/data_flow/DATA_FLOW_D4.md).
  Kryterium „kroki transformacji” w ścieżce spełnione częściowo (poziom
  metod) — odnotowane tam i w ograniczeniach.
- **Następny krok (wymaga zgody):** DATA_FLOW:D5 (pomiar: błędne przepływy,
  pominięcia, `undetermined`, czas/pamięć na konkretnej bibliotece i
  pilotażu; kwalifikacja przed ewentualnym włączeniem domyślnym) albo
  DESERIALIZATION:Z3 — powiązanie wyniku taint wejścia z wynikami łańcuchów
  (`flow` per wejście zamiast `not_analyzed`).

### DATA_FLOW:D5 — pomiar i kwalifikacja pilotażu taint — DONE (2026-11-07)

- **Zakres zatwierdzony (2026-11-07):** fixtures dla ośmiu kształtów z planu,
  wyrocznia prawdziwego JVM porównywana bez JDK, pomiar (pominięcia,
  nadmiar, `undetermined`, czas, pamięć) dla konkretnego pilotażu i wersji
  bibliotek, B4, publikacja zakresu; **`-taint` pozostaje opt-in** (decyzja
  użytkownika). Gałąź `data-flow-d5-qualification` od `archive-deserialization`
  (PR #12).
- **Zrealizowane:**
  1. `testdata/taint/src/demo/Shapes.java` (+ `Loader`, `Batch`): aliasing
     (druga referencja do buildera), tablice, zmienne lokalne do handlera
     wyjątku, obiekt wyjątku (`getMessage`), dwa strumienie w jednej
     metodzie, overload `load(String)`/`load(byte[])`, callback (lambda),
     klasa bez źródeł. Fixture budowany dla Javy 17
     (`scripts/build-taint-fixture.sh`).
  2. Wyrocznia: `testdata/taint/oracle/src` (`oracle.Canary`,
     `oracle.TaintOracle` — wywołuje każdy handler ze stubem `@PostMapping`
     z kanarkiem jako danymi żądania; fabryka `ObjectInputFilter` procesu
     zapisuje wejścia, które go odczytały: `read`/`rejected`),
     `scripts/taint-oracle.sh`, `observed.tsv`, `oracle.jar`, README.
  3. `internal/taint`: `TestTaintOracle` (pominięcia = 0, niepotwierdzone
     `flow_found` = 0, lista zaobserwowanych `undetermined` przypięta),
     `TestShapes`; `Analyzer.Stats` (wyszukiwania, odwiedzone wejścia,
     zatrzymania limitów) i czas wyszukiwania w logu `-taint`.
  4. `docs/TAINT_QUALIFICATION.MD`: wyniki wyroczni, kształty, XStream na
     przykładach niezależnie sprawdzonych, tabela pomiarów, co jest i czego
     nie jest zakwalifikowane, decyzja opt-in.
- **Weryfikacja:** `scripts/taint-oracle.sh` na OpenJDK 25 — 14 wejść
  odczytało kanarka (1 odrzucony przez własny filtr). Model na 21 wejściach:
  `flow_found`+odczyt 10, `undetermined`+odczyt 4 (`Codec#readDeep` limit
  głębokości, `Codec#readOther` pole, `Loader#inCallback` lambda,
  `Shapes#exceptionMessage` `Throwable#getMessage`), `no_flow_in_model`+brak
  odczytu 6, `undetermined`+brak 1; pominięć 0, niepotwierdzonych 0.
  Mutacja (źródła-parametry wyłączone) → `TestTaintOracle` zgłasza 5
  pominięć. Pomiar (offline, dwa przebiegi): WebGoat 23,9–24,0 s / 24,3–29,5 s,
  RSS 1736–1769 / 1780–1784 MB; petclinic 12,2–12,3 / 12,3–13,3 s, RSS
  1348–1399 / 1367–1368 MB (różnice w granicach szumu); wyszukiwanie: taint
  demo 66 µs, chains demo 0,3–0,6 ms, WebGoat 0,3 ms (2 wyszukiwania, 81
  wejść), 0 zatrzymań limitów. Werdykty bez zmian. `go test ./...`,
  `go vet ./...` — PASS.
- **Dokumentacja:** `docs/TAINT_QUALIFICATION.MD` (nowy), `docs/SUPPORT.MD`
  (wiersz Data flow and taint), README (sekcja D4, lista skryptów), `FAQ.md`,
  plan DATA_FLOW, ten rejestr. `FAQ_PL.MD` nietknięty.
- **Ograniczenia:** wyrocznia tylko dla serializacji Java (XStream nie
  używa `ObjectInputFilter` — sprawdzony na przykładach); fixture mały i
  pisany pod znane kształty — żadnego globalnego procentu; luki D4 bez
  zmian (pola, lambdy, wywołujący spoza aplikacji, niemodelowane wywołania
  jak `getMessage`, brak zależności sterowania); heap/context sensitivity i
  wątki — osobne etapy.
- **Wydanie i archiwum:** zmergowane jako PR #13 (2026-11-07); cały obszar
  (D0–D5) zarchiwizowany w
  [archive/2026-11-07/data_flow/](../archive/2026-11-07/data_flow/COMPLETION.md).
- **Następny krok:** obszar zamknięty. Potem inny obszar (CVE:F8, ROADMAP:B3, GRAPH_PRECISION:P0) —
  wymaga zgody.

### ASYNC_CALLBACKS + DATA_FLOW:D3 — model `AccessController.doPrivileged` — DONE for this scope (2026-11-05)

- **Zakres zatwierdzony:** (1) katalog ASYNC_CALLBACKS: sześć przeciążeń
  `doPrivileged`, nowy rodzaj wykonania `synchronous`, krawędź do
  `run()` istniejącą ścieżką A1/A2; (2) streszczenie biblioteczne D3:
  wynik `doPrivileged` = wartości zwracane przez `run()` akcji; reguła
  stopu przy zerowym efekcie (nie zadziałała).
- **Zrealizowane:** `internal/callback/catalog.go`
  (`ExecutionSynchronous`, wpisy dla `PrivilegedAction`/
  `PrivilegedExceptionAction` × {sam, z `AccessControlContext`, z
  kontekstem i `Permission...`}). `internal/dataflow/callresults.go`:
  `isDoPrivileged`, `resolvePrivileged` w `resolveCallResultOnceMode` —
  akcja z argumentu wywołania (po jednym kroku pola/wyniku wywołania):
  alokacja → własna, niemostkowa, bezargumentowa `run()` (więcej niż
  jedna → nieznane); lambda/referencja → implementacja bez argumentów
  (przechwycone wartości → nieznane); inne → nieznane z powodem.
  Wspólne miejsce, więc korzystają wszyscy odbiorcy wyników wywołań.
- **Weryfikacja:** fixture `fixture.jpms.app.Privileged` (klasa
  anonimowa, lambda, `PrivilegedExceptionAction`, akcja z parametru,
  `getDeclaredMethod` wewnątrz akcji). `TestDoPrivileged` (krawędzie do
  `run()`, brak krawędzi dla akcji z parametru, powiązanie konsumpcji
  z Codec, jawna luka dla parametru, refleksja → `Target#hit`).
  Mutacje: wyłączenie streszczenia, usunięcie filtra metod mostkowych,
  usunięcie wpisu katalogu — test pada w każdym przypadku. `go vet
  ./...`, `go test ./...`, `make build` — PASS. Offline `-profile
  spring`, po dwa stabilne przebiegi: krawędzie callbacków petclinic
  62→288, WebGoat 44→284 (rozwiązane identyczności 67→302 / 44→291);
  powiązane konsumpcje ServiceLoadera 46→52 / 100→106; krawędzie
  modelu 7→9 / 79; luki "does not resolve method doPrivileged" 13→0 /
  14→0, 2 akcje WebGoat nadal nie do udowodnienia; krawędzie refleksji
  bez zmian. Statusy CVE bez zmian, `results.json` bajtowo identyczny
  z baseline'em, czas bez zmian (~9,8 s / ~15 s).
- **Dokumentacja:** `docs/SUPPORT.MD` (ASYNC, DATA_FLOW), README,
  szablon raportu, ten rejestr.
- **Ograniczenia:** `doPrivilegedWithCombiner` i inne API
  uprzywilejowane poza katalogiem; akcja z parametru/pola bez kroku
  przez wywołujących (nie używa jeszcze `ResolveParameter`); lambdy z
  przechwyconymi wartościami; refleksja wewnątrz akcji z
  `try/catch` daje zbiór z `null` (nierozwiązany receiver).
- **Następny krok (wymaga zgody):** SERVICE_LOADING:S4 (ostatnia otwarta
  faza obszaru), dalsze modele B3, FRAMEWORKS:F3, albo ASYNC_CALLBACKS:A3
  (continuations/listenery).

### SERVICE_LOADING:S4 — Regresja i rozszerzenia API — DONE (2026-11-05); obszar SERVICE_LOADING zamknięty i zarchiwizowany

- **Zakres zatwierdzony:** wyrocznia prawdziwego JVM (kontrola B4),
  brakujące przypadki fixture (duplikat, dwa loadery), pomiar na stałym
  wejściu przez kolejne buildy, lista wspieranych sygnatur; rozszerzenie
  stream API świadomie pominięte (0 użyć na petclinic/WebGoat).
- **Zrealizowane:** `scripts/serviceloader-oracle.sh` +
  `testdata/oracle/src/oracle/OracleRunner.java` (poza analizowanym
  wejściem) → `testdata/oracle/expected.tsv`; `model-gaps.tsv`,
  `model-extras.tsv`; `internal/serviceload/oracle_test.go`
  (`TestServiceLoaderOracle`: brak oczekiwanego providera = błąd, chyba
  że zadeklarowana luka i analiza `Incomplete`; nadmiarowy provider =
  błąd, chyba że zadeklarowany z powodem; nieaktualne wpisy = błąd).
  Fixture: znaczniki `ORACLE-PROVIDER` w konstruktorach/`provider()`,
  usługa `Hasher` + `AutoHasher`, powtórzony wpis `AutoCodec`,
  `Regression#twoLoaders`; `TestRegression_duplicatesAndTwoLoaders`.
  `docs/SUPPORT.MD`: sekcja "ServiceLoader: exact supported signatures
  and scenarios". README fixture i README główne zaktualizowane.
- **Weryfikacja:** wyrocznia na JDK 25 (classpath, module-path z
  `--add-modules ALL-MODULE-PATH` i automatycznym modułem jako jar,
  mixed): 62 pary wejście/scenariusz, **0** nadmiarowych krawędzi,
  **1** zadeklarowana luka (`runThroughStreamCallback`,
  `map(Provider::get)`). Pierwsze uruchomienie ujawniło błąd w samej
  wyroczni (katalog bez `module-info` na module path nie jest modułem)
  — poprawiony, nie w modelu. Mutacje (filtr warstwy boot; pomijanie
  `META-INF/services` jawnego modułu) — test pada. Trzy istniejące
  asercje zaktualizowane o nową usługę (`uses` x2, materializacje +
  `AutoHasher`). `go vet ./...`, `go test ./...` — PASS. Pomiar
  historyczny (11 buildów od `30859e5`, offline): tabela w
  `archive/2026-11-05/service_loading/COMPLETION.md`; `results.json`
  bajtowo identyczny na wszystkich buildach obu aplikacji.
- **Dokumentacja:** `docs/SUPPORT.MD`, README, README fixture,
  `RECOMMENDATION/README.MD` (indeks), PLAN.md (tylko odnośnik do
  przeniesionego dokumentu), rekord COMPLETION.md, ten rejestr.
- **Następny krok (wymaga zgody):** inny obszar — kandydatury: dalsze
  modele B3, FRAMEWORKS:F3, ASYNC_CALLBACKS:A3, CVE:F0–F1 (B0: wpływ luk
  na werdykt), GRAPH_PRECISION (RTA, odblokowałoby też REFLECTION:R3).

### CVE:F0 — Kontrakty, zakres i przypadki odniesienia — DONE (kontrakt, 2026-11-05)

- **Zakres zatwierdzony:** całe F0 jako faza kontraktowa; baseline danych
  sieciowych jako manifest w repo + eksport poza repo (decyzja
  użytkownika). F1 rozpisane na F1a/F1b/F1c, każde z osobną zgodą.
- **Zrealizowane (F0.1–F0.6):**
  1. `docs/RESULTS_CONTRACT.MD`: osie ustalenia (tożsamość,
     `applicability`, precyzja mapowania z `role`/`coverage`/`review`,
     źródło dowodu, kompletność modelu, osiągalność), stany z opisem i
     przykładem, w tym v2-only `not_assessed`; rozdzielenie "brak metody"
     vs "brak ścieżki"; nazwy ustabilizowane wg propozycji "Wynik v2".
  2. `internal/findings`: stabilne kody `VERSION_UNKNOWN`,
     `CLASSPATH_CONFLICT`, `MAPPING_PARTIAL`, `MAPPING_ENTRYPOINT_ONLY`,
     `METHOD_NOT_FOUND`, `SOURCE_UNAVAILABLE`, `MODEL_INCOMPLETE`,
     `RUNTIME_UNSUPPORTED`; `TestReasonCodesDocumented` trzyma dokument i
     kod zgodne. Nic jeszcze ich nie ustawia.
  3. `testdata/cve-control/MANUAL_MAPPING_AUDIT.MD`: 39 wpisów —
     38 punktów wejścia (SnakeYAML, Commons Text, 36 × XStream ze
     wspólną kotwicą, z czego 2 zależą od drivera), 1 ujście (Log4Shell),
     0 pełnych mapowań kodu podatnego.
  4. `testdata/cve-control/cases.yaml`: 33 przypadki (mechanika
     `ARIADNE-TEST-*` + jakość ze źródłami), każde z 5 rozszerzeń i każdy
     wiersz macierzy §5; `TestControlSetCoversExtensions`.
  5. `testdata/baseline/`: wyniki v1, SHA-256 wejść, manifest 1405
     odpowiedzi sieciowych (35 MB, eksport 5,3 MB poza repo),
     `scripts/cve-baseline.sh`, `internal/httpcache` `ARIADNE_CACHE_TRACE`
     (lista trafień cache, bez wpływu na analizę; test).
  6. Decyzje w kontrakcie §5–§6: v1 domyślny bez zmiany znaczenia pól,
     v2 opt-in, eksport v1 z agregacją (pozytywny dowód widoczny,
     nierozstrzygnięta kopia nie znika pod NOT_REACHABLE, NOT_APPLICABLE
     tylko WebJar), stare mapowania = częściowe / tylko punkt wejścia.
- **Kryteria F0:** każdy stan ma opis i przykład, brak metody i brak
  ścieżki rozróżnione — tak; zestaw kontrolny obejmuje pięć rozszerzeń —
  tak (test); ustalono, które stare wyniki się zmienią —
  `testdata/baseline/VERDICT_CHANGE_ESTIMATE.MD`: 7/26 (petclinic) i 6/18
  (WebGoat) NOT_REACHABLE to "metoda nie znaleziona" → UNKNOWN w F1;
  szeroka reguła luk ("luka gdziekolwiek w jarze") objęłaby 25/26 i 17/18,
  więc kontrakt wymaga też osiągalności miejsca luki.
- **Weryfikacja:** `go test ./internal/findings ./internal/httpcache`,
  pełne `go test ./...`, `go vet ./...` — PASS. Baseline odtworzony z
  samego eksportu w pustym katalogu cache: oba `results.json` bajtowo
  identyczne, wszystkie hashe odpowiedzi zgodne. Szacunek zmian:
  tymczasowa sonda (test Go, usunięty).
- **Dokumentacja:** `docs/RESULTS_CONTRACT.MD` (nowy), `docs/SUPPORT.MD`
  (Interpretation), README (tabela statusów: NOT_REACHABLE obejmuje dziś
  "metoda nie znaleziona"), ten rejestr. FAQ bez zmian.
- **Ograniczenia:** kontrakt nie zmienia działania; audyt mapowań oparty na
  opisach advisory (odnośniki do ponownego sprawdzenia), bez strukturalnych
  pól w YAML; część oczekiwań w zestawie kontrolnym to jawne placeholdery
  do ustalenia przez fazę-właściciela (np. wycofane rekordy w F2).
- **Następny krok (wymaga zgody):** CVE:F1a — model komponentów
  (`Artifact`, `ComponentInstance`, SHA-256, stabilne identyfikatory,
  powiązanie klas z instancją); potem F1b (strukturalne mapowania, kody
  powodów w wynikach, METHOD_NOT_FOUND i reguła luk) i F1c
  (`-json-format v2`).

### CVE:F1a — model artefaktów i instancji komponentów — DONE for this slice; CVE:F1 IN_PROGRESS (2026-11-05)

- **Zakres zatwierdzony:** zadania F1.1–F1.3 i F1.6–F1.8 planu CVE:
  `internal/component` (`Artifact`, `Instance`, `IdentityEvidence`,
  `Conflict`), SHA-256 i stabilna lokalizacja, identyfikatory z treści i
  lokalizacji, jedna instancja na jar (dodatkowe `pom.properties` jako
  kandydaci osadzeni), powiązanie klas z instancją, inwentarz kopii
  niewybranych; bez zmian werdyktów i `results.json`.
- **Zrealizowane:** `internal/component/component.go` (`Build`,
  `InstanceForJar`/`InstanceForClass`, `Counts`); `internal/input`:
  `PreparedApp.LibEntries` (oryginalna ścieżka każdego wypakowanego jara:
  `BOOT-INF/lib/…`, `WEB-INF/lib/…`, `x.war!/WEB-INF/lib/…`, jar EAR-a),
  `JarPOMProperties`. Lokalizacje: `app:<nazwa wejścia>`,
  `deps:<plik>`, `app:<nazwa>!/<wpis>`. ID artefaktu = skrót SHA-256(treść,
  lokalizacja), ID instancji = skrót (artefakt, tożsamość). Konflikty z
  `hierarchy.Duplicates()` (JVM_GRAPH G1 — deterministyczny wybór już
  istniał), z jawnym założeniem kolejności classpath. CLI: linia
  `components:`; raport: sekcja "Components" + "Class conflicts".
- **Weryfikacja:** `internal/component`: `TestIDsStableAcrossTempPaths`
  (katalog klas i fat jar budowany dwa razy w różnych katalogach),
  `TestTwoInstancesOfOneComponent` (ten sam GAV w dwóch jarach → dwie
  instancje z własnymi dowodami, kandydat osadzony, konflikt z wybraną i
  przesłoniętą kopią), `TestUnidentifiedJar`; mutacje (ścieżka w ID;
  usunięcie kandydatów osadzonych) — testy padają. `go vet ./...`,
  `go test ./...` — PASS. Offline: petclinic 88 artefaktów / 87
  zidentyfikowanych / 31 konfliktów (wszystkie identyczne), WebGoat 127 /
  126 / 242 (163 o różnej treści); `results.json` obu bajtowo identyczny z
  baseline'em CVE:F0. petclinic jako fat jar `spring-petclinic-2.4.5.jar`
  (online raz, by uzupełnić cache OSV, potem offline): 76 artefaktów z
  lokalizacjami `BOOT-INF/lib`, 1 niezidentyfikowany
  (`spring-boot-jarmode-layertools`), tabela komponentów i wyniki
  identyczne w obu przebiegach. Kandydaci osadzeni: byte-buddy, ehcache,
  jruby-core.
- **Kryteria F1 a ten wycinek:** "dwie instancje tego samego komponentu"
  — osobne instancje i dowody tak, osobne *wyniki* to F1b/F1c; "losowa
  kolejność odczytu nie zmienia wybranej klasy" — zapewnione przez G1,
  inwentarz to raportuje.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz tożsamości), README (sekcja
  raportu), `docs/RESULTS_CONTRACT.MD` (oś tożsamości), szablon raportu,
  ten rejestr.
- **Następny krok (wymaga zgody):** CVE:F1b — strukturalne mapowania
  (`MethodSelector` z rolą i kompletnością), ustalenia per advisory ×
  instancja, kody powodów w wynikach (`METHOD_NOT_FOUND`,
  `MAPPING_ENTRYPOINT_ONLY`, `CLASSPATH_CONFLICT`) i reguła luk modelu
  (kontrakt §4); potem F1c (`-json-format v2`).

### CVE:F1b — strukturalne mapowania, kody powodów, ustalenia per instancja, reguła luk — DONE for this slice; CVE:F1 IN_PROGRESS (2026-11-05)

- **Zakres zatwierdzony:** zadania F1.4, F1.5 (bez eksportu v2), F1.10
  (agregacja v1) planu CVE i reguła luk kontraktu §4. Decyzja w trakcie
  (użytkownik, po pomiarze): reguła luk **report-only**.
- **Zrealizowane:**
  1. `vuln_methods.yaml`: forma obiektowa `{methods, role, coverage,
     review}` obok listy (lista = partial/reviewed); walidacja wartości;
     39 wpisów z rolami wg audytu F0 (entrypoint/sink, wszystkie
     partial). `vulnmap.Entry` niesie `Role/Coverage/Review`;
     `WithDefaults`: diff → vulnerable/partial/unreviewed, LLM →
     vulnerable/unknown/unreviewed (także wpisy ze starego cache).
  2. `reach`: `Result.Reasons` (v1, addytywne), `InstanceFinding` per
     dotknięta instancja komponentu (CVE:F1a). Węzeł przypisany do
     instancji klasy deklarującej metodę; kopia wybrana z innego jara,
     gdy dotknięta instancja ma własną przesłoniętą kopię, nie jest jej
     dowodem (`CLASSPATH_CONFLICT`). Agregacja §5: pozytywna ścieżka
     wygrywa, w przeciwnym razie nierozstrzygnięta instancja daje
     `UNKNOWN` (nigdy nie znika pod NOT_REACHABLE). `METHOD_NOT_FOUND` →
     `UNKNOWN`; `MAPPING_ENTRYPOINT_ONLY` z roli; `MAPPING_PARTIAL` na
     NOT_REACHABLE (opcja B: werdykt zachowany, v2 niesie kompletność).
  3. Reguła luk: `cmd/ariadne/gaps.go` zbiera nierozwiązane miejsca
     (refleksja invoke/newInstance i MethodHandle — nieograniczone;
     callback — podtypy interfejsu callbacku; ServiceLoader —
     zarejestrowani dostawcy), `Engine.SetGaps` zostawia osiągalne z
     korzeni (przebieg FRAMEWORK). Istotna luka na NOT_REACHABLE dodaje
     `MODEL_INCOMPLETE` i `model_gaps` (liczby wg rodzaju), werdykt
     bez zmian. CLI: linia `model gaps:`; raport: nagłówek, kolumna
     Reasons, sekcja "Findings per component instance".
- **Pomiar reguły luk (powód report-only):** osiągalne miejsca 424
  (376 nieograniczonych) petclinic, 526 (456) WebGoat. Reguła jak w
  kontrakcie: 19/19 i 12/12 NOT_REACHABLE → UNKNOWN (NOT_REACHABLE = 0);
  tylko miejsca ograniczone: 18/19 i 11/12. Etykieta na wszystkim nic
  nie rozróżnia; kontrakt §4 poprawiony, ARIADNE-TEST-0004 zachowuje
  docelową formę `UNKNOWN` z adnotacją `v1`.
- **Weryfikacja:** `internal/reach`: `TestContractCases` (0001–0003,
  0007, 0008, dwie instancje), `TestModelGapRule` (nieograniczona,
  ograniczona do dotkniętej / innej instancji, nieosiągalne miejsce,
  pozytywna ścieżka), `TestAggregateNeverHidesUnresolvedInstance`;
  `internal/vulnmap` (forma obiektowa, walidacja); `internal/report`
  (`TestJoinReasonsWithGaps`, `TestShowPerInstance`). Mutacje (wyłączona
  agregacja §5; `relevantGaps` zawsze puste) — testy padają. `go vet
  ./...`, `go test ./...` — PASS. Offline na baseline'ie CVE:F0
  (`-profile spring`): zmiany wyłącznie oczekiwane — petclinic 7
  NOT_REACHABLE → UNKNOWN/`METHOD_NOT_FOUND`, WebGoat 6 (dokładnie
  szacunek F0); 19/12 NOT_REACHABLE z `MAPPING_PARTIAL,MODEL_INCOMPLETE`;
  `MAPPING_ENTRYPOINT_ONLY` na 1 (petclinic) / 37 (WebGoat) pozytywnych
  wierszach; wiersze REACHABLE/FRAMEWORK poza `reasons` bajtowo zgodne z
  baseline'em (ścieżki, wejścia). Drugi przebieg WebGoat: `results.json`
  bajtowo identyczny. Przykład per instancja: CVE-2023-6378
  logback-classic NOT_REACHABLE, logback-core FRAMEWORK — osobne wyniki.
- **Kryteria F1 a ten wycinek:** "dwie instancje tego samego komponentu
  mają osobne wyniki i dowody" — tak (raport; eksport v2 to F1c);
  "losowa kolejność odczytu nie zmienia wybranej klasy" — G1/F1a;
  "v2 nie miesza niepewności mapowania z osiągalnością" — F1c;
  "uzasadnione zmiany werdyktów opisane" — tak (wyżej, README, kontrakt).
- **Dokumentacja:** `docs/RESULTS_CONTRACT.MD` (§3 emitowane kody, opcja
  B; §4 pomiar i report-only), `docs/SUPPORT.MD`, README (tabela
  statusów, `vuln_methods.yaml`, pola `reasons`/`model_gaps`), FAQ
  (statusy), `testdata/baseline/VERDICT_CHANGE_ESTIMATE.MD` (pomiar),
  `testdata/cve-control/cases.yaml` (pole `v1` dla 0004, 0005), ten
  rejestr. Baseline nie był regenerowany (pozostaje odniesieniem F0).
- **Ograniczenia:** reguła luk nie zmienia werdyktu, dopóki rozwiązywanie
  refleksji/MethodHandle nie zawęzi kandydatów; ograniczenia callbacków
  (wszystkie podtypy interfejsu) i ServiceLoadera (wszyscy dostawcy) są
  szerokie. `VERSION_UNKNOWN`, `SOURCE_UNAVAILABLE`, `RUNTIME_UNSUPPORTED`
  nieustawiane. Węzeł spoza dotkniętej instancji, ale z pokrewnego
  artefaktu (np. advisory na `spring-boot-starter-web`, kod w
  `spring-beans`), nadal się liczy, z adnotacją.
- **Następny krok (wymaga zgody):** CVE:F1c — `-json-format v2`
  (komponenty, ustalenia per instancja, pokrycie, odwołanie do
  snapshotu) i eksport v1 z v2.

### CVE:F1c — wynik v2 i eksport v1 — DONE; CVE:F1 DONE (2026-11-05)

- **Zakres zatwierdzony:** zadania F1.9 (`-json-format v2`) i F1.10
  (eksport v1 z v2) planu CVE.
- **Zrealizowane:**
  1. `internal/resultsv2`: `Document` (`schema_version` "2", `tool`,
     `snapshot`, `components`, `advisories`, `findings`, `coverage`;
     schemat w `docs/RESULTS_CONTRACT.MD` §5). `snapshot` to odwołanie
     do wejść (SHA-256 po stabilnych lokalizacjach i treści artefaktów,
     niezależny od ścieżek tymczasowych) — snapshot danych to CVE:F3.
     Tożsamość `identified`/`unidentified`/`conflict` (przesłonięta
     kopia o innej treści). Mapowanie jako osobna oś (`match_precision:
     name`, `source`, `role`, `coverage`, `review`); kody `MAPPING_*` nie
     trafiają do `reason_codes` ustaleń. `applicability` zawsze
     `affected` (OSV pytane dokładną wersją; zakresy to F4).
  2. `ExportV1`/`CheckV1`: werdykt v1 jako `reach.Aggregate` ustaleń
     (kontrakt §5). Silnik liczy wiersz v1 tą samą funkcją; CLI buduje
     v2 w każdym przebiegu i kończy się błędem, jeśli eksport nie zgadza
     się z v1.
  3. `reach`: by werdykt wiersza był agregatem ustaleń także w
     przypadkach brzegowych — instancja, której własny kod nie daje
     REACHABLE, sprawdza też kod pokrewnych, niedotkniętych artefaktów
     (z notatką; własny kod wygrywa remis); instancja z przesłoniętą
     kopią może mieć ścieżkę tylko przez taki kod (zostaje
     `CLASSPATH_CONFLICT`); degradacja LLM obejmuje ustalenia;
     werdykty bez dopasowania (WebJar, brak mapowania) dostają ustalenie
     per instancja (`fillFindings`). `InstanceFinding.Entry`.
  4. CLI: `-json-format v1|v2` (domyślnie v1, walidacja), linia
     `results: N advisories, M findings …; v1 export from v2 consistent`.
- **Weryfikacja:** `internal/resultsv2`: `TestBuild` (tożsamości, sortowanie,
  domyślne mapowania, brak kodów mapowania w ustaleniach, pokrycie,
  liczniki), `TestSnapshotIDIgnoresPaths`, `TestExportV1` (agregacja i
  wykrycie niezgodności); `internal/reach`: `TestAggregate`,
  `TestFindingsForUnmatchedVerdicts`, `TestRelatedArtifactEvidence`
  (mutacja: wyłączenie reguły pokrewnego artefaktu — test pada);
  `cmd/ariadne`: `TestJSONFormatFlag`. `go vet ./...`, `go test ./...` —
  PASS. Offline (`-profile spring`): v1 obu aplikacji **bajtowo identyczny**
  z F1b (binarka z `6033206` w osobnym worktree); v2 petclinic 148
  advisory / 155 ustaleń, WebGoat 160 / 168, 0 advisory bez instancji,
  eksport v1 zgodny w 100%; drugi przebieg v2 WebGoat bajtowo
  identyczny. Przykład: CVE-2023-6378 — logback-classic NOT_REACHABLE z
  `MODEL_INCOMPLETE`, logback-core FRAMEWORK, wiersz v1 FRAMEWORK.
- **Kryteria F1:** dwie instancje — osobne wyniki i dowody (v2); losowa
  kolejność odczytu — G1/F1a; v2 nie miesza niepewności mapowania z
  osiągalnością, ograniczenia eksportu v1 opisane (kontrakt §5);
  dotychczasowe polecenia bez zmian, zmiany werdyktów opisane (F1b).
  **F1 zamknięte.**
- **Dokumentacja:** kontrakt (§1, §5: schemat v2, ograniczenia v1),
  README (flaga, `results.json`), `docs/SUPPORT.MD`, FAQ, ten rejestr,
  nagłówek F1 w planie.
- **Ograniczenia:** `not_assessed` zdefiniowane, nieemitowane; brak
  snapshotu danych (F3), zakresów wersji (F4) i JDK (F8); raport
  Markdown nadal na v1 z sekcją per instancja.
- **Następny krok (wymaga zgody):** CVE:F2 — cache z metadanymi i
  polityką świeżości.

### CVE:F2 — cache z metadanymi i polityką świeżości — DONE (2026-11-05)

- **Zakres zatwierdzony:** zadania F2.1–F2.11 planu CVE.
- **Decyzja projektowa:** metadane w osobnym pliku `<klucz>.meta.json`
  obok niezmienionego `<klucz>.json`, związane z treścią przez SHA-256
  (niezgodny lub brakujący plik = wiek nieznany). Dzięki temu hashe
  odpowiedzi baseline'u F0 (`responses.tsv`) i eksport pozostają ważne;
  spójność "jednej jednostki" zapewnia hash, nie wspólny plik.
- **Zrealizowane:**
  1. `internal/httpcache`: `Meta` (rodzaj, metoda, URL, `fetched_at`,
     `validated_at`, SHA-256, rozmiar, ETag/Last-Modified — bez nagłówków
     żądania), zapis atomowy treści i metadanych; rodzaje `osv`, `pinned`
     (pełny 40-znakowy SHA commita — nie wygasa), `mutable`, `llm`
     (nigdy nie odświeżane); `Policy` (TTL domyślnie 24h, `MaxAge` jako
     twardy limit, `Refresh`, `AllowStale`), wstrzykiwany zegar, żądania
     warunkowe i 304, statystyki per rodzaj (`Freshness`, z najstarszym
     czasem walidacji zamiast wieku, by v2 był deterministyczny).
  2. `internal/cves`: `published`/`modified`, `Sources` (każdy rekord z
     datami, `withdrawn`, zakresami), `ResolveAll` z `Withdrawn`;
     wycofany rekord łączy aliasy i wersje, ale nie wnosi treści; gdy
     grupa nie ma aktywnego rekordu, pobierane są rekordy aliasów (404 =
     brak); grupa w całości wycofana — lista, bez analizy. Aliasy liczone
     od nowa przy każdym rozwiązaniu.
  3. CLI: `-data-ttl`, `-refresh`, `-allow-stale`, `-max-data-age`
     (`-refresh` + `-offline` = błąd walidacji), log `data freshness:`.
     Raport: nagłówek "Data freshness" z banerem dla stale/unknown,
     ostrzeżenia, sekcja "Withdrawn advisories"; v2:
     `coverage.data_freshness`, `coverage.withdrawn_advisories`.
  4. `vulnmap` LLM: indeks ostatniej odpowiedzi per (model, wersja
     promptu, CVE); zmieniony tekst advisory po odświeżeniu nie wysyła
     nowego zapytania — poprzednia odpowiedź użyta z ostrzeżeniem
     (`Reused`), `-llm-reask-changed` pyta ponownie.
- **Weryfikacja:** testy `internal/httpcache` (TTL i granica, pinned/LLM
  bez wygasania, ETag/304, nieudane odświeżenie z/bez `-allow-stale`,
  `-max-data-age` twardy także z `-allow-stale` i offline, wpisy bez
  metadanych i z niezgodnym hashem, `-refresh`, brak sekretów w
  metadanych; `-race`), `internal/cves` (0401 przez prawdziwego klienta,
  0402, wszystkie źródła wycofane), `internal/vulnmap`
  (`TestLLM_changedAdvisoryReusesPreviousAnswer`), `internal/report`,
  `internal/resultsv2`, `cmd/ariadne` (`TestFreshnessFlags`). `go vet
  ./...`, `go test ./...` — PASS. Na prawdziwych danych:
  - offline na `.cache` baseline'u: v1 obu aplikacji bajtowo identyczny z
    F1c, `.cache` nietknięty (1680 plików), stan "unknown" (149/161
    odpowiedzi OSV bez metadanych), `-offline -max-data-age 24h` kończy
    się jednoznacznym błędem;
  - online na kopii cache: petclinic 148 → **153** advisory (4 ×
    tomcat-embed-core 9.0.45, 1 × jackson-databind 2.11.4, wszystkie
    FRAMEWORK), WebGoat 160 → **164**; żaden istniejący werdykt się nie
    zmienił; drugi przebieg — wszystko z cache, ustalenia identyczne;
    `-refresh` — 154 odświeżone; potem `-offline -max-data-age 1h` —
    przechodzi; OSV nie wysyła ETag (rewalidacja = pełne pobranie);
  - `-llm` online z fikcyjnym kluczem na odświeżonych danych: 0 zapytań
    (0 skipped), 0 ponownych użyć (teksty się nie zmieniły), wyniki
    offline `-llm` bajtowo identyczne, `.cache/llm` nietknięty.
- **Kryteria F2:** nowe CVE po odświeżeniu pustej odpowiedzi — tak (test i
  dane rzeczywiste); wiek danych i awaria odświeżenia widoczne,
  jednoznaczne — tak; wycofane źródło i aktywny alias — tak (test; brak
  wycofanego rekordu w aplikacjach referencyjnych, OSV-WITHDRAWN-REAL
  nadal bez konkretnego rekordu); odświeżenie CVE bez automatycznych
  płatnych zapytań LLM — tak.
- **Dokumentacja:** README (krok OSV, cache, flagi, LLM), FAQ,
  `docs/SUPPORT.MD`, kontrakt §5 (pokrycie v2),
  `testdata/cve-control/cases.yaml` (`v1` dla 0401–0403 i
  OSV-WITHDRAWN-REAL), ten rejestr, nagłówek F2 w planie.
- **Ograniczenia:** brak `SOURCE_UNAVAILABLE` per ustalenie (świeżość jest
  per przebieg); zakresy przechowywane, nieoceniane (F4); brak snapshotu
  danych (F3); wpisy `pinned` sprzed F2 pozostają bez metadanych (nie
  wygasają, więc nie są pobierane); kontrole exploit/framework-check
  (LLM) nie mają reguły ponownego użycia — klucz zależy od ścieżki.
- **Następny krok (wymaga zgody):** CVE:F3 — snapshoty, odtwarzanie i
  unieważnianie wyników pochodnych.

### CVE:F3 — snapshoty, odtwarzanie i unieważnianie wyników pochodnych — DONE (2026-11-05)

- **Zakres zatwierdzony:** zadania F3.1–F3.8 planu CVE.
- **Zrealizowane:**
  1. `internal/snapshot`: `Manifest` (wersja formatu, `data_set_id` =
     SHA-256 po obserwacjach, nie po czasie; `created_at` osobno; narzędzie:
     wersja, Go, hash binarki, `dirty`/`reproducible`; wersje algorytmów;
     konfiguracja; wejścia po stabilnych lokalizacjach; obserwacje;
     wyprowadzenia; werdykty v1; stan świeżości), magazyn `objects/`
     adresowany SHA-256, `Recorder` (stan końcowy klucza: trafienie wygrywa
     z wcześniejszym chybieniem), `Replayer` (weryfikacja hasha przy każdym
     odczycie, lista porażek), `CheckTool`, `CheckInputs`, `ConfigDiff`,
     `Compare`.
  2. `internal/httpcache`: `Observation` (trafienie / chybienie / błąd z
     klasą `offline`/`too_old`/`other` — obserwacje negatywne, np. 404
     aliasu), haki `SetRecorder`/`SetReplay`/`SetNamespace`; w odtwarzaniu
     brak odczytu katalogu, zapisu i sieci, błędy odtwarzane z tą samą
     klasą dla `errors.Is`, `ErrNotInSnapshot`.
  3. Klucz wyprowadzenia per advisory: hash advisory (wszystkie źródła),
     dotknięte instancje (lokalizacje) i wersje, hash mapowania (metody,
     rola, pokrycie, przegląd), hashe treści użytych do mapowania (diffy,
     pliki, rekord LLM — zbierane per CVE przez kontekst w `vulnmap`),
     `vulnmap.ResolverVersion`, dla LLM model, wersja promptu i znacznik
     odpowiedzi ponownie użytej (F2).
  4. Wersje algorytmów: `callgraph.BuilderVersion`, `reach.EngineVersion`,
     `vulnmap.ResolverVersion`, wersje promptów LLM/exploit/framework,
     `results.v2`, format snapshotu.
  5. CLI: `-snapshot-out`, `-snapshot-in` (wymusza offline; z `-refresh` lub
     `-snapshot-out` — błąd walidacji), `-snapshot-compare`; odtwarzanie
     kończy się błędem, gdy snapshot nie odpowiedział na którekolwiek
     zapytanie (nawet gdy wywołujący połknąłby błąd jako ostrzeżenie).
     Raport: baner "Replayed", stan `replayed`, sekcja "Changes since
     snapshot"; v2: `snapshot.data_set`, `coverage.snapshot_changes`.
- **Znalezione i naprawione w trakcie:** uszkodzony obiekt początkowo nie
  przerywał odtwarzania (błąd pobrania diffu jest w `vulnmap` ostrzeżeniem
  per CVE) — stąd lista porażek `Replayer` sprawdzana przed zapisem
  wyników; Go oznacza zmodyfikowane drzewo jako `+dirty`, nie `-dirty` —
  wykrywanie poprawione.
- **Weryfikacja:** `internal/snapshot` (zapis/odtworzenie, obserwacja
  negatywna, uszkodzony obiekt i edytowany manifest — 0404, ID niezależne
  od czasu, `CheckTool` z wersją/algorytmem/binarką dirty, `CheckInputs`,
  `Compare` — zmiana advisory, mapowania, algorytmu, nowe/usunięte,
  porażki), `internal/httpcache` (`TestRecordThenReplayWithoutNetwork`),
  `cmd/ariadne` (`TestSnapshotFlags`); `-race`; `go vet ./...`, `go test
  ./...` — PASS. Na prawdziwych danych (offline, `.cache` nietknięty):
  - zapis nie zmienia wyników (v1 petclinic/WebGoat bajtowo jak F2);
    snapshot petclinic 23 MB / 1082 obserwacji, WebGoat 14 MB / 895;
    dwukrotny zapis WebGoat — ten sam `data_set_id`;
  - odtworzenie z pustym `-cache-dir` w przestrzeni nazw bez sieci
    (`unshare -rn`): v1 bajtowo identyczny, ustalenia v2 identyczne, dwa
    odtworzenia v2 bajtowo identyczne, katalog cache pusty; z `-llm` (114
    obserwacji LLM) — identyczny;
  - odrzucone: zmieniony `jackson-databind-2.11.4.jar` (hash), uszkodzony
    obiekt, inna binarka dirty; zmieniona konfiguracja (`-roots profile`) —
    dozwolona z ostrzeżeniem, `config_matches: false`;
  - `-snapshot-compare`: dane OSV odświeżone w F2 względem starego
    snapshotu — 5 nowych advisory i 1 ze zmienioną treścią
    (CVE-2026-54512), werdykty bez zmian; zmienione `coverage` jednego
    ręcznego mapowania (`-vuln-methods`) — unieważnione wyłącznie
    CVE-2022-1471.
- **Kryteria F3:** analiza ze snapshotu bez sieci — tak; zmieniony artefakt
  odrzucony — tak; zmiana advisory/mapowania/wersji algorytmu unieważnia
  właściwe wyniki — tak (dane rzeczywiste i testy); dokumentacja rozróżnia
  odtworzenie danych od środowiska — tak (README "Data snapshots", SUPPORT).
- **Dokumentacja:** README (flagi, sekcja "Data snapshots"), FAQ,
  `docs/SUPPORT.MD` (nowy wiersz), kontrakt §5, `testdata/baseline/README.MD`,
  `testdata/cve-control/cases.yaml` (`v1` dla 0404), ten rejestr, nagłówek
  F3 w planie.
- **Ograniczenia:** binaria wejściowe nie są archiwizowane (tylko hashe);
  analizator musi być tej samej wersji (dirty: ta sama binarka); Go/OS
  zapisywane, niewymuszane; wersje algorytmów podnoszone ręcznie; zmiana
  wersji analizatora unieważnia w porównaniu wszystkie wyprowadzenia;
  wyniki pochodne nie są cache'owane między przebiegami (mapowania liczone
  od nowa) — klucze służą wyjaśnieniu zmian, nie ponownemu użyciu.
- **Następny krok (wymaga zgody):** CVE:F4 — mapowania wersji i dokładnych
  sygnatur.

### CVE:F4 — mapowania wersji i dokładnych sygnatur — DONE (2026-11-05)

- **Zakres zatwierdzony:** zadania F4.1–F4.11 planu CVE. F4.5–F4.7
  (dokładny lookup w `internal/hierarchy`, CHA z deskryptorem, wywołania
  dziedziczone) były już zrobione w JVM_GRAPH G2 (`ResolveExact`,
  `ResolveByParams`) — tu wykorzystane, nie powtórzone.
- **Decyzja (plan, tabela decyzji "Komparator wersji Maven"):** własna
  implementacja w Go bez nowej zależności (brak pytania o licencję),
  zgodność ustalona z **zainstalowanym Maven 3.9.12** jako wyrocznią.
- **Zrealizowane:**
  1. `internal/mavenversion`: porządek Maven 3.9 (tokeny, podlisty,
     kwalifikatory, aliasy, normalizacja), `Regular` (SNAPSHOT i wersje
     nienumeryczne — nierozstrzygnięte, bez porównania tekstowego),
     `ParseRange` (składnia Maven, unie; goła wersja odrzucona),
     `Range.Contains`, `EvalOSV` (`ECOSYSTEM`: introduced/fixed/
     last_affected/limit, sortowane wg wersji). Dane referencyjne z Maven:
     197 wersji (przykłady specyfikacji + wszystkie wersje zależności
     aplikacji referencyjnych) → 38809 porównań, 197 struktur tokenów, 396
     przypadków zakresów; `scripts/maven-version-oracle.sh` je odtwarza
     (sprawdzone: wynik identyczny).
  2. `vulnmap`: selektory `Class#name` / `Class#name(params)` /
     `Class#name(params)ret` z walidacją deskryptorów (`ParseSelector`,
     `MappingPrecision`); schemat v2 tabeli ręcznej: `variants` z
     `package`, `versions` (zakres Maven), `methods`; ścisła walidacja
     (nieznane pola, identyfikatory, pakiety, zakresy, deskryptory, albo
     `methods` albo `variants`); stary YAML bez zmian znaczenia (bez
     zakresu, po nazwie, partial). LLM: role per metoda zachowane
     (`MethodRoles`), wpis `entrypoint`, gdy wszystkie metody to
     `entry_point`. CVE-2022-1471 w tabeli osadzonej wersjonowane do
     `(,2.0)`.
  3. `reach`: dopasowanie z deskryptorem (dokładne / po parametrach /
     po nazwie; podtyp nadpisuje tylko ten sam deskryptor), wybór wariantu
     per instancja z kodami `VERSION_UNKNOWN`, `MAPPING_VERSION_UNCOVERED`,
     `MAPPING_CONFLICT` (nowe, w kontrakcie §3), grupowanie instancji o
     tej samej liście metod, agregacja z ustaleń; `MatchPrecision`
     pozytywnych ustaleń.
  4. v2: `mapping.match_precision` (wyliczana), `variants`,
     `method_roles`; ustalenie: `match_precision`, `applicability_check`
     (zakresy OSV oceniane komparatorem obok odpowiedzi OSV).
  5. Fikstura `testdata/overloadfixture` (lib.jar/app.jar) kompilowana bez
     binarki `javac` — `scripts/JavaBuild.java` (moduł `jdk.compiler` przez
     `ToolProvider`, `-source/-target 17`, bo środowisko nie ma `ct.sym`),
     `scripts/build-overload-fixture.sh`; jary bajtowo powtarzalne.
  6. Wersje algorytmów: `vulnmap.resolver` 1→2, `reach.engine` 1→2; hash
     mapowania w kluczu wyprowadzenia (F3) obejmuje warianty i role.
- **Weryfikacja:** `internal/mavenversion` (Maven: porównania, tokeny,
  zakresy; `Regular`, błędne zakresy, `EvalOSV` z nieposortowanymi
  zdarzeniami i last_affected), `internal/vulnmap` (selektory, schemat v2 i
  8 rodzajów odrzuceń, role LLM), `internal/reach` (`TestExactSelectors` —
  0104 i dziedziczenie, `TestVersionedMappings` — 0101, brak pokrycia,
  konflikt, SNAPSHOT), `internal/resultsv2` (`TestCheckApplicability`).
  Mutacje: ignorowanie deskryptora, `Contains` zawsze `In` — testy padają.
  `go vet ./...`, `go test ./...` — PASS. Na prawdziwych danych (offline):
  - v1 **bajtowo identyczny** z F3 na petclinic, WebGoat i czterech
    demach (servlet, jakarta, micronaut, quarkus), także po wersjonowaniu
    CVE-2022-1471 (SnakeYAML 1.27/1.30 < 2.0; demo servlet nadal
    REACHABLE + `MAPPING_ENTRYPOINT_ONLY`);
  - `applicability_check`: 155/155 (petclinic) i 168/168 (WebGoat)
    ustaleń — zakresy OSV oceniane komparatorem dają `in`, zgodnie z OSV;
  - `-llm` (petclinic): jedyna zmiana — CVE-2016-1000027 dostaje
    `MAPPING_ENTRYPOINT_ONLY` (model oznaczył obie metody `entry_point`;
    sprawdzone w zapisanej odpowiedzi).
- **Kryteria F4:** dwa przeciążenia i dwa zakresy wersji — rozłączne
  dopasowania (test na skompilowanej fiksturze); brak doklejania
  przeciążenia po nazwie — tak dla selektorów z deskryptorem (G2 dla
  grafu); dziedziczone wywołania zachowują właściwe ciało — tak;
  brakująca metoda, nieznana wersja, konflikt — nierozstrzygnięte z kodem.
- **Dokumentacja:** README (krok 6, `results.json`), FAQ, `docs/SUPPORT.MD`
  (nowy wiersz), kontrakt (§3: dwa kody, `VERSION_UNKNOWN`; §5: pola v2),
  nagłówek `vuln_methods.yaml`, `testdata/cve-control/cases.yaml` (`v1` dla
  0101–0104 i CVE-2022-1471), ten rejestr, nagłówek F4 w planie.
- **Ograniczenia:** diffy i LLM nadal dają selektory po nazwie (obie strony
  poprawki — F5); w tabeli osadzonej wersjonowane jest tylko
  CVE-2022-1471; `applicability` to odpowiedź OSV, ocena zakresów jest
  kontrolą krzyżową (`not_affected` nie jest jeszcze emitowane); metody
  bridge dopasowywane jako zwykłe zadeklarowane metody. Uwaga na przyszłość:
  `scripts/JavaBuild.java` umożliwia teraz realną weryfikację bytecode,
  na którą czekał JVM_GRAPH G2 (bridge/default) — poza zakresem F4.
- **Następny krok (wymaga zgody):** CVE:F5 — obie strony poprawki i
  powiązanie z badanym wydaniem.

### CVE:F5a — obie strony poprawki, kandydaci ze starego kodu — DONE for this slice; CVE:F5 IN_PROGRESS (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** zadania F5.1–F5.4, F5.6, F5.10 i
  część F5.11 (klucze wyprowadzeń); źródło rodzica: GitHub REST API z
  opcjonalnym tokenem i fallbackiem do `.diff`. F5.5/F5.8 (weryfikacja w
  bytecode z deskryptorem), F5.7 (powiązanie commita z gałęzią/wydaniem),
  F5.9 (heurystyka przemianowań metod), prompt LLM z obiema stronami — F5b.
- **Zrealizowane:**
  1. `internal/javadiff`: `FileDiff{OldPath, NewPath, Change, OldLines,
     NewLines}` (added/modified/deleted/renamed; `rename from/to`,
     `new/deleted file mode`, `/dev/null` po obu stronach); treść hunku
     konsumowana według liczników z nagłówka (usunięta linia `-- x`
     nie jest nagłówkiem pliku); pliki usunięte i czyste rename są
     zwracane. `DeclaredMethods(src)`. Poprawka przy okazji: metoda
     otwarta i zamknięta w jednej linii (`void g() { … }`) jest teraz
     właścicielem tej linii (wcześniej linia nie miała właściciela).
  2. `internal/vulnmap/diffsource.go`: metadane commita
     `GithubAPIBaseURL/repos/{o}/{r}/commits/{sha}` (pełne SHA, rodzice;
     patch nadal z `.diff`); `GITHUB_TOKEN` (cmd) wysyłany, nie
     zapisywany w cache. Jeden rodzic → stara treść z
     `raw/<parent>/<old-path>`, nowa z `raw/<sha>/<new-path>`;
     klasyfikacja `changed`/`removed`/`added` po nazwie metody
     zadeklarowanej po każdej stronie; mapowanie = changed + removed pod
     nazwą klasy sprzed poprawki; `added` tylko w dowodach. Brak
     metadanych, merge (≥2 rodziców), root commit, nieczytelny stary plik
     → strona po poprawce (dawne zachowanie) i `BaseNewSideOnly` z
     powodem. Dowody: `Entry.FixCommits` (repo, SHA, parent, base,
     per plik ścieżki, zmiana, SHA-256 obu stron, metody z rodzajem).
  3. Błędy źródła per CVE: nieczytelny diff → `FixCommit.Error`, reszta
     commitów i CVE dalej; gdy żaden commit nie dał się odczytać i nie ma
     innego źródła → `Entry{SourceError}` → `UNKNOWN` +
     `SOURCE_UNAVAILABLE`. `ErrOffline`/`ErrTooOld`/`ErrNotInSnapshot`
     na `.diff` nadal przerywają przebieg (`isContractError`). Ścieżka
     LLM pomija nieczytelne commity zamiast przerywać.
  4. Nowy kod `MAPPING_NEW_SIDE_ONLY` (kontrakt §3), kwalifikator wiersza
     jak `MAPPING_PARTIAL` (`findings.MappingQualifier`, nie trafia do
     ustaleń per instancja); `SOURCE_UNAVAILABLE` po raz pierwszy
     emitowany. v2: `mapping.fix_commits`, `mapping.new_side_only`.
     `vulnmap.ResolverVersion` 2→3 (klucze wyprowadzeń F3). Log CLI:
     liczba mapowań tylko z nowej strony i nieczytelnych commitów.
- **Weryfikacja:** `internal/javadiff` (obie strony, rename, plik
  dodany/usunięty, hunk z samymi `-`, linia `--- x`, `DeclaredMethods`,
  metoda jednolinijkowa); `internal/vulnmap/f5_test.go` na fałszywym
  GitHubie: usunięta metoda jest kandydatem, helper nie; brak metadanych
  → nowa strona z powodem; merge commit; rename + usunięty + dodany plik;
  brak źródeł; nieczytelny commit per CVE i obok czytelnego; offline bez
  `.diff` → `ErrOffline`; offline bez metadanych → nowa strona; token
  wysyłany. `internal/reach/f5_test.go`: kody na wierszu i ich brak w
  ustaleniach per instancja; `SOURCE_UNAVAILABLE`. Mutacje: mapowanie
  także metod `added` (2 testy padają), parser ignorujący linie `-`
  (5 testów padają). Istniejące testy nie sięgają już do prawdziwego
  `api.github.com` (`GithubAPIBaseURL` wskazuje fałszywy serwer).
  `go vet ./...`, `go test ./...` — PASS.
- **Prawdziwe dane** (wejścia baseline CVE:F0, `-profile spring`):
  - offline na dotychczasowym `.cache` (bez metadanych): każde mapowanie
    z diffu dostaje `MAPPING_NEW_SIDE_ONLY` (petclinic 90, WebGoat 67);
    listy metod bez zmian poza poprawką jednolinijkowych metod (3/4 CVE);
  - metadane pobrane raz online (token z `gh auth token`, tylko w
    środowisku) i przeniesione do kopii starego cache **wyłącznie** jako
    wpisy `api.github.com` i `raw.githubusercontent.com` (OSV bez zmian,
    ten sam zestaw CVE 148/160); potem offline: petclinic 234 commity
    (227 z rodzicem, 7 merge), WebGoat 125 (110/15); metody: 232/402
    tylko-dodane usunięte z mapowań, 120/116 sprzed poprawki dodane;
    mapowanych CVE 91→92 i 104→105;
  - zmiany werdyktów (każda sprawdzona w `fix_commits`):
    CVE-2024-12798 UNKNOWN→FRAMEWORK na obu (logback usuwa
    `JaninoEventEvaluator*`; ścieżka do `JaninoEventEvaluatorBase#evaluate`
    — kryterium „usunięta metoda jest kandydatem”); CVE-2024-38821
    REACHABLE→FRAMEWORK na WebGoat (stara ścieżka kończyła się w
    `AbstractServerWebExchangeMatcherRegistry#matchers` z pliku *dodanego*
    przez linkowany commit z 2017 r.; teraz zmieniona
    `WebFilterChainProxy#filter` — kryterium „helper nie jest podatnym
    kodem”); CVE-2026-50193 UNKNOWN→NOT_REACHABLE na obu (zmieniony
    jednolinijkowy konstruktor `InternalNodeMapper#<init>` teraz
    znaleziony; `push`/`popOrNull` to dodane helpery, nieobecne w
    bytecode; wiersz ma `MAPPING_PARTIAL` i `MODEL_INCOMPLETE`);
  - v2 bajtowo powtarzalny (dwa przebiegi petclinic);
  - lokalny `.cache` uzupełniony o 1064 wpisy GitHub (API + raw), bez
    zmiany wpisów OSV — baseline F0 nadal odtwarzalny z eksportu.
- **Kryteria F5 a ten wycinek:** usunięta podatna metoda jest kandydatem —
  tak (test + CVE-2024-12798); nowy helper nie jest uznany za istniejący
  kod — tak (test + CVE-2024-38821); rename pliku, merge commit, brak
  źródeł — odrębne wyniki tak; **dwie gałęzie — nie** (F5b, F5.7);
  przypadki nierozstrzygnięte widoczne — tak (`MAPPING_NEW_SIDE_ONLY`,
  `SOURCE_UNAVAILABLE`).
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz mapowań), README krok 6,
  FAQ (krok 6), `docs/RESULTS_CONTRACT.MD` §3, legenda w szablonie
  raportu, `testdata/cve-control/cases.yaml` (`v1` dla 0201–0203 i
  CVE-2021-44228), PLAN.md (Phase 4), nagłówek F5 w planie CVE.
- **Ograniczenia:** metody z diffu nadal po nazwie, bez deskryptora,
  metody klas zagnieżdżonych przypisane do klasy najwyższego poziomu;
  `added` jest względne wobec rodzica commita — commit z dawną funkcją
  podlinkowany w advisory nadal może dać metody `changed`, które nie są
  podatnym kodem (powiązanie z wydaniem: F5b); metody abstrakcyjne/
  interfejsowe bez ciała nie są liczone jako zadeklarowane; bez tokena
  limit API 60/h (wynik: nowa strona z powodem, nie błąd);
  `MAPPING_NEW_SIDE_ONLY` nie zmienia werdyktu (jak `MAPPING_PARTIAL`).
- **Następny krok (wymaga zgody):** CVE:F5b — weryfikacja kandydatów w
  bytecode z deskryptorem i sources JAR (F5.5, F5.8), powiązanie commita
  z gałęzią/wydaniem i odrębne rodzaje odnośników (F5.7), heurystyka
  przemianowań (F5.9), obie strony w prompcie LLM bez uruchamiania LLM
  (F5.11).

### CVE:F5b — weryfikacja w bytecode, klasy zagnieżdżone, dopasowanie commita do wydania — DONE for this slice; CVE:F5 IN_PROGRESS (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** F5.5 (klasy zagnieżdżone, parametry
  ze źródła, deskryptor z bytecode bez zgadywania), F5.8 (weryfikacja w
  klasach badanego artefaktu; bez sources JAR), F5.7 w wariancie lekkim
  (typ referencji OSV, dopasowanie commita po obecności kandydatów), F5.9
  (heurystyka przemianowań, tylko dowody). Odłożone za zgodą: pełne
  powiązanie z gałęzią przez tagi git (GitHub compare) i obie strony w
  prompcie LLM (F5.11: zmiana wersji promptu = ponowne płatne pytania).
- **Zrealizowane:**
  1. `internal/javadiff`: `MethodRef{Class, Name, Params, StartLine,
     EndLine}`, `MethodsForLines`, `Declared`, `Package`. Klasy członkowskie
     jako `Outer$Inner` (także druga klasa najwyższego poziomu w pliku);
     linie klas anonimowych/lokalnych i lambd należą do metody otaczającej
     (nazwy `Outer$1` zależą od numeracji kompilatora — zmieniona
     oczekiwana wartość jednego starego testu, z komentarzem); parametry:
     proste nazwy, tablice, varargs, generyki wymazane, zmienne typu `?`.
  2. `internal/vulnmap`: `FixFile.Methods` jako lista `FixMethod{selector,
     kind, source_params, bytecode, possible_rename}` (pole v2 z F5a
     zmienione z mapy na listę, przed wydaniem); boki dopasowywane po
     `Class#name(params)` (zmiana sygnatury = removed + added);
     `Resolver.Bytecode` → `verify`: jedno pasujące przeciążenie →
     selektor z deskryptorem, kilka → `ambiguous`, żadne →
     `params_mismatch`, brak metody/klasy → `method_missing`/
     `class_missing`; konstruktory klas wewnętrznych/enumów z
     syntetycznymi parametrami wiodącymi. `noteReleaseFit`: commit bez
     obecnego kandydata pomijany w mapowaniu tylko, gdy inny commit ma
     obecnego; bez żadnego — wszyscy kandydaci zostają (`release_fit`
     puste). `collapseSelectors`: selektor po nazwie wygrywa z dokładnym
     dla tej samej `C#m`. `RefType` z referencji OSV (FIX, jeśli
     którakolwiek referencja commita jest FIX). `possible_rename`:
     removed↔added w tej samej klasie, Jaccard identyfikatorów ciała ≥ 0,6,
     tylko jednoznaczna para.
  3. `cmd/ariadne`: `declaredOverloads` z hierarchii (bez metod bridge i
     syntetycznych). `vulnmap.ResolverVersion` 3→4.
- **Weryfikacja:** `internal/javadiff` (klasy zagnieżdżone, parametry,
  lokalne/anonimowe/lambda, zakresy linii), `internal/vulnmap/f5b_test.go`
  (dokładny/ambiguous/params_mismatch/method_missing/class_missing; dwa
  commity — jeden poza wydaniem pominięty, bez obecnych brak wniosku; typ
  referencji; klasa zagnieżdżona + przemianowanie; `paramsCompatible`;
  `collapseSelectors`), `cmd/ariadne` `TestDeclaredOverloads`. Mutacje:
  zawsze pierwsze przeciążenie jako dokładne, ignorowanie `release_fit`,
  klasa bez ścieżki `$` — testy padają. `go vet ./...`, `go test ./...`
  — PASS.
- **Prawdziwe dane** (offline, ten sam `.cache` dla F5a i F5b, wejścia
  baseline, `-profile spring`): weryfikacja kandydatów petclinic 911
  exact / 1 ambiguous / 80 params_mismatch / 87 method_missing / 130
  class_missing, WebGoat 791/2/36/82/223; kandydaci w klasach
  członkowskich 191/194; commity pominięte przez `release_fit` 9/4;
  `possible_rename` 152/113 oznaczeń (tylko dowody). Typ referencji: 0 z
  337 linków do commitów w rekordach GHSA ma `FIX` (wszystkie `WEB`).
  Zmiany werdyktów (sprawdzone w `fix_commits`): CVE-2023-20873
  NOT_REACHABLE→FRAMEWORK na obu (zmieniona metoda w
  `ReactiveCloudFoundryActuatorAutoConfiguration$WebFilterChainPostProcessor`,
  wcześniej szukana w klasie zewnętrznej); CVE-2026-50193
  NOT_REACHABLE→UNKNOWN + METHOD_NOT_FOUND na obu (poprawka w
  `InternalNodeMapper$WrapperForSerializer`, nieobecnej w badanym
  jackson; `<init>` z F5a był konstruktorem nowej klasy zagnieżdżonej
  przypisanym do zewnętrznej); CVE-2026-41848 REACHABLE→FRAMEWORK na
  WebGoat (poprawka w `AntPathMatcher$AntPathStringMatcher#matchStrings`,
  nie w konstruktorze `AntPathMatcher`). 12/10 wierszy zmienia tylko
  ścieżkę (cel w klasie członkowskiej lub konkretne przeciążenie). v2
  bajtowo powtarzalny na obu aplikacjach.
- **Kryteria F5 po F5a+F5b:** usunięta metoda — kandydat (tak); helper —
  nie (tak); rename pliku, merge commit, brak źródeł — odrębne wyniki
  (tak); dwie gałęzie — częściowo: commit z innej linii wydań pomijany,
  gdy żaden jego kandydat nie istnieje w badanych klasach; commit z innej
  linii, którego metody też istnieją, nadal się dolicza; przypadki
  nierozstrzygnięte widoczne (tak).
- **Dokumentacja:** `docs/SUPPORT.MD`, README krok 6, FAQ krok 6,
  `docs/RESULTS_CONTRACT.MD` §5 (pola `fix_commits`), `cases.yaml`
  (0203), nagłówek F5 w planie CVE, ten rejestr.
- **Ograniczenia:** porównanie typów po prostych nazwach (typ o tej samej
  prostej nazwie z innego pakietu nie jest odróżniony); nagłówki z
  argumentami adnotacji w parametrach nie są rozpoznawane jako metody
  (wcześniejsze ograniczenie skanera); metody klas anonimowych/lokalnych
  mapowane na metodę otaczającą; brak sources JAR; typ referencji nie
  różnicuje danych GHSA.
- **Następny krok (wymaga zgody):** (zrealizowano wariant (a) jako CVE:F5c,
  zob. niżej) do wyboru — (a) CVE:F5c: powiązanie
  commita z linią wydań przez tagi git (GitHub compare: czy tag badanej
  wersji jest przodkiem rodzica) i obie strony w prompcie LLM (bez
  uruchamiania LLM; świadoma zmiana wersji promptu), po czym zamknięcie
  F5; albo (b) uznanie F5 za wystarczające i przejście do CVE:F6
  (komponenty osadzone) z F5.7-pełnym jako znanym ograniczeniem.

### CVE:F5c — linia wydań z historii git, obie strony poprawki w prompcie LLM — DONE; CVE:F5 DONE (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** F5.7 w pełni (powiązanie commita
  z linią wydań przez tagi git, GitHub compare) i F5.11 (obie strony
  poprawki i zakres wydania w prompcie LLM, świadoma zmiana wersji
  promptu, bez uruchamiania LLM).
- **Zrealizowane:**
  1. `internal/vulnmap/releaseline.go`: tag badanej wersji (z
     `cve.Affected`) szukany po dokładnej nazwie (`v{V}`, `{V}`,
     `{artifactId}-{V}`, `{repo}-{V}`, `release-{V}`, `v_{V}`,
     `version-{V}`) przez `git/matching-refs/tags/{name}` (200 z pustą
     listą = brak taga, odpowiedź cache'owalna; zapytanie prefiksowe
     zwraca też `v1.2.30` — liczy się tylko dokładne trafienie).
     Porównanie `compare/{tag}...{sha}?per_page=1&page=2` (strona 2:
     GitHub nie wysyła listy plików, ≈10 KB zamiast do ≈1 MB — pierwsza
     wersja ze stroną 1 zapisała 342 MB w cache; usunięte, pobrane
     ponownie). `ahead` → `descends_from_release`, `behind`/`identical`
     → `in_release`, `diverged` → `other_line`, z wyjątkiem: `behind_by
     == 1` i merge base = rodzic commita taga → `descends_from_release`
     z notatką `NoteOffBranchTag` (Tomcat „Tag 9.0.45”, Spring „Release
     v5.3.6” — commit taga poza gałęzią; bez tego wszystkie 192 commity
     Tomcata 9.0.x były „diverged”). Kilka wersji z tagami → nieznane
     (mapowanie jest per advisory). Błąd/brak w cache → nieznane z
     powodem, nigdy błąd globalny (jak metadane commita w F5a).
  2. `FixCommit`: `release_tag`, `release_line`, `release_line_reason`,
     `excluded` (`release_line`/`release_fit`). Kolejność: najpierw linia
     (commit `other_line` wykluczony, gdy inny commit advisory
     `descends_from_release`), potem reguła F5b liczona tylko względem
     niewykluczonych commitów. **Korekta w trakcie:** pierwsza wersja
     dawała historii pierwszeństwo nad regułą fit dla commitów
     `descends_from_release`; na prawdziwych danych dołożyło to metody
     Jacksona 3 (`tools.jackson.*`) do CVE jackson-databind 2.11.4 —
     w repozytoriach scalanych do przodu następna linia główna
     „pochodzi” od każdego starszego taga. Teraz `descends` nie
     wyłącza reguły fit.
  3. Prompt LLM v2 (`promptVersion` 1→2): sekcja `<fix_methods>` (per
     commit: metody changed/removed/added z obu stron, obecność w
     klasach, linia wydań, „not the fix of the analyzed release”),
     notatka na linii `# commit` w diffie dla commitu z innej linii,
     reguła 8 promptu systemowego (nie wskazywać metod `added` ani
     commitów oznaczonych jako nie-poprawka badanego wydania);
     `docs/phase4-llm.md` zsynchronizowany (test
     `TestSystemPromptMatchesApprovedDoc`). Cache LLM: stare odpowiedzi
     nie pasują do nowego klucza — ponowne pytania tylko z `-llm`.
  4. `vulnmap.ResolverVersion` 4→5; log przebiegu: liczniki linii i
     wykluczeń (`releaseLineSummary` w `cmd/ariadne`).
- **Weryfikacja:** `internal/vulnmap/f5c_test.go` — wykluczenie innej
  linii przy dokładnym tagu (prefiks `v1.2.30` nie liczy się), brak
  commitu `descends` → nic nie wykluczone (`in_release` tylko dowód),
  brak taga / kilka wersji / offline bez wpisów → nieznane bez błędu,
  interakcja z regułą fit (w obie strony), tag poza gałęzią (oraz
  `behind_by` 1 z innym merge base i `behind_by` 5 z merge base =
  rodzic → `other_line`), złoty prompt `user_d.golden`, notatka dla
  skróconego SHA w diffie dla LLM. Fałszywy serwer odpowiada tylko na
  `per_page=1&page=2`. Mutacje: dopasowanie prefiksowe taga, wykluczanie
  bez commitu `descends`, `diverged` jako `descends`, fit ignorujący
  wykluczenie, reguła off-branch bez `behind_by == 1` i bez sprawdzenia
  merge base — każda łapana. `go vet ./...`, `go test ./...` — PASS.
- **Prawdziwe dane** (wejścia baseline CVE:F0, `-profile spring`,
  `-json-format v2`): przed = HEAD F5b offline na `.cache`; wpisy tagów i
  compare pobrane raz online (`GITHUB_TOKEN` z `gh auth token`, tylko w
  środowisku, kopia cache z `-data-ttl 87600h`) i przeniesione do
  `.cache` **wyłącznie** jako wpisy `git/matching-refs` (116) i
  `compare` (326) — przebieg online odświeżył też 157 rekordów OSV i
  dodał 227 nowych w kopii; tych nie przeniesiono, zestaw CVE bez zmian
  (offline przed i po: petclinic 148 advisories / 155 ustaleń, WebGoat
  160 / 168). `.cache` 89→94 MB. Potem offline:
  - petclinic: 234 czytelne commity → 116 `descends` (83 przez tag poza
    gałęzią), 3 `in_release`, 109 `other_line`, 6 nieznane (brak taga:
    log4j2, snakeyaml, netty, http2, commons-fileupload); wykluczone 98
    przez linię (42 advisories), 7 przez fit;
  - WebGoat: 125 → 95 (41) / 5 / 20 / 5; wykluczone 9 linią, 4 fitem;
  - zmienione mapowania 7/2 advisories, wyłącznie zawężenia, każde
    sprawdzone w `fix_commits`: CVE-2023-46589 (Tomcat 9.0.45) 13 → 1
    metoda — zostaje poprawka 9.0.x `InputBuffer#realReadByte`, odpadają
    `Request#getParameter*` itd. z commitu „Implement parameter error
    handling changes” na linii 10.x/main (`behind_by` 1107);
    CVE-2023-24998 bez `Parameters#addParameter` (commit z innej linii);
    CVE-2022-42003 33 → 29 (commit z gałęzi 2.12 rozbieżny z 2.11.4);
    CVE-2022-22970, CVE-2024-24549, CVE-2025-24813, CVE-2025-53506 —
    selektor po nazwie z commitu innej linii znika, zostaje dokładny z
    deskryptorem;
  - werdykty bez zmian (petclinic 64/6/18/3/57, WebGoat 48/11/40/61),
    jedna ścieżka FRAMEWORK zmieniona (CVE-2023-46589, 24 → 29 kroków);
  - v2 bajtowo powtarzalny (dwa przebiegi offline na każdej aplikacji).
  - LLM nie uruchomiony (płatny); prompt sprawdzony testami złotymi.
- **Kryteria F5:** usunięta metoda — kandydat (F5a); helper — nie (F5a);
  rename pliku, merge commit, brak źródeł — odrębne wyniki (F5a);
  **dwie gałęzie — tak**: commit rozbieżny z tagiem badanej wersji jest
  pomijany, gdy inny commit pochodzi od taga (test + CVE-2023-46589);
  przypadki nierozstrzygnięte widoczne (`release_line_reason`, brak
  wykluczenia bez rozstrzygnięcia). F5 = DONE.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz mapowania), README krok 6
  (i źródło LLM), FAQ krok 6, `docs/RESULTS_CONTRACT.MD` §5 (pola
  `fix_commits`), `docs/phase4-llm.md` (prompt v2), `cases.yaml` (0203),
  nagłówek F5 w planie CVE, ten rejestr.
- **Ograniczenia:** tagi tylko według wymienionych wzorców (np. XStream
  `XSTREAM_1_4_19`, PostgreSQL `REL42.3.3` → nieznane); tylko GitHub;
  `descends_from_release` ≠ „ta sama linia” w repozytoriach scalanych do
  przodu (dlatego nie wyłącza reguły fit); jeden tag na advisory;
  commit z innej linii zostaje, gdy żaden nie pochodzi od taga; dane
  historii wymagają zapytań GitHub API (token zalecany); brak sources
  JAR (opcjonalne zadanie 8); nowy prompt bez płatnej weryfikacji na
  prawdziwych CVE.
- **Następny krok (wymaga zgody):** CVE:F6 — komponenty osadzone bez
  relokacji (zachowanie wszystkich `pom.properties`, inwentarz
  komponentów osadzonych, potwierdzanie obecności po klasach).

### CVE:F6a — komponenty osadzone potwierdzone klasami — DONE for this slice; CVE:F6 IN_PROGRESS (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** F6.1, F6.2, F6.4 (referencja z
  Maven Central — zgoda na pobieranie), F6.6, F6.8, F6.9, F6.7 dla wersji
  potwierdzonych; reguła „moduł tego samego projektu”. Poza zakresem
  (F6b): F6.3 SBOM, F6.5 zbiory wersji kandydujących i kandydaci bez
  `pom.properties`.
- **Zrealizowane:**
  1. `internal/component/embedded.go`: kandydaci (`Inventory.Embedded`)
     z każdego dodatkowego `pom.properties` JAR-a zależności;
     `isModuleOf` (ta sama wersja, ta sama grupa albo zagnieżdżona) →
     `module`, bez zapytań; reszta `library` → referencyjny JAR
     `repo1.maven.org/maven2/<g>/<a>/<v>/<a>-<v>.jar`
     (`ARIADNE_MAVEN_CENTRAL` zmienia korzeń) przez cache; porównanie
     wpisów `.class` (bez `META-INF/`, `module-info`): nazwa, potem bajty
     (CRC-32 + rozmiar) albo `classfile.Fingerprint`. Statusy:
     `confirmed` (≥ 50% klas referencji obecnych i ≥ 90% z nich takich
     samych), `content_differs`, `partial`, `classes_absent`,
     `reference_unavailable` (błąd pobrania — także brak w cache offline
     — nigdy błąd globalny). Potwierdzony → `Instance` z `Detection
     embedded`, `HostInstance`, `EmbeddedPath`; własność klas na
     poziomie klasy (`byClass`, `InstanceForClass`), `recount` przelicza
     liczby klas, pakiety, właścicieli konfliktów i `Shadowed`.
     `InstanceLocation`: `<lokalizacja hosta>!/META-INF/maven/<g>/<a>`.
  2. **Korekta w trakcie:** pierwsza wersja porównywała tylko bajty —
     na prawdziwych danych 0 z 656 klas jnr-ffi w jruby-core było
     bajtowo identycznych (ta sama długość, inna kolejność puli stałych:
     maven-shade przepisuje klasy przez ASM, gdy w JAR-ze jest jakakolwiek
     relokacja). Dodany `internal/classfile/fingerprint.go`: nagłówek,
     pola, metody (posortowane), instrukcje z operandami puli zastąpionymi
     tym, co nazywają, `ldc_w`/`goto_w`/`jsr_w` zwinięte, cele skoków
     jako indeksy instrukcji; bez atrybutów innych niż Code, tablic
     wyjątków i tablic switch. Wynik: 644/656 takich samych, 12 różnych
     (odwołania do relokowanego ASM).
  3. `internal/httpcache`: JAR wydania z Maven Central (nie SNAPSHOT) =
     `KindPinned`.
  4. `cmd/ariadne`: `ResolveEmbedded` po otwarciu cache, log liczników,
     GAV-y potwierdzonych instancji dodane do zapytania OSV (z
     deduplikacją listy GAV); `snapshot.go` używa `InstanceLocation`.
  5. v2: `components[].embedded_in`, detekcja `embedded`, lokalizacja
     instancji osadzonej; `coverage.embedded_candidates`. Raport: tabela
     „Embedded component candidates”, kolumna lokalizacji instancji,
     zaktualizowany opis sekcji Components.
- **Weryfikacja:** `internal/classfile/fingerprint_test.go` (klasa z
  dopisaną stałą — inne bajty, ten sam odcisk; inna klasa i zmieniony
  kod — inny; `ldc` vs `ldc_w` z przesuniętym skokiem — ten sam),
  `internal/component/embedded_test.go` (potwierdzona przez odcisk,
  własność klas, lokalizacja; `classes_absent`, `content_differs`,
  `partial`, `reference_unavailable`, moduł bez zapytania; dwie wersje →
  dwie instancje, konflikt przypisany instancjom osadzonym),
  `internal/reach/f6_test.go` (advisory przeciw bibliotece osadzonej →
  ustalenie REACHABLE na instancji osadzonej). Mutacje: tylko CRC,
  brak reguły modułu, brak progu `partial`, brak sprawdzenia treści,
  własność na poziomie JAR-a, surowe offsety skoków, `ldc_w` bez
  zwinięcia — każda łapana. `go vet ./...`, `go test ./...` — PASS.
- **Prawdziwe dane** (wejścia baseline CVE:F0, `-profile spring`,
  `-json-format v2`): referencje pobrane raz online na kopii cache
  (6 JAR-ów z Central); nowa lista GAV (z jnr-ffi, sizeof,
  terracotta-utilities-tools) zmienia klucz `querybatch` OSV, więc
  porównanie przed/po wykonano offline na **tej samej** kopii po
  uzupełnieniu (HEAD F5c vs F6a): petclinic 153/160, WebGoat 164/172
  advisories/ustaleń po obu stronach, werdykty identyczne, 0 ustaleń na
  instancjach osadzonych (żadna z trzech wersji nie ma advisory w OSV).
  Kandydaci: petclinic — `ehcache-3.9.2`: 5 modułów,
  `org.ehcache:sizeof:0.4.0` (66/66) i
  `org.terracotta:terracotta-utilities-tools:0.0.6` (22/22)
  potwierdzone, `offheap-store`/`statistics` `classes_absent`
  (relokowane do `org.ehcache.shadow`); byte-buddy: `byte-buddy-dep`
  moduł; WebGoat — `com.github.jnr:jnr-ffi:2.2.1` w
  `jruby-core-9.2.14.0` potwierdzony (656 klas, 644 takich samych),
  `jitescript` `classes_absent`. v2 bajtowo powtarzalny. `.cache`
  uzupełniony o 252 nowe wpisy (6 JAR-ów z Central, 229 OSV — nowy
  `querybatch` i rekordy, 3 diffy GitHub), bez nadpisywania istniejących;
  offline na `.cache`: petclinic 148→153, WebGoat 160→164 advisories —
  **wyłącznie** przez świeższą odpowiedź `querybatch` (Tomcat 9.0.45,
  undertow 2.2.4, jackson-databind 2.11.4), nie przez komponenty osadzone.
- **Kryteria F6:** osadzona biblioteka z zachowanymi pakietami dostaje
  odrębny wynik — tak (test silnika; na danych realnych instancje są,
  ustaleń brak, bo brak advisories); stare metadane bez klas nie dają
  identyfikacji — tak (`classes_absent`, test + 3 przypadki realne); dwie
  kopie w różnych wersjach nie są scalane — tak (test); pominięte lub
  nieznane wersje widoczne w pokryciu — tak (`embedded_candidates`).
  F6 pozostaje IN_PROGRESS: brak SBOM (F6.3) i zbiorów wersji
  kandydujących (F6.5).
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz tożsamości komponentów),
  README (krok rozpoznawania zależności, opis raportu, ograniczenie
  „Shaded/relocated”), FAQ krok 1, `docs/RESULTS_CONTRACT.MD`
  (`components[]`, `coverage.embedded_candidates`), szablon raportu,
  `cases.yaml` (0301, 0302), nagłówek F6 w planie CVE, ten rejestr.
- **Ograniczenia:** kandydaci tylko z `pom.properties`; sprawdzana tylko
  zadeklarowana wersja; referencja musi być na Maven Central pod tym
  GAV; progi stałe; klasy niepotwierdzonego kandydata zostają przy
  hoście; reguła modułu ufa grupie i wersji; relokacja to F7; odcisk nie
  obejmuje adnotacji, sygnatur generycznych, tablic wyjątków i switch.
- **Następny krok (wymaga zgody):** CVE:F6b (import SBOM CycloneDX JSON,
  wersje kandydujące / nieznane dla kandydatów `content_differs`) albo
  CVE:F7 (jawna relokacja — trzy realne przypadki `classes_absent` z
  F6a to gotowy materiał testowy).

### CVE:F7a — jawna relokacja pakietów — DONE; CVE:F7 DONE dla jawnej relokacji (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** F7.1 (plik `-embedded-components`,
  YAML, przypięty do SHA-256), F7.2 (konfiguracja builda — relokacje
  maven-shade z `pom.xml` JAR-a), F7.3–F7.6. Poza zakresem: F7.7
  (automatyczne odkrywanie relokacji — w planie osobny etap
  eksperymentalny) i ASM w jruby (bez `pom.properties`, materiał dla F6b).
- **Zrealizowane:**
  1. `internal/relocate`: reguły `from → to` (prefiks tylko na granicy
     pakietu, najdłuższy wygrywa, sprzeczne reguły odrzucone), nazwy
     binarne i wewnętrzne, deskryptory (tablice, klasy wewnętrzne),
     selektory, literały nazw klas, odwrotność, `Under`.
  2. `classfile.FingerprintWith`: odcisk z przepisaniem nazw w
     nagłówku, deskryptorach, referencjach puli i literałach;
     `RelocateClassForTesting` (przepisanie Utf8 w puli, bez ASM/JDK) do
     fixture'ów.
  3. `internal/component`: `shadeconfig.go` (każda `<relocation>`
     wtyczki `maven-shade-plugin`, wszystkie executions/profile, pomija
     placeholdery/wildcardy z notatką, ignoruje inne wtyczki);
     `declarations.go` (`LoadDeclarations` — ścisły schemat, wersja 1;
     `applyDeclarations` — dopięcie do kandydata albo nowy kandydat
     zadeklarowany, problemy: nieznany hash / niezgodny hash dla
     nazwanego JAR-a). **Korekta w trakcie:** zadeklarowane reguły
     najpierw były per kandydat; klasy jednego relokowanego komponentu
     odwołują się do innych relokowanych pakietów tego samego JAR-a, więc
     reguły są teraz cechą JAR-a (build-config ∪ deklaracje wszystkich
     komponentów JAR-a). `ResolveEmbedded`: obrazy klas referencji pod
     przeniesionymi nazwami, porównanie przez odwrotną relokację,
     `collision` (klasa przypisana już innemu potwierdzonemu
     kandydatowi), `relocation_conflict`; instancja z `Relocations`
     (reguły, pod którymi są jej własne klasy, ze źródłem).
  4. `internal/reach.matchSpecs`: dla instancji z relokacją selektory
     mapowania dopasowywane też po przepisaniu (wariantowe też).
  5. `cmd/ariadne`: flaga `-embedded-components`, plik w wejściach
     snapshotu, log (potwierdzone/relokowane, notatki z pom.xml),
     problemy deklaracji jako ostrzeżenia. v2: `components[].relocations`,
     `embedded_candidates[].declared/relocations`,
     `coverage.embedded_declaration_issues`. Raport: kolumna relokacji
     (oryginał → w JAR-ze, źródło), lista niezastosowanych deklaracji.
  6. `testdata/embedded/petclinic-ehcache.yaml` — deklaracja dla
     ehcache-3.9.2 (Gradle Shadow, brak konfiguracji w JAR-ze).
- **Weryfikacja:** `internal/relocate/relocate_test.go` (granica
  prefiksu, najdłuższa reguła, deskryptory, selektory, literały,
  odwrotność, konflikty, niepoprawne reguły),
  `internal/component/relocation_test.go` (bez deklaracji
  `classes_absent` z podpowiedzią; z deklaracją potwierdzony z regułą i
  źródłem, własność klas; z `pom.xml` bez deklaracji, z ignorowaniem
  relokacji innej wtyczki i notatką o placeholderze; zła mapa, niezgodny
  i nieznany hash, konflikt build-config vs deklaracja, kopia
  zminimalizowana `partial`, kolizja dwóch kandydatów, odrzucane pliki
  deklaracji), `internal/reach/f6_test.go` `TestRelocatedInstanceFinding`
  (aplikacja i biblioteka relokowane `reachfixture` →
  `shaded.reachfixture`; selektor oryginalny → REACHABLE na instancji
  osadzonej, ścieżka przez metodę relokowaną). Mutacje: brak
  przepisania selektorów, brak odwrotnej relokacji w odcisku, relokacje
  dowolnej wtyczki, brak kontroli kolizji, prefiks bez granicy pakietu,
  ignorowanie hasha deklaracji — każda łapana; „zachowaj placeholdery”
  przeżywa jako mutant równoważny (walidacja nazwy pakietu i tak je
  odrzuca z notatką). `go vet ./...`, `go test ./...` — PASS.
- **Prawdziwe dane** (`-profile spring`, `-json-format v2`, przed = HEAD
  F6a, po = F7a; nowa lista GAV zmienia klucz `querybatch` OSV, więc
  porównanie offline na kopii cache uzupełnionej raz online; do `.cache`
  przeniesione tylko 2 nowe wpisy `querybatch`):
  - petclinic z `-embedded-components testdata/embedded/petclinic-ehcache.yaml`:
    `org.terracotta:offheap-store:2.5.1` 167/167 i
    `org.terracotta:statistics:2.1` 130/130 (pakiety `statistics` i
    `context`) potwierdzone pod `org.ehcache.shadow.org.terracotta.*`;
    wszystkie 4 kandydaci-biblioteki ehcache potwierdzone, 0
    niepotwierdzonych;
  - WebGoat bez deklaracji: 2 reguły maven-shade z
    `META-INF/maven/org.jruby/jruby-core/pom.xml`;
    `me.qmx.jitescript:jitescript:0.4.1` potwierdzony 9/9 pod
    `org.jruby.me.qmx.jitescript`; jnr-ffi 656/656 identycznych (było
    644 — 12 klas odwołujących się do relokowanego ASM);
  - advisories/ustalenia i werdykty identyczne przed/po (petclinic
    153/160, WebGoat 164/172): żadna z nowych wersji nie ma advisory w
    OSV; v2 bajtowo powtarzalny; offline na `.cache` — te same liczby.
- **Kryteria F7:** relokowane klasy i deskryptory prowadzą do właściwych
  metod w grafie — tak (test silnika); raport pokazuje nazwy oryginalne
  i rzeczywiste oraz podstawę relokacji — tak; błędna mapa, kolizja,
  niezgodny hash, niejednoznaczność — bez potwierdzonego mapowania —
  tak (testy); zakres obejmuje jawną relokację, nie wszystkie shaded
  JAR-y — tak (SUPPORT, README).
- **Dokumentacja:** `docs/SUPPORT.MD`, README (flaga, krok rozpoznawania
  zależności, ograniczenie „Shaded/relocated”), FAQ krok 1,
  `docs/RESULTS_CONTRACT.MD`, szablon raportu, `cases.yaml`
  (SHADED-JACKSON-HADOOP, 0303, 0304), nagłówek F7 w planie CVE, ten
  rejestr.
- **Ograniczenia:** relokacja musi być jawna; `<includes>/<excludes>` i
  `rawString` shade nie modelowane; reguła z profilu pom.xml stosowana do
  całego JAR-a; weryfikacja kandydatów diffu (F5b) i grounding LLM
  szukają nazw oryginalnych; odcisk nie obejmuje adnotacji, sygnatur
  generycznych, tablic wyjątków i switch.
- **Następny krok (wymaga zgody):** CVE:F6b (SBOM, wersje kandydujące,
  ASM w jruby jako przypadek bez `pom.properties`), CVE:F8 (docelowy
  runtime/JDK) albo obszar spoza CVE z ROADMAP (np. GRAPH_PRECISION).

### CVE:F6b — SBOM, kandydaci z konfiguracji builda, zbiory wersji kandydujących — DONE; CVE:F6 DONE (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** F6.3 (import SBOM — CycloneDX
  JSON), F6.5 (wersja potwierdzona / zbiór kandydatów / nieznana),
  F6.7 dla zbiorów i wersji nieznanych, kandydaci bez `pom.properties`
  (ASM w jruby). Decyzje domyślne zaakceptowane: CycloneDX JSON
  (1.4–1.6), co najwyżej 8 porównywanych wydań na kandydata, pobieranie
  z Maven Central przez cache.
- **Zrealizowane:**
  1. `internal/component/versions.go`: lista wydań z `maven-metadata.xml`
     (cache, mutable; bez SNAPSHOT; porządek `mavenversion`), okno
     `versionWindow` (do 8 najbliższych zadeklarowanej; bez wersji —
     wszystkie tylko gdy ≤ 8, inaczej brak wyszukiwania),
     `searchVersions`: jedna pasująca → `confirmed` z wersją z klas,
     kilka → `version_candidates`, żadna → `version_unknown`. Szukane
     tylko po `content_differs` lub braku wersji; `partial`,
     `classes_absent`, `reference_unavailable` — nie. Brak listy wydań przy
     wersji zadeklarowanej zostawia `content_differs` z notatką.
  2. `embedded.go`: porównanie klas wydzielone do `compareVersion`
     (`versionCheck`, progi F6a bez zmian), `claims` po nazwie pakietu,
     `embeddedInstance` (ścieżki `build-config/…`, `sbom/…`, dowody),
     `Embedded.Source/SourceDetail/Version/VersionCandidates/
     VersionsChecked/VersionsPublished/VersionSearch`;
     `Instance.VersionCandidates`, `Instance.Label()` (`g:a:{v1,v2}`);
     `ConfirmedEmbeddedGAVs` — GAV dla każdej wersji kandydującej.
  3. `buildconfig.go`: z `pom.xml` hosta — `<artifactSet><includes>`
     (glob, `<excludes>`) wtyczki maven-shade ∩ `<dependencies>`
     (compile/runtime, bez optional/klasyfikatora), przechodnio przez pomy
     z Central (głębokość ≤ 4), `${…}` z właściwości, łańcucha rodziców i
     `dependencyManagement`; konfiguracja bez includes → brak kandydatów +
     notatka; dekoder XML czyta ISO-8859-1 (też w `shadeconfig.go`, co
     usunęło 3 błędy parsowania na WebGoat).
  4. `sbom.go`: `LoadSBOM` (tylko CycloneDX 1.4–1.6), `ApplySBOM`:
     powiązanie wyłącznie po SHA-256 z JAR-em zależności; zagnieżdżone
     komponenty Maven → kandydaci `sbom`; rozbieżna tożsamość lub GAV dla
     niezidentyfikowanego JAR-a → `issues`, tożsamość z JAR-a zostaje;
     reszta liczona i nieużywana.
  5. `internal/reach`: instancja z kandydatami dopasowana, gdy którakolwiek
     wersja jest w advisory; `markCandidates` dodaje `VERSION_CANDIDATES`;
     warianty mapowania tylko gdy wszystkie kandydatki wybierają ten sam
     (inaczej `VERSION_UNKNOWN`). Nowy kod `findings.VersionCandidates`.
  6. `internal/resultsv2`: `components[].version_candidates`, identity
     `version_candidates`; `embedded_candidates[]` z polami F6b;
     `coverage.sbom`; applicability `undetermined`, chyba że każda
     kandydatka jest objęta. `httpcache`: wydane pomy z Central są
     `KindPinned`.
  7. CLI: `-sbom` (plik jest wejściem snapshotu), log SBOM i liczników F6b;
     raport: kolumny „Named by” i „Version (classes)”, sekcja SBOM,
     poprawione ograniczenie „Shaded/relocated” w szablonie.
  8. `testdata/sbom/webgoat-jruby.cdx.json` — SBOM kontrolny (jruby-core po
     SHA-256; `asm-util` celowo jako 9.1; jeden komponent niepowiązany).
- **Weryfikacja:** `internal/component/f6b_test.go` (wersja z klas, zbiór,
  żadna, zadeklarowana pasuje bez wyszukiwania, partial bez wyszukiwania,
  okno, kolejność wydań; build-config: właściwość z rodzica,
  dependencyManagement, przechodnie, wykluczone/test/optional/spoza
  includes; brak includes; nierozwiązana wersja bez pobierania JAR-ów;
  SBOM: powiązanie, zagnieżdżenie, niepowiązane, rozbieżność, odrzucane
  pliki, purl), `internal/reach/f6_test.go`
  (`TestVersionCandidatesFinding`, `TestVersionCandidatesVariants`),
  `internal/resultsv2/f6b_test.go`; zaktualizowany licznik zapytań w
  `TestEmbeddedNotConfirmed`. `go vet ./...`, `go test ./...` — PASS.
  Mutacji nie uruchamiano osobno w tej sesji (testy obejmują przypadki
  ujemne każdej reguły).
- **Prawdziwe dane** (`-profile spring`, `-json-format v2`, przed = HEAD
  F7a, po = F6b, kopia `.cache` uzupełniona raz online; nowe wpisy
  dodane do `.cache` bez nadpisywania):
  - WebGoat: 5 kandydatów build-config z `jruby-core/pom.xml`
    (`asm.version` = 9.0 z `org.jruby:jruby-parent:9.2.14.0`; asm-tree i
    asm-analysis przez pom asm-commons) — `org.ow2.asm:asm`,
    `asm-commons`, `asm-util`, `asm-tree`, `asm-analysis` 9.0 potwierdzone
    pod `org.jruby.org.objectweb` (37+27+24+38+14 = wszystkie 140
    relokowanych klas ASM, identyczne). Z SBOM kontrolnym: `asm-util:9.1`
    nie pasuje (18/24), z 8 sąsiednich wydań pasują 9.0-beta i 9.0 (24/24)
    → `version_candidates {9.0-beta, 9.0}`; 1 komponent niepowiązany.
  - petclinic (z `testdata/embedded/petclinic-ehcache.yaml`): bez zmian;
    byte-buddy ma shade bez includes (notatka).
  - advisories/ustalenia i werdykty identyczne przed/po (petclinic
    153/160, WebGoat 164/172) — ASM 9.0/9.0-beta nie mają advisory w OSV;
    v2 bajtowo identyczny w dwóch przebiegach offline (WebGoat z SBOM),
    offline = online merytorycznie.
- **Kryteria F6:** odrębny wynik dla osadzonej biblioteki — tak (F6a);
  metadane bez klas nie dają potwierdzenia — tak; dwie kopie w dwóch
  wersjach nie są scalane — tak (F6a); pominięte/nieznane wersje widoczne
  w pokryciu — tak (`version_unknown`, `versions_checked`,
  `version_search`). Zadania F6.1–F6.9 wykonane; F6 DONE.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersze tożsamości i
  „Shaded/relocated components” — ten drugi był nieaktualny: „Not
  implemented”), README (krok rozpoznawania zależności, flaga `-sbom`,
  ograniczenie), FAQ krok 1, `docs/RESULTS_CONTRACT.MD` (oś tożsamości,
  applicability, `VERSION_CANDIDATES`, pola v2), szablon raportu,
  `cases.yaml` (0305–0307), nagłówek F6 w planie CVE, ten rejestr.
- **Ograniczenia:** tylko CycloneDX JSON (bez SPDX); konfiguracja builda
  tylko z maven-shade w `pom.xml` JAR-a (Gradle Shadow nie zostawia
  konfiguracji); okno 8 wydań, bez listy wydań offline brak wyszukiwania;
  zbiór kandydatów znaczy „klasy nie rozróżniają wersji”; mediacja wersji
  Mavena uproszczona (pierwsza znaleziona, BFS); profile pom i zakres
  `import` w `dependencyManagement` nie są modelowane; reguły F7a bez
  zmian (tylko jawna relokacja).
- **Następny krok (wymaga zgody):** CVE:F8a (docelowy runtime — potrzebny
  wybór pierwszego dostawcy JDK), CVE:F9 (kwalifikacja drugiego przyrostu
  F6–F7: m.in. atomowy zapis artefaktów i wspólny identyfikator analizy,
  błędy danych) albo obszar spoza CVE (GRAPH_PRECISION/RTA).

### CVE:F9a — integralność artefaktów wyniku, błędy danych, kontrole drugiego przyrostu — DONE for this slice; CVE:F9 IN_PROGRESS (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** F9.9 (atomowy zapis, wspólny
  identyfikator analizy), F9.8 (błędy danych — błąd globalny nie zostawia
  raportu wyglądającego na kompletny), F9.4/F9.5 (eksport v1 dla wielu
  instancji i wersji kandydujących; stabilność wyników). Praca na gałęzi
  `cve-f9a-qualification`. Format v1 bez zmian (brak pola na ID).
- **Zrealizowane:**
  1. Nowy pakiet `internal/atomicfile`: `Set.Add/Commit` — każdy plik
     zapisany w całości do pliku tymczasowego w swoim katalogu
     (`Sync`, `Close`, `0644`), dopiero potem `rename` każdego; błąd zapisu
     nie podmienia niczego i usuwa pliki tymczasowe; błąd między dwoma
     `rename` wymienia pliki już podmienione. `WriteFile` dla jednego pliku.
  2. `resultsv2`: pole `analysis_id` (SHA-256 treści dokumentu z pustym
     polem, bez czasu), `WithAnalysisID`, `Marshal`; `Write` atomowy.
     `report`: `ResultsJSON` (bajty v1), `Header.AnalysisID/ResultsFile/
     ResultsFormat/ResultsSHA256`, linia „Analysis ID” w nagłówku raportu;
     `WriteMarkdown`/`WriteResultsJSON` atomowe. `snapshot`: `Marshal`,
     `ManifestPath`, `Manifest.AnalysisID`; `Write` atomowy.
  3. `cmd/ariadne`: etap końcowy przebudowany — dokument v2 i kontrola
     v1/v2, potem ID, bajty wyników, raport z ID i SHA-256 pliku wyników,
     manifest snapshotu z ID, jeden `Commit`. **Naprawiony błąd:** wcześniej
     raport był zapisywany przed kontrolą v1/v2 i przed plikiem wyników —
     błąd w tych krokach zostawiał nowy raport bez pasującego JSON-a.
- **Weryfikacja:** `internal/atomicfile/atomicfile_test.go` (zapis całości,
  błąd zapisu nic nie podmienia i nie zostawia plików tymczasowych, błąd
  `rename` wymienia podmienione), `cmd/ariadne/outputs_test.go`
  (end-to-end na `testdata/reachfixture`: wspólne ID raportu/v2/manifestu i
  SHA-256 pliku wyników, to samo ID i bajtowo ten sam v2 w dwóch
  przebiegach, v1 z hashem w raporcie, nieudany zapis zostawia wcześniejszy
  raport, błąd OSV offline i uszkodzony / nieobsługiwany manifest nie
  zapisują nic), `internal/resultsv2/f6b_test.go` (`TestV1ExportWithCandidates`
  — eksport v1 z instancją potwierdzoną i `version_candidates` zgodny z
  agregatem silnika; `TestAnalysisID` — stabilne, zależne od treści).
  `go vet ./...`, `go test ./...` — PASS.
- **Prawdziwe dane** (offline na `.cache`, `-profile spring`, binarka z
  `main` 7dfe6ac vs gałąź): petclinic (z deklaracją ehcache) i WebGoat — v1
  bajtowo identyczny; v2 identyczny poza `analysis_id` i `tool.version`
  (`+dirty`); v2 bajtowo identyczny w dwóch przebiegach nowej binarki;
  ID w nagłówku raportu = `analysis_id`, SHA-256 w raporcie = hash pliku.
- **Przegląd F9.8 (błędy danych):** uszkodzony obiekt i edytowany manifest
  — odrzucane przez `-snapshot-in` (CVE:F3, testy `internal/snapshot`);
  nieczytelny i nieobsługiwany format manifestu — nowe testy CLI; niezgodny
  hash wejścia — CVE:F3; niedostępne OSV (offline bez cache, błąd sieci,
  HTTP 429) — błąd globalny `resolving CVEs`, teraz bez zapisu artefaktów;
  niedostępny commit poprawki — `SOURCE_UNAVAILABLE` per advisory (CVE:F5);
  częściowe mapowanie — `MAPPING_PARTIAL` (CVE:F1b). Rate limit OSV nie ma
  osobnego testu (ta sama ścieżka błędu co brak odpowiedzi).
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz raportów — także poprawiony
  nieaktualny opis „brak v2”; wiersz cache — poprawiony nieaktualny opis
  „brak TTL/snapshotów”), `docs/RESULTS_CONTRACT.MD` (`analysis_id`, zapis
  wspólny), README (rozdział o raporcie), nagłówek F9 w planie CVE, ten
  rejestr.
- **Ograniczenia:** v1 nie ma pola na ID (powiązanie tylko przez SHA-256 w
  raporcie); nieudany przebieg nie usuwa wcześniejszych plików (są starsze,
  nie częściowe); atomowość per plik na POSIX, nie dla zestawu przy awarii
  między `rename`; obiekty snapshotu zapisywane w trakcie (adresowane
  treścią, atomowe), manifest na końcu.
- **Następny krok (wymaga zgody):** reszta CVE:F9 przy kolejnym punkcie
  wydania (migracja części ręcznych mapowań do v2 — F9.3, porównanie
  aplikacji kontrolnych na wspólnym snapshotcie — F9.7), CVE:F8a (wymaga
  wyboru dostawcy JDK) albo GRAPH_PRECISION.

### CVE:F9b — porównanie aplikacji kontrolnych, migracja mapowań do v2, rate limit, testy bez sieci — DONE; CVE:F9 DONE dla F0–F7 (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** F9.7 (porównanie z bazą F0 na tych
  samych danych, każda zmiana wyjaśniona), F9.3 (CVE-2021-44228 i
  CVE-2022-42889 do schematu v2 z dokładnymi deskryptorami, weryfikacja w
  bytecode wydań z Central), F9.8 (rate limit / niedostępne OSV), F9.6
  (testy bez sieci). Gałąź `cve-f9b-qualification`.
- **Zrealizowane:**
  1. `scripts/cve-compare.sh OLD [NEW]` — buduje dwie wersje (worktree),
     uruchamia offline na tym samym `.cache` petclinic i WebGoat, wypisuje
     advisory tylko po jednej stronie i każdą zmianę statusu v1.
     Snapshot nie może być wspólnym źródłem (`-snapshot-in` odrzuca inną
     wersję analizatora), więc wspólnym źródłem danych jest cache; F0 na
     obecnym cache odtwarza bajtowo zapisaną bazę.
  2. `testdata/baseline/F9_COMPARISON.MD` — 45045d6 (F0) vs main e8ee758:
     petclinic 9 zmian + 5 nowych advisory, WebGoat 11 + 4; **wszystkie
     wyjaśnione**: (a) dane — nowe advisory pochodzą z nowszej odpowiedzi
     OSV, bo lista GAV (z komponentami osadzonymi F6/F7) zmienia klucz
     `querybatch`; potwierdzone na zapisanych odpowiedziach (dokładnie te
     GHSA, nic nie ubyło); (b) F1b `METHOD_NOT_FOUND` — dokładnie 7/6
     wierszy przewidzianych w `VERDICT_CHANGE_ESTIMATE.MD`; (c) F5 —
     CVE-2024-12798 (strona sprzed poprawki: usunięty plik), CVE-2023-20873
     (klasy zagnieżdżone), CVE-2024-38821 i CVE-2026-41848 REACHABLE →
     FRAMEWORK (zbyt szerokie mapowanie nazwami: 227 i 7 metod, ścieżki do
     metod niezwiązanych z poprawką; teraz 1 metoda zweryfikowana w
     bytecode, osiągalna przez framework). Żadna zmiana z grafu ani z
     komponentów osadzonych.
  3. `vuln_methods.yaml`: CVE-2021-44228 — warianty `[2.0-beta9,2.1)` (oba
     przeciążenia `JndiLookup#lookup`) i `[2.1,2.15.0)` (tylko
     `lookup(LogEvent,String)`); CVE-2022-42889 — `[1.5,1.10.0)`, 11
     instancyjnych przeciążeń `StringSubstitutor#replace` (3 statyczne nad
     mapą odrzucone). Sprawdzone w bytecode 17 wydań log4j-core i 7
     commons-text (jednorazowa sonda na `internal/classfile`, usunięta).
     `internal/vulnmap/f9b_test.go` przypina wybór wariantu dla każdej
     sprawdzonej wersji i brak pokrycia poza zakresem.
  4. `internal/cves/f9b_test.go`: HTTP 429 i 503 z OSV → błąd (nigdy pusta
     odpowiedź), błędna odpowiedź nie trafia do cache, po powrocie usługi
     advisory są.
- **Weryfikacja:** `go test ./...` z `HTTPS_PROXY`/`HTTP_PROXY` na
  zamknięty port i pustymi kluczami LLM/GitHub — PASS (żaden test nie
  sięga do sieci przez domyślny transport; testy httptest idą na
  127.0.0.1, wyłączony z proxy), `go vet ./...` — PASS. Petclinic i WebGoat
  v1 bajtowo identyczne z main (migracja ich nie dotyczy).
  `work/jpms-demo`: CVE-2022-42889 classpath NOT_REACHABLE / module-path
  REACHABLE jak przed migracją, `match_precision` name → exact.
- **Kryteria wydania F9 (przyrosty 1–2):** scenariusze kontrolne mają
  wyjaśnione wyniki (F9_COMPARISON, cases.yaml); niepewność nie jest
  bezpieczeństwem (`UNKNOWN`/`METHOD_NOT_FOUND`, `version_unknown`,
  `undetermined`); brak sieci offline i LLM w testach (sprawdzone);
  dowody i wejścia rozstrzygnięć (snapshot, `analysis_id`, evidence);
  dokumentacja rozróżnia działające/eksperymentalne/nieobsługiwane
  (SUPPORT).
- **Dokumentacja:** `docs/SUPPORT.MD` (mapowania wersjonowane),
  `docs/RESULTS_CONTRACT.MD` §6, `MANUAL_MAPPING_AUDIT.MD` (sekcja
  migracji), `testdata/baseline/README.MD`, `F9_COMPARISON.MD`, komentarze w
  `vuln_methods.yaml`, nagłówek F9 w planie CVE, ten rejestr.
- **Ograniczenia:** porównanie zależy od zawartości `.cache` (przy
  odświeżonym cache część „dane” trzeba wyjaśnić na nowo); proxy-test nie
  wykryje klienta z własnym transportem ignorującym zmienne proxy; 36
  wpisów XStream pozostaje w formacie legacy.
- **Następny krok (wymaga zgody):** CVE:F8a (docelowy runtime JDK —
  potrzebny wybór pierwszego dostawcy), po nim F9 dla trzeciego przyrostu;
  albo obszar spoza CVE (GRAPH_PRECISION/RTA).

### DESERIALIZATION:Z0+Z1 — katalog modeli i callbacki Java serialization — DONE (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** Z0 (kontrakt: katalog
  `DeserializerModel`, rodzaje dowodów, rola XStream bez zmian) i Z1
  (wejścia `ObjectInputStream`/`ObjectInput`, callbacki o dokładnych
  sygnaturach, konstrukcja według reguł serializacji, nieznany zbiór typów
  widoczny), aplikacja kontrolna z prawdziwym JVM. Gałąź
  `deserialization-z0-z1`. Ocena alternatyw: FRAMEWORKS:F3 (zmierzony
  zerowy efekt dispatchu proxy) i ASYNC_CALLBACKS:A3 (wymaga reszty
  DATA_FLOW:D3) odłożone.
- **Zrealizowane:**
  1. `internal/deserialize/model.go` (Z0): `Model` (ID, komponent, zakres
     wersji, wejścia jako `models.Signature`, reguła typów, `CallbackSpec`
     z dostępem i zakresem, pokrycie `modeled`/`entry-only`, zabezpieczenia,
     źródło), `EvidenceKind` (api-use, type-candidate, callback, filter,
     chain-candidate — Z1 tworzy trzy pierwsze, filtr wykrywa). `Catalog`:
     `jdk:java-serialization` (3 wejścia, 4 rodzaje callbacków) i
     `xstream:fromxml` (5 przeciążeń, entry-only).
  2. `sites.go`: `FindSites` — dokładne wejścia + `readObject`/
     `readUnshared` na podklasie `ObjectInputStream`; filtr
     `present_not_evaluated`, gdy w tej samej metodzie jest
     `setObjectInputFilter`, gdy odbiorca nadpisuje `resolveClass` albo gdy
     metoda tworzy taki strumień (`new`).
  3. `callbacks.go`: `Candidates` — klasy konkretne serializowalne (przez
     interfejsy/nadklasy wejścia, tabela klas JDK, JDK `*Exception`/`*Error`;
     reszta „unknown”, licznik z klasą wyjścia), bez enumów i rekordów;
     cele: prywatny niestatyczny `readObject`/`readObjectNoData` każdej klasy
     łańcucha, `readResolve` z regułami dostępu serializacji, Externalizable
     — publiczny konstruktor bezargumentowy i `readExternal`, konstruktor
     bezargumentowy pierwszej nieserializowalnej nadklasy. `Edges` — tylko
     dla modeli `modeled`.
  4. `internal/callgraph`: `BasisDeserialization`,
     `ApplyDeserializationEdges`. `cmd/ariadne`: krawędzie, log, każde
     wejście jako luka modelu `deserialization` (nieograniczona — callbacki
     JDK niemodelowane), raport „Deserialization entry points”, poprawione
     ograniczenie w szablonie.
  5. `scripts/prepare-deserialization-demo.sh` — aplikacja kontrolna
     (`Gadget` z prywatnym `readObject`, `WrongSig` z publicznym, `Ext`
     Externalizable, `Resolving` z `readResolve`, logback 1.2.3
     `LoggingEventVO`), uruchamiana na prawdziwym JVM (zapis, odczyt,
     asercje, które callbacki zadziałały).
- **Weryfikacja:** `internal/deserialize/deserialize_test.go` (callbacki:
  dokładne sygnatury i dostęp, łańcuch, konstruktor, dostęp `readResolve`
  między pakietami, Externalizable, enum/rekord/unknown; miejsca na OIS,
  ObjectInput, podklasie, filtry, XStream entry-only, odrzucone metody o tej
  samej nazwie; katalog), `internal/callgraph/deserialization_edges_test.go`.
  JVM (OpenJDK 25): odczyt uruchamia `Gadget.readObject`, `Ext.<init>`,
  `Ext.readExternal`, `Resolving.readResolve`, nie uruchamia
  `WrongSig.readObject` — dokładnie zgodnie z celami ariadne. ariadne na
  demo: CVE-2023-6378 NOT_REACHABLE → REACHABLE (`Main#read →
  LoggingEventVO#readObject`); CVE-2021-42550, -12798, -12801 → REACHABLE
  przez `Logger#readResolve` (w logback woła `LoggerFactory.getLogger`, czyli
  auto-konfigurację — realne zachowanie, krawędź „może wykonać”, strumień
  demo nie zawiera Loggera); strumienie `Hardened*InputStream` logback
  pokazane jako filtrowane. `go test ./...` (sieć zablokowana proxy) i
  `go vet ./...` — PASS.
- **Prawdziwe dane** (offline, `.cache`, `-profile spring`, main e8ee758 vs
  gałąź): petclinic 130 wejść (11 z filtrem), 2957 klas kandydatów, 227
  celów, 15436 krawędzi; WebGoat 142 (13), 3466, 296, 24272. Zmiany:
  CVE-2021-46877 (jackson-databind `NodeSerialization#readExternal`) i
  CVE-2023-6378 (logback `LoggingEventVO#readObject`) NOT_REACHABLE →
  FRAMEWORK na petclinic (wejście Spring `DefaultDeserializer` przez
  konwertery) i → REACHABLE na WebGoat (wejście JRuby
  `JavaProxy#marshal_load`); CVE-2022-45868 → FRAMEWORK na petclinic (H2
  `JdbcDataSource#readObject`, dalej długa ścieżka CHA). Nic więcej;
  reguła luk nadal obejmuje wszystkie NOT_REACHABLE (16/16, 10/10); v2
  bajtowo powtarzalny.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz deserializacji: Not
  implemented → Partial, z zakresem i dowodami), README (krawędzie
  deserializacji, baza `deserialization`, reguła luk, ograniczenia), FAQ
  (dlaczego nie ogranicza rzutowanie), `docs/RESULTS_CONTRACT.MD` §4,
  szablon raportu, plan DESERIALIZATION, ten rejestr.
- **Ograniczenia:** nie analiza łańcuchów gadżetów (Z3); callbacki JDK i
  to, co wołają na odczytanych obiektach, niemodelowane; kontrola danych
  wejściowych przez atakującego niemodelowana; filtry i `jdk.serialFilter`
  nieoceniane (Z2); `validateObject`, kanoniczne konstruktory rekordów,
  `readObjectOverride` poza zakresem; tabela JDK niepełna (910 / 553 klas
  „unknown”); XStream bez krawędzi callbacków.
- **Następny krok (wymaga zgody):** DESERIALIZATION:Z2 (ocena filtrów
  `ObjectInputFilter`/`resolveClass` dla znanych wzorców, pierwszy model
  biblioteki — XStream w wybranych wydaniach) albo inny obszar.

### DESERIALIZATION:Z2 — ocena filtrów i model XStream — DONE dla literalnych polityk w metodzie wejścia (2026-11-06)

- **Zakres zatwierdzony (2026-11-06):** model XStream zależny od wersji
  (callbacki, framework bezpieczeństwa, progi 1.4.7 / 1.4.18), ocena
  filtrów Java serialization (literalny `createFilter` na tym samym
  strumieniu, wzorce JEP 290), stany polityki w raporcie, aplikacje
  kontrolne na prawdziwym JVM. Gałąź `deserialization-z2`.
- **Zrealizowane:**
  1. `internal/deserialize/callbacks.go`: `ClassTargets(h, format)` —
     cele per klasa (potrzebne do odcinania klas); `FormatJava` jak w Z1,
     `FormatXStream`: każda klasa konkretna (też nieserializowalna) —
     `readResolve` z klasy lub nadklasy o dowolnym dostępie, prywatny
     `readObject` klas serializowalnych łańcucha, Externalizable:
     publiczny konstruktor i `readExternal`; bez innych konstruktorów.
     `Connect` — krawędzie per miejsce z pominięciem klas odrzuconych,
     liczniki `Candidates`/`Rejected`, próbka odrzuconych.
  2. `policy.go`: `ResolvePolicies` — fakty `internal/dataflow` metody
     wejścia (ponowny odczyt klasy). Java: `setObjectInputFilter` na tej
     samej alokacji strumienia z argumentem
     `ObjectInputFilter$Config#createFilter(String)` o stałym literale →
     `evaluated` (JEP 290: dokładna nazwa, `.*`, `.**`, prefiks `*`, `*`,
     negacja; limity pomijane; pierwsze dopasowanie; brak = UNDECIDED =
     dopuszczone; wzorce z modułem → unknown); inny filtr, inny/niepewny
     strumień, `resolveClass` → `unknown`. XStream: wersja z GAV klasy
     `com.thoughtworks.xstream.XStream`; < 1.4.7 `unrestricted`;
     1.4.7–1.4.17 start `ANY`; ≥ 1.4.18 start `NONE` (domyślna allowlista
     typów JDK); na tym samym `new XStream()` przed odczytem:
     `addPermission(NONE|ANY)` (czyści listę), `setupDefaultSecurity`,
     `allowTypes`/`denyTypes` (`Class[]`/`String[]` literalne),
     `*ByWildcard`, `*TypeHierarchy`; ocena od ostatnio dodanego, brak
     dopasowania = odrzucone; inne wywołania konfiguracji, instancja z
     pola/parametru, przekazanie instancji do innego kodu → `unknown`.
  3. Katalog: XStream modelowany (7 przeciążeń `fromXML`, `Format`),
     `Model.Format`; CLI: polityki, krawędzie obu formatów, log stanów;
     raport: kolumny polityki, wersji i odrzuconych klas.
  4. Fixture `testdata/deserialize` (stub API XStream + `fx.Cases` z
     wejściem na politykę, `scripts/build-deserialize-fixture.sh`, JAR-y w
     repo); `scripts/prepare-xstream-demo.sh` (XStream 1.4.5/1.4.10/1.4.20,
     z `--add-opens` dla JDK 25, asercje); demo deserializacji z odczytem
     filtrowanym.
- **Weryfikacja:** `internal/deserialize/policy_test.go` — na bytecode:
  Java brak/literał/dynamiczny/inny strumień; XStream w 1.4.5, 1.4.10,
  1.4.20 i bez wersji: domyślnie, `allowTypes`, wildcard, hierarchia,
  `setupDefaultSecurity`, `ANY` po `NONE`, `denyTypes`, pole, ucieczka;
  wzorce JEP 290 i wildcardy. (Jedno oczekiwanie testu poprawione: `fx.**`
  obejmuje podpakiety — model miał rację.) JVM (OpenJDK 25): 1.4.5 —
  wszystkie callbacki; 1.4.10 — samo `allowTypes(Gadget)` niczego nie
  ogranicza, `NONE`+`allowTypes(Gadget)` odrzuca `Plain`; 1.4.20 — domyślnie
  odrzuca `Gadget`, `allowTypes(Gadget)` dopuszcza tylko jego; filtr
  `demo.Gadget;!*` — `Gadget.readObject` działa, następna klasa
  odrzucona. ariadne na demach daje dokładnie te stany i cele.
  `go test ./...` (sieć zablokowana) i `go vet ./...` — PASS.
- **Prawdziwe dane** (offline, `.cache`, main e472f61 vs gałąź):
  werdykty bez zmian na petclinic i WebGoat. WebGoat: 143 wejścia (doszło
  przeciążenie `fromXML(String,Object)`), 4 wejścia XStream 1.4.5
  `unrestricted` (w tym lekcja vulnerable-components), +1088 krawędzi;
  advisory XStream były już REACHABLE przez punkt wejścia. Żadnego
  literalnego filtra Java w obu aplikacjach; 11/13 wejść z ograniczeniem
  to nadpisane `resolveClass` → `unknown`. v2 bajtowo powtarzalny.
- **Dokumentacja:** `docs/SUPPORT.MD`, README, FAQ, szablon raportu,
  `MANUAL_MAPPING_AUDIT.MD`, plan DESERIALIZATION, ten rejestr.
- **Ograniczenia:** tylko polityki literalne w metodzie wejścia; filtry
  procesowe, `resolveClass`, polityki budowane gdzie indziej,
  `allowTypesByRegExp`, własne `TypePermission` i konwertery — unknown;
  konwertery XStream dla typów JDK (klasyczne łańcuchy gadżetów)
  niemodelowane; inne biblioteki (Jackson, SnakeYAML, Kryo) bez modelu.
- **Następny krok (wymaga zgody):** DESERIALIZATION:Z3 (dane zewnętrzne i
  kandydaci łańcuchów), międzyproceduralne polityki (instancja XStream /
  filtr z pola lub parametru — wymaga DATA_FLOW:D3), albo inny obszar.
  (Zrealizowano Z3a — wpis niżej.)

### DESERIALIZATION:Z3a — kandydaci łańcuchów gadżetów (bez dopływu danych) — DONE (2026-11-07)

- **Zakres zatwierdzony (2026-11-07):** wersjonowany katalog kandydatów
  łańcuchów, stany `candidate_chain`/`conditions_unresolved`/
  `candidate_blocked` (+ `classes_absent`, `no_entry`), tożsamość klas przez
  inwentarz komponentów (F6/F7), strażniki z bytecode, budżet = jawnie
  skończony katalog; dopływ danych świadomie poza zakresem (wymaga
  DATA_FLOW:D4) i raportowany jako `not_analyzed`. Gałąź
  `deserialization-z3a` (odgałęziona od `deserialization-z2`, PR #5 jeszcze
  niescalony).
- **Zrealizowane:**
  1. `internal/deserialize/chains.go`: `ChainCatalog` (`z3a-1`) —
     CommonsCollections6 (commons-collections 3.x), CommonsCollections2/4
     (commons-collections4 4.0), CommonsBeanutils1, Groovy1, XStream
     `java.beans.EventHandler` (CVE-2013-7285); per łańcuch: komponenty,
     klasy wymagane, klasy JDK (wyzwalacz pierwszy), wyzwalacz, ujście,
     wydania, strażniki (`declares-method` z opcjonalnym wywołaniem,
     `not-serializable`, `declares-field`), limit JDK, warunki
     niesprawdzane, advisory, źródło. `EvaluateChains`: klasy po nazwie lub
     pod relokacją potwierdzonej instancji (F7a), tożsamość (inny/nieznany
     właściciel → unresolved), strażnik czytany z analizowanej klasy (także
     relokowanej), `-target-java` (Groovy1 ≤ 8 z niepewnym 8, CB1 ≤ 15),
     potem per wejście formatu: polityka Z2 (oceniona odrzucająca którąkolwiek
     klasę łańcucha, także JDK → blocked tylko tam; `unknown`/
     `version_unknown` → unresolved).
  2. `cmd/ariadne/deserchains.go` (adapter inwentarza), wywołanie po F6a/F7a,
     linia logu; raport „Candidate gadget chains” (wejścia aplikacji i z
     własnym powodem w tabeli, biblioteczne zliczone), ograniczenia w
     szablonie. Bez krawędzi, bez wpływu na werdykty, bez JSON v2.
  3. Poprawka tożsamości (osobny commit, wykryta tym demem):
     `input.resolveOne` — gdy żaden `pom.properties` w jarze nie pasuje do
     nazwy pliku, a `-deps-list` ma dokładnie ten plik, wygrywa lista.
     groovy 2.4.3 (jedyny `pom.properties`: przepakowany commons-cli 1.2) był
     odpytywany w OSV jako commons-cli 1.2 — jego advisory (m.in.
     CVE-2015-3253, CVE-2016-6814) w ogóle nie były sprawdzane.
  4. `scripts/prepare-deserialization-chains-demo.sh` — dwa buildy
     (vulnerable: cc 3.2.1, cc4 4.0, groovy 2.4.3, beanutils 1.9.4,
     XStream 1.4.6; fixed: 3.2.2, 4.1, 2.4.4, 1.9.4, 1.4.7); na prawdziwym
     JVM tylko round-trip pojedynczej klasy strażnika (żaden łańcuch nie jest
     budowany), z asercjami.
- **Weryfikacja:** `internal/deserialize/chains_test.go` (kandydat; strażnik
  3.2.2 i readObject bez wywołania strażnika; 4.0 vs 4.1; nieznany
  właściciel; kopia relokowana potwierdzona, jej strażnik, relokacja innego
  komponentu; JDK 0/7/8/11 i groovy 2.4.4; XStream ze strażnikiem / bez /
  bez klasy; filtr per wejście, allowlista; brak wejścia; poprawność
  katalogu), `internal/input` (lista wygrywa z obcym `pom.properties`, bez
  listy bez zmian). Strażniki potwierdzone najpierw na bytecode prawdziwych
  JAR-ów (3.2.1/3.2.2, 4.0/4.1, groovy 2.4.3/2.4.4, XStream 1.4.6/1.4.7,
  beanutils 1.9.4). JVM (OpenJDK 25): vulnerable — wszystkie round-tripy
  przechodzą; fixed — `UnsupportedOperationException` (cc3),
  `NotSerializableException` (cc4), `UnsupportedOperationException`
  (groovy), `ConversionException` (XStream), beanutils przechodzi (brak
  poprawki). ariadne: vulnerable — CC6, CC2/4 `candidate_chain` na
  `Main#read`, `candidate_blocked` na `Main#readFiltered`
  (`createFilter("demo.*;!*")` odrzuca `HashMap`/`PriorityQueue`), XStream
  `candidate_chain` na `readXml`, Groovy1 i CB1 `conditions_unresolved` bez
  `-target-java`, `candidate_blocked` z `-target-java 17`; fixed — wszystko
  `candidate_blocked` oprócz CB1. Przed poprawką tożsamości Groovy1 w
  vulnerable był `conditions_unresolved` z powodem „found in
  commons-cli:commons-cli, not identified” — mechanizm tożsamości zadziałał
  poprawnie i ujawnił błąd F1a. `go test ./...`, `go vet ./...` — PASS.
- **Prawdziwe dane** (offline, `.cache`, `-profile spring`, HEAD 87baaa6 vs
  gałąź): werdykty i JSON v2 bajtowo identyczne na petclinic i WebGoat (także
  po poprawce tożsamości — żaden ich jar nie miał tego wzorca); dwa przebiegi
  gałęzi bajtowo identyczne. petclinic: 4 łańcuchy Java `classes_absent`,
  XStream `no_entry`. WebGoat: 4 łańcuchy Java `classes_absent`, XStream
  EventHandler `candidate_chain` na 4 wejściach XStream 1.4.5, w tym
  `VulnerableComponentsLesson#completed` — exploit samej lekcji.
- **Dokumentacja:** `docs/SUPPORT.MD` (deserializacja, tożsamość Maven),
  README (tożsamość jarów, akapit Z3a, ograniczenia), FAQ („candidate” vs
  „exploitable”), `docs/RESULTS_CONTRACT.MD` §4, szablon raportu, komentarze
  `model.go`, plan DESERIALIZATION, ten rejestr.
- **Ograniczenia:** katalog pięciu łańcuchów, nie wyszukiwanie; brak
  dopływu danych (D4); `jdk.serialFilter` niewidoczny; kopia relokowana
  znajdowana tylko przy relokacjach potwierdzonej instancji; warunki typu
  „komparator w strumieniu” tylko opisane; stany łańcuchów nie trafiają do
  JSON v2 ani nie łączą się z wierszem advisory (Z4); łańcuchy XStream inne
  niż EventHandler (konwertery typów JDK, CVE-2021-*) poza katalogiem.
- **Następny krok (wymaga zgody):** DESERIALIZATION:Z4 (raportowanie: stany
  łańcuchów w JSON v2 i snapshot, powiązanie z wierszem advisory), DATA_FLOW:D3
  → D4 (dopływ danych — reszta Z3), albo inny obszar.
  (Zrealizowano Z4 — wpis niżej.)

### DESERIALIZATION:Z3 — dopływ danych do kandydatów łańcuchów — DONE (2026-11-07)

- **Zakres zatwierdzony (2026-11-07):** ostatni punkt Z3 („powiązać źródła
  danych z argumentem wejścia; osobno pokazać brak analizy flow i znalezioną
  ścieżkę”): wynik DATA_FLOW:D4 wejścia przenoszony na każdy wynik łańcuch ×
  wejście i na łańcuch; stan łańcucha bez zmian; v2, raport, wiersz CVE;
  wersja katalogu `z3a-1` → `z3-1`; aplikacja kontrolna łańcuchów
  rozszerzona o wejścia zasilane z żądania. Gałąź `deserialization-z3-flow`
  od `archive-data-flow-d4` (PR #10).
- **Zrealizowane:**
  1. `internal/deserialize/chains.go`: `ChainSite.Flow` = stan taint wejścia
     (`siteFlow`; `not_analyzed` bez `-taint`); `ChainPresence.Flow`
     (`chainFlow`: najmocniejszy stan — `flow_found` > `undetermined` >
     `no_flow_in_model` > `not_assessed` — wśród wejść, na których łańcuch
     nie jest zablokowany; gdy zablokowany wszędzie — wśród wszystkich) i
     `FlowCounts`; `ShownSites` pokazuje też wejście biblioteczne z
     `flow_found`; `ChainFlowCounts` do logu; warunki „attacker-controlled”
     w katalogu przeredagowane: sprawdzane tylko dla danych żądania.
     `ChainCatalogVersion = "z3-1"`.
  2. v2: `chains[].flow` (już nie stała), `chains[].flows_by_state`,
     `chains[].sites[].flow`, `findings[].deserialization_chains[].flow`.
  3. Raport: kolumna Flow w tabeli wejść łańcucha, „flow: …” w nagłówku
     łańcucha, opis znaczenia; wiersz CVE „gadget chain: … , request data
     flow_found” (tylko z `-taint`); log „flow per chain”.
  4. `scripts/prepare-deserialization-chains-demo.sh`: stuby `@RequestParam`/
     `@PostMapping` o nazwach Springa i `demo.Web` (`upload` → Base64 →
     `readObject`, `uploadXml` → `fromXML`), wywoływane z `main`.
- **Weryfikacja:** `TestChainFlow` (bez `-taint` wszystko `not_analyzed`;
  kandydat z wejściem `flow_found` i wejściem ze stałą; `flow_found` tylko na
  zablokowanym wejściu → łańcuch bierze stan z niezablokowanego; zablokowany
  wszędzie → `flow_found` zachowany, stan `candidate_blocked`; `undetermined`
  > `no_flow_in_model`; biblioteczne `flow_found` pokazane w raporcie),
  `TestChainFlowInV2`. Demo łańcuchów na OpenJDK 25 (obie wersje działają,
  sondy strażników jak dotąd), ariadne `-taint -offline`: podatna — CC6,
  CC2/4, XStream `candidate_chain` + `flow_found` przy `Web#upload`/
  `Web#uploadXml`, CB1/Groovy1 `conditions_unresolved` + `flow_found`,
  `Main#readXml` `no_flow_in_model`, `Main#read`/`readFiltered`
  `undetermined` (bajty z `ByteArrayOutputStream#toByteArray`, poza
  katalogiem D3 — luka uczciwie zgłoszona, nie „brak przepływu”); poprawiona
  — wszystkie łańcuchy zablokowane poza CB1, przepływy bez zmian. WebGoat:
  `xstream:eventhandler` `candidate_chain` + `flow_found` (wewnętrzne
  `XStream#fromXML` i `VulnerableComponentsLesson#completed`), wiersz
  CVE-2013-7285 „xstream:eventhandler candidate_chain, request data
  flow_found”. Werdykty WebGoat 42/51/10/61/0 i petclinic 3/71/16/57/6 bez
  zmian; v2 względem binarki `main`: tylko `chain_catalog`, nowe pola
  `flow`/`flows_by_state` (oraz `analysis_id`/wersja); z `-taint` względem
  bez: tylko pola flow/taint; dwa przebiegi z `-taint` bajtowo identyczne.
  `go test ./...`, `go vet ./...` — PASS.
- **Dokumentacja:** `docs/SUPPORT.MD` (wiersz Deserialization), README
  (sekcje łańcuchów i v2), `FAQ.md`, `docs/RESULTS_CONTRACT.MD`, szablon
  raportu, plan DESERIALIZATION, ten rejestr. `FAQ_PL.MD` nietknięty.
- **Ograniczenia:** tylko dane żądania (źródła D4); `flow_found` przy
  kandydacie nie sprawdza grafu obiektów, wartości pól ani konfiguracji
  środowiska; `no_flow_in_model` nie oznacza zaufanych bajtów; ograniczenia
  D4 (pola, lambdy, wywołujący spoza aplikacji) przechodzą na łańcuchy.
- **Wydanie i archiwum:** zmergowane jako PR #11 (`e818150`, 2026-11-07);
  cały obszar (Z0–Z4) zarchiwizowany w
  [archive/2026-11-07/deserialization/](../archive/2026-11-07/deserialization/COMPLETION.md).
- **Następny krok:** obszar zamknięty. Potem DATA_FLOW:D5 albo inny obszar (CVE:F8,
  ROADMAP:B3, GRAPH_PRECISION:P0) — wymaga zgody.

### DESERIALIZATION:Z4 — kwalifikacja i raportowanie — DONE (2026-11-07)

- **Zakres zatwierdzony (2026-11-07):** ocena łańcuchów per kopia klas
  (dwie instancje), JSON v2 (pola addytywne) z powiązaniem przy ustaleniu,
  wersje reguł w snapshotcie, notka przy wierszu advisory, fixtures z
  zapisanym oczekiwanym modelem i wyrocznią JVM, źródłowy zestaw CVE,
  kontrole B4. Gałąź `deserialization-z4` od `main` (3243863).
- **Zrealizowane:**
  1. `chains.go`: `EvaluateChains` ocenia każdą kopię — nazwy oryginalne i
     każdą potwierdzoną instancję komponentu łańcucha z relokacjami
     (`ChainPresence.Copy`, `.Instance`); strażnik czytany z kopii (także
     wywołanie strażnika relokowane tymi samymi regułami); oryginał
     raportowany jako `classes_absent` tylko gdy nie ma żadnej kopii;
     łańcuch z samych klas JDK dostaje instancję właściciela klasy strażnika.
     `model.go`: `ModelCatalogVersion = "z2-1"`.
  2. `internal/resultsv2/deserialization.go`: `Document.Deserialization`
     (wersje katalogów, wejścia z polityką, kopie łańcuchów z klasami,
     powodami, stanami per wejście, `flow`, podsumowanie),
     `Finding.DeserializationChains` (łańcuch nazywa advisory po ID/aliasie
     i należy do tej samej instancji). Bez zmian werdyktów i v1.
  3. Snapshot: `algorithms` `deserialize.models`, `deserialize.chains`.
     Raport: „gadget chain: …” w kolumnie Reasons wiersza advisory (klucz
     advisory + pakiet), kopia w nagłówku łańcucha, wersja katalogu modeli.
  4. `testdata/deserialize/oracle` + `scripts/deserialize-oracle.sh`:
     scenariusze z planu (prywatny callback, publiczny `readObject`,
     Externalizable, rekord, nieznana serializowalność, dwie instancje
     `lib.Cb`/`shaded.lib.Cb`, filtr), JAR-y i `observed.tsv` z OpenJDK 25,
     `model-gaps.tsv`/`model-extras.tsv`; `TestDeserializationOracle`.
  5. `testdata/deserialize/cves.yaml` (7 przypadków: chain/entry/callback,
     oczekiwania ze źródeł), `TestDeserializationControlSet` (słownik,
     łańcuchy w katalogu, każde advisory łańcucha ma przypadek),
     `testdata/deserialize/QUALIFICATION.MD` (porównanie).
- **Weryfikacja:** testy `chains_test.go` (dwie instancje w obu
  kierunkach, tylko kopia relokowana), `resultsv2` (`chainRefs`,
  `buildDeserialization`), wyrocznia (0 nadmiarowych krawędzi, 2
  zadeklarowane luki: konstruktor kanoniczny rekordu, klasa o nieznanej
  serializowalności; filtr i dwie instancje zgodne), zestaw CVE — każdy
  sprawdzalny przypadek zgodny (CVE-2015-4852 nie do sprawdzenia: komponent
  WebLogic). B4 (offline, ten sam `.cache`, main 3243863 vs gałąź):
  petclinic 3/71/16/57/6 i WebGoat 42/51/10/61/0 bez zmian; dema
  (vulnerable, fixed, vulnerable `-target-java 17`, deserialization-demo) bez
  zmian werdyktów; v2 identyczny poza polami Z4 (i wersją narzędzia),
  +3 % rozmiaru; dwa przebiegi bajtowo identyczne. WebGoat: CVE-2013-7285 →
  `xstream:eventhandler candidate_chain`. `go test ./...`, `go vet ./...` —
  PASS. `go mod tidy`: `gopkg.in/yaml.v3` teraz bezpośrednia zależność
  (import w teście).
- **Dokumentacja:** `docs/SUPPORT.MD`, `docs/RESULTS_CONTRACT.MD` (schemat
  v2), README, FAQ, szablon raportu, plan DESERIALIZATION, ten rejestr,
  `testdata/deserialize/oracle/README.md`, `QUALIFICATION.MD`.
- **Ograniczenia:** bez dopływu danych (reszta Z3, DATA_FLOW:D4); wyrocznia
  pokrywa tylko swoje scenariusze (brak obserwacji ≠ brak callbacku);
  wyrocznia dla XStream nie istnieje poza demem Z2; rekordy,
  `validateObject`, `readObjectOverride` dalej poza modelem.
- **Następny krok (wymaga zgody):** DATA_FLOW:D3 (pola instancyjne,
  wartości zwracane) → D4 (taint; reszta DESERIALIZATION:Z3), albo inny
  obszar (np. GRAPH_PRECISION P0, FRAMEWORKS F3).

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

### ROADMAP:B3 — koordynator iteracji, pierwszy przyrost — IN_PROGRESS (2026-10-25)

- **Zakres:** po zatwierdzeniu planu zbudowano `internal/analysis`, które
  łączy rejestr `internal/models`, hierarchię i graf wywołań. Zależności
  pozostają DAG-iem: `analysis` może importować `models` i `callgraph`,
  a żaden z tych pakietów nie importuje koordynatora.
- **Zachowanie:** dla osiągalnych metod wyszukiwane są wywołania
  dopasowane dokładnym `Signature` i GAV. Wstrzyknięty evaluator zwraca
  fakty z dokładnymi metodami docelowymi; koordynator dodaje krawędzie
  `BasisModel` z PC miejsca wywołania, propaguje osiągalność i wraca do
  nowych miejsc aż do punktu stałego. Powtarzane dowody PC są zachowane.
  `MaxIterations` ogranicza przebieg; jeśli po ostatnim przebiegu graf
  nadal się rozszerzył, wynik ma `Incomplete` i konkretną przyczynę.
  Brak celu w analizowanej hierarchii także jest jawnie raportowany.
- **Integracja:** CLI rejestruje evaluator ServiceLoader, wylicza korzenie
  zgodne z `-roots all/profile`, a następnie uruchamia koordynator przed
  reachability. Wycofano bezpośrednie wstawianie krawędzi S2; teraz
  `BasisModel` dodaje wyłącznie cele z osiągalnych call sites.
  Konstruktor i istniejący `<clinit>` zachowują odrębny rodzaj krawędzi.
  W raporcie liczba krawędzi modelu jest oddzielona od direct/CHA/init.
- **Granice tego przyrostu:** evaluator krawędzi providerów dotyczy czterech
  dokładnych sygnatur `ServiceLoader.load`/`loadInstalled`; trzy dokładne
  sygnatury konsumpcji (`iterator()`, `stream()`, `Provider.get()`) są
  dopasowywane wyłącznie po to, by zgłosić `Incomplete` bez zgadywania celu.
  Powiązanie iteratora/uchwytu z load site pozostaje niezaimplementowane.
  Cel krawędzi musi
  wskazywać metodę w analizowanej hierarchii — nie generujemy klas ani
  metod. Nie ma wyboru modeli z profilu ani krawędzi iteratora/
  `Provider.get` (wywołania są tylko raportowane jako incomplete),
  invalidation zależnych wyników po zmianie modelu, ani pomiaru RTA/
  points-to. Nierozwiązane argumenty, brak rejestracji providera oraz
  niezweryfikowana widoczność przez wskazany loader/JPMS są jawnie
  raportowane jako incomplete; przy loaderach innych niż classpath
  znane provider-y nadal są krawędziami możliwości z luką widoczności.
- **Zmiany:** `internal/analysis`, evaluator w `internal/serviceload`,
  rozszerzone fakty `internal/models`, `BasisModel` i licznik w
  `internal/callgraph`, integracja w `cmd/ariadne/main.go`, testy,
  `README.md`, szablon raportu, `docs/SUPPORT.MD` i ten rejestr.
- **Weryfikacja wykonana:**
  `GOCACHE=/tmp/ariadne-gocache go test ./internal/analysis ./internal/models ./internal/callgraph`
  — PASS; `GOCACHE=/tmp/ariadne-gocache go vet ./...` — PASS;
  `GOCACHE=/tmp/ariadne-gocache go test ./...` — PASS po zgodzie na
  wykonanie poza sandboxem, bo kilka istniejących testów tworzy lokalne
  serwery HTTP. Testy obejmują rozszerzenie grafu ujawniające następne
  osiągalne miejsce modelu, stały punkt, limit iteracji, nierozpoznany
  cel oraz brak ewaluacji nieosiągalnego/niedopasowanego wersją API.
  Integracyjnie `go run ./cmd/ariadne ... -offline` wykonano na
  petclinic i WebGoat: odpowiednio 3 i 80 krawędzi modelu, 2 iteracje;
  oba świeże `results.json` były bajtowo identyczne z zapisanymi
  baseline'ami (148 i 160 CVE). W obu przebiegach niekompletne
  ServiceLoader przypadki były widoczne w ostrzeżeniach. Dodatkowy
  przebieg petclinic z `-roots profile` zakończył się poprawnie
  (24 korzenie, 3 krawędzie modelu).
- **Następny krok:** dodać inne źródłowo uzasadnione modele i dopiero
  wtedy ustalić, czy potrzebna jest invalidation zależnych wyników;
  B3 pozostaje IN_PROGRESS.

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
   **Aktualizacja 2026-11-05:** `internal/analysis` (koordynator B3) już
   istnieje — zob. wpis "ROADMAP:B3 — koordynator iteracji" (IN_PROGRESS).
   SERVICE_LOADING:S2 jest DONE i zarchiwizowane
   (`archive/2026-11-05/service_loading/SERVICE_LOADING_S2.md`).
   SERVICE_LOADING:S3 ma dwa wycinki: inwentarz deskryptorów JPMS oraz
   `-jpms-scenario module-path` z aplikacją kontrolną
   `scripts/prepare-jpms-demo.sh` (CVE-2022-42889 NOT_REACHABLE vs
   REACHABLE), a wycinek 3 dodał `mixed`/`-jpms-module-path` i
   `ModuleLayer.boot()`; S3 zamknięte i zarchiwizowane
   (`archive/2026-11-05/service_loading/SERVICE_LOADING_S3.md`).
   Rozpoznawanie jawnych `ClassLoader` (korekta po S2) zrobione: 1/19
   i 5/32 miejsc. DATA_FLOW:D3 wycinek "parametry od wywołujących"
   zrobiony (load sites z usługą 13→20 / 31→38, statusy bez zmian).
   REFLECTION:R3 wycinek 1 zmierzony 2026-11-05: 0 możliwych krawędzi
   na petclinic/WebGoat, reguła stopu — nie budowany (zob. wpis).
   DATA_FLOW:D3 pola `final` z parametru konstruktora zrobione
   (load sites z usługą 20→25 / 38→43, krawędzie bez zmian).
   Model `AccessController.doPrivileged` zrobiony (krawędzie callbacków
   62→288 / 44→284, statusy CVE bez zmian). **SERVICE_LOADING S0–S4
   zamknięte i zarchiwizowane 2026-11-05**
   (`archive/2026-11-05/service_loading/COMPLETION.md`). Najbliższe
   kandydatury: dalsze modele B3, FRAMEWORKS:F3, ASYNC_CALLBACKS:A3,
   CVE:F0–F1 (wpływ luk na werdykt), GRAPH_PRECISION (RTA; odblokowałoby
   REFLECTION:R3). **CVE:F0 zrobione 2026-11-05 (kontrakt)**, CVE:F1a (model
   komponentów) i CVE:F1b (kody powodów, ustalenia per instancja, reguła
   luk report-only) oraz CVE:F1c (`-json-format v2`) zrobione — CVE:F1
   zamknięte; CVE:F2 (świeżość danych) i CVE:F3 (snapshoty) zrobione
   2026-11-05, CVE:F4 (wersje i sygnatury) również — następny krok w tym
   obszarze: CVE:F5.
   **Aktualizacja 2026-11-06:** CVE:F5 (F5a–F5c), F6 (F6a, F6b) i F7
   (F7a, jawna relokacja) zrobione — zob. wpisy wyżej. Następne
   kandydatury w CVE: F8a (wymaga wyboru dostawcy JDK) albo F9
   (kwalifikacja drugiego przyrostu).
3. Zaproponuj najbliższą wykonalną fazę lub minimalny wspólny kontrakt, przedstaw zakres, kolejność i priorytety, a następnie zaczekaj na wyraźną zgodę użytkownika. Do tego czasu wykonuj tylko przegląd do odczytu.
4. Po zatwierdzeniu wykonaj rezultat fazy, jego weryfikację i aktualizację dokumentacji. Jeżeli jest to faza kontraktowa, nie oznaczaj przez to funkcji jako zaimplementowanej.

Wersjonowane dowody i stan repozytorium są podstawą wznowienia. Poprzedni opis czatu nie zastępuje ich weryfikacji.
