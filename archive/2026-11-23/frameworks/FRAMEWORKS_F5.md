# FRAMEWORKS:F5 — DONE for this scope (2026-11-23)

## Zapis ukończenia

- **Przyrosty:** ocena (2026-11-22, PR #126), `f5-1` (PR #127), `f5-2` (PR #128), `f5-3` (PR #129).
- **Co dostarczono:**
  - `f5-1` — reguły korzeni/profili w snapshotcie: `profile.rules` (`profile.RulesVersion`) i `profile.builtin` (skrót wbudowanych YAML-i); `TestRulesVersion_tracksRuleSource` wymusza podbicie wersji przy zmianie kodu reguł (zadziałał w `f5-2` i `f5-3`).
  - `f5-2` — wyrocznia JVM Spring MVC `testdata/mvc` (wbudowany Tomcat 9.0.45 → `DispatcherServlet` → kontrolery, prawdziwe żądania HTTP): 14 z 26 ciał nieosiągalnych; naprawione przez adnotacje `@ExceptionHandler`/`@ModelAttribute`/`@InitBinder`, `overrides` `WebMvcConfigurer`/`HandlerInterceptor` i `roots.web_bodies` (gettery/konstruktor/settery ciał Jacksona).
  - `f5-3` — wyrocznia DI w dwóch scenariuszach profili i ze sprawdzeniem wstrzyknięć: 2 fałszywe negatywy (wyrażenia `@Profile` Springa 5.1+) i 2 błędne rozwiązania qualifierów (nazwa domyślna, `@Qualifier` klasy) naprawione.
- **Kryteria F5 → stan:**
  - Fixtures (11 przypadków): pokryte wyroczniami JVM 8 — złożony marker, parametry, XML, proxy, self-invocation, kontener przed kontrolerem, qualifier, dwa profile; częściowo 1 — wartości domyślne adnotacji (testy jednostkowe `internal/annotation`, `@Bean` init/destroy w `testdata/di`); **odłożone 2** — konsument wiadomości (wymaga brokera; listenery są korzeniami z adnotacji) i priorytety konfiguracji (brak modelu źródeł właściwości; `@Value` tylko raportowe).
  - Kontrolowane uruchomienia beanów/endpointów: DI (2 scenariusze), XML, AOP, zdarzenia, tx, JPA, Spring Data, MyBatis, servlety, Spring MVC — jedna wersja (Spring 5.3.6, Tomcat 9.0.45), wnioski nieprzenoszone na inne wersje.
  - Raportowanie: log i `report.md` (punkty wstrzyknięcia, katalog wejść, korzenie cyklu życia i ciał MVC, warunki `@Profile`); **fakty frameworka nie trafiają do v2 JSON** — znana luka, nie kryterium tej fazy blokujące.
  - Snapshot: scenariusz w `config` (`profile`, `roots`, `active-profiles`), reguły w `algorithms` (`f5-1`).
  - B4 i dokumentacja: tabela kwalifikacji, README, FAQ, PLAN.md (faza 7), SUPPORT.MD zaktualizowane w każdym przyroście.
- **Weryfikacja:** `scripts/mvc-oracle.sh`, `scripts/di-oracle.sh` (JDK 25), `TestMVCOracle`, `TestDIOracle`, `TestDIProfilesOracle`, `TestDIInjectionOracle`, `go test ./...`, `go vet ./...`; petclinic/WebGoat/JPetStore `main` vs gałąź (12 konfiguracji w `f5-2`, 6 w `f5-3`): 0 zmian werdyktów.
- **Pozostałe ograniczenia:** konsumenci wiadomości i priorytety konfiguracji bez wyroczni; kandydaci wstrzyknięć niefiltrowani przez `@Profile`; fallback po nazwie pola, `@Priority`, własne adnotacje qualifierów; typy generyczne i dostęp do pól w ciałach Jacksona; skanowanie komponentów nieczytane (kontroler spoza skanowania jest korzeniem); fakty frameworka poza v2 JSON. Szczegóły: wpisy `FRAMEWORKS:F5` w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.

## Tekst fazy z planu

## F5. Kwalifikacja

**Ocena 2026-11-22:** z 11 przypadków fixtures 5 pokrytych wyroczniami B4, 3 częściowo, 3 bez pokrycia (`testdata/di`, `xmlbeans`, `aop`, `tx`, `jpa`, `springdata`, `mybatis`, `servletdispatch`); brakuje kontrolera za `DispatcherServlet`, konsumenta wiadomości, qualifier/`@Primary`, dwóch scenariuszy profili i priorytetów konfiguracji; reguły korzeni/profili nie mają wersji w snapshotcie. Proponowane przyrosty `f5-1`…`f5-3` — zob. IMPLEMENTATION_STATUS. **`f5-1` DONE (2026-11-22):** `profile.rules` i `profile.builtin` w snapshotcie, test wymusza podbicie wersji. **`f5-2` DONE (2026-11-22):** wyrocznia Spring MVC (Tomcat → `DispatcherServlet` → kontroler) — 14 z 26 ciał nieosiągalnych, naprawione (`roots.web_bodies`, korzenie MVC); „kontener przed kontrolerem” pokryty. **`f5-3` DONE (2026-11-22):** qualifier/`@Primary` i dwa scenariusze profili w wyroczni DI — naprawione wyrażenia `@Profile` i dopasowanie qualifierów; pokryte 8 z 11 przypadków, odłożone: konsumenci wiadomości, priorytety konfiguracji.

- Fixtures: adnotacje z wartościami i defaults, złożony marker, parametry, XML, priorytety configu, qualifier, dwa profile, proxy, self-invocation, kontener przed kontrolerem i konsument wiadomości.
- Porównać w kontrolowanym uruchomieniu rzeczywiste beany/endpointy z modelem na kilku scenariuszach; nie przenosić wniosków między wersjami bez sprawdzenia.
- Raportować liczbę reguł, wiązań, nieznanych warunków i nierozstrzygniętych wejść. Zapis scenariusza i modeli w snapshotcie jest obowiązkowy.
- Wykonać B4 i migrację dokumentacji. Brak wsparcia fragmentu konfiguracji ogranicza wniosek negatywny, nie zamienia go w „bean nie istnieje”.
