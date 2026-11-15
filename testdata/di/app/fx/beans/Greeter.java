package fx.beans;

import fx.trace.Ran;
import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

/**
 * Constructor injection, field injection, setter injection and the
 * JSR-250 lifecycle annotations.
 */
@Service
public class Greeter {
    private final Repo repo;
    @Autowired
    private Widget widget;
    private Engine engine;

    public Greeter(Repo repo) { Ran.stack(); this.repo = repo; }

    @Autowired
    public void setEngine(Engine engine) { Ran.stack(); this.engine = engine; }

    @PostConstruct
    void init() { Ran.stack(); }

    @PreDestroy
    void shutdown() { Ran.stack(); }

    public String greet() { return "hello " + repo.name() + " " + widget.label; }
}
