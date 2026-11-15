#!/usr/bin/env bash
# Control app for CompletionStage continuations (ASYNC_CALLBACKS A3 slice
# 1, internal/callback's completion catalog): a plain Java application
# whose Render class (a java.util.function.Function, not a lambda) renders a
# template with commons-text 1.9's StringSubstitutor#replace (CVE-2022-42889,
# Text4Shell — a manual-table CVE in internal/vulnmap/vuln_methods.yaml).
# Packaged twice with one class differing:
#   - app-registered.jar: main hands a new Render to thenApply on a
#     CompletableFuture — the JVM runs it;
#   - app-unregistered.jar: main creates the Render but registers nothing
#     — the JVM never runs it.
# A lambda needed no model (the invokedynamic creating it already reaches
# its body); an object of a class handed to a continuation did.
#
# The script runs each on the real JVM, then runs bin/ariadne (-profile
# generic -roots profile, from main; OSV through the cache) and prints the
# verdict of CVE-2022-42889 and its path.
#
# Produces under work/continuations-demo/: app-{registered,unregistered}.jar,
# lib/*.jar, deps.txt.
#
# Usage: scripts/prepare-continuations-demo.sh [JAVA_HOME]   (any JDK 8+; default: java on PATH)
set -euo pipefail

if [ -n "${1:-}" ]; then
    export JAVA_HOME="$1"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/continuations-demo"
rm -rf "$work"
mkdir -p "$work/src/main/java/demo" "$work/variants/registered" "$work/variants/unregistered"
cd "$work"

# source/target rather than <release>: a JDK without ct.sym rejects --release.
cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>continuations-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>8</maven.compiler.source>
    <maven.compiler.target>8</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>org.apache.commons</groupId>
      <artifactId>commons-text</artifactId>
      <version>1.9</version>
    </dependency>
  </dependencies>
</project>
EOF

cat > src/main/java/demo/Render.java <<'EOF'
package demo;

import java.util.function.Function;
import org.apache.commons.text.StringSubstitutor;

/** Renders an order confirmation: a Function class, not a lambda. */
public class Render implements Function<String, String> {
    public String apply(String text) {
        return StringSubstitutor.createInterpolator().replace(text);
    }
}
EOF
cat > variants/registered/Main.java <<'EOF'
package demo;

import java.util.concurrent.CompletableFuture;

public class Main {
    public static void main(String[] args) {
        String out = CompletableFuture.supplyAsync(Main::order).thenApply(new Render()).join();
        System.out.println("rendered: " + out);
    }

    static String order() { return "order ${java:version}"; }
}
EOF
cat > variants/unregistered/Main.java <<'EOF'
package demo;

import java.util.concurrent.CompletableFuture;

public class Main {
    public static void main(String[] args) {
        Render render = new Render();
        String out = CompletableFuture.supplyAsync(Main::order).thenApply(String::trim).join();
        System.out.println("not rendered: " + out + " (" + render.getClass().getSimpleName() + " unused)");
    }

    static String order() { return "order ${java:version}"; }
}
EOF

java -version
mvn -q -B dependency:copy-dependencies -DoutputDirectory=lib -DincludeScope=runtime
mvn -q -B dependency:list -DoutputFile=deps.txt -DincludeScope=runtime
# One jar per variant: the shared classes plus that variant's Main
# (zipped with Python, so a JDK without the jar tool is enough).
for v in registered unregistered; do
    cp "variants/$v/Main.java" src/main/java/demo/Main.java
    mvn -q -B clean compile
    python3 -I - target/classes "app-$v.jar" <<'PY'
import os, sys, zipfile
classes, out = sys.argv[1:]
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    for d, _, files in sorted(os.walk(classes)):
        if d != classes:
            z.write(d, os.path.relpath(d, classes) + "/")
        for f in sorted(files):
            p = os.path.join(d, f)
            z.write(p, os.path.relpath(p, classes))
PY
done

for v in registered unregistered; do
    echo "== real JVM, $v"
    java -cp "app-$v.jar:lib/*" demo.Main 2>&1 | sed 's/^/  /'
done

echo "== ariadne"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
for v in registered unregistered; do
    out="$work/ariadne-$v"
    bin/ariadne -cache-dir "$root/.cache" -app "$work/app-$v.jar" -deps "$work/lib" \
        -deps-list "$work/deps.txt" -profile generic -roots profile -out "$out.md" -json "$out.json" 2>&1 |
        grep -E "callbacks: .*modeled" | sed -E 's/^.*callbacks:/  callbacks:/' || true
    python3 -I - "$out.json" "$v" <<'PY'
import json, sys
for r in json.load(open(sys.argv[1])):
    if r.get("cve") == "CVE-2022-42889":
        print(" ", sys.argv[2], r["cve"], r["status"], " > ".join(r.get("path") or [])[:400])
PY
done
