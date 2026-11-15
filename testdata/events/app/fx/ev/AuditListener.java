package fx.ev;

import fx.trace.Ran;
import org.springframework.context.event.EventListener;

/** No stereotype: EventsConfig#audit registers it. */
public class AuditListener {
    @EventListener
    public void onPlaced(OrderPlaced e) { Ran.stack(); }
}
