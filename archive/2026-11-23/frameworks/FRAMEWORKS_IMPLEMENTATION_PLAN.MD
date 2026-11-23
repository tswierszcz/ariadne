# Frameworki i konfiguracja — fazy F0–F5

Status: plan. Wymagania: [FRAMEWORKS.MD](FRAMEWORKS.MD). Prefiks F w tym pliku dotyczy frameworków; zapis „CVE F1” odnosi się do osobnego planu CVE.

## F0. Model frameworka i scenariusza — DONE (2026-10-23; formalnie zamknięte 2026-11-23)

**Zamknięcie 2026-11-23:** wykrycie frameworka oddzielone od aktywnego profilu (`-profile auto`, `Catalog`), kontrakt `FrameworkScenario`/`ConditionalBinding`/`ExternalEntry` z warunkiem true/false/unknown (`internal/profile/scenario.go`), pilot Spring na petclinic/WebGoat i fixtures; wersjonowanie reguł profili w snapshotcie (`f5-1`).

Zależności: B0–B1, JVM G0–G2 i kontrakt rejestru B3.

- W `internal/profile` oddzielić wykrycie dostępnego frameworka od aktywnego modelu wybranego scenariusza.
- Zaprojektować katalog modeli z wersją frameworka, obsługiwanymi adnotacjami/API i ograniczeniami. Uzgodnić migrację istniejących profili YAML i syntetycznych Spring Data edges.
- Wprowadzić proponowane `FrameworkScenario`, `BeanDefinition`, `ConditionalBinding`, `ExternalEntry`. Warunek ma stan true/false/unknown.
- Wybrać pilot Spring oparty o istniejące aplikacje i małe fixtures; nie deklarować wszystkich frameworków jako obsługiwanych przez jeden model.

Gotowe: obecny profil zachowuje działanie, a dostępność biblioteki i aktywna konfiguracja są osobnymi informacjami.

## F1. Adnotacje, zasoby i jawna konfiguracja — DONE (zarchiwizowane 2026-10-23)

Zrealizowane dla swojego faktycznego zakresu i zweryfikowane na realnym
bytecode; zobacz
[archive/2026-10-23/frameworks/FRAMEWORKS_F1.md](../../2026-10-23/frameworks/FRAMEWORKS_F1.md)
po pełny zapis (co zrobione, dowody, co celowo odłożone) i
`docs/SUPPORT.MD` po bieżący opis zachowania. Konflikty nazw beanów są
wykrywane, nie rozstrzygane; pełny priorytet wielu źródeł konfiguracji i
jawne wejście plików/właściwości scenariusza (poza `-active-profiles`)
pozostają nieruszone — żadny z tych punktów nie trafił do zrealizowanego
zakresu, odnotowane szczerze w archiwum.

## F2. Pierwsze DI i model zewnętrznych wejść — DONE for this scope (2026-11-23)

**Zamknięcie 2026-11-23 (decyzja użytkownika):** siedem wycinków (DI po typie/qualifier/primary, korzenie `web.xml`, metody `@Bean`, katalog wejść zewnętrznych, repozytoria Spring Data, fakty `@Value`, dispatch kontenera servletów) i późniejsze `s2`, `l1`/`l2`, `j1`; kryterium gotowości sprawdzone wyroczniami: bean/qualifier (`f5-3`, `TestDIInjectionOracle`), konflikt nazw (F1), nieznany profil → `unknown`, kontroler z helperem — publiczna metoda bez mapowania nie jest korzeniem (`f5-2`, `TestMVCOracle`); tożsamość parametrów wejściowych dla taint (`src:spring-mvc-param`, DATA_FLOW D4). Odłożone: programowalne fabryki (alternatywy/luki), kandydaci wstrzyknięć filtrowani przez `@Profile`.

Zależności: F1, DATA_FLOW D0–D2 dla prostych fabryk, B3.

- Modelować wybrane definicje komponentów i `@Bean`, konstrukcję, wstrzyknięcia i qualifier/primary według wybranej wersji Spring.
- Znane fabryki przekazują modelowane alokacje i typy. Niejednoznaczne lub programowalne fabryki zachowują alternatywy/luki.
- Wprowadzić oznaczone korzenie dla HTTP, servlet/filter/lifecycle i prostego scenariusza kontenera. Nie uznawać każdego publicznego helpera za endpoint.
- Katalog punktów wejścia może wskazać zewnętrzny parser/dispatch przed kodem aplikacji. Pokazywać podstawę i warunki aktywacji; nie przypisywać dowolnemu framework CVE pewnej ścieżki wyłącznie na podstawie obecności pakietu.
- Przekazać tożsamość parametrów wejściowych do przyszłych źródeł taint D4. Samo `roots=all` nie tworzy takich źródeł.

Gotowe: jeden bean/qualifier, konflikt, nieznany profil i kontroler z helperem mają właściwe wyniki. F0–F2 to pierwszy zakres frameworków.

Wycinek „dispatch kontenera servletów” (2026-11-08): model B3 `servlet-container:dispatch` (`internal/servletdispatch`, wersja `s1`) — ograniczony wyjątek od pominięcia odbiorców `javax.`/`jakarta.` w CHA dla `Servlet#service`, `Filter#doFilter` i wewnętrznego dispatchu szablonowego `HttpServlet` (`service(HttpServletRequest, …)`, `doGet`/`doPost`/…, tylko gdy wołającym jest sam `HttpServlet`). Cele: wyłącznie servlety i filtry zarejestrowane przez aplikację (web.xml, `@WebServlet`/`@WebFilter`, w aplikacji Spring Boot beany będące servletem/filtrem oraz katalog autokonfiguracji Boot sprawdzony w jej źródłach, z warunkami `@ConditionalOnClass` i `exclude`), metoda wybrana przez JVM na zarejestrowanej klasie; krawędzie z podstawą `model`, rejestracje z własną podstawą w logu; rejestracja nieznanej klasy (`FilterRegistrationBean` itp.) to niekompletny fakt modelu, nie zgadywanie. Wyrocznia JVM `testdata/servletdispatch` (`scripts/servletdispatch-oracle.sh`). Szczegóły i wpływ na petclinic/WebGoat: IMPLEMENTATION_STATUS, wpis FRAMEWORKS:F2 (dispatch kontenera servletów), oraz `testdata/graphbench/RESULTS.MD`.

## F3. Proxy i ograniczony model AOP — DONE for this scope (2026-11-23; odłożone: ogólny JDK Proxy; [zarchiwizowane](FRAMEWORKS_F3.md))

## F4. Rozszerzanie wejść i frameworków — DONE for this scope (2026-11-23; odłożone: listenery wiadomości, SpEL, osobny tryb klasyfikacji; [zarchiwizowane](FRAMEWORKS_F4.md))

## F5. Kwalifikacja — DONE for this scope (2026-11-23, `f5-1`…`f5-3`; odłożone: konsumenci wiadomości, priorytety konfiguracji; [zarchiwizowane](FRAMEWORKS_F5.md))
