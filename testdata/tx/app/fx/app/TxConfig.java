package fx.app;

import fx.svc.FxTransactionManager;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;
import org.springframework.transaction.PlatformTransactionManager;
import org.springframework.transaction.annotation.EnableTransactionManagement;

@Configuration
@ComponentScan("fx.svc")
@EnableTransactionManagement
public class TxConfig {
    @Bean
    public PlatformTransactionManager transactionManager() { return new FxTransactionManager(); }
}
