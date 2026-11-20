package cf;

import java.util.concurrent.CompletableFuture;

public class Api {
    /** Completes the future it is given (a parameter: resolved at the callers). */
    static void finish(CompletableFuture<Void> f) {
        f.complete(null);
    }
}
