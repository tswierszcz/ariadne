package fx.app;

import fx.beans.Engine;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;

@Configuration
@ComponentScan("fx.beans")
public class AppConfig {
    /** A @Bean with init and destroy methods named in the annotation. */
    @Bean(initMethod = "start", destroyMethod = "stop")
    public Engine engine() {
        return new Engine();
    }
}
