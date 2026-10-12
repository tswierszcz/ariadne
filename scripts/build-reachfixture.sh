#!/usr/bin/env bash
# Compiles testdata/reachfixture/src and packages each class into its own
# jar under testdata/reachfixture/, so internal/reach's tests can treat
# them as three separate inputs (one app, two dependencies) without
# hierarchy.Build's class-dedup logic merging them. The resulting jars
# are committed so `go test` never needs a JDK.
#
# Usage: scripts/build-reachfixture.sh [javac]
set -euo pipefail

JAVAC="${1:-javac}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture="$root/testdata/reachfixture"
src="$fixture/src"
jar_bin="$(dirname "$JAVAC")/jar"

outdir="$fixture/classes"
rm -rf "$outdir"
mkdir -p "$outdir"
"$JAVAC" --release 8 -d "$outdir" "$src"/reachfixture/*.java

(cd "$outdir" && "$jar_bin" cf "$fixture/appentry.jar" "reachfixture/AppEntry.class")
(cd "$outdir" && "$jar_bin" cf "$fixture/frameworkthing.jar" "reachfixture/FrameworkThing.class")
(cd "$outdir" && "$jar_bin" cf "$fixture/vulnerablelib.jar" "reachfixture/VulnerableLib.class" "reachfixture/VulnerableLibSub.class")
echo "wrote $fixture/{appentry,frameworkthing,vulnerablelib}.jar"

rm -rf "$outdir"
