package rx;

public class Idle extends Base {
    @Override
    public void onNext(String item) {
        Work.hit("Idle " + item);
    }
}
