# DYNAMIC_LINKAGE — obszar DONE (2026-11-23)

Plan i opis: [DYNAMIC_LINKAGE.MD](DYNAMIC_LINKAGE.MD), [DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD](DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD); archiwa faz: [L1](../../2026-11-09/dynamic_linkage/DYNAMIC_LINKAGE_L1.md), [L2](DYNAMIC_LINKAGE_L2.md).

| Faza | Stan | Najważniejsze |
|---|---|---|
| L0 | DONE | rodzaj uchwytu bootstrapu (`RefKind`), wartości `KindFunctional`/`KindHandle` odrębne od krawędzi, nieznany bootstrap jako licznik z diagnostyką |
| L1 | DONE (2026-11-09) | lambdy i referencje metod: krawędź u twórcy i materializacja przy wywołaniu SAM (`internal/samcall`, builder `4`) |
| L2 | DONE for this scope | stałe uchwyty, `find*`, uchwyty w polach i wynikach wywołań; dalsza rozbudowa zamknięta po pomiarze |
| L3 | DONE for this scope, niebudowane | kombinatory: 2 z 611 luk; `StringConcatFactory` z typem spoza JDK: 0 (petclinic) / 1 (WebGoat) miejsce |
| L4 | DONE for this scope | wyrocznie `samcall`, `async` (MethodHandle), dowody i wersje w snapshotcie |

**Jawne ograniczenia:** kombinatory MethodHandle, `CONSTANT_Dynamic`, własne bootstrapy, niejawne `toString()` w konkatenacji (zarówno `invokedynamic`, jak i `StringBuilder`) — niemodelowane; wiersze `docs/SUPPORT.MD` bez zmiany statusu.
