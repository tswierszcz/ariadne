package fx.svc;

import fx.trace.Ran;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** @Transactional on one method only. */
@Service
public class OrderService {
    @Transactional
    public int place() { Ran.stack(); return 1; }

    public int count() { Ran.stack(); return 0; }
}
