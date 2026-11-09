#!/usr/bin/env bash
# RUNTIME_NATIVE N1 end-to-end check on JPetStore 6 (scripts/prepare-jpetstore.sh),
# the same steps as scripts/runtime-trace-petclinic.sh: analyzes the WAR once
# to write JFR settings for exact entries of every mapped vulnerable method
# (-runtime-filter-out), runs the WAR in Jetty (jetty-runner, javax.servlet;
# fetched into OUT_DIR by Maven) on the real JVM under Java Flight Recorder,
# sends a fixed scenario of requests (catalog browsing, search, cart, sign-on
# with the application's own seeded demo account), stops it, converts the
# recording with scripts/JfrToTrace.java and runs bin/ariadne -offline
# -runtime-trace on it. Prints the report's "Runtime evidence" section.
# Nothing is committed: the trace depends on this machine's run.
#
# Usage: scripts/runtime-trace-jpetstore.sh [OUT_DIR] [PORT]   (java 25+ with jdk.jfr and jdk.compiler on PATH)
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="${1:-$(mktemp -d)}"
port="${2:-18081}"
w="$root/work/jpetstore"
jetty_version="9.4.57.v20241219"
mkdir -p "$out"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
analyze() {
    bin/ariadne -offline -app "$w/target/jpetstore.war" -deps-list "$w/deps.txt" -roots profile "$@"
}
analyze -runtime-filter-out "$out/vulnerable.jfc" -out "$out/static.md" -json "$out/static.json" 2>&1 | grep -E "runtime filter"
runner="$out/jetty-runner-$jetty_version.jar"
[ -f "$runner" ] || mvn -q -B dependency:copy -Dartifact="org.eclipse.jetty:jetty-runner:$jetty_version" -DoutputDirectory="$out"
# Spring 4.3's CGLIB defines classes through ClassLoader#defineClass, which
# JDK 16+ only allows with java.lang opened to the class path.
java --add-opens java.base/java.lang=ALL-UNNAMED "-XX:StartFlightRecording=filename=$out/jpetstore.jfr,settings=profile,settings=$out/vulnerable.jfc,+jdk.ClassLoad#enabled=true" \
    -jar "$runner" --port "$port" --path /jpetstore "$w/target/jpetstore.war" >"$out/app.log" 2>&1 &
pid=$!
trap 'kill "$pid" 2>/dev/null || true' EXIT
b="http://localhost:$port/jpetstore"
for _ in $(seq 1 90); do
    curl -s -o /dev/null "$b/" && break
    sleep 2
done
jar="$out/cookies.txt"
get() { curl -s -o /dev/null -b "$jar" -c "$jar" "$b$1"; }
a="/actions"
for _ in $(seq 1 50); do
    get "/"
    get "$a/Catalog.action"
    for c in FISH DOGS CATS REPTILES BIRDS; do get "$a/Catalog.action?viewCategory=&categoryId=$c"; done
    get "$a/Catalog.action?viewProduct=&productId=FI-SW-01"
    get "$a/Catalog.action?viewItem=&itemId=EST-1"
    get "$a/Catalog.action?searchProducts=&keyword=fish"
    get "$a/Cart.action?addItemToCart=&workingItemId=EST-1"
    get "$a/Cart.action?viewCart="
    get "$a/Account.action?signonForm="
    curl -s -o /dev/null -b "$jar" -c "$jar" -d "username=j2ee&password=j2ee&signon=Login" "$b$a/Account.action"
    get "$a/Order.action?newOrderForm="
    get "$a/Account.action?signoff="
done
grep -q "Context initialization failed" "$out/app.log" && { echo "JPetStore failed to start, see $out/app.log" >&2; exit 1; }
kill -TERM "$pid"
wait "$pid" 2>/dev/null || true
trap - EXIT
java scripts/JfrToTrace.java --jfr "$out/jpetstore.jfr" --out "$out/jpetstore.trace" --run-id jpetstore \
    --scenario "JPetStore 6.0.2 in Jetty $jetty_version: 50 x (catalog, 5 categories, product, item, search, add to cart, cart, sign-on form, sign-on, new order form, sign-off)" \
    --artifact "$w/app-classes.jar" --artifact-dir "$w/target/dependency"
analyze -runtime-trace "$out/jpetstore.trace" -out "$out/report.md" -json "$out/results.json" -json-format v2 2>&1 | grep -E "runtime trace|REACHABLE="
sed -n '/^## Runtime evidence/,/^## Withdrawn/p' "$out/report.md"
