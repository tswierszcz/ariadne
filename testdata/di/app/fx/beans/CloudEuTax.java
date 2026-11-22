package fx.beans;

import fx.trace.Ran;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

/** FRAMEWORKS F5 f5-3: created only when the profile expression "eu & cloud" holds. */
@Component
@Profile("eu & cloud")
public class CloudEuTax {
    public CloudEuTax() { Ran.stack(); }
}
