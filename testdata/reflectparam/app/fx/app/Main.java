package fx.app;

import fx.trace.Ran;

/**
 * REFLECTION R3 oracle (r3-3): Method objects resolved by literal lookups
 * in one method and invoked in another, which receives them as a
 * parameter. scripts/reflectparam-oracle.sh writes the record, cmd/ariadne
 * TestReflectParamOracle compares the graph with it.
 */
public final class Main {
    public static void main(String[] args) throws Exception {
        Ops.apply(Plus.CHECK, "a");
        Ops.apply(Minus.CHECK, "b");
        Ops.apply(Times.class.getDeclaredMethod("check", Object.class), "c");
        Ops.call(Greeter.class.getMethod("hello"), new Greeter());
        Ops.mixed(Known.class.getMethod("check", Object.class));
        String name = System.getProperty("fx.target", "fx.app.Secret");
        Ops.mixed(Class.forName(name).getMethod("check", Object.class));
        System.out.println("# " + (Unused.CHECK != null));
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the targets' stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
