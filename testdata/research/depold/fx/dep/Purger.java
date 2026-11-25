package fx.dep;

import java.nio.file.Files;
import java.nio.file.Path;

/** Case (j) target: the dependency's own Callback; the library never creates one. */
public class Purger implements Callback {
    @Override
    public void call(Path p) throws java.io.IOException {
        Files.deleteIfExists(p);
    }
}
