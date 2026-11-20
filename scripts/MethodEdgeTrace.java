import com.sun.jdi.AbsentInformationException;
import com.sun.jdi.Bootstrap;
import com.sun.jdi.IncompatibleThreadStateException;
import com.sun.jdi.Location;
import com.sun.jdi.Method;
import com.sun.jdi.ReferenceType;
import com.sun.jdi.StackFrame;
import com.sun.jdi.VirtualMachine;
import com.sun.jdi.connect.Connector;
import com.sun.jdi.connect.LaunchingConnector;
import com.sun.jdi.event.BreakpointEvent;
import com.sun.jdi.event.ClassPrepareEvent;
import com.sun.jdi.event.Event;
import com.sun.jdi.event.EventSet;
import com.sun.jdi.event.VMDeathEvent;
import com.sun.jdi.event.VMDisconnectEvent;
import com.sun.jdi.request.BreakpointRequest;
import com.sun.jdi.request.ClassPrepareRequest;
import com.sun.jdi.request.EventRequest;
import com.sun.jdi.request.EventRequestManager;
import java.io.InputStream;
import java.io.PrintStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;
import java.util.TreeSet;

/**
 * A JVM oracle for call-graph edges (ASYNC_CALLBACKS a3-3): like
 * MethodBreakTrace, but the breakpoints stay, and every entry of a method of
 * a named class (exact name or a ClassPrepareRequest-style pattern) records
 * the pair caller -> callee, the caller being the nearest frame that is not
 * a lambda's hidden class (ariadne links a lambda's creator straight to its
 * implementation). Writes the distinct pairs, sorted, one
 * "caller<TAB>callee" (owner#name+descriptor) per line; a pair whose caller
 * is a thread's first frame is the thread's entry. Slow — every entry
 * stops the thread: for small programs.
 *
 * Usage: java scripts/MethodEdgeTrace.java CLASSPATH MAIN_CLASS OUT CLASS...
 */
public class MethodEdgeTrace {
    public static void main(String[] a) throws Exception {
        LaunchingConnector c = Bootstrap.virtualMachineManager().defaultConnector();
        Map<String, Connector.Argument> args = c.defaultArguments();
        args.get("options").setValue("-cp " + a[0]);
        args.get("main").setValue(a[1]);
        VirtualMachine vm = c.launch(args);
        drain(vm.process().getInputStream(), System.out);
        drain(vm.process().getErrorStream(), System.err);
        EventRequestManager erm = vm.eventRequestManager();
        for (int i = 3; i < a.length; i++) {
            ClassPrepareRequest r = erm.createClassPrepareRequest();
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
                if (e instanceof ClassPrepareEvent cp) {
                    arm(erm, cp.referenceType());
                } else if (e instanceof BreakpointEvent be) {
                    Method m = be.location().method();
                    String callee = m.declaringType().name() + "#" + m.name() + m.signature();
                    String caller = "<thread>";
                    try {
                        List<StackFrame> frames = be.thread().frames();
                        for (int i = 1; i < frames.size(); i++) {
                            Method cm = frames.get(i).location().method();
                            if (!cm.declaringType().name().contains("$$Lambda")) {
                                caller = cm.declaringType().name() + "#" + cm.name() + cm.signature();
                                break;
                            }
                        }
                    } catch (IncompatibleThreadStateException ex) {
                        caller = "<unknown>";
                    }
                    seen.add(caller + "\t" + callee);
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

    private static void arm(EventRequestManager erm, ReferenceType t) {
        for (Method m : t.methods()) {
            if (m.isAbstract() || m.isNative()) {
                continue;
            }
            Location at = m.location();
            if (at == null || at.codeIndex() < 0) {
                continue;
            }
            BreakpointRequest b = erm.createBreakpointRequest(at);
            b.setSuspendPolicy(EventRequest.SUSPEND_EVENT_THREAD);
            b.enable();
        }
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
