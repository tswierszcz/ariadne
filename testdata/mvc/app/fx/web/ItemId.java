package fx.web;

import fx.trace.Ran;

public final class ItemId {
    private final String value;

    public ItemId(String value) { Ran.stack(); this.value = value; }

    public String value() { Ran.stack(); return value; }
}
