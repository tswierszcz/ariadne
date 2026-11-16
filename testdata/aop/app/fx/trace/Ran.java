package fx.trace;

import java.util.Collections;
import java.util.Set;
import java.util.TreeSet;

/**
 * Records, at a bean body, every frame on the stack (owner#name+desc):
 * each is a body the JVM ran to get there — the application's and
 * the container's.
 */
public final class Ran {
    public static final Set<String> bodies = Collections.synchronizedSet(new TreeSet<>());
    private static final StackWalker W = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);

    /** advice name + " <- " + the join point's declaring type#method, per advice run. */
    public static final Set<String> advised = Collections.synchronizedSet(new TreeSet<>());

    public static void advised(String advice, String type, String method) {
        advised.add(advice + " <- " + type + "#" + method);
    }

    public static void stack() {
        W.walk(s -> {
            s.skip(1).forEach(f -> bodies.add(f.getClassName() + "#" + f.getMethodName() + f.getDescriptor()));
            return null;
        });
    }
}
