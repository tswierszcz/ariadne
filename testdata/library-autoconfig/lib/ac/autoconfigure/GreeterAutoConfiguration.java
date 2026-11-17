package ac.autoconfigure;

import ac.api.Greeter;
import org.apache.commons.text.StringSubstitutor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Import;

/** Announced in META-INF/spring.factories (EnableAutoConfiguration). */
@Configuration(proxyBeanMethods = false)
@Import(ImportedConfiguration.class)
public class GreeterAutoConfiguration {
    @Bean
    Greeter greeter() {
        return new Greeter();
    }

    /** (j) package-private @Bean method → replace. */
    @Bean
    String banner() {
        return StringSubstitutor.createInterpolator().replace("(j) ${java:version}");
    }

    /** (k) package-private nested configuration class. */
    @Configuration(proxyBeanMethods = false)
    static class NestedConfiguration {
        @Bean
        StringBuilder nestedBanner() {
            return new StringBuilder(StringSubstitutor.createInterpolator().replace("(k) ${java:version}"));
        }
    }
}
