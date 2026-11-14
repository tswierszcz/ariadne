package fx.drv;

import fx.stub.StubPreparedStatement;
import fx.trace.Ran;
import java.sql.ResultSet;

public final class FxStatement extends StubPreparedStatement {
    @Override public ResultSet executeQuery() { Ran.here(); return new FxResultSet(); }
    @Override public void close() { Ran.here(); }
}
