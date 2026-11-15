package fx.app;

import fx.ev.AuditListener;
import fx.ev.FxTransactionManager;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.annotation.EnableScheduling;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.annotation.EnableTransactionManagement;

@Configuration
@ComponentScan("fx.ev")
@EnableScheduling
@EnableTransactionManagement
public class EventsConfig {
    @Bean
    public PlatformTransactionManager transactionManager() { return new FxTransactionManager(); }

    /** A listener registered by a @Bean method, not by a stereotype. */
    @Bean
    public AuditListener audit() { return new AuditListener(); }
}
