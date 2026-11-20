package af;

import java.util.concurrent.Executor;

/** Forwards every task it is given to an executor (the registration site). */
public class Pool {
    private final Executor executor;

    public Pool(Executor executor) {
        this.executor = executor;
    }

    public void submit(Runnable task) {
        executor.execute(task);
    }
}
