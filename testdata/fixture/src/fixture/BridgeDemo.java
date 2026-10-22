package fixture;

/**
 * Exercises the two JVM_GRAPH G2 scenarios no other fixture class
 * covers: an interface call through a generic type that gets a
 * synthetic bridge method (see StringBox), and a default method
 * inherited without being overridden (see Greeter#defaultGreeting).
 * Deliberately not wired into App#handle, whose exact call counts
 * several existing tests (classfile_test.go's
 * TestParseAndExtractCalls_handleMethod) pin precisely.
 *
 * Not yet exercised against real compiled bytecode: this repo's own
 * environment has no javac to regenerate fixture-8.jar/fixture-21.jar
 * (scripts/build-fixtures.sh). JVM_GRAPH G2's resolution logic is
 * already verified against a hand-built hierarchy instead
 * (internal/callgraph/build_test.go's TestCHA_overloadsAreNotConflated,
 * internal/hierarchy/hierarchy_test.go's
 * TestResolveExact_overloadsAndDiamondInterface) — once a JDK is
 * available, regenerate the fixture jars and add a cross-check against
 * this class's real bytecode: callBridge()'s invokeinterface edge
 * should CHA-expand to StringBox's synthetic bridge method
 * ("get()Ljava/lang/Object;"), whose own invokevirtual call should lead
 * to StringBox's real "get()Ljava/lang/String;"; callDefault()'s
 * invokeinterface edge through FormalGreeter/CasualGreeter should
 * resolve directly to Greeter#defaultGreeting itself (G2's
 * resolveTarget), since neither subtype overrides it.
 */
public class BridgeDemo {
    public String callBridge() {
        Box<String> box = new StringBox();
        return box.get();
    }

    public String callDefault() {
        Greeter g = new FormalGreeter();
        return g.defaultGreeting();
    }
}
