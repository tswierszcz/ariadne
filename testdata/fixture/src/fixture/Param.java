package fixture;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Stands in for a real parameter-level framework annotation (e.g. CDI's
 * {@code @Observes}), so parser tests for
 * RuntimeVisibleParameterAnnotations do not need a real CDI dependency on
 * the classpath. Unlike {@link Entry}, this is PARAMETER-targeted, not
 * METHOD-targeted.
 */
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.PARAMETER)
public @interface Param {
}
