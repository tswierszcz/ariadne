package cf;

public class Later implements Runnable {
    @Override
    public void run() {
        Work.hit("Later");
    }
}
