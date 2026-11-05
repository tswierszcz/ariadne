module fixture.jpms.app {
    requires fixture.jpms.api;
    uses fixture.jpms.api.Codec;
    uses fixture.jpms.api.Hasher;
}
