package sp;

import java.lang.reflect.Proxy;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.ServiceLoader;
import java.util.Set;
import javax.servlet.Filter;
import javax.servlet.FilterChain;
import javax.servlet.Servlet;
import javax.servlet.ServletContainerInitializer;
import javax.servlet.ServletContext;
import javax.servlet.ServletRequest;
import javax.servlet.ServletResponse;
import javax.servlet.annotation.HandlesTypes;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

/**
 * A minimal servlet container with no web.xml: it loads every
 * ServletContainerInitializer through ServiceLoader (as Jetty's and
 * Tomcat's annotation configuration do), hands each the classes of the
 * scanned class names that are subtypes of its @HandlesTypes types, and
 * records what the initializers register on the ServletContext
 * (addServlet/addFilter with an instance, a class or a class name). It then
 * dispatches requests through Filter#doFilter and Servlet#service. No class
 * name of a servlet, filter or initializer appears in this bytecode.
 */
final class Container {
    private final Map<String, Servlet> servlets = new HashMap<>();
    private final List<Filter> filters = new ArrayList<>();

    static Container start(List<String> scanned) throws Exception {
        Container c = new Container();
        ServletContext ctx = (ServletContext) Proxy.newProxyInstance(Container.class.getClassLoader(),
            new Class<?>[] {ServletContext.class}, (p, m, a) -> {
                switch (m.getName()) {
                    case "addServlet":
                        c.servlets.put((String) a[0], (Servlet) instance(a[1]));
                        return null;
                    case "addFilter":
                        c.filters.add((Filter) instance(a[1]));
                        return null;
                    default:
                        return null;
                }
            });
        for (ServletContainerInitializer sci : ServiceLoader.load(ServletContainerInitializer.class)) {
            Set<Class<?>> handled = new LinkedHashSet<>();
            HandlesTypes ht = sci.getClass().getAnnotation(HandlesTypes.class);
            for (String n : scanned) {
                Class<?> k = Class.forName(n);
                for (Class<?> t : ht == null ? new Class<?>[0] : ht.value()) {
                    if (t.isAssignableFrom(k) && k != t) {
                        handled.add(k);
                    }
                }
            }
            sci.onStartup(handled, ctx);
        }
        return c;
    }

    private static Object instance(Object o) throws Exception {
        if (o instanceof String) {
            return Class.forName((String) o).getDeclaredConstructor().newInstance();
        }
        if (o instanceof Class) {
            return ((Class<?>) o).getDeclaredConstructor().newInstance();
        }
        return o;
    }

    void request(String method, String name) throws Exception {
        HttpServletRequest req = (HttpServletRequest) Proxy.newProxyInstance(getClass().getClassLoader(),
            new Class<?>[] {HttpServletRequest.class}, (p, m, a) -> m.getName().equals("getMethod") ? method : defaultValue(m.getReturnType()));
        HttpServletResponse res = (HttpServletResponse) Proxy.newProxyInstance(getClass().getClassLoader(),
            new Class<?>[] {HttpServletResponse.class}, (p, m, a) -> defaultValue(m.getReturnType()));
        new Chain(filters, servlets.get(name)).doFilter(req, res);
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
        private int next;

        Chain(List<Filter> filters, Servlet servlet) {
            this.filters = filters;
            this.servlet = servlet;
        }

        @Override
        public void doFilter(ServletRequest req, ServletResponse res) throws java.io.IOException, javax.servlet.ServletException {
            if (next < filters.size()) {
                filters.get(next++).doFilter(req, res, this);
            } else {
                servlet.service(req, res);
            }
        }
    }
}
