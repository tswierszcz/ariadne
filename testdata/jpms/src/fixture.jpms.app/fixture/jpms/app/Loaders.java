package fixture.jpms.app;

import fixture.jpms.api.Codec;
import java.util.ServiceLoader;

// ServiceLoader.load(Class, ClassLoader) argument shapes: four recognized
// from the method's own bytecode, one (a parameter) left as a gap. Each
// consumes the loader in the same method so the iterator links to it.
public class Loaders {
    public static void context() {
        for (Codec c : ServiceLoader.load(Codec.class, Thread.currentThread().getContextClassLoader())) {
            System.out.println(c.encode("context"));
        }
    }

    public static void system() {
        for (Codec c : ServiceLoader.load(Codec.class, ClassLoader.getSystemClassLoader())) {
            System.out.println(c.encode("system"));
        }
    }

    public static void nullLoader() {
        for (Codec c : ServiceLoader.load(Codec.class, null)) {
            System.out.println(c.encode("nullLoader"));
        }
    }

    public static void defining() {
        for (Codec c : ServiceLoader.load(Codec.class, Loaders.class.getClassLoader())) {
            System.out.println(c.encode("defining"));
        }
    }

    public static void given(ClassLoader loader) {
        for (Codec c : ServiceLoader.load(Codec.class, loader)) {
            System.out.println(c.encode("given"));
        }
    }
}
