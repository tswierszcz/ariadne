"""Does the call graph contain a real path? (ASYNC_CALLBACKS a3-3, 2026-11-19)

Usage:
  python3 -I scripts/real-edge-paths.py GRAPH RESULTS.json EDGES ran.txt [API_PREFIX]

GRAPH is an ariadne -graph-out file, RESULTS.json the -json-format v2 file of
the same run, EDGES the caller<TAB>callee pairs a JVM run recorded
(java scripts/MethodEdgeTrace.java CP MAIN edges.txt 'pkg.*'), ran.txt the
methods that ran (MethodBreakTrace). An edge counts when the JVM took it, or
when it is a modeled cross-thread edge (basis callback: A2, a3-2, a3-3)
whose two ends both ran — a hand-off never shows as caller and callee on
one stack. Per REACHABLE finding: how many exposed-API entries that ran
reach the vulnerable method over such edges. Unlike path-divergence.py,
which checks the one reported (shortest) path, this asks whether any real
path exists in the graph. With API_PREFIX, prints that entry's path.
"""
import collections
import json
import sys


def main():
    graph, results, edges, ran_file = sys.argv[1:5]
    show = sys.argv[5] if len(sys.argv) > 5 else None
    ran = {l.strip() for l in open(ran_file) if l.strip() and not l.startswith("#")}
    real = {tuple(l.rstrip("\n").split("\t")) for l in open(edges) if "\t" in l}
    adj, kind = collections.defaultdict(set), {}
    for l in open(graph):
        p = l.rstrip("\n").split("\t")
        if p[0] != "edge":
            continue
        a, b, basis = p[1], p[2], p[4]
        if (a, b) in real or (basis == "callback" and a in ran and b in ran):
            adj[a].add(b)
            kind.setdefault((a, b), "real" if (a, b) in real else basis)

    def path(src, dst):
        prev, q = {src: None}, collections.deque([src])
        while q:
            u = q.popleft()
            if u == dst:
                out = []
                while u:
                    out.append(u)
                    u = prev[u]
                return out[::-1]
            for v in sorted(adj[u]):
                if v not in prev:
                    prev[v] = u
                    q.append(v)
        return None

    for f in json.load(open(results))["findings"]:
        if f["reachability"] != "REACHABLE":
            continue
        sink = f["path"][-1]
        entries = [e["path"][0] for e in f.get("exposed_api") or [] if e["path"][0] in ran]
        ok = [e for e in entries if path(e, sink)]
        print(f'{f["advisory_id"]}: {len(ok)} of {len(entries)} exposed-API entries that ran reach {sink.split("(")[0]} over real edges')
        for e in entries:
            if show and e.startswith(show):
                p = path(e, sink)
                print(f'  {e.split("(")[0]}: {"none" if not p else str(len(p)) + " methods"}')
                for a, b in zip(p or [], (p or [])[1:]):
                    print(f'    {kind[(a, b)]:9} {b.split("(")[0]}')


if __name__ == "__main__":
    main()
