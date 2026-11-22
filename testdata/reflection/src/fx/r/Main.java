package fx.r;

import fx.trace.Ran;
import java.lang.reflect.Method;

/**
 * REFLECTION R0–R2 and R4 oracle (ROADMAP B4): each case is one method, run on
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

    // Positive (R4 r4-1): forName(String) initializes the class — its
    // static initializer runs, and nothing else in the program touches it.
    static void forNameInit() throws Exception {
        Class.forName("fx.r.Inited");
    }

    // Negative (R4 r4-1): initialize=false loads without initializing.
    static void forNameNoInit() throws Exception {
        Class.forName("fx.r.NotInited", false, Main.class.getClassLoader());
    }

    // Positive (R4 r4-1): a class with no initializer of its own — the JVM
    // initializes its superclass first, so InitedBase's runs.
    static void forNameSuper() throws Exception {
        Class.forName("fx.r.InitedSub");
    }

    // Positive (R4 r4-1): initialize is not a constant (true here at run
    // time) — the class may be initialized.
    static void forNameMaybe(boolean init) throws Exception {
        Class.forName("fx.r.MaybeInited", init, Main.class.getClassLoader());
    }

    // Negative (R4): the lookup alone — possession of a Method is not
    // invocation; lookedUp never runs.
    static void lookupOnly() throws Exception {
        Target.class.getMethod("lookedUp");
    }

    // Negative (R4): getMethod sees public methods only — the JVM throws
    // NoSuchMethodException and packagePrivate never runs.
    static void getMethodNotPublic() throws Exception {
        try {
            Target.class.getMethod("packagePrivate").invoke(new Target());
        } catch (NoSuchMethodException expected) {
            // the lookup failed
        }
    }

    // Negative (R4): getDeclaredMethod sees only what Sub declares itself;
    // baseOnly is inherited, so the JVM throws NoSuchMethodException.
    static void declaredInherited() throws Exception {
        try {
            Sub.class.getDeclaredMethod("baseOnly").invoke(new Sub());
        } catch (NoSuchMethodException expected) {
            // the lookup failed
        }
    }

    // Negative (R4): getConstructor sees public constructors only.
    static void constructorNotPublic() throws Exception {
        try {
            Hidden.class.getConstructor().newInstance();
        } catch (NoSuchMethodException expected) {
            // the lookup failed
        }
    }

    // Accepted over-approximation (R4): a private method found by
    // getDeclaredMethod but invoked without setAccessible(true) — the JVM
    // throws IllegalAccessException. The model does not track setAccessible
    // and keeps the edge: an extra edge, never a missed one.
    static void privateWithoutAccess() throws Exception {
        try {
            Target.class.getDeclaredMethod("locked").invoke(new Target());
        } catch (IllegalAccessException expected) {
            // the access check failed
        }
    }

    public static void main(String[] args) throws Exception {
        forNameInit();
        forNameNoInit();
        forNameSuper();
        forNameMaybe(args.length == 0);
        literal();
        supertypeLookup();
        staticWithNull();
        declaredPrivate();
        cachedInField();
        constructor();
        wrongArity();
        lookupOnly();
        getMethodNotPublic();
        declaredInherited();
        constructorNotPublic();
        privateWithoutAccess();
        nameFromInput(args.length > 0 ? args[0] : "chosenAtRunTime");
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
