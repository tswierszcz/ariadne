#!/usr/bin/env bash
# Control app for the logback appender model (FRAMEWORKS:F2 l2,
# internal/logbackconfig): a minimal Spring Boot 2.4.5 application
# (logback 1.2.3, the version petclinic and WebGoat ship) that logs at INFO
# and prints the appenders logback actually attached, packaged three times
# with different configuration and identical classes:
#   - app-default.jar: no logback file, no logging.file.* — Spring Boot's
#     default configuration (ConsoleAppender only);
#   - app-file.jar: application.properties sets logging.file.name — Boot
#     adds its RollingFileAppender;
#   - app-smtp.jar: logback-spring.xml with Boot's console appender and a
#     SMTPAppender whose JaninoEventEvaluator (CVE-2024-12798) runs on
#     every event (it never sends: the expression is false for INFO).
#
# The script runs each on the real JVM (the app lists the root logger's
# appenders), then runs bin/ariadne (-profile spring -roots all; OSV
# through the cache) and prints the model's decision and the verdicts of
# the three logback advisories whose CHA paths went through unconfigured
# appenders on petclinic/WebGoat.
#
# Produces under work/logback-demo/: app-{default,file,smtp}.jar,
# target/dependency/*.jar, deps.txt.
#
# Usage: scripts/prepare-logback-demo.sh [JAVA_HOME]   (any JDK 8+; default: java on PATH)
set -euo pipefail

if [ -n "${1:-}" ]; then
    export JAVA_HOME="$1"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/logback-demo"
rm -rf "$work"
mkdir -p "$work/src/main/java/demo" "$work/variants/default" "$work/variants/file" "$work/variants/smtp"
cd "$work"

# source/target rather than <release>: a JDK without ct.sym rejects --release.
cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>logback-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>8</maven.compiler.source>
    <maven.compiler.target>8</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter</artifactId>
      <version>2.4.5</version>
    </dependency>
    <dependency>
      <groupId>org.codehaus.janino</groupId>
      <artifactId>janino</artifactId>
      <version>3.1.3</version>
    </dependency>
    <!-- SMTPAppender's own dependency (JavaMail; activation is no longer in the JDK) -->
    <dependency>
      <groupId>com.sun.mail</groupId>
      <artifactId>javax.mail</artifactId>
      <version>1.6.2</version>
    </dependency>
    <dependency>
      <groupId>com.sun.activation</groupId>
      <artifactId>javax.activation</artifactId>
      <version>1.2.0</version>
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

cat > src/main/java/demo/DemoApplication.java <<'EOF'
package demo;

import java.util.Iterator;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class DemoApplication {
    private static final Logger log = LoggerFactory.getLogger(DemoApplication.class);

    public static void main(String[] args) {
        SpringApplication.run(DemoApplication.class, args).close();
    }

    @org.springframework.context.annotation.Bean
    org.springframework.boot.CommandLineRunner run() {
        return args -> {
            log.info("order 42 placed");
            ch.qos.logback.classic.Logger root = (ch.qos.logback.classic.Logger) LoggerFactory.getLogger(Logger.ROOT_LOGGER_NAME);
            StringBuilder names = new StringBuilder();
            for (Iterator<?> it = root.iteratorForAppenders(); it.hasNext(); ) {
                names.append(' ').append(it.next().getClass().getName());
            }
            System.out.println("root appenders:" + names);
        };
    }
}
EOF

: > variants/default/application.properties
cat > variants/file/application.properties <<'EOF'
logging.file.name=target/demo.log
EOF
: > variants/smtp/application.properties
cat > variants/smtp/logback-spring.xml <<'EOF'
<configuration>
  <include resource="org/springframework/boot/logging/logback/defaults.xml"/>
  <include resource="org/springframework/boot/logging/logback/console-appender.xml"/>
  <appender name="MAIL" class="ch.qos.logback.classic.net.SMTPAppender">
    <smtpHost>localhost</smtpHost>
    <to>ops@example.invalid</to>
    <from>app@example.invalid</from>
    <evaluator class="ch.qos.logback.classic.boolex.JaninoEventEvaluator">
      <expression>level &gt;= ERROR</expression>
    </evaluator>
  </appender>
  <root level="INFO">
    <appender-ref ref="CONSOLE"/>
    <appender-ref ref="MAIL"/>
  </root>
</configuration>
EOF

java -version
mvn -q -B package -DskipTests
mvn -q -B dependency:copy-dependencies -DoutputDirectory=target/dependency -DincludeScope=runtime
mvn -q -B dependency:list -DoutputFile=deps.txt -DincludeScope=runtime

# One jar per variant: the same classes plus that variant's resources
# (zipped with Python, so a JDK without the jar tool is enough).
for v in default file smtp; do
    python3 -I - "target/classes" "variants/$v" "app-$v.jar" <<'PY'
import os, sys, zipfile
classes, res, out = sys.argv[1:]
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    for base in (classes, res):
        for d, _, files in os.walk(base):
            for f in sorted(files):
                p = os.path.join(d, f)
                if base == classes and f.endswith((".properties", ".xml")):
                    continue
                z.write(p, os.path.relpath(p, base))
PY
done

for v in default file smtp; do
    echo "== real JVM, $v"
    java -cp "app-$v.jar:target/dependency/*" demo.DemoApplication 2>&1 | grep -E "root appenders"
done

echo "== ariadne"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
for v in default file smtp; do
    out="$work/ariadne-$v"
    bin/ariadne -cache-dir "$root/.cache" -app "$work/app-$v.jar" -deps "$work/target/dependency" \
        -deps-list "$work/deps.txt" -profile spring -roots all -out "$out.md" -json "$out.json" 2>&1 |
        grep -E "logback appenders" || true
    python3 -I - "$out.json" "$v" <<'PY'
import json, sys
for r in json.load(open(sys.argv[1])):
    if r.get("cve") in ("CVE-2021-42550", "CVE-2023-6378", "CVE-2024-12798"):
        print(sys.argv[2], r["cve"], r["status"])
PY
done
