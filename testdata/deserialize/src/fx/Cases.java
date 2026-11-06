package fx;

import com.thoughtworks.xstream.XStream;
import com.thoughtworks.xstream.security.AnyTypePermission;
import com.thoughtworks.xstream.security.NoTypePermission;
import java.io.InputStream;
import java.io.ObjectInputFilter;
import java.io.ObjectInputStream;

/** One deserialization entry per method, each under another policy. */
public class Cases {
    private final XStream shared = new XStream();

    public Object javaNone(InputStream in) throws Exception {
        return new ObjectInputStream(in).readObject();
    }

    public Object javaFilter(InputStream in) throws Exception {
        ObjectInputStream ois = new ObjectInputStream(in);
        ois.setObjectInputFilter(ObjectInputFilter.Config.createFilter("fx.Allowed;fx.allowed.**;!*"));
        return ois.readObject();
    }

    public Object javaFilterDynamic(InputStream in, String pattern) throws Exception {
        ObjectInputStream ois = new ObjectInputStream(in);
        ois.setObjectInputFilter(ObjectInputFilter.Config.createFilter(pattern));
        return ois.readObject();
    }

    public Object javaFilterOtherStream(InputStream a, InputStream b) throws Exception {
        ObjectInputStream filtered = new ObjectInputStream(a);
        filtered.setObjectInputFilter(ObjectInputFilter.Config.createFilter("!*"));
        ObjectInputStream plain = new ObjectInputStream(b);
        return plain.readObject();
    }

    public Object xsPlain(String xml) {
        return new XStream().fromXML(xml);
    }

    public Object xsAllow(String xml) {
        XStream x = new XStream();
        x.allowTypes(new Class[] {Allowed.class});
        return x.fromXML(xml);
    }

    public Object xsWildcard(String xml) {
        XStream x = new XStream();
        x.addPermission(NoTypePermission.NONE);
        x.allowTypesByWildcard(new String[] {"fx.allowed.**"});
        return x.fromXML(xml);
    }

    public Object xsHierarchy(String xml) {
        XStream x = new XStream();
        x.addPermission(NoTypePermission.NONE);
        x.allowTypeHierarchy(Base.class);
        return x.fromXML(xml);
    }

    public Object xsDefaultSecurity(String xml) {
        XStream x = new XStream();
        XStream.setupDefaultSecurity(x);
        return x.fromXML(xml);
    }

    public Object xsAnyAfterNone(String xml) {
        XStream x = new XStream();
        x.addPermission(NoTypePermission.NONE);
        x.addPermission(AnyTypePermission.ANY);
        return x.fromXML(xml);
    }

    public Object xsDeny(String xml) {
        XStream x = new XStream();
        x.addPermission(NoTypePermission.NONE);
        x.allowTypesByWildcard(new String[] {"fx.**"});
        x.denyTypes(new Class[] {Bad.class});
        return x.fromXML(xml);
    }

    public Object xsField(String xml) {
        return shared.fromXML(xml);
    }

    public Object xsEscape(String xml) {
        XStream x = new XStream();
        configure(x);
        return x.fromXML(xml);
    }

    static void configure(XStream x) {
        x.allowTypes(new String[] {"fx.Bad"});
    }
}
