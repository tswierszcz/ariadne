package af;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

/** JVM oracle driver (scripts/async-forward-oracle.sh). */
public class Main {
    static synchronized void ran(String what) {
        System.out.println("ran " + what);
    }

    public static void main(String[] args) throws Exception {
        ExecutorService ex = Executors.newSingleThreadExecutor();
        Api.go(new Pool(ex));
        ex.shutdown();
        ex.awaitTermination(10, TimeUnit.SECONDS);
    }
}
