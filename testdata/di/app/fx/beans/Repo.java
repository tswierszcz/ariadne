package fx.beans;

import fx.trace.Ran;
import org.springframework.stereotype.Repository;

/** A stereotype bean created through its no-argument constructor. */
@Repository
public class Repo {
    public Repo() { Ran.stack(); }
    String name() { return "Leo"; }
}
