package lib;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.Serializable;

/** A library class; the second jar carries a relocated copy of it. */
public class Cb implements Serializable {
    private static final long serialVersionUID = 1L;

    private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        in.defaultReadObject();
        sc.Rec.ran("lib.Cb#readObject");
    }
}
