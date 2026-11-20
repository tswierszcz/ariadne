package ah;

import java.nio.ByteBuffer;
import java.nio.channels.Pipe;
import java.nio.channels.Selector;

/** JVM oracle driver (scripts/async-handoff-oracle.sh): the drains run on another thread. */
public class Main {
    static synchronized void ran(String what) {
        System.out.println("ran " + what);
    }

    public static void main(String[] args) throws Exception {
        Loop loop = new Loop();
        Swap swap = new Swap();
        Listeners listeners = new Listeners();
        Api.go(loop, swap, listeners);
        new Unused();

        Selector sel = Selector.open();
        Pipe pipe = Pipe.open();
        Sel.register(sel, pipe.source(), new Handler());
        pipe.sink().write(ByteBuffer.wrap(new byte[] {1}));

        Thread worker = new Thread(() -> {
            loop.drain();
            swap.process();
            try {
                Sel.loop(sel);
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
            Sel.dumpKeys(sel);
        });
        worker.start();
        worker.join();
        sel.close();
    }
}
