package fx.app;

/** A final field written once, in the constructor, with a known allocation. */
final class Holder {
    private final Shape shape;

    Holder() { this.shape = new Circle(); }

    double area() { return shape.area(); }
}
