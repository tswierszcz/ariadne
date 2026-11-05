package oracle;

import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import java.lang.reflect.Modifier;

// SERVICE_LOADING S4 oracle (ROADMAP B4): runs fixture entry methods on a
// real JVM; fixture providers print "ORACLE-PROVIDER <class>" when the
// JVM instantiates them. Not part of any analyzed input.
public final class OracleRunner {
    public static void main(String[] specs) throws Exception {
        for (String spec : specs) {
            String[] parts = spec.split("#");
            System.out.println("ORACLE-ENTRY " + spec);
            Class<?> cls = Class.forName(parts[0]);
            Method target = null;
            for (Method m : cls.getDeclaredMethods()) {
                if (m.getName().equals(parts[1])) {
                    target = m;
                }
            }
            if (target == null) {
                throw new IllegalArgumentException("no method " + spec);
            }
            target.setAccessible(true);
            Object[] args = target.getParameterTypes().length == 1 ? new Object[] {new String[0]} : new Object[0];
            Object receiver = null;
            if (!Modifier.isStatic(target.getModifiers())) {
                java.lang.reflect.Constructor<?> c = cls.getDeclaredConstructor();
                c.setAccessible(true);
                receiver = c.newInstance();
            }
            try {
                target.invoke(receiver, args);
            } catch (InvocationTargetException e) {
                System.out.println("ORACLE-ERROR " + e.getCause().getClass().getName());
            }
        }
    }
}
