package fx.ev;

import fx.trace.Ran;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

/** @EventListener on the exact payload type. */
@Component
public class MailListener {
    @EventListener
    public void onSignup(UserSignedUp e) { Ran.stack(); }
}
