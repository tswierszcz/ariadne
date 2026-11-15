package fx.app;

/** A static factory: its result's class is the allocation it returns. */
final class Shapes {
    static Shape square() { return new Square(); }
}
