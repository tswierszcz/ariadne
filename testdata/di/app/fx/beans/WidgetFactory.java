package fx.beans;

import fx.trace.Ran;
import org.springframework.beans.factory.FactoryBean;
import org.springframework.stereotype.Component;

/** A FactoryBean: the container asks it for the Widget Greeter needs. */
@Component
public class WidgetFactory implements FactoryBean<Widget> {
    public Widget getObject() { Ran.stack(); return new Widget("w"); }
    public Class<?> getObjectType() { return Widget.class; }
}
