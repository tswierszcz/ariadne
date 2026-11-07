package demo;

import java.io.ByteArrayInputStream;
import java.io.ObjectInputStream;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;

/**
 * DATA_FLOW D5's shapes: one handler (or a pair) per shape, each read by
 * its own entry so the model's result and the real JVM's observation
 * (scripts/taint-oracle.sh) can be compared per entry.
 */
public class Shapes {
    /** Aliasing: the token is appended through a second reference to the builder. */
    @PostMapping("/aliasing")
    public Object aliasing(@RequestParam("t") String token) throws Exception {
        StringBuilder sb = new StringBuilder();
        StringBuilder alias = sb;
        alias.append(token);
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(sb.toString())))) {
            return in.readObject();
        }
    }

    /** Arrays: the token travels in an array element. */
    @PostMapping("/arrays")
    public Object arrays(@RequestParam("t") String token) throws Exception {
        String[] parts = new String[] {"unused", token};
        byte[][] chunks = new byte[1][];
        chunks[0] = Codec.decode(parts[1]);
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(chunks[0]))) {
            return in.readObject();
        }
    }

    /** Exceptions (locals): the bytes decoded before a try are read in its handler. */
    @PostMapping("/exception-local")
    public Object exceptionLocal(@RequestParam("t") String token) throws Exception {
        byte[] b = Codec.decode(token);
        try {
            if (b.length > 0) {
                throw new IllegalStateException("take the handler");
            }
            return null;
        } catch (IllegalStateException e) {
            try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(b))) {
                return in.readObject();
            }
        }
    }

    /** Exceptions (the exception object): the token travels in the exception's message. */
    @PostMapping("/exception-message")
    public Object exceptionMessage(@RequestParam("t") String token) throws Exception {
        String carried;
        try {
            throw new IllegalArgumentException(token);
        } catch (IllegalArgumentException e) {
            carried = e.getMessage();
        }
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(carried)))) {
            return in.readObject();
        }
    }

    /** Two instances: two streams in one method, only the literal one is read. */
    @PostMapping("/two-instances")
    public Object twoInstances(@RequestParam("t") String token) throws Exception {
        ObjectInputStream fromRequest = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(token)));
        ObjectInputStream literal = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(Endpoints.HELLO)));
        fromRequest.close();
        return literal.readObject();
    }

    /** Overload: the String overload gets request data, the byte[] one a literal. */
    @PostMapping("/overload")
    public Object overload(@RequestParam("t") String token) throws Exception {
        Object a = Loader.load(token);
        Object b = Loader.load(Codec.decode(Endpoints.HELLO));
        return a + " " + b;
    }

    /** Callback: the token is read inside a Runnable's body (a lambda). */
    @PostMapping("/callback")
    public Object callback(@RequestParam("t") String token) throws Exception {
        Object[] out = new Object[1];
        Runnable r = () -> {
            try {
                out[0] = Loader.inCallback(token);
            } catch (Exception e) {
                throw new IllegalStateException(e);
            }
        };
        r.run();
        return out[0];
    }
}

final class Loader {
    static Object load(String s) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(s)))) {
            return in.readObject();
        }
    }

    static Object load(byte[] b) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(b))) {
            return in.readObject();
        }
    }

    static Object inCallback(String s) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(s)))) {
            return in.readObject();
        }
    }
}

/** No sources: a batch job reading a path given on its command line. */
final class Batch {
    static Object run(String encoded) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(encoded)))) {
            return in.readObject();
        }
    }

    public static void main(String[] args) throws Exception {
        System.out.println(run(args.length > 0 ? args[0] : Endpoints.HELLO));
    }
}
