package fx.research;

import java.nio.file.Path;

/** Case (e): a template method; subclasses supply handle. */
public abstract class Handler {
    public final void process(Path p) throws java.io.IOException {
        handle(p);
    }

    protected abstract void handle(Path p) throws java.io.IOException;
}
