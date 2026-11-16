#!/usr/bin/env bash
# Control app for the spring profile's message listener roots: a minimal
# Spring Boot 2.4.3 application with an in-memory ActiveMQ broker
# (spring-boot-starter-activemq, vm://) whose JMS listener renders the
# message with commons-text 1.9's StringSubstitutor#replace
# (CVE-2022-42889, Text4Shell — a manual-table CVE in
# internal/vulnmap/vuln_methods.yaml). main sends a message to the
# "orders" queue with JmsTemplate and waits for the listener, packaged
# twice with one class differing:
#   - app-listening.jar: OrderListener#onOrder is a @JmsListener — the
#     listener container runs it (DefaultMessageListenerContainer →
#     MessagingMessageListenerAdapter → InvocableHandlerMethod →
#     Method#invoke);
#   - app-silent.jar: the same method without the annotation — nothing
#     calls it.
# Under -roots profile a listener method is reachable only as a root of
# the profile: nothing in the application calls it.
#
# The script runs each on the real JVM (the listener prints when it runs),
# then runs bin/ariadne (-profile spring, -roots profile and -roots all;
# OSV through the cache) and prints the verdict of CVE-2022-42889.
#
# Produces under work/listeners-demo/: app-{listening,silent}.jar,
# lib/*.jar, deps.txt.
#
# Usage: scripts/prepare-listeners-demo.sh [JAVA_HOME]   (any JDK 8+; default: java on PATH)
set -euo pipefail

if [ -n "${1:-}" ]; then
    export JAVA_HOME="$1"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/listeners-demo"
rm -rf "$work"
mkdir -p "$work/src/main/java/demo" "$work/variants/listening" "$work/variants/silent"
cd "$work"

# source/target rather than <release>: a JDK without ct.sym rejects --release.
cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>2.4.3</version>
  </parent>
  <groupId>demo</groupId>
  <artifactId>listeners-demo</artifactId>
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
      <artifactId>spring-boot-starter-activemq</artifactId>
    </dependency>
    <dependency>
      <groupId>org.apache.activemq</groupId>
      <artifactId>activemq-broker</artifactId>
    </dependency>
    <dependency>
      <groupId>org.apache.commons</groupId>
      <artifactId>commons-text</artifactId>
      <version>1.9</version>
    </dependency>
  </dependencies>
</project>
EOF

cat > src/main/java/demo/DemoApplication.java <<'EOF'
package demo;

import java.util.concurrent.TimeUnit;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.jms.core.JmsTemplate;

@SpringBootApplication
public class DemoApplication {
    public static void main(String[] args) throws Exception {
        ConfigurableApplicationContext ctx = SpringApplication.run(DemoApplication.class, args);
        ctx.getBean(JmsTemplate.class).convertAndSend("orders", "order ${java:version}");
        OrderListener.done.await(10, TimeUnit.SECONDS);
        ctx.close();
    }
}
EOF

listener() { # variant annotation-line
    sed "s|@ANN@|$2|" > "variants/$1/OrderListener.java" <<'EOF'
package demo;

import java.util.concurrent.CountDownLatch;
import org.apache.commons.text.StringSubstitutor;
import org.springframework.jms.annotation.JmsListener;
import org.springframework.stereotype.Component;

/** Renders an incoming order message. */
@Component
public class OrderListener {
    static final CountDownLatch done = new CountDownLatch(1);

    @ANN@
    public void onOrder(String text) {
        System.out.println("listener ran: " + StringSubstitutor.createInterpolator().replace(text));
        done.countDown();
    }
}
EOF
}
listener listening '@JmsListener(destination = "orders")'
listener silent '// not a listener'

java -version
mvn -q -B dependency:copy-dependencies -DoutputDirectory=lib -DincludeScope=runtime
mvn -q -B dependency:list -DoutputFile=deps.txt -DincludeScope=runtime
# One jar per variant: the shared classes plus that variant's OrderListener
# (zipped with Python, so a JDK without the jar tool is enough; with
# directory entries, without which component scanning finds nothing).
for v in listening silent; do
    cp "variants/$v/OrderListener.java" src/main/java/demo/OrderListener.java
    mvn -q -B clean compile
    python3 -I - target/classes "app-$v.jar" <<'PY'
import os, sys, zipfile
classes, out = sys.argv[1:]
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    for d, _, files in sorted(os.walk(classes)):
        if d != classes:  # directory entries: Spring's component scan lists them
            z.write(d, os.path.relpath(d, classes) + "/")
        for f in sorted(files):
            p = os.path.join(d, f)
            z.write(p, os.path.relpath(p, classes))
PY
done

for v in listening silent; do
    echo "== real JVM, $v"
    java -cp "app-$v.jar:lib/*" demo.DemoApplication 2>&1 | grep -E "listener ran" || echo "  (listener did not run)"
done

echo "== ariadne"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
for v in listening silent; do
    for roots in profile all; do
        out="$work/ariadne-$v-$roots"
        bin/ariadne -cache-dir "$root/.cache" -app "$work/app-$v.jar" -deps "$work/lib" \
            -deps-list "$work/deps.txt" -profile spring -roots "$roots" -out "$out.md" -json "$out.json" > /dev/null 2>&1
        python3 -I - "$out.json" "$v -roots $roots" <<'PY'
import json, sys
for r in json.load(open(sys.argv[1])):
    if r.get("cve") == "CVE-2022-42889":
        print(" ", sys.argv[2], r["cve"], r["status"], " > ".join(r.get("path") or [])[:400])
PY
    done
done
