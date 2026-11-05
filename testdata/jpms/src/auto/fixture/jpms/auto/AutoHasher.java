package fixture.jpms.auto;

import fixture.jpms.api.Hasher;

public class AutoHasher implements Hasher {
    public AutoHasher() {
        System.out.println("ORACLE-PROVIDER fixture.jpms.auto.AutoHasher");
    }

    public int hash(String s) {
        return s.length();
    }
}
