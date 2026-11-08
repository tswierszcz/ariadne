package fx.app;

import fx.trace.Ran;
import javax.ariadnefx.Codec;
import javax.ariadnefx.Codecs;

public final class Main {
    static String run(String s) {
        Codec c = Codecs.open();
        return c.encode(s);
    }

    public static void main(String[] args) {
        System.out.println("# " + run("ariadne"));
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
