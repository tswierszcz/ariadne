package ah;

import java.nio.channels.SelectionKey;

public class Handler {
    public void ready(SelectionKey k) {
        Main.ran("Handler#ready");
    }
}
