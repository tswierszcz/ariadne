package fx.a;

import fx.trace.Ran;
import java.lang.invoke.MethodHandle;
import java.lang.invoke.MethodHandles;
import java.lang.invoke.MethodType;
import java.security.AccessController;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

/**
 * ASYNC_CALLBACKS A0–A2 and DYNAMIC_LINKAGE L2 (MethodHandle) oracle
 * (ROADMAP B4): one case per method, run on the real JVM by main;
 * scripts/async-oracle.sh records which bodies ran, and cmd/ariadne
 * TestAsyncOracle compares the call graph with that record.
 */
public final class Main {
    static final ExecutorService POOL = Executors.newSingleThreadExecutor();

    // Positive: a Thread with a Runnable allocated in place.
    static void thread() throws Exception {
        Thread t = new Thread(new Task());
        t.start();
        t.join();
    }

    // Positive: a lambda handed to Executor#execute.
    static void lambdaToExecutor() throws Exception {
        POOL.execute(() -> Ran.here());
        POOL.submit(() -> Ran.here()).get(); // waits for the task above
    }

    // Positive: a Callable handed to ExecutorService#submit.
    static void callable() throws Exception {
        POOL.submit(new Job()).get();
    }

    // Positive: a synchronous callback (AccessController#doPrivileged).
    @SuppressWarnings("removal")
    static void privileged() {
        AccessController.doPrivileged(new Action());
    }

    // Negative: allocated, never registered — the JVM never runs it.
    static void unused() {
        new Unused();
    }

    // Unresolved: the callback reaches the registration through a parameter.
    static void viaParameter() throws Exception {
        runOn(new FromParam());
    }

    static void runOn(Runnable r) throws Exception {
        Thread t = new Thread(r);
        t.start();
        t.join();
    }

    // Positive: MethodHandles.Lookup#findVirtual with literals.
    static void handleVirtual() throws Throwable {
        MethodHandle h = MethodHandles.lookup().findVirtual(Handled.class, "virt", MethodType.methodType(void.class));
        h.invoke(new Handled());
    }

    // Positive: findStatic with literals.
    static void handleStatic() throws Throwable {
        MethodHandle h = MethodHandles.lookup().findStatic(Handled.class, "stat", MethodType.methodType(int.class, int.class));
        int r = (int) h.invokeExact(1);
    }

    // Negative: a method that does not exist — the JVM throws, nothing runs.
    static void handleMissing() throws Throwable {
        try {
            MethodHandles.lookup().findVirtual(Handled.class, "missing", MethodType.methodType(void.class)).invoke(new Handled());
        } catch (NoSuchMethodException expected) {
            // no body ran
        }
    }

    // Unresolved: the method name comes from the input.
    static void handleFromInput(String name) throws Throwable {
        MethodHandles.lookup().findVirtual(Handled.class, name, MethodType.methodType(void.class)).invoke(new Handled());
    }

    public static void main(String[] args) throws Throwable {
        thread();
        lambdaToExecutor();
        callable();
        privileged();
        unused();
        viaParameter();
        handleVirtual();
        handleStatic();
        handleMissing();
        handleFromInput(args.length > 0 ? args[0] : "chosenAtRunTime");
        POOL.shutdown();
        POOL.awaitTermination(10, TimeUnit.SECONDS);
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
