package fx;

import io.micronaut.context.ApplicationContext;
import mn.api.Banner;
import mn.api.Component;
import mn.api.Startup;
import mn.api.Store;

/** A client: boots Micronaut and uses each of the library's beans once. */
public class Main {
    public static void main(String[] args) {
        try (ApplicationContext ctx = ApplicationContext.run()) {
            System.out.println(ctx.getBean(Banner.class).text());
            ctx.getEventPublisher(Startup.class).publishEvent(new Startup());
            System.out.println(ctx.getBean(Store.class).get("k"));
            System.out.println(ctx.getBean(Component.class).describe());
        }
    }
}
