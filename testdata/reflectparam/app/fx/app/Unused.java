package fx.app;

import fx.trace.Ran;
import java.lang.reflect.Method;

/** Negative: its Method is looked up but never passed to a helper. */
final class Unused {
    static final Method CHECK;
    static {
        try {
            CHECK = Unused.class.getDeclaredMethod("check", Object.class);
        } catch (NoSuchMethodException e) {
            throw new ExceptionInInitializerError(e);
        }
    }

    static boolean check(Object x) { Ran.stack(); return false; }
}
