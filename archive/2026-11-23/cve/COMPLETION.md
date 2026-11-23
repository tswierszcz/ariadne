# CVE (rozpoznawanie podatności) — obszar DONE (2026-11-23)

Wszystkie fazy F0–F10 rozstrzygnięte; plan i opis obszaru przeniesione tutaj:
[CVE.MD](CVE.MD), [CVE_IMPLEMENTATION_PLAN.MD](CVE_IMPLEMENTATION_PLAN.MD). Wcześniejsze
częściowe archiwa: [F10](../../2026-11-13/cve/CVE_F10.md),
[redukcja `UNKNOWN`](../../2026-11-16/cve/CVE_F10_UNKNOWN_REDUCTION.md), [F8](CVE_F8.md).

## Co dostarczono

| Faza | Stan | Najważniejsze |
|---|---|---|
| F0 | DONE (2026-11-05) | kontrakt wyników `docs/RESULTS_CONTRACT.MD`, zestaw kontrolny `testdata/cve-control` (33 przypadki), baseline |
| F1 | DONE (2026-11-05) | instancje komponentów, ustalenia per advisory × instancja, kody powodów, `-json-format v2`, v1 jako eksport |
| F2 | DONE (2026-11-05) | metadane cache, TTL, `-refresh`/`-allow-stale`/`-max-data-age`, wycofane rekordy i aliasy |
| F3 | DONE (2026-11-05) | snapshoty `-snapshot-out`/`-snapshot-in`/`-snapshot-compare`, klucze wyprowadzeń |
| F4 | DONE (2026-11-05) | komparator wersji Maven (wyrocznia Maven 3.9.12), wersjonowane mapowania, selektory z deskryptorem |
| F5 | DONE (2026-11-06) | obie strony diffu poprawki, klasy członkowskie, linia wydań z historii git |
| F6 | DONE (2026-11-06) | komponenty osadzone (`pom.properties`, SBOM CycloneDX, wersje kandydujące) |
| F7 | DONE dla jawnej relokacji (2026-11-06) | relokacja maven-shade i `-embedded-components` |
| F8 | DONE (2026-11-13…23) | identyfikacja runtime'u, kuratorowany katalog OpenJDK (Temurin, 2026), F8c zamknięte bez budowy (decyzja użytkownika) |
| F9 | DONE dla F0–F7 (2026-11-06) i F8 (2026-11-23) | integracja i kwalifikacja; dla F8: testy end-to-end CLI, dwie poprawki determinizmu snapshotu |
| F10 | DONE (2026-11-10…16) | mapowania bez commita poprawki, `release-diff`, redukcja `UNKNOWN` 49/45/8 → 1/0/0 |

## Weryfikacja

Szczegóły i polecenia w odpowiednich wpisach `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`
(rejestr faz); ostatni: „CVE:F9 dla przyrostu F8”.

## Jawne ograniczenia

- Automatyczne odkrywanie relokacji (F7 zadanie 7) — poza zakresem, osobny etap eksperymentalny.
- Runtime: tylko Temurin, katalog od 2026, brak werdyktu metodowego dla CVE JDK (F8c).
- Ręczne mapowania legacy nie są deklarowane jako kompletne przez samą zmianę formatu.
- Wiersze `docs/SUPPORT.MD` pozostają `Partial` tam, gdzie były.
