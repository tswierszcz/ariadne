package fx.app;

import java.lang.reflect.Method;

/**
 * The helpers that receive a Method as a parameter and invoke it — the
 * shape of Thymeleaf's BinaryOperationExpression#
 * doComposeBinaryOperationExpression.
 */
final class Ops {
    /** A static target: invoke(null, x). */
    static Object apply(Method check, Object x) throws Exception {
        return check.invoke(null, x);
    }

    /** An instance target passed alongside. */
    static Object call(Method m, Object target) throws Exception {
        return m.invoke(target);
    }

    /** One caller passes a Method found by a name known only at run time. */
    static Object mixed(Method m) throws Exception {
        return m.invoke(null, "z");
    }
}
