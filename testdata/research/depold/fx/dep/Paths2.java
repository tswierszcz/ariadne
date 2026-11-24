package fx.dep;

import java.nio.file.Path;

/** Case (g): version 1 of the same class (fx-dep-1.0.jar): the selected copy, since it sorts first. */
public final class Paths2 {
    private Paths2() {}

    public static Path resolve(Path dir, String name) {
        return dir.resolve(name);
    }

    public static Path legacyResolve(Path dir, String name) {
        return dir.resolve(name.replace('\\', '/'));
    }
}
