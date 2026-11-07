#!/usr/bin/env bash
# Generates, builds and probes a tiny application with the libraries of
# DESERIALIZATION Z3a's chain catalog — the control app for candidate
# gadget chains. Built twice:
#   work/chains-demo-vulnerable/  commons-collections 3.2.1, commons-collections4 4.0,
#                                 groovy 2.4.3, commons-beanutils 1.9.4, XStream 1.4.6
#   work/chains-demo-fixed/       commons-collections 3.2.2, commons-collections4 4.1,
#                                 groovy 2.4.4, commons-beanutils 1.9.4, XStream 1.4.7
#
# Entries (demo.Main), none of them fed by a request — the data flow is
# not analyzed anyway (DATA_FLOW:D4):
#   read          new ObjectInputStream(...).readObject()           policy none_seen
#   readFiltered  the same, with setObjectInputFilter(createFilter("demo.*;!*"))
#   readXml       new XStream().fromXML(...)
#
# No gadget chain is ever built or run. The probe only round-trips, on
# the real JVM, the single class each catalog guard is about, harmless
# by itself, to observe that the guard ariadne reads from the bytecode
# is what the release does:
#   cc3     an InvokerTransformer("toString") written with
#           enableUnsafeSerialization=true, read without it
#           (3.2.1: read; 3.2.2: UnsupportedOperationException)
#   cc4     writing an InvokerTransformer (4.0: written; 4.1: NotSerializableException)
#   groovy  a MethodClosure("abc", "length") (2.4.3: read; 2.4.4: UnsupportedOperationException)
#   xstream an XML java.beans.EventHandler, never invoked
#           (1.4.6: an EventHandler; 1.4.7: ConversionException)
#   beanutils a BeanComparator round-trip (both: read — no guard exists)
#
# Run ariadne on each with
#   -app target/classes -deps target/dependency -deps-list deps.txt -profile generic -roots profile
# (and -target-java 17 to see the JDK-bound chains blocked).
#
# Usage: scripts/prepare-deserialization-chains-demo.sh [JAVA_HOME]
set -euo pipefail
JAVA_HOME_ARG="${1:-${JAVA_HOME:-}}"
if [ -n "$JAVA_HOME_ARG" ]; then
    export JAVA_HOME="$JAVA_HOME_ARG"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
jvm_opts=(--add-opens java.base/java.util=ALL-UNNAMED --add-opens java.base/java.lang=ALL-UNNAMED
    --add-opens java.base/java.lang.reflect=ALL-UNNAMED --add-opens java.base/java.text=ALL-UNNAMED
    --add-opens java.base/java.io=ALL-UNNAMED --add-opens java.desktop/java.beans=ALL-UNNAMED --add-opens java.desktop/java.awt.font=ALL-UNNAMED)

build() { # variant cc3 cc4 groovy xstream
    local variant=$1 cc3=$2 cc4=$3 groovy=$4 xstream=$5 work="$root/work/chains-demo-$1"
    rm -rf "$work/src"
    mkdir -p "$work/src/main/java/demo"
    cd "$work"
    cat > pom.xml <<EOF
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>chains-demo</artifactId>
  <version>1.0</version>
  <properties>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency><groupId>commons-collections</groupId><artifactId>commons-collections</artifactId><version>$cc3</version></dependency>
    <dependency><groupId>org.apache.commons</groupId><artifactId>commons-collections4</artifactId><version>$cc4</version></dependency>
    <dependency><groupId>org.codehaus.groovy</groupId><artifactId>groovy</artifactId><version>$groovy</version></dependency>
    <dependency><groupId>commons-beanutils</groupId><artifactId>commons-beanutils</artifactId><version>1.9.4</version></dependency>
    <dependency><groupId>com.thoughtworks.xstream</groupId><artifactId>xstream</artifactId><version>$xstream</version></dependency>
  </dependencies>
  <build><plugins><plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-compiler-plugin</artifactId>
    <version>3.11.0</version>
  </plugin></plugins></build>
</project>
EOF
    cat > src/main/java/demo/Main.java <<'EOF'
package demo;

import com.thoughtworks.xstream.XStream;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.ObjectInputFilter;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;

public class Main {
    public static void main(String[] args) throws Exception {
        if (args.length > 0 && args[0].equals("probe")) {
            Probe.run();
            return;
        }
        // The entries, on harmless data.
        System.out.println(read(bytes("hello")));
        System.out.println(readFiltered(bytes("hello")));
        System.out.println(readXml("<string>hello</string>"));
    }

    static byte[] bytes(Object o) throws Exception {
        ByteArrayOutputStream b = new ByteArrayOutputStream();
        try (ObjectOutputStream out = new ObjectOutputStream(b)) {
            out.writeObject(o);
        }
        return b.toByteArray();
    }

    static Object read(byte[] data) throws Exception {
        ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(data));
        return in.readObject();
    }

    static Object readFiltered(byte[] data) throws Exception {
        ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(data));
        in.setObjectInputFilter(ObjectInputFilter.Config.createFilter("demo.*;!*"));
        return in.readObject();
    }

    static Object readXml(String xml) {
        XStream x = new XStream();
        return x.fromXML(xml);
    }
}
EOF
    cat > src/main/java/demo/Probe.java <<'EOF'
package demo;

import com.thoughtworks.xstream.XStream;

// Round-trips the one class each catalog guard is about. Nothing here
// builds or runs a gadget chain.
class Probe {
    static void run() {
        probe("cc3", () -> {
            Object t = new org.apache.commons.collections.functors.InvokerTransformer("toString", new Class[0], new Object[0]);
            System.setProperty("org.apache.commons.collections.enableUnsafeSerialization", "true");
            byte[] b = Main.bytes(t);
            System.clearProperty("org.apache.commons.collections.enableUnsafeSerialization");
            return Main.read(b);
        });
        probe("cc4", () -> Main.bytes(new org.apache.commons.collections4.functors.InvokerTransformer<Object, Object>("toString", new Class[0], new Object[0])));
        probe("groovy", () -> Main.read(Main.bytes(new org.codehaus.groovy.runtime.MethodClosure("abc", "length"))));
        probe("xstream", () -> new XStream().fromXML(
            "<java.beans.EventHandler><target class=\"string\">abc</target><action>length</action></java.beans.EventHandler>"));
        probe("beanutils", () -> Main.read(Main.bytes(new org.apache.commons.beanutils.BeanComparator<Object>("x", String.CASE_INSENSITIVE_ORDER))));
    }

    interface Step {
        Object call() throws Exception;
    }

    static void probe(String name, Step s) {
        try {
            Object o = s.call();
            System.out.println(name + ": ok " + (o instanceof byte[] ? "written" : o.getClass().getName()));
        } catch (Throwable e) {
            Throwable c = e;
            while (c.getCause() != null && c.getCause() != c) {
                c = c.getCause();
            }
            System.out.println(name + ": refused " + e.getClass().getSimpleName() + (c != e ? " (" + c.getClass().getSimpleName() + ")" : ""));
        }
    }
}
EOF
    mvn -q package -DskipTests
    mvn -q dependency:copy-dependencies -DoutputDirectory=target/dependency
    mvn -q dependency:list -DoutputFile=deps.txt
    sed -i 's/\x1b\[[0-9;]*m//g' deps.txt
    local cp="target/classes:target/dependency/*"
    echo "--- $variant: entries"
    java "${jvm_opts[@]}" -cp "$cp" demo.Main 2>/dev/null
    echo "--- $variant: probe"
    java "${jvm_opts[@]}" -cp "$cp" demo.Main probe 2>/dev/null | tee "$work/probe.txt"
    expect() { # pattern
        grep -q "^$1" "$work/probe.txt" || { echo "$variant: expected '$1'" >&2; exit 1; }
    }
    case $variant in
        vulnerable)
            expect "cc3: ok"; expect "cc4: ok written"; expect "groovy: ok"; expect "xstream: ok java.beans.EventHandler"; expect "beanutils: ok" ;;
        fixed)
            expect "cc3: refused UnsupportedOperationException"; expect "cc4: refused NotSerializableException"
            expect "groovy: refused UnsupportedOperationException"; expect "xstream: refused"; expect "beanutils: ok" ;;
    esac
    echo "done: $work"
}

java -version 2>&1 | head -1
build vulnerable 3.2.1 4.0 2.4.3 1.4.6
build fixed 3.2.2 4.1 2.4.4 1.4.7
