package px.client;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;
import px.archive.Unpacker;

/**
 * RH-A3 oracle client: public API only, local files. Prints one "effect:"
 * line per scenario: what the JVM did to the file system, independent of
 * any static claim.
 */
public class Main {
    public static void main(String[] a) throws Exception {
        Path base = Files.createTempDirectory("pilot");
        Path dest = Files.createDirectories(base.resolve("dest"));
        // normal input: one entry inside the directory
        run("normal", dest, base, "a.txt");
        // candidate input: an entry that climbs out of the directory
        run("traversal", dest, base, "../escaped.txt");
        // control: the API that writes where the caller says, by contract
        Unpacker.writeFile(base.resolve("chosen.txt"), new byte[] {1});
        System.out.println("effect: allowed-api wrote " + Files.exists(base.resolve("chosen.txt")));
        try (var s = Files.walk(base)) {
            s.sorted(java.util.Comparator.reverseOrder()).forEach(p -> p.toFile().delete());
        }
    }

    static void run(String scenario, Path dest, Path base, String name) throws Exception {
        ByteArrayOutputStream buf = new ByteArrayOutputStream();
        try (ZipOutputStream z = new ZipOutputStream(buf)) {
            z.putNextEntry(new ZipEntry(name));
            z.write("x".getBytes("UTF-8"));
            z.closeEntry();
        }
        String outcome;
        try {
            int n = new Unpacker(dest).unpack(new ByteArrayInputStream(buf.toByteArray()));
            outcome = "unpacked " + n;
        } catch (java.io.IOException e) {
            outcome = "refused";
        }
        boolean outside = Files.exists(base.resolve("escaped.txt"));
        System.out.println("effect: " + scenario + " " + outcome + ", file outside dest " + outside);
    }
}
