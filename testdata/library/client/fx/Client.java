package fx;

import lib.api.Codecs;
import lib.api.Formatter;
import lib.api.Pipeline;
import lib.api.Templates;
import lib.internal.Util;

/**
 * The oracle's client: calls every public entry of the library once with
 * a template, and passes its own Listener to Pipeline#run.
 */
public class Client {
    public static void main(String[] args) {
        String t = "v=${java:version}";
        System.out.println("(a) " + new Templates().render(t));
        System.out.println("(b) " + new Formatter().format(t));
        new Pipeline().run(text -> System.out.println("(d) client listener: " + text), t);
        System.out.println("(e) " + Codecs.decodeAll(t));
        System.out.println("(f) " + Util.expand(t));
    }
}
