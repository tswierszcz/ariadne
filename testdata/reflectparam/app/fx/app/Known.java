package fx.app;

import fx.trace.Ran;

/** Passed to Ops.mixed by a literal lookup: resolved although the site stays partial. */
public final class Known {
    public static boolean check(Object x) { Ran.stack(); return true; }
}
