package ah;

import java.io.IOException;
import java.nio.channels.SelectableChannel;
import java.nio.channels.SelectionKey;
import java.nio.channels.Selector;
import java.util.Iterator;

/** An NIO selection key's attachment: registered here, taken back by loop() when the key is ready. */
public class Sel {
    public static void register(Selector sel, SelectableChannel ch, Handler h) throws IOException {
        ch.configureBlocking(false);
        ch.register(sel, SelectionKey.OP_READ, h);
    }

    public static void loop(Selector sel) throws IOException {
        sel.select(5000);
        Iterator<SelectionKey> it = sel.selectedKeys().iterator();
        while (it.hasNext()) {
            SelectionKey k = it.next();
            it.remove();
            if (k.isReadable()) {
                ((Handler) k.attachment()).ready(k);
            }
        }
    }

    /** Walks every registered key: reads attachments but handles no readiness. */
    public static void dumpKeys(Selector sel) {
        for (SelectionKey k : sel.keys()) {
            Main.ran("key " + k.attachment().getClass().getSimpleName());
        }
    }
}
