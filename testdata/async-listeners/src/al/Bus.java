package al;

import java.util.ArrayList;
import java.util.List;
import java.util.function.Consumer;

/** An event source of JDK-typed listeners: registered listeners stay; publish walks them. */
public class Bus {
    private final List<Consumer<String>> listeners = new ArrayList<>();

    public void register(Consumer<String> l) {
        listeners.add(l);
    }

    public void publish(String e) {
        for (Consumer<String> l : listeners) {
            l.accept(e);
        }
    }
}
