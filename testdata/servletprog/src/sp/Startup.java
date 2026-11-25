package sp;

import javax.servlet.ServletContext;
import javax.servlet.ServletException;

/** Spring's WebApplicationInitializer in miniature. */
interface Startup {
    void onStartup(ServletContext ctx) throws ServletException;
}

/**
 * Spring's AbstractDispatcherServletInitializer: registers the servlet a
 * factory method returns, declared as an abstract servlet type.
 */
abstract class BaseStartup implements Startup {
    @Override
    public void onStartup(ServletContext ctx) throws ServletException {
        ctx.addServlet("main", createServlet());
    }

    protected FrameworkLikeServlet createServlet() {
        return new MainServlet();
    }
}

/** The application's initializer (Lavagna's DispatcherServletInitializer). */
final class AppStartup extends BaseStartup {
    @Override
    public void onStartup(ServletContext ctx) throws ServletException {
        super.onStartup(ctx);
        ctx.addServlet("ping", new PingServlet());
        addFilter(ctx, "audit", AuditFilter.class);
        ctx.addFilter("tx", "sp.TxFilter");
    }

    private static void addFilter(ServletContext ctx, String name, Class<? extends javax.servlet.Filter> filter) {
        ctx.addFilter(name, filter);
    }
}

/** Abstract and never subclassed: the container never creates it. */
abstract class UnusedStartup implements Startup {
    @Override
    public void onStartup(ServletContext ctx) {
        ctx.addServlet("ghost", new GhostServlet());
    }
}

/** Registers a servlet but is not a Startup: nothing runs it. */
final class NotAStartup {
    void register(ServletContext ctx) {
        ctx.addFilter("ghost", new GhostFilter());
    }
}
