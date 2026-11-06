#!/usr/bin/env bash
# Generates and builds a tiny, hand-written application that reads an
# object stream with java.io.ObjectInputStream — the control app for
# DESERIALIZATION Z1 (Java serialization callbacks).
#
# The stream it writes and reads holds:
#   - demo.Gadget: Serializable with a private readObject(ObjectInputStream)
#     that calls demo.Sink#touch — serialization runs it;
#   - demo.WrongSig: a *public* readObject(ObjectInputStream) — not a
#     serialization callback, never run by readObject;
#   - demo.Ext: Externalizable — its public no-arg constructor and
#     readExternal run;
#   - demo.Resolving: a readResolve() — runs after the object is read;
#   - ch.qos.logback.classic.spi.LoggingEventVO (logback-classic 1.2.3):
#     its private readObject is the code of CVE-2023-6378.
# Nothing in the application calls any of those methods directly: before
# DESERIALIZATION Z1 every one of them was unreachable from main.
#
# The script also runs the app with the real JVM (write, then read) and
# prints which callbacks actually ran, so the expected edges are
# observed, not assumed.
#
# Run ariadne with `-profile generic -roots profile` (main is the root).
#
# Produces under work/deserialization-demo/:
#   target/classes, target/dependency/*.jar, deps.txt
#
# Usage: scripts/prepare-deserialization-demo.sh [JAVA_HOME]
set -euo pipefail
JAVA_HOME_ARG="${1:-${JAVA_HOME:-}}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/deserialization-demo"
mkdir -p "$work/src/main/java/demo"
cd "$work"

cat > pom.xml <<'EOF'
<project xmlns="http://maven.apache.org/POM/4.0.0">
  <modelVersion>4.0.0</modelVersion>
  <groupId>demo</groupId>
  <artifactId>deserialization-demo</artifactId>
  <version>1.0</version>
  <packaging>jar</packaging>
  <properties>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
  </properties>
  <dependencies>
    <dependency>
      <groupId>ch.qos.logback</groupId>
      <artifactId>logback-classic</artifactId>
      <version>1.2.3</version>
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

    private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        in.defaultReadObject();
        Sink.touch("Gadget.readObject");
    }
}
EOF

cat > src/main/java/demo/WrongSig.java <<'EOF'
package demo;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.Serializable;

public class WrongSig implements Serializable {
    private static final long serialVersionUID = 1L;

    // public: not a serialization callback (it must be private).
    public void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        Sink.touch("WrongSig.readObject (must not run)");
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

cat > src/main/java/demo/Resolving.java <<'EOF'
package demo;

import java.io.Serializable;

public class Resolving implements Serializable {
    private static final long serialVersionUID = 1L;

    private Object readResolve() {
        Sink.touch("Resolving.readResolve");
        return this;
    }
}
EOF

cat > src/main/java/demo/Main.java <<'EOF'
package demo;

import ch.qos.logback.classic.Level;
import ch.qos.logback.classic.LoggerContext;
import ch.qos.logback.classic.spi.LoggingEvent;
import ch.qos.logback.classic.spi.LoggingEventVO;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;

public class Main {
    public static void main(String[] args) throws Exception {
        if (args.length == 2 && args[0].equals("write")) {
            write(args[1]);
        } else if (args.length == 2 && args[0].equals("read")) {
            read(args[1]);
        } else {
            System.err.println("usage: write|read FILE");
        }
    }

    // Writing never runs a read callback.
    static void write(String file) throws Exception {
        LoggerContext ctx = new LoggerContext();
        LoggingEvent ev = new LoggingEvent("demo", ctx.getLogger("demo"), Level.INFO, "hello", null, null);
        try (ObjectOutputStream out = new ObjectOutputStream(new FileOutputStream(file))) {
            out.writeObject(new Gadget());
            out.writeObject(new WrongSig());
            out.writeObject(new Ext());
            out.writeObject(new Resolving());
            out.writeObject(LoggingEventVO.build(ev));
        }
    }

    static void read(String file) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new FileInputStream(file))) {
            for (int i = 0; i < 5; i++) {
                Object o = in.readObject();
                System.out.println("read " + o.getClass().getName());
            }
        }
    }
}
EOF

if [ -n "$JAVA_HOME_ARG" ]; then
    export JAVA_HOME="$JAVA_HOME_ARG"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
java -version

mvn -q package -DskipTests
mvn -q dependency:copy-dependencies -DoutputDirectory=target/dependency
mvn -q dependency:list -DoutputFile=deps.txt
sed -i 's/\x1b\[[0-9;]*m//g' deps.txt

stream="$(mktemp)"
trap 'rm -f "$stream"' EXIT
echo "--- write (expect only Ext.<init> from constructing the object to write):"
java -cp "target/classes:target/dependency/*" demo.Main write "$stream"
echo "--- read (expect Gadget.readObject, Ext.<init>, Ext.readExternal, Resolving.readResolve; never WrongSig):"
out="$(java -cp "target/classes:target/dependency/*" demo.Main read "$stream")"
echo "$out"
for want in "Gadget.readObject" "Ext.<init>" "Ext.readExternal" "Resolving.readResolve"; do
    grep -q "callback ran: $want" <<<"$out" || { echo "expected callback $want did not run" >&2; exit 1; }
done
if grep -q "must not run" <<<"$out"; then
    echo "unexpected: WrongSig.readObject ran" >&2
    exit 1
fi
grep -q "read ch.qos.logback.classic.spi.LoggingEventVO" <<<"$out" || { echo "LoggingEventVO not read" >&2; exit 1; }

echo "done: $work/{target/classes,target/dependency/,deps.txt}"
