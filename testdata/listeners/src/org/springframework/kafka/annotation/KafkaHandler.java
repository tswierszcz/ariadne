package org.springframework.kafka.annotation;

import java.lang.annotation.*;

/** A name-only stand-in for the real annotation (the spring profile matches roots by name). */
@Retention(RetentionPolicy.RUNTIME)
@Target({ElementType.METHOD})

public @interface KafkaHandler {
    String[] value() default {};
    boolean isDefault() default false;
}
