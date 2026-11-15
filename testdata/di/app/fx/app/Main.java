package fx.app;

import fx.beans.Greeter;
import fx.beans.Tick;
import fx.trace.Ran;
import org.springframework.context.annotation.AnnotationConfigApplicationContext;

/**
 * Spring dependency injection and bean lifecycle oracle (ROADMAP B4 b4-7):
 * main starts an annotation-config context over fx.beans, asks it for one
 * bean, publishes an event and closes it. The container runs the beans'
 * constructors, injection methods, lifecycle annotations and callback
 * interfaces through reflection or its own interfaces; each of those bodies
 * records the whole stack. scripts/di-oracle.sh writes the record,
 * cmd/ariadne TestDIOracle compares the graph with it.
 */
public final class Main {
    public static void main(String[] args) {
        AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext(AppConfig.class);
        String greeting = ctx.getBean(Greeter.class).greet();
        ctx.publishEvent(new Tick());
        ctx.close();
        System.out.println("# " + greeting);
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the beans' stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
