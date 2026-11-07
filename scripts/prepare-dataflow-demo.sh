#!/usr/bin/env bash
# Generates, builds and runs a tiny application whose deserialization
# entries read data reaching them in different ways — the control app for
# DATA_FLOW D3's dependency summaries (what the data an entry reads
# depends on in its own method). No dependencies; built with
# scripts/JavaBuild.java into work/dataflow-demo/app.jar.
#
# Entries (demo.Api, each reads one Serializable value) and what ariadne's
# "Data read depends on" column is expected to say:
#   fromBytes(byte[])        the stream wraps the parameter       → parameter 0 (byte[])
#   fromBase64(String)       Base64 through a helper              → parameter 0 (java.lang.String)
#   fromField()              a stream kept in a field             → this, field demo.Api.kept
#   fromConstant()           a constant byte array                → — (nothing)
#   fromSource(Source, int)  an interface with two implementations → parameter 0, parameter 1,
#                                                                    result of demo.Source#bytes (conservative)
#   twoSites(String,String)  one helper called twice              → parameter 1 only
# The run prints each value read, so the app is shown to work.
#
# Usage: scripts/prepare-dataflow-demo.sh [JAVA_HOME]
#   then: bin/ariadne -app work/dataflow-demo/app.jar -profile generic -roots profile -offline \
#         -out work/dataflow-demo/report.md -json work/dataflow-demo/results.json -json-format v2
set -euo pipefail
JAVA_HOME_ARG="${1:-${JAVA_HOME:-}}"
if [ -n "$JAVA_HOME_ARG" ]; then
    export JAVA_HOME="$JAVA_HOME_ARG"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/dataflow-demo"
rm -rf "$work/src"
mkdir -p "$work/src/demo"
cat > "$work/src/demo/Codec.java" <<'J'
package demo;

import java.util.Base64;

final class Codec {
    static byte[] decode(String s) {
        return Base64.getDecoder().decode(s);
    }

    static String pick(String s) {
        return s;
    }
}
J
cat > "$work/src/demo/Source.java" <<'J'
package demo;

interface Source {
    byte[] bytes(int n);
}

class Stored implements Source {
    private final byte[] data;

    Stored(byte[] data) {
        this.data = data;
    }

    public byte[] bytes(int n) {
        return data;
    }
}

class Empty implements Source {
    public byte[] bytes(int n) {
        return new byte[n];
    }
}
J
cat > "$work/src/demo/Api.java" <<'J'
package demo;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;

public class Api {
    private final InputStream kept;

    Api(byte[] data) {
        this.kept = new ByteArrayInputStream(data);
    }

    static Object fromBytes(byte[] data) throws Exception {
        return new ObjectInputStream(new ByteArrayInputStream(data)).readObject();
    }

    static Object fromBase64(String b64) throws Exception {
        return new ObjectInputStream(new ByteArrayInputStream(Codec.decode(b64))).readObject();
    }

    Object fromField() throws Exception {
        return new ObjectInputStream(kept).readObject();
    }

    static Object fromConstant() throws Exception {
        byte[] b = {(byte) 0xac, (byte) 0xed, 0, 5, 0x74, 0, 2, 'h', 'i'};
        return new ObjectInputStream(new ByteArrayInputStream(b)).readObject();
    }

    static Object fromSource(Source src, int n) throws Exception {
        return new ObjectInputStream(new ByteArrayInputStream(src.bytes(n))).readObject();
    }

    static Object twoSites(String ignored, String b64) throws Exception {
        String a = Codec.pick(ignored);
        String b = Codec.pick(b64);
        System.out.println("ignored " + a.length());
        return new ObjectInputStream(new ByteArrayInputStream(Codec.decode(b))).readObject();
    }

    public static void main(String[] args) throws Exception {
        ByteArrayOutputStream out = new ByteArrayOutputStream();
        try (ObjectOutputStream o = new ObjectOutputStream(out)) {
            o.writeObject("hello");
        }
        byte[] data = out.toByteArray();
        String b64 = java.util.Base64.getEncoder().encodeToString(data);
        System.out.println("fromBytes " + fromBytes(data));
        System.out.println("fromBase64 " + fromBase64(b64));
        System.out.println("fromField " + new Api(data).fromField());
        System.out.println("fromConstant " + fromConstant());
        System.out.println("fromSource " + fromSource(new Stored(data), 0));
        System.out.println("twoSites " + twoSites("x", b64));
    }
}
J
java "$root/scripts/JavaBuild.java" "$work/src" "$work/app.jar" 11
java -cp "$work/app.jar" demo.Api | tee "$work/run.txt"
for m in fromBytes fromBase64 fromField fromConstant fromSource twoSites; do
    grep -q "^$m " "$work/run.txt" || { echo "$m did not run" >&2; exit 1; }
done
echo "done: $work/app.jar"
