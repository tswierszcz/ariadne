package ar;

/** An endpoint nothing creates: its task is a candidate, gated out. */
public class Idle implements Endpoint {
    @Override
    public Runnable onSelected() {
        return new IdleTask();
    }
}
