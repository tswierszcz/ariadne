package rx;

public class Work {
    static synchronized void hit(String s) {
        System.out.println(s);
    }

    static void consumed(String s) {
        hit("consumed " + s);
    }
}
