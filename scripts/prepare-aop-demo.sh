#!/usr/bin/env bash
# Control app for the Spring AOP advice model (FRAMEWORKS:F3,
# internal/springaop, spring:aop): a minimal Spring Boot 2.4.3 application
# (spring-boot-starter-aop) whose @Before advice renders the order text
# with commons-text 1.9's StringSubstitutor#replace (CVE-2022-42889,
# Text4Shell — a manual-table CVE in internal/vulnmap/vuln_methods.yaml).
# main calls OrderService#place on the bean, packaged three times:
#   - app-advised.jar: the @Aspect's pointcut is
#     execution(* demo.OrderService.place(..)) && args(text) — the advice
#     runs (text bound by name: Boot compiles with -parameters);
#   - app-unadvised.jar: the pointcut names OrderService#refund, which
#     nothing calls — the advice never runs;
#   - app-xml.jar: the same advice declared by <aop:config> in aop.xml
#     (and/args in the XML expression) on a plain class — the advice runs.
# An advice method is not a root of the spring profile, so under -roots
# profile it is reachable only through the model's edge from the call of
# the advised method.
#
# The script runs each on the real JVM (the advice prints when it runs),
# then runs bin/ariadne (-profile spring, -roots profile and -roots all;
# OSV through the cache) and prints the model's log line and the verdict
# of CVE-2022-42889.
#
# Produces under work/aop-demo/: app-{advised,unadvised,xml}.jar, lib/*.jar,
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
mkdir -p "$work/src/main/java/demo" "$work/variants"
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

aspect() { # variant method
    mkdir -p "variants/$1/java"
    sed "s/@M@/$2/" > "variants/$1/java/AuditAspect.java" <<'EOF'
package demo;

import org.apache.commons.text.StringSubstitutor;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Before;
import org.springframework.stereotype.Component;

/** Renders an audit line before the advised order operation. */
@Aspect
@Component
public class AuditAspect {
    @Before("execution(* demo.OrderService.@M@(..)) && args(text)")
    public void audit(String text) {
        System.out.println("advice ran: " + StringSubstitutor.createInterpolator().replace(text));
    }
}
EOF
}
aspect advised place
aspect unadvised refund

# The xml variant: a plain class made an aspect by <aop:config>.
mkdir -p variants/xml/java variants/xml/resources
cat > variants/xml/java/AuditAspect.java <<'EOF'
package demo;

import org.apache.commons.text.StringSubstitutor;

/** Renders an audit line; an aspect through aop.xml. */
public class AuditAspect {
    public void audit(String text) {
        System.out.println("advice ran: " + StringSubstitutor.createInterpolator().replace(text));
    }
}
EOF
cat > variants/xml/java/XmlConfig.java <<'EOF'
package demo;

import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.ImportResource;

@Configuration
@ImportResource("classpath:aop.xml")
public class XmlConfig {
}
EOF
cat > variants/xml/resources/aop.xml <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<beans xmlns="http://www.springframework.org/schema/beans"
       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
       xmlns:aop="http://www.springframework.org/schema/aop"
       xsi:schemaLocation="http://www.springframework.org/schema/beans http://www.springframework.org/schema/beans/spring-beans.xsd
                           http://www.springframework.org/schema/aop http://www.springframework.org/schema/aop/spring-aop.xsd">
    <bean id="auditAspect" class="demo.AuditAspect"/>
    <aop:config>
        <aop:aspect ref="auditAspect">
            <aop:before method="audit" pointcut="execution(* demo.OrderService.place(..)) and args(text)"/>
        </aop:aspect>
    </aop:config>
</beans>
EOF

java -version
mvn -q -B dependency:copy-dependencies -DoutputDirectory=lib -DincludeScope=runtime
mvn -q -B dependency:list -DoutputFile=deps.txt -DincludeScope=runtime
# One jar per variant: the shared classes plus that variant's classes and
# resources (zipped with Python, so a JDK without the jar tool is enough;
# with directory entries, without which component scanning finds nothing).
for v in advised unadvised xml; do
    rm -rf src/main/resources src/main/java/demo/XmlConfig.java
    cp variants/$v/java/*.java src/main/java/demo/
    if [ -d variants/$v/resources ]; then cp -r variants/$v/resources src/main/resources; fi
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

for v in advised unadvised xml; do
    echo "== real JVM, $v"
    java -cp "app-$v.jar:lib/*" demo.DemoApplication 2>&1 | grep -E "advice ran" || echo "  (advice did not run)"
done

echo "== ariadne"
cd "$root"
go build -o bin/ariadne ./cmd/ariadne
for v in advised unadvised xml; do
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
