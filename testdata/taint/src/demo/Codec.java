package demo;

import java.io.ByteArrayInputStream;
import java.io.ObjectInputStream;
import java.util.Base64;

final class Codec {
    static byte[] decode(String s) {
        return Base64.getDecoder().decode(s);
    }

    static Object read(byte[] b) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(b))) {
            return in.readObject();
        }
    }
    static Object readOther(byte[] b) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(b))) {
            return in.readObject();
        }
    }

    static Object readDeep(byte[] b) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(b))) {
            return in.readObject();
        }
    }

    static Object readShared(byte[] b) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(b))) {
            return in.readObject();
        }
    }
}
