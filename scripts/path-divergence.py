"""Where do reported paths leave the real execution? (LIBRARY_MODE, 2026-11-19)

Usage:
  python3 -I scripts/path-divergence.py classes RESULTS.json > classes.txt
      the classes on every REACHABLE finding's path and exposed_api path
      (trace them: java scripts/MethodBreakTrace.java CP MAIN ran.txt $(cat classes.txt))
  python3 -I scripts/path-divergence.py diverge RESULTS.json ran.txt [N]
      per path, the first method that never ran; prints how many paths are
      fully executed and the N most common divergence edges with their
      call kind/basis (v2 path_edges evidence), e.g. "interface/cha".

RESULTS.json is an ariadne -json-format v2 file; ran.txt one owner#name+desc
per line (MethodBreakTrace/MethodTrace output).
"""
import collections
import json
import sys


def paths(doc):
    for f in doc["findings"]:
        if f["reachability"] != "REACHABLE":
            continue
        yield f["path"], f.get("path_edges") or []
        for e in f.get("exposed_api") or []:
            yield e["path"], e.get("path_edges") or []


def main():
    cmd, doc = sys.argv[1], json.load(open(sys.argv[2]))
    if cmd == "classes":
        for c in sorted({m.split("#")[0] for p, _ in paths(doc) for m in p}):
            print(c)
        return
    ran = {l.strip() for l in open(sys.argv[3]) if l.strip() and not l.startswith("#")}
    top = int(sys.argv[4]) if len(sys.argv) > 4 else 15
    total, full, div, kind = 0, 0, collections.Counter(), {}
    for p, pe in paths(doc):
        total += 1
        i = next((k for k, m in enumerate(p) if m not in ran), None)
        if i is None:
            full += 1
            continue
        if i == 0:
            div[("", p[0])] += 1
            continue
        key = (p[i - 1], p[i])
        div[key] += 1
        ev = pe[i - 1].get("evidence", []) if i - 1 < len(pe) else []
        kind[key] = ",".join(sorted({x["kind"] + "/" + x["basis"] for x in ev}))
    print(f"{total} paths, {full} fully executed, {len(div)} distinct divergence edges")
    for (a, b), n in div.most_common(top):
        src = a.split("(")[0] if a else "(root never ran)"
        print(f"{n:5}  {src} -> {b.split('(')[0]}  [{kind.get((a, b), '')}]")


if __name__ == "__main__":
    main()
