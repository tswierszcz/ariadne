package cg;

import java.util.Set;
import java.util.TreeSet;

/** Records which method body ran, for the call site the driver set. */
public final class Trace {
    static String site = "";
    static final Set<String> hits = new TreeSet<>();
    private static final StackWalker W = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);

    static void hit() {
        StackWalker.StackFrame f = W.walk(s -> s.skip(1).findFirst().orElseThrow());
        hits.add(site + "\t" + f.getClassName() + "#" + f.getMethodName() + f.getDescriptor());
    }
}
