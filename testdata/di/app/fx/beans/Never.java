package fx.beans;

import fx.trace.Ran;
import org.springframework.context.annotation.Condition;
import org.springframework.context.annotation.ConditionContext;
import org.springframework.core.type.AnnotatedTypeMetadata;

/** A condition that never matches; the container does run matches. */
public class Never implements Condition {
    public boolean matches(ConditionContext ctx, AnnotatedTypeMetadata md) { Ran.stack(); return false; }
}
