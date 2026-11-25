package px.archive;

import px.archive.internal.Store;
import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Path;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

/**
 * Public API of the RH-A3 pilot target. Property under test: unpack writes
 * only inside the destination directory it was constructed with.
 */
public final class Unpacker {
    private final Path dest;

    public Unpacker(Path dest) {
        this.dest = dest;
    }

    /** Unpacks every file entry of a zip stream into the destination directory. */
    public int unpack(InputStream zip) throws IOException {
        int n = 0;
        try (ZipInputStream in = new ZipInputStream(zip)) {
            for (ZipEntry e; (e = in.getNextEntry()) != null; ) {
                if (e.isDirectory()) {
                    continue;
                }
                Store.write(Entries.target(dest, e.getName()), in.readAllBytes());
                n++;
            }
        }
        return n;
    }

    /**
     * Control (allowed by contract): writes data to the path the caller
     * names, anywhere. Reaching Store#write from here is not a defect.
     */
    public static void writeFile(Path path, byte[] data) throws IOException {
        Store.write(path, data);
    }

    /** Control (unresolved reflection): a filter class named by the caller. */
    public Object filter(String className) throws ReflectiveOperationException {
        return Class.forName(className).getDeclaredConstructor().newInstance();
    }
}
