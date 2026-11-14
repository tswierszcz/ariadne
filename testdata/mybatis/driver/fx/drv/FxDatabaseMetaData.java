package fx.drv;

import fx.stub.StubDatabaseMetaData;

public final class FxDatabaseMetaData extends StubDatabaseMetaData {
    @Override public boolean supportsMultipleResultSets() { return false; }
}
