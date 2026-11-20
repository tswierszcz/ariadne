package ar;

/** Like Jetty's SelectorProducer: the task is what the selected endpoint returns. */
public class SelProducer implements Producer {
    private final Endpoint[] ready;
    private int next;

    public SelProducer(Endpoint[] ready) {
        this.ready = ready;
    }

    @Override
    public Runnable produce() {
        if (next >= ready.length) {
            return null;
        }
        Runnable task = ready[next++].onSelected();
        return task;
    }
}
