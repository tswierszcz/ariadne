package fixture.provider;

import fixture.service.Decoder;

public final class DecoderProvider implements Decoder {
    public DecoderProvider() {
        System.out.println("ORACLE-PROVIDER fixture.provider.DecoderProvider");
    }

    @Override
    public String decode() {
        return "decoded";
    }
}
