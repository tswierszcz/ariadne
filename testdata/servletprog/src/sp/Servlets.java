package sp;

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

/** Spring's FrameworkServlet: the declared return type of the factory. */
abstract class FrameworkLikeServlet extends HttpServlet {
}

/** Spring's DispatcherServlet: reached through HttpServlet's template dispatch. */
final class MainServlet extends FrameworkLikeServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse res) {
        Ran.here();
    }
}

final class PingServlet extends GenericServlet {
    @Override
    public void service(ServletRequest req, ServletResponse res) {
        Ran.here();
    }
}

final class AuditFilter implements Filter {
    @Override
    public void doFilter(ServletRequest req, ServletResponse res, FilterChain chain) throws IOException, ServletException {
        Ran.here();
        chain.doFilter(req, res);
    }
}

final class TxFilter implements Filter {
    @Override
    public void doFilter(ServletRequest req, ServletResponse res, FilterChain chain) throws IOException, ServletException {
        Ran.here();
        chain.doFilter(req, res);
    }
}

/** Registered only by UnusedStartup, which nothing creates. */
final class GhostServlet extends GenericServlet {
    @Override
    public void service(ServletRequest req, ServletResponse res) {
        Ran.here();
    }
}

/** Registered only by NotAStartup, which nothing runs. */
final class GhostFilter implements Filter {
    @Override
    public void doFilter(ServletRequest req, ServletResponse res, FilterChain chain) {
        Ran.here();
    }
}
