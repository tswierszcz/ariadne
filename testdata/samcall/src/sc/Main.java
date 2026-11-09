package sc;

import java.util.List;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;

/** An application functional interface. */
interface Op {
    int apply(int x);
}

/** A non-final private field set by a setter and run by another method. */
final class Holder {
    private Runnable action;

    void set(Runnable action) {
        this.action = action;
    }

    void fire() {
        action.run();
    }
}

/** A final field set in the constructor. */
final class Task {
    private final Function<String, Integer> f;

    Task(Function<String, Integer> f) {
        this.f = f;
    }

    int go() {
        return f.apply("abc");
    }
}

class Base {
    int measure(String s) {
        Oracle.ran();
        return 1;
    }
}

/** Overrides the method the bound reference names: the JVM runs this one. */
final class Derived extends Base {
    @Override
    int measure(String s) {
        Oracle.ran();
        return s.length();
    }
}

final class Thing {
    Thing() {
        Oracle.ran();
    }
}

/** Calls through an interface: the helper's callers name Exec#exec. */
interface Exec {
    void exec(Runnable r);
}

final class DirectExec implements Exec {
    @Override
    public void exec(Runnable r) {
        r.run();
    }
}

public final class Main {
    public static void main(String[] args) {
        caseParam();
        caseReturn();
        caseField();
        caseFinalField();
        caseConstructorRef();
        caseEscape();
        caseCaptured();
        caseInterfaceHelper();
        unused();
    }

    /** sam-call through a parameter: the lambda is created here, run in run(). */
    static void caseParam() {
        run(x -> {
            Oracle.ran();
            return x + 1;
        }, 1);
    }

    static int run(Op op, int x) {
        return op.apply(x);
    }

    /** sam-call through a return value. */
    static void caseReturn() {
        Supplier<String> s = make();
        s.get();
    }

    static Supplier<String> make() {
        return () -> {
            Oracle.ran();
            return "s";
        };
    }

    /** A static method reference stored in a field by a setter. */
    static void caseField() {
        Holder h = new Holder();
        h.set(Main::target);
        h.fire();
    }

    static void target() {
        Oracle.ran();
    }

    /** A bound instance-method reference to an overridden method, in a final field. */
    static void caseFinalField() {
        Base b = new Derived();
        new Task(b::measure).go();
    }

    /** A constructor reference run by a helper. */
    static void caseConstructorRef() {
        build(Thing::new);
    }

    static Object build(Supplier<?> s) {
        return s.get();
    }

    /** escape: handed to List#forEach (JDK code runs it) from a helper. */
    static void caseEscape() {
        each(List.of("a"), v -> Oracle.ran());
    }

    static void each(List<String> l, Consumer<String> c) {
        l.forEach(c);
    }

    /** A lambda body runs a lambda it captured from its creator. */
    static void caseCaptured() {
        Op inner = x -> {
            Oracle.ran();
            return x * 2;
        };
        Runnable r = () -> inner.apply(3);
        r.run();
    }

    /** The helper is called through its interface. */
    static void caseInterfaceHelper() {
        Exec e = new DirectExec();
        e.exec(() -> Oracle.ran());
    }

    /** Creates a lambda nobody runs. */
    static Op unused() {
        return x -> {
            Oracle.ran();
            return 0;
        };
    }

    /** Never called: its lambda must not reach run()'s SAM call. */
    static void neverCalled() {
        run(x -> {
            Oracle.ran();
            return -1;
        }, 2);
    }
}
