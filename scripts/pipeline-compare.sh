#!/usr/bin/env bash
# RESEARCH_INTERFACE RI1: the shared static pipeline must not change the CVE
# scan. Builds two analyzer revisions and runs both, offline on the same
# cache, on petclinic, WebGoat, JPetStore and the library-mode control
# library, writing results.json (v2), report.md and the -graph-out file,
# then compares them. Durations ("in 12ms") are masked in report.md and the
# log, the analyzer version and the analysis id / results hash (which cover it) in the
# results, report and log, and the log is compared as a sorted set of lines; everything else
# must be byte-identical. Prints one line per run and
# the first lines of every difference; exit status 1 when anything differs.
#
# ONLY=GLOB limits the runs by name (e.g. ONLY='library-*').
#
# Usage: scripts/pipeline-compare.sh OLD_COMMIT [NEW_COMMIT|working-tree] [--cache-dir DIR] [--keep DIR]
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
old="$1"; shift
new="working-tree"
if [ $# -gt 0 ] && [ "${1#--}" = "$1" ]; then new="$1"; shift; fi
cache="$root/.cache"
keep=""
while [ $# -gt 0 ]; do
    case "$1" in
        --cache-dir) cache="$2"; shift 2 ;;
        --keep) keep="$2"; shift 2 ;;
        *) echo "unknown argument $1" >&2; exit 2 ;;
    esac
done
tmp="$(mktemp -d)"
cleanup() {
    for wt in "$tmp"/wt-*; do
        [ -d "$wt" ] && git -C "$root" worktree remove --force "$wt" >/dev/null 2>&1
    done
    if [ -n "$keep" ]; then mkdir -p "$keep" && cp -r "$tmp"/out "$keep"/; fi
    rm -rf "$tmp"
}
trap cleanup EXIT
build() { # commit|working-tree out
    if [ "$1" = working-tree ]; then (cd "$root" && go build -o "$2" ./cmd/ariadne); return; fi
    local wt="$tmp/wt-$(basename "$2")"
    git -C "$root" worktree add -q --detach "$wt" "$1"
    (cd "$wt" && go build -o "$2" ./cmd/ariadne)
}
build "$old" "$tmp/old-bin"
build "$new" "$tmp/new-bin"
cd "$root"
mask() { sed -E '/^- \*\*Generated:\*\*/d; /^- \*\*ariadne version:\*\*/d; /^- \*\*Analysis ID:\*\*/d; s/"analysis_id": "sha256:[0-9a-f]{64}"/"analysis_id": <id>/; s/^analysis sha256:[0-9a-f]{64}/analysis <id>/; s/"version": "[^"]*"/"version": <v>/; s#/tmp/ariadne-[a-z-]+-[0-9]+/#<tmpdir>/#g; s/[0-9]+(\.[0-9]+)?(ns|µs|us|ms|s|m[0-9.]+s)\b/<t>/g; s/[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9:.]+(Z|[+-][0-9:]+)/<ts>/g' "$1"; }
status=0
run() { # name cache-dir args...
    local name=$1 c=$2; shift 2
    case "$name" in ${ONLY:-*}) ;; *) return ;; esac
    mkdir -p "$tmp/out/$name"
    # Both sides run at once, each into its own directory under the same
    # file names (report.md names its results file).
    local pids=()
    for side in old new; do
        mkdir -p "$tmp/out/$name/$side"
        ( cd "$tmp/out/$name/$side" && "$tmp/$side-bin" -offline -cache-dir "$c" "${@/#@root@/$root}" -json-format v2 -json results.json -out report.md -graph-out graph.txt > run.log 2>&1 ) &
        pids+=($!)
    done
    for i in 0 1; do
        wait "${pids[$i]}" || { echo "$name: run $i failed"; tail -3 "$tmp/out/$name/$([ $i = 0 ] && echo old || echo new)/run.log"; status=1; return; }
    done
    local o="$tmp/out/$name" same=yes
    cmp -s "$o/old/graph.txt" "$o/new/graph.txt" || { same=no; echo "--- $name graph.txt differs:"; diff "$o/old/graph.txt" "$o/new/graph.txt" | head -20 || true; }
    # results.json: the analyzer version and the ids hashed over it differ
    # between two builds by design; everything else must not.
    diff -q <(mask "$o/old/results.json") <(mask "$o/new/results.json") >/dev/null || { same=no; echo "--- $name results.json differs (version and hashes masked):"; diff <(mask "$o/old/results.json") <(mask "$o/new/results.json") | head -20 || true; }
    diff -q <(mask "$o/old/report.md") <(mask "$o/new/report.md") >/dev/null || { same=no; echo "--- $name report.md differs (durations masked):"; diff <(mask "$o/old/report.md") <(mask "$o/new/report.md") | head -20 || true; }
    # The log is compared as a set of lines: RI1 moved the CVE stage after
    # the static pipeline, so its lines come later; every line must remain.
    diff -q <(mask "$o/old/run.log" | LC_ALL=C sort) <(mask "$o/new/run.log" | LC_ALL=C sort) >/dev/null || { same=no; echo "--- $name log lines differ (durations masked, sorted):"; diff <(mask "$o/old/run.log" | LC_ALL=C sort) <(mask "$o/new/run.log" | LC_ALL=C sort) | head -20 || true; }
    echo "$name: identical=$same ($(grep -E '^REACHABLE=' "$o/new/run.log" || true))"
    [ $same = yes ] || status=1
}
pc=(-app @root@/work/spring-petclinic/app-classes.jar -deps @root@/work/spring-petclinic/target/dependency -deps-list @root@/work/spring-petclinic/deps.txt -profile spring)
wg=(-app @root@/work/WebGoat/app-classes.jar -deps @root@/work/WebGoat/app-dependency -deps-list @root@/work/WebGoat/deps.txt -profile spring)
jp=(-app @root@/work/jpetstore/target/jpetstore.war -deps-list @root@/work/jpetstore/deps.txt -profile spring)
lib=(-app @root@/testdata/library/lib.jar -deps @root@/testdata/library/deps -profile generic -roots api)
run petclinic-all "$cache" "${pc[@]}"
run petclinic-profile "$cache" "${pc[@]}" -roots profile
run petclinic-rta "$cache" "${pc[@]}" -roots profile -callgraph-mode rta
run webgoat-all "$cache" "${wg[@]}"
run jpetstore-all "$cache" "${jp[@]}"
run library-api "$root/testdata/library/cache" "${lib[@]}"
run library-api-pta "$root/testdata/library/cache" "${lib[@]}" -callgraph-mode pta
exit $status
