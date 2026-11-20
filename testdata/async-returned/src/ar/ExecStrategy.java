package ar;

/** Like Jetty's ExecutionStrategy: callers hold the interface, CHA reaches every strategy. */
public interface ExecStrategy {
    void loop();
}
