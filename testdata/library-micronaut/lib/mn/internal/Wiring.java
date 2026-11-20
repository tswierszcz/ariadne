package mn.internal;

import jakarta.inject.Inject;
import jakarta.inject.Singleton;
import mn.api.Banner;
import mn.api.Component;

/** (w) an @Inject method the container runs when a client asks for a Component. */
@Singleton
class Wiring implements Component {
    private String described = "";

    @Inject
    void init(Banner banner) {
        described = Expand.expand("wiring " + banner.text() + " ${java:locale}");
    }

    @Override
    public String describe() {
        return described;
    }
}
