package fx.ev;

import fx.trace.Ran;
import org.springframework.context.event.EventListener;

/** Not a bean (no stereotype, no @Bean, not scanned as one): never runs. */
public class Unregistered {
    @EventListener
    public void onSignup(UserSignedUp e) { Ran.stack(); }
}
