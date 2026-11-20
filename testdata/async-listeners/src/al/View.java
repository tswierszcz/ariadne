package al;

import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;

public class View implements PropertyChangeListener {
    @Override
    public void propertyChange(PropertyChangeEvent evt) {
        Work.hit("view " + evt.getPropertyName());
    }
}
