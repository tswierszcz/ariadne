package fx.beans;

import fx.trace.Ran;
import org.springframework.stereotype.Component;

@Component("slow")
public class SlowStore implements Store {
    public SlowStore() { Ran.stack(); }

    @Override
    public String save() { Ran.stack(); return "SlowStore"; }
}
