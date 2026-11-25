package fx.research;

import java.nio.file.Path;
import java.util.concurrent.CompletableFuture;

/** Case (k): two callbacks registered on one future; recover runs only if it fails (a site condition). */
public final class Workers {
    private Workers() {}

    public static CompletableFuture<Path> later(CompletableFuture<Path> f) {
        f.thenAccept(Workers::archive);
        return f.exceptionally(Workers::recover);
    }

    static void archive(Path p) {
        // runs on every successful completion
    }

    static Path recover(Throwable t) {
        return null; // runs only when the future completes exceptionally
    }
}
