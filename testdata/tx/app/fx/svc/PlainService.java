package fx.svc;

import fx.trace.Ran;
import org.springframework.stereotype.Service;

/** Negative: no @Transactional — no proxy, no transaction. */
@Service
public class PlainService {
    public String hello() { Ran.stack(); return "hello"; }
}
