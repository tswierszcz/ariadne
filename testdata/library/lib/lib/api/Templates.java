package lib.api;

/** (a) Public API: renders a template through a package-private engine. */
public class Templates {
    public String render(String template) {
        return Engine.expand(prepare(template));
    }

    /** Protected in a class a client can extend: API. */
    protected String prepare(String template) {
        return template;
    }

    /** A public member class of a public class: API. */
    public static class Options {
        public boolean strict() {
            return false;
        }
    }
}
