package n4;

import java.lang.reflect.*;
import java.util.concurrent.*;

public class Main {
    public interface Svc { int op(int x); }

    public static int target(int x) { return x + 1; }           // reflection target
    public static class Real implements Svc { public int op(int x) { return x * 2; } }

    static final class Task implements Runnable {               // one task class, two registration sites
        final int v; Task(int v) { this.v = v; }
        public void run() { work(v); }
    }
    static void work(int v) { }                                  // async target

    static Svc proxy(String tag) {
        Svc real = new Real();
        return (Svc) Proxy.newProxyInstance(Main.class.getClassLoader(), new Class<?>[] {Svc.class},
            (p, m, a) -> m.invoke(real, a));
    }

    static void submitA(ExecutorService ex) throws Exception { ex.submit(new Task(1)).get(); }
    static void submitB(ExecutorService ex) throws Exception { ex.submit(new Task(2)).get(); }

    public static void main(String[] a) throws Exception {
        Method m = Main.class.getMethod("target", int.class);
        for (int i = 0; i < 3; i++) m.invoke(null, i);
        proxy("p1").op(1);
        proxy("p2").op(2);
        ExecutorService ex = Executors.newSingleThreadExecutor();
        submitA(ex);
        submitB(ex);
        ex.shutdown();
    }
}
