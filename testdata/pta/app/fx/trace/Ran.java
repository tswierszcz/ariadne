package fx.trace;

import java.util.Collections;
import java.util.Set;
import java.util.TreeSet;

/**
 * Records, at a target body, the call that entered it: "caller -> target",
 * each as owner#name+desc — which body each call site ran.
 */
public final class Ran {
    public static final Set<String> calls = Collections.synchronizedSet(new TreeSet<>());
    private static final StackWalker W = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);

    private static String key(StackWalker.StackFrame f) {
        return f.getClassName() + "#" + f.getMethodName() + f.getDescriptor();
    }

    public static void call() {
        W.walk(s -> {
            var frames = s.skip(1).limit(2).toList();
            calls.add(key(frames.get(1)) + " -> " + key(frames.get(0)));
            return null;
        });
    }
}
