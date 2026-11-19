package fx;

import cd.api.Banner;
import cd.api.Component;
import cd.api.Startup;
import cd.api.Store;
import org.jboss.weld.environment.se.Weld;
import org.jboss.weld.environment.se.WeldContainer;

/** The oracle's client: boots Weld SE with the library's bean archive and uses its beans. */
public class Main {
    public static void main(String[] args) {
        try (WeldContainer c = new Weld().initialize()) {
            System.out.println(c.select(Banner.class).get().text().isEmpty() ? "" : "banner");   // (t)
            c.event().select(Startup.class).fire(new Startup());                               // (u)
            System.out.println(c.select(Store.class).get().get().isEmpty() ? "" : "store");   // (v)
            System.out.println(c.select(Component.class).get().name().isEmpty() ? "" : "component"); // (w)
        }
    }
}
