package fx.research;

import fx.research.internal.Cleaner;
import java.nio.file.Path;

/** Case (e): inherits the public process, which reaches Cleaner#wipe through handle. */
public class DeletingHandler extends Handler {
    @Override
    protected void handle(Path p) throws java.io.IOException {
        Cleaner.wipe(p);
    }
}
