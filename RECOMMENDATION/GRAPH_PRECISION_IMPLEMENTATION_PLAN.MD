# Precyzja grafu — fazy P0–P4

Status: P0 DONE (2026-11-07, zarchiwizowane: [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P0.md); strategia, ustawienia grafu, `-graph-out`/`-graph-compare`, wyrocznia znanych celów, polityka nieznanych alokacji); P1 DONE (2026-11-08, zarchiwizowane: [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P1.md); katalog `internal/alloc` `p1-1`, luki, wyrocznia JVM; tylko raport); P2 DONE (2026-11-08, zarchiwizowane: [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P2.md); opt-in `-callgraph-mode rta`, punkt stały `p2-1`, fallback nieznanych alokacji, wyrocznia JVM `testdata/rta`); P4 DONE (2026-11-08, zarchiwizowane: [archive/2026-11-08/graph_precision/GRAPH_PRECISION_P4.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P4.md); benchmark `scripts/graph-benchmark.sh` + `TestGraphBenchmarkFixtures`, kwalifikacja w `testdata/graphbench/RESULTS.MD`, domyślny tryb pozostaje CHA); P3 nierozpoczęte (nieuzasadnione wynikami P4); zob. IMPLEMENTATION_STATUS. Wymagania: [GRAPH_PRECISION.MD](GRAPH_PRECISION.MD). Wspólna kolejność: [ROADMAP.MD](ROADMAP.MD).

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

## P4. Benchmark i kwalifikacja — DONE (zarchiwizowane 2026-11-08)

Benchmark CHA vs RTA (`scripts/graph-benchmark.sh`, `TestGraphBenchmarkFixtures`), ręczna
kwalifikacja każdej zmiany werdyktu, decyzja: domyślnie CHA; ponowna ocena po modelu dispatchu
kontenera servletów; zobacz
[archive/2026-11-08/graph_precision/GRAPH_PRECISION_P4.md](../archive/2026-11-08/graph_precision/GRAPH_PRECISION_P4.md)
po pełny zapis i `testdata/graphbench/RESULTS.MD` po bieżące wyniki.
