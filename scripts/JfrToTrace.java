import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.security.MessageDigest;
import java.time.Instant;
import java.util.*;
import java.util.zip.*;
import jdk.jfr.consumer.*;

/**
 * RUNTIME_NATIVE N0/N1 recorder adapter: converts a Java Flight Recorder recording into an
 * ariadne-runtime-trace-1 file (docs/RUNTIME_TRACE.MD). No Java agent: the data comes from
 * JFR's own events — jdk.ClassLoad (loaded classes with their defining loader),
 * jdk.ExecutionSample (periodic stack samples: methods seen executing, and caller → callee
 * relations measured as adjacent frames) and, on JDK 25+, jdk.MethodTrace (exact entries of
 * filtered methods, with the caller's stack).
 *
 * Only classes found in the given artifacts are recorded (scope "artifacts"); JDK frames
 * between two recorded frames make a "stack-through-runtime" relation. JFR does not say
 * where a class's bytes came from, so class hashes are a classpath lookup.
 *
 * Usage:
 *   java scripts/JfrToTrace.java --jfr rec.jfr --out run.trace --scenario "text"
 *        --artifact app.jar [--artifact dep.jar ...] [--artifact-dir dir-of-jars]
 *        [--run-id id] [--method-filter Class::method ...]
 *
 * Record with, e.g.:
 *   java -XX:StartFlightRecording=filename=rec.jfr,settings=profile,+jdk.ClassLoad#enabled=true \
 *        '-XX:StartFlightRecording=...,jdk.MethodTrace#enabled=true,jdk.MethodTrace#filter=a.B::m' ...
 */
public final class JfrToTrace {
    public static void main(String[] args) throws Exception {
        Path jfr = null, out = null;
        String scenario = null, runId = null;
        List<Path> artifacts = new ArrayList<>();
        List<String> filters = new ArrayList<>();
        for (int i = 0; i < args.length; i++) {
            String a = args[i];
            String v = i + 1 < args.length ? args[i + 1] : null;
            switch (a) {
                case "--jfr" -> { jfr = Path.of(v); i++; }
                case "--out" -> { out = Path.of(v); i++; }
                case "--scenario" -> { scenario = v; i++; }
                case "--run-id" -> { runId = v; i++; }
                case "--artifact" -> { artifacts.add(Path.of(v)); i++; }
                case "--artifact-dir" -> {
                    try (var s = Files.list(Path.of(v))) {
                        s.filter(p -> p.toString().endsWith(".jar")).sorted().forEach(artifacts::add);
                    }
                    i++;
                }
                case "--method-filter" -> { filters.add(v); i++; }
                default -> throw new IllegalArgumentException("unknown argument " + a);
            }
        }
        if (jfr == null || out == null || scenario == null || artifacts.isEmpty()) {
            System.err.println("usage: JfrToTrace --jfr F --out F --scenario S --artifact JAR [...]");
            System.exit(2);
        }
        new JfrToTrace().run(jfr, out, scenario, runId, artifacts, filters);
    }

    record ClassSrc(String artifact, String sha256) {}

    final Map<String, String> artifactSha = new LinkedHashMap<>(); // id -> sha
    final Map<String, String> artifactName = new LinkedHashMap<>();
    final Map<String, ClassSrc> classSrc = new HashMap<>(); // binary name -> first artifact with it
    final Map<String, String[]> loaders = new TreeMap<>(); // id -> {name, type}
    final Map<String, String[]> classes = new TreeMap<>(); // key loader/name -> {id, name, loader}
    final Map<String, String[]> methods = new TreeMap<>(); // key classKey#name+desc -> {id, classKey, name, desc}
    final Set<String> loaded = new TreeSet<>();
    final Map<String, int[]> executed = new TreeMap<>(); // methodKey + "\0" + basis -> count
    final Map<String, int[]> calls = new TreeMap<>(); // caller\0callee\0basis\0source\0bci\0line -> count
    int truncated, samples, traces, classLoads;
    long first = Long.MAX_VALUE, last = Long.MIN_VALUE;
    String jvmName = "", jvmVersion = "", jvmVendor = "", jvmVersionInfo = "", samplingPeriod = "";

    void run(Path jfr, Path out, String scenario, String runId, List<Path> artifacts, List<String> filters) throws Exception {
        int n = 0;
        for (Path p : artifacts) {
            String id = "a" + (++n);
            artifactName.put(id, p.getFileName().toString());
            artifactSha.put(id, sha256(Files.readAllBytes(p)));
            try (ZipFile z = new ZipFile(p.toFile())) {
                var es = Collections.list(z.entries());
                es.sort(Comparator.comparing(ZipEntry::getName));
                for (ZipEntry e : es) {
                    String name = e.getName();
                    if (!name.endsWith(".class") || name.startsWith("META-INF/") || name.endsWith("module-info.class")) continue;
                    String bin = name.substring(name.startsWith("BOOT-INF/classes/") ? 17 : 0, name.length() - 6).replace('/', '.');
                    if (classSrc.containsKey(bin)) continue; // first on the class path wins
                    try (InputStream in = z.getInputStream(e)) {
                        classSrc.put(bin, new ClassSrc(id, sha256(in.readAllBytes())));
                    }
                }
            }
        }
        try (RecordingFile f = new RecordingFile(jfr)) {
            while (f.hasMoreEvents()) {
                RecordedEvent e = f.readEvent();
                String t = e.getEventType().getName();
                switch (t) {
                    case "jdk.ClassLoad" -> classLoad(e);
                    case "jdk.ExecutionSample" -> sample(e);
                    case "jdk.MethodTrace" -> trace(e);
                    case "jdk.JVMInformation" -> { jvmName = e.getString("jvmName"); jvmVersionInfo = e.getString("jvmVersion"); }
                    case "jdk.InitialSystemProperty" -> {
                        switch (e.getString("key")) {
                            case "java.version" -> jvmVersion = e.getString("value");
                            case "java.vendor" -> jvmVendor = e.getString("value");
                            default -> { }
                        }
                    }
                    case "jdk.ActiveSetting" -> {
                        if ("period".equals(e.getString("name")) && samplingPeriod.isEmpty()) {
                            long typeId = e.getLong("id");
                            for (var et : f.readEventTypes()) {
                                if (et.getId() == typeId && et.getName().equals("jdk.ExecutionSample")) samplingPeriod = e.getString("value");
                            }
                        }
                    }
                    default -> { continue; }
                }
                long ts = e.getStartTime().toEpochMilli();
                first = Math.min(first, ts);
                last = Math.max(last, ts);
            }
        }
        if (jvmVersion.isEmpty()) jvmVersion = jvmVersionInfo; // jdk.InitialSystemProperty not recorded
        write(out, scenario, runId == null ? sha256(Files.readAllBytes(jfr)).substring(0, 16) : runId, filters);
        System.err.printf("JfrToTrace: %d class loads, %d samples, %d method traces → %d classes, %d methods, %d executed, %d calls (%d truncated stacks)%n",
                classLoads, samples, traces, classes.size(), methods.size(), executed.size(), calls.size(), truncated);
    }

    String loaderId(RecordedClassLoader l) {
        if (l == null) return "l0";
        String id = "l" + l.getId();
        loaders.putIfAbsent(id, new String[] {l.getName() == null ? "" : l.getName(), l.getType() == null ? "" : l.getType().getName()});
        return id;
    }

    /** The class's key, or null when it is outside the artifacts (out of scope). */
    String classKey(RecordedClass c) {
        if (c == null || !classSrc.containsKey(c.getName())) return null;
        String loader = loaderId(c.getClassLoader());
        if (loader.equals("l0")) loaders.putIfAbsent("l0", new String[] {"bootstrap", ""});
        String key = loader + "/" + c.getName();
        classes.putIfAbsent(key, new String[] {null, c.getName(), loader});
        return key;
    }

    String methodKey(RecordedMethod m) {
        if (m == null) return null;
        String ck = classKey(m.getType());
        if (ck == null) return null;
        String key = ck + "#" + m.getName() + m.getDescriptor();
        methods.putIfAbsent(key, new String[] {null, ck, m.getName(), m.getDescriptor()});
        return key;
    }

    void classLoad(RecordedEvent e) {
        classLoads++;
        RecordedClass c = e.getValue("loadedClass");
        String ck = classKey(c);
        if (ck != null) loaded.add(ck);
    }

    void sample(RecordedEvent e) {
        samples++;
        RecordedStackTrace st = e.getStackTrace();
        if (st == null) return;
        if (st.isTruncated()) truncated++;
        List<RecordedFrame> fr = st.getFrames();
        String[] keys = new String[fr.size()];
        for (int i = 0; i < fr.size(); i++) keys[i] = fr.get(i).isJavaFrame() ? methodKey(fr.get(i).getMethod()) : null;
        Set<String> seen = new HashSet<>();
        for (int i = 0; i < keys.length; i++) {
            if (keys[i] == null) continue;
            if (seen.add(keys[i])) executed.computeIfAbsent(keys[i] + "\0sampled", k -> new int[1])[0]++;
            int j = i + 1;
            while (j < keys.length && keys[j] == null) j++;
            if (j < keys.length) addCall(keys[j], keys[i], j == i + 1 ? "stack-adjacent" : "stack-through-runtime", "sampled", fr.get(j));
        }
    }

    void trace(RecordedEvent e) {
        traces++;
        String mk = methodKey(e.getValue("method"));
        if (mk == null) return;
        executed.computeIfAbsent(mk + "\0traced", k -> new int[1])[0]++;
        RecordedStackTrace st = e.getStackTrace();
        if (st == null) return;
        if (st.isTruncated()) truncated++;
        List<RecordedFrame> fr = st.getFrames();
        for (int j = 0; j < fr.size(); j++) {
            String ck = fr.get(j).isJavaFrame() ? methodKey(fr.get(j).getMethod()) : null;
            if (ck != null) {
                addCall(ck, mk, j == 0 ? "stack-adjacent" : "stack-through-runtime", "traced", fr.get(j));
                return;
            }
        }
    }

    void addCall(String caller, String callee, String basis, String source, RecordedFrame callerFrame) {
        int bci = callerFrame.getBytecodeIndex();
        int line = callerFrame.getLineNumber();
        String k = caller + "\0" + callee + "\0" + basis + "\0" + source + "\0" + bci + "\0" + line;
        calls.computeIfAbsent(k, x -> new int[1])[0]++;
    }

    void write(Path out, String scenario, String runId, List<String> filters) throws IOException {
        // stable ids
        int n = 0;
        for (String[] c : classes.values()) c[0] = "c" + (++n);
        n = 0;
        for (String[] m : methods.values()) m[0] = "m" + (++n);
        List<String> reasons = new ArrayList<>();
        boolean sampled = samples > 0;
        if (sampled) reasons.add("periodic stack sampling: a method that ran between samples is not recorded");
        if (traces > 0 || !filters.isEmpty()) reasons.add("exact entries only for the filtered methods");
        if (truncated > 0) reasons.add(truncated + " stack traces truncated by JFR's stack depth: frames below the cut are missing");
        reasons.add("JFR does not report where class bytes came from: class hashes are a classpath lookup");
        List<String> kinds = new ArrayList<>(List.of("loaded", "executed", "call"));
        try (Writer w = new BufferedWriter(new OutputStreamWriter(Files.newOutputStream(out), StandardCharsets.UTF_8))) {
            StringBuilder h = new StringBuilder("{\"record\":\"header\",\"format\":\"ariadne-runtime-trace-1\"");
            h.append(",\"run_id\":").append(q(runId)).append(",\"scenario\":").append(q(scenario));
            h.append(",\"started\":").append(q(first == Long.MAX_VALUE ? "" : Instant.ofEpochMilli(first).toString()));
            h.append(",\"ended\":").append(q(last == Long.MIN_VALUE ? "" : Instant.ofEpochMilli(last).toString()));
            h.append(",\"runtime\":{\"name\":").append(q(jvmName)).append(",\"version\":").append(q(jvmVersion)).append(",\"vendor\":").append(q(jvmVendor)).append("}");
            h.append(",\"recorder\":{\"name\":\"ariadne-jfr-convert\",\"version\":\"1\",\"source\":\"JFR jdk.ClassLoad, jdk.ExecutionSample, jdk.MethodTrace\"}");
            h.append(",\"scope\":{\"event_kinds\":").append(arr(kinds)).append(",\"classes\":\"artifacts\",\"method_filters\":").append(arr(filters)).append("}");
            if (sampled) {
                long ms = periodMillis(samplingPeriod);
                h.append(",\"sampling\":{\"kind\":\"periodic\",\"interval_ms\":").append(ms).append("}");
            } else {
                h.append(",\"sampling\":{\"kind\":\"none\",\"interval_ms\":0}");
            }
            h.append(",\"completeness\":{\"complete\":false,\"lost_events\":0,\"truncated_stacks\":").append(truncated).append(",\"reasons\":").append(arr(reasons)).append("}");
            h.append(",\"data_policy\":\"identifiers-only\"}");
            w.write(h + "\n");
            for (var a : artifactSha.entrySet()) {
                w.write("{\"record\":\"artifact\",\"id\":" + q(a.getKey()) + ",\"name\":" + q(artifactName.get(a.getKey())) + ",\"sha256\":" + q(a.getValue()) + "}\n");
            }
            for (var l : loaders.entrySet()) {
                w.write("{\"record\":\"loader\",\"id\":" + q(l.getKey()) + ",\"name\":" + q(l.getValue()[0]) + ",\"type\":" + q(l.getValue()[1]) + "}\n");
            }
            for (String[] c : classes.values()) {
                ClassSrc src = classSrc.get(c[1]);
                w.write("{\"record\":\"class\",\"id\":" + q(c[0]) + ",\"name\":" + q(c[1]) + ",\"loader\":" + q(c[2])
                        + ",\"artifact\":" + q(src.artifact()) + ",\"sha256\":" + q(src.sha256()) + ",\"source_basis\":\"classpath-lookup\"}\n");
            }
            for (String[] m : methods.values()) {
                w.write("{\"record\":\"method\",\"id\":" + q(m[0]) + ",\"class\":" + q(classes.get(m[1])[0]) + ",\"name\":" + q(m[2]) + ",\"desc\":" + q(m[3]) + "}\n");
            }
            int e = 0;
            for (String ck : loaded) {
                w.write("{\"record\":\"loaded\",\"id\":\"e" + (++e) + "\",\"class\":" + q(classes.get(ck)[0]) + "}\n");
            }
            for (var x : executed.entrySet()) {
                String[] p = x.getKey().split("\0");
                w.write("{\"record\":\"executed\",\"id\":\"e" + (++e) + "\",\"method\":" + q(methods.get(p[0])[0]) + ",\"basis\":" + q(p[1]) + ",\"count\":" + x.getValue()[0] + "}\n");
            }
            for (var x : calls.entrySet()) {
                String[] p = x.getKey().split("\0");
                StringBuilder b = new StringBuilder("{\"record\":\"call\",\"id\":\"e" + (++e) + "\",\"caller\":" + q(methods.get(p[0])[0]) + ",\"callee\":" + q(methods.get(p[1])[0])
                        + ",\"basis\":" + q(p[2]) + ",\"source\":" + q(p[3]) + ",\"count\":" + x.getValue()[0]);
                if (Integer.parseInt(p[4]) >= 0) b.append(",\"bci\":").append(p[4]);
                if (Integer.parseInt(p[5]) >= 0) b.append(",\"line\":").append(p[5]);
                w.write(b + "}\n");
            }
        }
    }

    static long periodMillis(String s) {
        // JFR settings like "10 ms", "20 ms", "1 s"; unknown → 0 (reported as unknown)
        String[] p = s.trim().split("\\s+");
        try {
            long v = Long.parseLong(p[0]);
            return p.length > 1 && p[1].startsWith("s") ? v * 1000 : v;
        } catch (RuntimeException ex) {
            return 0;
        }
    }

    static String sha256(byte[] b) throws Exception {
        return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(b));
    }

    static String arr(List<String> xs) {
        StringJoiner j = new StringJoiner(",", "[", "]");
        for (String x : xs) j.add(q(x));
        return j.toString();
    }

    static String q(String s) {
        StringBuilder b = new StringBuilder("\"");
        for (char c : s.toCharArray()) {
            switch (c) {
                case '"' -> b.append("\\\"");
                case '\\' -> b.append("\\\\");
                case '\n' -> b.append("\\n");
                case '\r' -> b.append("\\r");
                case '\t' -> b.append("\\t");
                default -> { if (c < 0x20) b.append(String.format("\\u%04x", (int) c)); else b.append(c); }
            }
        }
        return b.append('"').toString();
    }
}
