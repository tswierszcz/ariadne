#!/usr/bin/env bash
# Runtime evidence for Lavagna 1.1.9 (scripts/prepare-lavagna.sh), the steps
# of scripts/runtime-trace-jpetstore.sh: analyzes the input once to write JFR
# settings for exact entries of every mapped vulnerable method
# (-runtime-filter-out), assembles the exploded webapp (src/main/webapp,
# WEB-INF/classes from app-classes.jar, WEB-INF/lib the runtime
# dependencies), runs it in the provided Jetty 9.4.44
# (scripts/JettyExploded.java) on the real JVM under Java Flight Recorder in
# the dev profile (HSQLDB in memory, the dev seed's demo users), sends a fixed
# HTTP scenario through the REST API (demo login as "user", project, board,
# columns, cards with markdown descriptions, comments, action lists, moves,
# search, calendar feed, export), stops it, converts the recording with
# scripts/JfrToTrace.java and runs bin/ariadne -offline -runtime-trace on it
# (CHA, -roots all). Prints the runtime summary and every finding whose
# vulnerable method ran. Nothing is committed: the trace depends on this
# machine's run.
#
# Usage: scripts/runtime-trace-lavagna.sh [OUT_DIR] [PORT]   (java 25+ with jdk.jfr and jdk.compiler on PATH)
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="${1:-$(mktemp -d)}"
port="${2:-18082}"
w="$root/work/lavagna"
mkdir -p "$out"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
analyze() {
    bin/ariadne -offline -app "$w/app-classes.jar" -deps "$w/app-dependency" -deps-list "$w/deps.txt" -profile spring -cache-dir "$w/cache" "$@"
}
analyze -runtime-filter-out "$out/vulnerable.jfc" -out "$out/static.md" -json "$out/static.json" 2>&1 | grep -E "runtime filter"

# The exploded webapp; Jetty and the servlet API (scope provided) stay on
# the launcher's class path.
web="$out/webapp"
rm -rf "$web"
cp -r "$w/src/src/main/webapp" "$web"
mkdir -p "$web/WEB-INF/classes" "$web/WEB-INF/lib"
python3 -I - "$w" "$web" "$out/jetty-cp.txt" <<'PY'
import pathlib, re, shutil, sys, zipfile
w, web, cp = map(pathlib.Path, sys.argv[1:])
zipfile.ZipFile(w / 'app-classes.jar').extractall(web / 'WEB-INF/classes')
def jars(listing, scopes):
    for line in open(w / listing):
        m = re.match(r'\s+[^:\s]+:([^:]+):jar:(?:[^:]+:)?([^:]+):(\w+)', line)
        if m and m[3] in scopes:
            yield next((w / 'app-dependency').glob(f'{m[1]}-{m[2]}*.jar'))
for j in jars('deps-runtime.txt', ('compile', 'runtime')):
    shutil.copy(j, web / 'WEB-INF/lib')
cp.write_text(':'.join(str(j) for j in jars('deps-provided.txt', ('provided',))))
PY

java --add-opens java.base/java.lang=ALL-UNNAMED \
    "-XX:StartFlightRecording=filename=$out/lavagna.jfr,settings=profile,settings=$out/vulnerable.jfc,+jdk.ClassLoad#enabled=true" \
    -Ddatasource.dialect=HSQLDB -Ddatasource.url=jdbc:hsqldb:mem:lavagna -Ddatasource.username=sa -Ddatasource.password= \
    -Dspring.profiles.active=dev -cp "$(cat "$out/jetty-cp.txt")" scripts/JettyExploded.java "$web" "$port" >"$out/app.log" 2>&1 &
pid=$!
trap 'kill "$pid" 2>/dev/null || true' EXIT
for _ in $(seq 1 90); do
    grep -qE "Started @|Exception in thread" "$out/app.log" && break
    sleep 2
done
grep -q "Failed startup" "$out/app.log" && { echo "Lavagna failed to start, see $out/app.log" >&2; exit 1; }

# The HTTP scenario.
B="http://127.0.0.1:$port" cj="$out/cookies.txt" rounds=10
curl -s -o /dev/null -c "$cj" -b "$cj" "$B/login"
T=$(curl -s -i -c "$cj" -b "$cj" "$B/login" | grep -i '^x-csrf-token' | tr -d '\r' | awk '{print $2}')
curl -s -o /dev/null -c "$cj" -b "$cj" -d "username=user&password=user&reqUrl=/&_csrf=$T" "$B/login/demo/"
T=$(curl -s -i -c "$cj" -b "$cj" "$B/" | grep -i '^x-csrf-token' | tr -d '\r' | awk '{print $2}')
fail=0
j() { # METHOD PATH [JSON] -> body on stdout; counts non-2xx
    local out code
    out=$(curl -s -c "$cj" -b "$cj" -X "$1" -H "X-CSRF-TOKEN: $T" -H 'Content-Type: application/json' ${3:+--data "$3"} -w '\n%{http_code}' "$B$2")
    code=${out##*$'\n'}
    case $code in 2*) ;; *) fail=$((fail+1)); echo "  $1 $2 -> $code" >&2 ;; esac
    printf '%s' "${out%$'\n'*}" | sed '1s/^)]}.//'
}
num() { python3 -I -c "import json,sys;d=json.loads(sys.stdin.read() or 'null');print($1)"; }
j POST /api/project '{"shortName":"DEMO","name":"Demo project","description":"demo"}' >/dev/null
j POST /api/project/DEMO/board '{"shortName":"DB","name":"Demo board","description":"board"}' >/dev/null
open=$(j GET /api/project/DEMO/definitions | num "[x['id'] for x in d if x['value']=='OPEN'][0]")
j POST /api/board/DB/column "{\"name\":\"Todo\",\"definition\":$open}" >/dev/null
j POST /api/board/DB/column "{\"name\":\"Done\",\"definition\":$open}" >/dev/null
cols=$(j GET /api/board/DB/columns-in/BOARD | num "' '.join(str(c['id']) for c in d if c['name'] in ('Todo','Done'))")
read -r c1 c2 <<<"$cols"
for i in $(seq 1 "$rounds"); do
    card=$(j POST "/api/column/$c1/card" "{\"name\":\"Card $i <b>bold</b> **md**\",\"description\":\"# Title\\n\\nSome *markdown* with <script>alert(1)</script>\"}" >/dev/null; j GET "/api/column/$c1/card" | num "max(c['id'] for c in d)")
    j POST "/api/card/$card/description" '{"content":"# Title\n\nSome *markdown* with <script>alert(1)</script> and [a link](http://example.com)"}' >/dev/null
    j POST "/api/card/$card/comment" "{\"content\":\"comment $i with **bold** and <a href='javascript:x'>x</a>\"}" >/dev/null
    j POST "/api/card/$card/actionlist" '{"content":"checklist"}' >/dev/null
    j GET "/api/card/$card" >/dev/null; j GET "/api/card/$card/data" >/dev/null; j GET "/api/card/$card/activity" >/dev/null
    j GET "/api/card/$card/comments" >/dev/null; j GET "/api/card/$card/description" >/dev/null
    j POST "/api/card/$card/from-column/$c1/to-column/$c2" "{\"newContainer\":[$card]}" >/dev/null
    j POST "/api/card/$card" "{\"name\":\"Card $i renamed\"}" >/dev/null
    j GET "/api/board/DB" >/dev/null; j GET "/api/board/DB/task-statistics" >/dev/null
    j GET "/api/search/card?q=%5B%7B%22type%22%3A%22FREETEXT%22%2C%22value%22%3A%7B%22type%22%3A%22STRING%22%2C%22value%22%3A%22Card%22%7D%7D%5D&page=0" >/dev/null
    j GET "/api/search/autocomplete-card?term=Card" >/dev/null
    j GET "/api/search/user?term=us" >/dev/null
    j GET "/api/project/DEMO" >/dev/null; j GET /api/project >/dev/null
done
tok=$(j GET /api/calendar/token | num "d['token'] if isinstance(d,dict) else d")
j GET "/api/calendar/$tok/calendar.ics" >/dev/null
j GET /api/calendar/user >/dev/null
curl -s -o /dev/null -c "$cj" -b "$cj" -X POST -H "X-CSRF-TOKEN: $T" -w 'export %{http_code}\n' "$B/api/export"
echo "scenario: $rounds rounds, $fail non-2xx responses"

kill -TERM "$pid"
wait "$pid" 2>/dev/null || true
trap - EXIT
java scripts/JfrToTrace.java --jfr "$out/lavagna.jfr" --out "$out/lavagna.trace" --run-id lavagna \
    --scenario "Lavagna 1.1.9 exploded in Jetty 9.4.44 (dev, HSQLDB): demo login, project, board, 2 columns, $rounds x (card with markdown description, description update, comment, action list, card views, move, rename, board, statistics, search, autocomplete, user search, project), calendar token and .ics feed, export" \
    --artifact "$w/app-classes.jar" --artifact-dir "$w/app-dependency"
analyze -runtime-trace "$out/lavagna.trace" -out "$out/report.md" -json "$out/results.json" -json-format v2 2>&1 | grep -E "runtime trace|REACHABLE="
python3 -I - "$out/results.json" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
seen = sorted({(f['reachability'], f['advisory_id']) for f in d['findings'] if f.get('runtime_observed')})
print(f'{len(seen)} findings whose vulnerable method ran:', ', '.join(f'{a} ({r})' for r, a in seen))
bad = [a for r, a in seen if r == 'NOT_REACHABLE']
print('NOT_REACHABLE with an observed vulnerable method:', ', '.join(bad) or 'none')
PY
