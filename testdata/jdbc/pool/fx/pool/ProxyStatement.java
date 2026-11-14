package fx.pool;

import fx.trace.Ran;
import java.sql.PreparedStatement;
import java.sql.SQLException;

final class ProxyStatement extends AbstractProxyStatement {
    ProxyStatement(PreparedStatement delegate) { super(delegate); }
    @Override public void close() {
        Ran.here();
        try {
            delegate.close();
        } catch (SQLException e) {
            throw new IllegalStateException(e);
        }
    }
}
