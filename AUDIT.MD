# Audyt projektu Ariadne

## Zakres i metoda

Audyt objął dokumentację i statyczny przegląd ścieżki: wejście → hierarchia klas → graf wywołań → OSV → mapowanie CVE na metody → reachability → raport. Nie uruchamiano testów ani analiz przykładowych aplikacji. Ustalenia wskazują na możliwe błędy wyników lub niepełne deklaracje; nie są potwierdzeniem błędu runtime.

## Błędy i ryzyka

### 1. [Wysoki] Niejednoznaczne klasy na classpath mogą zmieniać wynik zależnie od kolejności odczytu

`internal/hierarchy.Build` odczytuje JAR-y równolegle, a przy duplikacie klasy zachowuje pierwszą odpowiedź, która dotrze z workerów (`internal/hierarchy/build.go`, okolice linii 52–80). Wybrana implementacja nie musi być stała między uruchomieniami; może zmienić graf i status CVE.

**Wpływ:** niestabilne wyniki, zwłaszcza przy konfliktach wersji lub klasach zduplikowanych w JAR-ach.

**Usprawnienie:** ustalić i udokumentować kolejność classpath, np. aplikacja przed zależnościami w jawnej kolejności, oraz raportować konflikt jako wymagający uwagi.

### 2. [Wysoki, warunkowy] Parser zawsze analizuje bazowe klasy z multi-release JAR-a

`internal/hierarchy/jarread.go` pomija wszystkie klasy z `META-INF/versions/`. Multi-release JAR może jednak dostarczać inną implementację zależnie od wersji Javy uruchamiającej aplikację.

**Wpływ:** dla takich zależności analiza może badać kod inny niż faktycznie uruchamiany, dając fałszywe wyniki w obie strony.

**Usprawnienie:** przyjmować wersję docelowego runtime’u i wybierać najwyższą obsługiwaną wersję; jeśli jej nie podano, jawnie oznaczać wynik jako niepewny.

### 3. [Wysoki, warunkowy] Łączenie modułów EAR może pomijać zależności albo klasy aplikacji

Przy ekstrakcji bibliotek `internal/input/source.go` używa `filepath.Base(...)`, a istniejący plik o tej samej nazwie pomija. Dwa różne JAR-y o tej samej nazwie mogą więc zostać scalone w jeden. Dodatkowo top-level JAR-y EAR są traktowane jako zależności, nie jako kod aplikacji. Dokumentacja wymienia drugie ograniczenie, ale nie wyjaśnia ryzyka kolizji nazw.

**Wpływ:** pominięte klasy lub zależności mogą prowadzić do brakujących CVE albo błędnej reachability.

**Usprawnienie:** zachować unikalną ścieżkę/własność modułu i zgłaszać kolizje; dla EAR obsługiwać moduły na podstawie deskryptora aplikacji.

### 4. [Wysoki dla `-roots profile`] Wykrywanie frameworków i warunków profilu opiera się na obecności klasy w analizowanym zbiorze

`internal/profile/profile.go` sprawdza obecność klas przez `h.Lookup`, bez sprawdzenia `IsApp`; profile mogą więc zostać aktywowane przez zależności. Z kolei `spring.yaml` warunkuje Jacksona adnotacją aplikacji, ale `appHasAnyAnnotation` sprawdza tylko adnotacje klasy i metody, nie parametrów.

**Wpływ:** przy `-roots profile` niepasujący profil może zmienić zbiory frameworków i korzeni; aplikacja używająca wyłącznie `@RequestBody` na parametrze może nie aktywować warunku Jacksona.

**Usprawnienie:** doprecyzować, które markery muszą pochodzić z kodu aplikacji, i uwzględniać adnotacje parametrów w warunku Jacksona.

### 5. [Średni] Mapowanie z diffów może przypisać CVE metody zmienione przez commit niezwiązany bezpośrednio z poprawką podatności

`internal/vulnmap/diffsource.go` próbuje każdy link GitHub do commita, także referencje nieoznaczone jako `FIX`, a następnie łączy metody ze wszystkich commitów.

**Wpływ:** metoda może zostać uznana za podatną na podstawie szerszego lub błędnie sklasyfikowanego commita, co zmienia status CVE.

**Usprawnienie:** rozróżniać dowody według typu referencji i commitów/gałęzi, a wynik diffu raportować z listą źródłowych commitów.

### 6. [Średni] Aplikacje Spring z `@Bean` mogą uznać metody konfiguracyjne za korzenie w `-roots profile`

Profil Spring traktuje `@Bean` jako adnotację korzenia (`internal/profile/profiles/spring.yaml`). Reguła obejmuje każdą metodę z adnotacją, niezależnie od widoczności; `@Bean` często występuje na metodach konfiguracyjnych, a nie na żądaniach użytkownika.

**Wpływ:** rozszerza zbiór osiągalnych ścieżek i może zawyżać reachability w trybie profilu.

**Usprawnienie:** oddzielić korzenie cyklu życia/konfiguracji od wejść obsługujących ruch i określić, które mają być traktowane jako źródła atakującego.

## Usprawnienia audytowalności

- Raportować konflikty klas oraz ostrzeżenia identyfikacji zależności. `ResolveDepJars` gromadzi `DepJar.Warnings` (`internal/input/resolve.go`), ale w przepływie CLI raportowane są niezidentyfikowane JAR-y oraz ostrzeżenia z hierarchii/LLM. Wybór GAV przy wielu `pom.properties` może być niewidoczny dla osoby oceniającej wynik.
- W raporcie podawać wersję docelowej Javy i ograniczenie multi-release JAR-ów. Czytelnik nie wie wtedy, czy analiza odpowiada runtime’owi aplikacji.
- Ujednolicić pomoc CLI z README. `-profile` i `-profile-file` w `cmd/ariadne/main.go` nadal opisują tylko Spring/generic, choć kod ładuje sześć profili, a README wymienia wszystkie.

## Zakres i ograniczenia audytu

README wymienia mechanizmy niewidoczne dla analizy, m.in. refleksję, część proxy i konfigurację. `NOT_REACHABLE` należy zatem rozumieć jako „nieosiągalne w modelowanym grafie”, a nie dowód braku podatnego przepływu. `ServiceLoader` pozostaje jawnie poza zakresem; domknięcie tej luki występuje w przyszłym planie.

Nie sprawdzono poprawności konkretnych CVE ani zachowania na rzeczywistych aplikacjach, ponieważ wymagałoby to uruchomienia analizy i porównania z ich runtime/classpath.
