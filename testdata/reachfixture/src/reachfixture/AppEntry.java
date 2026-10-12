package reachfixture;

/**
 * Stands in for application code. Three methods exercise the three
 * reachable-ish outcomes internal/reach needs to tell apart:
 *  - callsFrameworkWhichCallsVulnerable: app -> framework -> vulnerable,
 *    a multi-hop path that only exists by continuing *through* framework
 *    code -> FRAMEWORK.
 *  - callsVulnerableDirectly: app -> vulnerable, one hop, bypassing the
 *    framework package entirely -> REACHABLE.
 *  - callsFrameworkEntryPointDirectly: app -> a framework-tagged method,
 *    one hop, landing directly on it -> REACHABLE even though the target
 *    is itself in a framework package (the section 3.6/5 exception).
 */
public class AppEntry {
    public void callsFrameworkWhichCallsVulnerable() {
        new FrameworkThing().doWork();
    }

    public void callsVulnerableDirectly() {
        VulnerableLib.vulnerable();
    }

    public void callsFrameworkEntryPointDirectly() {
        new FrameworkThing().riskyMethod();
    }

    // The static type here is the subclass, which does not override
    // riskyInstance() — the bytecode's Methodref owner is the subclass,
    // even though the method body that actually runs is VulnerableLib's.
    public void callsInheritedMethodThroughSubclass() {
        new VulnerableLibSub().riskyInstance();
    }
}
