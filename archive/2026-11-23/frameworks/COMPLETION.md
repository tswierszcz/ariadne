# FRAMEWORKS — obszar DONE (2026-11-23)

Plan i opis obszaru: [FRAMEWORKS.MD](FRAMEWORKS.MD), [FRAMEWORKS_IMPLEMENTATION_PLAN.MD](FRAMEWORKS_IMPLEMENTATION_PLAN.MD). Archiwa faz: [F1](../../2026-10-23/frameworks/FRAMEWORKS_F1.md), [F3](FRAMEWORKS_F3.md), [F4](FRAMEWORKS_F4.md), [F5](FRAMEWORKS_F5.md).

| Faza | Stan | Najważniejsze |
|---|---|---|
| F0 | DONE | wykrycie frameworka ≠ aktywny profil; kontrakt scenariusza z warunkiem true/false/unknown |
| F1 | DONE (2026-10-23) | wartości adnotacji, trasy Spring MVC, nazwy beanów, `@Profile`, ograniczony `web.xml`/XML |
| F2 | DONE for this scope | DI (typ/qualifier/primary), korzenie `web.xml`, `@Bean`, katalog wejść, Spring Data, `@Value`, dispatch kontenera servletów, obiekty żądań, appendery logback/log4j, JDBC |
| F3 | DONE for this scope | proxy (MyBatis, Spring Data, Spring ORM, `@Transactional`), Spring AOP; odłożony ogólny JDK Proxy |
| F4 | DONE for this scope | zdarzenia i `@Scheduled`, profile innych frameworków, wersjonowana lista wejść; odłożone listenery wiadomości, SpEL |
| F5 | DONE for this scope | wersja reguł w snapshotcie, wyrocznie Spring MVC i scenariuszy DI; 8 z 11 przypadków z wyrocznią JVM |

**Weryfikacja:** wyrocznie JVM `di`, `xmlbeans`, `mvc`, `aop`, `events`, `tx`, `jpa`, `springdata`, `mybatis`, `servletdispatch`; pomiary na petclinic/WebGoat/JPetStore — wpisy FRAMEWORKS w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`. Fakty frameworka w v2 (`frameworks`).

**Jawne ograniczenia:** ogólny JDK Proxy, listenery wiadomości (broker), priorytety konfiguracji, SpEL i programowalne rejestracje, kandydaci wstrzyknięć bez filtra `@Profile`, skanowanie komponentów nieczytane; wiersze `docs/SUPPORT.MD` pozostają `Partial`/`Implemented` jak dotąd.
