package rt;

import java.util.ServiceLoader;

/** Drives every scenario on the real JVM and prints the method bodies that ran. */
public final class Oracle {
    static double total(Shape s) {
        return s.area();
    }

    public static void main(String[] args) throws Exception {
        total(new Square());
        total(new Circle());
        total(Factory.make());
        new StepA().go(2);
        for (Plugin p : ServiceLoader.load(Plugin.class)) {
            p.name();
        }
        // simple names only: "rt." is prepended at run time, so no constant spells a class name
        total(UnknownFactory.create("Dynamic"));
        total(UnknownFactory.create("Unnamed"));
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
