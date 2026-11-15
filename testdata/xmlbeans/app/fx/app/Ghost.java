package fx.app;

import fx.trace.Ran;

/** Negative: only an abstract definition names it — never instantiated. */
public class Ghost {
    public Ghost() { Ran.stack(); }
    public void setTag(String tag) { Ran.stack(); }
}
