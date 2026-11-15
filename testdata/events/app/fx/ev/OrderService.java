package fx.ev;

import org.springframework.context.ApplicationEventPublisher;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
public class OrderService {
    private final ApplicationEventPublisher events;

    public OrderService(ApplicationEventPublisher events) { this.events = events; }

    /** Publishes inside a transaction: the @TransactionalEventListener runs after commit. */
    @Transactional
    public void place() { events.publishEvent(new OrderPlaced()); }

    public void cancel() { events.publishEvent(new OrderCancelled(this)); }
}
