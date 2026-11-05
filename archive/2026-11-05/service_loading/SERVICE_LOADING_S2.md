# SERVICE_LOADING:S2 — Materializacja i krawędzie dispatch — archiwum

**Data archiwizacji:** 2026-11-05.
**Stan:** DONE dla zadeklarowanego zakresu (kryteria „Gotowe” z planu S2 mają
testy; pozostałe ograniczenia wypisane niżej są jawnymi lukami, nie ukrytą
pracą). Obszar SERVICE_LOADING jako całość pozostaje aktywny: S3 (JPMS) i S4
(regresja i rozszerzenia API) są otwarte.

Oryginalny tekst fazy (sketch sprzed implementacji) jest na końcu tego pliku;
w `RECOMMENDATION/SERVICE_LOADING_IMPLEMENTATION_PLAN.MD` zastąpiono go
wskaźnikiem. Bieżące zachowanie opisują `docs/SUPPORT.MD` (wiersze
"ServiceLoader and JPMS service models" i "Library model iteration
foundation") oraz README; szczegółowe wpisy per-przyrost zostają w
`RECOMMENDATION/IMPLEMENTATION_STATUS.MD`. Ten plik jest zapisem
historycznym, nie źródłem prawdy o bieżącym stanie.

## Commity

1. `30859e5` — krawędzie do konstruktorów providerów i `<clinit>`.
2. `2e3ab2f` — powiązanie konsumpcji (`iterator()`/`next()`/`Provider.get()`)
   z load site i descriptor-exact dispatch providerów przez koordynator B3.
3. `877c788` — provenance przez streszczenia D3 (pola statyczne, `final` pola
   instancji z konstruktora), wyniki helperów `invokestatic`/`invokespecial`
   z podstawieniem zwracanego parametru, liniowy
   `stream().findFirst().get().get()`; skompilowany fixture
   `testdata/serviceloader`.

## Co zostało zrobione

- Materializacja providera dopiero w operacji pobrania (`Iterator.next()`,
  `Provider.get()`), z PC tego miejsca; `load()`, `hasNext()` i stream bez
  terminala nie tworzą instancji.
- Krawędzie do publicznego bezargumentowego konstruktora i istniejącego
  `<clinit>`; jawna luka dla klasy abstrakcyjnej/interfejsu, braku lub
  niepublicznego konstruktora i braku klasy w hierarchii.
- Dalsze wywołania przez dokładny interfejs usługi: krawędź `BasisModel` do
  implementacji z rejestracji po sprawdzeniu assignability (koordynator B3,
  `internal/analysis`), tylko z osiągalnych call sites.
- Tabela "ServiceLoader consumption and dispatch sites" w raporcie.

## Kryteria „Gotowe” i dowody

| Kryterium | Dowód |
|---|---|
| Dodatnia ścieżka do `decode` | `TestCompiledServiceLoaderFixture`, `TestS2_linkedNextReachesProviderConstructorAndCHAServiceMethod` |
| Deklaracja bez użycia nie tworzy instancji | `loadWithoutConsuming()` w fixture |
| Brak celu | brakujący provider w fixture → jawny gap bez krawędzi |
| Powtórzony wpis | `TestFindMaterializations_duplicateRegistrationMaterializedOnce`, `TestParse_duplicatesAreFlaggedNotDropped` |

Weryfikacja końcowa (2026-11-05): `go vet ./...`, `go test ./...`, build;
`bin/ariadne -offline -profile spring` — petclinic 29 load sites, 57
consumption/dispatch (39 linked), 4 model edges, `3/62/26/51/6`, dwa
przebiegi bajtowo identyczne; WebGoat 48 load sites, 115 consumption/dispatch
(90 linked), 75 model edges, `42/44/18/56/0`. Ostatni przyrost nie zmienił
reachability na żadnej z aplikacji.

## Co nie zostało zrobione (jawne ograniczenia)

- Parametry z nieznanych wywołujących, mutowalne/arbitralne pola, wirtualne
  helper returns — `Incomplete`, bez zgadywanych krawędzi.
- Stream pipelines z callbackami (`map`/`filter`/…) — luka; plan przenosi
  rozszerzenia stream API do S4, współdzielone z ASYNC_CALLBACKS.
- Widoczność przez explicit/system/installed loadery — krawędź możliwości z
  luką widoczności.
- „Typy przekazać do RTA” — RTA nie istnieje (GRAPH_PRECISION); bazowy CHA
  dalej dodaje szersze cele.
- JPMS (`provides`/`provider()`) — S3.

---

## Oryginalny tekst fazy

## S2. Materializacja i krawędzie dispatch

Zależności: S1, JVM G3, rejestr B3.

- Dodać cele konstruktora/dopuszczalnej fabryki w odpowiednim miejscu użycia oraz związane inicjalizacje.
- Powiązać obiekt dostawcy z alokacją modelowaną. Jego wywołania metod rozwiązywać zgodnie z typem i deskryptorem.
- Typy przekazać do późniejszego RTA; nie tworzyć oddzielnej prywatnej listy instancji tylko dla ServiceLoader.
- Obsłużyć część iterator API i `Provider.get` z jawną listą wspieranych operacji; nieobsługiwany stream terminal zgłasza lukę.
- W raportach rekonstruować krok modelowany i źródło rejestracji. Brak implementacji daje lukę, nie brak podatności.

Gotowe: pozytywna ścieżka do `decode`, ujemny przypadek deklaracji bez użycia, brak celu i powtórzony wpis mają oczekiwane wyniki. To pierwszy zakres wydania usług classpath.

