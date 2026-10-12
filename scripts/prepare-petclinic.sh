#!/usr/bin/env bash
# Clones spring-petclinic at an old (pre-2021-06-01) commit, builds it with
# JDK 11, and produces everything Phase 1 needs under work/spring-petclinic/:
#   app-classes.jar, target/dependency/*.jar, deps.txt
#
# Usage: scripts/prepare-petclinic.sh [JAVA_HOME for JDK 11]
set -euo pipefail

JAVA11_HOME="${1:-${JAVA11_HOME:-}}"
if [ -z "$JAVA11_HOME" ]; then
    echo "usage: $0 <path to a JDK 11 JAVA_HOME>" >&2
    echo "(or set JAVA11_HOME)" >&2
    exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work"
mkdir -p "$work"
cd "$work"

if [ ! -d spring-petclinic ]; then
    git clone https://github.com/spring-projects/spring-petclinic
fi
cd spring-petclinic

commit="$(git rev-list -1 --before=2021-06-01 main)"
echo "checking out petclinic @ $commit"
git checkout -q "$commit"

export JAVA_HOME="$JAVA11_HOME"
export PATH="$JAVA_HOME/bin:$PATH"
java -version

./mvnw -q package -DskipTests
./mvnw -q dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory=target/dependency
./mvnw -q dependency:list -DincludeScope=runtime -DoutputFile=deps.txt
# dependency:list output has ANSI color codes when color output is on; strip them.
sed -i 's/\x1b\[[0-9;]*m//g' deps.txt

"$JAVA_HOME/bin/jar" cf app-classes.jar -C target/classes .

echo "done: $work/spring-petclinic/{app-classes.jar,deps.txt,target/dependency/}"
