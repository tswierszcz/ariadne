package fx.store;

import fx.trace.Ran;
import org.springframework.data.repository.Repository;

/**
 * A concrete base-repository-like class nothing creates or names: the
 * base-repository rule (framework:proxies x2) must not target it.
 */
public class OrphanRepository implements Repository<Object, Long> {
    public long count() { Ran.stack(); return -1; }
}
