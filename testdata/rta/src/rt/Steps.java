package rt;

/**
 * A dispatch cycle with allocation discovered late: StepB is created only in StepA#go, which is
 * reached only through an interface call on a StepA; StepC is created only in StepB#go.
 */
interface Step {
    void go(int n);
}

final class StepA implements Step {
    public void go(int n) {
        Ran.here();
        if (n > 0) {
            Step next = new StepB(this);
            next.go(n - 1);
        }
    }
}

final class StepB implements Step {
    private final Step back;

    StepB(Step back) {
        this.back = back;
    }

    public void go(int n) {
        Ran.here();
        Step last = new StepC();
        last.go(0);
        if (n > 0) {
            back.go(n - 1);
        }
    }
}

final class StepC implements Step {
    public void go(int n) {
        Ran.here();
    }
}
