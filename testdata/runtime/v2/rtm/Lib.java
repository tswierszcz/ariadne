package rtm;

/** Stands in for a vulnerable library method. */
public final class Lib {
    public static long parse(String s) {
        long h = 0;
        for (int i = 0; i < s.length(); i++) {
            h = h * 31 + s.charAt(i);
        }
        return h;
    }
}
