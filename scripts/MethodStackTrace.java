import com.sun.jdi.Bootstrap;
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
import java.io.InputStream;
import java.io.PrintStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Runs a Java program under JDI and, the first time each named method is
 * entered, records the thread's call stack: one frame per line,
 * innermost first, as owner#name+descriptor@bci (bci: the bytecode index
 * the frame is at — for a caller, its call site). A JVM oracle for the
 * real call chain into a method, to compare with a static graph's edges.
 *
 * Usage: java scripts/MethodStackTrace.java CLASSPATH MAIN_CLASS OUT OWNER#NAME...
 * (NAME without descriptor: every overload).
 */
public class MethodStackTrace {
    public static void main(String[] a) throws Exception {
        LaunchingConnector c = Bootstrap.virtualMachineManager().defaultConnector();
        Map<String, Connector.Argument> args = c.defaultArguments();
        args.get("options").setValue("-cp " + a[0]);
        args.get("main").setValue(a[1]);
        VirtualMachine vm = c.launch(args);
        drain(vm.process().getInputStream(), System.out);
        drain(vm.process().getErrorStream(), System.err);
        List<String> targets = new ArrayList<>();
        for (int i = 3; i < a.length; i++) {
            targets.add(a[i]);
            ClassPrepareRequest r = vm.eventRequestManager().createClassPrepareRequest();
            r.addClassFilter(a[i].substring(0, a[i].indexOf('#')));
            r.setSuspendPolicy(EventRequest.SUSPEND_EVENT_THREAD);
            r.enable();
        }
        List<String> out = new ArrayList<>();
        boolean done = false;
        vm.resume();
        while (!done) {
            EventSet set = vm.eventQueue().remove();
            for (Event e : set) {
                if (e instanceof ClassPrepareEvent cp) {
                    ReferenceType t = cp.referenceType();
                    for (String target : targets) {
                        String owner = target.substring(0, target.indexOf('#')), name = target.substring(target.indexOf('#') + 1);
                        if (!owner.equals(t.name())) {
                            continue;
                        }
                        for (Method m : t.methods()) {
                            if ((m.name() + m.signature()).startsWith(name) && m.location() != null) {
                                BreakpointRequest b = vm.eventRequestManager().createBreakpointRequest(m.location());
                                b.setSuspendPolicy(EventRequest.SUSPEND_EVENT_THREAD);
                                b.enable();
                            }
                        }
                    }
                } else if (e instanceof BreakpointEvent be) {
                    Method m = be.location().method();
                    out.add("== " + m.declaringType().name() + "#" + m.name() + m.signature() + " on thread " + be.thread().name());
                    for (StackFrame f : be.thread().frames()) {
                        Location l = f.location();
                        out.add(l.declaringType().name() + "#" + l.method().name() + l.method().signature() + "@" + l.codeIndex());
                    }
                    vm.eventRequestManager().deleteEventRequest(be.request());
                } else if (e instanceof VMDeathEvent || e instanceof VMDisconnectEvent) {
                    done = true;
                }
            }
            if (!done) {
                set.resume();
            }
        }
        Files.write(Path.of(a[2]), out);
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
