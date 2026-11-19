package ah;

/**
 * The callers: go() hands objects to the queues (a direct new, a lambda)
 * and adds a listener that is never fired; other() — never called — submits
 * C, which only other() reaches.
 */
public class Api {
    public static void go(Loop loop, Swap swap, Listeners listeners) {
        loop.submit(new A());
        loop.submit(() -> Main.ran("lambda"));
        swap.submit(new B());
        listeners.add(new L());
    }

    public static void other(Loop loop) {
        loop.submit(new C());
    }
}
