#!/usr/bin/env bash
# Generates, builds and runs a tiny application whose deserialization
# entries get their data from different places — the control app for
# DATA_FLOW D4's taint pilot (request data → the data a deserialization
# entry reads). No dependencies: the few Spring MVC / Servlet types the
# taint catalog names (annotations, HtmlUtils, ServletRequest) are small
# stubs with the real names, so the demo builds and runs offline; built
# with scripts/JavaBuild.java into work/taint-demo/app.jar.
#
# Entries (demo.Endpoints, each reads one value; demo.Codec.read for the
# first) and the taint result ariadne is expected to report:
#   userData(@RequestParam)  token → Base64 helper → Codec.read     → flow_found, guard none (2 call levels)
#   constant(@RequestParam)  a literal serialized string               → no_flow_in_model
#   twoSites(@RequestParam)  one helper called with the token and with
#                            a literal; only the literal is read        → no_flow_in_model
#   filtered(@RequestParam)  literal ObjectInputFilter allowing only
#                            java.lang.String (rejects demo.Payload)    → flow_found, guard effective
#   escaped(@RequestParam)   isEmpty() check + HtmlUtils.htmlEscape     → flow_found, guard unrelated
#   unknown(@RequestParam)   bytes from an interface with two
#                            implementations (not resolved)             → undetermined
#   servlet(ServletRequest)  req.getParameter("t")                      → flow_found (src:servlet-request)
#   implicit(String)         unannotated String of a handler            → flow_found (src:spring-mvc-implicit)
#   fromCli(String)          called only from main with args[0]         → no_flow_in_model, ends at main's parameter
# The run prints each value read, so the app is shown to work.
#
# Usage: scripts/prepare-taint-demo.sh [JAVA_HOME]
#   then: bin/ariadne -app work/taint-demo/app.jar -profile generic -roots profile -offline -taint \
#         -out work/taint-demo/report.md -json work/taint-demo/results.json -json-format v2
set -euo pipefail
JAVA_HOME_ARG="${1:-${JAVA_HOME:-}}"
if [ -n "$JAVA_HOME_ARG" ]; then
    export JAVA_HOME="$JAVA_HOME_ARG"
    export PATH="$JAVA_HOME/bin:$PATH"
fi
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$root/work/taint-demo"
rm -rf "$work/src"
mkdir -p "$work/src/demo" "$work/src/org/springframework/web/bind/annotation" "$work/src/org/springframework/web/util" "$work/src/javax/servlet"

for a in RequestParam PostMapping; do
    target=PARAMETER
    [ "$a" = PostMapping ] && target=METHOD
    cat > "$work/src/org/springframework/web/bind/annotation/$a.java" <<J
package org.springframework.web.bind.annotation;

import java.lang.annotation.*;

/** Stub of the Spring MVC annotation (same name, for the demo only). */
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.$target)
public @interface $a {
    String[] value() default {};
}
J
done
cat > "$work/src/org/springframework/web/util/HtmlUtils.java" <<'J'
package org.springframework.web.util;

/** Stub of Spring's HtmlUtils (for the demo only). */
public final class HtmlUtils {
    public static String htmlEscape(String s) {
        return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;").replace("\"", "&quot;");
    }
}
J
cat > "$work/src/javax/servlet/ServletRequest.java" <<'J'
package javax.servlet;

/** Stub of the Servlet API request (for the demo only). */
public interface ServletRequest {
    String getParameter(String name);
}
J
cat > "$work/src/demo/FakeRequest.java" <<'J'
package demo;

import javax.servlet.ServletRequest;

final class FakeRequest implements ServletRequest {
    private final String value;
    FakeRequest(String value) { this.value = value; }
    public String getParameter(String name) { return value; }
}
J
cat > "$work/src/demo/Payload.java" <<'J'
package demo;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.Serializable;

public class Payload implements Serializable {
    private static final long serialVersionUID = 1L;
    String note;
    private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        in.defaultReadObject();
        System.out.println("  Payload.readObject ran");
    }
}
J
cat > "$work/src/demo/Provider.java" <<'J'
package demo;

interface Provider {
    byte[] bytes();
}

final class LiteralProvider implements Provider {
    public byte[] bytes() { return Codec.decode(Endpoints.HELLO); }
}

final class OtherProvider implements Provider {
    public byte[] bytes() { return Codec.decode(Endpoints.HELLO); }
}
J
cat > "$work/src/demo/Codec.java" <<'J'
package demo;

import java.io.ByteArrayInputStream;
import java.io.ObjectInputStream;
import java.util.Base64;

final class Codec {
    static byte[] decode(String s) {
        return Base64.getDecoder().decode(s);
    }

    static Object read(byte[] b) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(b))) {
            return in.readObject();
        }
    }
}
J
cat > "$work/src/demo/Endpoints.java" <<'J'
package demo;

import java.io.ByteArrayInputStream;
import java.io.ObjectInputFilter;
import java.io.ObjectInputStream;
import javax.servlet.ServletRequest;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.util.HtmlUtils;

public class Endpoints {
    /** Java serialization of the String "hello". */
    static final String HELLO = "rO0ABXQABWhlbGxv";

    @PostMapping("/user-data")
    public Object userData(@RequestParam("token") String token) throws Exception {
        return Codec.read(Codec.decode(token));
    }

    @PostMapping("/constant")
    public Object constant(@RequestParam("ignored") String ignored) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode("rO0ABXQABWhlbGxv")))) {
            return in.readObject();
        }
    }

    @PostMapping("/two-sites")
    public Object twoSites(@RequestParam("token") String token) throws Exception {
        byte[] fromRequest = Codec.decode(token);
        byte[] literal = Codec.decode("rO0ABXQABWhlbGxv");
        System.out.println("  request bytes: " + fromRequest.length);
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(literal))) {
            return in.readObject();
        }
    }

    @PostMapping("/filtered")
    public Object filtered(@RequestParam("token") String token) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(token)))) {
            in.setObjectInputFilter(ObjectInputFilter.Config.createFilter("java.lang.String;!*"));
            return in.readObject();
        }
    }

    @PostMapping("/escaped")
    public Object escaped(@RequestParam("token") String token) throws Exception {
        if (token.isEmpty()) {
            return null;
        }
        String clean = HtmlUtils.htmlEscape(token);
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(clean)))) {
            return in.readObject();
        }
    }

    @PostMapping("/unknown")
    public Object unknown(@RequestParam("token") String token) throws Exception {
        Provider p = token.isEmpty() ? new LiteralProvider() : new OtherProvider();
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(p.bytes()))) {
            return in.readObject();
        }
    }

    @PostMapping("/servlet")
    public Object servlet(ServletRequest req) throws Exception {
        String t = req.getParameter("t");
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(t)))) {
            return in.readObject();
        }
    }

    @PostMapping("/implicit")
    public Object implicit(String token) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(token)))) {
            return in.readObject();
        }
    }

    static Object fromCli(String arg) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(arg)))) {
            return in.readObject();
        }
    }

    public static void main(String[] args) throws Exception {
        Endpoints e = new Endpoints();
        java.io.ByteArrayOutputStream bos = new java.io.ByteArrayOutputStream();
        try (java.io.ObjectOutputStream out = new java.io.ObjectOutputStream(bos)) {
            out.writeObject(new Payload());
        }
        String payload = java.util.Base64.getEncoder().encodeToString(bos.toByteArray());
        System.out.println("userData: " + e.userData(payload).getClass().getName());
        System.out.println("constant: " + e.constant(payload));
        System.out.println("twoSites: " + e.twoSites(payload));
        try {
            e.filtered(payload);
            System.out.println("filtered: Payload accepted (unexpected)");
        } catch (java.io.InvalidClassException rejected) {
            System.out.println("filtered: Payload rejected by the filter (" + rejected.getMessage() + ")");
        }
        System.out.println("filtered: " + e.filtered(HELLO));
        System.out.println("escaped: " + e.escaped(HELLO));
        System.out.println("unknown: " + e.unknown(payload));
        System.out.println("servlet: " + e.servlet(new FakeRequest(HELLO)));
        System.out.println("implicit: " + e.implicit(HELLO));
        System.out.println("fromCli: " + fromCli(args.length > 0 ? args[0] : HELLO));
    }
}
J
java "$root/scripts/JavaBuild.java" "$work/src" "$work/app.jar" 11
java -cp "$work/app.jar" demo.Endpoints | tee "$work/run.txt"
echo "done: $work/app.jar"
