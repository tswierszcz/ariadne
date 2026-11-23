#!/usr/bin/env bash
# RUNTIME_NATIVE N5 n5-1: builds testdata/native/app.jar from
# testdata/native/src (compiled only; there is no native library and the
# program is never run). The jar is committed; TestNativeBoundaries needs
# no JDK.
#
# Usage: scripts/native-fixture.sh [JAVA_HOME]
set -euo pipefail
JAVA_HOME_ARG="${1:-${JAVA_HOME:-}}"
if [ -n "$JAVA_HOME_ARG" ]; then
    export JAVA_HOME="$JAVA_HOME_ARG"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
java "$root/scripts/JavaBuild.java" "$root/testdata/native/src" "$root/testdata/native/app.jar" 17
