# FRAMEWORKS:F3 — DONE for this scope (2026-11-23)

## Zapis ukończenia

- **Decyzja użytkownika (2026-11-23)** po ocenie (PR #137): zamknąć.
- **Zrobione:** tożsamość proxy/target/interceptor (`internal/aop`, wycinek 1); katalog proxy MyBatis/Spring Data/Spring ORM/`@Transactional` (`framework:proxies` `x1`–`x3`, wyrocznie `mybatis`, `springdata`, `jpa`, `tx`); advice Spring AOP (`spring:aop` `a1`/`a2`, wyrocznia `testdata/aop`: CGLIB, self-invocation, final, JDK proxy); fallback Spring Data jako jawna reguła `synthetic_edges`; typy tworzone przez kontener w katalogu P1 z luką `container-proxy`.
- **Odłożone:** ogólny JDK Proxy z wiązaniem handlera i `Method` (wycinek 2) — pomiar 2026-11-09: 0 z 12 relacji `proxy` ma ograniczony zbiór interfejsów.
- **Weryfikacja:** wpisy FRAMEWORKS:F3 w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.

## Tekst fazy z planu

## F3. Proxy i ograniczony model AOP

**Ocena 2026-11-23:** punkty 1–5 zrobione wycinkami (`internal/aop`, `framework:proxies` `x1`–`x3`, `spring:aop` `a1`/`a2`, katalog P1); otwarty tylko ogólny JDK Proxy (wycinek 2, zmierzony 0/12, reguła stopu). Rekomendacja: zamknąć jako DONE for this scope. Zob. IMPLEMENTATION_STATUS.

Zależności: F2, DATA_FLOW D3 dla bardziej złożonych powiązań, REFLECTION R2 i JVM G2 w używanych przypadkach.

- Wprowadzić model tożsamości proxy, targetu oraz interceptora; dla ogólnego JDK Proxy powiązać handler i argument `Method` z właściwym wywołaniem interfejsu.
- Zachować różnice JDK/CGLIB, final/private i self-invocation. Nie tworzyć całego advice chain dla zwykłego bezpośredniego wywołania targetu.
- Dodać ograniczoną semantykę wybranych pointcutów i advice. Nieznane wyrażenia pozostają oznaczone; `around` może warunkować przekazanie dalej.
- Migrować Spring Data bridge do katalogu modeli z deskryptorami, powiązaniem receivera i wersji; zachować oryginalną regułę jako jawny fallback przybliżony.
- Udostępnić katalog typów tworzonych przez DI/proxy dla GRAPH_PRECISION; generowany typ wirtualny musi być oznaczony, nie udawać znalezionej klasy.

Gotowe: proxy/target/self-invocation, dwa handlery, warunkowe advice i generowany typ nie mieszają się w jedną ścieżkę.
