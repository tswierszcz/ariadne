package fx.a;

import fx.trace.Ran;

/** Allocated but never registered anywhere: no callback edge may reach it. */
public class Unused implements Runnable {
    public void run() { Ran.here(); }
}
