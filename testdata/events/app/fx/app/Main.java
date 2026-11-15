package fx.app;

import fx.ev.OrderService;
import fx.ev.SignupService;
import fx.ev.Ticker;
import fx.trace.Ran;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;

/**
 * Spring application events oracle (spring:events): main publishes events
 * through services — a payload object, an ApplicationEvent, one inside a
 * transaction — and waits for a @Scheduled method to run. The listeners'
 * and the scheduled bodies record the stack. scripts/events-oracle.sh
 * writes the record, cmd/ariadne TestEventsOracle compares the graph with it.
 */
public final class Main {
    public static void main(String[] args) throws Exception {
        AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext(EventsConfig.class);
        ctx.getBean(SignupService.class).signup("ann");
        OrderService orders = ctx.getBean(OrderService.class);
        orders.place();
        orders.cancel();
        ctx.getBean(Ticker.class).await();
        ctx.close();
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the recorded stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
