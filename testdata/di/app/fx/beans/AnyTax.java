package fx.beans;

import fx.trace.Ran;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

/** FRAMEWORKS F5 f5-3: two names — created when either is active. */
@Component
@Profile({"us", "cloud"})
public class AnyTax {
    public AnyTax() { Ran.stack(); }
}
