// Calls ServiceLoader.load(Codec.class) without `uses`: on the module path
// ServiceLoader throws ServiceConfigurationError, so no provider is loaded.
module fixture.jpms.nouses {
    requires fixture.jpms.api;
}
