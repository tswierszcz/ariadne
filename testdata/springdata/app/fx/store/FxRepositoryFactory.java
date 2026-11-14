package fx.store;

import fx.app.Pet;
import org.springframework.data.repository.core.EntityInformation;
import org.springframework.data.repository.core.RepositoryInformation;
import org.springframework.data.repository.core.RepositoryMetadata;
import org.springframework.data.repository.core.support.AbstractEntityInformation;
import org.springframework.data.repository.core.support.RepositoryFactorySupport;

/** A minimal Spring Data store module: one entity type, FxRepository as the base class. */
public class FxRepositoryFactory extends RepositoryFactorySupport {
    @Override
    @SuppressWarnings("unchecked")
    public <T, ID> EntityInformation<T, ID> getEntityInformation(Class<T> domainClass) {
        return (EntityInformation<T, ID>) new AbstractEntityInformation<Pet, Long>(Pet.class) {
            public Long getId(Pet entity) { return entity.id; }
            public Class<Long> getIdType() { return Long.class; }
        };
    }

    @Override
    protected Object getTargetRepository(RepositoryInformation metadata) {
        return new FxRepository<Pet, Long>(p -> p.id);
    }

    @Override
    protected Class<?> getRepositoryBaseClass(RepositoryMetadata metadata) {
        return FxRepository.class;
    }
}
