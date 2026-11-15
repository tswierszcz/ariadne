package fx.svc;

import fx.trace.Ran;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

@Service
@Transactional(readOnly = true)
public class CatalogImpl implements Catalog {
    public String title() { Ran.stack(); return "pets"; }
}
