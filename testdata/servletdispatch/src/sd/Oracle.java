package sd;

import java.nio.file.Path;

/** Starts the container on web.xml, sends three requests, and prints the method bodies that ran. */
public final class Oracle {
    public static void main(String[] args) throws Exception {
        Container c = Container.fromWebXml(Path.of(args.length > 0 ? args[0] : "WEB-INF/web.xml"));
        c.request("GET", "/hello");
        c.request("POST", "/hello");
        c.request("GET", "/plain");
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc)");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
