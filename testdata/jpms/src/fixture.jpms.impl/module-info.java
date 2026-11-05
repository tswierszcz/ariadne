module fixture.jpms.impl {
    requires transitive fixture.jpms.api;
    provides fixture.jpms.api.Codec with fixture.jpms.impl.ModuleCodec, fixture.jpms.impl.FactoryCodec;
}
