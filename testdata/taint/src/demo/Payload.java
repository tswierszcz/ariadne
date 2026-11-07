package demo;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.Serializable;

public class Payload implements Serializable {
    private static final long serialVersionUID = 1L;
    String note;
    private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        in.defaultReadObject();
        System.out.println("  Payload.readObject ran");
    }
}
