package rx;

public class Quiet extends Base {
    @Override
    public void onNext(String item) {
        Work.hit("Quiet " + item);
    }
}
