package fx.app;

import fx.trace.Ran;

/** A bean with parent="base": it inherits the parent's tag property. */
public class Child extends Base {
    public Child() { Ran.stack(); }
}
