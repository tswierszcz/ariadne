package fx.beans;

import org.springframework.context.ApplicationEvent;

public class Tick extends ApplicationEvent {
    public Tick() { super("tick"); }
}
