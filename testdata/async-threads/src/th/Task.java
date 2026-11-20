package th;

public class Task implements Runnable {
    @Override
    public void run() {
        Work.hit("Task");
    }
}
