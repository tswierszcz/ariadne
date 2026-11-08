package rtm;

/** Loaded without being initialized: no method of it runs. */
public final class LoadedOnly {
    static {
        System.out.println("never printed");
    }

    public static void run() {
        System.out.println("never run");
    }
}
