package fx.store;

import fx.trace.Ran;
import java.util.Collections;

/**
 * The provider's native entity manager factory: createEntityManager records
 * every frame on the stack and hands out an FxEntityManager.
 */
public class FxEntityManagerFactory implements javax.persistence.EntityManagerFactory {
    private boolean open = true;

    public javax.persistence.EntityManager createEntityManager() { Ran.stack(); return new FxEntityManager(); }
    public javax.persistence.EntityManager createEntityManager(java.util.Map a0) { Ran.stack(); return new FxEntityManager(); }
    public boolean isOpen() { return open; }
    public void close() { open = false; }
    public java.util.Map<java.lang.String, java.lang.Object> getProperties() { return Collections.emptyMap(); }

    // ---- unsupported
    public javax.persistence.Cache getCache() { throw new UnsupportedOperationException("getCache"); }
    public javax.persistence.criteria.CriteriaBuilder getCriteriaBuilder() { throw new UnsupportedOperationException("getCriteriaBuilder"); }
    public javax.persistence.EntityManager createEntityManager(javax.persistence.SynchronizationType a0, java.util.Map a1) { throw new UnsupportedOperationException("createEntityManager"); }
    public javax.persistence.EntityManager createEntityManager(javax.persistence.SynchronizationType a0) { throw new UnsupportedOperationException("createEntityManager"); }
    public javax.persistence.metamodel.Metamodel getMetamodel() { throw new UnsupportedOperationException("getMetamodel"); }
    public javax.persistence.PersistenceUnitUtil getPersistenceUnitUtil() { throw new UnsupportedOperationException("getPersistenceUnitUtil"); }
    public <T> T unwrap(java.lang.Class<T> a0) { throw new UnsupportedOperationException("unwrap"); }
    public <T> void addNamedEntityGraph(java.lang.String a0, javax.persistence.EntityGraph<T> a1) { throw new UnsupportedOperationException("addNamedEntityGraph"); }
    public void addNamedQuery(java.lang.String a0, javax.persistence.Query a1) { throw new UnsupportedOperationException("addNamedQuery"); }
}
