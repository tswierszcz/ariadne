package fx.pool;

import fx.trace.Ran;
import java.io.PrintWriter;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import java.util.logging.Logger;
import javax.sql.DataSource;

/** A pool: hands out proxy connections over the driver's (Hikari-shaped). */
public final class FxPool implements DataSource {
    private final String url;
    public FxPool(String url) { this.url = url; }
    public Connection getConnection() throws SQLException {
        Ran.here();
        return new ProxyConnection(DriverManager.getConnection(url));
    }
    public Connection getConnection(String user, String password) throws SQLException { return getConnection(); }
    public PrintWriter getLogWriter() { return null; }
    public void setLogWriter(PrintWriter out) { }
    public void setLoginTimeout(int seconds) { }
    public int getLoginTimeout() { return 0; }
    public Logger getParentLogger() { return Logger.getGlobal(); }
    public <T> T unwrap(Class<T> iface) { throw new UnsupportedOperationException(); }
    public boolean isWrapperFor(Class<?> iface) { return false; }
}
