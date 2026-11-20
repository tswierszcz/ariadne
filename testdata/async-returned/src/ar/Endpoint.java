package ar;

/** Like Jetty's ManagedSelector.Selectable: returns the task to run when selected. */
public interface Endpoint {
    Runnable onSelected();
}
