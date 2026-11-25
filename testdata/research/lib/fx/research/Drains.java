package fx.research;

import java.nio.file.Path;

/** Case (j): hands a path to a sink the caller passes (a parameter of the entry itself). */
public final class Drains {
    private Drains() {}

    public static void drain(Sink sink, Path p) throws java.io.IOException {
        sink.accept(p);
    }

    /** Case (j): a dependency-type callback as a parameter of the entry itself (lb2-7, not reported per entry). */
    public static void hand(fx.dep.Callback cb, Path p) throws java.io.IOException {
        cb.call(p);
    }
}
