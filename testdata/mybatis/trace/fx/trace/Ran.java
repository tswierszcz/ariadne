package fx.trace;

import java.util.Collections;
import java.util.Set;
import java.util.TreeSet;

/**
 * Records, at a driver body, every frame on the stack (owner#name+desc):
 * each is a body the JVM ran to get there — the application's, MyBatis's,
 * the driver's.
 */
public final class Ran {
    public static final Set<String> bodies = Collections.synchronizedSet(new TreeSet<>());
    private static final StackWalker W = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);

    public static void stack() {
        W.walk(s -> {
            s.skip(1).forEach(f -> bodies.add(f.getClassName() + "#" + f.getMethodName() + f.getDescriptor()));
            return null;
        });
    }
}
