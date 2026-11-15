package fx.app;

import fx.trace.Ran;

/** Reached only through a name known at run time: the gap that stays. */
public final class Secret {
    public static boolean check(Object x) { Ran.stack(); return false; }
}
