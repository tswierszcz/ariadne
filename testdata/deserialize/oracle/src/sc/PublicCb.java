package sc;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.Serializable;

/** A public readObject is not a serialization callback. */
public class PublicCb implements Serializable {
    private static final long serialVersionUID = 1L;

    public void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        Rec.ran("sc.PublicCb#readObject");
    }
}
