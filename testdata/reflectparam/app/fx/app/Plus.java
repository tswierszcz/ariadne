package fx.app;

import fx.trace.Ran;
import java.lang.reflect.Method;

/** Keeps its check Method in a static field, like Thymeleaf's AdditionSubtractionExpression. */
final class Plus {
    static final Method CHECK;
    static {
        try {
            CHECK = Plus.class.getDeclaredMethod("check", Object.class);
        } catch (NoSuchMethodException e) {
            throw new ExceptionInInitializerError(e);
        }
    }

    static boolean check(Object x) { Ran.stack(); return x != null; }
}
