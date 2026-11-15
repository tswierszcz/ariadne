package fx.trace;

import java.util.Collections;
import java.util.Set;
import java.util.TreeSet;

/** Records, at a body, the frame that called stack() (owner#name+desc) — from any thread. */
public final class Ran {
    public static final Set<String> bodies = Collections.synchronizedSet(new TreeSet<>());
    private static final StackWalker W = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);

    public static void here() {
        W.walk(s -> {
            s.skip(1).findFirst().ifPresent(f -> bodies.add(f.getClassName() + "#" + f.getMethodName() + f.getDescriptor()));
            return null;
        });
    }
}
