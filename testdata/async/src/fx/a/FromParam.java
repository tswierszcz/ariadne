package fx.a;

import fx.trace.Ran;

/** Reaches the registration only through a parameter. */
public class FromParam implements Runnable {
    public void run() { Ran.here(); }
}
