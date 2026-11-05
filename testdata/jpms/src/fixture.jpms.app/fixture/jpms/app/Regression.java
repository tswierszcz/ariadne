package fixture.jpms.app;

import fixture.jpms.api.Codec;
import fixture.jpms.api.Hasher;
import java.util.ServiceLoader;

// SERVICE_LOADING S4: two loaders for two services in one method must not
// mix their providers.
public class Regression {
    public static void twoLoaders() {
        ServiceLoader<Codec> codecs = ServiceLoader.load(Codec.class);
        ServiceLoader<Hasher> hashers = ServiceLoader.load(Hasher.class);
        for (Codec c : codecs) {
            System.out.println(c.encode("r"));
        }
        for (Hasher h : hashers) {
            System.out.println(h.hash("r"));
        }
    }
}
