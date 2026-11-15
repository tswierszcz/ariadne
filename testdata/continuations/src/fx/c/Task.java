package fx.c;

import fx.trace.Ran;

/** A Runnable class handed to runAsync. */
public class Task implements Runnable {
    public void run() { Ran.here(); }
}
