#!/usr/bin/env bash
# Clones Lavagna (digitalfondue/lavagna, a Kanban issue tracker; archived)
# at tag lavagna-1.1.9 — plain Spring 5.3 MVC without Boot (Java
# configuration, WebApplicationInitializer), embedded Jetty 9.4 started by
# io.lavagna.config.Launcher, HSQLDB in memory in the dev profile, model
# classes in Kotlin — and produces under work/lavagna/:
#   app-classes.jar: the Java and Kotlin classes and src/main/resources,
#   app-dependency/*.jar, deps.txt: the runtime dependencies plus the
#     "provided" ones (Jetty, servlet API, jopt-simple), which the
#     project's executable war (lavagna-jetty-console.war) bundles.
#
# First blind test on an application ariadne was never tuned on
# (2026-11-25): it found the samcall fold cycle (Jetty's
# HttpChannelListeners, x = combine(x, l)).
#
# Notes on quirks hit while building this on a JDK 25 runtime:
#  - The project's Kotlin 1.3.61 (and 1.9.25) compiler cannot parse the
#    runtime's version string (25.0.5+8-1-Debian) and stops. The Kotlin
#    sources are compiled with kotlin-maven-plugin 2.2.21 in language
#    version 1.9 (K2 rejects an open val assigned in init), with the
#    null-check assertions off; the classes it emits use only what the
#    project's kotlin-stdlib 1.3.61 has (checked: every kotlin/ class and
#    Intrinsics method they name is in that jar). The dependencies are
#    unchanged.
#  - maven-compiler-plugin runs javac through error-prone 2.3.3, which
#    does not run on a current JDK, so the Java sources are compiled with
#    scripts/JavaBuild.java (-source/-target 8) against the Kotlin classes,
#    with debug information (-g, Maven's default): Spring picks one of the
#    two TaskScheduler beans by the parameter name, without it Lavagna does
#    not start.
#  - The first `mvn` fetches from Maven Central (once); later runs can add
#    -o.
#
# Usage: scripts/prepare-lavagna.sh [JAVA_HOME]   (JDK 17+ with jdk.compiler)
set -euo pipefail

if [ -n "${1:-}" ]; then
    export JAVA_HOME="$1"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
TAG="lavagna-1.1.9"
KOTLIN="2.2.21"

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="$root/work/lavagna"
mkdir -p "$out"
cd "$out"
if [ ! -d src ]; then
    git clone -q --depth 1 --branch "$TAG" https://github.com/digitalfondue/lavagna.git src
fi
java -version

cd src
mvn=(mvn -B -q ${MVN_OFFLINE:+-o})
rm -rf target/classes ../app-dependency
"${mvn[@]}" org.jetbrains.kotlin:kotlin-maven-plugin:$KOTLIN:compile \
    -Dkotlin.compiler.jvmTarget=1.8 -Dkotlin.compiler.languageVersion=1.9 -Dkotlin.compiler.apiVersion=1.9 \
    '-Dkotlin.compiler.arguments=-Xno-param-assertions,-Xno-call-assertions,-Xno-receiver-assertions'
"${mvn[@]}" dependency:build-classpath -Dmdep.outputFile=../compile-cp.txt -Dmdep.includeScope=compile
"${mvn[@]}" dependency:copy-dependencies -DincludeScope=runtime -DoutputDirectory=../app-dependency \
    dependency:list -DincludeScope=runtime -DoutputFile=../deps-runtime.txt
"${mvn[@]}" dependency:copy-dependencies -DincludeScope=provided -DoutputDirectory=../app-dependency \
    dependency:list -DincludeScope=provided -DoutputFile=../deps-provided.txt
cd ..
cat deps-runtime.txt deps-provided.txt > deps.txt

java "$root/scripts/JavaBuild.java" src/src/main/java java-classes.jar 8 "$(cat compile-cp.txt):src/target/classes" g
python3 -I - <<'PY'
import pathlib, zipfile
seen = set()
with zipfile.ZipFile('app-classes.jar', 'w', zipfile.ZIP_DEFLATED) as out:
    def add(name, data):
        if name not in seen:
            seen.add(name)
            out.writestr(name, data)
    with zipfile.ZipFile('java-classes.jar') as z:
        for n in z.namelist():
            if not n.endswith('/') and not n.endswith('.kt'):
                add(n, z.read(n))
    for base in ['src/target/classes', 'src/src/main/resources']:
        for p in sorted(pathlib.Path(base).rglob('*')):
            if p.is_file():
                add(str(p.relative_to(base)), p.read_bytes())
classes = sum(1 for n in seen if n.endswith('.class'))
print(f'app-classes.jar: {classes} classes, {len(seen) - classes} resources')
PY
echo "dependencies: $(ls app-dependency | wc -l) jars"
