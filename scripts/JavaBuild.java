import javax.tools.*;
import java.nio.file.*;
import java.util.*;
import java.util.spi.ToolProvider;

/**
 * Compiles every .java file under a source directory and jars the classes,
 * using the JDK's own compiler and jar tool through the java.compiler /
 * ToolProvider APIs — for environments with a Java runtime that has the
 * jdk.compiler module but no javac/jar binaries.
 *
 * Usage: java scripts/JavaBuild.java SRC_DIR OUT_JAR TARGET [CLASSPATH]
 */
public class JavaBuild {
    public static void main(String[] a) throws Exception {
        Path src = Path.of(a[0]), out = Path.of(a[1]);
        Path classes = Files.createTempDirectory("javabuild");
        List<String> files = new ArrayList<>();
        try (var s = Files.walk(src)) { s.filter(p -> p.toString().endsWith(".java")).sorted().forEach(p -> files.add(p.toString())); }
        // -source/-target rather than --release: a runtime without lib/ct.sym
        // cannot check the API of an older release, only emit its class files.
        List<String> args = new ArrayList<>(List.of("-source", a[2], "-target", a[2], "-Xlint:-options", "-d", classes.toString()));
        if (a.length > 3) { args.add("-cp"); args.add(a[3]); }
        args.addAll(files);
        int rc = ToolProvider.findFirst("javac").orElseThrow().run(System.out, System.err, args.toArray(String[]::new));
        if (rc != 0) System.exit(rc);
        Files.deleteIfExists(out);
        rc = ToolProvider.findFirst("jar").orElseThrow().run(System.out, System.err, "--create", "--file", out.toString(), "--date", "2026-01-01T00:00:00Z", "-C", classes.toString(), ".");
        System.exit(rc);
    }
}
