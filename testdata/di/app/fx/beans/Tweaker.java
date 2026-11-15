package fx.beans;

import fx.trace.Ran;
import org.springframework.beans.factory.config.BeanFactoryPostProcessor;
import org.springframework.beans.factory.config.ConfigurableListableBeanFactory;
import org.springframework.stereotype.Component;

/** A BeanFactoryPostProcessor: the container calls it before creating beans. */
@Component
public class Tweaker implements BeanFactoryPostProcessor {
    public void postProcessBeanFactory(ConfigurableListableBeanFactory factory) { Ran.stack(); }
}
