package rx;

public class Printer extends Base {
    @Override
    public void onNext(String item) {
        Work.hit("Printer " + item);
    }
}
