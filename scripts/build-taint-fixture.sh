#!/usr/bin/env bash
# Compiles testdata/taint/src into testdata/taint/app.jar for
# internal/taint's tests (DATA_FLOW D4). The sources are the runnable
# control app of scripts/prepare-taint-demo.sh plus demo.Extra's edge
# cases (a lambda caller, a field, a depth limit, two callers of one sink
# helper). The jar is committed so `go test` never needs a JDK. Uses
# scripts/JavaBuild.java (java with the jdk.compiler module).
#
# Usage: scripts/build-taint-fixture.sh
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fx="$root/testdata/taint"
java "$root/scripts/JavaBuild.java" "$fx/src" "$fx/app.jar" 17
echo "wrote $fx/app.jar"
