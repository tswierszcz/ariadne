package rx;

/** onNext throws: the publisher ends the subscription with onError. */
public class Thrower extends Base {
    @Override
    public void onNext(String item) {
        Work.hit("Thrower " + item);
        throw new IllegalStateException("boom");
    }

    @Override
    public void onError(Throwable t) {
        Work.hit("Thrower onError " + t.getMessage());
        done.countDown();
    }
}
