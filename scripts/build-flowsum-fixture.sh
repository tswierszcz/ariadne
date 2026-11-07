#!/usr/bin/env bash
# Compiles testdata/flowsum/src into testdata/flowsum/fs.jar for
# internal/flowsum's tests (DATA_FLOW D3 dependency summaries). ext.Lib is
# compiled against but not jarred: it stands for a method outside the
# analyzed input. The jar is committed so `go test` never needs a JDK.
# Uses scripts/JavaBuild.java (java with the jdk.compiler module).
#
# Usage: scripts/build-flowsum-fixture.sh
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fx="$root/testdata/flowsum"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
java "$root/scripts/JavaBuild.java" "$fx/ext-src" "$tmp/ext.jar" 11
java "$root/scripts/JavaBuild.java" "$fx/src" "$fx/fs.jar" 11 "$tmp/ext.jar"
echo "wrote $fx/fs.jar"
