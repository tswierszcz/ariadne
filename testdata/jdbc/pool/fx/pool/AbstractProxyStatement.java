package fx.pool;

import fx.stub.StubPreparedStatement;
import fx.trace.Ran;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;

/** Declares executeQuery for the proxy that inherits it (Hikari's ProxyPreparedStatement shape). */
abstract class AbstractProxyStatement extends StubPreparedStatement {
    final PreparedStatement delegate;
    AbstractProxyStatement(PreparedStatement delegate) { this.delegate = delegate; }
    @Override public ResultSet executeQuery() {
        Ran.here();
        try {
            return delegate.executeQuery();
        } catch (SQLException e) {
            throw new IllegalStateException(e);
        }
    }
}
