package fx.app;

import fx.trace.Ran;
import org.springframework.context.support.ClassPathXmlApplicationContext;

/**
 * Spring XML bean definitions oracle (ROADMAP B4 b4-8): main starts a
 * context from fx/app/beans.xml, asks it for one bean and closes it. The
 * container runs the beans' constructors, property setters and init/destroy
 * methods through reflection; each of those bodies records the whole stack.
 * scripts/xmlbeans-oracle.sh writes the record, cmd/ariadne
 * TestXMLBeansOracle compares the graph with it.
 */
public final class Main {
    public static void main(String[] args) {
        ClassPathXmlApplicationContext ctx = new ClassPathXmlApplicationContext("fx/app/beans.xml");
        String said = ((Service) ctx.getBean("service")).say();
        ctx.close();
        System.out.println("# " + said);
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the beans' stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
