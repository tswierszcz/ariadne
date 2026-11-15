package fx.app;

import fx.trace.Ran;

/** A plain bean; setup() is the file's default-init-method. */
public class Repo {
    public Repo() { Ran.stack(); }
    public void setup() { Ran.stack(); }
    String name() { return "Leo"; }
}
