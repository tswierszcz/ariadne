package sd;

import java.io.IOException;
import javax.servlet.Filter;
import javax.servlet.FilterChain;
import javax.servlet.GenericServlet;
import javax.servlet.ServletException;
import javax.servlet.ServletRequest;
import javax.servlet.ServletResponse;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

/** Registered in web.xml: an HttpServlet reached through HttpServlet's own template dispatch. */
class HelloServlet extends HttpServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse res) {
        Ran.here();
        ETags.validate(req);
    }

    @Override
    protected void doPost(HttpServletRequest req, HttpServletResponse res) {
        Ran.here();
    }
}

/** The library code a request reaches (the analog of ServletWebRequest#validateIfNoneMatch). */
final class ETags {
    static boolean validate(HttpServletRequest req) {
        Ran.here();
        return req.getHeader("If-None-Match") == null;
    }
}

/** Registered in web.xml: a servlet with its own service(ServletRequest, ServletResponse). */
class PlainServlet extends GenericServlet {
    @Override
    public void service(ServletRequest req, ServletResponse res) {
        Ran.here();
    }
}

/** Registered in web.xml. */
class AuditFilter implements Filter {
    @Override
    public void doFilter(ServletRequest req, ServletResponse res, FilterChain chain) throws IOException, ServletException {
        Ran.here();
        chain.doFilter(req, res);
    }
}

/** On the class path, never registered: the container never dispatches to it. */
class UnusedServlet extends HttpServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse res) {
        Ran.here();
    }

    @Override
    public void service(ServletRequest req, ServletResponse res) {
        Ran.here();
    }
}

/** On the class path, never registered. */
class UnusedFilter implements Filter {
    @Override
    public void doFilter(ServletRequest req, ServletResponse res, FilterChain chain) {
        Ran.here();
    }
}
