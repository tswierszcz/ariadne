package org.springframework.amqp.rabbit.annotation;

import java.lang.annotation.*;

/** A name-only stand-in for the real annotation (the spring profile matches roots by name). */
@Retention(RetentionPolicy.RUNTIME)
@Target({ElementType.METHOD})

public @interface RabbitHandler {
    String[] value() default {};
}
