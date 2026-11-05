package fixture.jpms.app;

import fixture.jpms.api.Codec;
import java.util.ServiceLoader;

public class Main {
    public static void main(String[] args) {
        for (Codec c : ServiceLoader.load(Codec.class)) {
            System.out.println(c.encode("x"));
        }
    }
}
