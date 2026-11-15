package fx.store;

import fx.trace.Ran;
import java.util.Collections;
import java.util.List;

/**
 * The fixture's persistence provider entity manager (what Hibernate's
 * SessionImpl is): Spring ORM's proxies call it through Method#invoke. find
 * and createQuery record every frame on the stack; the rest of the
 * interface is unsupported.
 */
public class FxEntityManager implements javax.persistence.EntityManager {
    private boolean open = true;

    public <T> T find(java.lang.Class<T> a0, java.lang.Object a1) { Ran.stack(); return null; }
    public javax.persistence.Query createQuery(java.lang.String a0) { Ran.stack(); return new FxQuery(); }
    public boolean isOpen() { return open; }
    public void close() { open = false; }
    public java.util.Map<java.lang.String, java.lang.Object> getProperties() { return Collections.emptyMap(); }
    public boolean isJoinedToTransaction() { return false; }
    /** A resource-local transaction (Spring asks to tell JTA from resource-local). */
    public javax.persistence.EntityTransaction getTransaction() { return new FxTransaction(); }

    // ---- unsupported
    public boolean contains(java.lang.Object a0) { throw new UnsupportedOperationException("contains"); }
    public java.lang.Object getDelegate() { throw new UnsupportedOperationException("getDelegate"); }
    public javax.persistence.criteria.CriteriaBuilder getCriteriaBuilder() { throw new UnsupportedOperationException("getCriteriaBuilder"); }
    public javax.persistence.EntityGraph<?> createEntityGraph(java.lang.String a0) { throw new UnsupportedOperationException("createEntityGraph"); }
    public javax.persistence.EntityGraph<?> getEntityGraph(java.lang.String a0) { throw new UnsupportedOperationException("getEntityGraph"); }
    public javax.persistence.EntityManagerFactory getEntityManagerFactory() { throw new UnsupportedOperationException("getEntityManagerFactory"); }
    public javax.persistence.FlushModeType getFlushMode() { throw new UnsupportedOperationException("getFlushMode"); }
    public javax.persistence.LockModeType getLockMode(java.lang.Object a0) { throw new UnsupportedOperationException("getLockMode"); }
    public javax.persistence.metamodel.Metamodel getMetamodel() { throw new UnsupportedOperationException("getMetamodel"); }
    public javax.persistence.Query createNamedQuery(java.lang.String a0) { throw new UnsupportedOperationException("createNamedQuery"); }
    public javax.persistence.Query createNativeQuery(java.lang.String a0, java.lang.Class a1) { throw new UnsupportedOperationException("createNativeQuery"); }
    public javax.persistence.Query createNativeQuery(java.lang.String a0, java.lang.String a1) { throw new UnsupportedOperationException("createNativeQuery"); }
    public javax.persistence.Query createNativeQuery(java.lang.String a0) { throw new UnsupportedOperationException("createNativeQuery"); }
    public javax.persistence.Query createQuery(javax.persistence.criteria.CriteriaDelete a0) { throw new UnsupportedOperationException("createQuery"); }
    public javax.persistence.Query createQuery(javax.persistence.criteria.CriteriaUpdate a0) { throw new UnsupportedOperationException("createQuery"); }
    public javax.persistence.StoredProcedureQuery createNamedStoredProcedureQuery(java.lang.String a0) { throw new UnsupportedOperationException("createNamedStoredProcedureQuery"); }
    public javax.persistence.StoredProcedureQuery createStoredProcedureQuery(java.lang.String a0, java.lang.Class... a1) { throw new UnsupportedOperationException("createStoredProcedureQuery"); }
    public javax.persistence.StoredProcedureQuery createStoredProcedureQuery(java.lang.String a0, java.lang.String... a1) { throw new UnsupportedOperationException("createStoredProcedureQuery"); }
    public javax.persistence.StoredProcedureQuery createStoredProcedureQuery(java.lang.String a0) { throw new UnsupportedOperationException("createStoredProcedureQuery"); }
    public <T> java.util.List<javax.persistence.EntityGraph<? super T>> getEntityGraphs(java.lang.Class<T> a0) { throw new UnsupportedOperationException("getEntityGraphs"); }
    public <T> javax.persistence.EntityGraph<T> createEntityGraph(java.lang.Class<T> a0) { throw new UnsupportedOperationException("createEntityGraph"); }
    public <T> javax.persistence.TypedQuery<T> createNamedQuery(java.lang.String a0, java.lang.Class<T> a1) { throw new UnsupportedOperationException("createNamedQuery"); }
    public <T> javax.persistence.TypedQuery<T> createQuery(java.lang.String a0, java.lang.Class<T> a1) { throw new UnsupportedOperationException("createQuery"); }
    public <T> javax.persistence.TypedQuery<T> createQuery(javax.persistence.criteria.CriteriaQuery<T> a0) { throw new UnsupportedOperationException("createQuery"); }
    public <T> T find(java.lang.Class<T> a0, java.lang.Object a1, java.util.Map<java.lang.String, java.lang.Object> a2) { throw new UnsupportedOperationException("find"); }
    public <T> T find(java.lang.Class<T> a0, java.lang.Object a1, javax.persistence.LockModeType a2, java.util.Map<java.lang.String, java.lang.Object> a3) { throw new UnsupportedOperationException("find"); }
    public <T> T find(java.lang.Class<T> a0, java.lang.Object a1, javax.persistence.LockModeType a2) { throw new UnsupportedOperationException("find"); }
    public <T> T getReference(java.lang.Class<T> a0, java.lang.Object a1) { throw new UnsupportedOperationException("getReference"); }
    public <T> T merge(T a0) { throw new UnsupportedOperationException("merge"); }
    public <T> T unwrap(java.lang.Class<T> a0) { throw new UnsupportedOperationException("unwrap"); }
    public void clear() { throw new UnsupportedOperationException("clear"); }
    public void detach(java.lang.Object a0) { throw new UnsupportedOperationException("detach"); }
    public void flush() { throw new UnsupportedOperationException("flush"); }
    public void joinTransaction() { throw new UnsupportedOperationException("joinTransaction"); }
    public void lock(java.lang.Object a0, javax.persistence.LockModeType a1, java.util.Map<java.lang.String, java.lang.Object> a2) { throw new UnsupportedOperationException("lock"); }
    public void lock(java.lang.Object a0, javax.persistence.LockModeType a1) { throw new UnsupportedOperationException("lock"); }
    public void persist(java.lang.Object a0) { throw new UnsupportedOperationException("persist"); }
    public void refresh(java.lang.Object a0, java.util.Map<java.lang.String, java.lang.Object> a1) { throw new UnsupportedOperationException("refresh"); }
    public void refresh(java.lang.Object a0, javax.persistence.LockModeType a1, java.util.Map<java.lang.String, java.lang.Object> a2) { throw new UnsupportedOperationException("refresh"); }
    public void refresh(java.lang.Object a0, javax.persistence.LockModeType a1) { throw new UnsupportedOperationException("refresh"); }
    public void refresh(java.lang.Object a0) { throw new UnsupportedOperationException("refresh"); }
    public void remove(java.lang.Object a0) { throw new UnsupportedOperationException("remove"); }
    public void setFlushMode(javax.persistence.FlushModeType a0) { throw new UnsupportedOperationException("setFlushMode"); }
    public void setProperty(java.lang.String a0, java.lang.Object a1) { throw new UnsupportedOperationException("setProperty"); }
}
