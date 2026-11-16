package fx.aop.svc;

import fx.trace.Ran;
import org.springframework.stereotype.Component;

@Component
@Monitored
public class ReportService {
    public void render() { Ran.stack(); }
}
