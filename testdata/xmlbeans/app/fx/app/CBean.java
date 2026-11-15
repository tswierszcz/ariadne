package fx.app;

import fx.trace.Ran;

/** A constructor argument through the c: namespace. */
public class CBean {
    public CBean(String value) { Ran.stack(); }
}
