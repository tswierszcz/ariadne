package fx.other;

import fx.trace.Ran;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

/** Outside the component scan: never a bean, never a handler; /ghost is a 404. */
@RestController
public class GhostController {
    public GhostController() { Ran.stack(); }

    @GetMapping("/ghost")
    public String ghost() { Ran.stack(); return "ghost"; }
}
