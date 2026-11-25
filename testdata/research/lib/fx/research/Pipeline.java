package fx.research;

import java.nio.file.Path;

/** Case (j): hands a path to a sink the caller supplied to the constructor (a field). */
public final class Pipeline {
    private final Sink sink;

    public Pipeline(Sink sink) {
        this.sink = sink;
    }

    public void run(Path p) throws java.io.IOException {
        sink.accept(p);
    }
}
