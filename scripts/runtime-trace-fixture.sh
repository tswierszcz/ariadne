#!/usr/bin/env bash
# RUNTIME_NATIVE N0/N1 fixture. Compiles testdata/runtime/src into
# testdata/runtime/rtm.jar (and testdata/runtime/v2 — the same code with one
# method body changed — into rtm-v2.jar), runs rtm.Main on the real JVM under
# Java Flight Recorder (stack sampling, class loading, and jdk.MethodTrace
# exact entries of rtm.Lib::parse — JDK 25+, JEP 520), and converts the
# recording with scripts/JfrToTrace.java into testdata/runtime/run.trace.
# The jars and the trace are committed; cmd/ariadne TestRuntimeTrace imports
# the trace without a JDK. Sampling is periodic: a rerun produces different
# counts, not different conclusions.
#
# Usage: scripts/runtime-trace-fixture.sh [JAVA_HOME]  (JDK/JRE 25+ with jdk.compiler and jdk.jfr)
set -euo pipefail
JAVA_HOME_ARG="${1:-${JAVA_HOME:-}}"
if [ -n "$JAVA_HOME_ARG" ]; then
    export JAVA_HOME="$JAVA_HOME_ARG"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fx="$root/testdata/runtime"
java "$root/scripts/JavaBuild.java" "$fx/src" "$fx/rtm.jar" 17
java "$root/scripts/JavaBuild.java" "$fx/v2" "$fx/rtm-v2.jar" 17
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
java "-XX:StartFlightRecording=filename=$tmp/run.jfr,settings=profile,+jdk.ClassLoad#enabled=true,jdk.MethodTrace#enabled=true,jdk.MethodTrace#filter=rtm.Lib::parse" \
    -cp "$fx/rtm.jar" rtm.Main >/dev/null
java "$root/scripts/JfrToTrace.java" --jfr "$tmp/run.jfr" --out "$fx/run.trace" \
    --scenario "rtm.Main: shapes 800 ms, 1000 Lib.parse calls, LoadedOnly loaded, Plugin.work by reflection under the app loader and an isolated loader (600 ms each)" \
    --run-id fixture-1 --artifact "$fx/rtm.jar" --method-filter 'rtm.Lib::parse'
grep -c . "$fx/run.trace"
