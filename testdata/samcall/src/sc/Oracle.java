package sc;

import java.util.Set;
import java.util.TreeSet;

/**
 * DYNAMIC_LINKAGE L1 oracle: every lambda body / method-reference target of
 * this fixture calls {@link #ran()}, which records the body and the frame
 * that ran it — the nearest frame of the fixture below the body, skipping
 * the JVM's hidden lambda classes and JDK frames (List#forEach, …): the
 * caller -> callee relation a runtime trace attributes the call to.
 * scripts/samcall-oracle.sh writes the result to observed.tsv.
 */
public final class Oracle {
    private static final StackWalker WALKER = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);
    private static final Set<String> SEEN = new TreeSet<>();

    static void ran() {
        WALKER.walk(frames -> {
            var it = frames.skip(1).iterator(); // skip ran() itself
            StackWalker.StackFrame body = it.next();
            while (it.hasNext()) {
                StackWalker.StackFrame f = it.next();
                if (f.getClassName().startsWith("sc.")) {
                    SEEN.add(key(f) + "\t" + f.getByteCodeIndex() + "\t" + key(body));
                    break;
                }
            }
            return null;
        });
    }

    private static String key(StackWalker.StackFrame f) {
        return f.getClassName() + "#" + f.getMethodName() + f.getDescriptor();
    }

    public static void main(String[] args) {
        Main.main(args);
        System.out.println("# caller\tbci\tcallee (the body that ran) — recorded on the real JVM by sc.Oracle");
        for (String s : SEEN) {
            System.out.println(s);
        }
    }
}
