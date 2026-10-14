#!/usr/bin/env bash
# Generates and builds a tiny, hand-written Jakarta EE application (no
# Spring, no plain Servlet): a JAX-RS resource (@Path/@GET), an EJB
# (@Schedule), and a CDI observer method (@Observes), all three calling
# straight into org.yaml.snakeyaml:1.30's Yaml#load (CVE-2022-1471, the
# same manual-table seed CVE scripts/prepare-servlet-demo.sh and PLAN.md
# section 4 already use).
#
# This is the control app for the "jakarta" profile (PLAN.md Phase 7),
# the same role prepare-servlet-demo.sh plays for "servlet": small enough
# to generate from a script rather than clone, exercising on a real
# compiled application run through the real CLI:
#   - profile auto-detection picks "jakarta", not "generic", purely from
#     javax.ws.rs.Path being present on the classpath;
#   - -roots profile's root set narrows to exactly the resource's @GET
#     method, the EJB's @Schedule method (via roots.annotations, already
#     implemented for spring.yaml/servlet.yaml -- jakarta.yaml is just a
#     new user of it, not new matching logic), and the CDI observer
#     method (via roots.parameter_annotations, matching @Observes on the
#     method's own event parameter -- internal/classfile now reads
#     RuntimeVisibleParameterAnnotations);
#   - CVE-2022-1471 is REACHABLE either way.
#
# Produces under work/jakarta-demo/:
#   target/classes, target/dependency/*.jar, deps.txt
#
# Usage: scripts/prepare-jakarta-demo.sh [JAVA_HOME for JDK 11+]
set -euo pipefail

JAVA_HOME_ARG="${1:-${JAVA11_HOME:-${JAVA_HOME:-}}}"
if [ -z "$JAVA_HOME_ARG" ]; then
    echo "usage: $0 <path to a JDK 11+ JAVA_HOME>" >&2
    echo "(or set JAVA11_HOME/JAVA_HOME)" >&2
    exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/jakarta-demo"
mkdir -p "$work/src/main/java/demo"
cd "$work"

cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>jakarta-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>javax.ws.rs</groupId>
      <artifactId>javax.ws.rs-api</artifactId>
      <version>2.1.1</version>
    </dependency>
    <dependency>
      <groupId>javax.ejb</groupId>
      <artifactId>javax.ejb-api</artifactId>
      <version>3.2.2</version>
    </dependency>
    <dependency>
      <groupId>javax.enterprise</groupId>
      <artifactId>cdi-api</artifactId>
      <version>2.0.SP1</version>
    </dependency>
    <dependency>
      <groupId>org.yaml</groupId>
      <artifactId>snakeyaml</artifactId>
      <version>1.30</version>
    </dependency>
  </dependencies>
</project>
EOF

cat > src/main/java/demo/YamlResource.java <<'EOF'
package demo;

import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.QueryParam;
import org.yaml.snakeyaml.Yaml;

@Path("/yaml")
public class YamlResource {
    @GET
    public String get(@QueryParam("payload") String payload) {
        Object parsed = new Yaml().load(payload);
        return String.valueOf(parsed);
    }
}
EOF

cat > src/main/java/demo/ScheduledYamlJob.java <<'EOF'
package demo;

import javax.ejb.Schedule;
import javax.ejb.Stateless;
import org.yaml.snakeyaml.Yaml;

@Stateless
public class ScheduledYamlJob {
    @Schedule(hour = "*", minute = "*/15")
    public void run() {
        Object parsed = new Yaml().load("scheduled: true");
        System.out.println(parsed);
    }
}
EOF

cat > src/main/java/demo/YamlObserver.java <<'EOF'
package demo;

import javax.enterprise.event.Observes;
import org.yaml.snakeyaml.Yaml;

public class YamlObserver {
    public void onEvent(@Observes String payload) {
        Object parsed = new Yaml().load(payload);
        System.out.println(parsed);
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
