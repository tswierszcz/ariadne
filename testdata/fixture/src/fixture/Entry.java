package fixture;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Stands in for a real framework route annotation (e.g. Spring's
 * {@code @RequestMapping}), so profile/root detection tests do not need a
 * real framework dependency on the classpath.
 */
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
public @interface Entry {
}
