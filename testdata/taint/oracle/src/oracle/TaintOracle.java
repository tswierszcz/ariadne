package oracle;

import java.io.ByteArrayOutputStream;
import java.io.ObjectInputFilter;
import java.io.ObjectOutputStream;
import java.lang.reflect.Constructor;
import java.lang.reflect.Method;
import java.util.ArrayList;
import java.util.Base64;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

/**
 * DATA_FLOW D5 real-JVM oracle for internal/taint. Calls every Spring
 * MVC handler (stub @PostMapping) of the fixture's request classes with
 * a serialized Canary as request data — a Base64 String, a FakeRequest
 * returning it, or the raw bytes — and records, through a process-wide
 * filter factory, every entry (the method calling ObjectInputStream's
 * read) that actually read a Canary, also when the entry's own filter
 * then rejected it. Output (stdout, TSV): entry (owner#name+desc), then
 * "read" or "rejected".
 */
public final class TaintOracle {
    static final Map<String, String> seen = new TreeMap<>();
    static final StackWalker walker = StackWalker.getInstance(StackWalker.Option.RETAIN_CLASS_REFERENCE);

    static void record(String how) {
        StackWalker.StackFrame f = walker.walk(s -> s.filter(x -> {
            String c = x.getClassName();
            return !c.startsWith("java.") && !c.startsWith("jdk.") && !c.startsWith("oracle.");
        }).findFirst().orElse(null));
        if (f != null) {
            seen.putIfAbsent(f.getClassName() + "#" + f.getMethodName() + f.getDescriptor(), how);
        }
    }

    public static void main(String[] args) throws Exception {
        // Every stream's filter is wrapped: the entry's own filter (if
        // any) decides, and a Canary is recorded with that decision.
        ObjectInputFilter.Config.setSerialFilterFactory((current, next) -> info -> {
            ObjectInputFilter.Status st = next == null ? ObjectInputFilter.Status.UNDECIDED : next.checkInput(info);
            if (info.serialClass() == Canary.class) {
                record(st == ObjectInputFilter.Status.REJECTED ? "rejected" : "read");
            }
            return st;
        });
        ByteArrayOutputStream bos = new ByteArrayOutputStream();
        try (ObjectOutputStream out = new ObjectOutputStream(bos)) {
            out.writeObject(new Canary());
        }
        byte[] bytes = bos.toByteArray();
        String b64 = Base64.getEncoder().encodeToString(bytes);
        Class<?> fake = Class.forName("demo.FakeRequest");
        Constructor<?> fakeCtor = fake.getDeclaredConstructor(String.class);
        fakeCtor.setAccessible(true);
        List<String> called = new ArrayList<>();
        for (String cn : new String[] {"demo.Endpoints", "demo.Extra", "demo.Shapes"}) {
            Class<?> c = Class.forName(cn);
            Object target = c.getDeclaredConstructor().newInstance();
            Method[] ms = c.getDeclaredMethods();
            java.util.Arrays.sort(ms, Comparator.comparing(Method::getName));
            for (Method m : ms) {
                boolean handler = false;
                for (var a : m.getAnnotations()) {
                    if (a.annotationType().getName().equals("org.springframework.web.bind.annotation.PostMapping")) {
                        handler = true;
                    }
                }
                if (!handler) {
                    continue;
                }
                Object[] argv = new Object[m.getParameterCount()];
                Class<?>[] pt = m.getParameterTypes();
                for (int i = 0; i < pt.length; i++) {
                    if (pt[i] == String.class) {
                        argv[i] = b64;
                    } else if (pt[i] == byte[].class) {
                        argv[i] = bytes;
                    } else if (pt[i].getName().equals("javax.servlet.ServletRequest")) {
                        argv[i] = fakeCtor.newInstance(b64);
                    }
                }
                m.setAccessible(true);
                try {
                    m.invoke(target, argv);
                } catch (Throwable t) {
                    // a rejected or failing read is still recorded by the filter
                }
                called.add(cn + "#" + m.getName());
            }
            if (cn.equals("demo.Extra")) {
                // what keep() stored, read later through the field
                Method m = c.getDeclaredMethod("fromField");
                try {
                    m.invoke(target);
                } catch (Throwable t) {
                }
                called.add(cn + "#fromField");
            }
        }
        System.out.println("# handlers called with request data: " + String.join(", ", called));
        for (Map.Entry<String, String> e : seen.entrySet()) {
            System.out.println(e.getKey() + "\t" + e.getValue());
        }
    }
}
