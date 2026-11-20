package cf;

import java.util.function.Function;

public class Recover implements Function<Throwable, Object> {
    @Override
    public Object apply(Throwable o) {
        Work.hit("Recover");
        return o;
    }
}
