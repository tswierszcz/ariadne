package fx.web;

import fx.trace.Ran;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import org.springframework.web.servlet.HandlerInterceptor;
import org.springframework.web.servlet.ModelAndView;

/** Registered through WebMvcConfigurer#addInterceptors. */
public class AuditInterceptor implements HandlerInterceptor {
    public AuditInterceptor() { Ran.stack(); }

    @Override
    public boolean preHandle(HttpServletRequest req, HttpServletResponse res, Object handler) { Ran.stack(); return true; }

    @Override
    public void postHandle(HttpServletRequest req, HttpServletResponse res, Object handler, ModelAndView mv) { Ran.stack(); }

    @Override
    public void afterCompletion(HttpServletRequest req, HttpServletResponse res, Object handler, Exception ex) { Ran.stack(); }
}
