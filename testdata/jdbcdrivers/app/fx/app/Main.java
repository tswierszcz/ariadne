package fx.app;

import java.sql.Connection;
import java.sql.DriverManager;

/**
 * Opens one connection through DriverManager to a URL only the fixture's
 * own driver accepts. pgjdbc and MySQL Connector/J are registered too (and
 * listed first on the class path): DriverManager offers each of them the
 * URL, and each checks it and returns null.
 */
public class Main {
    public static void main(String[] args) throws Exception {
        Connection c = DriverManager.getConnection("jdbc:fx:oracle");
        System.out.println("connected: " + (c != null));
    }
}
