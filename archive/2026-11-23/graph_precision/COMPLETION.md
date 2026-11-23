# GRAPH_PRECISION — obszar DONE (2026-11-23)

Plan i opis: [GRAPH_PRECISION.MD](GRAPH_PRECISION.MD), [GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD](GRAPH_PRECISION_IMPLEMENTATION_PLAN.MD); archiwa faz: [P0](../../2026-11-08/graph_precision/GRAPH_PRECISION_P0.md), [P1](../../2026-11-08/graph_precision/GRAPH_PRECISION_P1.md), [P2](../../2026-11-08/graph_precision/GRAPH_PRECISION_P2.md), [P4](../../2026-11-08/graph_precision/GRAPH_PRECISION_P4.md).

| Faza | Stan | Najważniejsze |
|---|---|---|
| P0 | DONE (2026-11-07) | jawna strategia `-callgraph-mode cha`, `-graph-out`/`-graph-compare`, wyrocznia JVM `testdata/callgraph` |
| P1 | DONE (2026-11-08) | katalog typów tworzonych (`internal/alloc`) z pochodzeniem i lukami |
| P2 | DONE (2026-11-08) | opt-in `-callgraph-mode rta`, punkt stały, fallback nieznanych alokacji z lukami |
| P3 | DONE for this scope (2026-11-23) | eksperymentalne `-callgraph-mode pta` (`p3-1`): zawężanie odbiorników według alokacji; 0 zmian werdyktów, większa czułość niebudowana |
| P4 | DONE (2026-11-08) | benchmark CHA/RTA na trzech aplikacjach (`scripts/graph-benchmark.sh`, `testdata/graphbench`) |

**Decyzja trybu:** domyślnie CHA; RTA i PTA opt-in (RTA usuwa fałszywe ścieżki przy `-roots profile`, PTA nie zmienia werdyktów przy wyższym koszcie).

**Jawne ograniczenia:** brak kontekstu wywołań, pól niefinalnych i zawartości tablic w PTA; wiersz `docs/SUPPORT.MD` „RTA and points-to” pozostaje `Partial`.
