package th;

/** A Thread subclass overriding run(): no Runnable anywhere. */
public class Worker extends Thread {
    @Override
    public void run() {
        Work.hit("Worker");
    }
}
