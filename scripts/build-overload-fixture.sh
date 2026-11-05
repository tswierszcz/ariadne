#!/usr/bin/env bash
# Builds testdata/overloadfixture/{lib,app}.jar (CVE:F4 exact-selector and
# versioned-mapping tests) with scripts/JavaBuild.java — a Java runtime
# with jdk.compiler is enough, no javac binary needed.
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fx="$root/testdata/overloadfixture"
java "$root/scripts/JavaBuild.java" "$fx/src/lib" "$fx/lib.jar" 17
java "$root/scripts/JavaBuild.java" "$fx/src/app" "$fx/app.jar" 17 "$fx/lib.jar"
echo "built $fx/lib.jar and $fx/app.jar"
