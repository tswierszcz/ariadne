package fx.app;

import fx.trace.Ran;

/** An instance method invoked through Ops.call. */
public class Greeter {
    public String hello() { Ran.stack(); return "hello"; }
}
