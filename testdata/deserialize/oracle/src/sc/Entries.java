package sc;

import java.io.ByteArrayInputStream;
import java.io.InputStream;
import java.io.ObjectInputFilter;
import java.io.ObjectInputStream;

/** The deserialization entries of the oracle, one policy each. */
public class Entries {
    public static Object plain(InputStream in) throws Exception {
        return new ObjectInputStream(in).readObject();
    }

    public static Object filtered(InputStream in) throws Exception {
        ObjectInputStream ois = new ObjectInputStream(in);
        ois.setObjectInputFilter(ObjectInputFilter.Config.createFilter("sc.PrivateCb;!*"));
        return ois.readObject();
    }

    static InputStream of(byte[] b) {
        return new ByteArrayInputStream(b);
    }
}
