package sc;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.util.concurrent.LinkedBlockingQueue;

/**
 * Serializable only through a JDK superclass outside ariadne's built-in
 * table: its serializability is "unknown", so it is not a candidate.
 */
public class Queue extends LinkedBlockingQueue<Object> {
    private static final long serialVersionUID = 1L;

    private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        in.defaultReadObject();
        Rec.ran("sc.Queue#readObject");
    }
}
