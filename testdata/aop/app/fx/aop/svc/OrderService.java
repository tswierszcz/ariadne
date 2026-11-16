package fx.aop.svc;

import fx.trace.Ran;
import org.springframework.stereotype.Service;

/** No interface: Spring proxies it with CGLIB. */
@Service
public class OrderService {
    /** Calls internal() on this: the proxy is bypassed, no advice runs for it. */
    public void place(String who) {
        Ran.stack();
        internal(who);
    }

    public void internal(String who) { Ran.stack(); }

    @Audited("cancel")
    public void cancel(int id) { Ran.stack(); }

    /** Final: a CGLIB proxy cannot override it, no advice runs. */
    public final void fixed() { Ran.stack(); }
}
