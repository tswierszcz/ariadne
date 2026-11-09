package rtm;

import java.lang.reflect.Method;
import java.net.URL;
import java.net.URLClassLoader;

/** Runs every runtime-evidence case for a known time, so periodic samples see them. */
public final class Main {
    static long until(long ms) {
        return System.nanoTime() + ms * 1_000_000L;
    }

    /** Executed: called through an interface whose other implementation never runs. */
    static long shapes(long ms) {
        Shape s = new Square();
        long acc = 0, end = until(ms);
        while (System.nanoTime() < end) {
            acc += s.area(100_000);
        }
        return acc;
    }

    /** The "vulnerable" method, traced exactly with jdk.MethodTrace. */
    static long handle(String input) {
        return Lib.parse(input);
    }

    /** A reflective call whose method name is built at run time: invisible to the static graph. */
    static long reflective(ClassLoader cl, long ms) throws Exception {
        Class<?> c = Class.forName("rtm.Plugin", true, cl);
        Object p = c.getDeclaredConstructor().newInstance();
        Method m = c.getMethod(new StringBuilder("wo").append("rk").toString(), long.class);
        long acc = 0, end = until(ms);
        while (System.nanoTime() < end) {
            acc += (long) m.invoke(p, 100_000L);
        }
        return acc;
    }

    public static void main(String[] args) throws Exception {
        long acc = shapes(800);
        for (int i = 0; i < 1000; i++) {
            acc += handle("k" + i);
        }
        // loaded, never initialized or executed
        Class.forName("rtm.LoadedOnly", false, Main.class.getClassLoader());
        // the same class under the application loader and under an isolated loader
        acc += reflective(Main.class.getClassLoader(), 600);
        URL jar = Main.class.getProtectionDomain().getCodeSource().getLocation();
        try (URLClassLoader isolated = new URLClassLoader("isolated", new URL[] {jar}, ClassLoader.getPlatformClassLoader())) {
            acc += reflective(isolated, 600);
        }
        acc += Gaps.inherited(500);
        acc += Gaps.sam(500);
        System.out.println(acc);
    }
}
