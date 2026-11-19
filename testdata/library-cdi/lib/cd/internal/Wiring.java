package cd.internal;

import cd.api.Banner;
import cd.api.Component;
import jakarta.enterprise.context.Dependent;
import jakarta.inject.Inject;

/** (w) A package-private @Inject initializer method. */
@Dependent
class Wiring implements Component {
    private String name = "";

    @Inject
    void init(Banner banner) {
        name = Expand.expand("${java:runtime}") + banner.text();
    }

    @Override
    public String name() {
        return name;
    }
}
