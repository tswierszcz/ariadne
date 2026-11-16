package fx.aop.svc;

import fx.trace.Ran;
import org.springframework.stereotype.Repository;

/** Implements an interface: Spring proxies it with a JDK proxy. */
@Repository
public class JdbcRepo implements Repo {
    @Override
    public void save(Object o) { Ran.stack(); }

    @Override
    public Object find(long id) {
        Ran.stack();
        throw new IllegalStateException("no row " + id);
    }
}
