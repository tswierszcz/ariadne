#!/usr/bin/env bash
# Control app for the log4j 1.2 appender model (FRAMEWORKS:F2 l1,
# internal/log4jconfig): one tiny application logging one error through
# log4j 1.2.17, packaged twice with different log4j.properties:
#   - app-console.jar: a ConsoleAppender only;
#   - app-jdbc.jar: a ConsoleAppender and a JDBCAppender writing to an
#     in-memory HSQLDB 2.3.4 table.
# The classes are identical, so CHA alone gives both the same graph; the
# packaged configuration is what decides whether JDBCAppender
# (CVE-2022-23305, SQL built from the log message) ever runs.
#
# The script runs both on the real JVM (-Dlog4j.debug shows which
# appenders log4j creates; the app counts the rows JDBCAppender wrote), then
# runs bin/ariadne on both (-profile generic -roots all; OSV through the cache) and
# prints the CVE-2022-23305 and CVE-2021-4104 (JMSAppender) verdicts.
# Expected: jdbc REACHABLE / console NOT_REACHABLE for CVE-2022-23305;
# NOT_REACHABLE for CVE-2021-4104 in both (no JMSAppender configured).
#
# Produces under work/log4j1-demo/: app-console.jar, app-jdbc.jar,
# target/dependency/*.jar, deps.txt.
#
# Usage: scripts/prepare-log4j1-demo.sh [JAVA_HOME]   (any JDK 8+; default: java on PATH)
set -euo pipefail

if [ -n "${1:-}" ]; then
    export JAVA_HOME="$1"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/log4j1-demo"
rm -rf "$work"
mkdir -p "$work/src/main/java/demo" "$work/variants/console" "$work/variants/jdbc"
cd "$work"

# source/target rather than <release>: a JDK without ct.sym rejects --release.
cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>log4j1-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>8</maven.compiler.source>
    <maven.compiler.target>8</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>log4j</groupId>
      <artifactId>log4j</artifactId>
      <version>1.2.17</version>
    </dependency>
    <dependency>
      <groupId>org.hsqldb</groupId>
      <artifactId>hsqldb</artifactId>
      <version>2.3.4</version>
    </dependency>
  </dependencies>
  <build>
    <plugins>
      <plugin>
        <groupId>org.apache.maven.plugins</groupId>
        <artifactId>maven-compiler-plugin</artifactId>
        <version>3.11.0</version>
      </plugin>
    </plugins>
  </build>
</project>
EOF

cat > src/main/java/demo/Main.java <<'EOF'
package demo;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.ResultSet;
import org.apache.log4j.LogManager;
import org.apache.log4j.Logger;

public class Main {
    private static final String URL = "jdbc:hsqldb:mem:logs";

    public static void main(String[] args) throws Exception {
        try (Connection c = DriverManager.getConnection(URL, "SA", "")) {
            c.createStatement().execute("CREATE TABLE log (msg VARCHAR(200))");
        }
        Logger.getLogger(Main.class).error("order 42 failed");
        LogManager.shutdown();
        try (Connection c = DriverManager.getConnection(URL, "SA", "");
             ResultSet rs = c.createStatement().executeQuery("SELECT COUNT(*) FROM log")) {
            rs.next();
            System.out.println("rows written by JDBCAppender: " + rs.getInt(1));
        }
    }
}
EOF

cat > variants/console/log4j.properties <<'EOF'
log4j.rootLogger=INFO, stdout
log4j.appender.stdout=org.apache.log4j.ConsoleAppender
log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
log4j.appender.stdout.layout.ConversionPattern=%-5p %c - %m%n
EOF

cat > variants/jdbc/log4j.properties <<'EOF'
log4j.rootLogger=INFO, stdout, db
log4j.appender.stdout=org.apache.log4j.ConsoleAppender
log4j.appender.stdout.layout=org.apache.log4j.PatternLayout
log4j.appender.stdout.layout.ConversionPattern=%-5p %c - %m%n
log4j.appender.db=org.apache.log4j.jdbc.JDBCAppender
log4j.appender.db.driver=org.hsqldb.jdbc.JDBCDriver
log4j.appender.db.URL=jdbc:hsqldb:mem:logs
log4j.appender.db.user=SA
log4j.appender.db.password=
log4j.appender.db.sql=INSERT INTO log (msg) VALUES ('%m')
EOF

java -version
mvn -q -B package -DskipTests
mvn -q -B dependency:copy-dependencies -DoutputDirectory=target/dependency -DincludeScope=runtime
mvn -q -B dependency:list -DoutputFile=deps.txt -DincludeScope=runtime

# One jar per variant: the same classes plus that variant's configuration
# (zipped with Python, so a JDK without the jar tool is enough).
for v in console jdbc; do
    python3 -I - "target/classes" "variants/$v/log4j.properties" "app-$v.jar" <<'PY'
import os, sys, zipfile
classes, props, out = sys.argv[1:]
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    for d, _, files in os.walk(classes):
        for f in sorted(files):
            p = os.path.join(d, f)
            z.write(p, os.path.relpath(p, classes))
    z.write(props, "log4j.properties")
PY
done

cp="target/dependency/*"
for v in console jdbc; do
    echo "== real JVM, $v"
    java -Dlog4j.debug -cp "app-$v.jar:$cp" demo.Main 2>&1 | grep -E "Parsing appender named|rows written|order 42"
done

echo "== ariadne"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
for v in console jdbc; do
    out="$work/ariadne-$v"
    bin/ariadne -cache-dir "$root/.cache" -app "$work/app-$v.jar" -deps "$work/target/dependency" \
        -deps-list "$work/deps.txt" -profile generic -roots all -out "$out.md" -json "$out.json" 2>&1 |
        grep -E "log4j 1.2 appenders" || true
    python3 -I - "$out.json" "$v" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
rows = d["results"] if isinstance(d, dict) and "results" in d else d
for r in rows:
    if r.get("cve") in ("CVE-2022-23305", "CVE-2021-4104"):
        print(sys.argv[2], r["cve"], r["status"])
PY
done
