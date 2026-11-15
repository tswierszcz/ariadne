package fx.store;

import fx.trace.Ran;

/**
 * An entity manager nothing creates: no model may target it (the negative
 * case). Its find
 * records the stack were it ever to run.
 */
public class OrphanEntityManager extends FxEntityManager {
    @Override
    public <T> T find(Class<T> entityClass, Object primaryKey) { Ran.stack(); return null; }
}
