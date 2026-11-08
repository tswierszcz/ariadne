package rt;

public final class PluginImpl implements Plugin {
    public String name() {
        Ran.here();
        return "impl";
    }
}
