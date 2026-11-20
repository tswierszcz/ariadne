package cf;

public class Pending implements Runnable {
    @Override
    public void run() {
        Work.hit("Pending");
    }
}
