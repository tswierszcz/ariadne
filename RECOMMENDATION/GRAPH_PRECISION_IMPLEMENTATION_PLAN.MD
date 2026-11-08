# Precyzja grafu — fazy P0–P4

Status: P0 DONE (2026-11-07, zarchiwizowane: [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md); strategia, ustawienia grafu, `-graph-out`/`-graph-compare`, wyrocznia znanych celów, polityka nieznanych alokacji); P1 DONE (2026-11-08, zarchiwizowane: [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md); katalog `internal/alloc` `p1-1`, luki, wyrocznia JVM; tylko raport); P2 DONE (2026-11-08, zarchiwizowane: [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md); opt-in `-callgraph-mode rta`, punkt stały `p2-1`, fallback nieznanych alokacji, wyrocznia JVM `testdata/rta`); P3–P4 nierozpoczęte; zob. IMPLEMENTATION_STATUS. Wymagania: [GRAPH_PRECISION.MD](GRAPH_PRECISION.MD). Wspólna kolejność: [ROADMAP.MD](ROADMAP.MD).

## P0. Uczciwy baseline i tryby porównania — DONE (zarchiwizowane 2026-11-08)

Strategia dispatch oddzielona od budowy grafu (`-callgraph-mode cha`), ustawienia grafu w
metadanych, `-graph-out`/`-graph-compare`, wyrocznia znanych celów na prawdziwym JVM, polityka
nieznanych alokacji dla RTA; zobacz
[archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md)
po pełny zapis i `docs/SUPPORT.MD` po bieżący opis zachowania.

## P1. Katalog typów tworzonych jawnie i pośrednio — DONE (zarchiwizowane 2026-11-08)

Katalog `internal/alloc` (`p1-1`, tylko raport): klasy tworzone z pochodzeniem i miejscami,
luki dla niemodelowanych mechanizmów, wyrocznia JVM `testdata/alloc`; zobacz
[archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md)
po pełny zapis i `docs/SUPPORT.MD` po bieżący opis zachowania.

## P2. RTA i punkt stały — DONE (zarchiwizowane 2026-11-08)

Opt-in `-callgraph-mode rta`: punkt stały na katalogu P1 zawęża krawędzie CHA do klas
tworzonych przez osiągalny kod, z fallbackiem nieznanych alokacji i jawnymi lukami; wyrocznia
JVM `testdata/rta`; zobacz
[archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md)
po pełny zapis i `docs/SUPPORT.MD` po bieżący opis zachowania.

## P3. Ograniczona analiza points-to

Zależności: DATA_FLOW D3; P2 jest przydatnym baseline, nie koniecznym fundamentem każdej implementacji points-to.

- Rozszerzyć wspólną domenę obiektów, nie budować drugiego interpreter stosu. Klucz alokacji i przyjęta czułość kontekstu muszą być jawne.
- Modelować przepływ przez parametry, wyniki, pola, tablice i interfejsy w określonym zakresie. Nieznane zewnętrzne metody mają konserwatywną politykę.
- Dla konkretnego receivera wyznaczać możliwe cele i aktualizować graf/streszczenia do kontrolowanego punktu stałego.
- Rozdzielić obiekty, gdy istnieje wystarczający dowód. Brak aliasingu nie wynika z różnicy nazw zmiennych.
- Wprowadzić limity pamięci/czasu, poszerzanie i flagi niekompletności. Fallback nie może podawać ograniczonej listy jako pełnej.

Gotowe: dwa receivery, alias pól, tablica i fabryka dają uzasadnione różnice względem RTA. Metryki wskazują koszt oraz jakość, zanim zwiększy się czułość analizy.

## P4. Benchmark i kwalifikacja

- Osobno mierzyć błędne cele, pominięte oczekiwane cele, nieznane miejsca, zmiany CVE, czas i pamięć na fixtures i reprezentatywnych aplikacjach.
- Jeśli używany jest zewnętrzny silnik referencyjny, przypiąć wersję, zakres i konfigurację; jego wynik także nie jest automatycznie prawdą.
- Porównanie musi uwzględniać wersję modeli frameworka/refleksji i reguły JDK skip. Ulepszenie potwierdzać ręcznie opisanymi przypadkami, nie opinią LLM o krótszej ścieżce.
- Wykonać B4, snapshoty i dokumentację migracji. Zmiana domyślnej strategii jest osobną decyzją po pomiarze.

Nie podaje się z góry gwarantowanej redukcji fałszywych alarmów ani kosztu analizy całego classpath.
