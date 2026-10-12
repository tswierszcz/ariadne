package fixture;

public class FormalGreeter implements Greeter {
    @Override
    public String greet(String name) {
        return "Good day, " + name + ".";
    }
}
