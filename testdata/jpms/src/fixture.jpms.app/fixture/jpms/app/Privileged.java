package fixture.jpms.app;

import fixture.jpms.api.Codec;
import java.lang.reflect.Method;
import java.security.AccessController;
import java.security.PrivilegedAction;
import java.security.PrivilegedExceptionAction;
import java.util.ServiceLoader;

// AccessController.doPrivileged: runs action.run() synchronously and
// returns its result (ASYNC_CALLBACKS catalog + DATA_FLOW D3 summary).
@SuppressWarnings("removal")
public class Privileged {
    public static void anon() {
        ServiceLoader<Codec> loader = AccessController.doPrivileged(new PrivilegedAction<ServiceLoader<Codec>>() {
            @Override
            public ServiceLoader<Codec> run() {
                return ServiceLoader.load(Codec.class);
            }
        });
        for (Codec c : loader) {
            System.out.println(c.encode("anon"));
        }
    }

    public static void lambda() {
        PrivilegedAction<ServiceLoader<Codec>> action = () -> ServiceLoader.load(Codec.class);
        for (Codec c : AccessController.doPrivileged(action)) {
            System.out.println(c.encode("lambda"));
        }
    }

    public static void exception() throws Exception {
        PrivilegedExceptionAction<ServiceLoader<Codec>> action = () -> ServiceLoader.load(Codec.class);
        for (Codec c : AccessController.doPrivileged(action)) {
            System.out.println(c.encode("exception"));
        }
    }

    // The action comes from outside: not provable here.
    public static void given(PrivilegedAction<ServiceLoader<Codec>> action) {
        for (Codec c : AccessController.doPrivileged(action)) {
            System.out.println(c.encode("given"));
        }
    }

    // A reflective lookup performed inside the privileged action.
    public static void reflective() throws Exception {
        Method m = AccessController.doPrivileged(new PrivilegedExceptionAction<Method>() {
            @Override
            public Method run() throws Exception {
                return Target.class.getDeclaredMethod("hit");
            }
        });
        m.invoke(null);
    }

    static class Target {
        static void hit() {
        }
    }
}
