package fx.ev;

import fx.trace.Ran;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

/** @EventListener on a supertype of the published payload. */
@Component
public class MetricsListener {
    @EventListener
    public void any(DomainEvent e) { Ran.stack(); }
}
