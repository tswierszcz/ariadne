# Runtime, dynamiczne klasy i JNI — rekomendacja funkcjonalna

Status: plan. Fazy: [RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD](RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD). Ocena podatności wydania JDK pozostaje w [CVE](CVE.MD).

## Cel i stan obecny

Uzupełniać analizę statyczną informacjami o rzeczywistym ładowaniu klas i wywołaniach, a w obszarze kodu natywnego ujawniać granice modelu i wykorzystywać jawne streszczenia.

Obecna Ariadne jest analizatorem statycznym napisanym w Go. Nie ma agenta Java ani importu śladów; hierarchia identyfikuje klasy przez FQCN i nie rozróżnia niezależnych definicji przez różne loadery.

## Trzy przyrosty

### Import dowodów wykonania

Przyjąć plik śladu z metadanymi przebiegu, runtime’u, konfiguracji i hashem artefaktów. Odróżniać:

- klasę załadowaną, ale bez dowodu wykonania metody;
- wykonanie konkretnej metody;
- zaobserwowane bezpośrednie wywołanie caller → callee;
- relację zadania/zdarzenia przekazanego do innego wątku;
- informację próbkowaną, brakującą lub utraconą.

Ślad bez informacji o call site nie staje się automatycznie pełnym grafem wywołań. Zbieżność czasowa zdarzeń nie dowodzi związku przyczynowego.

### Opcjonalny agent Java i dynamiczne klasy

Agent uruchomiony przez użytkownika rejestruje zadeklarowany zakres zdarzeń i może opcjonalnie dostarczać metadane/bytecode generowanych klas. Statyczna analiza zwykłych JAR-ów nadal działa bez JDK.

Tożsamość klasy obejmuje kontekst loadera i zawartość. Dwie klasy o tej samej nazwie z różnymi loaderami nie są scalane. Identyfikator loadera z jednego przebiegu nie gwarantuje tożsamości w innym przebiegu.

Dodatkowe znane plugin JAR-y można także podać statycznie z jawną konfiguracją loaderów. Sama inwentaryzacja pluginu nie dowodzi jego aktywacji. Nieznany loader albo kod pobierany dopiero w runtime pozostaje luką.

### Granice JNI i modele natywne

Rozpoznawać deklaracje native, wywołania ładowania bibliotek i obsługiwane streszczenia Java ↔ native, w tym callback do Java. Jeśli brak modelu, raport wskazuje nieprześledzoną granicę. Zwykły bytecode parser nie analizuje automatycznie kodu C/C++ ani wszystkich dynamicznych rejestracji JNI.

## Prezentacja i ograniczenia

Wyniki statyczne i runtime są widoczne obok siebie: możliwa ścieżka, zaobserwowana metoda/relacja i zgodność z konkretnym artefaktem. Obserwacja wykonania kodu nie jest potwierdzeniem wykorzystania podatności.

Brak zdarzenia nigdy nie dowodzi nieosiągalności. Ślad zależy od ruchu, scenariusza, czasu, konfiguracji i instrumentacji. Niepełne lub próbkowane dane mają jawne oznaczenie; gubienie zdarzeń ogranicza siłę wniosków.

Domyślny zakres danych agenta obejmuje identyfikatory i relacje, nie treść argumentów i payloady. Rozmiar śladu, wpływ na aplikację i zakres instrumentacji są mierzone i raportowane, a sam użytkownik wskazuje uruchamiany scenariusz.

## Gotowość i źródła

MVP: wersjonowany importer z poprawnym dopasowaniem artefaktów i bez negatywnych wniosków z braku zdarzeń. Następnie agent dla małego katalogu klas/API. Pełny tracing native i dowolnych hidden/generated classes to odrębne rozszerzenia eksperymentalne.

- [Instrumentation: agenci Java](https://docs.oracle.com/en/java/javase/25/docs/api/java.instrument/java/lang/instrument/package-summary.html).
- [JVM: tożsamość i ładowanie klas](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-5.html#jvms-5.3).
- [JNI: funkcje i przejścia między środowiskami](https://docs.oracle.com/en/java/javase/25/docs/specs/jni/functions.html).
