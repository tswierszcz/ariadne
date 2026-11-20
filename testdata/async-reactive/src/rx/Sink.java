package rx;

public class Sink extends Base {
    @Override
    public void onNext(String item) {
        Work.hit("Sink " + item);
    }
}
