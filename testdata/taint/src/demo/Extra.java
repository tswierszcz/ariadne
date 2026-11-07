package demo;

import java.util.function.Function;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;

/** Edge cases for internal/taint's tests (not in the runnable demo). */
public class Extra {
    private String kept;

    /** A sink method reached only through a method reference. */
    static Object readString(String s) throws Exception {
        return Codec.read(Codec.decode(s));
    }

    @PostMapping("/lambda")
    public Object viaLambda(@RequestParam("t") String t) throws Exception {
        Function<String, Object> f = s -> {
            try {
                return readString(s);
            } catch (Exception e) {
                return null;
            }
        };
        return f.apply(t);
    }

    @PostMapping("/keep")
    public void keep(@RequestParam("t") String t) {
        kept = t;
    }

    /** Reads what keep() stored: through a field. */
    public Object fromField() throws Exception {
        return Codec.readOther(Codec.decode(kept));
    }

    @PostMapping("/deep")
    public Object deep(@RequestParam("t") String t) throws Exception {
        return d1(t);
    }

    static Object d1(String s) throws Exception { return d2(s); }
    static Object d2(String s) throws Exception { return d3(s); }
    static Object d3(String s) throws Exception { return d4(s); }
    static Object d4(String s) throws Exception { return d5(s); }
    static Object d5(String s) throws Exception { return d6(s); }
    static Object d6(String s) throws Exception { return d7(s); }
    static Object d7(String s) throws Exception { return d8(s); }
    static Object d8(String s) throws Exception { return d9(s); }
    static Object d9(String s) throws Exception { return Codec.readDeep(Codec.decode(s)); }

    /** Two callers of one sink helper: one with request data, one with a literal. */
    @PostMapping("/mixed")
    public Object mixed(@RequestParam("t") String t) throws Exception {
        return Codec.readShared(Codec.decode(t));
    }

    public Object literalOnly() throws Exception {
        return Codec.readShared(Codec.decode(Endpoints.HELLO));
    }
}
