package fx;

public class Derived extends Base {
    private Object readResolve() { return this; }
}
