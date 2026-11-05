package app.overload;

import lib.overload.B;
import lib.overload.C;
import lib.overload.Sub;

public class Main {
    public void callsBString() { B.parse("x"); }
    public void callsCStream() { C.parse((java.io.InputStream) null); }
    public void callsInheritedHandle() { new Sub().handle("x"); }
}
