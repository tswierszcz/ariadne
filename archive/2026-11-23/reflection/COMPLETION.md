# REFLECTION — obszar DONE (2026-11-23)

Plan i opis obszaru: [REFLECTION.MD](REFLECTION.MD), [REFLECTION_IMPLEMENTATION_PLAN.MD](REFLECTION_IMPLEMENTATION_PLAN.MD); wcześniejsze archiwum R4: [REFLECTION_R4.md](../../2026-11-22/reflection/REFLECTION_R4.md).

| Faza | Stan | Najważniejsze |
|---|---|---|
| R0+R1 | DONE | `Class.forName`/`getMethod`/`getDeclaredMethod` z literałami → fakty `KindConst`/`KindReflectiveMethod`, `KindArray` z semantyką referencji |
| R2 | DONE | `Method#invoke` i `Constructor#newInstance` → krawędzie `BasisReflection` do celu dyspozycji; luki przy niezgodności liczby argumentów |
| R3 | DONE for this scope (2026-11-23) | wycinek 3 (`r3-3`): `Method` przez parametr rozwiązywany u wywołujących, spis pochodzenia; konfiguracja, `setAccessible`, `Constructor` przez parametr — niebudowane (zerowy zysk w dwóch pomiarach) |
| R4 | DONE (2026-11-22) | `Class.forName` → `<clinit>` (`r4-1`); lookupy odrzucane przez JVM → luka (`r4-2`); wyrocznia JVM `testdata/reflection` |

**Weryfikacja:** wyrocznie `scripts/reflection-oracle.sh`, `scripts/reflectparam-oracle.sh`; pomiary na petclinic/WebGoat/JPetStore — wpisy REFLECTION w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.

**Jawne ograniczenia:** nazwy z czasu wykonania, `setAccessible` nieśledzone (nadaproksymacja), sprawdzanie dostępu zależne od wywołującego, `MethodHandles.Lookup#findClass`/`ensureInitialized`; wiersz `docs/SUPPORT.MD` „Reflection” pozostaje `Partial`.
