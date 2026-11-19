package ah;

import java.util.Queue;
import java.util.concurrent.ConcurrentLinkedQueue;

/** A work queue drained by removal (poll). */
public class Loop {
    private final Queue<Work> queue = new ConcurrentLinkedQueue<>();

    public void submit(Work w) {
        queue.offer(w);
    }

    public void drain() {
        Work w;
        while ((w = queue.poll()) != null) {
            w.work();
        }
    }
}
