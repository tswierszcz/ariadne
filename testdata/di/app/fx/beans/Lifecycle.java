package fx.beans;

import fx.trace.Ran;
import org.springframework.beans.factory.BeanNameAware;
import org.springframework.beans.factory.DisposableBean;
import org.springframework.beans.factory.InitializingBean;
import org.springframework.beans.factory.SmartInitializingSingleton;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.stereotype.Component;

/** The callback interfaces the container calls on a bean. */
@Component
public class Lifecycle implements BeanNameAware, ApplicationContextAware, InitializingBean, DisposableBean, SmartInitializingSingleton {
    public void setBeanName(String name) { Ran.stack(); }
    public void setApplicationContext(ApplicationContext ctx) { Ran.stack(); }
    public void afterPropertiesSet() { Ran.stack(); }
    public void afterSingletonsInstantiated() { Ran.stack(); }
    public void destroy() { Ran.stack(); }
}
