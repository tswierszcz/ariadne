#!/usr/bin/env bash
# Generates and builds a tiny, hand-written Micronaut (no Spring, no
# Servlet/JAX-RS) application: one @Controller/@Get method calling
# straight into org.yaml.snakeyaml:1.30's Yaml#load (CVE-2022-1471, the
# same manual-table seed CVE the servlet/jakarta control apps use).
#
# This is the control app for the "micronaut" profile (PLAN.md Phase 7),
# the same role prepare-servlet-demo.sh/prepare-jakarta-demo.sh play for
# "servlet"/"jakarta": small enough to generate from a script rather than
# clone, exercising on a real compiled application run through the real
# CLI:
#   - profile auto-detection picks "micronaut", not "generic", purely
#     from io.micronaut.context.ApplicationContext being present on the
#     classpath;
#   - -roots profile's root set narrows to exactly the controller's @Get
#     method (via roots.annotations, already implemented for
#     spring.yaml/jakarta.yaml -- micronaut.yaml is just a new user);
#   - CVE-2022-1471 is REACHABLE either way.
#
# Pinned to Micronaut 3.9.7, not the current 4.x/5.x line: Micronaut 5.x
# (as of this writing) ships class files compiled for Java 25, newer than
# any JDK this script can assume is on hand, while 3.x targets Java 8/11
# -- consistent with the Java 11 this project's other demo/prep scripts
# use. This is purely a "which jar can javac read", not a profile
# limitation: ariadne's own bytecode parser has no version ceiling.
#
# Produces under work/micronaut-demo/:
#   target/classes, target/dependency/*.jar, deps.txt
#
# Usage: scripts/prepare-micronaut-demo.sh [JAVA_HOME for JDK 11+]
set -euo pipefail

JAVA_HOME_ARG="${1:-${JAVA11_HOME:-${JAVA_HOME:-}}}"
if [ -z "$JAVA_HOME_ARG" ]; then
    echo "usage: $0 <path to a JDK 11+ JAVA_HOME>" >&2
    echo "(or set JAVA11_HOME/JAVA_HOME)" >&2
    exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/micronaut-demo"
mkdir -p "$work/src/main/java/demo"
cd "$work"

cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>micronaut-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>io.micronaut</groupId>
      <artifactId>micronaut-http</artifactId>
      <version>3.9.7</version>
    </dependency>
    <dependency>
      <groupId>io.micronaut</groupId>
      <artifactId>micronaut-context</artifactId>
      <version>3.9.7</version>
    </dependency>
    <dependency>
      <groupId>org.yaml</groupId>
      <artifactId>snakeyaml</artifactId>
      <version>1.30</version>
    </dependency>
  </dependencies>
</project>
EOF

cat > src/main/java/demo/YamlController.java <<'EOF'
package demo;

import io.micronaut.http.annotation.Controller;
import io.micronaut.http.annotation.Get;
import io.micronaut.http.annotation.QueryValue;
import org.yaml.snakeyaml.Yaml;

@Controller("/yaml")
public class YamlController {
    @Get
    public String get(@QueryValue String payload) {
        Object parsed = new Yaml().load(payload);
        return String.valueOf(parsed);
    }
}
EOF

export JAVA_HOME="$JAVA_HOME_ARG"
export PATH="$JAVA_HOME/bin:$PATH"
java -version

mvn -q package -DskipTests
mvn -q dependency:copy-dependencies -DoutputDirectory=target/dependency
mvn -q dependency:list -DoutputFile=deps.txt
# dependency:list output has ANSI color codes when color output is on; strip them.
sed -i 's/\x1b\[[0-9;]*m//g' deps.txt

echo "done: $work/{target/classes,target/dependency/,deps.txt}"
