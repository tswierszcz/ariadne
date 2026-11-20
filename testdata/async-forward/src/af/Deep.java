package af;

/** Submitted through Relay#relay and Relay#pass (three caller levels). */
public class Deep implements Runnable {
    public void run() {
        Main.ran("Deep");
    }
}
