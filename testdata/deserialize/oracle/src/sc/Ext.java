package sc;

import java.io.Externalizable;
import java.io.ObjectInput;
import java.io.ObjectOutput;

/** Externalizable: public no-arg constructor, then readExternal. */
public class Ext implements Externalizable {
    public Ext() {
        Rec.ran("sc.Ext#<init>");
    }

    @Override
    public void writeExternal(ObjectOutput out) {
    }

    @Override
    public void readExternal(ObjectInput in) {
        Rec.ran("sc.Ext#readExternal");
    }
}
