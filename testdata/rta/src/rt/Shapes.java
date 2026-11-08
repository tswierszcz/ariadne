package rt;

/** Three implementations created and called, one created only where nothing calls, one never created. */
interface Shape {
    double area();
}

final class Square implements Shape {
    public double area() {
        Ran.here();
        return 4;
    }
}

final class Circle implements Shape {
    public double area() {
        Ran.here();
        return 3.14;
    }
}

/** Created only in a factory method (indirect allocation). */
final class Triangle implements Shape {
    public double area() {
        Ran.here();
        return 0.5;
    }
}

/** Created only in a method nothing calls: RTA (roots profile) removes its edge. */
final class Hidden implements Shape {
    public double area() {
        Ran.here();
        return -1;
    }

    static Shape unused() {
        return new Hidden();
    }
}

/** Never created anywhere: RTA removes its edge in every scenario. */
final class Ghost implements Shape {
    public double area() {
        Ran.here();
        return -2;
    }
}

final class Factory {
    static Shape make() {
        return new Triangle();
    }
}
