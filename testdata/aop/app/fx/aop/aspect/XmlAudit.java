package fx.aop.aspect;

import fx.trace.Ran;

/** A plain class used as an <aop:aspect> in fx/aop/aop.xml. */
public class XmlAudit {
    public void audit() { Ran.stack(); }

    /** Its pointcut matches no method: never runs. */
    public void ghost() { Ran.stack(); }
}
