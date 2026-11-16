import com.sun.jdi.Bootstrap;
import com.sun.jdi.Method;
import com.sun.jdi.VirtualMachine;
import com.sun.jdi.connect.Connector;
import com.sun.jdi.connect.LaunchingConnector;
import com.sun.jdi.event.Event;
import com.sun.jdi.event.EventSet;
import com.sun.jdi.event.MethodEntryEvent;
import com.sun.jdi.event.VMDeathEvent;
import com.sun.jdi.event.VMDisconnectEvent;
import com.sun.jdi.request.EventRequest;
import com.sun.jdi.request.MethodEntryRequest;
import java.io.InputStream;
import java.io.PrintStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Map;
import java.util.TreeSet;

/**
 * Runs a Java program under the JDK's debugger interface (JDI) and writes
 * every method body entered in the classes the given prefixes match, as
 * owner#name+descriptor, one per line, sorted — a JVM oracle for library
 * code the program cannot print from.
 *
 * Usage: java scripts/MethodTrace.java CLASSPATH MAIN_CLASS OUT PREFIX...
 * (a prefix is a class-name pattern for MethodEntryRequest#addClassFilter,
 * e.g. "org.postgresql.*").
 */
public class MethodTrace {
    public static void main(String[] a) throws Exception {
        LaunchingConnector c = Bootstrap.virtualMachineManager().defaultConnector();
        Map<String, Connector.Argument> args = c.defaultArguments();
        args.get("options").setValue("-cp " + a[0]);
        args.get("main").setValue(a[1]);
        VirtualMachine vm = c.launch(args);
        drain(vm.process().getInputStream(), System.out);
        drain(vm.process().getErrorStream(), System.err);
        for (int i = 3; i < a.length; i++) {
            MethodEntryRequest r = vm.eventRequestManager().createMethodEntryRequest();
            r.addClassFilter(a[i]);
            r.setSuspendPolicy(EventRequest.SUSPEND_EVENT_THREAD);
            r.enable();
        }
        TreeSet<String> seen = new TreeSet<>();
        boolean done = false;
        vm.resume();
        while (!done) {
            EventSet set = vm.eventQueue().remove();
            for (Event e : set) {
                if (e instanceof MethodEntryEvent me) {
                    Method m = me.method();
                    seen.add(m.declaringType().name() + "#" + m.name() + m.signature());
                } else if (e instanceof VMDeathEvent || e instanceof VMDisconnectEvent) {
                    done = true;
                }
            }
            if (!done) {
                set.resume();
            }
        }
        Files.write(Path.of(a[2]), seen);
    }

    private static void drain(InputStream in, PrintStream out) {
        Thread t = new Thread(() -> {
            try {
                in.transferTo(out);
            } catch (Exception ignored) {
                // the target exited
            }
        });
        t.setDaemon(true);
        t.start();
    }
}
