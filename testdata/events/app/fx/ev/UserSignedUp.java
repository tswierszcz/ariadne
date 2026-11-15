package fx.ev;

/** A payload event (not an ApplicationEvent): Spring wraps it in PayloadApplicationEvent. */
public class UserSignedUp implements DomainEvent {
    public final String name;
    public UserSignedUp(String name) { this.name = name; }
}
