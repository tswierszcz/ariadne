package cf;

import java.util.function.Consumer;

/** A continuation of a future nothing completes. */
public class Never implements Consumer<Object> {
    @Override
    public void accept(Object o) {
        Work.hit("Never");
    }
}
