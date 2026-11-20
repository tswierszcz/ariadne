package al;

import java.beans.PropertyChangeListener;
import java.beans.PropertyChangeSupport;

/** Registers listeners, never fires. */
public class QuietModel {
    private final PropertyChangeSupport pcs = new PropertyChangeSupport(this);

    public void addListener(PropertyChangeListener l) {
        pcs.addPropertyChangeListener("y", l);
    }
}
