package fx.store;

import fx.trace.Ran;
import java.util.Collections;

/** The provider's query: getResultList records every frame on the stack. */
public class FxQuery implements javax.persistence.Query {
    public java.util.List getResultList() { Ran.stack(); return Collections.singletonList("Leo"); }

    // ---- unsupported
    public boolean isBound(javax.persistence.Parameter<?> a0) { throw new UnsupportedOperationException("isBound"); }
    public int executeUpdate() { throw new UnsupportedOperationException("executeUpdate"); }
    public int getFirstResult() { throw new UnsupportedOperationException("getFirstResult"); }
    public int getMaxResults() { throw new UnsupportedOperationException("getMaxResults"); }
    public java.lang.Object getParameterValue(int a0) { throw new UnsupportedOperationException("getParameterValue"); }
    public java.lang.Object getParameterValue(java.lang.String a0) { throw new UnsupportedOperationException("getParameterValue"); }
    public java.lang.Object getSingleResult() { throw new UnsupportedOperationException("getSingleResult"); }
    public java.util.Map<java.lang.String, java.lang.Object> getHints() { throw new UnsupportedOperationException("getHints"); }
    public java.util.Set<javax.persistence.Parameter<?>> getParameters() { throw new UnsupportedOperationException("getParameters"); }
    public javax.persistence.FlushModeType getFlushMode() { throw new UnsupportedOperationException("getFlushMode"); }
    public javax.persistence.LockModeType getLockMode() { throw new UnsupportedOperationException("getLockMode"); }
    public javax.persistence.Parameter<?> getParameter(int a0) { throw new UnsupportedOperationException("getParameter"); }
    public javax.persistence.Parameter<?> getParameter(java.lang.String a0) { throw new UnsupportedOperationException("getParameter"); }
    public javax.persistence.Query setFirstResult(int a0) { throw new UnsupportedOperationException("setFirstResult"); }
    public javax.persistence.Query setFlushMode(javax.persistence.FlushModeType a0) { throw new UnsupportedOperationException("setFlushMode"); }
    public javax.persistence.Query setHint(java.lang.String a0, java.lang.Object a1) { throw new UnsupportedOperationException("setHint"); }
    public javax.persistence.Query setLockMode(javax.persistence.LockModeType a0) { throw new UnsupportedOperationException("setLockMode"); }
    public javax.persistence.Query setMaxResults(int a0) { throw new UnsupportedOperationException("setMaxResults"); }
    public javax.persistence.Query setParameter(int a0, java.lang.Object a1) { throw new UnsupportedOperationException("setParameter"); }
    public javax.persistence.Query setParameter(int a0, java.util.Calendar a1, javax.persistence.TemporalType a2) { throw new UnsupportedOperationException("setParameter"); }
    public javax.persistence.Query setParameter(int a0, java.util.Date a1, javax.persistence.TemporalType a2) { throw new UnsupportedOperationException("setParameter"); }
    public javax.persistence.Query setParameter(java.lang.String a0, java.lang.Object a1) { throw new UnsupportedOperationException("setParameter"); }
    public javax.persistence.Query setParameter(java.lang.String a0, java.util.Calendar a1, javax.persistence.TemporalType a2) { throw new UnsupportedOperationException("setParameter"); }
    public javax.persistence.Query setParameter(java.lang.String a0, java.util.Date a1, javax.persistence.TemporalType a2) { throw new UnsupportedOperationException("setParameter"); }
    public javax.persistence.Query setParameter(javax.persistence.Parameter<java.util.Calendar> a0, java.util.Calendar a1, javax.persistence.TemporalType a2) { throw new UnsupportedOperationException("setParameter"); }
    public javax.persistence.Query setParameter(javax.persistence.Parameter<java.util.Date> a0, java.util.Date a1, javax.persistence.TemporalType a2) { throw new UnsupportedOperationException("setParameter"); }
    public <T> javax.persistence.Parameter<T> getParameter(int a0, java.lang.Class<T> a1) { throw new UnsupportedOperationException("getParameter"); }
    public <T> javax.persistence.Parameter<T> getParameter(java.lang.String a0, java.lang.Class<T> a1) { throw new UnsupportedOperationException("getParameter"); }
    public <T> javax.persistence.Query setParameter(javax.persistence.Parameter<T> a0, T a1) { throw new UnsupportedOperationException("setParameter"); }
    public <T> T getParameterValue(javax.persistence.Parameter<T> a0) { throw new UnsupportedOperationException("getParameterValue"); }
    public <T> T unwrap(java.lang.Class<T> a0) { throw new UnsupportedOperationException("unwrap"); }
}
