package th;

import java.util.concurrent.ThreadPoolExecutor;

public class Api {
    /** Never called. */
    static void configure(ThreadPoolExecutor pool) {
        pool.setRejectedExecutionHandler(new Unused());
    }
}
