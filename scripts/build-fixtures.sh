#!/usr/bin/env bash
# Compiles testdata/fixture/src for Java 8 and Java 21 bytecode and packages
# each into a jar under testdata/fixture/. The resulting .class/.jar files are
# committed to the repo so that `go test` never needs a JDK.
#
# As of JVM_GRAPH G5, testdata/fixture/src/fixture/ also has Box.java,
# StringBox.java and BridgeDemo.java (a generic interface forcing a
# synthetic bridge method) and a default method on Greeter.java — added
# to exercise JVM_GRAPH G2's bridge-method/default-method resolution
# against real compiled bytecode, but not yet compiled into the
# committed fixture-8.jar/fixture-21.jar: this repo's own working
# environment has no javac. Run this script (with a real javac8/javac21)
# to pick them up, then see BridgeDemo.java's own doc comment for the
# cross-check this should let a future session add.
#
# Usage: scripts/build-fixtures.sh [javac8] [javac21]
set -euo pipefail

JAVAC8="${1:-javac}"
JAVAC21="${2:-javac}"

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture="$root/testdata/fixture"
src="$fixture/src"

build() {
    local release="$1"
    local javac_bin="$2"
    local outdir="$fixture/classes-$release"
    local jar="$fixture/fixture-$release.jar"
    rm -rf "$outdir"
    mkdir -p "$outdir"
    echo "compiling fixture for --release $release with $javac_bin"
    local jar_bin="$(dirname "$javac_bin")/jar"
    "$javac_bin" --release "$release" -d "$outdir" "$src"/fixture/*.java
    (cd "$outdir" && "$jar_bin" cf "$jar" .)
    rm -rf "$outdir"
    echo "wrote $jar"
}

build 8 "$JAVAC8"
build 21 "$JAVAC21"
