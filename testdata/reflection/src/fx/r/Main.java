package fx.r;

import fx.trace.Ran;
import java.lang.reflect.Method;

/**
 * REFLECTION R0–R2 oracle (ROADMAP B4): each case is one method, run on
 * the real JVM by main; scripts/reflection-oracle.sh records which bodies
 * ran, and cmd/ariadne TestReflectionOracle compares the call graph with
 * that record.
 */
public final class Main {
    static final Method CACHED;
    static {
        try {
            CACHED = Target.class.getMethod("viaField");
        } catch (NoSuchMethodException e) {
            throw new IllegalStateException(e);
        }
    }

    // Positive: forName literal + getMethod literal + invoke on an allocation.
    static void literal() throws Exception {
        Class<?> c = Class.forName("fx.r.Target");
        Method m = c.getMethod("hello");
        m.invoke(new Target());
    }

    // Positive: the lookup is on Base, the target allocated as Sub — the JVM
    // dispatches to Sub#greet; Base#greet does not run.
    static void supertypeLookup() throws Exception {
        Method m = Base.class.getMethod("greet");
        m.invoke(new Sub());
    }

    // Positive: a static method, null target, literal Class[] parameter types.
    static void staticWithNull() throws Exception {
        Method m = Target.class.getMethod("stat", String.class);
        m.invoke(null, "x");
    }

    // Positive: a private method through getDeclaredMethod + setAccessible.
    static void declaredPrivate() throws Exception {
        Method m = Target.class.getDeclaredMethod("secret");
        m.setAccessible(true);
        m.invoke(new Target());
    }

    // Positive: the Method is cached in a static field and invoked elsewhere.
    static void cachedInField() throws Exception {
        CACHED.invoke(new Target());
    }

    // Positive: Constructor#newInstance of a forName literal.
    static void constructor() throws Exception {
        Object o = Class.forName("fx.r.Made").getConstructor().newInstance();
        ((Made) o).used();
    }

    // Negative: a proven argument-count mismatch — the JVM throws
    // IllegalArgumentException and twoArgs never runs.
    static void wrongArity() throws Exception {
        Method m = Target.class.getMethod("twoArgs", String.class, String.class);
        try {
            m.invoke(new Target(), new Object[] {"only one"});
        } catch (IllegalArgumentException expected) {
            // the call never reached the body
        }
    }

    // Unresolved: the method name comes from the input; the model cannot
    // know it and must report the site as a gap, not invent a target.
    static void nameFromInput(String name) throws Exception {
        Target.class.getMethod(name).invoke(new Target());
    }

    public static void main(String[] args) throws Exception {
        literal();
        supertypeLookup();
        staticWithNull();
        declaredPrivate();
        cachedInField();
        constructor();
        wrongArity();
        nameFromInput(args.length > 0 ? args[0] : "chosenAtRunTime");
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
