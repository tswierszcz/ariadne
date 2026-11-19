package fx;

import java.io.StringReader;
import org.apache.commons.configuration2.PropertiesConfiguration;
import org.apache.commons.configuration2.io.FileHandler;

/**
 * LIBRARY_MODE LB4 oracle for commons-configuration2 2.7: a user loads a
 * properties file with variable interpolation and reads two keys — the
 * library's documented use. Run by scripts/library-lb4-jvm-oracle.sh.
 */
public class ConfigClient {
    public static void main(String[] args) throws Exception {
        PropertiesConfiguration config = new PropertiesConfiguration();
        new FileHandler(config).load(new StringReader(
                "home = ${sys:user.home}\nmax = ${const:java.lang.Integer.MAX_VALUE}\n"));
        System.out.println("home=" + (config.getString("home") != null) + " max=" + config.getString("max"));
    }
}
