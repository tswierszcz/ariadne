package fx.app;

import fx.store.FxRepositoryFactory;
import fx.trace.Ran;

/**
 * Spring Data repository proxy (framework:proxies) oracle (ROADMAP B4):
 * main calls repository methods; the JVM runs Spring's JdkDynamicAopProxy and
 * Spring Data's interceptors down to the store's base repository, whose
 * bodies record the whole stack. scripts/springdata-oracle.sh writes the
 * record, cmd/ariadne TestSpringDataOracle compares the graph with it.
 */
public final class Main {
    static String query() {
        PetRepository pets = new FxRepositoryFactory().getRepository(PetRepository.class);
        pets.save(new Pet(1L, "Leo"));
        return pets.findById(1L).map(p -> p.name).orElse("?") + " of " + pets.count();
    }

    public static void main(String[] args) {
        System.out.println("# " + query());
        System.out.println("# method bodies that ran on the real JVM (owner#name+desc), from the store's stacks");
        for (String b : Ran.bodies) {
            System.out.println(b);
        }
    }
}
