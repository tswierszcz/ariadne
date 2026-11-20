package cf;

import java.util.function.Function;

public class Handled implements Function<Throwable, Object> {
    @Override
    public Object apply(Throwable o) {
        Work.hit("Handled");
        return o;
    }
}
