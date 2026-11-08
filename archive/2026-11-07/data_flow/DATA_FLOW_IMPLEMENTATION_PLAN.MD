# Przepływ wartości i danych — fazy D0–D5

Status: **DONE (D0–D5), zarchiwizowane 2026-11-07** — zob. [COMPLETION.md](COMPLETION.md). Wymagania: [DATA_FLOW.MD](DATA_FLOW.MD). Wspólne zasady: [ROADMAP.MD](../../../RECOMMENDATION/ROADMAP.MD). Szczegółowe wpisy faz: [IMPLEMENTATION_STATUS.MD](../../../RECOMMENDATION/IMPLEMENTATION_STATUS.MD).

## D0. Kontrakt IR, wartości i zakresów

Zależności: B0–B1, JVM G0–G2.

- Zaprojektować `internal/bytecodeir` i `internal/valueflow` jako proponowane pakiety wspólne; zachować metodę, deskryptor, offset, rodzaj operacji i pochodzenie.
- Określić domenę wartości: stałe, typy, identyfikatory alokacji, `null`, zbiory i `unknown`; rozdzielić dokładnie znaną stałą od jedynie górnego ograniczenia typu.
- Określić kolejność poszerzania i łączenia stanów oraz budżety. Nadmiar kandydatów zastępować odpowiednim stanem unknown z powodem, nie przyciętą listą udającą pełną.
- Zaprojektować API faktów dla modeli B3, tak aby nie zależało od CVE ani konkretnego frameworka.

Gotowe: kontrakt ma przykłady gałęzi, pętli, wywołania i nieobsługiwanej operacji; nie wymaga ukończenia taint.

## D1. CFG i semantyka kodu metody

- Rozszerzyć `internal/classfile/code.go` o zachowanie tabeli wyjątków i niezbędnych metadanych; uzgodnić współdzielenie z JVM G3.
- Udostępnić instrukcje i cele skoków z `opcodes.go`, walidować granice, switche i zgodność przejść. Reprezentować podstawowe bloki oraz normalne i wyjątkowe krawędzie.
- Modelować kategorie jedno-/dwusłowowe stosu, zmienne lokalne, `dup`/`swap`, obiekty niezainicjalizowane i argumenty wynikające z deskryptora.
- Nie mylić rzuconego wyjątku z normalnym wynikiem metody. Stare `jsr`/`ret` i inne nieobsługiwane warianty albo wdrożyć, albo oznaczać jako nieobsługiwane.
- Zachować opcjonalne numery linii do prezentacji; brak debug data nie zatrzymuje analizy.

Gotowe: CFG i stany poprawne dla gałęzi, pętli, wyjątków i szerokich typów; niepoprawne wejścia nie powodują paniki ani fikcyjnego braku przepływu.

## D2. Lokalne wartości i udostępnienie modelom

Zależności: D1.

- Wprowadzić monotoniczną analizę lokalną: stałe napisowe/klasowe, argumenty, przypisania, podstawowe tablice i obiekty.
- Dla odczytów/zapisów pól i elementów tablic określić początkowy zakres precyzji. Gdy nie można rozdzielić obiektów, zachować konserwatywne połączenie i informację o nim.
- Wprowadzić mały katalog streszczeń operacji niezbędnych do przenoszenia wartości; nie analizować całego JDK bez ograniczeń.
- Naprawić lub wyraźnie obsłużyć wyrównanie adnotacji parametrów do indeksów deskryptora przy syntetycznych parametrach, zanim będą używane jako źródła danych.
- Udostępnić fakty ServiceLoaderowi, refleksji, callbackom i uchwytom; każdy model otrzymuje identyfikator wartości/obiektu i miejsce wywołania.

Gotowe: literał przez zmienną, dwa warianty gałęzi i nieznany parametr dają właściwe fakty. Pierwszy przyrost B2 można wydać bez D4.

## D3. Przepływ między metodami i streszczenia — DONE (2026-11-07; streszczenia zależności `internal/flowsum`; points-to dla pól instancyjnych poza zakresem)

Zależności: D2, rejestr B3 i dokładny graf JVM G2.

- Zapisywać zależności parametr → wynik/pole oraz wartości zwracane przez proste fabryki. Uwzględnić rekurencję i aktualizację grafu.
- Wprowadzić klucz kontekstu/instancji wystarczający dla przyjętego zakresu, aby wyniki dwóch wywołań nie były dowolnie mieszane.
- Przenosić dane przez modele biblioteczne i zarejestrowane callbacki; nie zakładać, że wszystkie wartości przechodzą przez każdą metodę w identyczny sposób.
- Nieznana zewnętrzna metoda używa jawnej konserwatywnej polityki; zachować pochodzenie poszerzenia i niedostępne ciało.
- Wersjonować streszczenia i unieważniać je po zmianie kodu, modelu, targetu i konfiguracji; wykorzystać snapshoty CVE F3.

Gotowe: fabryka zwraca właściwą wartość, dwa call sites nie tworzą nieuzasadnionej ścieżki, rekurencja zbiega lub ujawnia limit.

## D4. Pierwsza analiza taint — DONE (zarchiwizowane 2026-11-07)

Pilotaż (dane żądania Spring MVC/JAX-RS/Servlet → dane czytane przez wejścia
deserializacji, `-taint` opt-in, tylko raport) zrealizowany i zweryfikowany;
zobacz [archive/2026-11-07/data_flow/DATA_FLOW_D4.md](DATA_FLOW_D4.md)
po pełny zapis (co zrobione, dowody, co celowo odłożone — m.in. kroki
transformacji wewnątrz metody) i `docs/SUPPORT.MD` po bieżący opis zachowania.

## D5. Pomiar i wydanie — DONE (2026-11-07; pilotaż taint: wyrocznia JVM `scripts/taint-oracle.sh`, `TestTaintOracle`/`TestShapes`, pomiar w `docs/TAINT_QUALIFICATION.MD`; `-taint` opt-in)

- Fixtures obejmują aliasing, pola, tablice, wyjątki, dwie instancje, overload, callback i brak źródeł. Zapisane artefakty pozwalają testować Go bez JDK.
- Dla konkretnej wersji biblioteki i konkretnego pilota mierzyć błędne przepływy, pominięcia, unknown, czas oraz pamięć. Nie wyprowadzać globalnego procentu skuteczności z kilku fixtures.
- Porównać z kontrolowanym uruchomieniem lub niezależnie sprawdzonym przykładem; brak zaobserwowanego flow nie dowodzi jego niemożliwości.
- Zastosować B4; opublikować obsługiwane instrukcje, biblioteki i granice. Nowy tryb taint pozostaje opt-in do kwalifikacji.

Wykonano (2026-11-07): osiem kształtów w `testdata/taint` (aliasing, pola, tablice, wyjątki, dwie instancje, overload, callback, brak źródeł); wyrocznia prawdziwego JVM (kanarek jako dane żądania, filtr procesowy) — 21 wejść: 0 pominięć, 0 niepotwierdzonych `flow_found`, 4 zaobserwowane przepływy jako `undetermined`; XStream sprawdzony na przykładach niezależnie zweryfikowanych (WebGoat, demo łańcuchów); czas i pamięć na 5 wejściach. Decyzja użytkownika: `-taint` pozostaje opt-in. Szczegóły: IMPLEMENTATION_STATUS, wpis DATA_FLOW:D5.

Zaawansowane heap/context sensitivity i analiza wszystkich interleavingów wątków są dalszymi etapami wymagającymi osobnego pomiaru, a nie ukrytym warunkiem MVP.
