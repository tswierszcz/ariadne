package fx.app;

import fx.trace.Ran;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import org.apache.catalina.Context;
import org.apache.catalina.startup.Tomcat;
import org.springframework.web.context.support.AnnotationConfigWebApplicationContext;
import org.springframework.web.servlet.DispatcherServlet;

/**
 * FRAMEWORKS F5 f5-2 oracle: a real embedded Tomcat 9 runs a Spring MVC
 * 5.3.6 DispatcherServlet, and real HTTP requests reach the controllers.
 * Every application body records its stack (fx.trace.Ran); main prints
 * every body that ran.
 */
public final class Main {
    public static void main(String[] args) throws Exception {
        Path base = Files.createTempDirectory("mvc-oracle");
        AnnotationConfigWebApplicationContext ctx = new AnnotationConfigWebApplicationContext();
        ctx.register(WebConfig.class);
        Tomcat tomcat = new Tomcat();
        tomcat.setBaseDir(base.toString());
        tomcat.setPort(0);
        Context context = tomcat.addContext("", base.toString());
        Tomcat.addServlet(context, "dispatcher", new DispatcherServlet(ctx));
        context.addServletMappingDecoded("/", "dispatcher");
        tomcat.getConnector();
        tomcat.start();
        int port = tomcat.getConnector().getLocalPort();
        try {
            System.out.println("# GET /hello -> " + call(port, "GET", "/hello", null));
            System.out.println("# GET /item/7 -> " + call(port, "GET", "/item/7", null));
            System.out.println("# POST /item -> " + call(port, "POST", "/item", "{\"name\":\"lamp\"}"));
            System.out.println("# GET /boom -> " + call(port, "GET", "/boom", null));
            System.out.println("# GET /ghost -> " + call(port, "GET", "/ghost", null));
        } finally {
            tomcat.stop();
            tomcat.destroy();
        }
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the application bodies' stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }

    private static String call(int port, String method, String path, String json) throws Exception {
        HttpURLConnection c = (HttpURLConnection) new URL("http://127.0.0.1:" + port + path).openConnection();
        c.setRequestMethod(method);
        if (json != null) {
            c.setDoOutput(true);
            c.setRequestProperty("Content-Type", "application/json");
            try (OutputStream o = c.getOutputStream()) {
                o.write(json.getBytes(StandardCharsets.UTF_8));
            }
        }
        int code = c.getResponseCode();
        InputStream in = code < 400 ? c.getInputStream() : c.getErrorStream();
        String body = in == null ? "" : new String(in.readAllBytes(), StandardCharsets.UTF_8);
        return code + " " + body;
    }
}
