package ar;

/** Like Jetty's EatWhatYouKill: runs whatever the producer returns. */
public class Strategy implements ExecStrategy {
    private final Producer producer;

    public Strategy(Producer producer) {
        this.producer = producer;
    }

    @Override
    public void loop() {
        Runnable task = produceTask();
        while (task != null) {
            runTask(task);
            task = produceTask();
        }
    }

    private Runnable produceTask() {
        return producer.produce();
    }

    private void runTask(Runnable task) {
        task.run();
    }
}
