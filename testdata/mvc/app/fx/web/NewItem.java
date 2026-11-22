package fx.web;

import fx.trace.Ran;

/** Read by Jackson: the no-argument constructor, then the setter. */
public class NewItem {
    private String name;

    public NewItem() { Ran.stack(); }

    public String getName() { Ran.stack(); return name; }

    public void setName(String name) { Ran.stack(); this.name = name; }
}
