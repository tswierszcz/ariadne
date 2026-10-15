# Frameworki, proxy i konfiguracja — rekomendacja funkcjonalna

Status: plan. Fazy: [FRAMEWORKS_IMPLEMENTATION_PLAN.MD](FRAMEWORKS_IMPLEMENTATION_PLAN.MD). Wspólne zasady: [ROADMAP.MD](ROADMAP.MD).

## Cel i stan obecny

Odtwarzać wybrane wejścia aplikacji i powiązania tworzone przez framework, wykorzystując jego wersję, metadane i zadeklarowany scenariusz konfiguracji.

`internal/profile` wykrywa sześć profili, dobiera korzenie, pakiety frameworka i reguły syntetyczne. `internal/classfile/annotations.go` zachowuje nazwy adnotacji, ale pomija ich wartości. To wystarcza do obecnych prostych profili, lecz nie do odtworzenia nazw beanów, qualifierów, warunków, tras i pointcutów.

Obecność frameworka na classpath oznacza dostępność mechanizmu, a nie potwierdzenie, że konkretny endpoint lub bean jest aktywny.

## Funkcje docelowe

- Odczyt wartości adnotacji, ich jawnie obsługiwanych metadanych oraz dostępnych zasobów konfiguracji.
- Scenariusz konfiguracji: aktywne profile, wskazane pliki i właściwości, ich priorytet oraz nierozstrzygnięte warunki. Nieznane wartości nie są automatycznie false.
- Ograniczony model DI: definicje beanów, typy, fabryki, konstruktory, qualifier i możliwe wiązania. Niejednoznaczne wstrzyknięcie pozostaje oznaczone.
- Proxy: powiązanie obiektu proxy, implementacji docelowej i znanych interceptorów. JDK proxy i CGLIB wymagają odmiennych reguł; nie każdy interfejs prowadzi do dowolnej klasy o podobnej nazwie.
- AOP: wybrane pointcuty i typy advice, uwzględnienie self-invocation oraz ograniczeń override. Nieznany pointcut jest warunkiem/luką.
- Wejścia: wybrane HTTP routes, filtry/listenery, lifecycle, zadania i konsumenci wiadomości. Każdy mechanizm ma osobny wersjonowany model.
- Konfiguracja deklaratywna: pierwszy zakres `web.xml` i ograniczony XML bean definitions; programowalne konfiguracje i wyrażenia tylko w zadeklarowanym podzbiorze.

### Scenariusze

1. Wstrzyknięty interfejs ma dwa beany, qualifier wskazuje jeden: analiza zachowuje uzasadnione wiązanie dla scenariusza.
2. Nieznany aktywny profil dopuszcza dwa warianty: raport pokazuje oba z warunkami, zamiast wybierać wygodniejszy.
3. Wywołanie przez proxy przechodzi przez advice, a zwykłe self-invocation wewnątrz obiektu nie dostaje takiej samej ścieżki bez podstawy modelu.
4. Parser HTTP kontenera może być używany przed kontrolerem: zewnętrzny model wejścia pokazuje osobną ścieżkę, zamiast szukać fikcyjnego wywołania parsera przez kontroler.

## Wyniki i integracja

Raport podaje aktywny model i jego wersję, scenariusz, źródło definicji, możliwe wiązania i warunki. Korzenie są oznaczane jako zewnętrzne, lifecycle, konfiguracyjne albo wszystkie metody aplikacji; nie są równoważnym dowodem dopływu danych atakującego.

MVP: Spring na wybranej wersji z prostymi bean definitions i jawnie wskazanym kontekstem. Kolejne wersje i inne frameworki wymagają osobnych fixtures. Włączenie modelowanych wejść jest jawne; dotychczasowa semantyka dwóch BFS i `FRAMEWORK` ma opisaną migrację.

## Granice i gotowość

Nie odtwarzamy pełnego wykonania kontenera ani wszystkich efektów warunkowej konfiguracji. Bez wskazanego środowiska raportujemy scenariusze możliwe. Ograniczenie modelu nie może służyć jako automatyczny dowód nieaktywności podatnego kodu.

Gotowość: bean z jednoznacznym qualifierem, konflikt, aktywny/nieznany warunek, proxy vs self-invocation i zewnętrzna ścieżka kontenera są rozróżnione.

## Źródła

- [Spring: proxy i ograniczenia AOP](https://docs.spring.io/spring-framework/reference/core/aop/proxying.html).
- [JDK Proxy](https://docs.oracle.com/en/java/javase/25/docs/api/java.base/java/lang/reflect/Proxy.html).

Bieżąca dokumentacja opisuje mechanizmy. Zachowanie wersji użytych w fixtures projektu musi zostać sprawdzone osobno; nie zakładać identyczności starych i nowych wydań Spring.
