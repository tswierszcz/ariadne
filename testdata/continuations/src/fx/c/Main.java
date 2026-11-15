package fx.c;

import fx.trace.Ran;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.CompletionStage;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.function.Consumer;

/**
 * CompletionStage continuations oracle (ASYNC_CALLBACKS A3): one case per
 * method; every callback body records itself (fx.trace.Ran). Written to
 * testdata/continuations/observed.txt by scripts/continuations-oracle.sh;
 * cmd/ariadne TestContinuationsOracle compares the graph with it.
 */
public final class Main {
    static final ExecutorService POOL = Executors.newFixedThreadPool(2);

    /** Two continuations after an async source. */
    static void chain() {
        CompletableFuture.supplyAsync(() -> { Ran.here(); return "a"; })
                .thenApply(Main::upper)
                .thenAccept(s -> Ran.here())
                .join();
    }

    static String upper(String s) { Ran.here(); return s.toUpperCase(); }

    /** thenCompose: the function returns the next stage. */
    static void compose() {
        CompletableFuture.completedFuture(1).thenCompose(Main::next).join();
    }

    static CompletableFuture<Integer> next(Integer i) { Ran.here(); return CompletableFuture.completedFuture(i + 1); }

    /** thenCombine: the callback is the second argument. */
    static void combine() {
        CompletableFuture<Integer> a = CompletableFuture.completedFuture(1);
        a.thenCombine(CompletableFuture.completedFuture(2), (x, y) -> { Ran.here(); return x + y; }).join();
    }

    /** A failing stage: the exceptional branch runs, the normal one does not. */
    static void failure() {
        CompletableFuture<String> f = CompletableFuture.supplyAsync(() -> { throw new IllegalStateException("x"); });
        f.thenApply(s -> { Ran.here(); return s; }).exceptionally(t -> { Ran.here(); return "recovered"; }).join();
        f.handle((v, t) -> { Ran.here(); return v; }).join();
        f.whenComplete((v, t) -> Ran.here()).exceptionally(t -> null).join();
    }

    /** A succeeding stage: its exceptionally callback never runs. */
    static void success() {
        CompletableFuture.completedFuture("ok").exceptionally(t -> { Ran.here(); return "never"; }).join();
    }

    /** *Async variants on an explicit executor; a Runnable class. */
    static void async() {
        CompletableFuture.runAsync(new Task(), POOL).thenRunAsync(Main::after, POOL).join();
    }

    static void after() { Ran.here(); }

    /** Called through the CompletionStage interface. */
    static void throughInterface() {
        CompletionStage<String> st = CompletableFuture.completedFuture("i");
        st.thenAccept(s -> Ran.here()).toCompletableFuture().join();
    }

    /** A future nothing completes: its continuation never runs. */
    static void never() {
        new CompletableFuture<String>().thenAccept(s -> Ran.here());
    }

    /** The callback reaches the registration through a parameter. */
    static void fromParam() {
        attach(CompletableFuture.completedFuture("p"), s -> Ran.here());
    }

    static void attach(CompletableFuture<String> f, Consumer<String> c) { f.thenAccept(c).join(); }

    public static void main(String[] args) {
        chain();
        compose();
        combine();
        failure();
        success();
        async();
        throughInterface();
        never();
        fromParam();
        POOL.shutdown();
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
