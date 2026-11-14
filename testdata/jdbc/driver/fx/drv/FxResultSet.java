package fx.drv;

import fx.stub.StubResultSet;
import fx.trace.Ran;

public final class FxResultSet extends StubResultSet {
    private int row;
    @Override public boolean next() { Ran.here(); return row++ == 0; }
    @Override public String getString(int column) { Ran.here(); return "fx"; }
    @Override public void close() { Ran.here(); }
}
