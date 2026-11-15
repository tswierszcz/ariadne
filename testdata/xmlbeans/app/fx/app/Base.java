package fx.app;

import fx.trace.Ran;

/** The class of an abstract parent definition, whose property a child inherits. */
public class Base {
    public void setTag(String tag) { Ran.stack(); }
}
