package al;

import java.util.concurrent.CopyOnWriteArrayList;

/** An indexed walk (get(i)) over a copy-on-write list. */
public class IndexedBus {
    private final CopyOnWriteArrayList<Listener> listeners = new CopyOnWriteArrayList<>();

    public void register(Listener l) {
        listeners.add(l);
    }

    public void publish(String e) {
        for (int i = 0; i < listeners.size(); i++) {
            listeners.get(i).onEvent(e);
        }
    }
}
