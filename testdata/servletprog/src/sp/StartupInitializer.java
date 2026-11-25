package sp;

import java.util.Set;
import javax.servlet.ServletContainerInitializer;
import javax.servlet.ServletContext;
import javax.servlet.ServletException;
import javax.servlet.annotation.HandlesTypes;

/**
 * Spring's SpringServletContainerInitializer in miniature: creates every
 * concrete class the container hands it (reflectively) and runs its
 * onStartup.
 */
@HandlesTypes(Startup.class)
public final class StartupInitializer implements ServletContainerInitializer {
    @Override
    public void onStartup(Set<Class<?>> classes, ServletContext ctx) throws ServletException {
        for (Class<?> c : classes) {
            if (c.isInterface() || java.lang.reflect.Modifier.isAbstract(c.getModifiers())) {
                continue;
            }
            try {
                ((Startup) c.getDeclaredConstructor().newInstance()).onStartup(ctx);
            } catch (ReflectiveOperationException e) {
                throw new ServletException(e);
            }
        }
    }
}
