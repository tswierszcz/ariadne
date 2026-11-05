import org.apache.maven.artifact.versioning.ComparableVersion;
import java.nio.file.*; import java.util.*;
public class Order {
  public static void main(String[] a) throws Exception {
    List<String> vs = new ArrayList<>(Files.readAllLines(Path.of(a[0])));
    vs.removeIf(String::isBlank);
    vs.sort((x, y) -> new ComparableVersion(x).compareTo(new ComparableVersion(y)));
    int rank = 0; String prev = null; StringBuilder sb = new StringBuilder();
    for (String v : vs) {
      if (prev != null && new ComparableVersion(prev).compareTo(new ComparableVersion(v)) != 0) rank++;
      sb.append(rank).append('\t').append(v).append('\n'); prev = v;
    }
    System.out.print(sb);
  }
}
