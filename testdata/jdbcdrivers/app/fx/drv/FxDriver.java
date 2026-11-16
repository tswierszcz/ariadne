package fx.drv;

import java.lang.reflect.Proxy;
import java.sql.Connection;
import java.sql.Driver;
import java.sql.DriverManager;
import java.sql.DriverPropertyInfo;
import java.util.Properties;
import java.util.logging.Logger;

/** Accepts jdbc:fx: URLs; the connection is a do-nothing proxy. */
public class FxDriver implements Driver {
    static {
        try {
            DriverManager.registerDriver(new FxDriver());
        } catch (Exception e) {
            throw new ExceptionInInitializerError(e);
        }
    }

    public Connection connect(String url, Properties info) {
        if (!acceptsURL(url)) {
            return null;
        }
        return (Connection) Proxy.newProxyInstance(FxDriver.class.getClassLoader(),
            new Class<?>[] {Connection.class}, (proxy, method, a) -> null);
    }

    public boolean acceptsURL(String url) { return url.startsWith("jdbc:fx:"); }

    public DriverPropertyInfo[] getPropertyInfo(String url, Properties info) { return new DriverPropertyInfo[0]; }

    public int getMajorVersion() { return 1; }

    public int getMinorVersion() { return 0; }

    public boolean jdbcCompliant() { return false; }

    public Logger getParentLogger() { return Logger.getGlobal(); }
}
