package lib.overload;

/** The vulnerable overload is parse(String); the app only calls parse(InputStream) (ARIADNE-TEST-0104). */
public class C {
    public static void parse(String s) {}
    public static void parse(java.io.InputStream in) {}
}
