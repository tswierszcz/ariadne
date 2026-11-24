package fx.dep;

import java.nio.file.Path;

/** Case (f): dependency code. Version 2 (fx-dep-2.0.jar): shadowed by version 1. */
public final class Paths2 {
    private Paths2() {}

    public static Path resolve(Path dir, String name) {
        return dir.resolve(name);
    }

    /** Version 2 only: nothing in the library calls it. */
    public static Path safeResolve(Path dir, String name) {
        Path p = dir.resolve(name).normalize();
        if (!p.startsWith(dir)) throw new IllegalArgumentException(name);
        return p;
    }
}
