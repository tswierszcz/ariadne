package org.springframework.web.bind.annotation;

import java.lang.annotation.*;

/** Stub of the Spring MVC annotation (same name, for the demo only). */
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.METHOD)
public @interface PostMapping {
    String[] value() default {};
}
