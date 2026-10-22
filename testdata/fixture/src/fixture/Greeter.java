package fixture;

/** An interface so CHA expansion has something to fan out to. */
public interface Greeter {
    String greet(String name);

    /**
     * A default method neither FormalGreeter nor CasualGreeter overrides
     * — JVM_GRAPH G2's default-method resolution path (see BridgeDemo):
     * a call through either implementation must resolve to this
     * declaration, the same as any other inherited, non-overridden
     * method.
     */
    default String defaultGreeting() {
        return "Hello, stranger.";
    }
}
