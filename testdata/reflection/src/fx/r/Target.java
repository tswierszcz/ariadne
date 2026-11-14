package fx.r;

import fx.trace.Ran;

public class Target {
    public void hello() { Ran.here(); }
    public static String stat(String s) { Ran.here(); return s; }
    private void secret() { Ran.here(); }
    public void viaField() { Ran.here(); }
    public void twoArgs(String a, String b) { Ran.here(); }
    public void chosenAtRunTime() { Ran.here(); }
}
