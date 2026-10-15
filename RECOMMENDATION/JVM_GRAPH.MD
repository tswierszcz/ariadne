# Podstawy grafu JVM — rekomendacja funkcjonalna

Status: plan. Implementacja: [JVM_GRAPH_IMPLEMENTATION_PLAN.MD](JVM_GRAPH_IMPLEMENTATION_PLAN.MD). Wspólne kontrakty: [ROADMAP.MD](ROADMAP.MD).

## Cel i stan obecny

Graf powinien prowadzić do ciała metody właściwego dla analizowanego wejścia i uwzględniać niejawne uruchomienie inicjalizatorów. Obecne `internal/classfile/calls.go` wyciąga wywołania `invoke*`; `internal/callgraph/build.go` wpisuje ich właściciela bez pełnego rozwiązania implementacji i rozszerza CHA po nazwie. `internal/hierarchy/jarread.go` pomija warianty multi-release. To wymaga wspólnego uporządkowania przed szerszym modelowaniem dynamiki.

## Zachowanie docelowe

1. Analiza zapisuje kolejność classpath i pochodzenie każdej wybranej klasy. Kolejność zakończenia równoległego odczytu nie decyduje o wyborze.
2. Dla wywołania odziedziczonego znajduje implementację zgodną z rodzajem instrukcji i deskryptorem. Nie kończy ścieżki na pustym symbolu podklasy, jeśli ciało metody jest dostępne w nadklasie.
3. Rozróżnia przeciążenia, metody bridge, konstruktory, wywołania `super`, statyczne i interfejsowe. Nie łączy metod wyłącznie na podstawie nazwy.
4. Dodaje relacje inicjalizacji przy odpowiednich operacjach użycia klasy; uwzględnia reguły nadklas i interfejsów zamiast inicjalizować wszystkie typy na classpath.
5. Dla multi-release JAR-a wybiera wariant zgodny z zadeklarowaną wersją docelowej Javy i manifestem. Bez targetu jawnie raportuje przyjęte ograniczenie.
6. Oznacza niezidentyfikowane klasy, brakujące ciała metod i konflikty, które mogą zmienić wynik.

### Przykłady

- `Child` dziedziczy `Parent.process()`, a metoda ta woła podatny parser: ścieżka przechodzi przez implementację `Parent.process()`.
- Odczyt niebędącego stałą pola statycznego może uruchomić `<clinit>` klasy deklarującej to pole; graf pokazuje ten warunek.
- Stała wkompilowana w kod klienta nie powoduje dodania inicjalizacji na podstawie samej nazwy pola w źródle.
- JAR z wariantem klasy dla Java 17 jest analizowany inaczej dla targetów 11 i 17, jeśli manifest i zawartość to uzasadniają.

## Prezentacja i ograniczenia

Raport pokazuje docelową wersję Java, regułę wyboru klas oraz ścieżki z krokiem `initializes` lub rozwiązaniem odziedziczonej metody. Nie twierdzi, że dokładnie odtworzono kolejność wykonania inicjalizacji lub jej jednorazowość przez zwykły graf możliwości.

Reguły klas należy oprzeć na specyfikacji JVM; nie zastępować ich wyszukiwaniem pierwszej metody o tej samej nazwie. Niestandardowe loadery i izolacja modułów EAR wymagają osobnego modelu lub jawnego ograniczenia. Ten plan ustala bazowy kontekst classpath i zgłasza nieznane konteksty.

## Zależności i gotowość

Wspólne części tożsamości klas i dokładnych metod z [CVE F1/F4](CVE_IMPLEMENTATION_PLAN.MD) realizuje się jeden raz. Ta sama informacja o target feature version służy do wyboru klas; nie jest identyfikatorem poziomu poprawek JDK.

Gotowość: stabilny wybór klas, poprawne ścieżki dziedziczenia i inicjalizacji oraz fixtures z pozytywnymi i negatywnymi przypadkami wariantów JAR. Zmiany liczby CVE muszą wynikać z wyjaśnionej zmiany grafu.

## Źródła

- [JVM: rozwiązywanie metod i inicjalizacja](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-5.html).
- [Format JAR i multi-release](https://docs.oracle.com/en/java/javase/25/docs/specs/jar/jar.html).

Źródła wyjaśniają mechanizmy; obsługiwane wersje należy osobno potwierdzić w fixtures.
