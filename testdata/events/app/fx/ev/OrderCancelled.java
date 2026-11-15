package fx.ev;

import org.springframework.context.ApplicationEvent;

/** A classic ApplicationEvent subclass. */
public class OrderCancelled extends ApplicationEvent {
    public OrderCancelled(Object source) { super(source); }
}
