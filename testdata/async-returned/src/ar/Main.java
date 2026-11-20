package ar;

import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class Main {
    public static void main(String[] args) throws Exception {
        Conn c = new Conn();
        ExecStrategy s = new Strategy(new SelProducer(new Endpoint[] {c, c}));
        Thread t = new Thread(s::loop);
        t.start();
        t.join();
        c.closer(); // returned, never run
        ExecutorService ex = Executors.newSingleThreadExecutor();
        System.out.println(ex.submit(new Factory().job()).get());
        ex.shutdown();
    }
}
