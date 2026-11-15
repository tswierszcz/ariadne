#!/usr/bin/env bash
# Control app for the Spring application events model (FRAMEWORKS:F4
# slice 1, internal/springevents, spring:events): a minimal Spring Boot
# 2.4.3 application whose @TransactionalEventListener renders a template
# with commons-text 1.9's StringSubstitutor#replace (CVE-2022-42889,
# Text4Shell — a manual-table CVE in internal/vulnmap/vuln_methods.yaml).
# main places an order in a transaction (OrderService#place, @Transactional
# over a minimal transaction manager with synchronization), packaged twice
# with one class differing:
#   - app-published.jar: place() publishes an OrderPlaced — the listener
#     runs after commit;
#   - app-unpublished.jar: place() publishes an OrderAudited — nothing
#     publishes OrderPlaced, the listener never runs.
# A @TransactionalEventListener is not a root of the spring profile (only
# @EventListener itself is), so under -roots profile the listener is
# reachable only through the model's edge from the publishing call.
#
# The script runs each on the real JVM (the listener prints when it runs),
# then runs bin/ariadne (-profile spring, -roots profile and -roots all;
# OSV through the cache) and prints the model's log line and the verdict
# of CVE-2022-42889.
#
# Produces under work/events-demo/: app-{published,unpublished}.jar,
# lib/*.jar, deps.txt.
#
# Usage: scripts/prepare-events-demo.sh [JAVA_HOME]   (any JDK 8+; default: java on PATH)
set -euo pipefail

if [ -n "${1:-}" ]; then
    export JAVA_HOME="$1"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/events-demo"
rm -rf "$work"
mkdir -p "$work/src/main/java/demo" "$work/variants/published" "$work/variants/unpublished"
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
  <artifactId>events-demo</artifactId>
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
    </dependency>
    <dependency>
      <groupId>org.springframework</groupId>
      <artifactId>spring-tx</artifactId>
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

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.annotation.EnableTransactionManagement;

@SpringBootApplication
@EnableTransactionManagement
public class DemoApplication {
    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = SpringApplication.run(DemoApplication.class, args);
        ctx.getBean(OrderService.class).place("order ${java:version}");
        ctx.close();
    }

    @Bean
    PlatformTransactionManager transactionManager() { return new DemoTransactionManager(); }
}
EOF
cat > src/main/java/demo/DemoTransactionManager.java <<'EOF'
package demo;

import org.springframework.transaction.TransactionDefinition;
import org.springframework.transaction.support.AbstractPlatformTransactionManager;
import org.springframework.transaction.support.DefaultTransactionStatus;

/** No resource: begins and commits nothing, but runs transaction synchronization. */
public class DemoTransactionManager extends AbstractPlatformTransactionManager {
    protected Object doGetTransaction() { return new Object(); }
    protected void doBegin(Object transaction, TransactionDefinition definition) {}
    protected void doCommit(DefaultTransactionStatus status) {}
    protected void doRollback(DefaultTransactionStatus status) {}
}
EOF
cat > src/main/java/demo/OrderPlaced.java <<'EOF'
package demo;

public class OrderPlaced {
    public final String text;
    public OrderPlaced(String text) { this.text = text; }
}
EOF
cat > src/main/java/demo/OrderAudited.java <<'EOF'
package demo;

public class OrderAudited {
    public final String text;
    public OrderAudited(String text) { this.text = text; }
}
EOF
cat > src/main/java/demo/ConfirmationListener.java <<'EOF'
package demo;

import org.apache.commons.text.StringSubstitutor;
import org.springframework.stereotype.Component;
import org.springframework.transaction.event.TransactionalEventListener;

/** Renders the confirmation after the order's transaction commits. */
@Component
public class ConfirmationListener {
    @TransactionalEventListener
    public void onPlaced(OrderPlaced e) {
        System.out.println("listener ran: " + StringSubstitutor.createInterpolator().replace(e.text));
    }
}
EOF
for v in published unpublished; do
    if [ $v = published ]; then event=OrderPlaced; else event=OrderAudited; fi
    cat > variants/$v/OrderService.java <<EOF
package demo;

import org.springframework.context.ApplicationEventPublisher;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class OrderService {
    private final ApplicationEventPublisher events;

    public OrderService(ApplicationEventPublisher events) { this.events = events; }

    @Transactional
    public void place(String text) { events.publishEvent(new $event(text)); }
}
EOF
done

java -version
mvn -q -B dependency:copy-dependencies -DoutputDirectory=lib -DincludeScope=runtime
mvn -q -B dependency:list -DoutputFile=deps.txt -DincludeScope=runtime
# One jar per variant: the shared classes plus that variant's OrderService
# (zipped with Python, so a JDK without the jar tool is enough; with
# directory entries, without which component scanning finds nothing).
for v in published unpublished; do
    cp "variants/$v/OrderService.java" src/main/java/demo/OrderService.java
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

for v in published unpublished; do
    echo "== real JVM, $v"
    java -cp "app-$v.jar:lib/*" demo.DemoApplication 2>&1 | grep -E "listener ran" || echo "  (listener did not run)"
done

echo "== ariadne"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
for v in published unpublished; do
    for roots in profile all; do
        out="$work/ariadne-$v-$roots"
        bin/ariadne -cache-dir "$root/.cache" -app "$work/app-$v.jar" -deps "$work/lib" \
            -deps-list "$work/deps.txt" -profile spring -roots "$roots" -out "$out.md" -json "$out.json" 2>&1 |
            grep -E "Spring events \(" | sed -E 's/^.*Spring events/  Spring events/' || true
        python3 -I - "$out.json" "$v -roots $roots" <<'PY'
import json, sys
for r in json.load(open(sys.argv[1])):
    if r.get("cve") == "CVE-2022-42889":
        print(" ", sys.argv[2], r["cve"], r["status"], " > ".join(r.get("path") or [])[:400])
PY
    done
done
