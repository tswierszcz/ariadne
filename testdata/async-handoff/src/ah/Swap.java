package ah;

import java.util.ArrayDeque;
import java.util.Deque;

/**
 * Double buffering: submit adds to one deque, process swaps them, iterates
 * the other and clears it (Jetty's ManagedSelector); the catch inside the
 * loop jumps back to its head.
 */
public class Swap {
    private Deque<Work> pending = new ArrayDeque<>();
    private Deque<Work> active = new ArrayDeque<>();

    public synchronized void submit(Work w) {
        pending.add(w);
    }

    public void process() {
        synchronized (this) {
            Deque<Work> t = pending;
            pending = active;
            active = t;
        }
        for (Work w : active) {
            try {
                w.work();
            } catch (RuntimeException e) {
                Main.ran("failed " + e);
            }
        }
        active.clear();
    }
}
