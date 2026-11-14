package fx.trace;

import java.util.Set;
import java.util.TreeSet;

/** Records every method body that ran (owner#name+desc). */
public final class Ran {
    public static final Set<String> bodies = new TreeSet<>();
    private static final StackWalker W = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);

    public static void here() {
        StackWalker.StackFrame f = W.walk(s -> s.skip(1).findFirst().orElseThrow());
        bodies.add(f.getClassName() + "#" + f.getMethodName() + f.getDescriptor());
    }
}
