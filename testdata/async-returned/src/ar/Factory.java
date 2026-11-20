package ar;

import java.util.concurrent.Callable;

/** Returns a Callable that an executor runs (a registration site). */
public class Factory {
    public Callable<String> job() {
        return () -> Work.job();
    }
}
