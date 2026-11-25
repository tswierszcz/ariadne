package fx.research;

import fx.dep.Callback;
import java.nio.file.Path;

/** Case (j): calls a dependency-type callback the caller supplied to the constructor (a field: supplied dispatch). */
public final class Relay {
    private final Callback cb;

    public Relay(Callback cb) {
        this.cb = cb;
    }

    public void send(Path p) throws java.io.IOException {
        cb.call(p);
    }
}
