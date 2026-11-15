package fx.store;

/** A resource-local transaction that is never active. */
public class FxTransaction implements javax.persistence.EntityTransaction {
    public void begin() { throw new UnsupportedOperationException("begin"); }
    public void commit() { throw new UnsupportedOperationException("commit"); }
    public void rollback() { throw new UnsupportedOperationException("rollback"); }
    public void setRollbackOnly() { throw new UnsupportedOperationException("setRollbackOnly"); }
    public boolean getRollbackOnly() { return false; }
    public boolean isActive() { return false; }
}
