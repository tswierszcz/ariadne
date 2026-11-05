package fixture.provider;

import fixture.service.Decoder;

public final class DecoderProvider implements Decoder {
    public DecoderProvider() {}

    @Override
    public String decode() {
        return "decoded";
    }
}
