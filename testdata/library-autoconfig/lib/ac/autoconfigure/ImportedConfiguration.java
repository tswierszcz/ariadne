package ac.autoconfigure;

import javax.annotation.PostConstruct;
import org.apache.commons.text.StringSubstitutor;
import org.springframework.context.annotation.Configuration;

/** (l) imported by GreeterAutoConfiguration; its @PostConstruct → replace. */
@Configuration(proxyBeanMethods = false)
class ImportedConfiguration {
    @PostConstruct
    void init() {
        System.out.println(StringSubstitutor.createInterpolator().replace("(l) ${java:version}"));
    }
}
