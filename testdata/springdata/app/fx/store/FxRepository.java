package fx.store;

import fx.trace.Ran;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import org.springframework.data.repository.CrudRepository;

/**
 * The store module's base repository class (what SimpleJpaRepository is for
 * JPA): Spring Data's proxy dispatches CRUD methods to it. Its bodies record
 * every frame on the stack.
 */
public class FxRepository<T, ID> implements CrudRepository<T, ID> {
    private final Map<Object, T> rows = new LinkedHashMap<>();
    private final java.util.function.Function<T, ID> idOf;

    public FxRepository(java.util.function.Function<T, ID> idOf) { this.idOf = idOf; }

    public <S extends T> S save(S entity) { Ran.stack(); rows.put(idOf.apply(entity), entity); return entity; }
    public <S extends T> Iterable<S> saveAll(Iterable<S> entities) { entities.forEach(this::save); return entities; }
    public Optional<T> findById(ID id) { Ran.stack(); return Optional.ofNullable(rows.get(id)); }
    public boolean existsById(ID id) { return rows.containsKey(id); }
    public Iterable<T> findAll() { return new ArrayList<>(rows.values()); }
    public Iterable<T> findAllById(Iterable<ID> ids) { List<T> out = new ArrayList<>(); ids.forEach(i -> findById(i).ifPresent(out::add)); return out; }
    public long count() { Ran.stack(); return rows.size(); }
    public void deleteById(ID id) { rows.remove(id); }
    public void delete(T entity) { rows.remove(idOf.apply(entity)); }
    public void deleteAll(Iterable<? extends T> entities) { entities.forEach(this::delete); }
    public void deleteAll() { rows.clear(); }
}
