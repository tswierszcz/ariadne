#!/usr/bin/env bash
# RUNTIME_NATIVE N1 end-to-end check on a real application: runs the
# prepared petclinic (scripts/prepare-petclinic.sh) on the real JVM under
# Java Flight Recorder — stack sampling, class loading, and exact entries
# (jdk.MethodTrace, JDK 25+) of ServletWebRequest#validateIfNoneMatch
# (CVE-2024-38809) and DispatcherServlet#doDispatch —, sends a fixed
# scenario of requests (pages, then static resources with If-None-Match /
# If-Modified-Since), stops it, converts the recording with
# scripts/JfrToTrace.java and runs bin/ariadne -offline -runtime-trace on it.
# Prints the report's "Runtime evidence" section. Nothing is committed: the
# trace depends on this machine's run (timing and sampling).
#
# Usage: scripts/runtime-trace-petclinic.sh [OUT_DIR] [PORT]   (java 25+ with jdk.jfr and jdk.compiler on PATH)
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="${1:-$(mktemp -d)}"
port="${2:-18080}"
w="$root/work/spring-petclinic"
mkdir -p "$out"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
filters='org.springframework.web.context.request.ServletWebRequest::validateIfNoneMatch;org.springframework.web.servlet.DispatcherServlet::doDispatch'
java "-XX:StartFlightRecording=filename=$out/petclinic.jfr,settings=profile,+jdk.ClassLoad#enabled=true,jdk.MethodTrace#enabled=true,jdk.MethodTrace#filter=$filters" \
    -cp "$w/app-classes.jar:$w/target/dependency/*" org.springframework.samples.petclinic.PetClinicApplication --server.port="$port" \
    >"$out/app.log" 2>&1 &
pid=$!
trap 'kill "$pid" 2>/dev/null || true' EXIT
b="http://localhost:$port"
for _ in $(seq 1 90); do
    curl -s -o /dev/null "$b/" && break
    sleep 2
done
for p in / /owners/find "/owners?lastName=" /owners/1 /owners/2/pets/2/visits/new /vets.html /vets /oups; do
    curl -s -o /dev/null "$b$p"
done
for _ in $(seq 1 200); do
    for r in /resources/css/petclinic.css /resources/images/pets.png /webjars/jquery/2.2.4/jquery.min.js; do
        curl -s -o /dev/null -H 'If-None-Match: "abc", W/"def"' -H 'If-Modified-Since: Sat, 01 Jan 2000 00:00:00 GMT' "$b$r"
    done
    curl -s -o /dev/null "$b/owners/1"
done
kill -TERM "$pid"
wait "$pid" 2>/dev/null || true
trap - EXIT
java scripts/JfrToTrace.java --jfr "$out/petclinic.jfr" --out "$out/petclinic.trace" --run-id petclinic \
    --scenario "PetClinicApplication: 8 pages once, then 200 x (3 static resources with If-None-Match and If-Modified-Since, /owners/1)" \
    --artifact "$w/app-classes.jar" --artifact-dir "$w/target/dependency" \
    --method-filter 'org.springframework.web.context.request.ServletWebRequest::validateIfNoneMatch' \
    --method-filter 'org.springframework.web.servlet.DispatcherServlet::doDispatch'
bin/ariadne -offline -app "$w/app-classes.jar" -deps "$w/target/dependency" -deps-list "$w/deps.txt" -profile spring \
    -runtime-trace "$out/petclinic.trace" -out "$out/report.md" -json "$out/results.json" -json-format v2 2>&1 | grep -E "runtime trace|REACHABLE="
sed -n '/^## Runtime evidence/,/^## Withdrawn/p' "$out/report.md"
