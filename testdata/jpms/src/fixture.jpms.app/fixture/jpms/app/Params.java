package fixture.jpms.app;

import fixture.jpms.api.Codec;
import java.util.ServiceLoader;
import java.util.function.Consumer;

// DATA_FLOW D3 slice "parameters from callers": ServiceLoader arguments
// that are parameters of the method calling load().
public class Params {
    // Two callers pass different Class literals: a candidate set.
    private static <T> ServiceLoader<T> loadFor(Class<T> c) {
        return ServiceLoader.load(c);
    }

    public static void viaCodec() {
        for (Codec c : loadFor(Codec.class)) {
            System.out.println(c.encode("p"));
        }
    }

    public static void viaRunnable() {
        for (Runnable r : loadFor(Runnable.class)) {
            r.run();
        }
    }

    // Jackson's ObjectMapper.findModules()/secureGetServiceLoader shape:
    // the loader is null two caller levels up.
    public static void modules() {
        findAll(null);
    }

    public static void findAll(ClassLoader cl) {
        for (Codec c : secure(Codec.class, cl)) {
            System.out.println(c.encode("m"));
        }
    }

    private static ServiceLoader<Codec> secure(Class<Codec> k, ClassLoader cl) {
        return ServiceLoader.load(k, cl);
    }

    // Self-recursion converges to the non-recursive caller's value.
    public static void recEntry() {
        rec(Codec.class, 3);
    }

    private static void rec(Class<?> c, int n) {
        if (n > 0) {
            rec(c, n - 1);
        }
        ServiceLoader.load(c).iterator();
    }

    // Used as a method reference: invokedynamic callers are not traced.
    public static void lambdaUser() {
        Consumer<Class<?>> f = Params::viaLambda;
        f.accept(Codec.class);
    }

    static void viaLambda(Class<?> c) {
        ServiceLoader.load(c).iterator();
    }

    // No caller in the analyzed input.
    public static void orphan(Class<?> c) {
        ServiceLoader.load(c).iterator();
    }

    // A virtual method overridden in a subclass: no single exact target.
    public static class Base {
        public void go(Class<?> c) {
            ServiceLoader.load(c).iterator();
        }
    }

    public static class Sub extends Base {
        @Override
        public void go(Class<?> c) {
        }
    }

    public static void virtualCaller() {
        new Base().go(Codec.class);
    }
}
