package fx.other;

import fx.stub.StubPreparedStatement;
import fx.trace.Ran;
import java.sql.ResultSet;

/** Never created anywhere: never a target. */
public final class OrphanStatement extends StubPreparedStatement {
    @Override public ResultSet executeQuery() { Ran.here(); return null; }
}
