package fx.app;

import fx.trace.Ran;

/** A lazy-init bean nothing asks for: not created on this run. */
public class Sleepy {
    public Sleepy() { Ran.stack(); }
}
