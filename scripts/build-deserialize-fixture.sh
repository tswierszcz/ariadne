#!/usr/bin/env bash
# Compiles testdata/deserialize/src into two jars for
# internal/deserialize's policy tests (DESERIALIZATION Z2):
#   fx.jar          the cases (one entry per policy) and candidate classes
#   xstream-stub.jar a signature-only stub of XStream's configuration API
# The jars are committed so `go test` never needs a JDK. Uses
# scripts/JavaBuild.java (java with the jdk.compiler module).
#
# Usage: scripts/build-deserialize-fixture.sh
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fx="$root/testdata/deserialize"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/stub" "$tmp/fx"
cp -r "$fx/src/com" "$tmp/stub/"
cp -r "$fx/src/fx" "$tmp/fx/"
java "$root/scripts/JavaBuild.java" "$tmp/stub" "$fx/xstream-stub.jar" 11
java "$root/scripts/JavaBuild.java" "$tmp/fx" "$fx/fx.jar" 11 "$fx/xstream-stub.jar"
echo "wrote $fx/{fx,xstream-stub}.jar"
