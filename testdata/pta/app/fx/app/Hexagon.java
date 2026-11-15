package fx.app;

import fx.trace.Ran;

public class Hexagon implements Shape {
    public double area() { Ran.call(); return 1; }
}
