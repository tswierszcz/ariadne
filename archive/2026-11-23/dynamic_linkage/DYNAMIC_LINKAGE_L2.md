# DYNAMIC_LINKAGE:L2 — DONE for this scope (2026-11-23)

## Zapis ukończenia

- **Zrobione:** stałe uchwyty i katalog `Lookup.findStatic/findVirtual/findConstructor` na wspólnej analizie wartości (`internal/methodhandle`), uchwyt w tej samej metodzie, w polu statycznym, w polu instancyjnym i jako wynik wywołania statycznego; posiadanie ≠ wywołanie; nierozstrzygnięte miejsca jako luki `methodhandle`.
- **Zamknięcie (2026-11-23, decyzja użytkownika):** dalsza rozbudowa niebudowana — pomiar 2026-11-21: rozwiązane 0 z 10 miejsc `invoke`/`invokeExact` na petclinic i 2 z 46 na WebGoat; `MethodHandle` to 2 z 611 istotnych miejsc luk petclinic (`b0-5`) — pełne L2 nie zmieniłoby żadnego oznaczenia. Reszta to kombinatory (L3) i nazwy z czasu wykonania.
- **Weryfikacja:** wyrocznia `scripts/async-oracle.sh` (MethodHandle, `b4-2`), wpisy DYNAMIC_LINKAGE:L2 w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.
- **Ograniczenia:** adaptacja typów `invoke` poza zgodnością dokładną, pola z nieznanym zapisem, kombinatory — L3 (nieoceniony).

## Tekst fazy z planu

## L2. Podstawowe MethodHandle

**Pomiar 2026-11-21 — dalsza rozbudowa nie podjęta (decyzja użytkownika):** rozwiązane 0 z 10 miejsc
`invoke`/`invokeExact` na petclinic i 2 z 46 na WebGoat (głównie JRuby); pola instancyjne i wyniki
wywołań statycznych są już obsługiwane, reszta to kombinatory (L3), pola z nieznanym zapisem, nazwy
z czasu wykonania. `MethodHandle` to 2 z 611 istotnych miejsc luk petclinic (`b0-5`) — pełne L2/L3 nie
zmieniłoby żadnego oznaczenia. Zob. IMPLEMENTATION_STATUS.

- Obsłużyć stałe uchwyty i mały katalog `Lookup.findStatic/findVirtual/findConstructor`; korzystać ze wspólnej analizy nazw i typów, zamiast duplikować refleksję.
- Dla `invokeExact` sprawdzać znany MethodType, a dla `invoke` wdrożyć zadeklarowany zakres adaptacji. Nierozstrzygnięta zgodność typu jest warunkiem.
- Powiązać wywołanie uchwytu z rzeczywistym receiverem, typami argumentów i inicjalizacją. Sam lookup nie wywołuje ciała celu.
- Zarejestrować modelowane alokacje i nieobsługiwane warianty.

Gotowe: różnica posiadanie/invocation, stały cel, niezgodny typ i nieznany lookup są widoczne. Zakres L0–L2 jest MVP szerszego dynamic linkage.
