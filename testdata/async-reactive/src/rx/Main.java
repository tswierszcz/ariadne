package rx;

import java.util.concurrent.CompletableFuture;
import java.util.concurrent.SubmissionPublisher;
import java.util.concurrent.TimeUnit;

public class Main {
    public static void main(String[] args) throws Exception {
        // A pipeline that emits: subscriber, processor chain, consume.
        SubmissionPublisher<String> pub = new SubmissionPublisher<>();
        Printer printer = new Printer();
        pub.subscribe(printer);
        Upper upper = new Upper();
        pub.subscribe(upper);
        Sink sink = new Sink();
        upper.subscribe(sink);
        CompletableFuture<Void> consumed = pub.consume(s -> Work.consumed(s));
        Thrower thrower = new Thrower();
        pub.subscribe(thrower);
        new Idle(); // a subscriber nothing subscribes
        pub.submit("a");
        pub.close();
        printer.done.await(5, TimeUnit.SECONDS);
        sink.done.await(5, TimeUnit.SECONDS);
        thrower.done.await(5, TimeUnit.SECONDS);
        consumed.get(5, TimeUnit.SECONDS);

        // Subscribed, closed, never emits an item.
        SubmissionPublisher<String> quiet = new SubmissionPublisher<>();
        Quiet q = new Quiet();
        quiet.subscribe(q);
        quiet.close();
        q.done.await(5, TimeUnit.SECONDS);

        // Fails.
        SubmissionPublisher<String> bad = new SubmissionPublisher<>();
        Failing f = new Failing();
        bad.subscribe(f);
        bad.closeExceptionally(new RuntimeException("bad"));
        f.done.await(5, TimeUnit.SECONDS);
    }
}
