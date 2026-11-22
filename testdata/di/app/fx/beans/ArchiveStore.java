package fx.beans;

import fx.trace.Ran;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

@Component
@Qualifier("archive")
public class ArchiveStore implements Store {
    public ArchiveStore() { Ran.stack(); }

    @Override
    public String save() { Ran.stack(); return "ArchiveStore"; }
}
