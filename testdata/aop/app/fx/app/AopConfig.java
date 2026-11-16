package fx.app;

import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.EnableAspectJAutoProxy;

/** JDK proxies for beans with interfaces (the default), CGLIB otherwise. */
@Configuration
@ComponentScan("fx.aop")
@EnableAspectJAutoProxy
public class AopConfig {
}
