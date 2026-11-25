package sp;

import java.util.List;

/**
 * Starts the container over the fixture's class names (what a container's
 * class scan finds), sends two requests, and prints the method bodies that
 * ran.
 */
public final class Oracle {
    public static void main(String[] args) throws Exception {
        List<String> scanned = List.of("sp.AppStartup", "sp.BaseStartup", "sp.UnusedStartup", "sp.NotAStartup", "sp.MainServlet");
        Container c = Container.start(scanned);
        c.request("GET", "main");
        c.request("GET", "ping");
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
