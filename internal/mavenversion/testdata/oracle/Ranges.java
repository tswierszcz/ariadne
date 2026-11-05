import org.apache.maven.artifact.versioning.*;
public class Ranges {
  public static void main(String[] a) throws Exception {
    String[] ranges = {"[1.0,2.0)", "(,1.0]", "[1.5]", "[1.0,)", "(1.0,2.0]", "[1,2),[3,4)", "(,1.2.13),[2.0,2.0.5)", "[2.0.0-RC1,2.0.0]", "[9.0.0,9.0.46)", "(,2.0)", "[2.11.4.1,)", "[1.4.18,1.4.20)"};
    String[] versions = {"0.9", "1", "1.0", "1.0.0", "1.0-ga", "1.0.1", "1.2.12", "1.2.13", "1.5", "1.5.0", "1.9.9", "1.99", "2", "2.0", "2.0-rc1", "2.0.0-RC1", "2.0.0-M1", "2.0.0", "2.0.4", "2.0.5", "2.0.SP1", "2.11.4", "2.11.4.1", "3", "3.5", "4.0", "9.0.45", "9.0.46", "9.0.100", "1.4.18", "1.4.19", "1.4.20", "30"};
    for (String r : ranges) for (String v : versions)
      System.out.println(r + "\t" + v + "\t" + (VersionRange.createFromVersionSpec(r).containsVersion(new DefaultArtifactVersion(v)) ? "in" : "out"));
  }
}
