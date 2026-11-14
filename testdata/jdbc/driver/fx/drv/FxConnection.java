package fx.drv;

import fx.stub.StubConnection;
import fx.trace.Ran;
import java.sql.PreparedStatement;

public final class FxConnection extends StubConnection {
    @Override public PreparedStatement prepareStatement(String sql) { Ran.here(); return new FxStatement(); }
    @Override public void close() { Ran.here(); }
}
