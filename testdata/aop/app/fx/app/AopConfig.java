package fx.app;

import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.EnableAspectJAutoProxy;
import org.springframework.context.annotation.ImportResource;

/** JDK proxies for beans with interfaces (the default), CGLIB otherwise; fx/aop/aop.xml adds an <aop:config> aspect. */
@Configuration
@ComponentScan("fx.aop")
@EnableAspectJAutoProxy
@ImportResource("classpath:fx/aop/aop.xml")
public class AopConfig {
}
