package al;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.ObjectInputStream;
import java.io.ObjectOutputStream;
import java.lang.reflect.Proxy;
import java.util.ServiceLoader;

/** Creates objects in each way the catalog models (and a dynamic proxy, a gap); prints the runtime classes. */
public final class Oracle {
    static Shape factory() {
        return new Circle();
    }

    static Shape unused() {
        return new Never();
    }

    public static void main(String[] args) throws Exception {
        Shape a = new Square();
        Shape b = new BigSquare();
        Shape c = factory();
        Shape d = (Shape) Class.forName("al.Hexagon").getConstructor().newInstance();
        for (Plugin p : ServiceLoader.load(Plugin.class)) {
            p.name();
        }
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        try (ObjectOutputStream out = new ObjectOutputStream(bos)) {
            out.writeObject(new Saved());
        }
        // the write side's `new Saved()` records nothing: Saved has no constructor recording
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(bos.toByteArray()))) {
            in.readObject();
        }
        Shape proxy = (Shape) Proxy.newProxyInstance(Oracle.class.getClassLoader(), new Class<?>[] {Shape.class}, (p, m, x) -> 0.0);
        Made.it(proxy);
        Runnable r = () -> { };
        r.run();
        System.out.println("# created on the real JVM (runtime class names; proxy classes are generated)");
        for (String cl : Made.classes) {
            System.out.println(cl.startsWith("jdk.proxy") || cl.contains("$Proxy") ? "<proxy>" : cl);
        }
    }
}
