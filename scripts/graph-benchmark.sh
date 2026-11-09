#!/usr/bin/env bash
# GRAPH_PRECISION P4: call-graph strategy benchmark on the prepared real
# applications (scripts/prepare-petclinic.sh, scripts/prepare-webgoat.sh),
# offline, -profile spring. For every app, roots scenario (profile, all)
# and mode (cha, rta) it runs bin/ariadne and records, per run: wall time
# and peak RSS of the analyzer process, graph size, CHA edges (and with
# RTA: removed / kept only by the unknown-allocation fallback), live
# classes, active unknown-allocation gaps, reachable model-gap sites, the
# v1 verdict counts, the servlets/filters the app registers and the edges
# the B3 models add (servlet container dispatch), the SAM-call edges
# (DYNAMIC_LINKAGE L1, added / left out for an unreachable creator), and — RTA against the CHA run of the same scenario,
# via -graph-compare — the edges removed and every finding whose verdict
# or path changed. Each configuration runs REPEAT measured times with no
# graph file written or read (time and memory are the minimum and maximum
# over them; every other figure must be identical, or the script fails),
# plus one unmeasured run that writes (CHA) or compares against (RTA) the
# graph file — reading a several-hundred-MB graph file would otherwise be
# charged to the strategy.
#
# Writes testdata/graphbench/results.tsv, verdict-changes.tsv and
# versions.txt. Fewer edges or fewer REACHABLE results are not by
# themselves an improvement: every verdict change is explained by hand in
# testdata/graphbench/RESULTS.MD.
#
# Usage: scripts/graph-benchmark.sh [--repeat N] [--cache-dir DIR] [--out DIR]
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
repeat=2
cache="$root/.cache"
out="$root/testdata/graphbench"
while [ $# -gt 0 ]; do
    case "$1" in
        --repeat) repeat="$2"; shift 2 ;;
        --cache-dir) cache="$2"; shift 2 ;;
        --out) out="$2"; shift 2 ;;
        *) echo "unknown argument $1" >&2; exit 2 ;;
    esac
done
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
mkdir -p "$out"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
python3 -I - "$root" "$cache" "$out" "$work" "$repeat" <<'PY'
import json, os, re, subprocess, sys, time

root, cache, out, work, repeat = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4], int(sys.argv[5])
apps = {
    "petclinic": ("work/spring-petclinic", "work/spring-petclinic/target/dependency"),
    "webgoat": ("work/WebGoat", "work/WebGoat/app-dependency"),
}

def run(app, roots, mode, rep, graph):
    w, deps = apps[app]
    base = os.path.join(work, f"{app}-{roots}-{mode}-{rep}")
    args = [os.path.join(root, "bin/ariadne"), "-offline", "-cache-dir", cache,
            "-app", f"{w}/app-classes.jar", "-deps", deps, "-deps-list", f"{w}/deps.txt",
            "-profile", "spring", "-roots", roots, "-callgraph-mode", mode,
            "-out", base + ".md", "-json", base + ".json", "-json-format", "v2"]
    if graph and mode == "cha":
        args += ["-graph-out", os.path.join(work, f"{app}-{roots}-cha.graph")]
    if graph and mode != "cha":
        args += ["-graph-compare", os.path.join(work, f"{app}-{roots}-cha.graph")]
    t0 = time.monotonic()
    with open(base + ".log", "w") as log:
        p = subprocess.Popen(args, stdout=log, stderr=subprocess.STDOUT, cwd=root)
        _, status, ru = os.wait4(p.pid, 0)
    wall = time.monotonic() - t0
    if status != 0:
        sys.exit(f"{app} {roots} {mode}: exit status {status}, see {base}.log")
    logtext = open(base + ".log").read()
    v2 = json.load(open(base + ".json"))
    cg = v2["callgraph"]
    m = re.search(r"REACHABLE=(\d+) FRAMEWORK=(\d+) NOT_REACHABLE=(\d+) UNKNOWN=(\d+) NOT_APPLICABLE=(\d+)", logtext)
    g = re.search(r"model gaps: (\d+) unresolved dynamic sites, (\d+) reachable", logtext)
    sd = re.search(r"servlet dispatch \((\S+)\): (\d+) registered servlets/filters", logtext)
    me = re.search(r"models: \d+ facts, (\d+) edges added", logtext)
    se = re.search(r"SAM-call edges \(DYNAMIC_LINKAGE L1\): (\d+) added, (\d+) already in the graph, (\d+) left out", logtext)
    sv = re.search(r"SAM calls \(DYNAMIC_LINKAGE L1, (\S+)\)", logtext)
    ct = cg.get("created_types", {})
    rta = cg.get("rta") or {}
    row = {
        "nodes": cg["nodes"], "edges": cg["edges"],
        "cha_edges": rta.get("cha_edges_before", ct.get("cha_edges")),
        "cha_removed": rta.get("cha_edges_removed", 0),
        "cha_kept_by_fallback": rta.get("cha_edges_kept_by_fallback", 0),
        "reachable_methods": rta.get("reachable_methods", ""),
        "live_classes": rta.get("live_classes", ""),
        "live_only_by_fallback": rta.get("live_only_by_fallback", ""),
        "active_alloc_gaps": ",".join(f"{k}={v}" for k, v in sorted((rta.get("active_gaps") or {}).items())),
        "catalog_types": ct.get("types"), "catalog_gaps": sum((ct.get("gaps_by_kind") or {}).values()),
        "model_gap_sites": g.group(1), "model_gap_sites_reachable": g.group(2),
        "servlet_registrations": sd.group(2), "model_edges": me.group(1),
        "sam_edges": se.group(1), "sam_edges_gated": se.group(3),
        "R/F/NR/U/NA": "/".join(m.groups()),
    }
    changes = []
    comp = cg.get("comparison")
    if comp:
        row["edges_removed_vs_cha"] = sum((comp.get("removed_by_basis") or {}).values())
        row["edges_added_vs_cha"] = sum((comp.get("added_by_basis") or {}).values())
        fc = comp.get("findings_changed") or []
        row["findings_changed_vs_cha"] = len(fc)
        row["verdicts_changed_vs_cha"] = sum(1 for f in fc if f["old_status"] != f["new_status"])
        for f in fc:
            if f["old_status"] != f["new_status"]:
                changes.append((app, roots, mode, f["id"], f["old_status"], f["new_status"], " | ".join(f.get("because") or [])))
    versions = {"tool": v2["tool"]["version"], "builder": cg["builder"], "skip_rule": cg["skip_rule"],
                "unknown_allocation_policy": cg["unknown_allocation_policy"], "catalog": ct.get("catalog"),
                "rta": rta.get("version", ""), "servlet_dispatch": sd.group(1), "samcall": sv.group(1)}
    return wall, ru.ru_maxrss / 1024.0, row, changes, versions

rows, allchanges, versions = [], [], {}
for app in apps:
    for roots in ("profile", "all"):
        for mode in ("cha", "rta"):
            walls, rss, ref = [], [], None
            for rep in range(repeat):
                w, r, row, _, ver = run(app, roots, mode, rep, False)
                walls.append(w); rss.append(r)
                if ref is None:
                    ref = row
                elif row != ref:
                    sys.exit(f"{app} {roots} {mode}: run {rep} differs from run 0: {row} vs {ref}")
                versions[mode] = ver
            _, _, grow, ch, _ = run(app, roots, mode, repeat, True)
            for k, v in grow.items():
                if k in ref and ref[k] != v:
                    sys.exit(f"{app} {roots} {mode}: the graph-file run differs in {k}: {v} vs {ref[k]}")
                ref[k] = v
            ref = dict(app=app, roots=roots, mode=mode,
                       wall_s=f"{min(walls):.1f}-{max(walls):.1f}", peak_rss_mb=f"{min(rss):.0f}-{max(rss):.0f}", **ref)
            rows.append(ref); allchanges += ch
            print(app, roots, mode, ref["wall_s"], "s", ref["peak_rss_mb"], "MB", ref["R/F/NR/U/NA"], file=sys.stderr)

cols = []
for r in rows:
    for k in r:
        if k not in cols:
            cols.append(k)
with open(os.path.join(out, "results.tsv"), "w") as f:
    f.write("\t".join(cols) + "\n")
    for r in rows:
        f.write("\t".join(str(r.get(c, "")) for c in cols) + "\n")
with open(os.path.join(out, "verdict-changes.tsv"), "w") as f:
    f.write("app\troots\tmode\tadvisory\tcha\tother\texplained_by_edges\n")
    for c in sorted(allchanges):
        f.write("\t".join(c) + "\n")
commit = subprocess.run(["git", "-C", root, "rev-parse", "HEAD"], capture_output=True, text=True).stdout.strip()
dirty = subprocess.run(["git", "-C", root, "status", "--porcelain", "--untracked-files=no"], capture_output=True, text=True).stdout.strip()
goversion = subprocess.run(["go", "version"], capture_output=True, text=True).stdout.strip()
with open(os.path.join(out, "versions.txt"), "w") as f:
    f.write(f"commit\t{commit}{' (dirty)' if dirty else ''}\n{goversion}\nrepeat\t{repeat}\nprofile\tspring\noffline\ttrue\n")
    for mode, v in sorted(versions.items()):
        for k, val in v.items():
            f.write(f"{mode}.{k}\t{val}\n")
PY
cat "$out/results.tsv"
