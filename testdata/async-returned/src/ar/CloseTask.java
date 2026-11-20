package ar;

public class CloseTask implements Runnable {
    @Override
    public void run() {
        Work.close();
    }
}
