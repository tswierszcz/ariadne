# Plan CI floty — WYCOFANY (2026-11-25)

Plan: [FLEET_CI_PLAN.MD](FLEET_CI_PLAN.MD).

**Decyzja użytkownika z 2026-11-25:** uruchamianie ariadne w CI dla floty artefaktów (biblioteki i aplikacje) oraz raport zbiorczy nie będą częścią Ariadne. Będą robione z poziomu Tekmerion.

- Plan był propozycją czekającą na decyzję w połowie grudnia; 4 otwarte decyzje zostają nierozstrzygnięte.
- Wycofanie dotyczy tylko dokumentacji planu. To, co zostało zrobione, zostaje bez zmian: prototyp raportu zbiorczego `cmd/ariadne-fleet` (z testem `TestFleetReport`) i jego syntetyczne dane `testdata/fleet`. Dalszy rozwój po stronie Ariadne nie jest planowany.
- Ariadne dostarcza dalej to, czego potrzebuje taki orkiestrator: wynik v2 (`results.json`) dla każdego artefaktu, tryb biblioteki, snapshoty i interfejs badawczy (`research index`/`query`).
