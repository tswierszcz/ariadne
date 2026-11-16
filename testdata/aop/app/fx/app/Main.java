package fx.app;

import fx.aop.svc.OrderService;
import fx.aop.svc.ReportService;
import fx.aop.svc.Repo;
import fx.trace.Ran;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;

/**
 * Spring AOP oracle (spring:aop): main calls advised beans — a class
 * proxied with CGLIB (a self-invocation, a final method, an annotated
 * method), a bean behind a JDK proxy (a normal return and a thrown
 * exception) and a class-annotated bean. Advice and target bodies record
 * the stack. scripts/aop-oracle.sh writes the record, cmd/ariadne
 * TestAOPOracle compares the graph with it.
 */
public final class Main {
    public static void main(String[] args) throws Exception {
        AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext(AopConfig.class);
        OrderService orders = ctx.getBean(OrderService.class);
        orders.place("ann");
        orders.cancel(7);
        orders.fixed();
        Repo repo = ctx.getBean(Repo.class);
        repo.save("order");
        try {
            repo.find(7L);
        } catch (IllegalStateException expected) {
            // @AfterThrowing ran
        }
        ctx.getBean(ReportService.class).render();
        ctx.close();
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the recorded stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
        System.out.println("# advice runs: advice <- join point (declaring type#method)");
        for (String a : Ran.advised) {
            System.out.println("advised " + a);
        }
    }
}
