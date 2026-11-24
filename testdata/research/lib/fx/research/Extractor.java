package fx.research;

import fx.dep.Paths2;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;

/** Public API: unpacks one entry of a stream into a directory. */
public class Extractor {
    /** Case (a): public entry, reaches the private writeEntry and Files.write. */
    public void extract(InputStream in, Path dir) throws IOException {
        String name = new String(in.readNBytes(in.read()), "UTF-8");
        byte[] body = in.readAllBytes();
        writeEntry(Paths2.resolve(dir, name), body);
    }

    /** Case (b) target: private, called only from extract. */
    private void writeEntry(Path target, byte[] body) throws IOException {
        Files.write(target, body);
    }

    /** Case (c): private, nothing calls it. */
    private void unused(Path target) throws IOException {
        Files.delete(target);
    }

    /** Case (d): reflective plug-in lookup with a name from the caller (unresolved). */
    public Object plugin(String className) throws ReflectiveOperationException {
        return Class.forName(className).getDeclaredConstructor().newInstance();
    }

    /** Overloads: only the (String) one reaches writeEntry. */
    public void save(String name, byte[] body, Path dir) throws IOException {
        writeEntry(dir.resolve(name), body);
    }

    public void save(byte[] body) {
        // does nothing: an overload that reaches no target
    }
}
