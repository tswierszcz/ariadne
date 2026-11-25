# Rekomendacje rozwoju Ariadne — indeks

Status wszystkich nowych obszarów: plan, bez implementacji w ramach przygotowania tych dokumentów.

Punkt wyjścia: [oryginalne rekomendacje](MD). Wspólne zależności, zasady i kolejność: [ROADMAP.MD](ROADMAP.MD). Rozpoznawanie CVE (dawna osobna para dokumentów) jest zamknięte i zarchiwizowane — zob. niżej.

`PLAN.md`'s Phase 9/12/13 (z sesji 2026-10-14, przed powstaniem tego folderu) zawierają wcześniejsze, lżejsze notatki o części tego samego zakresu — ServiceLoader, refleksja, proxy/frameworki, Go, RTA. Ten folder jest ich bardziej rygorystycznym rozwinięciem, nie konkurencyjnym planem; nowa praca nad tymi tematami powinna zaczynać się tutaj, nie w `PLAN.md`.

Do realizacji w nowej sesji użyj [promptu startowego i skilla](START_IMPLEMENTATION.MD).
Postęp zapisuje [IMPLEMENTATION_STATUS.MD](IMPLEMENTATION_STATUS.MD), a faktycznie dostępne
możliwości i ograniczenia opisuje [SUPPORT.MD](../docs/SUPPORT.MD).

Poradnik dla człowieka: [INSTRUCTION FOR USER.MD](../INSTRUCTION%20FOR%20USER.MD).
AI ma pomijać ten poradnik przy starcie i odczytywać go wyłącznie na wyraźną
prośbę użytkownika dotyczącą jego treści.

## Dokumenty według obszaru

**Proces badawczy Recon → Hunt → Validate — PLANOWANE (2026-11-25).**
[Plan wsparcia Ariadne](RESEARCH_HARNESS_IMPLEMENTATION_PLAN.MD) wiąże istniejący
interfejs RI z procesem Tekmerion, określa wspólne kamienie milowe H0–H5 i
akceptację dowodów po adaptacji `cloudflare/security-audit-skill`.

**Podstawy grafu JVM (inicjalizacja, dziedziczenie, wybór klas, multi-release) —
DONE (2026-10-22), wszystkie fazy G0–G5.** Dawne `JVM_GRAPH.MD`/
`JVM_GRAPH_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-10-22/jvm_graph/`](../archive/2026-10-22/jvm_graph/COMPLETION.md)
(`COMPLETION.md` tam ma pełne podsumowanie: co dostarczono, jak zweryfikowano,
jedno wciąż otwarte ograniczenie czekające na JDK). Szczegółowe wpisy per-fazowe
zostają w [IMPLEMENTATION_STATUS.MD](IMPLEMENTATION_STATUS.MD).

**Rozpoznawanie podatności — DONE (2026-11-23), wszystkie fazy F0–F10.** Dawne
`CVE.MD`/`CVE_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-11-23/cve/`](../archive/2026-11-23/cve/COMPLETION.md) (`COMPLETION.md` tam:
co dostarczono w każdej fazie, kwalifikacja F9, decyzja o F8c, jawne ograniczenia). Wiersze
w `docs/SUPPORT.MD` pozostają `Partial` tam, gdzie były.

**Tryb biblioteki — DONE (2026-11-23), wszystkie fazy LB0–LB4** (LB2–LB4 dla swojego zakresu).
Dawne `LIBRARY_MODE.MD`/`LIBRARY_MODE_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-11-23/library_mode/`](../archive/2026-11-23/library_mode/COMPLETION.md). Na przyszłość:
tryb `rerun` dla `-alt-versions`, inne strategie wyboru wersji.

**Precyzja grafu — DONE (2026-11-23), wszystkie fazy P0–P4** (P3 dla swojego zakresu). Dawne
`GRAPH_PRECISION.MD`/`GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-11-23/graph_precision/`](../archive/2026-11-23/graph_precision/COMPLETION.md). Domyślny
tryb grafu pozostaje CHA.

**Wiązanie dynamiczne — DONE (2026-11-23), wszystkie fazy L0–L4** (L2–L4 dla swojego zakresu, L3
niebudowane po pomiarze). Dawne `DYNAMIC_LINKAGE.MD`/`DYNAMIC_LINKAGE_IMPLEMENTATION_PLAN.MD` są
zarchiwizowane w [`archive/2026-11-23/dynamic_linkage/`](../archive/2026-11-23/dynamic_linkage/COMPLETION.md).

**Frameworki — DONE (2026-11-23), wszystkie fazy F0–F5** (F2–F5 dla swojego zakresu). Dawne
`FRAMEWORKS.MD`/`FRAMEWORKS_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-11-23/frameworks/`](../archive/2026-11-23/frameworks/COMPLETION.md) (`COMPLETION.md`
tam: fazy, wyrocznie JVM, jawne ograniczenia). Wiersze w `docs/SUPPORT.MD` bez zmiany statusu.

**Refleksja — DONE (2026-11-23), wszystkie fazy R0–R4** (R3 dla swojego zakresu). Dawne
`REFLECTION.MD`/`REFLECTION_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-11-23/reflection/`](../archive/2026-11-23/reflection/COMPLETION.md). Wiersz w
`docs/SUPPORT.MD` pozostaje `Partial`.

**ServiceLoader i moduły JPMS — DONE (2026-11-05), wszystkie fazy S0–S4.** Dawne
`SERVICE_LOADING.MD`/`SERVICE_LOADING_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-11-05/service_loading/`](../archive/2026-11-05/service_loading/COMPLETION.md)
(`COMPLETION.md` tam: co dostarczono, wyrocznia JVM, pomiar historyczny, jawne
ograniczenia). Wiersz w `docs/SUPPORT.MD` pozostaje `Partial`.

**Deserializacja i kandydaci łańcuchów gadgetów — DONE (2026-11-07), wszystkie fazy
Z0–Z4.** Dawne `DESERIALIZATION.MD`/`DESERIALIZATION_IMPLEMENTATION_PLAN.MD` są
zarchiwizowane w
[`archive/2026-11-07/deserialization/`](../archive/2026-11-07/deserialization/COMPLETION.md)
(`COMPLETION.md` tam: co dostarczono, wyrocznia JVM, zestaw CVE, aplikacje kontrolne,
jawne ograniczenia). Wiersz w `docs/SUPPORT.MD` pozostaje `Partial`.

**Callbacki i wykonywanie asynchroniczne — DONE (2026-11-20), wszystkie fazy A0–A4.** Dawne
`ASYNC_CALLBACKS.MD`/`ASYNC_CALLBACKS_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-11-20/async_callbacks/`](../archive/2026-11-20/async_callbacks/COMPLETION.md)
(`COMPLETION.md` tam: co dostarczono, wyrocznie JVM, kwalifikacja
`testdata/async-qualify/RESULTS.MD`, jawne ograniczenia). Wiersz w `docs/SUPPORT.MD` pozostaje
`Partial`.

**Przepływ wartości i danych — DONE (2026-11-07), wszystkie fazy D0–D5.** Dawne
`DATA_FLOW.MD`/`DATA_FLOW_IMPLEMENTATION_PLAN.MD` są zarchiwizowane w
[`archive/2026-11-07/data_flow/`](../archive/2026-11-07/data_flow/COMPLETION.md)
(`COMPLETION.md` tam: co dostarczono, wyrocznia JVM taint, aplikacje kontrolne, jawne
ograniczenia; kwalifikacja: `docs/TAINT_QUALIFICATION.MD`). Wiersz w `docs/SUPPORT.MD`
pozostaje `Partial`.

| Obszar | Plan funkcjonalny | Plan implementacyjny |
|---|---|---|
| Interfejs badawczy dla Tekmerion — kontrakt draft.1, CLI planowane | [Specyfikacja, schemat i przykłady](../docs/research-interface/README.md) | [RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD](RESEARCH_INTERFACE_IMPLEMENTATION_PLAN.MD) |
| Obserwacja runtime’u, dynamiczne klasy i JNI | [RUNTIME_NATIVE.MD](RUNTIME_NATIVE.MD) | [RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD](RUNTIME_NATIVE_IMPLEMENTATION_PLAN.MD) |
| Inne języki: pilotaż integracji Go | [ECOSYSTEMS.MD](ECOSYSTEMS.MD) | [ECOSYSTEMS_IMPLEMENTATION_PLAN.MD](ECOSYSTEMS_IMPLEMENTATION_PLAN.MD) |

## Pokrycie pierwotnych rekomendacji

| Punkt dokumentu źródłowego | Dokument odpowiedzialny |
|---|---|
| Inicjalizacja klas `<clinit>` | JVM_GRAPH — **DONE, zarchiwizowane** (G3) |
| Wywołania odziedziczonych metod | JVM_GRAPH — **DONE, zarchiwizowane** (G2); wspólna realizacja z CVE F4 nie została naruszona (F4 samo nieodnotowane) |
| ServiceLoader i JPMS | SERVICE_LOADING — **DONE, zarchiwizowane** (S0–S4) |
| Refleksja przez literały, zmienne i konfigurację | REFLECTION — **DONE, zarchiwizowane** (R0–R4); wartości dostarcza DATA_FLOW, konfigurację FRAMEWORKS (**DONE, zarchiwizowane**) |
| Proxy, AOP i dependency injection | FRAMEWORKS — **DONE, zarchiwizowane** (F0–F5) |
| Callbacki i kod asynchroniczny | ASYNC_CALLBACKS — **DONE, zarchiwizowane** (A0–A4) |
| MethodHandle i pozostałe invokedynamic | DYNAMIC_LINKAGE — **DONE, zarchiwizowane** (L0–L4; L3 niebudowane po pomiarze) |
| Deserializacja | DESERIALIZATION — **DONE, zarchiwizowane** (Z0–Z4); dopływ danych z DATA_FLOW D4 |
| Konfiguracja i wejścia frameworka | FRAMEWORKS — **DONE, zarchiwizowane** (F0–F5) |
| Dynamiczne ładowanie klas i JNI | RUNTIME_NATIVE; statyczne nazwy klas obsługuje również REFLECTION |
| RTA i analiza możliwych obiektów | GRAPH_PRECISION — **DONE, zarchiwizowane** (P0–P4) |
| Taint analysis | DATA_FLOW — **DONE, zarchiwizowane** (D0–D5); pilotaż `-taint` opt-in, kwalifikacja w `docs/TAINT_QUALIFICATION.MD` |
| Obserwacja działającej aplikacji | RUNTIME_NATIVE |
| Rozpoznawanie samych podatności — pięć punktów | CVE — **DONE, zarchiwizowane** (F0–F10) |
| Inne języki, w szczególności Go | ECOSYSTEMS |
| Pomiar skuteczności i kolejność realizacji | ROADMAP i kryteria poszczególnych faz |

Multi-release JAR-y i stabilny wybór klas (JVM_GRAPH, teraz zarchiwizowane jako DONE) były uzupełnieniem niezbędnym do wiarygodnego grafu — **nie są deklaracją, że refleksja, callbacki czy JNI zostały rozwiązane**; `archive/2026-10-22/jvm_graph/COMPLETION.md` potwierdza explicite, że te wiersze `docs/SUPPORT.MD` wciąż są `Not implemented`.

## Jak korzystać z planów

1. Zacząć od ROADMAP i wspólnych kontraktów, aby nie budować odrębnych formatów dowodów, klas i metod dla każdego mechanizmu.
2. Dla obszaru przeczytać dokument funkcjonalny, następnie jego fazy i wskazane zależności.
3. Realizować pierwszy zakres danego mechanizmu, a rozszerzenia eksperymentalne dopuścić dopiero po pomiarze.
4. Po wdrożeniu aktualizować stan konkretnej fazy i zapisywać dowody weryfikacji. Samo utworzenie dokumentu nie oznacza ukończenia fazy.

Plany obejmują dokumentację wymagań i prac. Proponowane flagi, pakiety, stany i formaty nie są obecnie dostępnymi funkcjami. Nie podano terminów ani procentu gwarantowanego pokrycia.
