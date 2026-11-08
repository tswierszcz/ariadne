package javax.ariadnefx.spi;

import javax.ariadnefx.Codec;

/** A service-provider interface in a javax.* package: CHA never expands calls on it. */
public interface CodecProvider {
    Codec create();
}
