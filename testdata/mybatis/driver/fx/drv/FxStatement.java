package fx.drv;

import fx.stub.StubPreparedStatement;
import fx.trace.Ran;
import java.sql.Connection;
import java.sql.ResultSet;

public final class FxStatement extends StubPreparedStatement {
    private final Connection connection;
    private ResultSet current;
    FxStatement(Connection connection) { this.connection = connection; }
    @Override public boolean execute() { Ran.stack(); current = new FxResultSet(); return true; }
    @Override public ResultSet executeQuery() { Ran.stack(); return current = new FxResultSet(); }
    @Override public ResultSet getResultSet() { return current; }
    @Override public boolean getMoreResults() { current = null; return false; }
    @Override public int getUpdateCount() { return -1; }
    @Override public Connection getConnection() { return connection; }
    @Override public void close() { Ran.stack(); }
}
