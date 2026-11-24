package fx.client;

import fx.research.DeletingHandler;
import fx.research.Extractor;
import java.io.ByteArrayInputStream;
import java.nio.file.Files;
import java.nio.file.Path;

/** Oracle client: public API calls only, in a temporary directory. */
public class Main {
    public static void main(String[] a) throws Exception {
        Path dir = Files.createTempDirectory("research-fx");
        byte[] name = "out.txt".getBytes("UTF-8");
        byte[] msg = new byte[1 + name.length + 2];
        msg[0] = (byte) name.length;
        System.arraycopy(name, 0, msg, 1, name.length);
        msg[msg.length - 2] = 'h';
        msg[msg.length - 1] = 'i';
        new Extractor().extract(new ByteArrayInputStream(msg), dir);
        new Extractor().save(new byte[0]);
        new DeletingHandler().process(dir.resolve("out.txt"));
        Files.delete(dir);
    }
}
