#!/usr/bin/env bash
# Generates and builds a tiny, hand-written, plain-Servlet (no Spring)
# application: one HttpServlet whose doGet calls straight into
# org.yaml.snakeyaml:1.30's Yaml#load (CVE-2022-1471, the same manual-table
# seed CVE PLAN.md section 4 and testdata/reachfixture both use).
#
# This exists as the non-Spring control app for the "servlet" profile
# (PLAN.md Phase 7): it is small enough to generate from this script rather
# than clone, and its only purpose is to exercise, on a real compiled
# application run through the real CLI (not a synthetic in-memory test),
# three things a Spring-only target app can't by itself:
#   - profile auto-detection picks "servlet", not "generic", purely from
#     the presence of javax.servlet.http.HttpServlet on the classpath;
#   - -roots profile's root set narrows to exactly the servlet's own
#     doGet (via roots.overrides -> Profile.overrideRoots), not "every
#     app method";
#   - CVE-2022-1471 is REACHABLE either way, confirming the narrower root
#     set still finds the same path as -roots all.
#
# Produces under work/servlet-demo/:
#   target/classes, target/dependency/*.jar, deps.txt
#
# Usage: scripts/prepare-servlet-demo.sh [JAVA_HOME for JDK 11+]
set -euo pipefail

JAVA_HOME_ARG="${1:-${JAVA11_HOME:-${JAVA_HOME:-}}}"
if [ -z "$JAVA_HOME_ARG" ]; then
    echo "usage: $0 <path to a JDK 11+ JAVA_HOME>" >&2
    echo "(or set JAVA11_HOME/JAVA_HOME)" >&2
    exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/servlet-demo"
mkdir -p "$work/src/main/java/demo"
cd "$work"

cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>servlet-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>javax.servlet</groupId>
      <artifactId>javax.servlet-api</artifactId>
      <version>4.0.1</version>
    </dependency>
    <dependency>
      <groupId>org.yaml</groupId>
      <artifactId>snakeyaml</artifactId>
      <version>1.30</version>
    </dependency>
  </dependencies>
</project>
EOF

cat > src/main/java/demo/YamlServlet.java <<'EOF'
package demo;

import java.io.IOException;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import org.yaml.snakeyaml.Yaml;

public class YamlServlet extends HttpServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse resp)
            throws ServletException, IOException {
        Object parsed = new Yaml().load(req.getParameter("payload"));
        resp.getWriter().write(String.valueOf(parsed));
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
