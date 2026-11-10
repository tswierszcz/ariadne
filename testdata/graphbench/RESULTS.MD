# GRAPH_PRECISION P4 — call-graph strategy benchmark (CHA vs RTA)

**Rerun 2026-11-10 with the framework proxy catalog (`framework:proxies` `x1`)**: the first table
under "Real applications" is this run. JPetStore `-roots all` CHA 0/10/13/12/0 → 1/9/13/12/0
(CVE-2022-41853 `REACHABLE` through the application's mapper calls; RTA keeps it `FRAMEWORK`, a new
`verdict-changes.tsv` row); petclinic and WebGoat verdicts unchanged. See "Framework proxies".

**Rerun 2026-11-10 with logback's Joran components (`l2-2`) and builder `6`**: the first table
under "Real applications" is this run. petclinic and WebGoat: CVE-2024-12798 `FRAMEWORK` →
`NOT_REACHABLE` in every mode (no evaluator configured), CHA 6/72/12/57/6 → 6/71/13/57/6 and
54/37/13/60/0 → 54/36/14/60/0; JPetStore verdicts unchanged (+9 edges builder `5` had wrongly
dropped). See "logback Joran components".

**Rerun 2026-11-10 with the logback appender model (`l2`, `internal/logbackconfig`) and builder
`5`**: the first table under "Real applications" is this run. Edges: petclinic −288 (CHA), WebGoat
−755, JPetStore −7; **no verdict changed** in any app or mode (`verdict-changes.tsv` unchanged). The
logback paths through unconfigured appenders are gone, but two logback advisories stay `FRAMEWORK`
through CHA over logback's `LifeCycle` calls — see "logback appender model".

**Rerun 2026-11-10 with the log4j 1.2 appender model (`l1`, `internal/log4jconfig`)**: the first
table under "Real applications" is this run. JPetStore CHA 225 706 → 225 536 edges, verdicts
3/9/11/12/0 → 0/10/13/12/0 (the three false positives through log4j's appender dispatch are gone);
RTA verdicts unchanged, so its 6 JPetStore entries in `verdict-changes.tsv` (RTA removing those
CHA paths) are gone too. petclinic and WebGoat (no log4j 1.2) unchanged. See "JPetStore", "log4j
1.2 appender model".

**Rerun 2026-11-10 after seven manual sink mappings (HSQLDB CVE-2022-41853, six log4j 1.2
advisories)**: the first table under "Real applications" is this run. Graphs unchanged; verdicts
change only where those components are present — JPetStore 7 `UNKNOWN` resolved (3 CHA false
positives through log4j's appender dispatch, which RTA removes), WebGoat CVE-2022-41853 `UNKNOWN`
→ `REACHABLE` (a true positive: a SQL injection lesson runs request SQL on HSQLDB 2.5.1). See
"JPetStore", "HSQLDB and log4j 1.2 mapped".

**Rerun 2026-11-09 with a third application, JPetStore 6.0.2 (`scripts/prepare-jpetstore.sh`)**: a
WAR (`-app` is the WAR itself, so `web.xml` is read) with Stripes, Spring 4.3 XML configuration,
MyBatis and HSQLDB. The first table under "Real applications" is this run (petclinic and WebGoat
rerun with it: same graphs and verdicts as the `j1` run, times within its spread). No model of
the last increments changes a JPetStore verdict; the two scenarios of roots differ by 8 advisories,
and RTA changes 2 — see "JPetStore" below. **Then fixed:** the web.xml lifecycle methods are
roots of `-roots all` too; the JPetStore `all` rows of the table are after the fix (before:
CHA 225 288 edges, 0/1/15/19/0; RTA 0/0/16/19/0), petclinic and WebGoat unchanged (no web.xml).

**Rerun 2026-11-09 after the JDBC objects (`jdbc:objects` `j1`, `internal/jdbcdispatch`)**: a
virtual/interface call on a `java.sql` connection, statement, result set, metadata object, LOB or
`javax.sql.DataSource` reaches the bodies of the implementations reachable code creates — a
pool's proxies, a driver's classes (README, "JDBC objects"; measured on a runtime trace in
[TRACE_GAPS.MD](TRACE_GAPS.MD): all 22 JDBC relations the graph missed are edges now). The first
table under "Real applications" is this rerun. **No verdict or path changed** in any mode or
roots scenario (`-graph-compare` against `fac65dc`: petclinic +11 342 edges, WebGoat +4 882 /
+4 883, none removed; `verdict-changes.tsv` unchanged). The "Model edges" column counts only the
registry models' edges, not the creator-gated JDBC or servlet-object edges (the log lines
`JDBC objects …`: petclinic 12 560 added, WebGoat 4 589). On petclinic that column drops 5 326 → 4 108: the 1 218
edges the service-provider model used to add from JDBC calls into H2's own classes are now added
first by the JDBC gate (12 560 − 1 218 = the 11 342 net); on WebGoat it grows 1 927 → 2 219 /
1 932 → 2 225 (library-model call sites reachable through the new edges). **Times:** the machine was slower for
this whole run (petclinic 12.5–14.1 s against 9.7–10.4 s for `u1`); measured back to back on
WebGoat (`-roots profile`, CHA) the analyzer without and with the model takes 19.9–20.4 s and
20.4 s — within the spread; the model itself logs 150–180 ms.

**Rerun 2026-11-09 after the Undertow feature model (`internal/bootfeatures`, `u1`)**: in a
Spring Boot application with embedded Undertow, the classes of Undertow features the packaged
configuration and the application code leave off (HTTP/2, AJP, PROXY protocol, Undertow's own
authentication, its HTTP client, write timeout; TLS here stays on — its switch is an environment
placeholder) are not CHA/RTA targets (README, "Undertow features a Spring Boot application leaves
off"). The first table under "Real applications" is this rerun. **WebGoat: 4 Undertow advisories
`REACHABLE` → `NOT_REACHABLE`**, and one more RTA-versus-CHA change (CVE-2023-5685); petclinic
(Tomcat): identical graph. See "Undertow features (u1)" below.

**Rerun 2026-11-09 after the servlet request/response objects (`servlet-container:dispatch`
`s2`)**: calls on `javax.servlet` request/response/wrapper/filter-chain objects reach the bodies
of the implementations reachable code creates (README, "Request, response and filter-chain
objects"; measured on a runtime trace in [TRACE_GAPS.MD](TRACE_GAPS.MD)). The first table under
"Real applications" is this rerun; the builder `4` / `s1` table is kept below it. **WebGoat:
15 Undertow advisories `FRAMEWORK` → `REACHABLE`** in every mode and roots scenario, each checked
by hand in "Servlet request/response objects" below (4 plausible, 11 through CHA
over-approximation inside Undertow); petclinic: no verdict change; the RTA-versus-CHA changes
are the same 4 + 2.

**Rerun 2026-11-09 after SAM-call materialization (builder `4`, DYNAMIC_LINKAGE L1,
`internal/samcall`)**: a lambda/method reference run through its functional interface in
another method than the one creating it gets a `sam` edge from the calling site (README "SAM
calls"; measured on a runtime trace in [TRACE_GAPS.MD](TRACE_GAPS.MD)). The first table under
"Real applications" is this rerun; the builder `3` table is kept below it. **No verdict changed**
in any mode or roots scenario, against builder `3` or between RTA and CHA (the same 4 + 2
RTA-versus-CHA changes, `verdict-changes.tsv` unchanged); see "SAM-call materialization" below.

**Rerun 2026-11-09 after CHA's inherited implementations (builder `3`)**: for each concrete
subtype that only inherits the called method, CHA also targets the body JVM method selection
finds for it, including one in a superclass outside the call's type (found by a runtime
trace, [TRACE_GAPS.MD](TRACE_GAPS.MD)). The tables under "Real applications" are this rerun;
the builder `2` figures are kept below them. No verdict changed in any configuration; the
RTA-versus-CHA verdict changes are the same 4 + 2.

**Rerun 2026-11-08 after the servlet container dispatch model** (`servlet-container:dispatch`,
`internal/servletdispatch`, FRAMEWORKS F2): the sections below describe that rerun
(its table is kept as "Builder `2`"); the P4 figures without the model are kept in "P4 baseline" for comparison, and the
section "Servlet container dispatch" explains what changed and why.

Measured 2026-11-08 with `scripts/graph-benchmark.sh --repeat 3` (real applications) and
`go test ./cmd/ariadne -run TestGraphBenchmarkFixtures -v` (fixtures with a real-JVM
oracle). **A measurement, not ground truth:** fewer edges or fewer `REACHABLE`/`FRAMEWORK`
results are not an improvement by themselves; every verdict change below is explained by
hand from the bytecode, not by an LLM's opinion of a shorter path.

| File | Content |
|---|---|
| `results.tsv` | one row per app × roots scenario × mode: wall time and peak RSS (min–max of 3 measured runs), graph size, CHA edges removed / kept only by the unknown-allocation fallback, live classes, active unknown-allocation gaps, created-type catalog size and gaps, model-gap sites (all / reachable), v1 verdict counts `R/F/NR/U/NA`, registered servlets/filters and edges added by B3 models, SAM-call edges added and left out (`sam_edges`, `sam_edges_gated`), and against the CHA run of the same scenario the edges removed/added and the findings whose verdict or path changed |
| `verdict-changes.tsv` | every verdict change against CHA, with the removed edges that explain it (from `-graph-compare`) |
| `versions.txt` | analyzer commit, Go version, and per mode the graph settings and rule versions: builder `4`, skip rule (`java.`, `javax.`, `jdk.`, `sun.` receivers not expanded — the servlet dispatch model is the bounded exception), unknown-allocation policy, catalog `p1-2`, RTA `p2-1`, servlet dispatch `s1`, SAM-call rules `l1-1` |

## Setup

- Inputs: the prepared petclinic (Spring Boot 2.4.5, embedded Tomcat 9.0.45) and WebGoat
  (Spring Boot, Undertow) of `scripts/prepare-petclinic.sh` / `scripts/prepare-webgoat.sh`,
  `-profile spring`, offline on the same OSV cache, `-deps-list` given.
- Each configuration: 3 measured runs that write and read no graph file, plus 1 unmeasured
  run that writes (CHA) or compares against (RTA) the graph — reading a several-hundred-MB
  graph file costs ~1 GB of memory, which must not be charged to the strategy. Every figure
  other than time and memory was identical across the 4 runs (the script checks).
- The builder `4` rerun's `versions.txt` names `651b3b5 (dirty)`: the uncommitted changes were
  this change (SAM-call materialization, its docs and tests), measured as committed.
- `versions.txt` of the P4 run named a dirty commit (`a64edf7`, P4): the uncommitted changes at measurement
  time were the servlet dispatch model and its wiring, docs and tests. One later change (Spring
  servlet/filter beans count only in a Spring Boot application) does not apply to either app
  — neither declares such a bean; the per-finding comparison below was run with the final
  binary and gives the same counts.
- No external reference engine (e.g. SootUp) was used: none is pinned in this repository,
  and its result would not be ground truth either. The ground truth available is the
  real-JVM oracles of the fixtures and hand-checked bytecode for each verdict change.

## Fixtures (real-JVM oracle)

`TestGraphBenchmarkFixtures`: per fixture, the *traced* method bodies (those that record
when they run) the JVM ran (`observed`), how many traced bodies the graph reaches from
`main`, the observed bodies it does not reach (**missed expected targets**) and the reached
traced bodies the JVM did not run (**extra** — an over-approximation, or code this run did
not exercise).

| Fixture | Roots | Mode | Observed | Reached | Missed | Extra |
|---|---|---|---|---|---|---|
| `testdata/callgraph` (overloads, default method, bridges, inherited body, super call) | profile / all | cha | 13 | 13 of 14 | 0 | 0 |
| | profile / all | rta | 13 | 13 of 14 | 0 | 0 |
| `testdata/rta` (3 implementations, factory, dispatch cycle with late allocation, ServiceLoader provider, reflective factory) | profile | cha | 9 | 11 of 11 | 0 | 2 (`Ghost#area` never created, `Hidden#area` created only in uncalled code) |
| | profile | rta | 9 | 8 of 11 | 1 (`Unnamed#area`: created by reflection from a name built at run time, named nowhere — covered by the active `reflection-unresolved` gap) | 0 |
| | all | cha | 9 | 11 of 11 | 0 | 2 (same) |
| | all | rta | 9 | 11 of 11 | 0 | 2 (with `-roots all`, `total(Shape)` is a root: any `Shape` may come in) |
| `testdata/servletdispatch` (a container that creates web.xml-declared servlets/filters by name and dispatches through the real `javax.servlet` API: an `HttpServlet` via its template dispatch, a `GenericServlet`, a filter; an unregistered servlet and filter on the class path) | profile / all | cha | 5 | 5 of 8 | 0 | 0 |
| | profile / all | rta | 5 | 5 of 8 | 0 | 0 |

The test fails if CHA misses anything, if RTA misses anything but the documented case, if
RTA reaches a body CHA does not, or if the graph reaches the unregistered `sd.UnusedServlet` /
`sd.UnusedFilter` (the 3 traced bodies not reached). Without the servlet dispatch model
(analyzer `a64edf7`) none of the 5 observed `servletdispatch` bodies is reached from `main`
in either mode.

## Real applications

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | Model edges | SAM-call edges (left out: creator unreachable) | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 14.7–14.8 | 1357–1379 | 1 027 108 | — | — | 705 of 1096 | 5 | 4 571 | 1 678 (2 361) | 6/71/13/57/6 |
| petclinic | profile | rta | 13.2–14.2 | 1342–1364 | 991 892 | 35 216 of 321 532 (34 854) | 17 451 (5 796) | 661 of 1097 | 5 | 4 571 | 1 678 (2 361) | 6/67/17/57/6 |
| petclinic | all | cha | 12.5–13.0 | 1357–1426 | 1 027 108 | — | — | 705 of 1096 | 5 | 4 571 | 1 678 (2 361) | 6/71/13/57/6 |
| petclinic | all | rta | 13.4–13.6 | 1343–1400 | 1 021 212 | 5 896 of 321 532 (57 892) | 21 864 (9 567) | 703 of 1097 | 5 | 4 571 | 1 678 (2 361) | 6/71/13/57/6 |
| WebGoat | profile | cha | 19.2–20.2 | 1696–1814 | 1 434 047 | — | — | 771 of 1194 | 0 | 2 623 | 1 597 (2 104) | 54/36/14/60/0 |
| WebGoat | profile | rta | 20.1–20.5 | 1624–1721 | 1 351 506 | 82 541 of 569 025 (161 228) | 25 287 (10 581) | 695 of 1201 | 0 | 2 623 | 1 597 (2 104) | 53/34/17/60/0 |
| WebGoat | all | cha | 19.7–20.0 | 1749–1801 | 1 434 052 | — | — | 772 of 1194 | 0 | 2 629 | 1 598 (2 099) | 54/36/14/60/0 |
| WebGoat | all | rta | 20.4–20.8 | 1619–1783 | 1 419 286 | 14 766 of 569 025 (102 240) | 32 262 (13 137) | 774 of 1201 | 0 | 2 629 | 1 598 (2 099) | 54/36/14/60/0 |
| JPetStore | profile | cha | 3.5–3.6 | 295–313 | 225 554 | — | — | 161 of 364 | 1 | 93 | 0 (0) | 0/10/13/12/0 |
| JPetStore | profile | rta | 3.6–3.7 | 299–311 | 212 454 | 13 100 of 59 469 (16 226) | 3 205 (1 263) | 144 of 365 | 1 | 93 | 0 (0) | 0/9/14/12/0 |
| JPetStore | all | cha | 3.6–3.8 | 303–318 | 225 713 | — | — | 167 of 364 | 1 | 401 | 0 (0) | 1/9/13/12/0 |
| JPetStore | all | rta | 3.3–3.3 | 293–305 | 212 792 | 12 921 of 59 469 (16 401) | 3 229 (1 276) | 145 of 365 | 1 | 401 | 0 (0) | 0/9/14/12/0 |

<details><summary>JDBC objects <code>j1</code>, petclinic and WebGoat only (first run, analyzer <code>jdbc-objects</code>)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | Model edges | SAM-call edges (left out: creator unreachable) | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 13.6–14.1 | 1359–1386 | 1 027 168 | — | — | 681 of 1096 | 5 | 4 108 | 1 629 (2 524) | 6/72/12/57/6 |
| petclinic | profile | rta | 13.0–13.6 | 1345–1381 | 991 057 | 36 111 of 322 104 (35 727) | 17 385 (5 845) | 639 of 1097 | 5 | 4 108 | 1 629 (2 524) | 6/68/16/57/6 |
| petclinic | all | cha | 12.5–12.9 | 1331–1374 | 1 027 168 | — | — | 681 of 1096 | 5 | 4 108 | 1 629 (2 524) | 6/72/12/57/6 |
| petclinic | all | rta | 13.7–13.8 | 1323–1386 | 1 021 275 | 5 893 of 322 104 (59 461) | 21 852 (9 620) | 679 of 1097 | 5 | 4 108 | 1 629 (2 524) | 6/72/12/57/6 |
| WebGoat | profile | cha | 19.8–20.0 | 1653–1712 | 1 434 527 | — | — | 742 of 1194 | 0 | 2 219 | 1 557 (2 269) | 53/37/13/61/0 |
| WebGoat | profile | rta | 20.6–20.9 | 1779–1802 | 1 350 616 | 83 911 of 570 065 (162 221) | 25 180 (10 659) | 667 of 1201 | 0 | 2 219 | 1 557 (2 269) | 52/35/16/61/0 |
| WebGoat | all | cha | 19.1–19.8 | 1700–1851 | 1 434 532 | — | — | 743 of 1194 | 0 | 2 225 | 1 558 (2 264) | 53/37/13/61/0 |
| WebGoat | all | rta | 20.5–21.1 | 1680–1720 | 1 419 769 | 14 763 of 570 065 (104 410) | 32 224 (13 214) | 745 of 1201 | 0 | 2 225 | 1 558 (2 264) | 53/37/13/61/0 |

</details>

<details><summary>Undertow feature model <code>u1</code>, before the JDBC objects (analyzer <code>fac65dc</code>)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | Model edges | SAM-call edges (left out: creator unreachable) | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 9.8 | 1356–1372 | 1 015 826 | — | — | 681 of 1096 | 5 | 5 326 | 1 629 (2 525) | 6/72/12/57/6 |
| petclinic | profile | rta | 10.2–10.3 | 1351–1423 | 979 715 | 36 111 of 322 104 (35 727) | 17 379 (5 846) | 639 of 1097 | 5 | 5 326 | 1 629 (2 525) | 6/68/16/57/6 |
| petclinic | all | cha | 9.7–9.8 | 1391–1427 | 1 015 826 | — | — | 681 of 1096 | 5 | 5 326 | 1 629 (2 525) | 6/72/12/57/6 |
| petclinic | all | rta | 10.3–10.4 | 1355–1378 | 1 009 933 | 5 893 of 322 104 (59 461) | 21 851 (9 626) | 679 of 1097 | 5 | 5 326 | 1 629 (2 525) | 6/72/12/57/6 |
| WebGoat | profile | cha | 15.6–15.7 | 1706–1753 | 1 429 644 | — | — | 742 of 1194 | 0 | 1 927 | 1 557 (2 270) | 53/37/13/61/0 |
| WebGoat | profile | rta | 16.1–16.3 | 1699–1797 | 1 345 733 | 83 911 of 570 065 (162 221) | 25 173 (10 660) | 667 of 1201 | 0 | 1 927 | 1 557 (2 270) | 52/35/16/61/0 |
| WebGoat | all | cha | 15.5–16.1 | 1616–1726 | 1 429 650 | — | — | 743 of 1194 | 0 | 1 932 | 1 558 (2 265) | 53/37/13/61/0 |
| WebGoat | all | rta | 16.0–16.2 | 1647–1804 | 1 414 887 | 14 763 of 570 065 (104 410) | 32 223 (13 220) | 745 of 1201 | 0 | 1 932 | 1 558 (2 265) | 53/37/13/61/0 |

Against `s2` without the feature model (`-graph-compare`, both roots scenarios): petclinic
identical (no Undertow); WebGoat 3 215 CHA edges removed, none added. **Times:** the machine was
slower for this whole run than for the `s2` table below (petclinic, which this change does not
touch, 9.7–9.8 s against 5.8–6.4 s); measured back to back on WebGoat (`-roots profile`, CHA)
the analyzer without and with the model takes 14.8–15.2 s and 15.2–15.4 s — within the run-to-run
spread.

</details>

<details><summary>Servlet model <code>s2</code>, before the Undertow feature model (analyzer <code>8298256</code>)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | Model edges | SAM-call edges (left out: creator unreachable) | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 5.8–5.9 | 1363–1392 | 1 015 826 | — | — | 681 of 1096 | 5 | 5 326 | 1 629 (2 525) | 6/72/12/57/6 |
| petclinic | profile | rta | 6.1–6.3 | 1327–1476 | 979 715 | 36 111 of 322 104 (35 727) | 17 379 (5 846) | 639 of 1097 | 5 | 5 326 | 1 629 (2 525) | 6/68/16/57/6 |
| petclinic | all | cha | 6.0–6.4 | 1363–1466 | 1 015 826 | — | — | 681 of 1096 | 5 | 5 326 | 1 629 (2 525) | 6/72/12/57/6 |
| petclinic | all | rta | 6.4–6.5 | 1363–1388 | 1 009 933 | 5 893 of 322 104 (59 461) | 21 851 (9 626) | 679 of 1097 | 5 | 5 326 | 1 629 (2 525) | 6/72/12/57/6 |
| WebGoat | profile | cha | 8.8–9.2 | 1706–1728 | 1 432 859 | — | — | 745 of 1194 | 0 | 1 927 | 1 557 (2 270) | 57/37/9/61/0 |
| WebGoat | profile | rta | 9.2–9.3 | 1684–1794 | 1 350 750 | 82 109 of 573 201 (160 476) | 25 343 (10 571) | 671 of 1201 | 0 | 1 927 | 1 557 (2 270) | 57/35/11/61/0 |
| WebGoat | all | cha | 8.8–8.9 | 1702–1818 | 1 432 865 | — | — | 746 of 1194 | 0 | 1 932 | 1 558 (2 265) | 57/37/9/61/0 |
| WebGoat | all | rta | 9.3–9.4 | 1712–1777 | 1 418 102 | 14 763 of 573 201 (102 750) | 32 228 (13 018) | 748 of 1201 | 0 | 1 932 | 1 558 (2 265) | 57/37/9/61/0 |

Against `s1` (`-graph-compare`, both roots scenarios): +5 870 edges on petclinic (5 866 servlet
object edges — the "Model edges" column counts only the registry models' edges), +3 447 on
WebGoat, nothing removed; petclinic registers one more filter (`ResourceUrlEncodingFilter`).
**Times are not comparable with the tables below:** this machine was idle for this run and
busy with other analyses for the builder `4` run; measured back to back on WebGoat
(`-roots profile`, CHA), `s1` and `s2` both take 8.4–8.6 s — the model costs about 0.1 s
(building its 6 649–10 159 candidates: 50–70 ms).

</details>

<details><summary>Builder <code>4</code>, servlet model <code>s1</code> (before the request/response objects, analyzer <code>0018f2d</code>)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | Model edges | SAM-call edges (left out: creator unreachable) | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 10.4–11.1 | 1336–1364 | 1 009 956 | — | — | 681 of 1096 | 4 | 5 325 | 1 626 (2 530) | 6/72/12/57/6 |
| petclinic | profile | rta | 11.0–11.1 | 1371–1437 | 973 845 | 36 111 of 322 104 (35 732) | 17 373 (5 869) | 639 of 1097 | 4 | 5 325 | 1 626 (2 530) | 6/68/16/57/6 |
| petclinic | all | cha | 10.6–10.7 | 1339–1425 | 1 009 956 | — | — | 681 of 1096 | 4 | 5 325 | 1 626 (2 530) | 6/72/12/57/6 |
| petclinic | all | rta | 11.1–11.2 | 1362–1457 | 1 004 063 | 5 893 of 322 104 (59 466) | 21 847 (9 651) | 679 of 1097 | 4 | 5 325 | 1 626 (2 530) | 6/72/12/57/6 |
| WebGoat | profile | cha | 16.2–16.3 | 1732–1814 | 1 429 412 | — | — | 745 of 1194 | 0 | 1 927 | 1 557 (2 270) | 42/52/9/61/0 |
| WebGoat | profile | rta | 16.9–17.1 | 1759–1770 | 1 347 303 | 82 109 of 573 201 (160 476) | 25 337 (10 573) | 671 of 1201 | 0 | 1 927 | 1 557 (2 270) | 42/50/11/61/0 |
| WebGoat | all | cha | 16.2–16.3 | 1710–1798 | 1 429 418 | — | — | 746 of 1194 | 0 | 1 932 | 1 558 (2 265) | 42/52/9/61/0 |
| WebGoat | all | rta | 17.1–17.2 | 1704–1808 | 1 414 655 | 14 763 of 573 201 (102 750) | 32 228 (13 030) | 748 of 1201 | 0 | 1 932 | 1 558 (2 265) | 42/52/9/61/0 |

Against builder `3`: +1 626 (petclinic) / +1 557–1 558 (WebGoat) `sam` edges and +171 / +77
library-model edges (the B3 models reach more call sites through bodies now reachable at their
real calling site), nothing removed; the same verdict counts. The analysis is **faster** than
with builder `3` (petclinic 14.6 → 10.4–11.1 s, WebGoat 23.2–24.0 → 16.2–16.3 s) at about the
same peak memory: the SAM step itself takes 1.9 s / 2.4 s, and the class-file parser no longer
allocates a 32 KiB buffer per attribute, which made re-reading classes (here, in DATA_FLOW and
in hierarchy building) allocation-bound. Before that fix and the memoization below, the first
benchmark of this change measured WebGoat at 25.0–25.1 s and +200 MB, with the SAM step
taking 12.5 s.

</details>

<details><summary>Builder <code>3</code> (before SAM-call materialization, analyzer <code>651b3b5</code>)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | Model edges | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 14.6–14.8 | 1364–1392 | 1 008 159 | — | — | 679 of 1096 | 4 | 5 154 | 6/72/12/57/6 |
| petclinic | profile | rta | 14.7–15.2 | 1338–1393 | 971 408 | 36 751 of 322 104 (36 445) | 17 275 (5 969) | 635 of 1097 | 4 | 5 154 | 6/68/16/57/6 |
| petclinic | all | cha | 14.5–14.6 | 1361–1376 | 1 008 159 | — | — | 679 of 1096 | 4 | 5 154 | 6/72/12/57/6 |
| petclinic | all | rta | 15.0–15.1 | 1358–1374 | 1 002 249 | 5 910 of 322 104 (60 047) | 21 837 (9 786) | 677 of 1097 | 4 | 5 154 | 6/72/12/57/6 |
| WebGoat | profile | cha | 23.2–24.0 | 1766–1795 | 1 427 778 | — | — | 744 of 1194 | 0 | 1 850 | 42/52/9/61/0 |
| WebGoat | profile | rta | 24.5–25.5 | 1768–1772 | 1 345 079 | 82 699 of 573 201 (161 035) | 25 229 (10 669) | 668 of 1201 | 0 | 1 850 | 42/50/11/61/0 |
| WebGoat | all | cha | 23.6–23.9 | 1762–1770 | 1 427 783 | — | — | 745 of 1194 | 0 | 1 855 | 42/52/9/61/0 |
| WebGoat | all | rta | 24.6–24.9 | 1768–1824 | 1 413 003 | 14 780 of 573 201 (103 372) | 32 212 (13 162) | 747 of 1201 | 0 | 1 855 | 42/52/9/61/0 |

</details>

Against builder `2`: +11 713 (petclinic) / +16 439 (WebGoat) CHA-basis edges, a few more
reachable gap sites and model edges (code reached through the newly targeted bodies), the same
verdict counts. Computing the dispatch targets once per owner/name/descriptor instead of per
call site (`callgraph.BuildWith`) keeps the extra resolution from costing time: without it,
WebGoat's analysis took 47–48 s.

<details><summary>Builder <code>2</code> (before the inherited-implementation fix, analyzer <code>1f7d7f0</code>)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | Model edges | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 17.3–18.2 | 1320–1435 | 996 446 | — | — | 678 of 1096 | 4 | 5 127 | 6/72/12/57/6 |
| petclinic | profile | rta | 16.6–16.7 | 1350–1403 | 956 465 | 39 981 of 310 418 (35 374) | 17 106 (6 071) | 625 of 1097 | 4 | 5 127 | 6/68/16/57/6 |
| petclinic | all | cha | 14.1–16.5 | 1363–1397 | 996 446 | — | — | 678 of 1096 | 4 | 5 127 | 6/72/12/57/6 |
| petclinic | all | rta | 16.5–18.7 | 1380–1403 | 990 536 | 5 910 of 310 418 (59 454) | 21 835 (9 923) | 676 of 1097 | 4 | 5 127 | 6/72/12/57/6 |
| WebGoat | profile | cha | 28.0–32.6 | 1783–1841 | 1 411 339 | — | — | 742 of 1194 | 0 | 1 843 | 42/52/9/61/0 |
| WebGoat | profile | rta | 27.8–29.8 | 1774–1821 | 1 327 988 | 83 351 of 556 769 (158 682) | 25 211 (10 671) | 662 of 1201 | 0 | 1 843 | 42/50/11/61/0 |
| WebGoat | all | cha | 27.2–29.8 | 1735–1802 | 1 411 344 | — | — | 743 of 1194 | 0 | 1 848 | 42/52/9/61/0 |
| WebGoat | all | rta | 27.8–32.8 | 1756–1795 | 1 396 558 | 14 786 of 556 769 (101 407) | 32 210 (13 181) | 745 of 1201 | 0 | 1 848 | 42/52/9/61/0 |

</details>

<details><summary>Before the service-provider dispatch model (servlet dispatch only, analyzer <code>9a460a6</code>)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | Servlets/filters registered | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 15.2–15.8 | 1381–1411 | 991 340 | — | — | 606 of 1096 | 4 | 3/74/13/57/6 |
| petclinic | profile | rta | 15.2–15.4 | 1351–1419 | 941 815 | 49 525 of 310 418 (48 701) | 16 011 (7 379) | 541 of 1097 | 4 | 3/70/17/57/6 |
| petclinic | all | cha | 15.2–15.3 | 1321–1367 | 991 340 | — | — | 606 of 1096 | 4 | 3/74/13/57/6 |
| petclinic | all | rta | 14.1–15.4 | 1373–1415 | 985 273 | 6 067 of 310 418 (71 771) | 21 662 (11 168) | 604 of 1097 | 4 | 3/74/13/57/6 |
| WebGoat | profile | cha | 23.0–24.9 | 1760–1784 | 1 409 574 | — | — | 676 of 1194 | 0 | 42/51/10/61/0 |
| WebGoat | profile | rta | 23.8–25.8 | 1766–1798 | 1 314 196 | 95 378 of 556 769 (178 188) | 24 398 (11 752) | 590 of 1201 | 0 | 42/49/12/61/0 |
| WebGoat | all | cha | 23.1–23.8 | 1766–1797 | 1 409 575 | — | — | 677 of 1194 | 0 | 42/51/10/61/0 |
| WebGoat | all | rta | 24.0–25.1 | 1772–1817 | 1 394 752 | 14 823 of 556 769 (122 892) | 32 123 (13 998) | 680 of 1201 | 0 | 42/51/10/61/0 |

</details>

<details><summary>P4 baseline (analyzer <code>a64edf7</code>, without the servlet dispatch model)</summary>

| App | Roots | Mode | Time (s) | Peak RSS (MB) | Edges | CHA edges removed (kept only by fallback) | Live classes (only by fallback) | Reachable gap sites | R/F/NR/U/NA |
|---|---|---|---|---|---|---|---|---|---|
| petclinic | profile | cha | 13.7–17.9 | 1335–1367 | 991 328 | — | — | 596 of 1096 | 3/71/16/57/6 |
| petclinic | profile | rta | 14.1–14.7 | 1315–1385 | 938 767 | 52 561 of 310 418 (50 305) | 15 818 (7 562) | 532 of 1097 | 3/65/22/57/6 |
| petclinic | all | cha | 13.5–14.9 | 1387–1394 | 991 328 | — | — | 596 of 1096 | 3/71/16/57/6 |
| petclinic | all | rta | 13.9–14.0 | 1359–1435 | 985 215 | 6 113 of 310 418 (78 194) | 21 603 (11 609) | 594 of 1097 | 3/71/16/57/6 |
| WebGoat | profile | cha | 23.1–23.4 | 1779–1829 | 1 409 574 | — | — | 676 of 1194 | 42/51/10/61/0 |
| WebGoat | profile | rta | 23.8–24.1 | 1783–1806 | 1 314 196 | 95 378 of 556 769 (178 188) | 24 398 (11 752) | 590 of 1201 | 42/49/12/61/0 |
| WebGoat | all | cha | 23.1–23.4 | 1772–1810 | 1 409 575 | — | — | 677 of 1194 | 42/51/10/61/0 |
| WebGoat | all | rta | 23.8–24.3 | 1776–1819 | 1 394 752 | 14 823 of 556 769 (122 892) | 32 123 (13 998) | 680 of 1201 | 42/51/10/61/0 |

</details>

- **Cost:** RTA adds about 0–2 s; peak memory is within run-to-run noise. The machine was
  busier during this run (load average ~5) than for the tables below, so absolute times are
  not comparable across tables; measured back to back on WebGoat, the service-provider
  dispatch model's own cost is 0.6–2 s, inside the main analyzer's own spread (23.5–27.2 s).
- **Edges:** `-roots profile` removes 13% (petclinic) / 15% (WebGoat) of the CHA edges; `-roots
  all` 2–3%. No CHA edge is ever added by RTA. Under RTA, 1 (petclinic) / 7 (WebGoat) more gap
  sites exist — the `Unsafe`/constructor-handle `allocation` gaps RTA introduces.
- **Unknown sites:** the fallback carries a large share of what RTA keeps (in `-roots
  profile`, 35 374 of the 270 437 CHA edges kept on petclinic, 158 682 of 473 418 on
  WebGoat): most of the narrowing RTA could do is given back to unresolved reflection.
- **Verdicts:** `-roots all` (the default roots) changes no verdict in either app. `-roots
  profile` moves 4 (petclinic) and 2 (WebGoat) findings from `FRAMEWORK` to
  `NOT_REACHABLE`, all carrying `MODEL_INCOMPLETE` — the same findings as before the
  service-provider dispatch model. No `REACHABLE` result changes anywhere.

## Undertow features (`internal/bootfeatures`, `u1`)

WebGoat's packaged configuration (`application-webgoat.properties`) sets no HTTP/2, AJP, PROXY
protocol, write timeout, Undertow login configuration or HTTP client, and its application code
references none of their APIs; `server.ssl.enabled=${WEBGOAT_SSLENABLED:false}` is an
environment placeholder, so TLS stays on. The 11 artifacts of the `s2` section below, after the
model (both roots scenarios, CHA):

| Advisory | Feature on the `s2` path | After `u1` | Check |
|---|---|---|---|
| CVE-2021-3859 | HTTP/2 HPACK and frame parsing | `NOT_REACHABLE` (`MODEL_INCOMPLETE`) | **correct for the packaged configuration**: HTTP/2 is off unless `server.http2.enabled` is set |
| CVE-2024-6162 | AJP request parsing | `NOT_REACHABLE` (`MODEL_INCOMPLETE`) | **correct**: Spring Boot never configures an AJP listener; the application adds none |
| CVE-2024-7885 | PROXY protocol listener | `NOT_REACHABLE` (`MODEL_INCOMPLETE`) | **correct**: the PROXY protocol is a listener flag only application code sets |
| CVE-2023-1973 | Undertow form authentication | `NOT_REACHABLE` (`MODEL_INCOMPLETE`) | **correct**: WebGoat authenticates with Spring Security; no Undertow login configuration |
| CVE-2023-1108 | TLS conduit | `REACHABLE` | TLS is the environment's choice (`WEBGOAT_SSLENABLED`); **plausible when TLS is on** |
| CVE-2024-1635 | (was: write-timeout conduit) | `REACHABLE`, now through `HttpReadListener#newRequest` → `ParseState#reset` | the broad mapping includes the HTTP/1.1 parser, which runs on every request — **plausible**, like CVE-2024-1459 |
| CVE-2026-28367, CVE-2026-28368, CVE-2026-28369 | (was: Undertow HTTP client) | `REACHABLE`, now ending at `HttpReadListener#handleEventWithNoRunningRequest` | the mapped server read loop runs on every request — **plausible target**, reached through a generic `ChannelListeners#invokeChannelListener` hop |
| CVE-2021-3690 | WebSocket pong handling | `REACHABLE` | still a **CHA artifact** (stream close → any listener → `BufferedBinaryMessage$1`); WebSockets are on (undertow-websockets-jsr on the class path), so the feature model does not apply |
| CVE-2023-5685 | XNIO future cancellation | `REACHABLE` in CHA; `NOT_REACHABLE` with RTA (`-roots profile`, `verdict-changes.tsv`) | still a **CHA artifact** (`dispatchRequest` → `HttpContinue` handler → WebSocket exchange); RTA removes it — the one new RTA-versus-CHA change, a correct one |

So of the 11: 4 removed, 5 now plausible or configuration-dependent, 2 artifacts left. A
deployment that enables HTTP/2, AJP or the PROXY protocol outside the packaged files (environment
variables, command line) is outside what the input shows; the `NOT_REACHABLE` results carry
`MODEL_INCOMPLETE`.

## Servlet request/response objects (`servlet-container:dispatch` `s2`)

What changed: a call on a `javax.servlet` request, response, wrapper or `FilterChain` object now
reaches the bodies of the implementations reachable code creates — Tomcat's
`RequestFacade`/`Request`/`ResponseFacade`/`Response`/`ApplicationFilterChain` on petclinic,
Undertow's `HttpServletRequestImpl`/`HttpServletResponseImpl`/`FilterHandler$FilterChainImpl` on
WebGoat, and the wrappers of filters that run. On petclinic no verdict changes (32 findings get
a different path, all `FRAMEWORK` before and after: the shortest path now enters Tomcat through
`ServletRequestAttributes#getAttribute` → `Request#getAttribute`). On WebGoat, in both roots
scenarios and both modes, **15 Undertow advisories go from `FRAMEWORK` to `REACHABLE`**: a path
now starts in application code (`Welcome#welcome` calling `request.getSession()`,
`AjaxAuthenticationEntryPoint#commence` calling `response.sendError(…)`) and continues into
Undertow. Every first hop is real — the application really calls that method on Undertow's own
request/response object. What follows it, checked advisory by advisory:

| Advisory | Vulnerable code (mapped) | Path after the first hop | Verdict |
|---|---|---|---|
| CVE-2023-4639 | `Cookies#parseCookie` | `getSession` → `ServletContextImpl#getSession` → `SessionCookieConfig#findSessionId` → `HttpServerExchange#getRequestCookie` → cookie parsing | **plausible**: a session lookup parses the request's cookies |
| CVE-2024-3884, CVE-2024-4027 | a broad mapping (57 methods) including `HttpServletRequestImpl#getQueryString` | Spring Security's `LoginUrlAuthenticationEntryPoint#buildHttpsRedirectUrlForRequest` → `getQueryString` | **plausible** as a call of a mapped method; the mapping itself is broad (a mapped accessor is not the vulnerable parsing) |
| CVE-2024-1459 | `HttpRequestParser#handlePath`, `ParseState#reset` | `sendError` → error dispatch → `HttpServerExchange#endExchange` → `HttpReadListener#exchangeComplete` → `newRequest` → `ParseState#reset` | **plausible**: ending an exchange on a keep-alive connection resets the parser for the next request |
| CVE-2024-1635 | `WriteTimeoutStreamSinkConduit` | `sendError` → `responseDone` → `ServletOutputStreamImpl#close` → `ConduitStreamSinkChannel#shutdownWrites` → `WriteTimeoutStreamSinkConduit#terminateWrites` | **CHA artifact for this input**: the conduit exists only with a write timeout configured (CHA over every `StreamSinkConduit`) |
| CVE-2023-1108 | `SslConduit#wrapAndFlip` | the same stream close → `SslConduit#write` | **CHA artifact**: TLS conduit, WebGoat serves plain HTTP |
| CVE-2021-3690, CVE-2021-3859, CVE-2023-5685, CVE-2024-6162, CVE-2024-7885, CVE-2026-28367, CVE-2026-28368, CVE-2026-28369 | WebSocket pong handling, HTTP/2 HPACK, XNIO future cancel, AJP decoding, PROXY protocol, the HTTP client | `endExchange` → `Http2ServerConnection#terminateRequestChannel` → `AbstractFramedStreamSourceChannel#close` → CHA over every `ChannelListener#handleEvent` (WebSocket, AJP, HTTP client, PROXY listeners); CVE-2023-5685 through `HttpContinue`/`AsyncWebSocketHttpServerExchange` | **CHA artifacts**: an HTTP/1.1 servlet exchange is not an HTTP/2 stream, and the listener fan-out has no receiver evidence |
| CVE-2023-1973 | Undertow's `FormAuthenticationMechanism` | error dispatch → `ServletInitialHandler#dispatchRequest` → `AuthenticationCallHandler` → Undertow's own security context | **CHA artifact for this input**: WebGoat authenticates with Spring Security, Undertow's form authentication is not configured |

So 4 plausible, 11 artifacts of CHA expansion *inside* Undertow (receivers not under the skip
rule, expanded to every implementation on the class path). The model did not create those paths;
it made them start in application code. Before `s2` the same vulnerable methods were already
reached from framework roots through the same internals (`FRAMEWORK`). Narrowing Undertow's
internal dispatch (listeners, conduits, protocol channels) is separate work; RTA does not remove
them either (the classes are created by reachable Undertow code). Petclinic's Tomcat paths do not
produce such changes on this input.

## SAM-call materialization (builder `4`)

What changed against builder `3` (CHA, `-graph-compare` with the builder `3` graph of the same
app and roots scenario, `651b3b5`):

| App | Roots | Edges added | Findings changed | Verdicts changed |
|---|---|---|---|---|
| petclinic | profile / all | `sam` 1 626, `model` 171 | 2 (path only) | 0 |
| WebGoat | profile | `sam` 1 557, `model` 77 | 1 (path only) | 0 |
| WebGoat | all | `sam` 1 558, `model` 77 | 1 (path only) | 0 |

The paths that changed, checked by hand in the bytecode — each a shorter path through a body
the JVM runs at that site:

- petclinic, CVE-2023-20860 and CVE-2026-47884 (`spring-webmvc`, still `FRAMEWORK`): the path now
  goes `SpringApplicationRunListeners#doWithListeners` → `lambda$contextLoaded$4`. `contextLoaded`
  creates `listener -> listener.contextLoaded(context)` and passes it to `doWithListeners`, which
  hands it to `List#forEach` (an "escape" site: the JDK runs it). The petclinic trace measured
  the sibling lambdas `lambda$starting$0` / `lambda$environmentPrepared$2` from the same site.
- WebGoat, CVE-2025-41249 (still `FRAMEWORK`): `Lazy#getNullable` → `JpaQueryMethod#lambda$new$0`.
  `JpaQueryMethod`'s constructor passes `() -> …` to `Lazy.of`, the `Lazy` constructor stores it
  in its final field `supplier`, and `getNullable` calls `supplier.get()`.

`sam_edges_gated` (2 530 / 2 265) counts candidates left out because no method creating the
value is reachable from the roots. The resolution is context-insensitive (a helper's parameter
gets every lambda any caller passes). A first version without that check changed one verdict on
petclinic: CVE-2021-23463 (H2 `JdbcSQLXML#setResult`) went from `NOT_REACHABLE` to `FRAMEWORK`
through `JdbcTemplate#execute(ConnectionCallback)` "running" `AbstractJdbcInsert`'s lambda,
reached from `EmbeddedDatabaseConnection#isEmbedded`, which passes a lambda of its own. The
creating method `executeInsertAndReturnKeyHolderInternal` is reachable from nothing, so no
execution of the reachable code holds that lambda. With the check the verdict is
`NOT_REACHABLE` again, as with builder `3`.

Cost, measured while building it (each fix kept the edges identical):

| Version | SAM step petclinic / WebGoat | Cause |
|---|---|---|
| first | 9.2 s / — | each method's class re-read and parsed per analysis |
| + class cache, parallel prefetch, parser buffer | 2.8 s / 12.5 s | `ResolveParameter` expanding every caller chain anew (64 callers × 6 levels) |
| + direct callers, each link memoized | 3.3 s / > 100 s | JRuby's interpreter (`Iterable#iterator` behind > 60 levels of `IRubyObject` parameters): cycle-cut outcomes recomputed per visit |
| + per-site memo of cycle-cut outcomes, declared-type pruning, 20 000-link budget | 1.9 s / 2.4 s | — |

Peak memory went up by 300 MB in the first version because `dataflow.CallerIndex` copied a whole
`hierarchy.Method` into each of its ~600 000 caller entries. It now keeps a pointer, and the facts
are resolved in batches of 1 024 methods, at most 4 096 kept.

## Servlet container dispatch (model `servlet-container:dispatch`, `s1`)

**Registered servlets/filters.** petclinic (a Spring Boot application, `@SpringBootApplication`
on `PetClinicApplication`): `DispatcherServlet` (DispatcherServletAutoConfiguration),
`OrderedCharacterEncodingFilter` (HttpEncodingAutoConfiguration), `OrderedFormContentFilter`
and `OrderedRequestContextFilter` (WebMvcAutoConfiguration); no `web.xml`, no
`@WebServlet`/`@WebFilter`, no servlet/filter bean, Spring Security absent (so no
`DelegatingFilterProxy`). WebGoat: **none** — the prepared input is `webgoat-container` plus
the vulnerable-components lesson (`scripts/prepare-webgoat.sh`); its launcher
`StartWebGoat` (`@SpringBootApplication`, `@ServletComponentScan`) lives in `webgoat-server`,
which is not part of the analyzed input, so nothing in the input enables Boot
auto-configuration and the model registers nothing (WebGoat's graph and verdicts are
identical to P4). This is the model working as specified, not a measurement of WebGoat's real
deployment.

**Edges (petclinic, 12 added by the model, identical in all four runs):**
`ApplicationFilterChain#internalDoFilter` → `HttpServlet#service(ServletRequest, …)` (the
body JVM selects for `DispatcherServlet`) and → `OncePerRequestFilter#doFilter` (the body
of the three filters); `HttpServlet#service(…)`/`#doHead` → `FrameworkServlet#service(Http…)`,
`#doGet/#doPost/#doPut/#doDelete/#doOptions/#doTrace`. Two more `Servlet#service` sites in
Spring MVC — `SimpleServletHandlerAdapter#handle` and
`ServletWrappingController#handleRequestInternal` — also get the registered set: the model
expands every reachable `Servlet#service` call to the registered servlets, so these are a
bounded over-approximation (they dispatch to a servlet wrapped as a handler, which petclinic
does not define). `DispatcherServlet#doDispatch` and `ResourceHttpRequestHandler#handleRequest`
are now reachable from `main` in both modes (BFS over the exported graphs; not reachable with
`a64edf7`).

**Verdict changes against P4 (same input, same mode; v1 per finding, `a64edf7` vs this
analyzer):** WebGoat none. petclinic — every change is `NOT_REACHABLE` → `FRAMEWORK`:

| Advisory | Vulnerable code | New path (from the servlet dispatch on) | Scenarios | Check |
|---|---|---|---|---|
| CVE-2024-38819 | `ResourceHttpRequestHandler#isInvalidPath` (mapped from the fix diff) | `internalDoFilter` → `HttpServlet#service` → `FrameworkServlet#service` → `DispatcherServlet#doDispatch` → `HttpRequestHandlerAdapter#handle` → `ResourceHttpRequestHandler#handleRequest` → `#getResource` → `#isInvalidPath` | all four | The method runs on every static-resource request (petclinic serves `/resources/**` and webjars through `ResourceHttpRequestHandler`). **Correct** for the mapped method; whether the advisory's functional-endpoint condition applies is outside reachability |
| CVE-2026-41844 | `DefaultRequestToViewNameTranslator#getViewName` | `… DispatcherServlet#doDispatch` → `#applyDefaultViewName` → `#getDefaultViewName` → `#getViewName` | all four | `applyDefaultViewName` runs after every handler; `getDefaultViewName` only when the handler set no view. A real path, **plausible** (not exercised here) |
| CVE-2023-20883 | `WelcomePageHandlerMapping#getHandlerInternal` and its welcome-page lookup | CHA: `internalDoFilter` → `OncePerRequestFilter#doFilter` → CHA on `doFilterInternal` to `CorsFilter` → `HandlerMappingIntrospector` → `AbstractHandlerMapping#getHandler` → `WelcomePageHandlerMapping#getHandlerInternal`; RTA: `… DispatcherServlet#doDispatch` → `#getHandler` → `AbstractHandlerMapping#getHandler` → `WelcomePageHandlerMapping#getHandlerInternal` | all four | Boot registers `WelcomePageHandlerMapping` and `DispatcherServlet` consults it for each request it has not matched earlier. **Correct** (RTA's path is the real one; CHA's shortest path goes through `CorsFilter`, which petclinic does not register — a CHA artifact on an otherwise reachable method) |
| CVE-2024-22259 | `UriComponentsBuilder#fromOriginHeader/#fromHttpUrl` | RTA: `… DispatcherServlet#doDispatch` → `HandlerExecutionChain#applyPreHandle` → `AbstractHandlerMapping$CorsInterceptor#preHandle` → `DefaultCorsProcessor#processRequest` → `CorsUtils#isCorsRequest` → `fromOriginHeader` | profile / rta only (CHA already `FRAMEWORK` through the WebFlux `CorsWebFilter` artifact) | P4's "undetermined": the servlet path is now visible. Spring MVC runs the CORS processor for a request with an `Origin` header when a CORS configuration applies or for a pre-flight request; the path is real, **plausible** for an attacker-sent pre-flight (not exercised here) |
| CVE-2024-38809 | `ServletWebRequest#validateIfNoneMatch` | profile / rta: reported path is still the shortest one, through the reactive `WriteBarrier` → `DeferredResultSubscriber#onNext` (that subscriber is now live: WebMvc.fn code reachable from `DispatcherServlet` creates it); the real path `… DispatcherServlet#doDispatch` → `HttpRequestHandlerAdapter#handle` → `ResourceHttpRequestHandler#handleRequest` → `ServletWebRequest#checkNotModified(J)` → `#checkNotModified(String, J)` → `#validateIfNoneMatch` is in the graph in both modes | profile / rta only (CHA already `FRAMEWORK`) | P4's **false negative is fixed**: `FRAMEWORK` in every petclinic scenario, and the method is reachable through the request path that really runs it. The report shows the shortest path, which is not the real one — read it as "reachable", not as the route |

So the model adds 3 `FRAMEWORK` results to every petclinic scenario and 2 more under `-roots
profile` with RTA, all through the request path; none is a known false positive, and no
`REACHABLE` result changes (the `FRAMEWORK`/`REACHABLE` split is unchanged: the new paths
run entirely through framework code). The petclinic RTA-vs-CHA changes drop from 6 to 4 (the
four Tomcat valve/realm findings, still correct).

## Service-provider dispatch (model `service-provider:dispatch`, `p1`)

Found by a runtime trace (RUNTIME_NATIVE N1): on petclinic the JVM ran
`HibernateValidator#buildValidatorFactory` (called by Hibernate's
`AbstractConfigurationImpl#buildValidatorFactory` through
`javax.validation.spi.ValidationProvider`, a skipped receiver) while the graph stopped there.
The model expands calls on a skipped interface that is a registered `META-INF/services` service
(to its registered providers) or a product of one (an interface of the same API family reached
through the service's return types; to implementations in the providers' jars) — see the README.

**Catalog.** petclinic: 8 skipped service interfaces with registered providers, 1 096 call
signatures, 5 127 model edges in total (servlet dispatch included); WebGoat: 11 services,
1 105 signatures, 1 843–1 848 model edges. Families on petclinic include Bean Validation
(`javax.validation`, Hibernate Validator), JDBC (`java.sql`, H2 and MySQL drivers) and JCache
(`javax.cache`, Ehcache). A first version that took every interface of the family (not only
products) also sent `javax.servlet` calls to spring-web's servlet classes — a family brought
in by `ServletContainerInitializer`, which hands nothing out — and turned CVE-2021-22118 into a
false `FRAMEWORK`; the return-type closure removed it.

**Verdict changes against the analyzer without the model** (same input, `-roots all`, CHA,
`-graph-compare`; identical under `-roots profile` and RTA), each checked against a JFR
recording of petclinic with exact entries (`jdk.MethodTrace`) of the methods involved:

| App | Advisory | Change | New edge on the path | Runtime check | Verdict |
|---|---|---|---|---|---|
| petclinic, WebGoat | CVE-2025-35036 (Hibernate Validator) | `NOT_REACHABLE` → `FRAMEWORK` | `MessageInterpolatorFactory#getObject` → `AbstractConfigurationImpl#getDefaultMessageInterpolator` (a `javax.validation.Configuration` product) | `HibernateValidator#buildValidatorFactory` entered 3 times at startup; the advisory's mapped methods entered (26 exact entries in the vulnerable-method recording) | **correct**: a static false negative removed |
| petclinic | CVE-2021-46877, CVE-2022-45868, CVE-2023-6378 | `FRAMEWORK` → `REACHABLE` | `CacheConfiguration#lambda$petclinicCacheConfigurationCustomizer$0` (application code) → `Eh107CacheManager#createCache` (a `javax.cache.CacheManager` product of Ehcache's `CachingProvider`) | `Eh107CacheManager#createCache` entered (the first hop is real); the rest of the path — Ehcache's disk store `FileBasedStateRepository#loadMaps` → deserialization → `NodeSerialization#readExternal` — is pre-existing CHA/deserialization over-approximation: 0 entries | first hop **correct**; the `REACHABLE` result is as plausible as the downstream path, i.e. not (petclinic's caches are heap-only) |

No `REACHABLE` result is lost; RTA-versus-CHA changes are unchanged (the 4 + 2 above).

## JPetStore (WAR, `web.xml` registrations)

MyBatis JPetStore 6.0.2 (`scripts/prepare-jpetstore.sh`: built with `-source/-target 8` on the
current JDK, the WAR analyzed with `-deps-list`; 17/17 dependency jars identified, 35 advisories,
18 mapped). Runtime trace: [TRACE_GAPS.MD](TRACE_GAPS.MD), "A second application".

**Each model, one at a time.** The same input through the analyzer after every model merged so
far, each graph compared with the previous one (`-graph-compare`, CHA):

| Analyzer | Model added | `-roots profile` edges | `-roots all` edges | Verdicts (both scenarios) |
|---|---|---|---|---|
| `95ae6ec` | (P4 baseline) | 221 346 | 221 346 | profile 0/9/7/19/0, all 0/1/15/19/0 |
| `436695c` | servlet container dispatch `s1` | +0 | +0 | unchanged |
| `f27c8f3` | service-provider dispatch `p1` | +98 | +7 | unchanged |
| `36b361f` | CHA inherited bodies (builder `3`) | +3 190 | +3 190 | unchanged (1 path shorter) |
| `2aa71d9` | SAM-call materialization (builder `4`) | +0 | +0 | unchanged (no lambdas: Java 6 sources) |
| `8298256` | servlet objects `s2` | +255 | +0 | unchanged |
| `fac65dc` | Undertow features `u1` | +0 | +0 | unchanged (Jetty/Tomcat app) |
| `jdbc-objects` | JDBC objects `j1` | +811 | +739 | unchanged |

The servlet models have little to work with in a WAR: the container is not part of the input,
so nothing calls `Servlet#service`/`Filter#doFilter` and no request, response or filter-chain
class has a creator. Only `StripesFilter` is registered: Stripes' `DispatcherServlet` extends
`javax.servlet.http.HttpServlet`, and the servlet API is `provided` (not in the WAR), so the
hierarchy cannot tell it is a servlet — harmless here (registration only matters for container
dispatch), but a WAR's servlets are generally not recognized without the API jar. The web.xml
lifecycle methods are roots under `-roots profile` either way.

**The two root scenarios.** 8 advisories are `FRAMEWORK` with `-roots profile` and
`NOT_REACHABLE` with the default `-roots all`: CVE-2022-22950, CVE-2022-22965, CVE-2022-22970,
CVE-2023-20861, CVE-2023-20863, CVE-2024-38808, CVE-2024-38820, CVE-2026-41848. Every profile
path starts at `ContextLoaderListener#contextInitialized` or `#contextDestroyed` — a library class
the application registers in `web.xml`, so a root of `-roots profile` but not of `-roots all`
(every *application* method). Checked against the runtime trace: the mapped methods of
CVE-2022-22965/CVE-2022-22970 (`CachedIntrospectionResults#<init>`) and CVE-2026-41848
(`AntPathMatcher$AntPathStringMatcher#matchStrings`) **ran**; the other 5 were traced on entry and
not entered in this scenario. So the default scenario gives **3 false negatives** here, where
`-roots profile` is right (`FRAMEWORK`: Spring's own bootstrap, not application code). The
`-roots profile` paths are plausible: `ContextLoader#initWebApplicationContext` →
`SingletonBeanFactoryLocator` → `PathMatchingResourcePatternResolver` → `AntPathMatcher`;
`ContextCleanupListener` → `DisposableBeanAdapter` → `BeanUtils#getPropertyDescriptors` →
`CachedIntrospectionResults`; bean creation → `evaluateBeanDefinitionString` → SpEL for the SpEL
advisories (CVE-2024-38820's path goes through a CHA `afterPropertiesSet`, below).

**Fixed (2026-11-09):** the web.xml lifecycle methods are now roots of `-roots all` too (every
application method plus them; log line `roots: all (N methods: every application method, M more
from web.xml)`). JPetStore `-roots all`: 312 roots (6 from web.xml), +418 edges (the creator
gates see more reachable creators), the 8 advisories above `NOT_REACHABLE` → `FRAMEWORK` — the
same verdicts as `-roots profile` (0/9/7/19/0 CHA, 0/8/8/19/0 RTA). petclinic and WebGoat have no
`web.xml`: graphs and verdicts unchanged in the benchmark.

**RTA versus CHA** (`verdict-changes.tsv`; before the fix — after it, the `-roots all` row is
CVE-2024-38820 like `-roots profile`, and CVE-2026-41851 is `FRAMEWORK` in both modes through
Spring's bootstrap):

| Roots | Advisory | Why CHA reached it | Check | Verdict |
|---|---|---|---|---|
| profile | CVE-2024-38820 | `invokeInitMethods` → CHA to `ContentNegotiationManagerFactoryBean#afterPropertiesSet` → `MappingMediaTypeFileExtensionResolver` | JPetStore configures no Spring MVC; nothing reachable creates that factory bean; the trace shows the mapped method traced and not entered | **correct** |
| all | CVE-2026-41851 | Stripes `Log#error` → `Log4JLogger#error` → log4j `AppenderAttachableImpl` → CHA to every `Appender` → … → SpEL | The CHA path is an artifact (no appender of JPetStore's log4j configuration evaluates SpEL); RTA removes it at `Log4JLogger`, which commons-logging's `LogFactoryImpl` creates reflectively (the trace has that `Constructor#newInstance` relation) — RTA drops a body the JVM runs. `SpelParserConfiguration#<init>` does run, from Spring's bootstrap, which `-roots all` does not root | artifact removed; **false negative** of the `-roots all` scenario (as above), not of RTA's narrowing |

**HSQLDB and log4j 1.2 mapped (2026-11-10).** 7 of JPetStore's unmapped advisories now have
manual sink mappings ([audit](../cve-control/MANUAL_MAPPING_AUDIT.MD), "HSQLDB and log4j 1.2
sinks"): 25 of 35 mapped. All 7 were `UNKNOWN` (no mapping); now, the same in both root
scenarios:

| Advisory | Mapped (sink) | CHA | RTA | Check (trace: every mapped method traced on entry, **none entered**) |
|---|---|---|---|---|
| CVE-2022-41853 (HSQLDB) | `Routine#getMethods`, `#invokeJavaMethod` | REACHABLE | FRAMEWORK | CHA: Stripes `Log#trace` → log4j → CHA to every `Appender` → `JDBCAppender#execute` → `JDBCStatement#executeUpdate` → … → `StatementSchema#getResult` → `Routine#resolveReferences` → `getMethods` — an artifact (JPetStore configures only a `ConsoleAppender`). RTA: `ContextLoaderListener#contextDestroyed` → Spring `EmbeddedDatabaseFactory#shutdownDatabase` → `Statement#execute` → HSQLDB's statement execution → `getMethods`. That path exists for *some* SQL text (a statement creating a Java routine); the graph does not model the text (here `SHUTDOWN`), so `FRAMEWORK` is "may run", not proof — the exploit condition (attacker-controlled SQL) is outside the analysis either way |
| CVE-2021-4104 | `JMSAppender#activateOptions`, `#lookup` | REACHABLE | NOT_REACHABLE | CHA: log4j's `PropertySetter#activate` → CHA to every `OptionHandler#activateOptions` — artifact; nothing reachable creates a `JMSAppender` (log4j creates appenders reflectively from its configuration file, which names only `ConsoleAppender`): RTA correct |
| CVE-2022-23305 | `JDBCAppender#getLogStatement`, `#flushBuffer`, `#execute` | REACHABLE | NOT_REACHABLE | the same `Appender#doAppend` artifact as above; RTA correct |
| CVE-2019-17571, CVE-2022-23302, CVE-2022-23307, CVE-2023-26464 | `SocketNode#run`, `JMSSink`, Chainsaw `Slurper#run` | NOT_REACHABLE | NOT_REACHABLE | receivers the application never starts; correct (`MODEL_INCOMPLETE` kept: reflection gaps) |

Verdicts: CHA 0/9/7/19/0 → **3/9/11/12/0**, RTA 0/8/8/19/0 → **0/9/14/12/0**. The 3 CHA
`REACHABLE` are false positives through log4j's appender dispatch, the case RTA already removed
for CVE-2026-41851; RTA gets all 7 right except that HSQLDB's `FRAMEWORK` is SQL-text-insensitive.

**What this says about `j1`:** nothing new. The `Statement` edge on both HSQLDB paths exists
without `j1` — `service-provider:dispatch` (`p1`) expands `java.sql` *product* interfaces of the
`java.sql.Driver` HSQLDB registers in `META-INF/services` to HSQLDB's own implementations. The
analyzer before `j1` (`95d7c85^`) with the new table (`-vuln-methods`) gives the same verdicts and
paths for all 7 in `-roots profile`; in `-roots all` it differs only by the web.xml roots fixed
since (the RTA path starts at a web.xml listener). An application whose driver is not registered
as a service, or whose statements come from a pool, is where `j1` matters; JPetStore is neither.

**WebGoat (HSQLDB 2.5.1) changes too:** CVE-2022-41853 `UNKNOWN` → `REACHABLE` in all four
configurations (CHA 53/37/13/61/0 → 54/37/13/60/0; RTA one `UNKNOWN` → `REACHABLE` likewise). A
true positive with its exploit condition met: `SqlInjectionLesson4#completed(@RequestParam String
query)` runs the request parameter with `Statement#executeUpdate`, so attacker SQL can create and
call a Java routine. The reported (shortest) path starts at `LessonConnectionInvocationHandler#invoke`;
the analyzer before `j1` gives the same verdict (the `Statement` edge again from `p1`). petclinic
(H2, logback) unchanged.

**log4j 1.2 appender model (`l1`, 2026-11-10).** The appenders JPetStore's
`log4j.properties` does not name (all but `ConsoleAppender`) are no longer CHA/RTA targets
(`internal/log4jconfig`; README, "log4j 1.2 appenders"). `-graph-compare` with the analyzer before
it, both root scenarios: CHA −170 edges, RTA −106, none added. CHA verdicts 3/9/11/12/0 →
**0/10/13/12/0**: CVE-2021-4104 and CVE-2022-23305 `REACHABLE` → `NOT_REACHABLE`, CVE-2022-41853
`REACHABLE` → `FRAMEWORK` through the same Spring shutdown path RTA reports — the three CHA false
positives of the table above are gone, and CHA now agrees with RTA on every JPetStore verdict
but CVE-2024-38820 (checked above: RTA's change is correct). RTA verdicts unchanged (it had removed these paths
already). Control app `scripts/prepare-log4j1-demo.sh` (same classes, console-only vs
`JDBCAppender` configuration, run on the JVM): CVE-2022-23305 `NOT_REACHABLE`/`REACHABLE`,
CVE-2021-4104 `NOT_REACHABLE` in both — matching the JVM; before the model all four were
`REACHABLE`.

**logback appender model (`l2`, builder `5`, 2026-11-10).** petclinic and WebGoat have no logback
file, so Spring Boot's default configuration decides: `ConsoleAppender` on, the other 15 appenders
in the input off (`internal/logbackconfig`). `-graph-compare` with `main`:

| App | CHA edges | RTA edges | Findings changed |
|---|---|---|---|
| petclinic | −288 (both roots) | −165 profile, −293 all | CVE-2023-6378 (logback-core instance) `FRAMEWORK` → `NOT_REACHABLE` (CHA, and RTA `all`) |
| WebGoat | −755 (both roots) | −162 profile, −760 all | the same |
| JPetStore | −7 (builder `5` over `l1`'s exclusions) | −7 | none |

No advisory verdict changes. The three appender paths checked by hand (all artifacts: Boot's
defaults configure no `DBAppender`, `SMTPAppender` or file appender) are gone, but:
- CVE-2023-6378 stays `REACHABLE` through its logback-classic instance (`LoggingEventVO#readObject`
  from a real deserialization entry — ehcache's `FileBasedStateRepository` on petclinic, the
  vulnerable-components lesson on WebGoat);
- CVE-2021-42550 and CVE-2024-12798 stay `FRAMEWORK` through **another** CHA over-approximation:
  logback's `LifeCycle#start`/`#stop`/`#isStarted` calls (`StatusListenerConfigHelper#initAndAddListener`
  → every `LifeCycle` → `JaninoEventEvaluatorBase#start`; `ContextJNDISelector` →
  `LifeCycleManager#reset` → `ConnectionSourceBase#isStarted`). Those components are created by
  Joran from the configuration too; the model covers appenders only.

Control app `scripts/prepare-logback-demo.sh` (Spring Boot 2.4.5, logback 1.2.3; the same classes
with Boot's defaults, with `logging.file.name`, and with a `logback-spring.xml` adding a
`SMTPAppender` with a Janino evaluator): the model's appenders equal the root logger's appenders
the JVM prints in all three. CVE-2024-12798's path goes through `SMTPAppender#append` →
`JaninoEventEvaluatorBase#evaluate` only in the SMTP variant; in the other two it is the
`LifeCycle#start` path above.

**logback Joran components (`l2-2`, builder `6`, 2026-11-10).** The configuration now also decides
logback's Joran-only components (concrete `LifeCycle` classes no code creates with `new`, minus
appenders and pattern converters): on petclinic and WebGoat 30 of 32 are off. `-graph-compare` with
`main` (before `l2`):

| App | CHA edges | RTA edges | Findings changed vs `l2-1` |
|---|---|---|---|
| petclinic | −582 (both roots) | −247 profile, −587 all | CVE-2024-12798 `FRAMEWORK` → `NOT_REACHABLE` (all 4 runs) |
| WebGoat | +4 / −1 046 | +4 / −241 profile, −1 051 all | the same |
| JPetStore | +9 / −7 | +9 / −7 | none |

The `+` edges are builder `6`: builder `5` also dropped a kept class's body inherited from an
excluded declaring class (log4j's `ConsoleAppender` lost `WriterAppender#append`; logback's
`OutputStreamAppender#append` likewise) — a false narrowing, corrected; against `main` no edge
the model should keep is missing. CVE-2024-12798 checked: Boot's defaults create no evaluator, so
`JaninoEventEvaluator` never exists — `NOT_REACHABLE` is right. CVE-2021-42550 stays `FRAMEWORK`
through `StaticLoggerBinder` → `ContextJNDISelector#getLoggerContext`/`findConfigFileURL`: logback
creates that selector with `new`, but only when `-Dlogback.ContextSelector=JNDI` is set — real code,
switched by a JVM setting the input does not show. Control app, fourth variant (`EvaluatorFilter`
with a class-less `<evaluator>`): the JVM logs `Assuming default type [JaninoEventEvaluator] for
[evaluator] property`; the model keeps it on and CVE-2024-12798 stays `FRAMEWORK` there and in the
SMTP variant, `NOT_REACHABLE` in the default and file variants.

## Framework proxies (`framework:proxies` `x1`, FRAMEWORKS F3 slice 3)

2026-11-10: calls on MyBatis mappers, `SqlSessionTemplate`'s `SqlSession` proxy, Spring Data
repositories and Spring ORM's JPA interfaces reach the framework's `InvocationHandler`, creator-gated
(README, "Framework proxies"). `-graph-compare` with `main`:

| App | Detected | Edges added (CHA profile / all) | Verdicts |
|---|---|---|---|
| JPetStore | 7 mappers (mapper XML namespaces) | +16 / +175 | `-roots all` CHA: CVE-2022-41853 `FRAMEWORK` → **`REACHABLE`**; others unchanged |
| petclinic | 4 repositories | +522 / +522 (RTA profile +1 082) | unchanged; 63 paths changed (CHA all) |
| WebGoat | — (Spring Data repositories and JPA) | +562 / +562 (RTA profile +1 127) | unchanged; 23 paths changed |

Checked by hand:
- **JPetStore CVE-2022-41853** — `CatalogService` → mapper → `MapperProxy#invoke` →
  `MapperMethod#execute` → `DefaultSqlSession` → `BatchExecutor` → `JDBCStatement` → HSQLDB's
  statement execution → `Routine#getMethods`: the application's own queries run on HSQLDB 2.3.4,
  so the Java-routine code is reachable from application code (that attacker SQL is needed to
  exploit it stays outside the analysis; JPetStore's mappers use `#{}` parameters). Under
  `-roots profile` the path starts at a framework entry, so the label stays `FRAMEWORK`; under RTA
  the `SqlSession` behind `MapperMethod` is Spring's `SqlSessionTemplate`, created only
  reflectively, so the path does not continue — both expected.
- **petclinic/WebGoat paths** — every changed path goes through an application call on a repository
  → `JdkDynamicAopProxy#invoke` → CHA over Spring AOP's `MethodInterceptor`/`TargetSource`
  (`EventPublicationInterceptor`, `LazyInitTargetSource` → `YamlMapFactoryBean` …): shorter paths
  to the same `FRAMEWORK` verdicts, through interceptors petclinic does not configure — CHA
  over-approximation inside Spring AOP, not new reachability.
- **Runtime traces** — petclinic (new recording): `proxy` relations not in the graph 13 → 0, all
  relations not in the graph 148 → 127; JPetStore: 8 → 2 (`SqlSessionTemplate` →
  `SqlSessionInterceptor`, gated out: the template is created by Spring's reflection).

## Every verdict change, checked by hand

`FRAMEWORK` → `NOT_REACHABLE` under RTA, `-roots profile` (`verdict-changes.tsv`). With the
servlet dispatch model the petclinic rows for CVE-2024-22259 and CVE-2024-38809 are gone
(both `FRAMEWORK` in both modes, see above); the WebGoat CVE-2024-38809 row stays, because
nothing in WebGoat's analyzed input registers `DispatcherServlet`. The table below keeps the
P4 analysis of all 8 for the record:

| App | Advisory | Vulnerable code | Why CHA reached it | Check | Verdict |
|---|---|---|---|---|---|
| petclinic | CVE-2022-45143 | `JsonErrorReportValve#report` | `ErrorReportValve#invoke` → CHA to the subclass | No class of the input references `JsonErrorReportValve` (bytecode scan of every jar, slashed and dotted names): Tomcat creates valves only from `server.xml`/`context.xml`, which embedded Spring Boot does not use | **correct** |
| petclinic | CVE-2025-55752, CVE-2026-53404 | `RewriteValve#invoke/#parse`, `RewriteRule#evaluate` | `LifecycleBase#start` → CHA to `RewriteValve#startInternal` | Same: `RewriteValve` is referenced by no other class; configuration-only | **correct** |
| petclinic | CVE-2026-55957 | `JNDIRealm#bindAsUser` | `BasicAuthenticator#doAuthenticate` → CHA to `JNDIRealm#authenticate` | Only `MBeanFactory#createJNDIRealm` (a JMX management operation, unreachable) does `new JNDIRealm`; realms otherwise come from `server.xml` | **correct** |
| WebGoat | CVE-2021-22118 | reactive `DefaultPartHttpMessageReader` | WebFlux `DefaultServerWebExchange#initFormData` → CHA into the reactive multipart readers | The reactive stack is present in `spring-web` but WebGoat is a servlet MVC application: no reachable code creates a WebFlux exchange or these readers | **correct** |
| petclinic | CVE-2024-22259 | `UriComponentsBuilder#fromHttpUrl/#fromOriginHeader` | WebFlux `DefaultWebFilterChain#invokeFilter` → CHA to `CorsWebFilter#filter` | The CHA path was an artifact (WebFlux, unused). Whether a servlet path exists cannot be told from either graph: servlet request dispatch is cut in both modes (see below) | artifact removed; real reachability **undetermined** |
| petclinic, WebGoat | CVE-2024-38809 | `ServletWebRequest#validateIfNoneMatch` (ETag parsing) | reactive `ChannelSendOperator$WriteBarrier#emitCachedSignals` → CHA on `Subscriber#onNext` to the servlet-functional `PublisherEntityResponse$DeferredResultSubscriber` | The CHA path was an artifact (that subscriber is created only by WebMvc.fn publisher responses, which neither app uses). **But the method really runs:** `ResourceHttpRequestHandler#handleRequest` → `ServletWebRequest#checkNotModified(J)` → `validateIfNoneMatch` on every static-resource request. Neither graph reaches it, because `ApplicationFilterChain#internalDoFilter` → `javax.servlet.Servlet#service` is a `javax.*` receiver the skip rule never expands, so `DispatcherServlet` is unreachable in both modes | **false negative** under RTA (CHA's `FRAMEWORK` was right only by accident) |

So of P4's 8 changes: 5 correct, 1 undetermined, 2 (one advisory, both apps) a false
negative. After the servlet dispatch model, 6 remain: the 5 correct ones and WebGoat's
CVE-2024-38809, still a false negative of this *input* (its launcher is not analyzed;
`-roots profile` has no root that reaches Undertow's dispatch registered to anything).
The false negative is not RTA's fixed point going wrong — RTA removed exactly the absurd
edge it should — but it shows that a narrower graph turns a gap that CHA's
over-approximation happened to paper over (the unexpanded servlet dispatch) into a
`NOT_REACHABLE`. `MODEL_INCOMPLETE` is attached, but its gap counts name reflection,
deserialization and the like, not the skipped servlet dispatch.

## Decision: the default stays CHA

RTA stays opt-in (`-callgraph-mode rta`); the default strategy is not changed. Re-evaluated
after the servlet dispatch model:

1. **The petclinic false negative is gone** (CVE-2024-38809 is `FRAMEWORK` in every
   scenario, through the real request path), and P4's undetermined CVE-2024-22259 now has a
   visible servlet path. What remains of RTA's verdict changes on petclinic are the four
   correct Tomcat removals.
2. **But a false negative of the same kind remains wherever the input does not show the
   registration** — WebGoat's analyzed input lacks its Boot launcher, and the model's own
   gaps (servlets/filters registered by a `ServletContainerInitializer`, Actuator, Spring
   Security's inner chain, `ServletContext#addServlet`, `FilterRegistrationBean` with an
   unknown class) are not all reported as gaps of a finding. RTA turns such a missing
   dispatch into `NOT_REACHABLE`; CHA's over-approximation often still papers over it.
3. **With the default `-roots all`, RTA still changes no verdict** on either app: the default
   user would pay the extra pass and gain nothing visible.
4. **The fallback is a heuristic:** a class created by reflection and named only in a
   configuration file is not live (it is the reason the four Tomcat removals are right for
   these apps — and would be wrong for an app whose `server.xml` declares those valves).

Points-to (GRAPH_PRECISION P3) is still not justified by these numbers: on both apps the
fallback, not the dispatch precision, decides most of what RTA keeps.

## Using RTA (migration)

- Opt in with `-callgraph-mode rta`; nothing changes for a run without it. Use it as a
  second opinion with `-roots profile`, always next to a CHA run of the same input and
  configuration: `-graph-out` from the CHA run, `-graph-compare` in the RTA run, then read
  the "Call graph compared" section — every finding whose verdict or path changed, with the
  removed edges that explain it.
- What changes in the outputs: the report gets an "RTA" section; v2 gets `callgraph.mode =
  "rta"`, `callgraph.unknown_allocation_policy = "cha-fallback-with-gap"` and
  `callgraph.rta` (`docs/RESULTS_CONTRACT.MD`); reachable `Unsafe`/constructor-handle sites
  count as `allocation` model gaps of `NOT_REACHABLE` findings; snapshots record
  `alloc.catalog` and `callgraph.rta` among the algorithm versions (only in RTA mode), so a
  snapshot comparison flags a change of those rules.
- Treat an RTA-only `NOT_REACHABLE` as "no path in the narrowed model", never as proof:
  check whether the removed edge's class is created by configuration (XML, properties,
  `META-INF` files other than `services`) or reached through a framework dispatch the graph
  does not expand — for servlet containers, check the `servlet dispatch` log line: a servlet
  or filter missing from it (registered outside the analyzed input, by a container
  initializer or programmatically) is dispatched to by nothing in the graph.
