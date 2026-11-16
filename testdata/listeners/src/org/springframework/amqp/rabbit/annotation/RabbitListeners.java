package org.springframework.amqp.rabbit.annotation;

import java.lang.annotation.*;

@Retention(RetentionPolicy.RUNTIME)
@Target({ElementType.METHOD, ElementType.TYPE})
public @interface RabbitListeners {
    RabbitListener[] value();
}
