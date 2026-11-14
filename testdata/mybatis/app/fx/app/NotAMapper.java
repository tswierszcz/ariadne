package fx.app;

/** Not a mapper (no @Mapper, no mapper XML): the proxy catalog must not link it to MapperProxy. */
public interface NotAMapper {
    String name();
}
