#!/usr/bin/env bash
# Generates and builds a tiny, hand-written Quarkus application: a
# jakarta.ws.rs (@Path/@GET) REST resource, an io.quarkus.scheduler
# (@Scheduled) job, and a MicroProfile Reactive Messaging (@Incoming)
# consumer, all three calling straight into org.yaml.snakeyaml:1.30's
# Yaml#load (CVE-2022-1471, the same manual-table seed CVE the other
# control apps use).
#
# This is the control app for the "quarkus" profile (PLAN.md Phase 7),
# the same role prepare-servlet-demo.sh/prepare-jakarta-demo.sh/
# prepare-micronaut-demo.sh play for "servlet"/"jakarta"/"micronaut":
# small enough to generate from a script rather than clone, exercising on
# a real compiled application run through the real CLI:
#   - the app auto-detects as "jakarta+quarkus", not "generic" -- Quarkus's
#     own REST layer is plain JAX-RS (already covered by jakarta.yaml),
#     and this app also carries io.quarkus.runtime.Quarkus/Application,
#     quarkus.yaml's own detect markers;
#   - -roots profile's root set narrows to exactly the three annotated
#     methods below (via roots.annotations -- already implemented for
#     spring.yaml/jakarta.yaml/micronaut.yaml; quarkus.yaml is a fourth
#     user), confirmed directly via Profile.AppRoots, not just the
#     log's entry-point count (the report itself only records one
#     shortest path per REACHABLE CVE, so it alone can't show all three);
#   - CVE-2022-1471 is REACHABLE either way.
#
# Needs JDK 17+ (Quarkus 3.x's own jars are compiled for it) and a
# network connection the first time (pulls ~100 transitive dependencies,
# mostly Vert.x/Netty -- Quarkus's own runtime, not anything this demo's
# three classes use directly).
#
# Produces under work/quarkus-demo/:
#   target/classes, target/dependency/*.jar, deps.txt
#
# Usage: scripts/prepare-quarkus-demo.sh [JAVA_HOME for JDK 17+]
set -euo pipefail

JAVA_HOME_ARG="${1:-${JAVA17_HOME:-${JAVA_HOME:-}}}"
if [ -z "$JAVA_HOME_ARG" ]; then
    echo "usage: $0 <path to a JDK 17+ JAVA_HOME>" >&2
    echo "(or set JAVA17_HOME/JAVA_HOME)" >&2
    exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/quarkus-demo"
mkdir -p "$work/src/main/java/demo"
cd "$work"

cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>quarkus-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>17</maven.compiler.source>
    <maven.compiler.target>17</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>io.quarkus</groupId>
      <artifactId>quarkus-core</artifactId>
      <version>3.9.5</version>
    </dependency>
    <dependency>
      <groupId>io.quarkus</groupId>
      <artifactId>quarkus-resteasy-reactive</artifactId>
      <version>3.9.5</version>
    </dependency>
    <dependency>
      <groupId>io.quarkus</groupId>
      <artifactId>quarkus-scheduler-api</artifactId>
      <version>3.9.5</version>
    </dependency>
    <dependency>
      <groupId>org.eclipse.microprofile.reactive.messaging</groupId>
      <artifactId>microprofile-reactive-messaging-api</artifactId>
      <version>3.0.1</version>
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

import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.QueryParam;
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

import io.quarkus.scheduler.Scheduled;
import org.yaml.snakeyaml.Yaml;

public class ScheduledYamlJob {
    @Scheduled(every = "15m")
    void run() {
        Object parsed = new Yaml().load("scheduled: true");
        System.out.println(parsed);
    }
}
EOF

cat > src/main/java/demo/YamlMessageConsumer.java <<'EOF'
package demo;

import org.eclipse.microprofile.reactive.messaging.Incoming;
import org.yaml.snakeyaml.Yaml;

public class YamlMessageConsumer {
    @Incoming("yaml-in")
    public void consume(String payload) {
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
