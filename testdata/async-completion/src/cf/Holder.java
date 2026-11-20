package cf;

import java.util.concurrent.CompletableFuture;

/** A future in a final field; done() would complete it but never runs. */
public class Holder {
    private final CompletableFuture<Void> f = new CompletableFuture<>();

    void watch() {
        f.thenRun(new Pending());
    }

    void done() {
        f.complete(null);
    }
}
