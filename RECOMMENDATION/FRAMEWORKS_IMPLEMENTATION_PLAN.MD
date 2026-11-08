# Frameworki i konfiguracja — fazy F0–F5

Status: plan. Wymagania: [FRAMEWORKS.MD](FRAMEWORKS.MD). Prefiks F w tym pliku dotyczy frameworków; zapis „CVE F1” odnosi się do osobnego planu CVE.

## F0. Model frameworka i scenariusza

Zależności: B0–B1, JVM G0–G2 i kontrakt rejestru B3.

- W `internal/profile` oddzielić wykrycie dostępnego frameworka od aktywnego modelu wybranego scenariusza.
- Zaprojektować katalog modeli z wersją frameworka, obsługiwanymi adnotacjami/API i ograniczeniami. Uzgodnić migrację istniejących profili YAML i syntetycznych Spring Data edges.
- Wprowadzić proponowane `FrameworkScenario`, `BeanDefinition`, `ConditionalBinding`, `ExternalEntry`. Warunek ma stan true/false/unknown.
- Wybrać pilot Spring oparty o istniejące aplikacje i małe fixtures; nie deklarować wszystkich frameworków jako obsługiwanych przez jeden model.

Gotowe: obecny profil zachowuje działanie, a dostępność biblioteki i aktywna konfiguracja są osobnymi informacjami.

## F1. Adnotacje, zasoby i jawna konfiguracja — DONE (zarchiwizowane 2026-10-23)

Zrealizowane dla swojego faktycznego zakresu i zweryfikowane na realnym
bytecode; zobacz
[archive/2026-10-23/frameworks/FRAMEWORKS_F1.md](../archive/2026-10-23/frameworks/FRAMEWORKS_F1.md)
po pełny zapis (co zrobione, dowody, co celowo odłożone) i
`docs/SUPPORT.MD` po bieżący opis zachowania. Konflikty nazw beanów są
wykrywane, nie rozstrzygane; pełny priorytet wielu źródeł konfiguracji i
jawne wejście plików/właściwości scenariusza (poza `-active-profiles`)
pozostają nieruszone — żadny z tych punktów nie trafił do zrealizowanego
zakresu, odnotowane szczerze w archiwum.

## F2. Pierwsze DI i model zewnętrznych wejść

Zależności: F1, DATA_FLOW D0–D2 dla prostych fabryk, B3.

- Modelować wybrane definicje komponentów i `@Bean`, konstrukcję, wstrzyknięcia i qualifier/primary według wybranej wersji Spring.
- Znane fabryki przekazują modelowane alokacje i typy. Niejednoznaczne lub programowalne fabryki zachowują alternatywy/luki.
- Wprowadzić oznaczone korzenie dla HTTP, servlet/filter/lifecycle i prostego scenariusza kontenera. Nie uznawać każdego publicznego helpera za endpoint.
- Katalog punktów wejścia może wskazać zewnętrzny parser/dispatch przed kodem aplikacji. Pokazywać podstawę i warunki aktywacji; nie przypisywać dowolnemu framework CVE pewnej ścieżki wyłącznie na podstawie obecności pakietu.
- Przekazać tożsamość parametrów wejściowych do przyszłych źródeł taint D4. Samo `roots=all` nie tworzy takich źródeł.

Gotowe: jeden bean/qualifier, konflikt, nieznany profil i kontroler z helperem mają właściwe wyniki. F0–F2 to pierwszy zakres frameworków.

Wycinek „dispatch kontenera servletów” (2026-11-08): model B3 `servlet-container:dispatch` (`internal/servletdispatch`, wersja `s1`) — ograniczony wyjątek od pominięcia odbiorców `javax.`/`jakarta.` w CHA dla `Servlet#service`, `Filter#doFilter` i wewnętrznego dispatchu szablonowego `HttpServlet` (`service(HttpServletRequest, …)`, `doGet`/`doPost`/…, tylko gdy wołającym jest sam `HttpServlet`). Cele: wyłącznie servlety i filtry zarejestrowane przez aplikację (web.xml, `@WebServlet`/`@WebFilter`, w aplikacji Spring Boot beany będące servletem/filtrem oraz katalog autokonfiguracji Boot sprawdzony w jej źródłach, z warunkami `@ConditionalOnClass` i `exclude`), metoda wybrana przez JVM na zarejestrowanej klasie; krawędzie z podstawą `model`, rejestracje z własną podstawą w logu; rejestracja nieznanej klasy (`FilterRegistrationBean` itp.) to niekompletny fakt modelu, nie zgadywanie. Wyrocznia JVM `testdata/servletdispatch` (`scripts/servletdispatch-oracle.sh`). Szczegóły i wpływ na petclinic/WebGoat: IMPLEMENTATION_STATUS, wpis FRAMEWORKS:F2 (dispatch kontenera servletów), oraz `testdata/graphbench/RESULTS.MD`.

## F3. Proxy i ograniczony model AOP

Zależności: F2, DATA_FLOW D3 dla bardziej złożonych powiązań, REFLECTION R2 i JVM G2 w używanych przypadkach.

- Wprowadzić model tożsamości proxy, targetu oraz interceptora; dla ogólnego JDK Proxy powiązać handler i argument `Method` z właściwym wywołaniem interfejsu.
- Zachować różnice JDK/CGLIB, final/private i self-invocation. Nie tworzyć całego advice chain dla zwykłego bezpośredniego wywołania targetu.
- Dodać ograniczoną semantykę wybranych pointcutów i advice. Nieznane wyrażenia pozostają oznaczone; `around` może warunkować przekazanie dalej.
- Migrować Spring Data bridge do katalogu modeli z deskryptorami, powiązaniem receivera i wersji; zachować oryginalną regułę jako jawny fallback przybliżony.
- Udostępnić katalog typów tworzonych przez DI/proxy dla GRAPH_PRECISION; generowany typ wirtualny musi być oznaczony, nie udawać znalezionej klasy.

Gotowe: proxy/target/self-invocation, dwa handlery, warunkowe advice i generowany typ nie mieszają się w jedną ścieżkę.

## F4. Rozszerzanie wejść i frameworków

- Dodać osobne modele schedulerów i event/message listeners, współdzieląc mechanikę wywołania callbacków z ASYNC.
- Kolejne Servlet/Jakarta/Micronaut/Quarkus modele wdrażać na podstawie ich rzeczywistych artefaktów i wersji; sam wygenerowany bytecode nie dowodzi, że każda ścieżka konfiguracji została odtworzona.
- Programowalne rejestracje, dynamiczne profile, SpEL i inne złożone wyrażenia obejmować dopiero w jawnym zakresie; nie wykonywać aplikacji podczas analizy statycznej.
- Opracować zmianę `reach/bfs.go`/klasyfikacji dla modelowanych wejść jako osobny tryb, z zachowaniem wyjaśnienia granic frameworka i eksportu legacy.

Gotowe: lista wspieranych wejść jest wersjonowana; nowe wejścia nie zmieniają niejawnie znaczenia wszystkich istniejących wyników.

## F5. Kwalifikacja

- Fixtures: adnotacje z wartościami i defaults, złożony marker, parametry, XML, priorytety configu, qualifier, dwa profile, proxy, self-invocation, kontener przed kontrolerem i konsument wiadomości.
- Porównać w kontrolowanym uruchomieniu rzeczywiste beany/endpointy z modelem na kilku scenariuszach; nie przenosić wniosków między wersjami bez sprawdzenia.
- Raportować liczbę reguł, wiązań, nieznanych warunków i nierozstrzygniętych wejść. Zapis scenariusza i modeli w snapshotcie jest obowiązkowy.
- Wykonać B4 i migrację dokumentacji. Brak wsparcia fragmentu konfiguracji ogranicza wniosek negatywny, nie zamienia go w „bean nie istnieje”.
