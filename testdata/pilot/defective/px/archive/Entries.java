package px.archive;

import java.nio.file.Path;

/** Defective variant: the entry name is resolved as given ("../x" leaves the directory). */
final class Entries {
    private Entries() {}

    static Path target(Path dest, String name) {
        return dest.resolve(name);
    }
}
