package al;

import java.util.function.Consumer;

/** A listener of a JDK type: CHA never expands Consumer#accept. */
public class D implements Consumer<String> {
    @Override
    public void accept(String e) {
        Work.hit("D");
    }
}
