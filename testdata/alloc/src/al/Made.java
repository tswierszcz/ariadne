package al;

import java.util.Set;
import java.util.TreeSet;

/** Records the runtime class of every object the fixture creates. */
public final class Made {
    static final Set<String> classes = new TreeSet<>();

    static void it(Object o) {
        classes.add(o.getClass().getName());
    }
}
