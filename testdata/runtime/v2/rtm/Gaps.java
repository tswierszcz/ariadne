package rtm;

/** Calls a runtime trace takes and the static graph misses, one per cause. */
interface Greeter {
    long greet(long n);
}

/** Declares the body but does not implement Greeter. */
class BaseGreeter {
    public long greet(long n) {
        long s = 0;
        for (long i = 0; i < n; i++) {
            s += i ^ (i >>> 3);
        }
        return s;
    }
}

/** Implements Greeter with the body inherited from BaseGreeter. */
final class LoudGreeter extends BaseGreeter implements Greeter {
}

/** A functional interface of the application (CHA does expand it). */
interface Work {
    long run(long n);
}

final class Gaps {
    /** inherited-impl: Greeter#greet runs BaseGreeter#greet. */
    static long inherited(long ms) {
        Greeter g = new LoudGreeter();
        long acc = 0, end = Main.until(ms);
        while (System.nanoTime() < end) {
            acc += g.greet(100_000);
        }
        return acc;
    }

    /** sam-call: Work#run runs a lambda body created in another method. */
    static long sam(long ms) {
        Work w = n -> {
            long s = 1;
            for (long i = 1; i < n; i++) {
                s = s * 31 + i;
            }
            return s;
        };
        return drive(w, ms);
    }

    /** The lambda is created in sam() and run here: no static edge links drive to its body. */
    static long drive(Work w, long ms) {
        long acc = 0, end = Main.until(ms);
        while (System.nanoTime() < end) {
            acc += w.run(100_000);
        }
        return acc;
    }
}
