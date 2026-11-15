package fx.ev;

import fx.trace.Ran;
import org.springframework.context.ApplicationListener;
import org.springframework.stereotype.Component;

/** The interface form. */
@Component
public class CancelListener implements ApplicationListener<OrderCancelled> {
    public void onApplicationEvent(OrderCancelled e) { Ran.stack(); }
}
