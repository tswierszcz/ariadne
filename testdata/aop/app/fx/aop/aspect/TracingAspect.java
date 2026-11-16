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

    /** args(id): id is the one parameter (bound by inference, no debug info). */
    @Before("execution(* fx.aop.svc.Repo+.*(..)) && args(id)")
    public void withArgs(long id) { Ran.stack(); }

    @Pointcut(value = "execution(* fx.aop.svc.Repo+.find(..)) && args(id)", argNames = "id")
    public void byId(long id) {}

    /** A parameterized named pointcut. */
    @Before("byId(id)")
    public void beforeFind(long id) { Ran.stack(); }

    /** bean(name): the default name of the @Repository JdbcRepo. */
    @Before("bean(jdbc*) && execution(* name())")
    public void byBean() { Ran.stack(); }

    /** bean(name) of no bean: never runs. */
    @Before("bean(nothing*)")
    public void noBean() { Ran.stack(); }

    /** target(Type). */
    @Before("target(fx.aop.svc.ReportService)")
    public void byTarget() { Ran.stack(); }

    /** this(variable): bound to the ReportService parameter. */
    @Before("this(svc) && execution(* render(..))")
    public void byThis(fx.aop.svc.ReportService svc) { Ran.stack(); }

    /** @target(Annotation). */
    @Before("@target(fx.aop.svc.Monitored)")
    public void byTargetAnn() { Ran.stack(); }

    /** A method annotation pattern inside execution. */
    @Before("execution(@fx.aop.svc.Audited * *(..))")
    public void byExecAnn() { Ran.stack(); }

    /** returning String: Repo#name only (save is void, find throws). */
    @AfterReturning(pointcut = "execution(* fx.aop.svc.Repo+.*(..))", returning = "s")
    public void onString(String s) { Ran.stack(); }

    /** args(String): save's Object parameter may hold a String at run time (and does). */
    @Before("execution(* fx.aop.svc..*(..)) && args(String)")
    public void stringArg() { Ran.stack(); }
}
