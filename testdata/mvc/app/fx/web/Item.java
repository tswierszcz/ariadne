package fx.web;

import fx.trace.Ran;

/** Written by Jackson through its getters. */
public class Item {
    private final String id;
    private final String name;

    public Item(String id, String name) { Ran.stack(); this.id = id; this.name = name; }

    public String getId() { Ran.stack(); return id; }

    public String getName() { Ran.stack(); return name; }
}
