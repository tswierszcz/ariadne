package fx.app;

import fx.trace.Ran;

/** Its check Method is looked up at the call site and passed directly. */
final class Times {
    static boolean check(Object x) { Ran.stack(); return true; }
}
