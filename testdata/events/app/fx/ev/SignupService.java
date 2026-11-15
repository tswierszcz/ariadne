package fx.ev;

import org.springframework.context.ApplicationEventPublisher;
import org.springframework.stereotype.Service;

@Service
public class SignupService {
    private final ApplicationEventPublisher events;

    public SignupService(ApplicationEventPublisher events) { this.events = events; }

    public void signup(String name) { events.publishEvent(new UserSignedUp(name)); }
}
