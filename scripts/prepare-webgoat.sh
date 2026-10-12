#!/usr/bin/env bash
# Clones WebGoat at tag v8.2.2, builds the webgoat-container module and the
# vulnerable-components lesson (old XStream 1.4.5, the control case) with
# JDK 17, and produces under work/WebGoat/:
#   app-classes.jar (container + vulnerable-components classes combined),
#   app-dependency/*.jar, deps.txt
#
# Notes on quirks hit while building this:
#  - webgoat-lessons/cross-site-scripting runs the asciidoctor-maven-plugin
#    at build time (to render its lesson plan to HTML) and that plugin's
#    embedded JRuby fails under JDK 17's module system. We don't need that
#    lesson, so we build only the modules we do need (-pl), which skips it
#    entirely rather than patching around it.
#  - webgoat-lessons/vulnerable-components depends on webgoat-container as a
#    "provided" dependency, so webgoat-container must be `mvn install`-ed to
#    the local repo first, or dependency:copy-dependencies on the lesson
#    module fails to resolve it.
#
# Usage: scripts/prepare-webgoat.sh [JAVA_HOME for JDK 17+] [tag]
set -euo pipefail

JAVA17_HOME="${1:-${JAVA17_HOME:-}}"
TAG="${2:-v8.2.2}"
if [ -z "$JAVA17_HOME" ]; then
    echo "usage: $0 <path to a JDK 17+ JAVA_HOME> [tag]" >&2
    echo "(or set JAVA17_HOME)" >&2
    exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work"
mkdir -p "$work"
cd "$work"

if [ ! -d WebGoat ]; then
    git clone https://github.com/WebGoat/WebGoat
fi
cd WebGoat

echo "checking out WebGoat @ $TAG"
git checkout -q "$TAG"

export JAVA_HOME="$JAVA17_HOME"
export PATH="$JAVA_HOME/bin:$PATH"
java -version

mvn -q -pl webgoat-container -DskipTests install
mvn -q -pl webgoat-lessons/vulnerable-components -DskipTests package

rm -rf app-dependency
mvn -q -pl webgoat-container dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory="$PWD/app-dependency"
mvn -q -pl webgoat-lessons/vulnerable-components dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory="$PWD/app-dependency"

mvn -q -pl webgoat-container dependency:list -DincludeScope=runtime -DoutputFile="$PWD/deps-container.txt"
mvn -q -pl webgoat-lessons/vulnerable-components dependency:list -DincludeScope=runtime -DoutputFile="$PWD/deps-vc.txt"
sed 's/\x1b\[[0-9;]*m//g' deps-container.txt deps-vc.txt | grep -E '^   [a-zA-Z0-9_.-]+:' | sort -u > deps.txt
rm -f deps-container.txt deps-vc.txt

"$JAVA_HOME/bin/jar" cf app-classes.jar -C webgoat-container/target/classes .
(cd webgoat-lessons/vulnerable-components/target/classes && "$JAVA_HOME/bin/jar" uf "$PWD/../../../../app-classes.jar" .)

echo "done: $work/WebGoat/{app-classes.jar,deps.txt,app-dependency/}"
grep -i xstream deps.txt || echo "WARNING: xstream not found in deps.txt"
