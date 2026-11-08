package javax.ariadnefx;

/** A product of the provider: handed out by CodecProvider#create. */
public interface Codec {
    String encode(String s);
}
