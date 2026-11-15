package fx.app;

import fx.trace.Ran;

/**
 * GRAPH_PRECISION P3 oracle (p3-1): each method below is one call site of
 * Shape#area. Circle, Square and Hexagon are all created by reachable
 * code, so CHA and RTA give every site all three targets; receiver
 * narrowing keeps only the ones the receiver's allocations select — and
 * every target the JVM ran. scripts/pta-oracle.sh writes the record,
 * cmd/ariadne TestPTAOracle compares the graphs with it.
 */
public final class Main {
    static double twoReceiversA() { Shape a = new Circle(); return a.area(); }

    static double twoReceiversB() { Shape b = new Square(); return b.area(); }

    static double factory() { return Shapes.square().area(); }

    static double field() { return new Holder().area(); }

    private static double param(Shape s) { return s.area(); }

    static double array() { Shape[] shapes = { new Hexagon() }; return shapes[0].area(); }

    static Shape pick(String which) { return which.isEmpty() ? new Circle() : new Hexagon(); }

    static double twoAllocations(String which) { return pick(which).area(); }

    /** Not final: any code may store any Shape here — no proof. */
    static Shape current;

    static double unknown() { return current.area(); }

    public static void main(String[] args) {
        double sum = twoReceiversA() + twoReceiversB() + factory() + field()
            + param(new Circle()) + param(new Square()) + array() + twoAllocations(args.length > 0 ? args[0] : "x");
        current = new Square();
        sum += unknown() + new Wrapper(new Square()).area();
        System.out.println("# " + sum);
        System.out.println("# calls the real JVM made into Shape#area implementations (caller -> target)");
        for (String c : Ran.calls) {
            System.out.println(c);
        }
    }
}
