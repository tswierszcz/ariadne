package reachfixture;

/** Stands in for a plain (non-framework) dependency jar. */
public class VulnerableLib {
    public static void vulnerable() {
    }

    // Called only from FrameworkThing#doWork, never directly by app
    // code -> the pure FRAMEWORK case (no REACHABLE path exists at all).
    public static void frameworkOnlyVulnerable() {
    }

    // Never called from anywhere in this fixture -> NOT_REACHABLE.
    public static void unreachable() {
    }

    // An instance method a subclass inherits without overriding, so a
    // call through the subclass's static type still runs this code.
    public void riskyInstance() {
    }
}
