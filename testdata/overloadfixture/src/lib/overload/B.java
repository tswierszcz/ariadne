package lib.overload;

/** Two versions of one API: parse(String) before 2.0, parse(Options, String) from 2.0 (ARIADNE-TEST-0101). */
public class B {
    public static void parse(String s) {}
    public static void parse(Options o, String s) {}
}
