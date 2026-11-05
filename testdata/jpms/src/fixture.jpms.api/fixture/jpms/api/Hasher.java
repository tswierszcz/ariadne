package fixture.jpms.api;

// A second service, so two loaders in one method can be told apart.
public interface Hasher {
    int hash(String s);
}
