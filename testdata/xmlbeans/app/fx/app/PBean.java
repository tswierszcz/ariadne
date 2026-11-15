package fx.app;

import fx.trace.Ran;

/** Properties through the p: namespace (a value and a -ref). */
public class PBean {
    public void setLabel(String label) { Ran.stack(); }
    public void setRepo(Repo repo) { Ran.stack(); }
}
