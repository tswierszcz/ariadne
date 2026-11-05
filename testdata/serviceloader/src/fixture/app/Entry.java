package fixture.app;

import fixture.service.Decoder;
import java.util.ServiceLoader;
import java.util.stream.Stream;

public final class Entry {
    private static final ServiceLoader<Decoder> LOADER = ServiceLoader.load(Decoder.class);
    private static ServiceLoader<Decoder> CYCLE_A = Entry.CYCLE_B;
    private static ServiceLoader<Decoder> CYCLE_B = Entry.CYCLE_A;
    private ServiceLoader<Decoder> mutableLoader;
    private final ServiceLoader<Decoder> instanceLoader;

    public Entry() {
        this.instanceLoader = ServiceLoader.load(Decoder.class);
    }

    public static String run() {
        return createLoader().iterator().next().decode();
    }

    public static String runThroughParameter() {
        return passLoader(ServiceLoader.load(Decoder.class)).iterator().next().decode();
    }

    public static String runThroughStream() {
        return ServiceLoader.load(Decoder.class).stream().findFirst().get().get().decode();
    }

    public static Stream<ServiceLoader.Provider<Decoder>> streamWithoutConsuming() {
        return ServiceLoader.load(Decoder.class).stream();
    }

    public static String runThroughStreamCallback() {
        return ServiceLoader.load(Decoder.class).stream().map(provider -> provider.get()).findFirst().get().decode();
    }

    public static String runWithUnknownParameter(ServiceLoader<Decoder> loader) {
        return passLoader(loader).iterator().next().decode();
    }

    public static String runThroughStaticField() {
        return LOADER.iterator().next().decode();
    }

    public String runThroughFinalInstanceField() {
        return instanceLoader.iterator().next().decode();
    }

    public static String runThroughCyclicField() {
        return CYCLE_A.iterator().next().decode();
    }

    public String runThroughMutableField() {
        return mutableLoader.iterator().next().decode();
    }

    public static ServiceLoader<Decoder> loadWithoutConsuming() {
        return ServiceLoader.load(Decoder.class);
    }

    public static String runWithUnknownLoader() {
        return unknownLoader().iterator().next().decode();
    }

    private static ServiceLoader<Decoder> createLoader() {
        return ServiceLoader.load(Decoder.class);
    }

    private static ServiceLoader<Decoder> passLoader(ServiceLoader<Decoder> loader) {
        return loader;
    }

    private static ServiceLoader<Decoder> unknownLoader() {
        throw new UnsupportedOperationException("no known loader value");
    }
}
