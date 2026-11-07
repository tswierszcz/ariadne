package fs;

import java.io.ByteArrayInputStream;
import java.io.ObjectInputStream;
import java.util.Base64;

/** DATA_FLOW D3 dependency-summary cases (internal/flowsum tests). */
public class Cases {
    static void sink(Object o) {
    }

    // Factory/identity helper called from two sites: no mixing.
    static String id(String s) {
        return s;
    }

    static void twoSites(String a, String b) {
        String x = id(a);
        String y = id(b);
        sink(y);
        sink(x);
    }

    // A transformation through a catalog method: depends on the
    // argument, no new source.
    static byte[] decode(String s) {
        return Base64.getDecoder().decode(s);
    }

    // An unmodeled external method: conservative.
    static Object external(Object o) {
        return ext.Lib.thing(o);
    }

    // Recursion: converges.
    static String rec(String s, int n) {
        if (n == 0) {
            return s;
        }
        return rec(s.concat("x"), n - 1);
    }

    // String concatenation (invokedynamic StringConcatFactory).
    static String concat(String a, int n) {
        return a + "-" + n;
    }

    // A constant: no dependence.
    static String constant(String ignored) {
        return "fixed";
    }

    // Through an object's constructor and getter.
    static String holder(String p) {
        return new Holder(p).get();
    }

    // Through a single implementation (CHA) and two (widened).
    static String single(Codec c, String s) {
        return c.apply(s);
    }

    static String multi(Shape sh, String s) {
        return sh.name(s);
    }

    // A deserialization entry fed by a parameter.
    static Object read(byte[] data) throws Exception {
        return new ObjectInputStream(new ByteArrayInputStream(data)).readObject();
    }

    static Object readDecoded(String b64) throws Exception {
        return read(decode(b64));
    }

    // A user method that mutates an argument object.
    static void fill(StringBuilder sb, String s) {
        sb.append(s);
    }

    static String filled(String p) {
        StringBuilder sb = new StringBuilder();
        fill(sb, p);
        return sb.toString();
    }

    // A field written from a parameter, read elsewhere.
    static String shared;

    static void store(String v) {
        shared = v;
    }

    static String load() {
        return shared;
    }

    // An exception handler: the value assigned before the throw flows on.
    static String guarded(String p) {
        String r = "none";
        try {
            r = p;
            Integer.parseInt(p);
        } catch (NumberFormatException e) {
            return r;
        }
        return "ok";
    }
}
