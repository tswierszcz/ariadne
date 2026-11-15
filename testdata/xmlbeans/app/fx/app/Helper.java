package fx.app;

import fx.trace.Ran;

/** An inner bean, declared inside Service's <property>. */
public class Helper {
    public Helper() { Ran.stack(); }
}
