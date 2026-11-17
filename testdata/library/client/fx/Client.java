package fx;

import lib.api.Codecs;
import lib.api.Expander;
import lib.api.Formatter;
import lib.api.Pipeline;
import lib.api.SafeExpander;
import lib.api.StrictExpander;
import lib.api.Templates;
import lib.internal.Util;

/**
 * The oracle's client: calls every public entry of the library once with
 * a template, and passes its own Listener to Pipeline#run. With the
 * argument "strict" it runs case (i) alone, so its trace shows what the
 * inherited replace runs on a StrictExpander.
 */
public class Client {
    public static void main(String[] args) {
        String t = "v=${java:version}";
        if (args.length > 0 && args[0].equals("strict")) {
            System.out.println("(i) " + new StrictExpander().replace(t));
            return;
        }
        System.out.println("(a) " + new Templates().render(t));
        System.out.println("(b) " + new Formatter().format(t));
        new Pipeline().run(text -> System.out.println("(d) client listener: " + text), t);
        System.out.println("(e) " + Codecs.decodeAll(t));
        System.out.println("(f) " + Util.expand(t));
        System.out.println("(g) " + new Expander().replace(t));
        System.out.println("(h) " + new SafeExpander().replace(t));
        System.out.println("(i) " + new StrictExpander().replace(t));
    }
}
