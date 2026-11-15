package fx.beans;

import fx.trace.Ran;
import org.springframework.context.annotation.Conditional;
import org.springframework.stereotype.Component;

/** Negative: a bean whose condition fails — never created. */
@Component
@Conditional(Never.class)
public class Unmet {
    public Unmet() { Ran.stack(); }
}
