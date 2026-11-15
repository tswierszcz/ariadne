package fx.app;

import fx.trace.Ran;
import java.lang.reflect.Method;

final class Minus {
    static final Method CHECK;
    static {
        try {
            CHECK = Minus.class.getDeclaredMethod("check", Object.class);
        } catch (NoSuchMethodException e) {
            throw new ExceptionInInitializerError(e);
        }
    }

    static boolean check(Object x) { Ran.stack(); return x == null; }
}
