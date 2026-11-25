package fx.research.internal;

import java.nio.file.Files;
import java.nio.file.Path;

/** Case (e) target. */
public final class Cleaner {
    private Cleaner() {}

    public static void wipe(Path p) throws java.io.IOException {
        Files.deleteIfExists(p);
    }

    /** Case (j) target: only PurgingSink calls it. */
    public static void purge(Path p) throws java.io.IOException {
        Files.deleteIfExists(p);
    }
}
