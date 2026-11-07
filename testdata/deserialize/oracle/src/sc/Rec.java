package sc;

/** Called by every callback, so the oracle sees what ran. */
public final class Rec {
    public static final StringBuilder LOG = new StringBuilder();

    public static void ran(String who) {
        LOG.append(who).append(',');
    }
}
