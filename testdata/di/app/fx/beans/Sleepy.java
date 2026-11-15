package fx.beans;

import fx.trace.Ran;
import org.springframework.context.annotation.Lazy;
import org.springframework.stereotype.Component;

/** Negative: a lazy bean nothing asks for — never created. */
@Component
@Lazy
public class Sleepy {
    public Sleepy() { Ran.stack(); }
}
