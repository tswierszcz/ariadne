#!/usr/bin/env bash
# Clones MyBatis JPetStore 6 at tag jpetstore-6.0.2 (2016: a WAR with
# web.xml-registered Stripes filter/servlet and Spring's
# ContextLoaderListener, Spring 4.3 XML configuration, MyBatis over an
# embedded HSQLDB, log4j 1.2), builds it, and produces under
# work/jpetstore/:
#   app-classes.jar (WEB-INF/classes), target/dependency/*.jar (runtime
#   scope, the WAR's WEB-INF/lib), deps.txt, and the WAR itself
#   (target/jpetstore.war) for -app and for the runtime trace.
#
# The release targets Java 6 bytecode; a current JDK's javac no longer
# emits it, so the build is forced to -source/-target 8
# (-Dmaven.compiler.source/target; not --release, which needs the JDK's ct.sym),
# with no change to the sources. app-classes.jar is zipped with Python, so
# a JDK without the jar tool is enough.
#
# Usage: scripts/prepare-jpetstore.sh [JAVA_HOME]   (any JDK 8+; default: java on PATH)
set -euo pipefail

if [ -n "${1:-}" ]; then
    export JAVA_HOME="$1"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
TAG="jpetstore-6.0.2"

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work"
mkdir -p "$work"
cd "$work"

if [ ! -d jpetstore ]; then
    git clone https://github.com/mybatis/jpetstore-6 jpetstore
fi
cd jpetstore
echo "checking out JPetStore @ $TAG"
git checkout -q "$TAG"
java -version

mvnflags=(-q -B -Dmaven.test.skip=true -Dmaven.compiler.source=8 -Dmaven.compiler.target=8 -Dmaven.compiler.testSource=8 -Dmaven.compiler.testTarget=8
    -Dlicense.skip=true -Drat.skip=true -Dcheckstyle.skip=true -Denforcer.skip=true -Djacoco.skip=true)
# maven-war-plugin 3.0.0 (the release's own) fails to load on current JDKs;
# 3.4.0 is invoked directly, with the pom's configuration.
mvn "${mvnflags[@]}" compile org.apache.maven.plugins:maven-war-plugin:3.4.0:war
mvn "${mvnflags[@]}" dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory=target/dependency
mvn "${mvnflags[@]}" dependency:list -DincludeScope=runtime -DoutputFile=deps.txt
sed -i 's/\x1b\[[0-9;]*m//g' deps.txt

python3 -I - <<'PY'
import os, zipfile
with zipfile.ZipFile("app-classes.jar", "w", zipfile.ZIP_DEFLATED) as z:
    for d, _, files in sorted(os.walk("target/classes")):
        for f in sorted(files):
            p = os.path.join(d, f)
            z.write(p, os.path.relpath(p, "target/classes"))
PY

echo "done: $work/jpetstore/{app-classes.jar,deps.txt,target/dependency/,target/jpetstore.war}"
