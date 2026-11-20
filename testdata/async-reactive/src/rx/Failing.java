package rx;

public class Failing extends Base {
    @Override
    public void onNext(String item) {
        Work.hit("Failing " + item);
    }
}
