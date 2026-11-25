package px.archive.internal;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

/** The write method both variants reach. */
public final class Store {
    private Store() {}

    public static void write(Path target, byte[] data) throws IOException {
        Path parent = target.getParent();
        if (parent != null) {
            Files.createDirectories(parent);
        }
        Files.write(target, data);
    }

    /** Control (dead): nothing calls it. */
    static void overwrite(Path target, byte[] data) throws IOException {
        Files.write(target, data);
    }
}
