package fx.research;

import fx.research.internal.Cleaner;
import java.nio.file.Path;

/** Case (j): the library's only Sink; package-private and never created, so only a client's sink can run at the call sites. */
final class PurgingSink implements Sink {
    private PurgingSink() {}

    @Override
    public void accept(Path p) throws java.io.IOException {
        Cleaner.purge(p);
    }
}
