package lib.overload;

/** handle(String) is declared here and inherited by Sub, which does not override it. */
public class Base {
    public void handle(String s) {}
    public void handle(Integer i) {}
}
