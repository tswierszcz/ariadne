package reachfixture;

/** Stands in for a dependency jar tagged as a framework package. */
public class FrameworkThing {
    public void doWork() {
        VulnerableLib.vulnerable();
        VulnerableLib.frameworkOnlyVulnerable();
    }

    public void riskyMethod() {
        // A framework method that is itself the vulnerable one, called
        // directly by the app.
    }
}
