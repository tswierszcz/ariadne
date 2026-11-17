package ac.autoconfigure;

import org.apache.commons.text.StringSubstitutor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/** (o) announced only in AutoConfiguration.imports (Spring Boot 2.7+). */
@Configuration(proxyBeanMethods = false)
class ModernAutoConfiguration {
    @Bean
    String modernBanner() {
        return StringSubstitutor.createInterpolator().replace("(o) ${java:version}");
    }
}
