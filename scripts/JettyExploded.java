import java.net.InetSocketAddress;
import org.eclipse.jetty.annotations.AnnotationConfiguration;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.webapp.*;

/**
 * Runs an exploded web application directory in Jetty 9.4 on 127.0.0.1 with
 * the configurations io.lavagna.config.Launcher uses (annotation scanning, so
 * Spring's ServletContainerInitializer starts the application). Jetty comes
 * from the class path.
 *
 * Usage: java -cp JETTY_JARS scripts/JettyExploded.java WEBAPP_DIR PORT
 */
public class JettyExploded {
    public static void main(String[] a) throws Exception {
        Server server = new Server(new InetSocketAddress("127.0.0.1", Integer.parseInt(a[1])));
        WebAppContext w = new WebAppContext();
        w.setContextPath("/");
        w.setServer(server);
        w.setWar(a[0]);
        w.setConfigurations(new Configuration[] { new WebInfConfiguration(), new WebXmlConfiguration(),
            new MetaInfConfiguration(), new AnnotationConfiguration(), new JettyWebXmlConfiguration() });
        server.setHandler(w);
        server.start();
        server.join();
    }
}
