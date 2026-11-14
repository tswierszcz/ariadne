package fx.drv;

import fx.stub.StubResultSetMetaData;
import java.sql.Types;

public final class FxMetaData extends StubResultSetMetaData {
    @Override public int getColumnCount() { return 1; }
    @Override public String getColumnLabel(int column) { return "NAME"; }
    @Override public String getColumnName(int column) { return "NAME"; }
    @Override public int getColumnType(int column) { return Types.VARCHAR; }
    @Override public String getColumnClassName(int column) { return "java.lang.String"; }
}
