package sc;

import java.io.ByteArrayOutputStream;
import java.io.ObjectOutputStream;
import java.io.Serializable;

/**
 * Writes one object of each scenario class and reads it back through each
 * entry, printing "entry<TAB>object<TAB>callbacks that ran" per read.
 */
public class Oracle {
    public static void main(String[] args) throws Exception {
        Object[] objects = {new PrivateCb(), new PublicCb(), new Ext(), new Point(1), new Queue(),
            Class.forName("lib.Cb").getDeclaredConstructor().newInstance(),
            Class.forName("shaded.lib.Cb").getDeclaredConstructor().newInstance()};
        for (String entry : new String[] {"plain", "filtered"}) {
            for (Object o : objects) {
                ByteArrayOutputStream b = new ByteArrayOutputStream();
                try (ObjectOutputStream out = new ObjectOutputStream(b)) {
                    out.writeObject((Serializable) o);
                }
                Rec.LOG.setLength(0);
                String outcome = "";
                try {
                    if (entry.equals("plain")) {
                        Entries.plain(Entries.of(b.toByteArray()));
                    } else {
                        Entries.filtered(Entries.of(b.toByteArray()));
                    }
                } catch (Exception e) {
                    outcome = "\tREJECTED " + e.getClass().getName();
                }
                String ran = Rec.LOG.length() == 0 ? "-" : Rec.LOG.substring(0, Rec.LOG.length() - 1);
                System.out.println("sc.Entries#" + entry + "\t" + o.getClass().getName() + "\t" + ran + outcome);
            }
        }
    }
}
