#!/usr/bin/env bash
# Clones WebGoat at tag v8.2.2, builds the webgoat-container module and the
# vulnerable-components lesson (old XStream 1.4.5, the control case) with
# JDK 17, and produces under work/WebGoat/:
#   app-classes.jar (container + vulnerable-components classes combined,
#     plus webgoat-server's two classes: the @SpringBootApplication launcher
#     StartWebGoat and HSQLDBDatabaseConfig),
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
#  - webgoat-server (the launcher module of the real deployable) depends on
#    every lesson, including cross-site-scripting, so it is not built with
#    Maven: its two source files are compiled with javac against the
#    container and its runtime dependencies, Lombok as the annotation
#    processor (the version the parent pom pins, from the local repository
#    the container build filled; override with LOMBOK_JAR). Without the
#    launcher no class enables Boot auto-configuration, so ariadne registers
#    no DispatcherServlet and the whole MVC request path is unreachable.
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

lombok_version="$(sed -n 's:.*<lombok.version>\(.*\)</lombok.version>.*:\1:p' pom.xml)"
LOMBOK_JAR="${LOMBOK_JAR:-$HOME/.m2/repository/org/projectlombok/lombok/$lombok_version/lombok-$lombok_version.jar}"
[ -f "$LOMBOK_JAR" ] || mvn -q dependency:get -Dartifact="org.projectlombok:lombok:$lombok_version"
rm -rf webgoat-server/target/classes
mkdir -p webgoat-server/target/classes
"$JAVA_HOME/bin/javac" -nowarn -source 15 -target 15 -processorpath "$LOMBOK_JAR" \
    -cp "webgoat-container/target/classes:$(ls app-dependency/*.jar | tr '\n' ':')$LOMBOK_JAR" \
    -d webgoat-server/target/classes webgoat-server/src/main/java/org/owasp/webgoat/*.java

"$JAVA_HOME/bin/jar" cf app-classes.jar -C webgoat-container/target/classes .
(cd webgoat-lessons/vulnerable-components/target/classes && "$JAVA_HOME/bin/jar" uf "$PWD/../../../../app-classes.jar" .)
"$JAVA_HOME/bin/jar" uf app-classes.jar -C webgoat-server/target/classes .

echo "done: $work/WebGoat/{app-classes.jar,deps.txt,app-dependency/}"
grep -i xstream deps.txt || echo "WARNING: xstream not found in deps.txt"
unzip -l app-classes.jar | grep -q org/owasp/webgoat/StartWebGoat.class || echo "WARNING: StartWebGoat (Boot launcher) not in app-classes.jar"
