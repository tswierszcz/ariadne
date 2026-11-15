package fx.ev;

import fx.trace.Ran;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

/** Registered, but nothing publishes its event type: never runs. */
@Component
public class GhostListener {
    @EventListener
    public void onGhost(NeverPublished e) { Ran.stack(); }
}
