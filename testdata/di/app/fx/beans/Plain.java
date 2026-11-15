package fx.beans;

import fx.trace.Ran;

/** Negative: no stereotype, so not a bean — never created. */
public class Plain {
    public Plain() { Ran.stack(); }
}
