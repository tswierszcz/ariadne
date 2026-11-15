package fx.beans;

import fx.trace.Ran;
import org.springframework.context.ApplicationListener;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

/** An ApplicationListener and an @EventListener method for the same event. */
@Component
public class TickListener implements ApplicationListener<Tick> {
    public void onApplicationEvent(Tick tick) { Ran.stack(); }

    @EventListener
    public void onTick(Tick tick) { Ran.stack(); }
}
