package fixture;

/**
 * Box&lt;String&gt; forces javac to emit a synthetic bridge method
 * {@code Object get()} (the type-erased interface signature) that
 * delegates to this class's own {@code String get()} via invokevirtual.
 * This is the exact shape JVM_GRAPH G2's exact-descriptor CHA matching
 * needs to handle correctly without special-casing: the bridge method
 * itself satisfies Box#get()Ljava/lang/Object;'s exact descriptor and
 * resolves like any other declared method, and its own delegating call
 * carries the analysis on to the real, covariant-signature body (see
 * internal/hierarchy/resolve.go's ResolveExact doc comment, and
 * internal/callgraph/build.go's CHA-expansion comment).
 *
 * Once a JDK is available to regenerate fixture-8.jar/fixture-21.jar
 * (scripts/build-fixtures.sh), this is intended to let
 * TestCHA_overloadsAreNotConflated's hand-built-hierarchy coverage be
 * cross-checked against real compiled bytecode — see BridgeDemo.
 */
public class StringBox implements Box<String> {
    @Override
    public String get() {
        return "boxed";
    }
}
