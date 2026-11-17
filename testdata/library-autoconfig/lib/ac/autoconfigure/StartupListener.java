package ac.autoconfigure;

import org.apache.commons.text.StringSubstitutor;
import org.springframework.boot.context.event.ApplicationStartedEvent;
import org.springframework.context.ApplicationListener;

/** (m) announced in META-INF/spring.factories (ApplicationListener). */
class StartupListener implements ApplicationListener<ApplicationStartedEvent> {
    @Override
    public void onApplicationEvent(ApplicationStartedEvent event) {
        System.out.println(StringSubstitutor.createInterpolator().replace("(m) ${java:version}"));
    }
}
