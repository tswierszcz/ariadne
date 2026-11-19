package cd.api;

/** A value the library produces for CDI clients. */
public final class Banner {
    private final String text;
    public Banner(String text) { this.text = text; }
    public String text() { return text; }
}
