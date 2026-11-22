package fx.beans;

import fx.trace.Ran;
import org.springframework.stereotype.Component;

@Component
public class ColdStore implements Store {
    public ColdStore() { Ran.stack(); }

    @Override
    public String save() { Ran.stack(); return "ColdStore"; }
}
