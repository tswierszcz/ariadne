# ServiceLoader i JPMS — fazy S0–S4

Status: **DONE (S0–S4), zarchiwizowane 2026-11-05** — zob. [COMPLETION.md](COMPLETION.md). Wymagania: [SERVICE_LOADING.MD](SERVICE_LOADING.MD). Wspólna kolejność: [ROADMAP.MD](../../../RECOMMENDATION/ROADMAP.MD).

## S0. Inwentaryzacja zasobów i kontrakt zdarzeń

Zależności: B0–B1, JVM G1.

- W `internal/input` utworzyć indeks zasobów z tożsamością artefaktu i modułu; nie ograniczać wejścia do listy klas.
- Dodać parser plików usług: komentarze, białe znaki, duplikaty, błędne nazwy i brakująca klasa. Zachować zawartość i hash źródła.
- Zdefiniować `ServiceRegistration`, `ServiceLoadSite`, `ProviderMaterialization`. Rejestracja nie jest krawędzią wykonania.

Gotowe: poprawna lista kandydatów i diagnostyka bez tworzenia fikcyjnych wywołań.

## S1. Argument usługi i lokalne użycie loadera

Zależności: S0, DATA_FLOW D0–D2, JVM G2.

- W rejestrze modeli B3 dodać wybrane sygnatury `load` i wartości `Class` argumentu.
- Powiązać zwrócony loader z usługą, a wybrane wywołania iteratora/Provider z tym loaderem.
- Dla literalnej klasy w MVP użyć wspólnych wartości lokalnych. Zmienna o nieznanej wartości pozostaje nieznana, z jawnym powodem i limitem.
- Odróżnić classpath, system loader i jawnie podany loader. Nie przyjmować bez dowodu, że dowolny loader widzi wszystkie artefakty.

Gotowe: dwa loadery dla różnych usług nie mieszają dostawców; niewykorzystany loader nie uruchamia konstruktora.

## S2. Materializacja i krawędzie dispatch — DONE (2026-11-05)

Zarchiwizowane: [archive/2026-11-05/service_loading/SERVICE_LOADING_S2.md](SERVICE_LOADING_S2.md).

## S3. JPMS i warstwy — DONE (2026-11-05)

Zarchiwizowane: [archive/2026-11-05/service_loading/SERVICE_LOADING_S3.md](SERVICE_LOADING_S3.md).

## S4. Regresja i rozszerzenia API — DONE (2026-11-05)

Zrealizowane: zob. [COMPLETION.md](COMPLETION.md) (wyrocznia JVM, fixtures, pomiar historyczny, lista sygnatur; rozszerzenie stream API świadomie niezrobione).


- Dodać fixtures dla usług z konstruktorem, fabryką, Provider, duplikatem, dwoma loaderami i modułami; ich regeneracja wymaga odpowiedniego JDK, zwykłe testy czytają gotowe artefakty.
- Rozszerzać stream API dopiero z modelem operacji materializujących i terminalnych, współdzielonym z ASYNC/operacjami callbacków tam, gdzie ma zastosowanie.
- Zmierzyć wpływ na graf i wyniki CVE na stałym wejściu. Zastosować kontrole B4 i opisać wspierane sygnatury.

Poza MVP: dowolne dynamiczne module layers i pełna semantyka wszystkich niestandardowych loaderów. Są to jawne braki, nie powód do rejestrowania wszystkich dostawców jako wykonanych.
