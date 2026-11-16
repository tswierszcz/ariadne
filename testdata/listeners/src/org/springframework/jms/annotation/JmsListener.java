package org.springframework.jms.annotation;

import java.lang.annotation.*;

/** A name-only stand-in for the real annotation (the spring profile matches roots by name). */
@Retention(RetentionPolicy.RUNTIME)
@Target({ElementType.METHOD, ElementType.ANNOTATION_TYPE})
@java.lang.annotation.Repeatable(JmsListeners.class)
public @interface JmsListener {
    String[] value() default {};
}
