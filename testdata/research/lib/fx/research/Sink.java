package fx.research;

import java.nio.file.Path;

/** Case (j): a callback type a client implements. */
public interface Sink {
    void accept(Path p) throws java.io.IOException;
}
