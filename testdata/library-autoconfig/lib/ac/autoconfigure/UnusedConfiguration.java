package ac.autoconfigure;

import org.apache.commons.text.StringSubstitutor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/** (n) never announced or imported: Spring never creates it. */
@Configuration(proxyBeanMethods = false)
class UnusedConfiguration {
    @Bean
    String unusedBanner() {
        return StringSubstitutor.createInterpolator().replace("(n) ${java:version}");
    }
}
