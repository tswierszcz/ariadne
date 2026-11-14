package fx.app;

import fx.trace.Ran;
import java.lang.reflect.Proxy;
import org.apache.ibatis.datasource.unpooled.UnpooledDataSource;
import org.apache.ibatis.mapping.Environment;
import org.apache.ibatis.session.Configuration;
import org.apache.ibatis.session.SqlSession;
import org.apache.ibatis.session.SqlSessionFactory;
import org.apache.ibatis.session.SqlSessionFactoryBuilder;
import org.apache.ibatis.transaction.jdbc.JdbcTransactionFactory;

/**
 * MyBatis mapper proxy (framework:proxies) oracle (ROADMAP B4): main calls
 * a mapper method; the JVM runs MyBatis's MapperProxy and executor down to
 * the driver, whose bodies record the whole stack. scripts/mybatis-oracle.sh
 * writes the record, cmd/ariadne TestMyBatisOracle compares the graph with it.
 */
public final class Main {
    static String query() {
        UnpooledDataSource ds = new UnpooledDataSource("fx.drv.FxDriver", "jdbc:fx:pets", null, null);
        Configuration cfg = new Configuration(new Environment("oracle", new JdbcTransactionFactory(), ds));
        cfg.addMapper(PetMapper.class);
        SqlSessionFactory factory = new SqlSessionFactoryBuilder().build(cfg);
        try (SqlSession session = factory.openSession()) {
            return session.getMapper(PetMapper.class).name();
        }
    }

    // Negative for the catalog, unresolved for the graph: a JDK proxy of an
    // interface that is not a mapper, with the application's own handler.
    static String ownProxy() {
        NotAMapper p = (NotAMapper) Proxy.newProxyInstance(Main.class.getClassLoader(), new Class<?>[] {NotAMapper.class},
            (proxy, method, args) -> {
                Ran.stack();
                return "own";
            });
        return p.name();
    }

    public static void main(String[] args) {
        System.out.println("# " + query());
        System.out.println("# " + ownProxy());
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the driver's stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
