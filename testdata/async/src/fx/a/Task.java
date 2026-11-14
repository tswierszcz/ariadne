package fx.a;

import fx.trace.Ran;

public class Task implements Runnable {
    public void run() { Ran.here(); }
}
