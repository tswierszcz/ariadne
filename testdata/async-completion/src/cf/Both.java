package cf;

import java.util.function.BiConsumer;

/** whenComplete: runs on either outcome. */
public class Both implements BiConsumer<Object, Throwable> {
    @Override
    public void accept(Object o, Throwable t) {
        Work.hit("Both");
    }
}
