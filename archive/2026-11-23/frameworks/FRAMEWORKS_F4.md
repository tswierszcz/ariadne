# FRAMEWORKS:F4 — DONE for this scope (2026-11-23)

## Zapis ukończenia

- **Decyzja użytkownika (2026-11-23)** po ocenie (PR #137): zamknąć.
- **Zrobione:** zdarzenia aplikacyjne i `@Scheduled`/`@Schedules` (`spring:events` `ev1`, wyrocznia `testdata/events`); profile servlet/jakarta/micronaut/quarkus/stripes z aplikacjami kontrolnymi; dispatch kontenera servletów; Spring MVC (`f5-2`); wersjonowana lista wejść (`profile.rules`/`profile.builtin`, `f5-1`).
- **Odłożone:** model listenerów wiadomości (dziś korzenie z adnotacji), programowalne rejestracje i SpEL, osobny tryb klasyfikacji `reach` dla modelowanych wejść — brak zmierzonego przypadku.
- **Weryfikacja:** wpisy FRAMEWORKS:F4/F5 w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.

## Tekst fazy z planu

## F4. Rozszerzanie wejść i frameworków

**Ocena 2026-11-23:** zdarzenia i `@Scheduled` (`ev1`), profile innych frameworków z aplikacjami kontrolnymi, wersjonowana lista wejść (`f5-1`); odłożone: model listenerów wiadomości, SpEL/programowalne rejestracje, osobny tryb klasyfikacji. Rekomendacja: zamknąć jako DONE for this scope. Zob. IMPLEMENTATION_STATUS.

- Dodać osobne modele schedulerów i event/message listeners, współdzieląc mechanikę wywołania callbacków z ASYNC.
- Kolejne Servlet/Jakarta/Micronaut/Quarkus modele wdrażać na podstawie ich rzeczywistych artefaktów i wersji; sam wygenerowany bytecode nie dowodzi, że każda ścieżka konfiguracji została odtworzona.
- Programowalne rejestracje, dynamiczne profile, SpEL i inne złożone wyrażenia obejmować dopiero w jawnym zakresie; nie wykonywać aplikacji podczas analizy statycznej.
- Opracować zmianę `reach/bfs.go`/klasyfikacji dla modelowanych wejść jako osobny tryb, z zachowaniem wyjaśnienia granic frameworka i eksportu legacy.

Gotowe: lista wspieranych wejść jest wersjonowana; nowe wejścia nie zmieniają niejawnie znaczenia wszystkich istniejących wyników.
