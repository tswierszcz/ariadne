package fx.a;

import fx.trace.Ran;
import java.util.concurrent.Callable;

public class Job implements Callable<String> {
    public String call() { Ran.here(); return "job"; }
}
