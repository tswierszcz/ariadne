# FRAMEWORKS:F1 — Adnotacje, zasoby i jawna konfiguracja — archiwum

**Data archiwizacji:** 2026-10-23.
**Stan:** DONE dla faktycznie zrealizowanego zakresu (patrz "Co nie zostało zrobione"
poniżej — to nie jest 1:1 realizacja każdego punktu z oryginalnego planu F1).

Oryginalny tekst fazy F1 (sketch sprzed implementacji) był w
`RECOMMENDATION/FRAMEWORKS_IMPLEMENTATION_PLAN.MD`; zastąpiony w tym pliku
wskaźnikiem do tego archiwum. Pełny, bieżący opis zachowania użytkownika jest w
`docs/SUPPORT.MD` (wiersze "Profiles and selected roots" i "Configuration (web.xml,
XML bean definitions)"); ten plik jest wyłącznie zapisem historycznym, nie
źródłem prawdy o bieżącym stanie.

## Commity

1. `31052b5` — Część 1: model wartości adnotacji (`internal/classfile`).
2. `b00c762` — Część 2: rozwiązywanie między klasami (`internal/annotation`, nowy pakiet).
3. `9c1da4a` — Część 3: realne wartości Spring MVC (`internal/profile/springvalues.go`) + flaga `-active-profiles`.
4. `ec1ae4e` — Część 4: ograniczony, bezpieczny odczyt `web.xml` (`internal/webxml`, nowy pakiet).
5. `0438c55` — Część 5: XML bean definitions (`internal/springxml`, nowy pakiet) + wykrywanie konfliktów nazw beanów.

## Co zostało zrobione

- **Model wartości adnotacji** (`internal/classfile/annotations.go`, `pool.go`,
  `poolread.go`, `classfile.go`): każdy kształt `element_value` (JVM spec
  4.7.16.1) — stałe proste, enum, literał klasy, tablica, zagnieżdżona adnotacja
  — zamiast tylko nazwy. `RuntimeInvisibleAnnotations` czytane i scalane.
  Adnotacje na polach (wcześniej brak). `AnnotationDefault` przechowywane.
  Stare listy FQCN zachowane przez adapter (`annotationNames`) — żaden
  istniejący konsument nie wymagał zmiany.
- **`internal/annotation`** (nowy pakiet): `ResolveDefault`/`EffectiveValue`
  (wartość domyślna elementu), `ClassAnnotations` (ponowny odczyt klasy przez
  hierarchię), `RepeatableContainer`/`ExpandRepeatable` (rozpakowanie
  `@Repeatable` z weryfikacją, nie tylko po kształcie tablicy).
- **`internal/profile/springvalues.go`**: `Routes` (prawdziwa ścieżka/ścieżki i
  czasownik HTTP każdego mapowania Spring MVC), `BeanNames` (deklarowana nazwa
  beana + `@Primary`), `ProfileConditions` (deklarowane nazwy `@Profile`,
  włącznie z negacją `!x`) + nowa flaga CLI `-active-profiles` (brak flagi →
  `ConditionUnknown`, nigdy zgadywane `false`).
- **`internal/webxml`** (nowy pakiet) + `internal/input.PreparedApp.WebXMLs`:
  ograniczony, bezpieczny odczyt `WEB-INF/web.xml` (WAR i każdy moduł EAR) —
  `<servlet>`/`<servlet-mapping>`/`<filter>`/`<filter-mapping>`/`<listener>`/
  `<context-param>`, z resztą elementów nazwaną w `Skipped`, nigdy po prostu
  odrzucaną.
- **`internal/springxml`** (nowy pakiet) + `internal/input.FindBeansXML`:
  ograniczony odczyt klasycznego Spring XML wiring (`<bean id="..."
  class="...">`), skanowanie tylko własnych wpisów `app-classes.jar` (nigdy
  zależności).
- **Wykrywanie konfliktów** (`internal/report`'s `buildConfigConflicts`):
  bean nazwany przez więcej niż jedno źródło (adnotacja, XML) z **inną**
  klasą — zgłoszony jako kandydaci, nigdy scalony w jednego zwycięzcę.
- Pięć nowych sekcji raportu: "Spring MVC routes", "Spring bean names",
  "Spring @Profile conditions", "web.xml", "Spring XML bean definitions",
  "Configuration conflicts".

## Dowód

`bin/ariadne -offline` na petclinic: 17 realnych tras HTTP (np.
`OwnerController#showOwner` → `/owners/{ownerId}` GET), 7 stereotypów beanów
(wszystkie bez jawnej nazwy — poprawnie, nie zgadywane). WebGoat: 19 tras, 20
stereotypów. Realny WAR zbudowany z już skompilowanego `app-classes.jar`
petclinic (to środowisko nie ma JDK) + ręcznie napisany `web.xml` i
`applicationContext.xml`: wszystkie elementy poprawnie wyekstrahowane, elementy
nieznane poprawnie nazwane w "Skipped", zero fałszywych konfliktów. Dwa
przebiegi na każdym z tych wejść dają bajt-identyczny `results.json`. Zero
zmian w REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN na petclinic/WebGoat w całym
ciągu pięciu przyrostów. `go build/vet/test ./...` i `go test -race ./...`
zielone po każdym przyroście.

## Co NIE zostało zrobione (celowo, nie przez przeoczenie)

- **Rozstrzyganie konfliktów** (wybór zwycięzcy) — tylko detekcja. Prawdziwe
  zachowanie Springa (`allow-bean-definition-overriding`) jest konfigurowalne
  i nie jest modelowane.
- **Pełny priorytet wielu źródeł konfiguracji** — nie ma mechanizmu łączącego
  `web.xml`/adnotacje/XML beany/properties w jeden scenariusz z priorytetem,
  tylko detekcja konfliktu nazw beanów.
- **Jawne wejście plików/właściwości scenariusza** poza `-active-profiles` —
  nietknięte (np. wskazanie konkretnego `applicationContext.xml` lub
  `application.properties` z linii komend).
- **Zasoby i pochodzenie modułu** w `internal/input` — nietknięte poza samym
  `web.xml`/XML beanów capture.
- **`<property>`/`<constructor-arg>` wiring** w XML beanach — nic ich nie
  konsumuje, nie zostały zamodelowane.
- **Krawędzie grafu wywołań** z deklaracji `web.xml`/XML beanów (serwlet/bean
  jako root) — celowo nietknięte, inny krok niż inwentarz.
- **DI/wiring** (`@Autowired`/`@Qualifier` na punkcie wstrzyknięcia,
  konstruktor/pole) — to `FRAMEWORKS:F2`'s własny zakres, zależny od tego, co
  F1 tutaj dostarczył (`BeanNames`, inwentarz XML beanów).

## Następny krok

`FRAMEWORKS:F2` (DI, `@Bean`, wiązanie qualifier/primary) — może korzystać z
`BeanNames` i inwentarza XML beanów jako dodatkowych źródeł. Wymaga propozycji
i zgody użytkownika przed startem, jak każda następna faza.
