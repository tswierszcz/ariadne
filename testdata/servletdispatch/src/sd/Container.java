package sd;

import java.lang.reflect.Proxy;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import javax.servlet.Filter;
import javax.servlet.FilterChain;
import javax.servlet.Servlet;
import javax.servlet.ServletRequest;
import javax.servlet.ServletResponse;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import javax.xml.parsers.DocumentBuilderFactory;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.w3c.dom.NodeList;

/**
 * A minimal servlet container: it reads WEB-INF/web.xml, creates the declared
 * servlets and filters reflectively from their class names (as Tomcat's
 * StandardWrapper/ApplicationFilterConfig do), and dispatches a request
 * through javax.servlet.Filter#doFilter and javax.servlet.Servlet#service —
 * the same interface calls as Tomcat's ApplicationFilterChain#internalDoFilter.
 * No class name of a servlet or filter appears in this bytecode.
 */
final class Container {
    private final Map<String, Servlet> servlets = new HashMap<>();
    private final List<Filter> filters = new ArrayList<>();

    static Container fromWebXml(Path webXml) throws Exception {
        Document d = DocumentBuilderFactory.newInstance().newDocumentBuilder().parse(webXml.toFile());
        Container c = new Container();
        Map<String, Servlet> byName = new HashMap<>();
        NodeList ss = d.getElementsByTagName("servlet");
        for (int i = 0; i < ss.getLength(); i++) {
            Element e = (Element) ss.item(i);
            byName.put(text(e, "servlet-name"), (Servlet) create(text(e, "servlet-class")));
        }
        NodeList ms = d.getElementsByTagName("servlet-mapping");
        for (int i = 0; i < ms.getLength(); i++) {
            Element e = (Element) ms.item(i);
            c.servlets.put(text(e, "url-pattern"), byName.get(text(e, "servlet-name")));
        }
        NodeList fs = d.getElementsByTagName("filter");
        for (int i = 0; i < fs.getLength(); i++) {
            c.filters.add((Filter) create(text((Element) fs.item(i), "filter-class")));
        }
        return c;
    }

    private static Object create(String className) throws Exception {
        return Class.forName(className).getDeclaredConstructor().newInstance();
    }

    private static String text(Element e, String tag) {
        return e.getElementsByTagName(tag).item(0).getTextContent().trim();
    }

    void request(String method, String path) throws Exception {
        HttpServletRequest req = (HttpServletRequest) Proxy.newProxyInstance(getClass().getClassLoader(),
            new Class<?>[] {HttpServletRequest.class}, (p, m, a) -> m.getName().equals("getMethod") ? method : defaultValue(m.getReturnType()));
        HttpServletResponse res = (HttpServletResponse) Proxy.newProxyInstance(getClass().getClassLoader(),
            new Class<?>[] {HttpServletResponse.class}, (p, m, a) -> defaultValue(m.getReturnType()));
        new Chain(filters, servlets.get(path)).doFilter(req, res);
    }

    private static Object defaultValue(Class<?> t) {
        if (t == boolean.class) return false;
        if (t == int.class) return 0;
        if (t == long.class) return 0L;
        return null;
    }

    /** ApplicationFilterChain's shape: each filter, then the servlet. */
    static final class Chain implements FilterChain {
        private final List<Filter> filters;
        private final Servlet servlet;
        private int pos;

        Chain(List<Filter> filters, Servlet servlet) {
            this.filters = filters;
            this.servlet = servlet;
        }

        @Override
        public void doFilter(ServletRequest req, ServletResponse res) throws java.io.IOException, javax.servlet.ServletException {
            if (pos < filters.size()) {
                filters.get(pos++).doFilter(req, res, this);
                return;
            }
            servlet.service(req, res);
        }
    }
}
