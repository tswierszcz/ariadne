package lib.api;

/** A callback the client implements. */
public interface Listener {
    void onEvent(String text);
}
