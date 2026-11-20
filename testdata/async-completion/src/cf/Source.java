package cf;

public class Source implements Runnable {
    @Override
    public void run() {
        Work.hit("Source");
    }
}
