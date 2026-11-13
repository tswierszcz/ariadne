#!/usr/bin/env bash
# Clones WebGoat at tag v8.2.2, builds what its real deployable
# (webgoat-server) contains with JDK 17, and produces under work/WebGoat/:
#   app-classes.jar: webgoat-container (classes and resources), the classes
#     of every lesson (webgoat-lessons/*, the vulnerable-components lesson
#     with old XStream 1.4.5 included) and webgoat-server's two classes —
#     the @SpringBootApplication launcher StartWebGoat and
#     HSQLDBDatabaseConfig,
#   app-dependency/*.jar, deps.txt: webgoat-server's runtime dependencies
#     (the set Maven resolves for the deployable, so one version per
#     artifact), without WebGoat's own modules and without
#     spring-boot-devtools (optional; Boot's repackaged jar leaves it out).
#
# Notes on quirks hit while building this:
#  - webgoat-lessons/cross-site-scripting binds the asciidoctor-maven-plugin
#    to generate-resources (to render its lesson plan to HTML), and that
#    plugin's embedded JRuby fails under JDK 17's module system. The lessons
#    are therefore compiled with the plugin goals called directly
#    (resources:resources compiler:compile), which runs no other plugin,
#    instead of a lifecycle phase.
#  - The lessons depend on webgoat-container ("provided", and its test jar)
#    and webgoat-server on every lesson, so the container is `mvn install`-ed,
#    the parent poms are installed (-N), and each lesson's classes are
#    installed as its jar (install:install-file) before webgoat-server is
#    compiled and its dependencies resolved.
#  - Lesson resources (lesson plans, HTML, i18n, SQL migrations) are not put
#    into app-classes.jar: ariadne reads none of them, and the lessons'
#    i18n files share names.
#  - Without the launcher no class enables Boot auto-configuration, so
#    ariadne registers no DispatcherServlet and the whole MVC request path
#    is unreachable.
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
wg="$PWD"

echo "checking out WebGoat @ $TAG"
git checkout -q "$TAG"

export JAVA_HOME="$JAVA17_HOME"
export PATH="$JAVA_HOME/bin:$PATH"
java -version

mvn -q -N install
mvn -q -pl webgoat-container -DskipTests install
mvn -q -f webgoat-lessons/pom.xml -N install -DskipTests

lessons=()
for d in webgoat-lessons/*/; do
    [ -f "$d/pom.xml" ] && lessons+=("${d%/}")
done
mvn -q -pl "$(IFS=,; echo "${lessons[*]}")" resources:resources compiler:compile
for d in "${lessons[@]}"; do
    name="$(basename "$d")"
    "$JAVA_HOME/bin/jar" cf "$d/target/$name.jar" -C "$d/target/classes" .
    mvn -q install:install-file -Dfile="$d/target/$name.jar" -DpomFile="$d/pom.xml"
done
mvn -q -pl webgoat-server resources:resources compiler:compile

exclude=(-DexcludeGroupIds=org.owasp.webgoat,org.owasp.webgoat.lesson -DexcludeArtifactIds=spring-boot-devtools)
rm -rf app-dependency
mvn -q -pl webgoat-server dependency:copy-dependencies -DincludeScope=runtime "${exclude[@]}" -DoutputDirectory="$wg/app-dependency"
mvn -q -pl webgoat-server dependency:list -DincludeScope=runtime "${exclude[@]}" -DoutputFile="$wg/deps-server.txt"
sed 's/\x1b\[[0-9;]*m//g' deps-server.txt | grep -E '^   [a-zA-Z0-9_.-]+:' | sort -u > deps.txt
rm -f deps-server.txt

"$JAVA_HOME/bin/jar" cf app-classes.jar -C webgoat-container/target/classes .
for d in "${lessons[@]}" webgoat-server; do
    (cd "$d/target/classes" && find . -name '*.class' | sort > ../classes.lst && "$JAVA_HOME/bin/jar" uf "$wg/app-classes.jar" @../classes.lst)
done

echo "done: $work/WebGoat/{app-classes.jar,deps.txt,app-dependency/}"
grep -i xstream deps.txt || echo "WARNING: xstream not found in deps.txt"
unzip -l app-classes.jar | grep -q org/owasp/webgoat/StartWebGoat.class || echo "WARNING: StartWebGoat (Boot launcher) not in app-classes.jar"
unzip -l app-classes.jar | grep -q org/owasp/webgoat/sql_injection/ || echo "WARNING: lesson classes (sql-injection) not in app-classes.jar"
