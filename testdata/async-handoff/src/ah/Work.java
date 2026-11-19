package ah;

/** What the queues hold; the drains call work() on each element. */
public interface Work {
    void work();
}
