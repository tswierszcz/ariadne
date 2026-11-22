package fx.beans;

import fx.trace.Ran;
import javax.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

/**
 * FRAMEWORKS F5 f5-3: four Store beans, four ways Spring picks one. main
 * prints what the container injected ("# inject ..."), which
 * TestDIInjectionOracle compares with profile.ResolveInjections.
 */
@Component
public class Shop {
    /** No qualifier: the @Primary bean. */
    @Autowired
    Store store;

    /** A qualifier naming a bean's explicit name. */
    @Autowired
    @Qualifier("slow")
    Store slow;

    /** A qualifier the bean class itself carries. */
    @Autowired
    @Qualifier("archive")
    Store archive;

    /** A qualifier naming a bean's default name (its decapitalized simple name). */
    @Autowired
    @Qualifier("coldStore")
    Store cold;

    public Shop() { Ran.stack(); }

    @PostConstruct
    void report() {
        Ran.stack();
        for (Object[] p : new Object[][] {{"store", store}, {"slow", slow}, {"archive", archive}, {"cold", cold}}) {
            Store s = (Store) p[1];
            s.save();
            System.out.println("# inject fx.beans.Shop#" + p[0] + " = " + s.getClass().getName());
        }
    }
}
