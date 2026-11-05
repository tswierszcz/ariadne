package fixture.jpms.nouses;

import fixture.jpms.api.Codec;
import java.util.ServiceLoader;

public class NoUses {
    public static void main(String[] args) {
        for (Codec c : ServiceLoader.load(Codec.class)) {
            System.out.println(c.encode("y"));
        }
    }
}
