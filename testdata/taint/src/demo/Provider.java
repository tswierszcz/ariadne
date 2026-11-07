package demo;

interface Provider {
    byte[] bytes();
}

final class LiteralProvider implements Provider {
    public byte[] bytes() { return Codec.decode(Endpoints.HELLO); }
}

final class OtherProvider implements Provider {
    public byte[] bytes() { return Codec.decode(Endpoints.HELLO); }
}
