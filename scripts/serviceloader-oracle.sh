#!/usr/bin/env bash
# SERVICE_LOADING S4 / ROADMAP B4: a real-JVM oracle for the ServiceLoader
# model. Runs fixture entry methods (testdata/serviceloader and
# testdata/jpms) on the real JVM in the class-path, module-path and mixed
# scenarios, records which providers the JVM actually instantiated, and
# writes testdata/oracle/expected.tsv:
#   <scenario>\t<Class#method>\t<sorted providers or ->[\tERROR <exception>]
# internal/serviceload/oracle_test.go compares ariadne's model edges with
# it; ordinary `go test` reads the checked-in file and needs no JDK.
#
# Usage: scripts/serviceloader-oracle.sh   (needs `java` 11+ on PATH; the
# fixtures' own classes are expected to be compiled already, see the
# READMEs under testdata/serviceloader and testdata/jpms)
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root/testdata"
J="java --module jdk.compiler/com.sun.tools.javac.Main"
rm -rf oracle/classes
$J -source 8 -target 8 -nowarn -d oracle/classes oracle/src/oracle/OracleRunner.java 2>&1 | grep -v warning || true

# On the module path an exploded directory without module-info is not a
# module at all, so the automatic module is packed as a jar with the same
# Automatic-Module-Name the Go tests use (internal/serviceload's
# buildJPMSJars).
python3 - <<'PY'
import os, zipfile
src = "jpms/classes/auto"
with zipfile.ZipFile("oracle/classes/fixture-auto-1.2.3.jar", "w") as z:
    z.writestr("META-INF/MANIFEST.MF", "Manifest-Version: 1.0\r\nAutomatic-Module-Name: fixture.jpms.automatic\r\n\r\n")
    for d, _, files in os.walk(src):
        for f in files:
            p = os.path.join(d, f)
            z.write(p, os.path.relpath(p, src))
PY

S=serviceloader/classes
M=jpms/classes/mods
A=jpms/classes/auto

SL_ENTRIES=(run runThroughParameter runThroughStream streamWithoutConsuming
  runThroughStreamCallback runThroughStaticField runThroughFinalInstanceField
  runThroughCyclicField runThroughMutableField loadWithoutConsuming runWithUnknownLoader)
JPMS_ENTRIES=(Main#main Layers#boot Loaders#context Loaders#system Loaders#nullLoader
  Loaders#defining Params#viaCodec Params#viaRunnable Params#modules Params#recEntry
  Captures#jackson Captures#boxes Captures#mutable Privileged#anon Privileged#lambda
  Privileged#exception Regression#twoLoaders)

specs_sl=(); for e in "${SL_ENTRIES[@]}"; do specs_sl+=("fixture.app.Entry#$e"); done
specs_jpms=(); for e in "${JPMS_ENTRIES[@]}"; do specs_jpms+=("fixture.jpms.app.$e"); done

opens="--add-opens fixture.jpms.app/fixture.jpms.app=ALL-UNNAMED"
run() { # scenario, java args...
    local scenario=$1; shift
    java "$@" | python3 -c '
import sys
scenario, entry, seen, err = sys.argv[1], None, set(), ""
def flush():
    if entry is not None:
        line = "%s\t%s\t%s" % (scenario, entry, ",".join(sorted(seen)) or "-")
        print(line + ("\tERROR " + err if err else ""))
for raw in sys.stdin:
    parts = raw.split()
    if not parts:
        continue
    if parts[0] == "ORACLE-ENTRY":
        flush(); entry, seen, err = parts[1], set(), ""
    elif parts[0] == "ORACLE-PROVIDER":
        seen.add(parts[1])
    elif parts[0] == "ORACLE-ERROR":
        err = parts[1]
flush()
' "$scenario"
}

{
    run classpath -cp "oracle/classes:$S" oracle.OracleRunner "${specs_sl[@]}"
    run classpath -cp "oracle/classes:$M/fixture.jpms.app:$M/fixture.jpms.api:$M/fixture.jpms.impl:$A" \
        oracle.OracleRunner "${specs_jpms[@]}"
    run module-path -cp oracle/classes \
        --module-path "$M/fixture.jpms.app:$M/fixture.jpms.api:$M/fixture.jpms.impl:oracle/classes/fixture-auto-1.2.3.jar:$M/fixture.jpms.nouses" \
        --add-modules ALL-MODULE-PATH $opens --add-opens fixture.jpms.nouses/fixture.jpms.nouses=ALL-UNNAMED \
        oracle.OracleRunner "${specs_jpms[@]}" fixture.jpms.nouses.NoUses#main
    run mixed -cp "oracle/classes:$A" \
        --module-path "$M/fixture.jpms.app:$M/fixture.jpms.api:$M/fixture.jpms.impl" \
        --add-modules ALL-MODULE-PATH $opens \
        oracle.OracleRunner "${specs_jpms[@]}"
} > oracle/expected.tsv
rm -rf oracle/classes
echo "wrote $root/testdata/oracle/expected.tsv ($(wc -l < oracle/expected.tsv) entries)"
