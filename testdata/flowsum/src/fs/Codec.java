package fs;

interface Codec {
    String apply(String s);
}

class Upper implements Codec {
    public String apply(String s) {
        return s.toUpperCase();
    }
}
