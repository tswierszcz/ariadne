package ac.api;

/** The starter's public API: a bean the autoconfiguration creates. Reaches nothing vulnerable. */
public class Greeter {
    public String greet(String name) {
        return "hello " + name;
    }
}
