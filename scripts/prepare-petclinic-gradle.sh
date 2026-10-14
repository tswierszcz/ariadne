#!/usr/bin/env bash
# Clones spring-petclinic at commit 4df621b (2021-05-10, "Make build work
# with Gradle") -- the oldest commit in the repo's history with a working
# build.gradle, and still Spring Boot 2.4.5, the same version
# prepare-petclinic.sh's Maven-built target uses (commit e7c879e,
# 2021-05-14, 4 days later). Builds it with Gradle instead of Maven, using
# scripts/gradle-copy-deps.init.gradle in place of
# dependency:copy-dependencies/dependency:list, and produces under
# work/spring-petclinic-gradle/:
#   app-classes.jar, dependency/*.jar, deps.txt
#
# This exists to verify ariadne's Gradle input support against an app
# whose Maven-built twin already has a known baseline (148 CVEs, 3
# REACHABLE under -profile spring -- see README.md/PLAN.md), not to find
# new CVEs: run ariadne against both and compare.
#
# Note: pom.xml at this exact commit has already moved to Spring Boot
# 2.5.6 (the Maven and Gradle build files were briefly out of sync during
# this migration), so the Gradle-resolved dependency set is not expected
# to match the Maven baseline jar-for-jar -- only to be close enough, from
# the same vulnerable era, to sanity-check the new input path. Verified
# once (see PLAN.md Phase 7's Gradle notes): 80/80 deps identified, 144
# CVEs, 3/56/29/50/6 REACHABLE/FRAMEWORK/NOT_REACHABLE/UNKNOWN/
# NOT_APPLICABLE -- REACHABLE matches the Maven baseline exactly; the
# small differences in the other counts track the 80-vs-87-dependency gap.
#
# Build quirk fixed below, not upstream: as committed, build.gradle's
# buildscript block pulls in `ro.isdc.wro4j.gradle:wro4j-gradle-plugin
# 1.8.0.Beta4` (for LESS->CSS compilation) with no repository to resolve
# it from, and the artifact itself no longer exists on Maven Central or
# the Gradle Plugin Portal either -- it was a JCenter-only release, and
# JCenter shut down in 2021/2022. There is no live source to restore it
# from, so this script drops the wro4j plugin and its `webResources`
# block entirely (CSS/LESS bundling, irrelevant to bytecode analysis) and
# remaps its `webjarsRuntime` dependencies (the actual webjars jars,
# e.g. org.webjars:bootstrap -- relevant, since those are exactly the
# CVEs covered by the NOT_APPLICABLE/WebJar status) to the stock
# `runtimeOnly` configuration, which is all ariadne's -deps needs them on.
#
# Usage: scripts/prepare-petclinic-gradle.sh [JAVA_HOME for JDK 11]
set -euo pipefail

JAVA11_HOME="${1:-${JAVA11_HOME:-}}"
if [ -z "$JAVA11_HOME" ]; then
    echo "usage: $0 <path to a JDK 11 JAVA_HOME>" >&2
    echo "(or set JAVA11_HOME)" >&2
    exit 1
fi

COMMIT="4df621b41ed3013e527d4037d83a6cf756efd784"

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
init_script="$root/scripts/gradle-copy-deps.init.gradle"
work="$root/work"
mkdir -p "$work"
cd "$work"

if [ ! -d spring-petclinic-gradle ]; then
    git clone https://github.com/spring-projects/spring-petclinic spring-petclinic-gradle
fi
cd spring-petclinic-gradle

echo "checking out petclinic @ $COMMIT"
git checkout -q "$COMMIT"

# Dead-plugin workaround -- see the big comment above.
python3 - <<'PYEOF'
import re
path = "build.gradle"
s = open(path).read()
s = s.replace("""buildscript {
    dependencies {
        classpath 'ro.isdc.wro4j.gradle:wro4j-gradle-plugin:1.8.0.Beta4'
    }
}

""", "")
s = s.replace("apply plugin: 'wro4j'\n", "")
s = s.replace("webjarsRuntime ", "runtimeOnly ")
s = re.sub(r"\nwebResources \{.*?\n\}\n", "\n", s, flags=re.S)
open(path, "w").write(s)
PYEOF

export JAVA_HOME="$JAVA11_HOME"
export PATH="$JAVA_HOME/bin:$PATH"
java -version

./gradlew -q classes

rm -rf dependency
./gradlew -q -I "$init_script" ariadneCopyDeps ariadneDepsList \
    -PariadneOutputDir="$PWD/dependency" \
    -PariadneDepsListFile="$PWD/deps.txt"

"$JAVA_HOME/bin/jar" cf app-classes.jar -C build/classes/java/main .

echo "done: $work/spring-petclinic-gradle/{app-classes.jar,deps.txt,dependency/}"
