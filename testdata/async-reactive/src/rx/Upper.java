package rx;

import java.util.concurrent.Flow;
import java.util.concurrent.SubmissionPublisher;

/** A Flow.Processor: republishes each item upper-cased (submit on this). */
public class Upper extends SubmissionPublisher<String> implements Flow.Processor<String, String> {
    @Override
    public void onSubscribe(Flow.Subscription s) {
        s.request(Long.MAX_VALUE);
    }

    @Override
    public void onNext(String item) {
        submit(item.toUpperCase());
    }

    @Override
    public void onError(Throwable t) {
        closeExceptionally(t);
    }

    @Override
    public void onComplete() {
        close();
    }
}
