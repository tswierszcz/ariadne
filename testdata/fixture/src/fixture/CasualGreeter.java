package fixture;

public class CasualGreeter implements Greeter {
    @Override
    public String greet(String name) {
        return "Hey " + name + "!";
    }
}
