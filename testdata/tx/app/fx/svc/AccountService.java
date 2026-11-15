package fx.svc;

import fx.trace.Ran;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** @Transactional on the class: every public method runs in a transaction (a CGLIB proxy). */
@Service
@Transactional
public class AccountService {
    public String find(String id) { Ran.stack(); return "account " + id; }

    public void fail() { Ran.stack(); throw new IllegalStateException("rolled back"); }
}
