package fx.pool;

import fx.stub.StubConnection;
import fx.trace.Ran;
import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.SQLException;

final class ProxyConnection extends StubConnection {
    private final Connection delegate;
    ProxyConnection(Connection delegate) { this.delegate = delegate; }
    @Override public PreparedStatement prepareStatement(String sql) {
        Ran.here();
        try {
            return new ProxyStatement(delegate.prepareStatement(sql));
        } catch (SQLException e) {
            throw new IllegalStateException(e);
        }
    }
    @Override public void close() {
        Ran.here();
        try {
            delegate.close();
        } catch (SQLException e) {
            throw new IllegalStateException(e);
        }
    }
}
