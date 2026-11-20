package af;

/** The callers: a direct `new`, a lambda, and a task relayed two levels. */
public class Api {
    public static void go(Pool pool) {
        pool.submit(new Task());
        pool.submit(() -> Main.ran("lambda"));
        new Relay(pool).relay(new Deep());
    }
}
