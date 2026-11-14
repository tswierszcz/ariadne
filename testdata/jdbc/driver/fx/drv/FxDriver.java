package fx.drv;

import fx.trace.Ran;
import java.sql.Connection;
import java.sql.Driver;
import java.sql.DriverManager;
import java.sql.DriverPropertyInfo;
import java.sql.SQLException;
import java.util.Properties;
import java.util.logging.Logger;

/** The driver, registered in META-INF/services/java.sql.Driver. */
public final class FxDriver implements Driver {
    static {
        // JDBC 4: DriverManager instantiates the registered drivers through
        // ServiceLoader; each registers itself here.
        try {
            DriverManager.registerDriver(new FxDriver());
        } catch (SQLException e) {
            throw new ExceptionInInitializerError(e);
        }
    }

    public Connection connect(String url, Properties info) {
        if (!acceptsURL(url)) return null;
        Ran.here();
        return new FxConnection();
    }
    public boolean acceptsURL(String url) { return url.startsWith("jdbc:fx:"); }
    public DriverPropertyInfo[] getPropertyInfo(String url, Properties info) { return new DriverPropertyInfo[0]; }
    public int getMajorVersion() { return 1; }
    public int getMinorVersion() { return 0; }
    public boolean jdbcCompliant() { return false; }
    public Logger getParentLogger() { return Logger.getGlobal(); }
}
