package fx.ev;

import fx.trace.Ran;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

/** @Scheduled once, and twice (the class file then holds @Schedules). */
@Component
public class Ticker {
    private final CountDownLatch ran = new CountDownLatch(2);

    @Scheduled(fixedDelay = 10)
    public void tick() { Ran.stack(); ran.countDown(); }

    @Scheduled(fixedDelay = 10)
    @Scheduled(initialDelay = 5, fixedDelay = 1000)
    public void twice() { Ran.stack(); ran.countDown(); }

    public void await() throws InterruptedException {
        if (!ran.await(10, TimeUnit.SECONDS)) {
            throw new IllegalStateException("scheduled methods did not run");
        }
    }
}
