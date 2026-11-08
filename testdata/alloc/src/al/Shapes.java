package al;

import java.io.IOException;
import java.io.ObjectInputStream;
import java.io.Serializable;

interface Shape {
    double area();
}

/** Created with an explicit `new`. */
class Square implements Shape {
    Square() { Made.it(this); }
    public double area() { return 1; }
}

/** Created only inside a factory method. */
class Circle implements Shape {
    Circle() { Made.it(this); }
    public double area() { return 3.14; }
}

/** Created only by reflection with a literal class name. */
class Hexagon implements Shape {
    public Hexagon() { Made.it(this); }
    public double area() { return 2.6; }
}

/** Created only by deserialization: no constructor of its own runs, its readObject does. */
class Saved implements Serializable, Shape {
    private static final long serialVersionUID = 1L;
    public double area() { return 0; }
    private void readObject(ObjectInputStream in) throws IOException, ClassNotFoundException {
        in.defaultReadObject();
        Made.it(this);
    }
}

/** A subclass created with `new`: its superclass constructor records the runtime class. */
class BigSquare extends Square {
    BigSquare() { super(); }
}

/** Allocated only in a method nothing calls: in the catalog, never observed. */
class Never implements Shape {
    Never() { Made.it(this); }
    public double area() { return -1; }
}

/** Abstract: never a created class. */
abstract class AbstractShape implements Shape {
}
