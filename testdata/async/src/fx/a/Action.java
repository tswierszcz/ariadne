package fx.a;

import fx.trace.Ran;
import java.security.PrivilegedAction;

public class Action implements PrivilegedAction<String> {
    public String run() { Ran.here(); return "action"; }
}
