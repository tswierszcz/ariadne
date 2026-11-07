package demo;

import java.io.ByteArrayInputStream;
import java.io.ObjectInputFilter;
import java.io.ObjectInputStream;
import javax.servlet.ServletRequest;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.util.HtmlUtils;

public class Endpoints {
    /** Java serialization of the String "hello". */
    static final String HELLO = "rO0ABXQABWhlbGxv";

    @PostMapping("/user-data")
    public Object userData(@RequestParam("token") String token) throws Exception {
        return Codec.read(Codec.decode(token));
    }

    @PostMapping("/constant")
    public Object constant(@RequestParam("ignored") String ignored) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode("rO0ABXQABWhlbGxv")))) {
            return in.readObject();
        }
    }

    @PostMapping("/two-sites")
    public Object twoSites(@RequestParam("token") String token) throws Exception {
        byte[] fromRequest = Codec.decode(token);
        byte[] literal = Codec.decode("rO0ABXQABWhlbGxv");
        System.out.println("  request bytes: " + fromRequest.length);
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(literal))) {
            return in.readObject();
        }
    }

    @PostMapping("/filtered")
    public Object filtered(@RequestParam("token") String token) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(token)))) {
            in.setObjectInputFilter(ObjectInputFilter.Config.createFilter("java.lang.String;!*"));
            return in.readObject();
        }
    }

    @PostMapping("/escaped")
    public Object escaped(@RequestParam("token") String token) throws Exception {
        if (token.isEmpty()) {
            return null;
        }
        String clean = HtmlUtils.htmlEscape(token);
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(clean)))) {
            return in.readObject();
        }
    }

    @PostMapping("/unknown")
    public Object unknown(@RequestParam("token") String token) throws Exception {
        Provider p = token.isEmpty() ? new LiteralProvider() : new OtherProvider();
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(p.bytes()))) {
            return in.readObject();
        }
    }

    @PostMapping("/servlet")
    public Object servlet(ServletRequest req) throws Exception {
        String t = req.getParameter("t");
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(t)))) {
            return in.readObject();
        }
    }

    @PostMapping("/implicit")
    public Object implicit(String token) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(token)))) {
            return in.readObject();
        }
    }

    static Object fromCli(String arg) throws Exception {
        try (ObjectInputStream in = new ObjectInputStream(new ByteArrayInputStream(Codec.decode(arg)))) {
            return in.readObject();
        }
    }

    public static void main(String[] args) throws Exception {
        Endpoints e = new Endpoints();
        java.io.ByteArrayOutputStream bos = new java.io.ByteArrayOutputStream();
        try (java.io.ObjectOutputStream out = new java.io.ObjectOutputStream(bos)) {
            out.writeObject(new Payload());
        }
        String payload = java.util.Base64.getEncoder().encodeToString(bos.toByteArray());
        System.out.println("userData: " + e.userData(payload).getClass().getName());
        System.out.println("constant: " + e.constant(payload));
        System.out.println("twoSites: " + e.twoSites(payload));
        try {
            e.filtered(payload);
            System.out.println("filtered: Payload accepted (unexpected)");
        } catch (java.io.InvalidClassException rejected) {
            System.out.println("filtered: Payload rejected by the filter (" + rejected.getMessage() + ")");
        }
        System.out.println("filtered: " + e.filtered(HELLO));
        System.out.println("escaped: " + e.escaped(HELLO));
        System.out.println("unknown: " + e.unknown(payload));
        System.out.println("servlet: " + e.servlet(new FakeRequest(HELLO)));
        System.out.println("implicit: " + e.implicit(HELLO));
        System.out.println("fromCli: " + fromCli(args.length > 0 ? args[0] : HELLO));
    }
}
