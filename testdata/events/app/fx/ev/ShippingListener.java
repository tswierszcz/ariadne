package fx.ev;

import fx.trace.Ran;
import org.springframework.stereotype.Component;
import org.springframework.transaction.event.TransactionalEventListener;

/** Runs after the publishing transaction commits. */
@Component
public class ShippingListener {
    @TransactionalEventListener
    public void onPlaced(OrderPlaced e) { Ran.stack(); }
}
