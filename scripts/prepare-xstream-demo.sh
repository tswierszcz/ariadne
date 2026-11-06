#!/usr/bin/env bash
# Generates, builds and runs a tiny application that reads XML with
# XStream#fromXML — the control app for DESERIALIZATION Z2's XStream
# model. Built three times, against XStream 1.4.5 (no security
# framework), 1.4.10 (security framework, any type allowed by default)
# and 1.4.20 (default allowlist), into work/xstream-demo-<version>/.
#
# Classes read:
#   demo.Gadget   Serializable, private readObject  (SerializableConverter)
#   demo.Plain    not Serializable, private readResolve
#   demo.Ext      Externalizable: public constructor + readExternal
# Entries (demo.Main): read <xml> — new XStream(), nothing configured;
# and, from 1.4.7 on (the API does not exist before), demo.Configured:
#   allowOne    — allowTypes(new Class[]{Gadget.class}) only
#   noneAllowOne — addPermission(NoTypePermission.NONE) + allowTypes(Gadget)
# Each run prints which callbacks ran and which types XStream refused, so
# the expected edges per release are observed, not assumed.
#
# Run ariadne with `-profile generic -roots profile` on each directory.
#
# Usage: scripts/prepare-xstream-demo.sh [JAVA_HOME]
set -euo pipefail
JAVA_HOME_ARG="${1:-${JAVA_HOME:-}}"
if [ -n "$JAVA_HOME_ARG" ]; then
    export JAVA_HOME="$JAVA_HOME_ARG"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# XStream 1.4.x reflects into JDK internals; on JDK 16+ that needs opens.
jvm_opts=(--add-opens java.base/java.util=ALL-UNNAMED --add-opens java.base/java.lang=ALL-UNNAMED
    --add-opens java.base/java.lang.reflect=ALL-UNNAMED --add-opens java.base/java.text=ALL-UNNAMED
    --add-opens java.base/java.io=ALL-UNNAMED --add-opens java.desktop/java.awt.font=ALL-UNNAMED)

build() { # version
    local v=$1 work="$root/work/xstream-demo-$1"
    rm -rf "$work/src"
    mkdir -p "$work/src/main/java/demo"
    cd "$work"
    cat > pom.xml <<EOF
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>xstream-demo</artifactId>
  <version>1.0</version>
  <properties>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>com.thoughtworks.xstream</groupId>
      <artifactId>xstream</artifactId>
      <version>$v</version>
    </dependency>
  </dependencies>
  <build><plugins><plugin>
    <groupId>org.apache.maven.plugins</groupId>
    <artifactId>maven-compiler-plugin</artifactId>
    <version>3.11.0</version>
  </plugin></plugins></build>
</project>
EOF
    cat > src/main/java/demo/Sink.java <<'EOF'
package demo;

public final class Sink {
    static void touch(String who) {
        System.out.println("callback ran: " + who);
    }
}
EOF
    cat > src/main/java/demo/Gadget.java <<'EOF'
package demo;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.Serializable;

public class Gadget implements Serializable {
    private static final long serialVersionUID = 1L;
    int n = 1;

    private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        in.defaultReadObject();
        Sink.touch("Gadget.readObject");
    }
}
EOF
    cat > src/main/java/demo/Plain.java <<'EOF'
package demo;

public class Plain {
    int n = 2;

    private Object readResolve() {
        Sink.touch("Plain.readResolve");
        return this;
    }
}
EOF
    cat > src/main/java/demo/Ext.java <<'EOF'
package demo;

import java.io.Externalizable;
import java.io.ObjectInput;
import java.io.ObjectOutput;

public class Ext implements Externalizable {
    public Ext() {
        Sink.touch("Ext.<init>");
    }

    @Override
    public void writeExternal(ObjectOutput out) {
    }

    @Override
    public void readExternal(ObjectInput in) {
        Sink.touch("Ext.readExternal");
    }
}
EOF
    cat > src/main/java/demo/Main.java <<'EOF'
package demo;

import com.thoughtworks.xstream.XStream;

public class Main {
    public static void main(String[] args) throws Exception {
        String mode = args[0];
        if (mode.equals("write")) {
            // Writing runs no read callback (Ext's constructor runs to create it).
            System.out.println(writer().toXML(new Object[] {new Gadget(), new Plain(), new Ext()}).replace("\n", ""));
            return;
        }
        String xml = args[1];
        if (mode.equals("read")) {
            read(xml);
        } else {
            Configured.read(mode, xml);
        }
    }

    static XStream writer() {
        return new XStream();
    }

    // Each entry configures and reads in the same method (ariadne
    // evaluates a policy only there, on the same instance).
    static void read(String xml) {
        XStream x = new XStream();
        try {
            x.fromXML(xml);
            System.out.println("read ok");
        } catch (RuntimeException e) {
            refused(e);
        }
    }

    static void refused(RuntimeException e) {
        System.out.println("refused: " + e.getClass().getSimpleName() + ": " + firstLine(e.getMessage()));
    }

    static String firstLine(String s) {
        return s == null ? "" : s.split("\n")[0];
    }
}
EOF
    if [ "$(printf '%s\n' 1.4.7 "$v" | sort -V | head -1)" = 1.4.7 ]; then
        cat > src/main/java/demo/Configured.java <<'EOF'
package demo;

import com.thoughtworks.xstream.XStream;
import com.thoughtworks.xstream.security.NoTypePermission;

class Configured {
    static void read(String mode, String xml) {
        try {
            if (mode.equals("allowOne")) {
                allowOne(xml);
            } else if (mode.equals("noneAllowOne")) {
                noneAllowOne(xml);
            }
            System.out.println("read ok");
        } catch (RuntimeException e) {
            Main.refused(e);
        }
    }

    static Object allowOne(String xml) {
        XStream x = new XStream();
        x.allowTypes(new Class[] {Gadget.class});
        return x.fromXML(xml);
    }

    static Object noneAllowOne(String xml) {
        XStream x = new XStream();
        x.addPermission(NoTypePermission.NONE);
        x.allowTypes(new Class[] {Gadget.class, Object[].class});
        return x.fromXML(xml);
    }
}
EOF
    else
        cat > src/main/java/demo/Configured.java <<'EOF'
package demo;

class Configured {
    static void read(String mode, String xml) {
        System.out.println("no security API in this XStream release");
    }
}
EOF
    fi
    mvn -q package -DskipTests
    mvn -q dependency:copy-dependencies -DoutputDirectory=target/dependency
    mvn -q dependency:list -DoutputFile=deps.txt
    sed -i 's/\x1b\[[0-9;]*m//g' deps.txt
    local cp="target/classes:target/dependency/*"
    local xml
    xml="$(java "${jvm_opts[@]}" -cp "$cp" demo.Main write 2>/dev/null | grep '<object-array')"
    [ -n "$xml" ] || { echo "XStream $v: writing the XML failed" >&2; exit 1; }
    for mode in read allowOne noneAllowOne; do
        echo "--- XStream $v, $mode:"
        java "${jvm_opts[@]}" -cp "$cp" demo.Main "$mode" "$xml" 2>/dev/null | grep -v "^$" | tee "$work/run-$mode.txt" || true
    done
    # The expectations ariadne's model encodes, checked on the real run.
    expect() { # mode pattern present(1)/absent(0)
        if grep -q "$2" "$work/run-$1.txt"; then [ "$3" = 1 ]; else [ "$3" = 0 ]; fi ||
            { echo "XStream $v $1: expected '$2' $( [ "$3" = 1 ] && echo present || echo absent)" >&2; exit 1; }
    }
    case $v in
        1.4.5)
            for cb in Gadget.readObject Plain.readResolve Ext.\<init\> Ext.readExternal; do expect read "$cb" 1; done ;;
        1.4.10)
            for m in read allowOne; do for cb in Gadget.readObject Plain.readResolve Ext.readExternal; do expect $m "$cb" 1; done; done
            expect noneAllowOne Gadget.readObject 1; expect noneAllowOne Plain.readResolve 0 ;;
        1.4.20)
            expect read "callback ran" 0; expect read ForbiddenClassException 1
            for m in allowOne noneAllowOne; do expect $m Gadget.readObject 1; expect $m Plain.readResolve 0; done ;;
    esac
    echo "done: $work"
}

java -version 2>&1 | head -1
for v in 1.4.5 1.4.10 1.4.20; do
    build "$v"
done
