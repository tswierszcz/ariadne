package sd;

import java.util.Set;
import java.util.TreeSet;

/** Records every method body that ran (owner#name+desc), via the caller's stack frame. */
public final class Ran {
    static final Set<String> bodies = new TreeSet<>();
    private static final StackWalker W = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);

    static void here() {
        StackWalker.StackFrame f = W.walk(s -> s.skip(1).findFirst().orElseThrow());
        bodies.add(f.getClassName() + "#" + f.getMethodName() + f.getDescriptor());
    }
}
