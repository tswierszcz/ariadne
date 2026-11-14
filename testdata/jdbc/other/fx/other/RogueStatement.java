package fx.other;

import fx.stub.StubPreparedStatement;
import fx.trace.Ran;
import java.sql.ResultSet;

/** Outside the driver and the pool, created only by a method nothing calls: never a target. */
public final class RogueStatement extends StubPreparedStatement {
    @Override public ResultSet executeQuery() { Ran.here(); return null; }
    static RogueStatement neverCalled() { return new RogueStatement(); }
}
