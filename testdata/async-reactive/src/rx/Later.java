package rx;

public class Later extends Base {
    @Override
    public void onNext(String item) {
        Work.hit("Later " + item);
    }
}
