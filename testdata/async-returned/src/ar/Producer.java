package ar;

/** Like Jetty's ExecutionStrategy.Producer: hands out the next task. */
public interface Producer {
    Runnable produce();
}
