package lib.api;

/** (d) Calls the client's Listener: only client code runs there. */
public class Pipeline {
    public void run(Listener listener, String text) {
        listener.onEvent(text);
    }
}
