package fx.nat;

/**
 * RUNTIME_NATIVE N5 n5-1 fixture: native boundaries. Compiled only — the
 * native library does not exist and the program is never run.
 */
public final class Main {
    static {
        System.loadLibrary("fxcodec");
    }

    /** Reachable from main: a model gap of kind "native". */
    static native int encode(byte[] data);

    /** Called by nothing: not reachable, not a model gap the rule considers. */
    static native void unused();

    static void loadByName(String name) {
        Runtime.getRuntime().loadLibrary(name);
    }

    public static void main(String[] args) {
        encode(new byte[0]);
        loadByName(args.length > 0 ? args[0] : "fxplugin");
    }
}
