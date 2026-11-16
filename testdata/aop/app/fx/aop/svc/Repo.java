package fx.aop.svc;

public interface Repo {
    void save(Object o);

    Object find(long id);
}
