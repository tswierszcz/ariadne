package fx.svc;

import fx.trace.Ran;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.TransactionDefinition;
import org.springframework.transaction.TransactionStatus;
import org.springframework.transaction.support.SimpleTransactionStatus;

/** The fixture's transaction manager (what DataSourceTransactionManager is): its bodies record the stack. */
public class FxTransactionManager implements PlatformTransactionManager {
    public TransactionStatus getTransaction(TransactionDefinition definition) { Ran.stack(); return new SimpleTransactionStatus(); }
    public void commit(TransactionStatus status) { Ran.stack(); }
    public void rollback(TransactionStatus status) { Ran.stack(); }
}
