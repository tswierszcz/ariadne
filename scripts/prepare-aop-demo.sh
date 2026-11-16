#!/usr/bin/env bash
# Control app for the Spring AOP advice model (FRAMEWORKS:F3,
# internal/springaop, spring:aop): a minimal Spring Boot 2.4.3 application
# (spring-boot-starter-aop) whose @Before advice renders the order text
# with commons-text 1.9's StringSubstitutor#replace (CVE-2022-42889,
# Text4Shell — a manual-table CVE in internal/vulnmap/vuln_methods.yaml).
# main calls OrderService#place on the bean, packaged twice with one class
# differing:
#   - app-advised.jar: the aspect's pointcut is
#     execution(* demo.OrderService.place(..)) — the advice runs;
#   - app-unadvised.jar: the pointcut names OrderService#refund, which
#     nothing calls — the advice never runs.
# An advice method is not a root of the spring profile, so under -roots
# profile it is reachable only through the model's edge from the call of
# the advised method.
#
# The script runs each on the real JVM (the advice prints when it runs),
# then runs bin/ariadne (-profile spring, -roots profile and -roots all;
# OSV through the cache) and prints the model's log line and the verdict
# of CVE-2022-42889.
#
# Produces under work/aop-demo/: app-{advised,unadvised}.jar, lib/*.jar,
# deps.txt.
#
# Usage: scripts/prepare-aop-demo.sh [JAVA_HOME]   (any JDK 8+; default: java on PATH)
set -euo pipefail

if [ -n "${1:-}" ]; then
    export JAVA_HOME="$1"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/aop-demo"
rm -rf "$work"
mkdir -p "$work/src/main/java/demo" "$work/variants/advised" "$work/variants/unadvised"
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
  <artifactId>aop-demo</artifactId>
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
      <artifactId>spring-boot-starter-aop</artifactId>
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

@SpringBootApplication
public class DemoApplication {
    public static void main(String[] args) {
        ConfigurableApplicationContext ctx = SpringApplication.run(DemoApplication.class, args);
        ctx.getBean(OrderService.class).place("order ${java:version}");
        ctx.close();
    }
}
EOF
cat > src/main/java/demo/OrderService.java <<'EOF'
package demo;

import org.springframework.stereotype.Service;

@Service
public class OrderService {
    public void place(String text) { System.out.println("placed: " + text); }

    public void refund(String text) { System.out.println("refunded: " + text); }
}
EOF
for v in advised unadvised; do
    if [ $v = advised ]; then m=place; else m=refund; fi
    cat > variants/$v/AuditAspect.java <<EOF
package demo;

import org.apache.commons.text.StringSubstitutor;
import org.aspectj.lang.JoinPoint;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Before;
import org.springframework.stereotype.Component;

/** Renders an audit line before the advised order operation. */
@Aspect
@Component
public class AuditAspect {
    @Before("execution(* demo.OrderService.$m(..))")
    public void audit(JoinPoint jp) {
        String text = String.valueOf(jp.getArgs()[0]);
        System.out.println("advice ran: " + StringSubstitutor.createInterpolator().replace(text));
    }
}
EOF
done

java -version
mvn -q -B dependency:copy-dependencies -DoutputDirectory=lib -DincludeScope=runtime
mvn -q -B dependency:list -DoutputFile=deps.txt -DincludeScope=runtime
# One jar per variant: the shared classes plus that variant's AuditAspect
# (zipped with Python, so a JDK without the jar tool is enough; with
# directory entries, without which component scanning finds nothing).
for v in advised unadvised; do
    cp "variants/$v/AuditAspect.java" src/main/java/demo/AuditAspect.java
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

for v in advised unadvised; do
    echo "== real JVM, $v"
    java -cp "app-$v.jar:lib/*" demo.DemoApplication 2>&1 | grep -E "advice ran" || echo "  (advice did not run)"
done

echo "== ariadne"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
for v in advised unadvised; do
    for roots in profile all; do
        out="$work/ariadne-$v-$roots"
        bin/ariadne -cache-dir "$root/.cache" -app "$work/app-$v.jar" -deps "$work/lib" \
            -deps-list "$work/deps.txt" -profile spring -roots "$roots" -out "$out.md" -json "$out.json" 2>&1 |
            grep -E "Spring AOP \(" | sed -E 's/^.*Spring AOP/  Spring AOP/' || true
        python3 -I - "$out.json" "$v -roots $roots" <<'PY'
import json, sys
for r in json.load(open(sys.argv[1])):
    if r.get("cve") == "CVE-2022-42889":
        print(" ", sys.argv[2], r["cve"], r["status"], " > ".join(r.get("path") or [])[:400])
PY
    done
done
