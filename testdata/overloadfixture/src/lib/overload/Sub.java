package lib.overload;

/** Overrides only handle(Integer): handle(String) still runs Base's body. */
public class Sub extends Base {
    @Override
    public void handle(Integer i) {}
}
