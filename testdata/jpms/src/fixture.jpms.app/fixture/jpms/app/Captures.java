package fixture.jpms.app;

import fixture.jpms.api.Codec;
import java.security.PrivilegedAction;
import java.util.ServiceLoader;

// DATA_FLOW D3 slice "final fields set from constructor parameters".
public class Captures {
    // Jackson ObjectMapper#secureGetServiceLoader shape: an anonymous
    // PrivilegedAction captures the Class and ClassLoader parameters in
    // synthetic final fields (val$k, val$cl).
    public static void jackson() {
        secure(Codec.class, null).run();
    }

    private static PrivilegedAction<ServiceLoader<Codec>> secure(final Class<Codec> k, final ClassLoader cl) {
        return new PrivilegedAction<ServiceLoader<Codec>>() {
            @Override
            public ServiceLoader<Codec> run() {
                return ServiceLoader.load(k, cl);
            }
        };
    }

    // Two `new` sites with different values: a candidate set.
    static final class Box {
        private final Class<?> c;

        Box(Class<?> c) {
            this.c = c;
        }

        void go() {
            ServiceLoader.load(c).iterator();
        }
    }

    public static void boxes() {
        new Box(Codec.class).go();
        new Box(Runnable.class).go();
    }

    // Not final: any method could change it, so it stays unresolved.
    static final class Mutable {
        Class<?> c;

        Mutable(Class<?> c) {
            this.c = c;
        }

        void go() {
            ServiceLoader.load(c).iterator();
        }
    }

    public static void mutable() {
        new Mutable(Codec.class).go();
    }
}
