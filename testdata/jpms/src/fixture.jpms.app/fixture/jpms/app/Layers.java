package fixture.jpms.app;

import fixture.jpms.api.Codec;
import java.util.ServiceLoader;

// SERVICE_LOADING S3 slice 3: load(ModuleLayer, Class).
public class Layers {
    // The boot layer: named modules only, never the class path.
    public static void boot() {
        for (Codec c : ServiceLoader.load(ModuleLayer.boot(), Codec.class)) {
            System.out.println(c.encode("boot"));
        }
    }

    // A layer only known at run time: stays a visibility gap.
    public static void given(ModuleLayer layer) {
        for (Codec c : ServiceLoader.load(layer, Codec.class)) {
            System.out.println(c.encode("given"));
        }
    }
}
