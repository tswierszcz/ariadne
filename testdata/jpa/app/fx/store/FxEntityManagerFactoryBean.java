package fx.store;

import javax.persistence.EntityManagerFactory;
import org.springframework.orm.jpa.AbstractEntityManagerFactoryBean;

/**
 * What LocalContainerEntityManagerFactoryBean is for a real provider: Spring
 * wraps the native factory in its managed-factory proxy
 * (ManagedEntityManagerFactoryInvocationHandler), and every entity manager
 * that proxy hands out in an application-managed extended one
 * (ExtendedEntityManagerInvocationHandler).
 */
public class FxEntityManagerFactoryBean extends AbstractEntityManagerFactoryBean {
    @Override
    protected EntityManagerFactory createNativeEntityManagerFactory() {
        return new FxEntityManagerFactory();
    }
}
