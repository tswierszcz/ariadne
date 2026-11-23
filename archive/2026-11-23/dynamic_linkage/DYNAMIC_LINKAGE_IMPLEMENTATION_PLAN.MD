# Dynamiczne wiązanie — fazy L0–L4

Status: L1 DONE (2026-11-09, [zarchiwizowane](../../2026-11-09/dynamic_linkage/DYNAMIC_LINKAGE_L1.md)); L0 i część L2 — zob. IMPLEMENTATION_STATUS; reszta: plan. Wymagania: [DYNAMIC_LINKAGE.MD](DYNAMIC_LINKAGE.MD). Zależności wspólne: [ROADMAP.MD](../../../RECOMMENDATION/ROADMAP.MD).

## L0. Reprezentacja bootstrapów i uchwytów — DONE

Zależności: JVM G0–G3, DATA_FLOW D0–D2, B3.

- Rozszerzyć eksport metadanych z `internal/classfile/bootstrap.go`, `pool*` i `calls.go`: indeks bootstrapu, argumenty, rodzaj uchwytu, pełny typ i offset użycia.
- Zachować dynamiczne stałe, ich zależności i brakujące dane bez wykonywania bootstrapu. Uwzględnić ponowne użycia i cykle przy modelowaniu zależności.
- Wprowadzić model `FunctionalValue`/`HandleValue`, odrębny od gotowej krawędzi wywołania.

Gotowe: rodzaj REF_invokeVirtual/REF_invokeStatic/REF_newInvokeSpecial jest dostępny dla modelu; nieznany bootstrap ma diagnostykę przy konkretnym miejscu.

## L1. Lambdy i method references — DONE (zarchiwizowane 2026-11-09)

Wartość funkcyjna w miejscu utworzenia i krawędzie `sam` od wywołań interfejsu funkcyjnego do
implementacji (`internal/samcall`, builder `4`); zobacz
[archive/2026-11-09/dynamic_linkage/DYNAMIC_LINKAGE_L1.md](../../2026-11-09/dynamic_linkage/DYNAMIC_LINKAGE_L1.md)
po pełny zapis i `docs/SUPPORT.MD` po bieżący opis zachowania.

## L2. Podstawowe MethodHandle — DONE for this scope (2026-11-23; dalsza rozbudowa niebudowana po pomiarze; [zarchiwizowane](DYNAMIC_LINKAGE_L2.md))

## L3. Adaptacje i wybrane bootstrapy — DONE for this scope, niebudowane (2026-11-23)

**Zamknięcie 2026-11-23 po pomiarze:** kombinatory `bindTo`/`asType` — MethodHandle to 2 z 611 istotnych miejsc luk petclinic (`b0-5`), rozwiązane 0/10 i 2/46 miejsc `invoke` (L2). `StringConcatFactory` — miejsca konkatenacji `invokedynamic` z argumentem typu spoza JDK (jedyne, dla których model dałby krawędź `toString`; odbiorcy `java.*` nie są rozwijani): petclinic z zależnościami 0 (3 miejsca konkatenacji, żadne z obiektem), WebGoat z zależnościami **1** (119 miejsc, 4 z obiektem, 1 typ spoza JDK). Biblioteki kompilowane pod Javę 8 konkatenują przez `StringBuilder.append(Object)` — niejawne `toString()` to tam ogólna luka „wywołania z JDK”, nie `invokedynamic`. Nieznane bootstrapy zostają licznikiem `UnresolvedDynamic` z diagnostyką. Model niebudowany.

Zależności: DATA_FLOW D3 dla użycia między metodami.

- Rozszerzać uchwyty o `bindTo`, `asType` i wybrane kombinatory wraz z mapą parametrów. Nie udawać zgodności wszystkich kombinatorów przez pominięcie adaptacji.
- Utworzyć rejestr bootstrap models przypięty do sygnatur i wersji; osobno `StringConcatFactory` z właściwym zakresem konwersji obiektów, opcjonalne inne znane bootstrapy i `CONSTANT_Dynamic`.
- Arbitralnego kodu bootstrapu nie wykonywać. Jeśli jego ciało jest dostępne, możliwa jest ograniczona analiza statyczna; wynik wiązania pozostaje nieznany, dopóki nie ma modelu lub dowodu.
- MutableCallSite i aktualizacje celów modelować konserwatywnie jako zbiór możliwych celów w danym zakresie. Nie zamrażać pierwszego celu jako jedynego bez podstawy.

Gotowe: adaptacja zachowuje mapowanie argumentów, nieobsługiwany kombinator daje unknown, a kolejność modeli nie zmienia wyniku.

## L4. Weryfikacja i raportowanie — DONE for this scope (2026-11-23)

**Zamknięcie 2026-11-23:** dla zbudowanego zakresu (L0–L2) wyrocznie JVM istnieją (`testdata/samcall` — lambdy użyte/nieużyte, referencje wirtualne i konstruktorów; `testdata/async` — stały uchwyt, `b4-2`), dowody krawędzi w ścieżkach (`path_edges`, basis `sam`/`methodhandle`) i wersje w snapshotcie (`samcall.rules`, `methodhandle.edges`, `b3-4`), luki według B0. Fixtures dla `bindTo`, własnego bootstrapu, dynamicznej stałej i konkatenacji nie powstają, bo L3 nie jest budowane.

- Fixtures: lambda nieużyta/użyta, virtual reference, constructor reference, stały handle, `bindTo`, `invokeExact` mismatch, własny bootstrap, dynamiczna stała i konwersja obiektu w konkatenacji.
- Dane runtime’u mogą potwierdzać konkretne wiązanie, ale nie usuwają innych statycznie dopuszczonych celów w odmiennym scenariuszu.
- Dołączyć dowody do ścieżek i snapshotów, określić wpływ luk na wynik negatywny zgodnie z B0.
- Wykonać kontrole B4 i porównanie grafu przed/po z uzasadnieniem zmian. Zwykłe testy Go korzystają z przygotowanych klas, generacja fixtures ma jawne wymaganie JDK.
