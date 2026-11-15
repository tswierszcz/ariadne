package fx.app;

/** A final field written from a constructor parameter: the writer's argument is not tracked — no proof. */
final class Wrapper {
    private final Shape shape;

    Wrapper(Shape shape) { this.shape = shape; }

    double area() { return shape.area(); }
}
