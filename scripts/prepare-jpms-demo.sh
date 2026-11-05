#!/usr/bin/env bash
# Generates and builds a tiny, hand-written JPMS application whose only
# ServiceLoader provider is declared through module-info `provides` (no
# META-INF/services file) and is created through `public static
# provider()`, which calls org.apache.commons:commons-text:1.9's
# StringSubstitutor#replace (CVE-2022-42889, Text4Shell — a manual-table
# CVE in internal/vulnmap/vuln_methods.yaml).
#
# This is the control app for SERVICE_LOADING S3 slice 2 (-jpms-scenario):
# the same bytecode has different reachability depending on how it is
# launched, which no other control app can show:
#   - -jpms-scenario classpath: on the class path the JVM ignores
#     module-info, ServiceLoader finds no provider, provider() never runs
#     -> CVE-2022-42889 is not reachable from main;
#   - -jpms-scenario module-path: ServiceLoader uses the `provides` entry
#     and calls provider() -> REACHABLE through a model edge.
# Run ariadne with `-profile generic -roots profile` so that main is the
# only root (with -roots all, provider() is itself a root).
#
# The script also launches the app both ways with the real JVM, so the
# expected difference is observed, not assumed.
#
# Produces under work/jpms-demo/:
#   target/classes (with module-info.class), target/dependency/*.jar, deps.txt
#
# Usage: scripts/prepare-jpms-demo.sh [JAVA_HOME for JDK 11+]
set -euo pipefail

JAVA_HOME_ARG="${1:-${JAVA11_HOME:-${JAVA_HOME:-}}}"
if [ -z "$JAVA_HOME_ARG" ]; then
    echo "usage: $0 <path to a JDK 11+ JAVA_HOME>" >&2
    echo "(or set JAVA11_HOME/JAVA_HOME)" >&2
    exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/jpms-demo"
mkdir -p "$work/src/main/java/demo"
cd "$work"

# source/target rather than <release>: some JDK builds (e.g. Debian's
# without ct.sym) reject --release.
cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>jpms-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>org.apache.commons</groupId>
      <artifactId>commons-text</artifactId>
      <version>1.9</version>
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

cat > src/main/java/module-info.java <<'EOF'
module demo.jpms {
    requires org.apache.commons.text;
    uses demo.Codec;
    // Deliberately no META-INF/services/demo.Codec: this provider exists
    // only for a module-path launch.
    provides demo.Codec with demo.InterpolatingCodec;
}
EOF

cat > src/main/java/demo/Codec.java <<'EOF'
package demo;

public interface Codec {
    String encode(String s);
}
EOF

cat > src/main/java/demo/InterpolatingCodec.java <<'EOF'
package demo;

import org.apache.commons.text.StringSubstitutor;

public final class InterpolatingCodec implements Codec {
    private final String prefix;

    private InterpolatingCodec(String prefix) {
        this.prefix = prefix;
    }

    // ServiceLoader's module-path provider method: used instead of a
    // constructor. Interpolating a configurable template is the
    // Text4Shell-vulnerable pattern.
    public static Codec provider() {
        String template = System.getProperty("demo.prefix", "${sys:user.name}");
        return new InterpolatingCodec(StringSubstitutor.createInterpolator().replace(template));
    }

    @Override
    public String encode(String s) {
        return prefix + ":" + s;
    }
}
EOF

cat > src/main/java/demo/Main.java <<'EOF'
package demo;

import java.util.ServiceLoader;

public class Main {
    public static void main(String[] args) {
        int found = 0;
        for (Codec c : ServiceLoader.load(Codec.class)) {
            System.out.println(c.encode("hello"));
            found++;
        }
        System.out.println("providers found: " + found);
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

echo "--- class path launch (module-info ignored; expect 0 providers):"
java -cp "target/classes:target/dependency/*" demo.Main
echo "--- module path launch (expect 1 provider via provider()):"
java --module-path "target/classes:target/dependency" -m demo.jpms/demo.Main

echo "done: $work/{target/classes,target/dependency/,deps.txt}"
