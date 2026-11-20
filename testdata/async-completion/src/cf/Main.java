package cf;

import java.util.concurrent.CompletableFuture;

public class Main {
    public static void main(String[] args) throws Exception {
        // Completed normally: then* runs, exceptionally does not.
        CompletableFuture<Object> ok = new CompletableFuture<>();
        ok.thenApply(new Upper());
        ok.exceptionally(new Recover());
        ok.complete("x");

        // Never completed: nothing runs.
        CompletableFuture<Object> never = new CompletableFuture<>();
        never.thenAccept(new Never());

        // Failed: exceptionally and whenComplete run, thenApply does not.
        CompletableFuture<Object> failing = new CompletableFuture<>();
        failing.thenApply(new Skipped());
        failing.exceptionally(new Handled());
        failing.whenComplete(new Both());
        failing.completeExceptionally(new RuntimeException("no"));

        // Completed by a method it is handed to.
        CompletableFuture<Void> later = new CompletableFuture<>();
        later.thenRun(new Later());
        Api.finish(later);

        // A field future whose completing method never runs.
        new Holder().watch();

        // A source stage: unchanged (completes when its task does).
        CompletableFuture.runAsync(new Source()).get();
    }
}
