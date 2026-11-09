#!/usr/bin/env bash
# DYNAMIC_LINKAGE L1 (SAM-call materialization) oracle. Compiles
# testdata/samcall/src with the real javac into testdata/samcall/sc.jar and
# runs sc.Oracle on the real JVM: every lambda body and method-reference
# target records itself and the nearest fixture frame below it (hidden
# lambda classes and JDK frames skipped, as a runtime trace attributes the
# call). Writes testdata/samcall/observed.tsv (caller, bytecode index,
# body). The jar and the TSV are committed; internal/samcall TestSAMFixture
# checks the call graph against them without a JDK.
#
# Usage: scripts/samcall-oracle.sh [JAVA_HOME]   (JDK 17+ with jdk.compiler)
set -euo pipefail
JAVA_HOME_ARG="${1:-${JAVA_HOME:-}}"
if [ -n "$JAVA_HOME_ARG" ]; then
    export JAVA_HOME="$JAVA_HOME_ARG"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fx="$root/testdata/samcall"
java "$root/scripts/JavaBuild.java" "$fx/src" "$fx/sc.jar" 17
java -cp "$fx/sc.jar" sc.Oracle > "$fx/observed.tsv"
cat "$fx/observed.tsv"
