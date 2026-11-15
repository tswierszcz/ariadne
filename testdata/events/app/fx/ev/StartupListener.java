package fx.ev;

import fx.trace.Ran;
import org.springframework.context.event.ContextRefreshedEvent;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

/** No parameter: the event type comes from the annotation's classes element; the container publishes it. */
@Component
public class StartupListener {
    @EventListener(ContextRefreshedEvent.class)
    public void ready() { Ran.stack(); }
}
