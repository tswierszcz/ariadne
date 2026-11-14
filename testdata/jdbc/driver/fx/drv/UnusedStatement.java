package fx.drv;

import fx.stub.StubPreparedStatement;
import fx.trace.Ran;
import java.sql.ResultSet;

/**
 * In the registered driver's jar, created only by a method nothing calls:
 * the JVM never runs it. The service-provider dispatch model expands a
 * product interface to the provider jar's implementations without a creator
 * gate (as testdata/spidispatch's UnusedCodec), so it is a graph target by
 * design.
 */
public final class UnusedStatement extends StubPreparedStatement {
    @Override public ResultSet executeQuery() { Ran.here(); return null; }
    static UnusedStatement neverCalled() { return new UnusedStatement(); }
}
