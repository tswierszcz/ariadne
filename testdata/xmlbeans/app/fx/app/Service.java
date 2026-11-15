package fx.app;

import fx.trace.Ran;

/**
 * A constructor argument, two properties (one an inner bean) and explicit
 * init-method/destroy-method. The other constructor is never used.
 */
public class Service {
    private final Repo repo;
    private String name;
    private Helper helper;

    public Service(Repo repo) { Ran.stack(); this.repo = repo; }
    public Service() { Ran.stack(); this.repo = null; }

    public void setName(String name) { Ran.stack(); this.name = name; }
    public void setHelper(Helper helper) { Ran.stack(); this.helper = helper; }
    public void start() { Ran.stack(); }
    public void stop() { Ran.stack(); }

    public String say() { return name + " " + repo.name() + (helper != null ? " with a helper" : ""); }
}
