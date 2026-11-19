package af;

/** Passes a task on through two methods before it reaches Pool#submit. */
public class Relay {
    private final Pool pool;

    public Relay(Pool pool) {
        this.pool = pool;
    }

    public void relay(Runnable r) {
        pass(r);
    }

    void pass(Runnable r) {
        pool.submit(r);
    }
}
