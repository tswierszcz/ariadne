package fx.ev;

import org.springframework.transaction.TransactionDefinition;
import org.springframework.transaction.support.AbstractPlatformTransactionManager;
import org.springframework.transaction.support.DefaultTransactionStatus;

/** A transaction manager with synchronization (what DataSourceTransactionManager is), so @TransactionalEventListener fires. */
public class FxTransactionManager extends AbstractPlatformTransactionManager {
    protected Object doGetTransaction() { return new Object(); }
    protected void doBegin(Object transaction, TransactionDefinition definition) {}
    protected void doCommit(DefaultTransactionStatus status) {}
    protected void doRollback(DefaultTransactionStatus status) {}
}
