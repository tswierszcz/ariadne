package ar;

/**
 * Like Jetty's ProduceConsume: a strategy nothing constructs. CHA reaches
 * its loop through ExecStrategy, but its producer comes only from its
 * constructor's callers — none — so it runs no producer's task.
 */
public class OtherStrategy implements ExecStrategy {
    private final Producer producer;

    public OtherStrategy(Producer producer) {
        this.producer = producer;
    }

    @Override
    public void loop() {
        Runnable task = producer.produce();
        if (task != null) {
            task.run();
        }
    }
}
