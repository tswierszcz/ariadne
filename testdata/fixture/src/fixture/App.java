package fixture;

import java.util.function.Function;

/**
 * Tiny fixture exercising every call-graph edge kind ariadne's bytecode
 * walker needs to recognise:
 *  - invokespecial:   the FormalGreeter/CasualGreeter constructors
 *  - invokeinterface: greeter.greet(...) through the Greeter interface
 *  - invokestatic:    Util.shout(...)
 *  - invokedynamic:   the lambda passed to greetWith, and the Util::shout
 *                      method reference, both via LambdaMetafactory
 *
 * {@link #handle} stands in for a framework-detected entry point (see
 * {@link Entry}); {@link #main} stands in for the plain-main entry point.
 */
public class App {
    public static void main(String[] args) {
        new App().handle("world");
    }

    @Entry
    public String handle(String name) {
        Greeter formal = new FormalGreeter();
        Greeter casual = new CasualGreeter();

        String a = formal.greet(name);
        String b = casual.greet(name);

        String c = greetWith(name, n -> "Lambda says hi, " + n);
        String d = greetWith(name, Util::shout);

        return a + b + c + d;
    }

    private String greetWith(String name, Function<String, String> f) {
        return f.apply(name);
    }
}
