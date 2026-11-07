# DATA_FLOW:D4 — Pierwsza analiza taint — archiwum

**Data archiwizacji:** 2026-11-07.
**Stan:** DONE dla zatwierdzonego pilotażu (dane żądania HTTP → dane czytane
przez wejścia deserializacji), flaga `-taint` opt-in, wynik tylko w raporcie.
Jedno kryterium planu spełnione częściowo — patrz „Co nie zostało zrobione”.

Oryginalny tekst fazy D4 (poniżej, sprzed implementacji, z dopiskiem
„Wykonano”) był w `RECOMMENDATION/DATA_FLOW_IMPLEMENTATION_PLAN.MD` (od zamknięcia obszaru w tym katalogu); zastąpiony
tam wskaźnikiem do tego pliku. Bieżący opis zachowania: `docs/SUPPORT.MD`
(wiersz „Data flow and taint”); pełny wpis z weryfikacją:
`RECOMMENDATION/IMPLEMENTATION_STATUS.MD`, „DATA_FLOW:D4 — pierwsza analiza
taint”. Ten plik jest zapisem historycznym, nie źródłem prawdy o stanie.

## Commity

- `c97724f` — Implement DATA_FLOW D4: first taint analysis (request data →
  deserialization entries); zmergowane jako PR #9 (`003e9d3`, 2026-11-07).

## Co zostało zrobione

- `internal/taint` (katalog `taint-1`): źródła Spring MVC (adnotowane i
  niejawne parametry handlerów), JAX-RS `@*Param`, akcesory `ServletRequest`
  (`javax`/`jakarta`); wyszukiwanie wstecz po streszczeniach D3 per miejsce
  wywołania w klasach aplikacji (≤ 8 poziomów, ≤ 2000 węzłów); stany
  `flow_found`/`no_flow_in_model`/`undetermined`/`not_assessed`; ścieżka
  (metoda, wejście, PC wywołania); luki i wejścia spoza katalogu.
- Zabezpieczenia specyficzne dla modelu: tylko polityka Z2 wejścia
  (`effective`/`partial`/`unknown`); zamiana napisów, escaping/encoding i
  sprawdzenia na gałęzi → `unrelated`, nigdy nie zamykają przepływu.
- `internal/flowsum`: `Options.Opaque` (akcesory żądania pozostają źródłami).
- `internal/deserialize/taint.go`: ujścia i mapowanie polityki; CLI `-taint`;
  v2 `deserialization.taint_catalog`/`entries[].taint`/
  `summary.entries_by_taint`; sekcja raportu; snapshot `taint.catalog`.

## Dowody

- `go test ./...`, `go vet ./...` — PASS; testy `internal/taint` na fixture
  `testdata/taint` (`scripts/build-taint-fixture.sh`), testy v2 i raportu.
- `scripts/prepare-taint-demo.sh` na OpenJDK 25: 9/9 wejść w oczekiwanym
  stanie — pięć przypadków kryterium „Gotowe” rozróżnionych.
- WebGoat/petclinic offline: werdykty bez zmian; v2 bez `-taint` jak `main`
  (poza `analysis_id`), z `-taint` różni się tylko polami taint; dwa przebiegi
  bajtowo identyczne; WebGoat `VulnerableComponentsLesson#completed`
  `flow_found` z `@RequestParam payload` (guard `unrelated`); brak
  mierzalnego narzutu czasu.

## Co nie zostało zrobione

- „Rekonstruować ścieżkę danych z parametrami i **krokami transformacji**”:
  ścieżka jest na poziomie metod (parametry i miejsca wywołań), bez kroków
  wewnątrz metody — streszczenia D3 nie zachowują operacji pośrednich.
- „Osobny graf danych”: wykorzystano streszczenia D3 zamiast osobnego grafu;
  jedna rodzina ujść (wejścia deserializacji).
- Przepływy przez pola, wywołujący spoza klas aplikacji, lambdy/referencje
  metod i niemodelowane wywołania → `undetermined`; brak zależności sterowania.
- Wyniki łańcuchów (DESERIALIZATION Z3a) nadal `flow: not_analyzed` —
  powiązanie to reszta DESERIALIZATION:Z3. Pomiar i kwalifikacja — DATA_FLOW:D5.

## Oryginalny tekst fazy

## D4. Pierwsza analiza taint — DONE (2026-11-07; pilotaż: źródła Spring MVC/JAX-RS/Servlet → dane czytane przez wejścia deserializacji; `-taint` opt-in, tylko raport)

Zależności: D3, CVE F4–F5 dla mapowania, FRAMEWORK F1–F2 dla źródeł scenariusza.

- Dodać `internal/taint`: osobny graf danych, fakty źródeł/sinków, zakres argumentów i podstawę zaufania.
- Zdefiniować jeden pilotaż, np. request parameter → argument wybranego API deserializacji, bez deklarowania dowodu dla wszystkich CVE tej biblioteki.
- Wprowadzić reguły propagacji i zabezpieczeń specyficzne dla modelu. Warunkowy check tylko na jednej gałęzi nie oczyszcza danych w całej metodzie.
- Rekonstruować ścieżkę danych z parametrami i krokami transformacji, zachowując poprawne powiązanie wywołania z powrotem.
- W JSON v2 dodać osobną ocenę flow, powody luk i identyfikatory reguł; zachować dotychczasową ocenę LLM jako odrębny typ dowodu.

Gotowe: dane użytkownika docierające do sinka, stała, właściwe zabezpieczenie, zabezpieczenie niezwiązane z podatnością i nieznany fragment mają odrębne oczekiwane wyniki.

Wykonano (2026-11-07): `flow_found`/guard `none`, `no_flow_in_model`, `flow_found`/guard `effective` (polityka Z2), `flow_found`/guard `unrelated` (escaping + sprawdzenie na gałęzi), `undetermined` (interfejs z dwiema implementacjami) — aplikacja kontrolna `scripts/prepare-taint-demo.sh` i fixture `testdata/taint`. Ścieżka ma poziom metod (bez kroków transformacji wewnątrz metody); przepływy przez pola, wywołujący spoza aplikacji i lambdy są lukami. Szczegóły i pomiar: IMPLEMENTATION_STATUS, wpis DATA_FLOW:D4.
