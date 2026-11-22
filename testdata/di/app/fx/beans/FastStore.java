package fx.beans;

import fx.trace.Ran;
import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Component;

@Component
@Primary
public class FastStore implements Store {
    public FastStore() { Ran.stack(); }

    @Override
    public String save() { Ran.stack(); return "FastStore"; }
}
