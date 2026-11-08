package rtm;

/** Loaded by two loaders in the same run. */
public final class Plugin {
    public long work(long n) {
        long s = 0;
        for (long i = 0; i < n; i++) {
            s ^= i * 2654435761L;
        }
        return s;
    }
}
