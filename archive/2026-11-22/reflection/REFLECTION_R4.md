# REFLECTION:R4 — zakończone (2026-11-22)

## Zapis ukończenia

- **Data:** `r4-1` 2026-11-21 (PR #124), `r4-2` 2026-11-22.
- **Co dostarczono:** `Class.forName` z literałem → krawędź do `<clinit>` (`r4-1`); lookupy odrzucane przez JVM (`getMethod` niepublicznej, `getDeclaredMethod` odziedziczonej, `getConstructor` niepublicznego) → luka zamiast krawędzi (`r4-2`); brakujące przypadki wyroczni JVM (samo wyszukanie, niedostępna metoda, `initialize` true/false). Integracja z `internal/reach` (B0 `b0-1`…`b0-5`), wersje w snapshotach (`b3-4`) i wyrocznia (`b4-1`) były już spełnione przed `r4-1`.
- **Pliki:** `internal/reflectinvoke/{forname,edges,resolve,interproc,newinstance_edges,newinstance_resolve}.go`, `internal/hierarchy/resolve.go` (`ResolvePublicByParams`), `testdata/reflection`, `cmd/ariadne/reflection_oracle_test.go`, `scripts/reflection-oracle.sh`.
- **Weryfikacja:** `scripts/reflection-oracle.sh` (JDK 25), `go test ./...`, `go vet ./...`; petclinic/WebGoat/JPetStore `main` vs gałąź w 12 konfiguracjach (`-roots profile`/`all` × CHA/RTA) — wyniki v1 bajtowo identyczne w obu przyrostach.
- **Pozostałe ograniczenia:** `setAccessible` nieśledzone (jawna nadaproksymacja), sprawdzanie dostępu zależne od wywołującego, `ClassLoader#loadClass`/`Lookup#findClass`/`ensureInitialized`, nazwy nieliteralne. Szczegóły: wpisy `REFLECTION:R4` w `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`.

## Tekst fazy z planu

## R4. Kwalifikacja i integracja

**Ocena 2026-11-21:** integracja z `internal/reach` (B0 `b0-1`…`b0-5`), wersje w snapshotach (`b3-4`) i wyrocznia JVM (`b4-1`) już są; zostaje `Class.forName` → `<clinit>` (inicjalizacja przez refleksję niemodelowana) i przypadki fixtures: `initialize` true/false, samo wyszukanie, niedostępna metoda. Selektywność reguły luk (29/29 `NOT_REACHABLE` oznaczonych na petclinic) nie zależy od R4. Zob. IMPLEMENTATION_STATUS. **`r4-1` DONE (2026-11-21):** `Class.forName` → `<clinit>` z przypadkami true/false/nadklasa/niestały w wyroczni JVM; zostają przypadki fixtures: samo wyszukanie, niedostępna metoda. **`r4-2` DONE (2026-11-22) — R4 DONE:** samo wyszukanie bez wywołania, `getMethod` metody niepublicznej, `getDeclaredMethod` metody odziedziczonej, `getConstructor` konstruktora niepublicznego (trzy ostatnie dawały fałszywe krawędzie — teraz luka zamiast krawędzi, reguła `r4-2`) oraz `getDeclaredMethod` prywatnej bez `setAccessible` jako jawna nadaproksymacja (krawędź zostaje). Zob. IMPLEMENTATION_STATUS.

- Fixtures obejmują inicjalizację true/false, samo wyszukanie, invoke, konstruktor, dziedziczenie, overload, factory, nieznany parametr i niedostępną metodę.
- W `internal/reach` i raporcie przenieść wpływ unresolved sites zgodnie z B0; nie globalnie wyciszać wszystkich CVE ani udawać, że luka jest nieistotna.
- Wersjonować modele i dołączyć ich wejścia do snapshotów CVE F3.
- Wykonać kontrole B4 i udokumentować pokrycie według API/wersji; pomiary jakości przed rozszerzaniem heurystyk.
