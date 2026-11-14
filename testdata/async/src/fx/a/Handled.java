package fx.a;

import fx.trace.Ran;

public class Handled {
    public Handled() { Ran.here(); }
    public void virt() { Ran.here(); }
    public static int stat(int x) { Ran.here(); return x + 1; }
    public void chosenAtRunTime() { Ran.here(); }
}
