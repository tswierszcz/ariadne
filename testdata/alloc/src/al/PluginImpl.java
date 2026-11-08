package al;

/** The provider: created only by ServiceLoader. */
public class PluginImpl implements Plugin {
    public PluginImpl() { Made.it(this); }
    public String name() { return "plugin"; }
}
