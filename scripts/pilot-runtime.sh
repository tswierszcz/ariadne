#!/usr/bin/env bash
# RESEARCH_INTERFACE RI6 on the pilot control target (testdata/pilot, RH-A3):
# for each variant, indexes it (research index -roots api), writes a trace
# configuration for the pilot's targets (research trace-config), runs the
# oracle client under JFR with that filter (JDK 25+, jdk.MethodTrace),
# converts the recording (scripts/JfrToTrace.java) to
# testdata/pilot/runtime-VARIANT.trace and correlates it with the bundle
# (research correlate). The traces are committed; TestPilotRuntime (Go)
# repeats the index, configuration and correlation without a JDK.
#
# Usage: scripts/pilot-runtime.sh [WORK_DIR]
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fx="$root/testdata/pilot"
work="${1:-$(mktemp -d)}"
mkdir -p "$work/empty-cache"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
for v in defective fixed; do
    rm -rf "$work/bundle-$v"
    bin/ariadne research index -app "$fx/$v.jar" -roots api -profile generic -cache-dir "$work/empty-cache" -out "$work/bundle-$v" 2>/dev/null
    python3 - "$work/bundle-$v" "$work/request-$v.json" <<'PY'
import json, sys
bundle, out = sys.argv[1], sys.argv[2]
analysis = json.load(open(bundle + "/manifest.json"))["analysis_id"]
inst = next(json.loads(l)["instance_id"] for l in open(bundle + "/components.jsonl") if '"app:' in l)
P = "Ljava/nio/file/Path;"
def t(i, c, n, d): return {"id": i, "selector": {"instance_id": inst, "class": c, "name": n, "descriptor": d}}
json.dump({"schema_version": "ariadne-research-trace-request-1", "analysis_id": analysis, "request_id": "pilot-runtime",
           "targets": [t("write", "px.archive.internal.Store", "write", "(" + P + "[B)V"),
                       t("target", "px.archive.Entries", "target", "(" + P + "Ljava/lang/String;)" + P),
                       t("dead", "px.archive.internal.Store", "overwrite", "(" + P + "[B)V"),
                       t("filter", "px.archive.Unpacker", "filter", "(Ljava/lang/String;)Ljava/lang/Object;")]},
          open(out, "w"), indent=2)
PY
    bin/ariadne research trace-config -bundle "$work/bundle-$v" -request "$work/request-$v.json" -out "$work/config-$v.json" -jfc "$work/trace-$v.jfc"
    java "-XX:StartFlightRecording=filename=$work/$v.jfr,settings=profile,settings=$work/trace-$v.jfc,+jdk.ClassLoad#enabled=true" \
        -cp "$fx/client.jar:$fx/$v.jar" px.client.Main >/dev/null
    java scripts/JfrToTrace.java --jfr "$work/$v.jfr" --out "$fx/runtime-$v.trace" --run-id "pilot-$v" \
        --scenario "px.client.Main with $v.jar: normal entry, ../escaped.txt, writeFile" --artifact "$fx/$v.jar" --artifact "$fx/client.jar"
    bin/ariadne research correlate -bundle "$work/bundle-$v" -trace "$fx/runtime-$v.trace" -config "$work/config-$v.json" -out "$work/correlation-$v.json"
done
echo "work files in $work"
