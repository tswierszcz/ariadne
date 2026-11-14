package fx.drv;

import fx.stub.StubPreparedStatement;
import fx.trace.Ran;
import java.sql.ResultSet;

/**
 * In the registered driver's jar, created only by a method nothing calls:
 * the JVM never runs it. (The service-provider dispatch model expands a
 * product interface to the provider jar's implementations without a creator
 * gate, as for testdata/spidispatch's UnusedCodec; here no call on
 * java.sql.Driver is in the input, so that expansion does not happen.)
 */
public final class UnusedStatement extends StubPreparedStatement {
    @Override public ResultSet executeQuery() { Ran.here(); return null; }
    static UnusedStatement neverCalled() { return new UnusedStatement(); }
}
