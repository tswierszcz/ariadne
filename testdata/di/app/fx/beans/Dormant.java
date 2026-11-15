package fx.beans;

import fx.trace.Ran;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

/** Negative: a bean for a profile that is never active — never created. */
@Component
@Profile("never")
public class Dormant {
    public Dormant() { Ran.stack(); }
}
