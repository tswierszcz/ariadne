package af;

/** A Runnable nobody submits: it must get no edge. */
public class Unused implements Runnable {
    public void run() {
        Main.ran("Unused");
    }
}
