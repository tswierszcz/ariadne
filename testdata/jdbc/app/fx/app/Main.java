package fx.app;

import fx.pool.FxPool;
import fx.trace.Ran;
import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import javax.sql.DataSource;

/**
 * JDBC objects (jdbc:objects) and driver dispatch (service-provider:dispatch)
 * oracle (ROADMAP B4): the application queries through the java.sql
 * interfaces, once with DriverManager and once through a pool's DataSource;
 * scripts/jdbc-oracle.sh records which bodies ran, and cmd/ariadne
 * TestJDBCOracle compares the call graph with that record.
 */
public final class Main {
    static String query(Connection c) throws Exception {
        try (PreparedStatement ps = c.prepareStatement("select 1"); ResultSet rs = ps.executeQuery()) {
            rs.next();
            return rs.getString(1);
        }
    }

    static String direct() throws Exception {
        try (Connection c = DriverManager.getConnection("jdbc:fx:direct")) {
            return query(c);
        }
    }

    static String pooled() throws Exception {
        DataSource ds = new FxPool("jdbc:fx:pooled");
        try (Connection c = ds.getConnection()) {
            return query(c);
        }
    }

    public static void main(String[] args) throws Exception {
        direct();
        pooled();
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
