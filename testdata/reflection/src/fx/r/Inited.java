package fx.r;

import fx.trace.Ran;

/** REFLECTION R4 r4-1: its static initializer records that it ran. */
public class Inited {
    static { Ran.here(); }
}
