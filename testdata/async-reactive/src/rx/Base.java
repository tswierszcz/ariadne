package rx;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.Flow;

/** A subscriber that requests everything and counts down when done. */
public abstract class Base implements Flow.Subscriber<String> {
    final CountDownLatch done = new CountDownLatch(1);

    @Override
    public void onSubscribe(Flow.Subscription s) {
        s.request(Long.MAX_VALUE);
    }

    @Override
    public void onError(Throwable t) {
        Work.hit(getClass().getSimpleName() + " error");
        done.countDown();
    }

    @Override
    public void onComplete() {
        Work.hit(getClass().getSimpleName() + " complete");
        done.countDown();
    }
}
