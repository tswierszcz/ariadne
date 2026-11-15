package fx.beans;

import fx.trace.Ran;
import org.springframework.beans.factory.config.BeanPostProcessor;
import org.springframework.stereotype.Component;

/** A BeanPostProcessor: the container calls it around every bean's initialization. */
@Component
public class Tracer implements BeanPostProcessor {
    public Object postProcessBeforeInitialization(Object bean, String name) { Ran.stack(); return bean; }
}
