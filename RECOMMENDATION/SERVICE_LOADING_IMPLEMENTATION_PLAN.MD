# ServiceLoader i JPMS — fazy S0–S4

Status: plan. Wymagania: [SERVICE_LOADING.MD](SERVICE_LOADING.MD). Wspólna kolejność: [ROADMAP.MD](ROADMAP.MD).

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

## S2. Materializacja i krawędzie dispatch

Zależności: S1, JVM G3, rejestr B3.

- Dodać cele konstruktora/dopuszczalnej fabryki w odpowiednim miejscu użycia oraz związane inicjalizacje.
- Powiązać obiekt dostawcy z alokacją modelowaną. Jego wywołania metod rozwiązywać zgodnie z typem i deskryptorem.
- Typy przekazać do późniejszego RTA; nie tworzyć oddzielnej prywatnej listy instancji tylko dla ServiceLoader.
- Obsłużyć część iterator API i `Provider.get` z jawną listą wspieranych operacji; nieobsługiwany stream terminal zgłasza lukę.
- W raportach rekonstruować krok modelowany i źródło rejestracji. Brak implementacji daje lukę, nie brak podatności.

Gotowe: pozytywna ścieżka do `decode`, ujemny przypadek deklaracji bez użycia, brak celu i powtórzony wpis mają oczekiwane wyniki. To pierwszy zakres wydania usług classpath.

## S3. JPMS i warstwy

Zależności: S2, JVM G4; zaawansowane wartości mogą korzystać z DATA_FLOW D3.

- Rozszerzyć `internal/classfile` o dane atrybutu `Module`, `uses` i `provides`. Zachować deskryptor modułu poza zwykłym indeksem klas.
- Utworzyć jawny model module path i czytelności w analizowanym scenariuszu. Brak scenariusza nie oznacza poprawnego odtworzenia pełnej konfiguracji runtime’u.
- Rozwiązywać właściwe reguły dostawcy, m.in. metodę `provider()` dla odpowiednich rodzajów modułów; nie stosować jej automatycznie do każdego wpisu classpath.
- Dodać ograniczoną deklarację warstw i rodziców. Dynamicznie budowane nieznane warstwy oznaczać jako lukę do uzupełnienia śladem runtime’u.
- Zachować wszystkie dane w snapshotcie CVE F3; uwzględnić wariant modułu w kluczu modelu.

Gotowe: classpath, named module, automatic module i niedostępny provider są rozróżnione. Nieznana warstwa nie daje pewnego negatywnego wyniku.

## S4. Regresja i rozszerzenia API

- Dodać fixtures dla usług z konstruktorem, fabryką, Provider, duplikatem, dwoma loaderami i modułami; ich regeneracja wymaga odpowiedniego JDK, zwykłe testy czytają gotowe artefakty.
- Rozszerzać stream API dopiero z modelem operacji materializujących i terminalnych, współdzielonym z ASYNC/operacjami callbacków tam, gdzie ma zastosowanie.
- Zmierzyć wpływ na graf i wyniki CVE na stałym wejściu. Zastosować kontrole B4 i opisać wspierane sygnatury.

Poza MVP: dowolne dynamiczne module layers i pełna semantyka wszystkich niestandardowych loaderów. Są to jawne braki, nie powód do rejestrowania wszystkich dostawców jako wykonanych.
