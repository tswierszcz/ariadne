package lib.api;

/** (a) Public API: renders a template through a package-private engine. */
public class Templates {
    public String render(String template) {
        return Engine.expand(template);
    }
}
