package al;

import java.util.ArrayList;
import java.util.List;
import java.util.function.Consumer;

/** Registers, notifies — but nothing ever calls publish. */
public class SilentBus {
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
