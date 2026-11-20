package th;

/** A subclass of Worker started through a Worker-typed variable. */
public class Special extends Worker {
    @Override
    public void run() {
        Work.hit("Special");
    }
}
