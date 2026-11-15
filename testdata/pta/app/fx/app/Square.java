package fx.app;

import fx.trace.Ran;

public class Square implements Shape {
    public double area() { Ran.call(); return 1; }
}
