package cf;

import java.util.function.Function;

public class Skipped implements Function<Object, Object> {
    @Override
    public Object apply(Object o) {
        Work.hit("Skipped");
        return o;
    }
}
