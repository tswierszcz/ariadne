package fx.app;

import fx.store.FxEntityManagerFactory;
import fx.store.FxEntityManagerFactoryBean;
import fx.trace.Ran;
import java.util.List;
import javax.persistence.EntityManager;
import javax.persistence.EntityManagerFactory;
import org.springframework.orm.jpa.SharedEntityManagerCreator;

/**
 * Spring ORM JPA proxy (framework:proxies spring-orm:*) oracle (ROADMAP B4):
 * main uses the three proxies Spring hands an application — the shared
 * entity manager (@PersistenceContext), the query it defers, and the
 * managed entity manager factory with the extended entity managers it
 * creates. The JVM runs Spring's invocation handlers down to the
 * provider's bodies, which record the whole stack.
 * scripts/jpa-oracle.sh writes the record, cmd/ariadne TestJPAOracle
 * compares the graph with it.
 */
public final class Main {
    /** @PersistenceContext: SharedEntityManagerInvocationHandler, then DeferredQueryInvocationHandler. */
    static String shared(EntityManagerFactory emf) {
        EntityManager em = SharedEntityManagerCreator.createSharedEntityManager(emf);
        Object pet = em.find(Pet.class, 1L);
        List<?> names = em.createQuery("select p.name from Pet p").getResultList();
        return pet + " " + names;
    }

    /** @PersistenceUnit: ManagedEntityManagerFactoryInvocationHandler, then ExtendedEntityManagerInvocationHandler. */
    static String managed() {
        FxEntityManagerFactoryBean bean = new FxEntityManagerFactoryBean();
        bean.afterPropertiesSet();
        EntityManager em = bean.getObject().createEntityManager();
        try {
            return String.valueOf(em.find(Pet.class, 2L));
        } finally {
            em.close();
        }
    }

    public static void main(String[] args) {
        System.out.println("# " + shared(new FxEntityManagerFactory()) + " / " + managed());
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the provider's stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
