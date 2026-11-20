package ah;

import java.util.ArrayList;
import java.util.List;

/** A listener list: fire() walks it without removing — a broadcast, not a drain. */
public class Listeners {
    private final List<Work> listeners = new ArrayList<>();

    public void add(Work w) {
        listeners.add(w);
    }

    public void fire() {
        for (Work w : listeners) {
            w.work();
        }
    }
}
