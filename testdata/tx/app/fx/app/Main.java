package fx.app;

import fx.svc.AccountService;
import fx.svc.Catalog;
import fx.svc.OrderService;
import fx.svc.PlainService;
import fx.trace.Ran;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;

/**
 * Spring transactional proxy oracle (framework:proxies spring-tx:transactional):
 * main calls @Transactional services through the beans the context hands
 * out (CGLIB and JDK proxies), and a plain service. The JVM runs Spring's
 * TransactionInterceptor around the transactional calls, which calls the
 * transaction manager; its bodies and the services' record the stack.
 * scripts/tx-oracle.sh writes the record, cmd/ariadne TestTxOracle compares
 * the graph with it.
 */
public final class Main {
    static String run(AnnotationConfigApplicationContext ctx) {
        AccountService accounts = ctx.getBean(AccountService.class);
        OrderService orders = ctx.getBean(OrderService.class);
        Catalog catalog = ctx.getBean(Catalog.class);
        PlainService plain = ctx.getBean(PlainService.class);
        String out = accounts.find("1") + " " + orders.place() + " " + orders.count() + " " + catalog.title() + " " + plain.hello();
        try {
            accounts.fail();
        } catch (IllegalStateException e) {
            out += " " + e.getMessage();
        }
        return out;
    }

    public static void main(String[] args) {
        AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext(TxConfig.class);
        String out = run(ctx);
        ctx.close();
        System.out.println("# " + out);
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the recorded stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
