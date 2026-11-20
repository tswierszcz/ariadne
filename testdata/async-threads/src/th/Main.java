package th;

import java.util.concurrent.SynchronousQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

public class Main {
    public static void main(String[] args) throws Exception {
        Worker w = new Worker();
        w.start();
        w.join();
        Worker s = new Special();
        s.start();
        s.join();
        new Idle();

        // One thread, no queue: a second task while the first runs is rejected.
        ThreadPoolExecutor pool = new ThreadPoolExecutor(1, 1, 0, TimeUnit.SECONDS, new SynchronousQueue<>(), new Reject());
        pool.execute(() -> sleep());
        pool.execute(new Task());
        pool.shutdown();
        pool.awaitTermination(5, TimeUnit.SECONDS);
    }

    static void sleep() {
        try {
            Thread.sleep(300);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
