package al;

import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;

/** Registered with a model whose property never changes. */
public class Unseen implements PropertyChangeListener {
    @Override
    public void propertyChange(PropertyChangeEvent evt) {
        Work.hit("unseen");
    }
}
