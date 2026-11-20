package th;

/** Created, never started. */
public class Idle extends Thread {
    @Override
    public void run() {
        Work.hit("Idle");
    }
}
