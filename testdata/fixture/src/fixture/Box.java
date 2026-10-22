package fixture;

/**
 * A generic interface, implemented below with a concrete type argument
 * specifically so javac emits a synthetic bridge method (JVM_GRAPH G2) —
 * not yet exercised against real compiled bytecode in this repo (see
 * StringBox and BridgeDemo's own doc comments, and
 * scripts/build-fixtures.sh: this environment has no javac to regenerate
 * fixture-8.jar/fixture-21.jar with these new classes).
 */
public interface Box<T> {
    T get();
}
