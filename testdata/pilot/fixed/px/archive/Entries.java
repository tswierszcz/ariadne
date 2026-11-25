package px.archive;

import java.io.IOException;
import java.nio.file.Path;

/** Corrected variant: a name that resolves outside the directory is refused. */
final class Entries {
    private Entries() {}

    static Path target(Path dest, String name) throws IOException {
        Path root = dest.toAbsolutePath().normalize();
        Path t = root.resolve(name).normalize();
        if (!t.startsWith(root)) {
            throw new IOException("zip entry outside the destination: " + name);
        }
        return t;
    }
}
