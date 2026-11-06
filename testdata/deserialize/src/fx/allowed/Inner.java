package fx.allowed;

public class Inner {
    private Object readResolve() { return this; }
}
