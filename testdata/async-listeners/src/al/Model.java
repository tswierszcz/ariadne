package al;

import java.beans.PropertyChangeListener;
import java.beans.PropertyChangeSupport;

/** A bean with bound properties. */
public class Model {
    private final PropertyChangeSupport pcs = new PropertyChangeSupport(this);
    private int x;

    public void addListener(PropertyChangeListener l) {
        pcs.addPropertyChangeListener(l);
    }

    public void setX(int v) {
        int old = x;
        x = v;
        pcs.firePropertyChange("x", old, v);
    }
}
