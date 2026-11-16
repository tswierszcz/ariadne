package fx.aop.aspect;

import fx.aop.svc.Audited;
import fx.trace.Ran;
import org.aspectj.lang.JoinPoint;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.AfterReturning;
import org.aspectj.lang.annotation.AfterThrowing;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Before;
import org.aspectj.lang.annotation.Pointcut;
import org.springframework.stereotype.Component;

@Aspect
@Component
public class TracingAspect {
    @Pointcut("execution(public * fx.aop.svc..*Service.*(..))")
    public void services() {}

    /** Every public method of a *Service class (a named pointcut). */
    @Before("services()")
    public void before(JoinPoint jp) {
        Ran.stack();
        Ran.advised("before", jp.getSignature().getDeclaringTypeName(), jp.getSignature().getName());
    }

    /** After Repo#save returns (a subtype pattern). */
    @AfterReturning(pointcut = "execution(* fx.aop.svc.Repo+.save(..))", returning = "r")
    public void afterSave(Object r) { Ran.stack(); }

    /** When a JdbcRepo method throws. */
    @AfterThrowing(pointcut = "within(fx.aop.svc.JdbcRepo)", throwing = "e")
    public void onError(Exception e) { Ran.stack(); }

    /** Around methods annotated @Audited (the binding form). */
    @Around("@annotation(audited)")
    public Object audit(ProceedingJoinPoint pjp, Audited audited) throws Throwable {
        Ran.stack();
        return pjp.proceed();
    }

    /** Methods of classes annotated @Monitored. */
    @Before("@within(fx.aop.svc.Monitored)")
    public void monitor() { Ran.stack(); }

    /** Matches no method: never runs. */
    @Before("execution(* fx.aop.svc.OrderService.never*(..))")
    public void ghost() { Ran.stack(); }

    /** args(...) is not evaluated by spring:aop a1: runs, undecided (the declared gap). */
    @Before("execution(* fx.aop.svc.Repo+.*(..)) && args(id)")
    public void withArgs(long id) { Ran.stack(); }
}
