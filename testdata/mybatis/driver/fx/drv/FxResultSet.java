package fx.drv;

import fx.stub.StubResultSet;
import fx.trace.Ran;
import java.sql.ResultSetMetaData;

public final class FxResultSet extends StubResultSet {
    private int row;
    @Override public boolean next() { Ran.stack(); return row++ == 0; }
    @Override public String getString(String column) { Ran.stack(); return "fx"; }
    @Override public String getString(int column) { Ran.stack(); return "fx"; }
    @Override public boolean wasNull() { return false; }
    @Override public int getType() { return TYPE_FORWARD_ONLY; }
    @Override public ResultSetMetaData getMetaData() { return new FxMetaData(); }
    @Override public boolean isClosed() { return row > 1; }
    @Override public void close() { Ran.stack(); }
}
