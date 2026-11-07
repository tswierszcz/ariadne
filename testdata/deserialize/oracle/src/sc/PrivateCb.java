package sc;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.Serializable;

/** Private readObject: run by serialization, never called directly. */
public class PrivateCb implements Serializable {
    private static final long serialVersionUID = 1L;

    private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        in.defaultReadObject();
        Rec.ran("sc.PrivateCb#readObject");
    }

    private Object readResolve() {
        Rec.ran("sc.PrivateCb#readResolve");
        return this;
    }
}
