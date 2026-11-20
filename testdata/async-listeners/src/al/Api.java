package al;

public class Api {
    /** Never called: D is registered nowhere that runs. */
    static void later(Bus bus) {
        bus.register(new D());
    }
}
