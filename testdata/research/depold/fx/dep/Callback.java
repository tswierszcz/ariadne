package fx.dep;

import java.nio.file.Path;

/** Case (j): a dependency callback type the library's caller supplies. */
public interface Callback {
    void call(Path p) throws java.io.IOException;
}
