package lib.api;

/** A service the library loads with ServiceLoader. */
public interface Codec {
    String decode(String s);
}
