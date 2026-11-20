package ar;

public class IdleTask implements Runnable {
    @Override
    public void run() {
        Work.idle();
    }
}
