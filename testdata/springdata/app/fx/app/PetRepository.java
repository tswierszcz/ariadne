package fx.app;

import org.springframework.data.repository.CrudRepository;

/** A Spring Data repository: Spring implements it with a JDK proxy (JdkDynamicAopProxy). */
public interface PetRepository extends CrudRepository<Pet, Long> {
}
