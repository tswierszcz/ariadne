package fx.aop.aspect;

import fx.trace.Ran;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Before;

/** @Aspect without @Component: not a bean, advises nothing. */
@Aspect
public class Unregistered {
    @Before("execution(* fx.aop.svc..*(..))")
    public void never() { Ran.stack(); }
}
