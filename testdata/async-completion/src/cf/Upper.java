package cf;

import java.util.function.Function;

public class Upper implements Function<Object, Object> {
    @Override
    public Object apply(Object o) {
        Work.hit("Upper");
        return o;
    }
}
